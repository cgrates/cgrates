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
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRates(tpid, id string) (rates []*utils.TPRate, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDestinationRates(tpid, id string,
	paginator *utils.Paginator) (dRates []*utils.TPDestinationRate, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRatingPlans(string, string, *utils.Paginator) (rPlans []*utils.TPRatingPlan, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPRatingProfiles(filter *utils.TPRatingProfile) (rProfiles []*utils.TPRatingProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPSharedGroups(tpid, id string) (sGroups []*utils.TPSharedGroups, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActions(tpid, id string) (actions []*utils.TPActions, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActionPlans(tpid, id string) (aPlans []*utils.TPActionPlan, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPActionTriggers(tpid, id string) (aTriggers []*utils.TPActionTriggers, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPAccountActions(filter *utils.TPAccountActions) (accounts []*utils.TPAccountActions, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPResources(tpid, tenant, id string) (resources []*utils.TPResourceProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPStats(tpid, tenant, id string) (stats []*utils.TPStatProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPThresholds(tpid, tenant, id string) (ths []*utils.TPThresholdProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPFilters(tpid, tenant, id string) (fltrs []*utils.TPFilterProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPSuppliers(tpid, tenant, id string) (supps []*utils.TPSupplierProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPAttributes(tpid, tenant, id string) (attrs []*utils.TPAttributeProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPChargers(tpid, tenant, id string) (attrs []*utils.TPChargerProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDispatcherProfiles(tpid, tenant, id string) (attrs []*utils.TPDispatcherProfile, err error) {
	return nil, utils.ErrNotImplemented
}
func (ms *MapStorage) GetTPDispatcherHosts(tpid, tenant, id string) (attrs []*utils.TPDispatcherHost, err error) {
	return nil, utils.ErrNotImplemented
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
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRates(rates []*utils.TPRate) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDestinationRates(dRates []*utils.TPDestinationRate) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRatingPlans(ratingPlans []*utils.TPRatingPlan) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPRatingProfiles(ratingProfiles []*utils.TPRatingProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPSharedGroups(groups []*utils.TPSharedGroups) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActions(acts []*utils.TPActions) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActionPlans(aPlans []*utils.TPActionPlan) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPActionTriggers(aTriggers []*utils.TPActionTriggers) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPAccountActions(accActions []*utils.TPAccountActions) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPResources(resources []*utils.TPResourceProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPStats(stats []*utils.TPStatProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPThresholds(thresholds []*utils.TPThresholdProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPSuppliers(suppliers []*utils.TPSupplierProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPChargers(cpps []*utils.TPChargerProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDispatcherProfiles(dpps []*utils.TPDispatcherProfile) (err error) {
	return utils.ErrNotImplemented
}
func (ms *MapStorage) SetTPDispatcherHosts(dpps []*utils.TPDispatcherHost) (err error) {
	return utils.ErrNotImplemented
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
