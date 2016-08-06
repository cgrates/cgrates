/*
Real-time Charging System for Telecom & ISP environments
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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type ResourceUsage struct {
	ID         string
	UsageTime  time.Time
	UsageUnits float64
}

// ResourceLimit represents a limit imposed for accessing a resource (eg: new calls)
type ResourceLimit struct {
	ID             string                    // Identifier of this limit
	Filters        []*RequestFilter          // Filters for the request
	ActivationTime time.Time                 // Time when this limit becomes active
	Weight         float64                   // Weight to sort the ResourceLimits
	Limit          float64                   // Limit value
	ActionTriggers ActionTriggers            // Thresholds to check after changing Limit
	UsageTTL       time.Duration             // Expire usage after this duration
	Usage          map[string]*ResourceUsage //Keep a record of usage, bounded with timestamps so we can expire too long records
	usageCounter   float64                   // internal counter representing real usage
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
	rls := &ResourceLimiterService{stringIndexes: make(map[string]map[string]utils.StringMap), dataDB: dataDB, cdrStatS: cdrStatS}
	return rls, nil
}

// ResourcesLimiter is the service handling channel limits
type ResourceLimiterService struct {
	sync.RWMutex
	stringIndexes map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue]utils.StringMap[resourceID]
	dataDB        AccountingStorage                     // So we can load the data in cache and index it
	cdrStatS      rpcclient.RpcClientConnection
}

// Index cached ResourceLimits with MetaString filter types
func (rls *ResourceLimiterService) indexStringFilters(rlIDs []string) error {
	utils.Logger.Info("<RLs> Start indexing string filters")
	newStringIndexes := make(map[string]map[string]utils.StringMap) // Index it transactionally
	var cacheIDsToIndex []string                                    // Cache keys of RLs to be indexed
	if rlIDs == nil {
		cacheIDsToIndex = CacheGetEntriesKeys(utils.ResourceLimitsPrefix)
	} else {
		for _, rlID := range rlIDs {
			cacheIDsToIndex = append(cacheIDsToIndex, utils.ResourceLimitsPrefix+rlID)
		}
	}
	for _, cacheKey := range cacheIDsToIndex {
		x, err := CacheGet(cacheKey)
		if err != nil {
			return err
		}
		rl := x.(*ResourceLimit)
		var hasMetaString bool
		for _, fltr := range rl.Filters {
			if fltr.Type != MetaString {
				continue
			}
			hasMetaString = true // Mark that we found at least one metatring so we don't need to index globally
			if _, hastIt := newStringIndexes[fltr.FieldName]; !hastIt {
				newStringIndexes[fltr.FieldName] = make(map[string]utils.StringMap)
			}
			for _, fldVal := range fltr.Values {
				if _, hasIt := newStringIndexes[fltr.FieldName][fldVal]; !hasIt {
					newStringIndexes[fltr.FieldName][fldVal] = make(utils.StringMap)
				}
				newStringIndexes[fltr.FieldName][fldVal][rl.ID] = true
			}
		}
		if !hasMetaString {
			if _, hasIt := newStringIndexes[utils.NOT_AVAILABLE]; !hasIt {
				newStringIndexes[utils.NOT_AVAILABLE] = make(map[string]utils.StringMap)
			}
			if _, hasIt := newStringIndexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE]; !hasIt {
				newStringIndexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = make(utils.StringMap)
			}
			newStringIndexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE][rl.ID] = true // Fields without real field index will be located in map[NOT_AVAILABLE][NOT_AVAILABLE][rl.ID]
		}
	}
	rls.Lock()
	defer rls.Unlock()
	if rlIDs == nil { // We have rebuilt complete index
		rls.stringIndexes = newStringIndexes
		return nil
	}
	// Merge the indexes since we have only performed limited indexing
	for fldNameKey, mpFldName := range newStringIndexes {
		if _, hasIt := rls.stringIndexes[fldNameKey]; !hasIt {
			rls.stringIndexes[fldNameKey] = mpFldName
		} else {
			for fldValKey, strMap := range newStringIndexes[fldNameKey] {
				if _, hasIt := rls.stringIndexes[fldNameKey][fldValKey]; !hasIt {
					rls.stringIndexes[fldNameKey][fldValKey] = strMap
				} else {
					for resIDKey := range newStringIndexes[fldNameKey][fldValKey] {
						rls.stringIndexes[fldNameKey][fldValKey][resIDKey] = true
					}
				}
			}
		}
	}
	utils.Logger.Info("<RLs> Done indexing string filters")
	return nil
}

// Called when cache/re-caching is necessary
func (rls *ResourceLimiterService) cacheResourceLimits(loadID string, rlIDs []string) error {
	if len(rlIDs) == 0 {
		return nil
	}
	if rlIDs == nil {
		utils.Logger.Info("<RLs> Start caching all resource limits")
	} else if len(rlIDs) != 0 {
		utils.Logger.Info(fmt.Sprintf("<RLs> Start caching resource limits with ids: %+v", rlIDs))
	}
	if err := rls.dataDB.CacheAccountingPrefixValues(loadID, map[string][]string{utils.ResourceLimitsPrefix: rlIDs}); err != nil {
		return err
	}
	utils.Logger.Info("<RLs> Done caching resource limits")
	return rls.indexStringFilters(rlIDs)
}

// Called to start the service
func (rls *ResourceLimiterService) Start() error {
	if err := rls.cacheResourceLimits("ResourceLimiterServiceStart", nil); err != nil {
		return err
	}
	return nil
}

// Called to shutdown the service
func (rls *ResourceLimiterService) Shutdown() error {
	return nil
}

// RPC Methods available internally

// Called when a session or another event needs to
func (rls *ResourceLimiterService) V1InitiateResourceUsage(attrs utils.AttrRLsResourceUsage, reply *string) error {
	rls.Lock() // Unknown number of RLs updated
	defer rls.Unlock()
	matchingResources := make(map[string]*ResourceLimit)
	for fldName, fieldValIf := range attrs.Event {
		_, hasIt := rls.stringIndexes[fldName]
		if !hasIt {
			continue
		}
		strVal, canCast := utils.CastFieldIfToString(fieldValIf)
		if !canCast {
			return fmt.Errorf("Cannot cast field: %s into string", fldName)
		}
		if _, hasIt := rls.stringIndexes[fldName][strVal]; !hasIt {
			continue
		}
		for resName := range rls.stringIndexes[fldName][strVal] {
			if _, hasIt := matchingResources[resName]; hasIt { // Already checked this RL
				continue
			}
			x, err := CacheGet(utils.ResourceLimitsPrefix + resName)
			if err != nil {
				return err
			}
			rl := x.(*ResourceLimit)
			for _, fltr := range rl.Filters {
				if pass, err := fltr.Pass(attrs, "", rls.cdrStatS); err != nil {
					return utils.NewErrServerError(err)
				} else if !pass {
					continue
				}
				if rl.Limit < rl.UsedUnits()+attrs.RequestedUnits {
					continue // Not offering any usage units
				}
				if err := rl.RecordUsage(&ResourceUsage{ID: attrs.ResourceUsageID, UsageTime: time.Now(), UsageUnits: attrs.RequestedUnits}); err != nil {
					return err // Should not happen
				}
				matchingResources[rl.ID] = rl // Cannot save it here since we could have errors after and resource will remain unused
			}
		}
	}
	if len(matchingResources) == 0 {
		return utils.ErrResourceUnavailable
	}
	for _, rl := range matchingResources {
		CacheSet(utils.ResourceLimitsPrefix+rl.ID, rl)
	}
	*reply = utils.OK
	return nil
}

// Cache/Re-cache
func (rls *ResourceLimiterService) V1CacheResourceLimits(attrs *utils.AttrRLsCache, reply *string) error {
	if err := rls.cacheResourceLimits(attrs.LoadID, attrs.ResourceLimitIDs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
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
