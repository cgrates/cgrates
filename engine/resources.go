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
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
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

	lkID string // holds the reference towards guardian lock key
}

// TenantID returns unique identifier of the ResourceProfile in a multi-tenant environment
func (rp *ResourceProfile) TenantID() string {
	return utils.ConcatenatedKey(rp.Tenant, rp.ID)
}

// resourceProfileLockKey returns the ID used to lock a resourceProfile with guardian
func resourceProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheResourceProfiles, tnt, id)
}

// lock will lock the resourceProfile using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (rp *ResourceProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			resourceProfileLockKey(rp.Tenant, rp.ID))
	}
	rp.lkID = lkID
}

// unlock will unlock the resourceProfile and clear rp.lkID
func (rp *ResourceProfile) unlock() {
	if rp.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(rp.lkID)
	rp.lkID = utils.EmptyString
}

// isLocked returns the locks status of this resourceProfile
func (rp *ResourceProfile) isLocked() bool {
	return rp.lkID != utils.EmptyString
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
	TTLIdx []string         // holds ordered list of ResourceIDs based on their TTL, empty if feature is disableda
	lkID   string           // ID of the lock used when matching the resource
	ttl    *time.Duration   // time to leave for this resource, picked up on each Resource initialization out of config
	tUsage *float64         // sum of all usages
	dirty  *bool            // the usages were modified, needs save, *bool so we only save if enabled in config
	rPrf   *ResourceProfile // for ordering purposes
}

// resourceLockKey returns the ID used to lock a resource with guardian
func resourceLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheResources, tnt, id)
}

// lock will lock the resource using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (r *Resource) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			resourceLockKey(r.Tenant, r.ID))
	}
	r.lkID = lkID
}

// unlock will unlock the resource and clear r.lkID
func (r *Resource) unlock() {
	if r.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(r.lkID)
	r.lkID = utils.EmptyString
}

// isLocked returns the locks status of this resource
func (r *Resource) isLocked() bool {
	return r.lkID != utils.EmptyString
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

// TotalUsage returns the sum of all usage units
// Exported to be used in FilterS
func (r *Resource) TotalUsage() (tU float64) {
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

// unlock will unlock resources part of this slice
func (rs Resources) unlock() {
	for _, r := range rs {
		r.unlock()
		if r.rPrf != nil {
			r.rPrf.unlock()
		}
	}
}

// resIDsMp returns a map of resource IDs which is used for caching
func (rs Resources) resIDsMp() (mp utils.StringMap) {
	mp = make(utils.StringMap)
	for _, r := range rs {
		mp[r.ID] = true
	}
	return mp
}

func (rs Resources) tenatIDs() []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.TenantID()
	}
	return ids
}

func (rs Resources) IDs() []string {
	ids := make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	return ids
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

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) allocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	// Simulate resource usage
	for _, r := range rs {
		r.removeExpiredUnits()
		if _, hasID := r.Usages[ru.ID]; hasID && !dryRun { // update
			r.clearUsage(ru.ID) // clearUsage returns error only when ru.ID does not exist in the Usages map
		}
		if r.rPrf == nil {
			err = fmt.Errorf("empty configuration for resourceID: %s", r.TenantID())
			return
		}
		if alcMessage == utils.EmptyString &&
			(r.rPrf.Limit >= r.TotalUsage()+ru.Units || r.rPrf.Limit == -1) {
			alcMessage = utils.FirstNonEmpty(r.rPrf.AllocationMessage, r.rPrf.ID)
		}
	}
	if alcMessage == "" {
		err = utils.ErrResourceUnavailable
		return
	}
	if dryRun {
		return
	}
	if err = rs.recordUsage(ru); err != nil {
		return utils.EmptyString, err
	}
	return
}

// NewResourceService  returns a new ResourceService
func NewResourceService(dm *DataManager, cgrcfg *config.CGRConfig,
	filterS *FilterS, connMgr *ConnManager) (*ResourceService, error) {
	return &ResourceService{dm: dm,
		storedResources: make(utils.StringMap),
		cgrcfg:          cgrcfg,
		filterS:         filterS,
		loopStoped:      make(chan struct{}),
		stopBackup:      make(chan struct{}),
		connMgr:         connMgr}, nil

}

// ResourceService is the service handling resources
type ResourceService struct {
	dm              *DataManager // So we can load the data in cache and index it
	filterS         *FilterS
	storedResources utils.StringMap // keep a record of resources which need saving, map[resID]bool
	srMux           sync.RWMutex    // protects storedResources
	cgrcfg          *config.CGRConfig
	stopBackup      chan struct{} // control storing process
	loopStoped      chan struct{}
	connMgr         *ConnManager
}

// Called to start the service
func (rS *ResourceService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ResourceS))
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

