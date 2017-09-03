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

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ResourceCfg represents the user configuration for the resource
type ResourceCfg struct {
	ID                 string                    // identifier of this resource
	Filters            []*RequestFilter          // filters for the request
	ActivationInterval *utils.ActivationInterval // time when this resource becomes active and expires
	UsageTTL           time.Duration             // auto-expire the usage after this duration
	Limit              float64                   // limit value
	AllocationMessage  string                    // message returned by the winning resource on allocation
	Blocker            bool                      // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64  // Weight to sort the resources
	Thresholds         []string // Thresholds to check after changing Limit
}

// ResourceUsage represents an usage counted
type ResourceUsage struct {
	ID         string // Unique identifier of this ResourceUsage, Eg: FreeSWITCH UUID
	ExpiryTime time.Time
	Units      float64 // Number of units used
}

// isActive checks ExpiryTime at some time
func (ru *ResourceUsage) isActive(atTime time.Time) bool {
	return ru.ExpiryTime.IsZero() || ru.ExpiryTime.Sub(atTime) > 0
}

// Resource represents a resource in the system
// not thread safe, needs locking at process level
type Resource struct {
	ID     string
	Usages map[string]*ResourceUsage
	TTLIdx []string     // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
	tUsage *float64     // sum of all usages
	dirty  *bool        // the usages were modified, needs save, *bool so we only save if enabled in config
	rCfg   *ResourceCfg // for ordering purposes
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
		return fmt.Errorf("duplicate resource usage with id: %s", ru.ID)
	}
	r.Usages[ru.ID] = ru
	if r.tUsage != nil {
		*r.tUsage += ru.Units
	}
	return
}

// clearUsage clears the usage for an ID
func (r *Resource) clearUsage(ruID string) (err error) {
	ru, hasIt := r.Usages[ruID]
	if !hasIt {
		return fmt.Errorf("Cannot find usage record with id: %s", ruID)
	}
	delete(r.Usages, ruID)
	if r.tUsage != nil {
		*r.tUsage -= ru.Units
	}
	return
}

// Resources is an orderable list of Resources based on Weight
type Resources []*Resource

