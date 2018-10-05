/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ResourceProfile represents the user configuration for the resource
type ResourceProfile struct {
	Tenant             string
	ID                 string // identifier of this resource
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // time when this resource becomes active and expires
	UsageTTL           time.Duration             // auto-expire the usage after this duration
	Limit              float64                   // limit value
	AllocationMessage  string                    // message returned by the winning resource on allocation
	Blocker            bool                      // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64  // Weight to sort the resources
	ThresholdIDs       []string // Thresholds to check after changing Limit
}

// TenantID returns unique identifier of the ResourceProfile in a multi-tenant environment
func (rp *ResourceProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

// ResourceUsage represents an usage counted
type ResourceUsage struct {
	Tenant     string
	ID         string // Unique identifier of this ResourceUsage, Eg: FreeSWITCH UUID
	ExpiryTime time.Time
	Units      float64 // Number of units used
}

func (ru *ResourceUsage) TenantID() string {
	return utils.ConcatenatedKey(ru.Tenant, ru.ID)
}

// isActive checks ExpiryTime at some time
func (ru *ResourceUsage) isActive(atTime time.Time) bool {
	return ru.ExpiryTime.IsZero() || ru.ExpiryTime.Sub(atTime) > 0
}

// clone duplicates ru
func (ru *ResourceUsage) Clone() (cln *ResourceUsage) {
	cln = new(ResourceUsage)
	*cln = *ru
	return
}

// Resource represents a resource in the system
// not thread safe, needs locking at process level
type Resource struct {
	Tenant string
	ID     string
	Usages map[string]*ResourceUsage
	TTLIdx []string         // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
	ttl    *time.Duration   // time to leave for this resource, picked up on each Resource initialization out of config
	tUsage *float64         // sum of all usages
	dirty  *bool            // the usages were modified, needs save, *bool so we only save if enabled in config
	rPrf   *ResourceProfile // for ordering purposes
}

// TenantID returns the unique ID in a multi-tenant environment
func (r *Resource) TenantID() string {
	return utils.ConcatenatedKey(r.Tenant, r.ID)
}

// removeExpiredUnits removes units which are expired from the resource
func (r *Resource) removeExpiredUnits() {
	var firstActive int
	for _, rID := range r.TTLIdx {
		if r, has := r.Usages[rID]; has && r.isActive(time.Now()) {
			break
		}
		firstActive += 1
	}
	if firstActive == 0 {
		return
	}
	for _, rID := range r.TTLIdx[:firstActive] {
		ru, has := r.Usages[rID]
		if !has {
			continue
		}
		delete(r.Usages, rID)
		*r.tUsage -= ru.Units
		if *r.tUsage < 0 { // something went wrong
			utils.Logger.Warning(
				fmt.Sprintf("resetting total usage for resourceID: %s, usage smaller than 0: %f", r.ID, *r.tUsage))
			r.tUsage = nil
		}
	}
	r.TTLIdx = r.TTLIdx[firstActive:]
	r.tUsage = nil
}

// totalUsage returns the sum of all usage units
func (r *Resource) totalUsage() (tU float64) {
	if r.tUsage == nil {
		var tu float64
		for _, ru := range r.Usages {
			tu += ru.Units
		}
		r.tUsage = &tu
	}
	if r.tUsage != nil {
		tU = *r.tUsage
	}
	return
}

// recordUsage records a new usage
func (r *Resource) recordUsage(ru *ResourceUsage) (err error) {
	if _, hasID := r.Usages[ru.ID]; hasID {
		return fmt.Errorf("duplicate resource usage with id: %s", ru.TenantID())
	}
	if r.ttl != nil && *r.ttl != -1 {
		if *r.ttl == 0 {
			return // no recording for ttl of 0
		}
		ru = ru.Clone() // don't influence the initial ru
		ru.ExpiryTime = time.Now().Add(*r.ttl)
	}
	r.Usages[ru.ID] = ru
	if r.tUsage != nil {
		*r.tUsage += ru.Units
	}
	if !ru.ExpiryTime.IsZero() {
		r.TTLIdx = append(r.TTLIdx, ru.ID)
	}
	return
}

// clearUsage clears the usage for an ID
func (r *Resource) clearUsage(ruID string) (err error) {
	ru, hasIt := r.Usages[ruID]
	if !hasIt {
		return fmt.Errorf("cannot find usage record with id: %s", ruID)
	}
	if !ru.ExpiryTime.IsZero() {
		for i, ruIDIdx := range r.TTLIdx {
			if ruIDIdx == ruID {
				r.TTLIdx = append(r.TTLIdx[:i], r.TTLIdx[i+1:]...)
				break
			}
		}
	}
	if r.tUsage != nil {
		*r.tUsage -= ru.Units
	}
	delete(r.Usages, ruID)
	return
}

// Resources is an orderable list of Resources based on Weight
type Resources []*Resource

// sort based on Weight
func (rs Resources) Sort() {
	sort.Slice(rs, func(i, j int) bool { return rs[i].rPrf.Weight > rs[j].rPrf.Weight })
}

// recordUsage will record the usage in all the resource limits, failing back on errors
func (rs Resources) recordUsage(ru *ResourceUsage) (err error) {
	var nonReservedIdx int // index of first resource not reserved
	for _, r := range rs {
		if err = r.recordUsage(ru); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<ResourceLimits>, err: %s", err.Error()))
			break
		}
		nonReservedIdx += 1
	}
	if err != nil {
		for _, r := range rs[:nonReservedIdx] {
			r.clearUsage(ru.ID) // best effort
		}
	}
	return
}