// backup will regularly store resources changed to dataDB
func (rS *ResourceService) runBackup() {
	storeInterval := rS.cgrcfg.ResourceSCfg().StoreInterval
	if storeInterval <= 0 {
		rS.loopStoped <- struct{}{}
		return
	}
	for {
		rS.storeResources()
		select {
		case <-rS.stopBackup:
			rS.loopStoped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
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
		rIf, ok := Cache.Get(utils.CacheResources, rID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache resource with ID: %s", utils.ResourceS, rID))
			continue
		}
		r := rIf.(*Resource)
		r.lock(utils.EmptyString)
		if err := rS.storeResource(r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		r.unlock()
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

// StoreResource stores the resource in DB and corrects dirty flag
func (rS *ResourceService) storeResource(r *Resource) (err error) {
	if r.dirty == nil || !*r.dirty {
		return
	}
	if err = rS.dm.SetResource(r); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed saving Resource with ID: %s, error: %s",
				r.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if err = rS.dm.CacheDataFromDB(utils.ResourcesPrefix, []string{r.TenantID()}, true); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed caching Resource with ID: %s, error: %s",
				r.TenantID(), err.Error()))
		return
	}
	*r.dirty = false
	return
}

// storeMatchedResources will store the list of resources based on the StoreInterval
func (rS *ResourceService) storeMatchedResources(mtcRLs Resources) (err error) {
	if rS.cgrcfg.ResourceSCfg().StoreInterval == 0 {
		return
	}
	if rS.cgrcfg.ResourceSCfg().StoreInterval > 0 {
		rS.srMux.Lock()
		defer rS.srMux.Unlock()
	}
	for _, r := range mtcRLs {
		if r.dirty != nil {
			*r.dirty = true // mark it to be saved
			if rS.cgrcfg.ResourceSCfg().StoreInterval > 0 {
				rS.storedResources[r.TenantID()] = true
				continue
			}
			if err = rS.storeResource(r); err != nil {
				return
			}
		}

	}
	return
}

// processThresholds will pass the event for resource to ThresholdS
func (rS *ResourceService) processThresholds(rs Resources, argDispatcher *utils.ArgDispatcher) (err error) {
	if len(rS.cgrcfg.ResourceSCfg().ThresholdSConns) == 0 {
		return
	}

	var withErrs bool
	for _, r := range rs {
		var thIDs []string
		if len(r.rPrf.ThresholdIDs) != 0 {
			if len(r.rPrf.ThresholdIDs) == 1 &&
				r.rPrf.ThresholdIDs[0] == utils.META_NONE {
				continue
			}
			thIDs = r.rPrf.ThresholdIDs
		}

		thEv := &ArgsProcessEvent{ThresholdIDs: thIDs,
			CGREvent: &utils.CGREvent{
				Tenant: r.Tenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					utils.EventType:  utils.ResourceUpdate,
					utils.ResourceID: r.ID,
					utils.Usage:      r.TotalUsage(),
				},
			},
			ArgDispatcher: argDispatcher,
		}
		var tIDs []string
		if err = rS.connMgr.Call(rS.cgrcfg.ResourceSCfg().ThresholdSConns, nil,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with %s.",
					utils.ResourceS, err.Error(), thEv, utils.ThresholdS))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceService) matchingResourcesForEvent(ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (rs Resources, err error) {
	var rIDs utils.StringMap
	if x, ok := Cache.Get(utils.CacheEventResources, evUUID); ok { // The ResourceIDs were cached as utils.StringMap{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		rIDs = x.(utils.StringMap)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				Cache.Remove(utils.CacheEventResources, evUUID,
					cacheCommit(utils.NonTransactional), utils.NonTransactional)
			}
		}()
	} else { // select the resourceIDs out of dataDB
		rIDs, err = MatchingItemIDsForEvent(ev.Event,
			rS.cgrcfg.ResourceSCfg().StringIndexedFields,
			rS.cgrcfg.ResourceSCfg().PrefixIndexedFields,
			rS.dm, utils.CacheResourceFilterIndexes, ev.Tenant,
			rS.cgrcfg.ResourceSCfg().IndexedSelects,
			rS.cgrcfg.ResourceSCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheEventResources, evUUID, nil, nil, true, "") // cache negative match
			}
			return
		}
	}
	evNm := utils.MapStorage{utils.MetaReq: ev.Event}
	rs = make(Resources, 0, len(rIDs))
	for resName := range rIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			resourceProfileLockKey(ev.Tenant, resName))
		var rPrf *ResourceProfile
		if rPrf, err = rS.dm.GetResourceProfile(ev.Tenant, resName,
			true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			rs.unlock()
			return
		}
		rPrf.lock(lkPrflID)
		if rPrf.ActivationInterval != nil && ev.Time != nil &&
			!rPrf.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			rPrf.unlock()
			continue
		}
		if pass, err := rS.filterS.Pass(ev.Tenant, rPrf.FilterIDs,
			evNm); err != nil {
			rPrf.unlock()
			rs.unlock()
			return nil, err
		} else if !pass {
			rPrf.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			resourceLockKey(rPrf.Tenant, rPrf.ID))
		r, err := rS.dm.GetResource(rPrf.Tenant, rPrf.ID, true, true, "")
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			rPrf.unlock()
			rs.unlock()
			return nil, err
		}
		r.lock(lkID) // pass the lock into resource so we have it as reference
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
		rs = append(rs, r)
	}

	if len(rs) == 0 {
		return nil, utils.ErrNotFound
	}
	// All good, convert from Map to Slice so we can sort
	rs.Sort()
	for i, r := range rs {
		if r.rPrf.Blocker && i != len(rs)-1 { // blocker will stop processing and we are not at last index
			Resources(rs[i+1:]).unlock()
			rs = rs[:i+1]
			break
		}
	}
	Cache.Set(utils.CacheEventResources, evUUID, rs.resIDsMp(), nil, true, "")
	return
}