// sort based on Weight
func (rs Resources) Sort() {
	sort.Slice(rs, func(i, j int) bool { return rs[i].rCfg.Weight > rs[j].rCfg.Weight })
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
func (rs Resources) clearUsage(ruID string) (err error) {
	for _, r := range rs {
		if errClear := r.clearUsage(ruID); errClear != nil {
			utils.Logger.Warning(fmt.Sprintf("<ResourceLimits>, err: %s", errClear.Error()))
			err = errClear
		}
	}
	return
}

// ids returns list of resource IDs in resources
func (rs Resources) ids() (ids []string) {
	ids = make([]string, len(rs))
	for i, r := range rs {
		ids[i] = r.ID
	}
	return
}

// AllocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) AllocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return utils.META_NONE, nil
	}
	lockIDs := utils.PrefixSliceItems(rs.ids(), utils.ResourcesPrefix)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	// Simulate resource usage
	for _, r := range rs {
		if r.rCfg.Limit >= r.totalUsage()+ru.Units {
			if alcMessage == "" {
				alcMessage = r.rCfg.AllocationMessage
			}
			if alcMessage == "" { // rl.AllocationMessage is not populated
				alcMessage = r.rCfg.ID
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
	return
}

// Pas the config as a whole so we can ask access concurrently
func NewResourceService(cfg *config.CGRConfig, dataDB DataDB, statS rpcclient.RpcClientConnection) (*ResourceService, error) {
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	return &ResourceService{dataDB: dataDB, statS: statS}, nil
}

// ResourceService is the service handling resources
type ResourceService struct {
	cfg             *config.CGRConfig
	dataDB          DataDB // So we can load the data in cache and index it
	statS           rpcclient.RpcClientConnection
	eventResources  map[string][]string // map[ruID][]string{rID} for faster queries
	erMux           sync.RWMutex
	storedResources utils.StringMap // keep a record of resources which need saving, map[resID]bool
	srMux           sync.RWMutex
	stopBackup      chan struct{} // control storing process
	backupInterval  time.Duration
}

// Called to start the service
func (rS *ResourceService) ListenAndServe(exitChan chan bool) error {
	go rS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Called to shutdown the service
func (rS *ResourceService) ServiceShutdown() error {
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
	if err = rS.dataDB.SetResource(r); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed saving Resource with ID: %s, error: %s",
				r.ID, err.Error()))
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
		if rIf, ok := cache.Get(utils.ResourcesPrefix + rID); !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<ResourceS> failed retrieving from cache resource with ID: %s"))
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
	if rS.backupInterval <= 0 {
		return
	}
	for {
		select {
		case <-rS.stopBackup:
			return
		}
		rS.storeResources()
	}
	time.Sleep(rS.backupInterval)
}

// cachedResourcesForEvent attempts to retrieve cached resources for an event
// returns nil if event not cached or errors occur
func (rS *ResourceService) cachedResourcesForEvent(evUUID string) (rs Resources) {
	rS.erMux.RLock()
	rIDs, has := rS.eventResources[evUUID]
	rS.erMux.RUnlock()
	if !has {
		return nil
	}
	lockIDs := utils.PrefixSliceItems(rIDs, utils.ResourcesPrefix)
	guardian.Guardian.GuardIDs(rS.cfg.LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	rs = make(Resources, len(rIDs))
	for i, rID := range rIDs {
		if r, err := rS.dataDB.GetResource(rID, false, ""); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ResourceS> force-uncaching resources for evUUID: <%s>, error: <%s>",
					evUUID, err.Error()))
			rS.erMux.Lock()
			delete(rS.eventResources, evUUID)
			rS.erMux.Unlock()
			return nil
		} else {
			rs[i] = r
		}
	}
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceService) matchingResourcesForEvent(ev map[string]interface{}) (rs Resources, err error) {
	matchingResources := make(map[string]*Resource)
	rIDs, err := matchingItemIDsForEvent(ev, rS.dataDB, utils.ResourcesIndex)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(rIDs.Slice(), utils.ResourcesPrefix)
	guardian.Guardian.GuardIDs(rS.cfg.LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for resName := range rIDs {
		rCfg, err := rS.dataDB.GetResourceCfg(resName, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if rCfg.ActivationInterval != nil &&
			!rCfg.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		passAllFilters := true
		for _, fltr := range rCfg.Filters {
			if pass, err := fltr.Pass(ev, "", rS.statS); err != nil {
				return nil, err
			} else if !pass {
				passAllFilters = false
				continue
			}
		}
		if !passAllFilters {
			continue
		}
		r, err := rS.dataDB.GetResource(rCfg.ID, false, "")
		if err != nil {
			return nil, err
		}
		if rCfg.Stored {
			r.dirty = utils.BoolPointer(false)
		}
		r.rCfg = rCfg
		matchingResources[rCfg.ID] = r // Cannot save it here since we could have errors after and resource will remain unused
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
		if r.rCfg.Blocker { // blocker will stop processing
			rs = rs[:i+1]
			break
		}
	}
	return
}

// V1ResourcesForEvent returns active resource configs matching the event
func (rS *ResourceService) V1ResourcesForEvent(ev map[string]interface{}, reply *[]*ResourceCfg) error {
	matchingRLForEv, err := rS.matchingResourcesForEvent(ev)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if len(matchingRLForEv) == 0 {
		return utils.ErrNotFound
	}
	for _, r := range matchingRLForEv {
		*reply = append(*reply, r.rCfg)
	}
	return nil
}

// V1AllowUsage queries service to find if an Usage is allowed
func (rS *ResourceService) V1AllowUsage(args utils.AttrRLsResourceUsage, allow *bool) (err error) {
	mtcRLs := rS.cachedResourcesForEvent(args.UsageID)
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(args.Event); err != nil {
			return err
		}
	}
	if _, err = mtcRLs.AllocateResource(
		&ResourceUsage{ID: args.UsageID,
			Units: args.Units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			return // not error but still not allowed
		}
		return err
	}
	*allow = true
	return
}

// V1AllocateResource is called when a resource requires allocation
func (rS *ResourceService) V1AllocateResource(args utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs := rS.cachedResourcesForEvent(args.UsageID)
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(args.Event); err != nil {
			return
		}
	}
	alcMsg, err := mtcRLs.AllocateResource(&ResourceUsage{ID: args.UsageID, Units: args.Units}, false)
	if err != nil {
		return err
	}
	// index it for matching out of cache
	rS.erMux.Lock()
	rS.eventResources[args.UsageID] = mtcRLs.ids()
	rS.erMux.Unlock()
	// index it for storing
	rS.srMux.Lock()
	for _, r := range mtcRLs {
		if rS.backupInterval == -1 {
			rS.StoreResource(r)
		} else if r.dirty != nil {
			*r.dirty = true // mark it to be saved
		}
		rS.storedResources[r.ID] = true
	}
	rS.srMux.Unlock()
	*reply = alcMsg
	return
}

// V1ReleaseResource is called when we need to clear an allocation
func (rS *ResourceService) V1ReleaseResource(args utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs := rS.cachedResourcesForEvent(args.UsageID)
	if mtcRLs == nil {
		if mtcRLs, err = rS.matchingResourcesForEvent(args.Event); err != nil {
			return
		}
	}
	mtcRLs.clearUsage(args.UsageID)
	rS.erMux.Lock()
	delete(rS.eventResources, args.UsageID)
	rS.erMux.Unlock()
	if rS.backupInterval != -1 {
		rS.srMux.Lock()
	}
	for _, r := range mtcRLs {
		if r.dirty != nil {
			if rS.backupInterval == -1 {
				rS.StoreResource(r)
			} else {
				*r.dirty = true // mark it to be saved
				rS.storedResources[r.ID] = true
			}
		}
	}
	if rS.backupInterval != -1 {
		rS.srMux.Unlock()
	}
	*reply = utils.OK
	return nil
}
