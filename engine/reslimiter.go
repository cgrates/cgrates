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
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type ResourceUsage struct {
	ID         string    // Unique identifier of this resourceUsage, Eg: FreeSWITCH UUID
	UsageTime  time.Time // So we can expire it later
	UsageUnits float64   // Number of units used
}

// ResourceLimit represents a limit imposed for accessing a resource (eg: new calls)
type ResourceLimit struct {
	ID             string           // Identifier of this limit
	Filters        []*RequestFilter // Filters for the request
	ActivationTime time.Time        // Time when this limit becomes active
	ExpiryTime     time.Time
	Weight         float64                   // Weight to sort the ResourceLimits
	Limit          float64                   // Limit value
	ActionTriggers ActionTriggers            // Thresholds to check after changing Limit
	UsageTTL       time.Duration             // Expire usage after this duration
	Usage          map[string]*ResourceUsage // Keep a record of usage, bounded with timestamps so we can expire too long records
	usageCounter   float64                   // internal counter aggregating real usage of ResourceLimit
}

func (rl *ResourceLimit) removeExpiredUnits() {
	for ruID, rv := range rl.Usage {
		if time.Now().Sub(rv.UsageTime) <= rl.UsageTTL {
			continue // not expired
		}
		delete(rl.Usage, ruID)
		rl.usageCounter -= rv.UsageUnits
	}
}

func (rl *ResourceLimit) UsedUnits() float64 {
	if rl.UsageTTL != 0 {
		rl.removeExpiredUnits()
	}
	return rl.usageCounter
}

func (rl *ResourceLimit) RecordUsage(ru *ResourceUsage) error {
	if _, hasID := rl.Usage[ru.ID]; hasID {
		return fmt.Errorf("Duplicate resource usage with id: %s", ru.ID)
	}
	rl.Usage[ru.ID] = ru
	rl.usageCounter += ru.UsageUnits
	return nil
}

func (rl *ResourceLimit) RemoveUsage(ruID string) error {
	ru, hasIt := rl.Usage[ruID]
	if !hasIt {
		return fmt.Errorf("Cannot find usage record with id: %s", ruID)
	}
	delete(rl.Usage, ru.ID)
	rl.usageCounter -= ru.UsageUnits
	return nil
}

// Pas the config as a whole so we can ask access concurrently
func NewResourceLimiterService(cfg *config.CGRConfig, dataDB AccountingStorage, cdrStatS rpcclient.RpcClientConnection) (*ResourceLimiterService, error) {
	if cdrStatS != nil && reflect.ValueOf(cdrStatS).IsNil() {
		cdrStatS = nil
	}
	rls := &ResourceLimiterService{dataDB: dataDB, cdrStatS: cdrStatS}
	return rls, nil
}

// ResourcesLimiter is the service handling channel limits
type ResourceLimiterService struct {
	sync.RWMutex
	dataDB   AccountingStorage // So we can load the data in cache and index it
	cdrStatS rpcclient.RpcClientConnection
}

func (rls *ResourceLimiterService) matchingResourceLimitsForEvent(ev map[string]interface{}) (map[string]*ResourceLimit, error) {
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
			now := time.Now()
			if rl.ActivationTime.After(now) || (!rl.ExpiryTime.IsZero() && rl.ExpiryTime.Before(now)) { // not active
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
		now := time.Now()
		if rl.ActivationTime.After(now) || (!rl.ExpiryTime.IsZero() && rl.ExpiryTime.Before(now)) { // not active
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
	return matchingResources, nil
}

// Called to start the service
func (rls *ResourceLimiterService) ListenAndServe() error {
	return nil
}

// Called to shutdown the service
func (rls *ResourceLimiterService) ServiceShutdown() error {
	return nil
}

// RPC Methods

func (rls *ResourceLimiterService) V1ResourceLimitsForEvent(ev map[string]interface{}, reply *[]*ResourceLimit) error {
	rls.Lock() // Unknown number of RLs updated
	defer rls.Unlock()
	matchingRLForEv, err := rls.matchingResourceLimitsForEvent(ev)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	retRLs := make([]*ResourceLimit, len(matchingRLForEv))
	i := 0
	for _, rl := range matchingRLForEv {
		retRLs[i] = rl
		i++
	}
	*reply = retRLs
	return nil
}

// Alias API for external use
func (rls *ResourceLimiterService) ResourceLimitsForEvent(ev map[string]interface{}, reply *[]*ResourceLimit) error {
	return rls.V1ResourceLimitsForEvent(ev, reply)
}

// Called when a session or another event needs to
func (rls *ResourceLimiterService) V1InitiateResourceUsage(attrs utils.AttrRLsResourceUsage, reply *string) error {
	rls.Lock() // Unknown number of RLs updated
	defer rls.Unlock()
	matchingRLForEv, err := rls.matchingResourceLimitsForEvent(attrs.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	for rlID, rl := range matchingRLForEv {
		if rl.Limit < rl.UsedUnits()+attrs.RequestedUnits {
			delete(matchingRLForEv, rlID)
		}
		if err := rl.RecordUsage(&ResourceUsage{ID: attrs.ResourceUsageID, UsageTime: time.Now(), UsageUnits: attrs.RequestedUnits}); err != nil {
			return err // Should not happen
		}
	}
	if len(matchingRLForEv) == 0 {
		return utils.ErrResourceUnavailable
	}
	for _, rl := range matchingRLForEv {
		cache.Set(utils.ResourceLimitsPrefix+rl.ID, rl, true, "") // no real reason for a transaction
	}
	*reply = utils.OK
	return nil
}

// Alias for externam methods
func (rls *ResourceLimiterService) InitiateResourceUsage(attrs utils.AttrRLsResourceUsage, reply *string) error {
	return rls.V1InitiateResourceUsage(attrs, reply)
}

func (rls *ResourceLimiterService) V1TerminateResourceUsage(attrs utils.AttrRLsResourceUsage, reply *string) error {
	rls.Lock() // Unknown number of RLs updated
	defer rls.Unlock()
	matchingRLForEv, err := rls.matchingResourceLimitsForEvent(attrs.Event)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	for _, rl := range matchingRLForEv {
		rl.RemoveUsage(attrs.ResourceUsageID)
	}
	*reply = utils.OK
	return nil
}

// Alias for external methods
func (rls *ResourceLimiterService) TerminateResourceUsage(attrs utils.AttrRLsResourceUsage, reply *string) error {
	return rls.V1TerminateResourceUsage(attrs, reply)
}

// Make the service available as RPC internally
func (rls *ResourceLimiterService) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(rls).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
