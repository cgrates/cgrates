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

type ResourceUsage struct {
	ID    string    // Unique identifier of this ResourceUsage, Eg: FreeSWITCH UUID
	Time  time.Time // So we can expire it later
	Units float64   // Number of units used
}

// ResourceLimit represents a limit imposed for accessing a resource (eg: new calls)
type ResourceLimit struct {
	sync.RWMutex
	ID                 string                    // Identifier of this limit
	Filters            []*RequestFilter          // Filters for the request
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	ExpiryTime         time.Time
	Weight             float64                   // Weight to sort the ResourceLimits
	Limit              float64                   // Limit value
	ActionTriggers     ActionTriggers            // Thresholds to check after changing Limit
	UsageTTL           time.Duration             // Expire usage after this duration
	AllocationMessage  string                    // message returned by the winning resourceLimit on allocation
	Usage              map[string]*ResourceUsage // Keep a record of usage, bounded with timestamps so we can expire too long records
	TotalUsage         float64                   // internal counter aggregating real usage of ResourceLimit
}

func (rl *ResourceLimit) removeExpiredUnits() {
	for ruID, rv := range rl.Usage {
		if time.Now().Sub(rv.Time) <= rl.UsageTTL {
			continue // not expired
		}
		delete(rl.Usage, ruID)
		rl.TotalUsage -= rv.Units
	}
}

func (rl *ResourceLimit) UsedUnits() float64 {
	if rl.UsageTTL != 0 {
		rl.removeExpiredUnits()
	}
	return rl.TotalUsage
}

func (rl *ResourceLimit) RecordUsage(ru *ResourceUsage) (err error) {
	if _, hasID := rl.Usage[ru.ID]; hasID {
		return fmt.Errorf("Duplicate resource usage with id: %s", ru.ID)
	}
	rl.Usage[ru.ID] = ru
	rl.TotalUsage += ru.Units
	return
}

func (rl *ResourceLimit) ClearUsage(ruID string) error {
	ru, hasIt := rl.Usage[ruID]
	if !hasIt {
		return fmt.Errorf("Cannot find usage record with id: %s", ruID)
	}
	delete(rl.Usage, ru.ID)
	rl.TotalUsage -= ru.Units
	return nil
}

// ResourceLimits is an ordered list of ResourceLimits based on Weight
type ResourceLimits []*ResourceLimit

// sort based on Weight
func (rls ResourceLimits) Sort() {
	sort.Slice(rls, func(i, j int) bool { return rls[i].Weight > rls[j].Weight })
}

// RecordUsage will record the usage in all the resource limits, failing back on errors
func (rls ResourceLimits) RecordUsage(ru *ResourceUsage) (err error) {
	var failedAtIdx int
	for i, rl := range rls {
		if err = rl.RecordUsage(ru); err != nil {
			failedAtIdx = i
			break
		}
	}
	if err != nil {
		for _, rl := range rls[:failedAtIdx] {
			rl.ClearUsage(ru.ID) // best effort
			rl.TotalUsage -= ru.Units
		}
	}
	return
}

// ClearUsage gives back the units to the pool
func (rls ResourceLimits) ClearUsage(ruID string) {
	for _, rl := range rls {
		rl.Lock()
		defer rl.Unlock()
	}
	for _, rl := range rls {
		if err := rl.ClearUsage(ruID); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<ResourceLimits>, err: %s", err.Error()))
		}
	}
	return
}

// AllocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rls ResourceLimits) AllocateResource(ru *ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rls) == 0 {
		return utils.META_NONE, nil
	}
	// lock resources so we can safely take decisions, need all to be locked before proceeding
	for _, rl := range rls {
		if dryRun {
			rl.RLock()
			defer rl.RUnlock()
		} else {
			rl.Lock()
			defer rl.Unlock()
		}
	}
	// Simulate resource usage
	for _, rl := range rls {
		if rl.Limit >= rl.UsedUnits()+ru.Units {
			if alcMessage == "" {
				alcMessage = rl.AllocationMessage
			}
			if alcMessage == "" { // rl.AllocationMessage is not populated
				alcMessage = rl.ID
			}
		}
	}
	if alcMessage == "" {
		return "", utils.ErrResourceUnavailable
	}
	if dryRun {
		return
	}
	err = rls.RecordUsage(ru)
	return
}

// Pas the config as a whole so we can ask access concurrently
func NewResourceLimiterService(cfg *config.CGRConfig, dataDB DataDB, cdrStatS rpcclient.RpcClientConnection) (*ResourceLimiterService, error) {
	if cdrStatS != nil && reflect.ValueOf(cdrStatS).IsNil() {
		cdrStatS = nil
	}
	return &ResourceLimiterService{dataDB: dataDB, cdrStatS: cdrStatS}, nil
}

