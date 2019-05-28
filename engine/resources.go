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

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) allocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	lockIDs := utils.PrefixSliceItems(rs.tenatIDs(), utils.ResourcesPrefix)
	guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
		// Simulate resource usage
		for _, r := range rs {
			r.removeExpiredUnits()
			if _, hasID := r.Usages[ru.ID]; hasID && !dryRun { // update
				r.clearUsage(ru.ID)
			}
			if r.rPrf == nil {
				err = fmt.Errorf("empty configuration for resourceID: %s", r.TenantID())
				return
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
			err = utils.ErrResourceUnavailable
			return
		}
		if dryRun {
			return
		}
		err = rs.recordUsage(ru)
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, lockIDs...)
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
	storedResources     utils.StringMap // keep a record of resources which need saving, map[resID]bool
	srMux               sync.RWMutex    // protects storedResources
	storeInterval       time.Duration   // interval to dump data on
	stopBackup          chan struct{}   // control storing process
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

// processThresholds will pass the event for resource to ThresholdS
func (rS *ResourceService) processThresholds(r *Resource, argDispatcher *utils.ArgDispatcher) (err error) {
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
		CGREvent: &utils.CGREvent{
			Tenant: r.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				utils.EventType:  utils.ResourceUpdate,
				utils.ResourceID: r.ID,
				utils.Usage:      r.totalUsage(),
			},
		},
		ArgDispatcher: argDispatcher,
	}
	var tIDs []string
	if err = rS.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing event %+v with %s.",
				utils.ResourceS, err.Error(), thEv, utils.ThresholdS))
	}
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceService) matchingResourcesForEvent(ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (rs Resources, err error) {
	matchingResources := make(map[string]*Resource)
	var isCached bool
	var rIDs utils.StringMap
	if x, ok := Cache.Get(utils.CacheEventResources, evUUID); ok { // The ResourceIDs were cached as utils.StringMap{"resID":bool}
		isCached = true
		if x == nil {
			return nil, utils.ErrNotFound
		}
		rIDs = x.(utils.StringMap)
	} else { // select the resourceIDs out of dataDB
		rIDs, err = MatchingItemIDsForEvent(ev.Event, rS.stringIndexedFields, rS.prefixIndexedFields,
			rS.dm, utils.CacheResourceFilterIndexes, ev.Tenant, rS.filterS.cfg.ResourceSCfg().IndexedSelects)
	}
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheEventResources, evUUID, nil, nil, true, "") // cache negative match
		}
		return
	}
	lockIDs := utils.PrefixSliceItems(rs.IDs(), utils.ResourcesPrefix)
	guardian.Guardian.Guard(func() (gIface interface{}, gErr error) {
		for resName := range rIDs {
			var rPrf *ResourceProfile
			if rPrf, err = rS.dm.GetResourceProfile(ev.Tenant, resName,
				true, true, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					continue
				}
				return
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
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, lockIDs...)
	if err != nil {
		if isCached {
			Cache.Remove(utils.CacheEventResources, evUUID,
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
		}
		return
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

	var alcMsg string
	if alcMsg, err = mtcRLs.allocateResource(
		&ResourceUsage{Tenant: args.CGREvent.Tenant, ID: args.UsageID,
			Units: args.Units}, false); err != nil {
		return
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
		rS.processThresholds(r, args.ArgDispatcher)
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
		return err
	}
	mtcRLs.clearUsage(args.UsageID)

	// Handle storing
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
		rS.processThresholds(r, args.ArgDispatcher)
	}
	if rS.storeInterval != -1 {
		rS.srMux.Unlock()
	}

	*reply = utils.OK
	return
}

// GetResource returns a resource configuration
func (rS *ResourceService) V1GetResource(arg *utils.TenantID, reply *Resource) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if res, err := rS.dm.GetResource(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = *res
	}
	return nil
}