// clearUsage gives back the units to the pool
func (rs Resources) clearUsage(ruTntID string) (err error) {
	for _, r := range rs {
		if errClear := r.clearUsage(ruTntID); errClear != nil &&
			r.ttl != nil && *r.ttl != 0 { // we only consider not found error in case of ttl different than 0
			utils.Logger.Warning(fmt.Sprintf("<ResourceLimits>, clear ruID: %s, err: %s", ruTntID, errClear.Error()))
			err = errClear
		}
	}
	return
}

// tenantIDs returns list of TenantIDs in resources
func (rs Resources) tenantIDs() []*utils.TenantID {
	tntIDs := make([]*utils.TenantID, len(rs))
	for i, r := range rs {
		tntIDs[i] = &utils.TenantID{r.Tenant, r.ID}
	}
	return tntIDs
}

func (rs Resources) tenatIDsStr() []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.TenantID()
	}
	return ids
}

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) allocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	lockIDs := utils.PrefixSliceItems(rs.tenatIDsStr(), utils.ResourcesPrefix)
	guardian.Guardian.GuardIDs(config.CgrConfig().GeneralCfg().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	// Simulate resource usage
	for _, r := range rs {
		r.removeExpiredUnits()
		if _, hasID := r.Usages[ru.ID]; hasID { // update
			r.clearUsage(ru.ID)
		}
		if r.rPrf == nil {
			return "", fmt.Errorf("empty configuration for resourceID: %s", r.TenantID())
		}
		if r.rPrf.Limit >= r.totalUsage()+ru.Units {
			if alcMessage == "" {
				if r.rPrf.AllocationMessage != "" {
					alcMessage = r.rPrf.AllocationMessage
				} else {
					alcMessage = r.rPrf.ID
				}
			}
		}
	}
	if alcMessage == "" {
		return "", utils.ErrResourceUnavailable
	}
	if dryRun {
		return
	}
	err = rs.recordUsage(ru)
	if err != nil {
		return
	}
	return
}

// Pas the config as a whole so we can ask access concurrently
func NewResourceService(dm *DataManager, storeInterval time.Duration,
	thdS rpcclient.RpcClientConnection, filterS *FilterS,
	stringIndexedFields, prefixIndexedFields *[]string) (*ResourceService, error) {
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	return &ResourceService{dm: dm, thdS: thdS,
		lcEventResources:    make(map[string][]*utils.TenantID),
		storedResources:     make(utils.StringMap),
		storeInterval:       storeInterval,
		filterS:             filterS,
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields,
		stopBackup:          make(chan struct{})}, nil
}