// V1ResourcesForEvent returns active resource configs matching the event
func (rS *ResourceService) V1ResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *Resources) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.UsageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	// RPC caching
	if config.CgrConfig().CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent, args.TenantID())
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*Resources)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(args.CGREvent, args.UsageID, args.UsageTTL); err != nil {
		return err
	}
	*reply = mtcRLs
	mtcRLs.unlock()
	return
}

// V1AuthorizeResources queries service to find if an Usage is allowed
func (rS *ResourceService) V1AuthorizeResources(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.UsageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	// RPC caching
	if config.CgrConfig().CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources, args.TenantID())
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(args.CGREvent, args.UsageID, args.UsageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	var alcMessage string
	if alcMessage, err = mtcRLs.allocateResource(
		&ResourceUsage{
			Tenant: args.CGREvent.Tenant,
			ID:     args.UsageID,
			Units:  args.Units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			err = utils.ErrResourceUnauthorized
		}
		return
	}
	*reply = alcMessage
	return
}

// V1AllocateResource is called when a resource requires allocation
func (rS *ResourceService) V1AllocateResource(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.UsageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	// RPC caching
	if config.CgrConfig().CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources, args.TenantID())
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(args.CGREvent, args.UsageID,
		args.UsageTTL); err != nil {
		return err
	}
	defer mtcRLs.unlock()

	var alcMsg string
	if alcMsg, err = mtcRLs.allocateResource(
		&ResourceUsage{Tenant: args.CGREvent.Tenant, ID: args.UsageID,
			Units: args.Units}, false); err != nil {
		return
	}

	// index it for storing
	if err = rS.storeMatchedResources(mtcRLs); err != nil {
		return
	}
	if err = rS.processThresholds(mtcRLs, args.ArgDispatcher); err != nil {
		return
	}
	*reply = alcMsg
	return
}

// V1ReleaseResource is called when we need to clear an allocation
func (rS *ResourceService) V1ReleaseResource(args utils.ArgRSv1ResourceUsage, reply *string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args.CGREvent, []string{utils.Tenant, utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.UsageID == "" {
		return utils.NewErrMandatoryIeMissing(utils.UsageID)
	}

	// RPC caching
	if config.CgrConfig().CacheCfg()[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources, args.TenantID())
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var mtcRLs Resources
	if mtcRLs, err = rS.matchingResourcesForEvent(args.CGREvent, args.UsageID,
		args.UsageTTL); err != nil {
		return
	}
	defer mtcRLs.unlock()

	if err = mtcRLs.clearUsage(args.UsageID); err != nil {
		return
	}

	// Handle storing
	if err = rS.storeMatchedResources(mtcRLs); err != nil {
		return
	}
	if err = rS.processThresholds(mtcRLs, args.ArgDispatcher); err != nil {
		return
	}

	*reply = utils.OK
	return
}

// GetResource returns a resource configuration
func (rS *ResourceService) V1GetResource(arg *utils.TenantID, reply *Resource) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cgrcfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		resourceLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	res, err := rS.dm.GetResource(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return err
	} else {
		*reply = *res
	}
	return nil
}

// Reload stops the backupLoop and restarts it
func (rS *ResourceService) Reload() {
	close(rS.stopBackup)
	<-rS.loopStoped // wait until the loop is done
	rS.stopBackup = make(chan struct{})
	go rS.runBackup()
}

// StartLoop starts the gorutine with the backup loop
func (rS *ResourceService) StartLoop() {
	go rS.runBackup()
}
