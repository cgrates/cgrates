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
	"github.com/cgrates/cgrates/utils"
)

//implement LoadReader interface
func (ms *MapStorage) GetTpIds(colName string) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (ms *MapStorage) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
	filters map[string]string, paginator *utils.PaginatorWithSearch) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (ms *MapStorage) GetTPTimings(tpid, id string) (timings []*utils.ApierTPTiming, err error) {
	key := utils.TBLTPTimings + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key = utils.TBLTPTimings + utils.ConcatenatedKey(tpid, id)
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.ApierTPTiming
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			timings = append(timings, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(timings) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPDestinations(tpid, id string) (dsts []*utils.TPDestination, err error) {
	key := utils.TBLTPDestinations + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPDestination
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			dsts = append(dsts, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(dsts) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPRates(tpid, id string) (rates []*utils.TPRate, err error) {
	key := utils.TBLTPRates + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPRate
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			for _, rs := range result.RateSlots {
				rs.SetDurations()
			}
			rates = append(rates, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(rates) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPDestinationRates(tpid, id string,
	paginator *utils.Paginator) (dRates []*utils.TPDestinationRate, err error) {
	key := utils.TBLTPDestinationRates + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPDestinationRate
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			dRates = append(dRates, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(dRates) == 0 {
		return nil, utils.ErrNotFound
	}
	// handle paginator
	return
}
func (ms *MapStorage) GetTPRatingPlans(string, string, *utils.Paginator) (rPlans []*utils.TPRatingPlan, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) (rProfiles []*utils.TPRatingProfile, err error) {
	return nil, utils.ErrNotImplemented
}

func (ms *MapStorage) GetTPSharedGroups(tpid, id string) (sGroups []*utils.TPSharedGroups, err error) {
	key := utils.TBLTPSharedGroups + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPSharedGroups
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			sGroups = append(sGroups, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(sGroups) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPActions(tpid, id string) (actions []*utils.TPActions, err error) {
	key := utils.TBLTPActions + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPActions
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			actions = append(actions, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(actions) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPActionPlans(tpid, id string) (aPlans []*utils.TPActionPlan, err error) {
	key := utils.TBLTPActionPlans + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPActionPlan
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			aPlans = append(aPlans, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(aPlans) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPActionTriggers(tpid, id string) (aTriggers []*utils.TPActionTriggers, err error) {
	key := utils.TBLTPActionTriggers + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tpid
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPActionTriggers
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			aTriggers = append(aTriggers, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(aTriggers) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}
func (ms *MapStorage) GetTPAccountActions(filter *utils.TPAccountActions) (accounts []*utils.TPAccountActions, err error) {
	return nil, utils.ErrNotImplemented
}

func (ms *MapStorage) GetTPResources(tpid, tenant, id string) (resources []*utils.TPResourceProfile, err error) {
	key := utils.TBLTPResources + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPResourceProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			resources = append(resources, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(resources) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPStats(tpid, tenant, id string) (stats []*utils.TPStatProfile, err error) {
	key := utils.TBLTPStats + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPStatProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			stats = append(stats, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(stats) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPThresholds(tpid, tenant, id string) (ths []*utils.TPThresholdProfile, err error) {
	key := utils.TBLTPThresholds + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPThresholdProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			ths = append(ths, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(ths) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPFilters(tpid, tenant, id string) (fltrs []*utils.TPFilterProfile, err error) {
	key := utils.TBLTPFilters + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPFilterProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			fltrs = append(fltrs, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(fltrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPSuppliers(tpid, tenant, id string) (supps []*utils.TPSupplierProfile, err error) {
	key := utils.TBLTPSuppliers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPSupplierProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			supps = append(supps, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(supps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPAttributes(tpid, tenant, id string) (attrs []*utils.TPAttributeProfile, err error) {
	key := utils.TBLTPAttributes + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPAttributeProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			attrs = append(attrs, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(attrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPChargers(tpid, tenant, id string) (cpps []*utils.TPChargerProfile, err error) {
	key := utils.TBLTPChargers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPChargerProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			cpps = append(cpps, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(cpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPDispatcherProfiles(tpid, tenant, id string) (dpps []*utils.TPDispatcherProfile, err error) {
	key := utils.TBLTPDispatchers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPDispatcherProfile
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			dpps = append(dpps, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(dpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (ms *MapStorage) GetTPDispatcherHosts(tpid, tenant, id string) (dpps []*utils.TPDispatcherHost, err error) {
	key := utils.TBLTPDispatcherHosts + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		if values, ok := ms.dict[id]; ok {
			var result *utils.TPDispatcherHost
			if err = ms.ms.Unmarshal(values, &result); err != nil {
				return nil, err
			}
			dpps = append(dpps, result)
		} else {
			return nil, utils.ErrNotFound
		}
	}
	if len(dpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

//implement LoadWriter interface
func (ms *MapStorage) RemTpData(table, tpid string, args map[string]string) (err error) {
	if table == utils.EmptyString {
		return ms.Flush(utils.EmptyString)
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := table + utils.CONCATENATED_KEY_SEP + tpid
	if args != nil {
		for _, val := range args {
			key += utils.CONCATENATED_KEY_SEP + val
		}
	}
	ids, _ := ms.GetKeysForPrefix(key)
	for _, id := range ids {
		delete(ms.dict, id)
	}
	return
}

func (ms *MapStorage) SetTPTimings(timings []*utils.ApierTPTiming) (err error) {
	if len(timings) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, timing := range timings {
		result, err := ms.ms.Marshal(timing)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPTimings, timing.TPid, timing.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPDestinations(dests []*utils.TPDestination) (err error) {
	if len(dests) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, destination := range dests {
		result, err := ms.ms.Marshal(destination)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPDestinations, destination.TPid, destination.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPRates(rates []*utils.TPRate) (err error) {
	if len(rates) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, rate := range rates {
		result, err := ms.ms.Marshal(rate)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPRates, rate.TPid, rate.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPDestinationRates(dRates []*utils.TPDestinationRate) (err error) {
	if len(dRates) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, dRate := range dRates {
		result, err := ms.ms.Marshal(dRate)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPDestinationRates, dRate.TPid, dRate.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPRatingPlans(ratingPlans []*utils.TPRatingPlan) (err error) {
	return utils.ErrNotImplemented
}

func (ms *MapStorage) SetTPRatingProfiles(ratingProfiles []*utils.TPRatingProfile) (err error) {
	return utils.ErrNotImplemented
}

func (ms *MapStorage) SetTPSharedGroups(groups []*utils.TPSharedGroups) (err error) {
	if len(groups) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, group := range groups {
		result, err := ms.ms.Marshal(group)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPSharedGroups, group.TPid, group.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPActions(acts []*utils.TPActions) (err error) {
	if len(acts) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, action := range acts {
		result, err := ms.ms.Marshal(action)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPActions, action.TPid, action.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPActionPlans(aPlans []*utils.TPActionPlan) (err error) {
	if len(aPlans) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, aPlan := range aPlans {
		result, err := ms.ms.Marshal(aPlan)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPActionPlans, aPlan.TPid, aPlan.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPActionTriggers(aTriggers []*utils.TPActionTriggers) (err error) {
	if len(aTriggers) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, aTrigger := range aTriggers {
		result, err := ms.ms.Marshal(aTrigger)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPActionTriggers, aTrigger.TPid, aTrigger.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPAccountActions(accActions []*utils.TPAccountActions) (err error) {
	return utils.ErrNotImplemented
}

func (ms *MapStorage) SetTPResources(resources []*utils.TPResourceProfile) (err error) {
	if len(resources) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, resource := range resources {
		result, err := ms.ms.Marshal(resource)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPResources, resource.TPid, resource.Tenant, resource.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPStats(stats []*utils.TPStatProfile) (err error) {
	if len(stats) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, stat := range stats {
		result, err := ms.ms.Marshal(stat)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPStats, stat.TPid, stat.Tenant, stat.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPThresholds(thresholds []*utils.TPThresholdProfile) (err error) {
	if len(thresholds) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, threshold := range thresholds {
		result, err := ms.ms.Marshal(threshold)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPThresholds, threshold.TPid, threshold.Tenant, threshold.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	if len(filters) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, filter := range filters {
		result, err := ms.ms.Marshal(filter)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPFilters, filter.TPid, filter.Tenant, filter.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPSuppliers(suppliers []*utils.TPSupplierProfile) (err error) {
	if len(suppliers) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, supplier := range suppliers {
		result, err := ms.ms.Marshal(supplier)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPSuppliers, supplier.TPid, supplier.Tenant, supplier.ID)] = result
	}
	return
}

func (ms *MapStorage) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	if len(attributes) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, attribute := range attributes {
		result, err := ms.ms.Marshal(attribute)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPAttributes, attribute.TPid, attribute.Tenant, attribute.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPChargers(cpps []*utils.TPChargerProfile) (err error) {
	if len(cpps) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, cpp := range cpps {
		result, err := ms.ms.Marshal(cpp)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPChargers, cpp.TPid, cpp.Tenant, cpp.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPDispatcherProfiles(dpps []*utils.TPDispatcherProfile) (err error) {
	if len(dpps) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, dpp := range dpps {
		result, err := ms.ms.Marshal(dpp)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPDispatchers, dpp.TPid, dpp.Tenant, dpp.ID)] = result
	}
	return
}
func (ms *MapStorage) SetTPDispatcherHosts(dpps []*utils.TPDispatcherHost) (err error) {
	if len(dpps) == 0 {
		return nil
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, dpp := range dpps {
		result, err := ms.ms.Marshal(dpp)
		if err != nil {
			return err
		}
		ms.dict[utils.ConcatenatedKey(utils.TBLTPDispatcherHosts, dpp.TPid, dpp.Tenant, dpp.ID)] = result
	}
	return
}

//implement CdrStorage interface
func (ms *MapStorage) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) RemoveSMCost(smc *SMCost) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) GetCDRs(filter *utils.CDRsFilter, remove bool) (cdrs []*CDR, count int64, err error) {
	return nil, 0, utils.ErrNotImplemented
}

func (ms *MapStorage) GetSMCosts(cgrid, runid, originHost, originIDPrfx string) (smCosts []*SMCost, err error) {
	return nil, utils.ErrNotImplemented
}