// ResourceService is the service handling resources
type ResourceService struct {
	dm                  *DataManager                  // So we can load the data in cache and index it
	thdS                rpcclient.RpcClientConnection // allows applying filters based on stats
	filterS             *FilterS
	stringIndexedFields *[]string // speed up query on indexes
	prefixIndexedFields *[]string
	lcEventResources    map[string][]*utils.TenantID // cache recording resources for events in alocation phase
	lcERMux             sync.RWMutex                 // protects the lcEventResources
	storedResources     utils.StringMap              // keep a record of resources which need saving, map[resID]bool
	srMux               sync.RWMutex                 // protects storedResources
	storeInterval       time.Duration                // interval to dump data on
	stopBackup          chan struct{}                // control storing process
}

// Called to start the service
func (rS *ResourceService) ListenAndServe(exitChan chan bool) error {
	go rS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Called to shutdown the service
func (rS *ResourceService) Shutdown() error {
	utils.Logger.Info("<ResourceS> service shutdown initialized")
	close(rS.stopBackup)
	rS.storeResources()
	utils.Logger.Info("<ResourceS> service shutdown complete")
	return nil
}

// StoreResource stores the resource in DB and corrects dirty flag
func (rS *ResourceService) StoreResource(r *Resource) (err error) {
	if r.dirty == nil || !*r.dirty {
		return
	}
	if err = rS.dm.SetResource(r); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed saving Resource with ID: %s, error: %s",
				r.ID, err.Error()))
		return
	} else {
		*r.dirty = false
	}
	return
}

// storeResources represents one task of complete backup
func (rS *ResourceService) storeResources() {
	var failedRIDs []string
	for { // don't stop untill we store all dirty resources
		rS.srMux.Lock()
		rID := rS.storedResources.GetOne()
		if rID != "" {
			delete(rS.storedResources, rID)
		}
		rS.srMux.Unlock()
		if rID == "" {
			break // no more keys, backup completed
		}
		if rIf, ok := Cache.Get(utils.CacheResources, rID); !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<ResourceS> failed retrieving from cache resource with ID: %s", rID))
		} else if err := rS.StoreResource(rIf.(*Resource)); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		// randomize the CPU load and give up thread control
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		rS.srMux.Lock()
		for _, rID := range failedRIDs {
			rS.storedResources[rID] = true
		}
		rS.srMux.Unlock()
	}
}

// backup will regularly store resources changed to dataDB
func (rS *ResourceService) runBackup() {
	if rS.storeInterval <= 0 {
		return
	}
	for {
		select {
		case <-rS.stopBackup:
			return
		default:
		}
		rS.storeResources()
		time.Sleep(rS.storeInterval)
	}
}

