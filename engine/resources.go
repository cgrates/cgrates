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
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

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

// ResourceCfg represents the user configuration for the resource
type ResourceCfg struct {
	ID                 string                    // identifier of this resource
	Filters            []*RequestFilter          // filters for the request
	ActivationInterval *utils.ActivationInterval // time when this resource becomes active and expires
	UsageTTL           time.Duration             // expire the usage after this duration
	Limit              float64                   // limit value
	AllocationMessage  string                    // message returned by the winning resource on allocation
	Blocker            bool                      // blocker flag to stop processing on filters matched
	Stored             bool
	Weight             float64  // Weight to sort the resources
	Thresholds         []string // Thresholds to check after changing Limit
}

// StoredResourceUsages is stored on demand into dataDB
type StoredResourceUsages map[string]*ResourceUsage

func NewResource(rCfg *ResourceCfg) *Resource {
	return &Resource{rCfg: rCfg,
		usages: make(map[string]*ResourceUsage)}
}

// Resource represents a resource in the system
type Resource struct {
	sync.RWMutex
	rCfg      *ResourceCfg
	usages    map[string]*ResourceUsage
	ttlUsages []*ResourceUsage // used to speed up expirying of the usages
	tUsage    *float64         // sum of all usages
	dirty     bool             // the usages were modified, needs save
}

// removeExpiredUnits removes units which are expired
func (r *Resource) removeExpiredUnits() {
	var firstActive int
	r.RLock()
	for _, rv := range r.ttlUsages {
		if rv.isActive(time.Now()) {
			break
		}
		firstActive += 1
	}
	r.RUnlock()
	if firstActive == 0 {
		return
	}
	r.Lock()
	for _, ru := range r.ttlUsages[:firstActive] {
		delete(r.usages, ru.ID)
		*r.tUsage -= ru.Units
		ru = nil // empty it so we avoid memleak
	}
	r.ttlUsages = r.ttlUsages[firstActive:]
	r.Unlock()
}

// totalUsage returns the sum of all usage units
func (r *Resource) totalUsage() float64 {
	if r.tUsage == nil {
		var tu float64
		for _, ru := range r.usages {
			tu += ru.Units
		}
		r.tUsage = &tu
	}
	return *r.tUsage
}

// recordUsage records a new usage
func (r *Resource) recordUsage(ru *ResourceUsage) (err error) {
	if _, hasID := r.usages[ru.ID]; hasID {
		return fmt.Errorf("Duplicate resource usage with id: %s", ru.ID)
	}
	r.usages[ru.ID] = ru
	*r.tUsage += ru.Units
	return
}

// clearUsage clears the usage for an ID
func (r *Resource) clearUsage(ruID string) error {
	ru, hasIt := r.usages[ruID]
	if !hasIt {
		return fmt.Errorf("Cannot find usage record with id: %s", ruID)
	}
	delete(r.usages, ru.ID)
	*r.tUsage -= ru.Units
	return nil
}

// Resources is an ordered list of Resources based on Weight
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

// AllocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) AllocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return utils.META_NONE, nil
	}
	// lock resources so we can safely take decisions, need all to be locked before proceeding
	for _, r := range rs {
		if dryRun { // dryRun only needs read
			r.RLock()
			defer r.RUnlock()
		} else {
			r.Lock()
			defer r.Unlock()
		}
	}
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
func NewResourceService(cfg *config.CGRConfig, dataDB DataDB,
	statS rpcclient.RpcClientConnection) (*ResourceService, error) {
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	return &ResourceService{dataDB: dataDB, statS: statS}, nil
}

// ResourceService is the service handling channel limits
type ResourceService struct {
	dataDB DataDB // So we can load the data in cache and index it
	statS  rpcclient.RpcClientConnection
}

// Called to start the service
func (rS *ResourceService) ListenAndServe() error {
	return nil
}

// Called to shutdown the service
func (rS *ResourceService) ServiceShutdown() error {
	return nil
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceService) matchingResourcesForEvent(
	ev map[string]interface{}) (rs Resources, err error) {
	matchingResources := make(map[string]*Resource)
	rlIDs, err := matchingItemIDsForEvent(ev, rS.dataDB, utils.ResourcesIndex)
	if err != nil {
		return nil, err
	}
	for resName := range rlIDs {
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
				return nil, utils.NewErrServerError(err)
			} else if !pass {
				passAllFilters = false
				continue
			}
		}
		if !passAllFilters {
			continue
		}
		matchingResources[rCfg.ID] = NewResource(rCfg) // Cannot save it here since we could have errors after and resource will remain unused
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
		if r.rCfg.Blocker {
			rs = rs[:i+1]
		}
	}
	return
}

// V1ResourcesForEvent returns active resource limits matching the event
func (rS *ResourceService) V1ResourcesForEvent(ev map[string]interface{},
	reply *[]*ResourceCfg) error {
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
func (rS *ResourceService) V1AllowUsage(args utils.AttrRLsResourceUsage,
	allow *bool) (err error) {
	mtcRLs, err := rS.matchingResourcesForEvent(args.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, err = mtcRLs.AllocateResource(
		&ResourceUsage{ID: args.UsageID,
			Units: args.Units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			return // not error but still not allowed
		}
		return utils.NewErrServerError(err)
	}
	*allow = true
	return
}

// V1AllocateResource is called when a resource needs allocation
func (rS *ResourceService) V1AllocateResource(args utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs, err := rS.matchingResourcesForEvent(args.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if alcMsg, err := mtcRLs.AllocateResource(
		&ResourceUsage{ID: args.UsageID,
			Units: args.Units}, false); err != nil {
		return err
	} else {
		*reply = alcMsg
	}
	return
}

// V1ReleaseResource is called when we need to clear an allocation
func (rS *ResourceService) V1ReleaseResource(attrs utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs, err := rS.matchingResourcesForEvent(attrs.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	mtcRLs.clearUsage(attrs.UsageID)
	*reply = utils.OK
	return nil
}