// ResourcesLimiter is the service handling channel limits
type ResourceLimiterService struct {
	dataDB   DataDB // So we can load the data in cache and index it
	cdrStatS rpcclient.RpcClientConnection
}

// Called to start the service
func (rls *ResourceLimiterService) ListenAndServe() error {
	return nil
}

// Called to shutdown the service
func (rls *ResourceLimiterService) ServiceShutdown() error {
	return nil
}

// matchingResourceLimitsForEvent returns ordered list of matching resources which are active by the time of the call
func (rls *ResourceLimiterService) matchingResourceLimitsForEvent(ev map[string]interface{}) (resLimits ResourceLimits, err error) {
	matchingResources := make(map[string]*ResourceLimit)
	for fldName, fieldValIf := range ev {
		fldVal, canCast := utils.CastFieldIfToString(fieldValIf)
		if !canCast {
			return nil, fmt.Errorf("Cannot cast field: %s into string", fldName)
		}
		rlIDs, err := rls.dataDB.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey(fldName, fldVal))
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		for resName := range rlIDs {
			if _, hasIt := matchingResources[resName]; hasIt { // Already checked this RL
				continue
			}
			rl, err := rls.dataDB.GetResourceLimit(resName, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					continue
				}
				return nil, err
			}
			if rl.ActivationInterval != nil && !rl.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
				continue
			}
			passAllFilters := true
			for _, fltr := range rl.Filters {
				if pass, err := fltr.Pass(ev, "", rls.cdrStatS); err != nil {
					return nil, utils.NewErrServerError(err)
				} else if !pass {
					passAllFilters = false
					continue
				}
			}
			if passAllFilters {
				matchingResources[rl.ID] = rl // Cannot save it here since we could have errors after and resource will remain unused
			}
		}
	}
	// Check un-indexed resources
	uIdxRLIDs, err := rls.dataDB.MatchReqFilterIndex(utils.ResourceLimitsIndex, utils.ConcatenatedKey(utils.NOT_AVAILABLE, utils.NOT_AVAILABLE))
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}
	for resName := range uIdxRLIDs {
		if _, hasIt := matchingResources[resName]; hasIt { // Already checked this RL
			continue
		}
		rl, err := rls.dataDB.GetResourceLimit(resName, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if rl.ActivationInterval != nil && !rl.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		for _, fltr := range rl.Filters {
			if pass, err := fltr.Pass(ev, "", rls.cdrStatS); err != nil {
				return nil, utils.NewErrServerError(err)
			} else if !pass {
				continue
			}
			matchingResources[rl.ID] = rl // Cannot save it here since we could have errors after and resource will remain unused
		}
	}
	resLimits = make(ResourceLimits, len(matchingResources))
	i := 0
	for _, rl := range matchingResources {
		resLimits[i] = rl
		i++
	}
	resLimits.Sort()
	return resLimits, nil
}

// V1ResourceLimitsForEvent returns active resource limits matching the event
func (rls *ResourceLimiterService) V1ResourceLimitsForEvent(ev map[string]interface{}, reply *[]*ResourceLimit) error {
	matchingRLForEv, err := rls.matchingResourceLimitsForEvent(ev)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = matchingRLForEv
	return nil
}

func (rls *ResourceLimiterService) V1AllowUsage(args utils.AttrRLsResourceUsage, allow *bool) (err error) {
	mtcRLs, err := rls.matchingResourceLimitsForEvent(args.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, err = mtcRLs.AllocateResource(&ResourceUsage{ID: args.UsageID,
		Time: time.Now(), Units: args.Units}, true); err != nil {
		if err == utils.ErrResourceUnavailable {
			return // not error but still not allowed
		}
		return utils.NewErrServerError(err)
	}
	*allow = true
	return
}

// V1InitiateResourceUsage is called when a session or another event needs to consume
func (rls *ResourceLimiterService) V1AllocateResource(args utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs, err := rls.matchingResourceLimitsForEvent(args.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if alcMsg, err := mtcRLs.AllocateResource(&ResourceUsage{ID: args.UsageID,
		Time: time.Now(), Units: args.Units}, false); err != nil {
		return err
	} else {
		*reply = alcMsg
	}
	return
}

func (rls *ResourceLimiterService) V1ReleaseResource(attrs utils.AttrRLsResourceUsage, reply *string) (err error) {
	mtcRLs, err := rls.matchingResourceLimitsForEvent(attrs.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	mtcRLs.ClearUsage(attrs.UsageID)
	*reply = utils.OK
	return nil
}