// cachedResourcesForEvent attempts to retrieve cached resources for an event
// returns nil if event not cached or errors occur
// returns []Resource if negative reply was cached
func (rS *ResourceService) cachedResourcesForEvent(evUUID string) (rs Resources) {
	var shortCached bool
	rS.lcERMux.RLock()
	rIDs, has := rS.lcEventResources[evUUID]
	rS.lcERMux.RUnlock()
	if !has {
		if rIDsIf, has := Cache.Get(utils.CacheEventResources, evUUID); !has {
			return nil
		} else if rIDsIf != nil {
			rIDs = rIDsIf.([]*utils.TenantID)
		}
		shortCached = true
	}
	rs = make(Resources, len(rIDs))
	if len(rIDs) == 0 {
		return
	}
	lockIDs := make([]string, len(rIDs))
	for i, rTid := range rIDs {
		lockIDs[i] = utils.ResourcesPrefix + rTid.TenantID()
	}
	guardian.Guardian.GuardIDs(config.CgrConfig().GeneralCfg().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for i, rTid := range rIDs {
		if r, err := rS.dm.GetResource(rTid.Tenant, rTid.ID, true, true, ""); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ResourceS> force-uncaching resources for evUUID: <%s>, error: <%s>",
					evUUID, err.Error()))
			// on errors, cleanup cache so we recache
			if shortCached {
				Cache.Remove(utils.CacheEventResources, evUUID, true, "")
			} else {
				rS.lcERMux.Lock()
				delete(rS.lcEventResources, evUUID)
				rS.lcERMux.Unlock()
			}
			return nil
		} else {
			rs[i] = r
		}
	}
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceService) matchingResourcesForEvent(ev *utils.CGREvent, usageTTL *time.Duration) (rs Resources, err error) {
	matchingResources := make(map[string]*Resource)
	rIDs, err := matchingItemIDsForEvent(ev.Event, rS.stringIndexedFields, rS.prefixIndexedFields,
		rS.dm, utils.CacheResourceFilterIndexes, ev.Tenant, rS.filterS.cfg.FilterSCfg().IndexedSelects)
	if err != nil {
		return nil, err
	}
	for resName := range rIDs {
		rPrf, err := rS.dm.GetResourceProfile(ev.Tenant, resName, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if rPrf.ActivationInterval != nil && ev.Time != nil &&
			!rPrf.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			continue
		}
		if pass, err := rS.filterS.Pass(ev.Tenant, rPrf.FilterIDs,
			config.NewNavigableMap(ev.Event)); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		r, err := rS.dm.GetResource(rPrf.Tenant, rPrf.ID, true, true, "")
		if err != nil {
			return nil, err
		}
		if rPrf.Stored && r.dirty == nil {
			r.dirty = utils.BoolPointer(false)
		}
		if usageTTL != nil {
			if *usageTTL != 0 {
				r.ttl = usageTTL
			}
		} else if rPrf.UsageTTL >= 0 {
			r.ttl = utils.DurationPointer(rPrf.UsageTTL)
		}
		r.rPrf = rPrf
		matchingResources[rPrf.ID] = r
	}
	// All good, convert from Map to Slice so we can sort
	rs = make(Resources, len(matchingResources))
	i := 0
	for _, r := range matchingResources {
		rs[i] = r
		i++
	}
	rs.Sort()
	for i, r := range rs {
		if r.rPrf.Blocker { // blocker will stop processing
			rs = rs[:i+1]
			break
		}
	}
	return
}

// processThresholds will pass the event for resource to ThresholdS
func (rS *ResourceService) processThresholds(r *Resource) (err error) {
	if rS.thdS == nil {
		return
	}
	var thIDs []string
	if len(r.rPrf.ThresholdIDs) != 0 {
		if len(r.rPrf.ThresholdIDs) == 1 && r.rPrf.ThresholdIDs[0] == utils.META_NONE {
			return
		}
		thIDs = r.rPrf.ThresholdIDs
	}
	thEv := &ArgsProcessEvent{ThresholdIDs: thIDs,
		CGREvent: utils.CGREvent{
			Tenant: r.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				utils.EventType:  utils.ResourceUpdate,
				utils.ResourceID: r.ID,
				utils.Usage:      r.totalUsage()}}}
	var tIDs []string
	if err = rS.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
	}
	return
}

// V1ResourcesForEvent returns active resource configs matching the event
func (rS *ResourceService) V1ResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *Resources) (err error) {
	if missing := utils.MissingStructFields(&args.CGREvent, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	var mtcRLs Resources
	if args.UsageID != "" { // only cached if UsageID is present
		mtcRLs = rS.cachedResourcesForEvent(args.TenantID())
	}
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(&args.CGREvent, args.UsageTTL); err != nil {
			return err
		}
		Cache.Set(utils.CacheEventResources, args.TenantID(), mtcRLs.tenantIDs(), nil, true, "")
	}
	if len(mtcRLs) == 0 {
		return utils.ErrNotFound
	}
	*reply = mtcRLs
	return
}

// V1AuthorizeResources queries service to find if an Usage is allowed
func (rS *ResourceService) V1AuthorizeResources(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	var alcMessage string
	if missing := utils.MissingStructFields(&args.CGREvent, []string{"Tenant"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if missing := utils.MissingStructFields(&args, []string{"UsageID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	mtcRLs := rS.cachedResourcesForEvent(args.TenantID())
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(&args.CGREvent, args.UsageTTL); err != nil {
			return err
		}
		Cache.Set(utils.CacheEventResources, args.TenantID(), mtcRLs.tenantIDs(), nil, true, "")
	}
	if alcMessage, err = mtcRLs.allocateResource(
		&ResourceUsage{
			Tenant: args.CGREvent.Tenant,
			ID:     args.UsageID,
			Units:  args.Units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			err = utils.ErrResourceUnauthorized
			Cache.Set(utils.CacheEventResources, args.UsageID, nil, nil, true, "")
			return
		}
	}
	*reply = alcMessage
	return
}

// V1AllocateResource is called when a resource requires allocation
func (rS *ResourceService) V1AllocateResource(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	if missing := utils.MissingStructFields(&args.CGREvent, []string{"Tenant"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if missing := utils.MissingStructFields(&args, []string{"UsageID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	var wasCached bool
	mtcRLs := rS.cachedResourcesForEvent(args.UsageID)
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(&args.CGREvent, args.UsageTTL); err != nil {
			return
		}
	} else {
		wasCached = true
	}
	alcMsg, err := mtcRLs.allocateResource(
		&ResourceUsage{Tenant: args.CGREvent.Tenant, ID: args.UsageID, Units: args.Units}, false)
	if err != nil {
		return
	}

	// index it for matching out of cache
	var wasShortCached bool
	if wasCached {
		if _, has := Cache.Get(utils.CacheEventResources, args.UsageID); has {
			// remove from short cache to populate event cache
			wasShortCached = true
			Cache.Remove(utils.CacheEventResources, args.UsageID, true, "")
		}
	}
	if wasShortCached || !wasCached {
		rS.lcERMux.Lock()
		rS.lcEventResources[args.UsageID] = mtcRLs.tenantIDs()
		rS.lcERMux.Unlock()
	}
	// index it for storing
	for _, r := range mtcRLs {
		if rS.storeInterval == 0 || r.dirty == nil {
			continue
		}
		if rS.storeInterval == -1 {
			rS.StoreResource(r)
		} else {
			*r.dirty = true // mark it to be saved
			rS.srMux.Lock()
			rS.storedResources[r.TenantID()] = true
			rS.srMux.Unlock()
		}
		rS.processThresholds(r)
	}
	*reply = alcMsg
	return
}

// V1ReleaseResource is called when we need to clear an allocation
func (rS *ResourceService) V1ReleaseResource(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	if missing := utils.MissingStructFields(&args.CGREvent, []string{"Tenant"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if missing := utils.MissingStructFields(&args, []string{"UsageID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	mtcRLs := rS.cachedResourcesForEvent(args.UsageID)
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(&args.CGREvent, args.UsageTTL); err != nil {
			return
		}
	}
	mtcRLs.clearUsage(args.UsageID)
	rS.lcERMux.Lock()
	delete(rS.lcEventResources, args.UsageID)
	rS.lcERMux.Unlock()
	if rS.storeInterval != -1 {
		rS.srMux.Lock()
	}
	for _, r := range mtcRLs {
		if r.dirty != nil {
			if rS.storeInterval == -1 {
				rS.StoreResource(r)
			} else {
				*r.dirty = true // mark it to be saved
				rS.storedResources[r.TenantID()] = true
			}
		}
		rS.processThresholds(r)
	}
	if rS.storeInterval != -1 {
		rS.srMux.Unlock()
	}
	*reply = utils.OK
	return nil
}
