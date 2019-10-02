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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

//implement LoadReader interface
func (iDB *InternalDB) GetTpIds(colName string) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
	filters map[string]string, paginator *utils.PaginatorWithSearch) (ids []string, err error) {
	key := table + utils.CONCATENATED_KEY_SEP + tpid
	fullIDs := iDB.db.GetItemIDs(table, key)
	switch table {
	// in case of account action we have the id the following form : loadid:tenant:account
	// so we need to treat it as a special case
	case utils.TBLTPAccountActions:
		for _, fullID := range fullIDs {
			var buildedID string
			sliceID := strings.Split(fullID[len(key)+1:], utils.CONCATENATED_KEY_SEP)
			for _, key := range distinct {
				switch key {
				case "loadid":
					if len(buildedID) == 0 {
						buildedID += sliceID[0]
					} else {
						buildedID += utils.CONCATENATED_KEY_SEP + sliceID[0]
					}
				case "tenant":
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[1]
				case "account":
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[2]

				}
			}
			ids = append(ids, buildedID)
		}
		// in case of rating profile we have the id in the following form : loadid:tenant:category:subject
	// so we need to treat it as a special case
	case utils.TBLTPRateProfiles:
		for _, fullID := range fullIDs {
			var buildedID string
			sliceID := strings.Split(fullID[len(key)+1:], utils.CONCATENATED_KEY_SEP)
			for _, key := range distinct {
				switch key {
				case "loadid":
					if len(buildedID) == 0 {
						buildedID += sliceID[0]
					} else {
						buildedID += utils.CONCATENATED_KEY_SEP + sliceID[0]
					}
				case "tenant":
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[1]
				case "category":
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[2]
				case "subject":
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[3]

				}
			}
			ids = append(ids, buildedID)
		}
	default:
		for _, fullID := range fullIDs {
			var buildedID string
			sliceID := strings.Split(fullID[len(key)+1:], utils.CONCATENATED_KEY_SEP)
			for i := 0; i < len(distinct); i++ {
				if len(buildedID) == 0 {
					buildedID += sliceID[len(sliceID)-i-1]
				} else {
					buildedID += utils.CONCATENATED_KEY_SEP + sliceID[len(sliceID)-i+1]
				}
			}
			ids = append(ids, buildedID)
		}
	}

	return
}

func (iDB *InternalDB) GetTPTimings(tpid, id string) (timings []*utils.ApierTPTiming, err error) {
	key := utils.TBLTPTimings + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}

	ids := iDB.db.GetItemIDs(utils.TBLTPTimings, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPTimings, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		timings = append(timings, x.(*utils.ApierTPTiming))
	}
	if len(timings) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPDestinations(tpid, id string) (dsts []*utils.TPDestination, err error) {
	key := utils.TBLTPDestinations + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPDestinations, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPDestinations, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		dsts = append(dsts, x.(*utils.TPDestination))
	}

	if len(dsts) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPRates(tpid, id string) (rates []*utils.TPRate, err error) {
	key := utils.TBLTPRates + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPRates, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPRates, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		result := x.(*utils.TPRate)
		for _, rs := range result.RateSlots {
			rs.SetDurations()
		}
		rates = append(rates, result)
	}

	if len(rates) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPDestinationRates(tpid, id string,
	paginator *utils.Paginator) (dRates []*utils.TPDestinationRate, err error) {
	key := utils.TBLTPDestinationRates + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPDestinationRates, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPDestinationRates, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}

		dRates = append(dRates, x.(*utils.TPDestinationRate))
	}

	if len(dRates) == 0 {
		return nil, utils.ErrNotFound
	}
	if paginator != nil {
		var limit, offset int
		if paginator.Limit != nil && *paginator.Limit > 0 {
			limit = *paginator.Limit
		}
		if paginator.Offset != nil && *paginator.Offset > 0 {
			offset = *paginator.Offset
		}
		if limit == 0 && offset == 0 {
			return dRates, nil
		}
		if offset > len(dRates) {
			return
		}
		if offset != 0 {
			limit = limit + offset
		}
		if limit == 0 {
			limit = len(dRates[offset:])
		} else if limit > len(dRates) {
			limit = len(dRates)
		}
		dRates = dRates[offset:limit]
	}
	return
}

func (iDB *InternalDB) GetTPRatingPlans(tpid, id string, paginator *utils.Paginator) (rPlans []*utils.TPRatingPlan, err error) {
	key := utils.TBLTPRatingPlans + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPRatingPlans, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPRatingPlans, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		rPlans = append(rPlans, x.(*utils.TPRatingPlan))
	}

	if len(rPlans) == 0 {
		return nil, utils.ErrNotFound
	}
	if paginator != nil {
		var limit, offset int
		if paginator.Limit != nil && *paginator.Limit > 0 {
			limit = *paginator.Limit
		}
		if paginator.Offset != nil && *paginator.Offset > 0 {
			offset = *paginator.Offset
		}
		if limit == 0 && offset == 0 {
			return rPlans, nil
		}
		if offset > len(rPlans) {
			return
		}
		if offset != 0 {
			limit = limit + offset
		}
		if limit == 0 {
			limit = len(rPlans[offset:])
		} else if limit > len(rPlans) {
			limit = len(rPlans)
		}
		rPlans = rPlans[offset:limit]
	}
	return
}

func (iDB *InternalDB) GetTPRatingProfiles(filter *utils.TPRatingProfile) (rProfiles []*utils.TPRatingProfile, err error) {
	key := utils.TBLTPRateProfiles + utils.CONCATENATED_KEY_SEP + filter.TPid

	if filter.LoadId != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.LoadId
	}
	if filter.Tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.Tenant
	}
	if filter.Category != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.Category
	}
	if filter.Subject != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.Subject
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPRateProfiles, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPRateProfiles, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		rProfiles = append(rProfiles, x.(*utils.TPRatingProfile))
	}

	if len(rProfiles) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPSharedGroups(tpid, id string) (sGroups []*utils.TPSharedGroups, err error) {
	key := utils.TBLTPSharedGroups + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPSharedGroups, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPSharedGroups, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		sGroups = append(sGroups, x.(*utils.TPSharedGroups))
	}

	if len(sGroups) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPActions(tpid, id string) (actions []*utils.TPActions, err error) {
	key := utils.TBLTPActions + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPActions, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPActions, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		actions = append(actions, x.(*utils.TPActions))

	}
	if len(actions) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPActionPlans(tpid, id string) (aPlans []*utils.TPActionPlan, err error) {
	key := utils.TBLTPActionPlans + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPActionPlans, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPActionPlans, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		aPlans = append(aPlans, x.(*utils.TPActionPlan))

	}
	if len(aPlans) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPActionTriggers(tpid, id string) (aTriggers []*utils.TPActionTriggers, err error) {
	key := utils.TBLTPActionTriggers + utils.CONCATENATED_KEY_SEP + tpid
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPActionTriggers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPActionTriggers, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		aTriggers = append(aTriggers, x.(*utils.TPActionTriggers))
	}
	if len(aTriggers) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}
func (iDB *InternalDB) GetTPAccountActions(filter *utils.TPAccountActions) (accounts []*utils.TPAccountActions, err error) {
	key := utils.TBLTPAccountActions + utils.CONCATENATED_KEY_SEP + filter.TPid

	if filter.LoadId != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.LoadId
	}
	if filter.Tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.Tenant
	}
	if filter.Account != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + filter.Account
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPAccountActions, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPAccountActions, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		accounts = append(accounts, x.(*utils.TPAccountActions))
	}

	if len(accounts) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPResources(tpid, tenant, id string) (resources []*utils.TPResourceProfile, err error) {
	key := utils.TBLTPResources + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPResources, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPResources, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		resources = append(resources, x.(*utils.TPResourceProfile))

	}
	if len(resources) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPStats(tpid, tenant, id string) (stats []*utils.TPStatProfile, err error) {
	key := utils.TBLTPStats + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPStats, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPStats, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		stats = append(stats, x.(*utils.TPStatProfile))

	}
	if len(stats) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPThresholds(tpid, tenant, id string) (ths []*utils.TPThresholdProfile, err error) {
	key := utils.TBLTPThresholds + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPThresholds, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPThresholds, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		ths = append(ths, x.(*utils.TPThresholdProfile))

	}
	if len(ths) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPFilters(tpid, tenant, id string) (fltrs []*utils.TPFilterProfile, err error) {
	key := utils.TBLTPFilters + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPFilters, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPFilters, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		fltrs = append(fltrs, x.(*utils.TPFilterProfile))

	}
	if len(fltrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPSuppliers(tpid, tenant, id string) (supps []*utils.TPSupplierProfile, err error) {
	key := utils.TBLTPSuppliers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPSuppliers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPSuppliers, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		supps = append(supps, x.(*utils.TPSupplierProfile))

	}
	if len(supps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPAttributes(tpid, tenant, id string) (attrs []*utils.TPAttributeProfile, err error) {
	key := utils.TBLTPAttributes + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPAttributes, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPAttributes, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		attrs = append(attrs, x.(*utils.TPAttributeProfile))

	}
	if len(attrs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPChargers(tpid, tenant, id string) (cpps []*utils.TPChargerProfile, err error) {
	key := utils.TBLTPChargers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPChargers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPChargers, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		cpps = append(cpps, x.(*utils.TPChargerProfile))

	}
	if len(cpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPDispatcherProfiles(tpid, tenant, id string) (dpps []*utils.TPDispatcherProfile, err error) {
	key := utils.TBLTPDispatchers + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPDispatchers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPDispatchers, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		dpps = append(dpps, x.(*utils.TPDispatcherProfile))

	}
	if len(dpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPDispatcherHosts(tpid, tenant, id string) (dpps []*utils.TPDispatcherHost, err error) {
	key := utils.TBLTPDispatcherHosts + utils.CONCATENATED_KEY_SEP + tpid
	if tenant != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + tenant
	}
	if id != utils.EmptyString {
		key += utils.CONCATENATED_KEY_SEP + id
	}
	ids := iDB.db.GetItemIDs(utils.TBLTPDispatcherHosts, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.TBLTPDispatcherHosts, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		dpps = append(dpps, x.(*utils.TPDispatcherHost))

	}
	if len(dpps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

//implement LoadWriter interface
func (iDB *InternalDB) RemTpData(table, tpid string, args map[string]string) (err error) {
	if table == utils.EmptyString {
		return iDB.Flush(utils.EmptyString)
	}
	key := table + utils.CONCATENATED_KEY_SEP + tpid
	if args != nil {
		for _, val := range args {
			key += utils.CONCATENATED_KEY_SEP + val
		}
	}
	ids := iDB.db.GetItemIDs(table, key)
	for _, id := range ids {
		iDB.db.Remove(table, id,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPTimings(timings []*utils.ApierTPTiming) (err error) {
	if len(timings) == 0 {
		return nil
	}
	for _, timing := range timings {
		iDB.db.Set(utils.TBLTPTimings, utils.ConcatenatedKey(utils.TBLTPTimings, timing.TPid, timing.ID), timing, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDestinations(dests []*utils.TPDestination) (err error) {
	if len(dests) == 0 {
		return nil
	}
	for _, destination := range dests {
		iDB.db.Set(utils.TBLTPDestinations, utils.ConcatenatedKey(utils.TBLTPDestinations, destination.TPid, destination.ID), destination, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRates(rates []*utils.TPRate) (err error) {
	if len(rates) == 0 {
		return nil
	}
	for _, rate := range rates {
		iDB.db.Set(utils.TBLTPRates, utils.ConcatenatedKey(utils.TBLTPRates, rate.TPid, rate.ID), rate, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPDestinationRates(dRates []*utils.TPDestinationRate) (err error) {
	if len(dRates) == 0 {
		return nil
	}
	for _, dRate := range dRates {
		iDB.db.Set(utils.TBLTPDestinationRates, utils.ConcatenatedKey(utils.TBLTPDestinationRates, dRate.TPid, dRate.ID), dRate, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRatingPlans(ratingPlans []*utils.TPRatingPlan) (err error) {
	if len(ratingPlans) == 0 {
		return nil
	}
	for _, rPlan := range ratingPlans {
		iDB.db.Set(utils.TBLTPRatingPlans, utils.ConcatenatedKey(utils.TBLTPRatingPlans, rPlan.TPid, rPlan.ID), rPlan, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRatingProfiles(ratingProfiles []*utils.TPRatingProfile) (err error) {
	if len(ratingProfiles) == 0 {
		return nil
	}
	for _, rProfile := range ratingProfiles {
		iDB.db.Set(utils.TBLTPRateProfiles, utils.ConcatenatedKey(utils.TBLTPRateProfiles, rProfile.TPid,
			rProfile.LoadId, rProfile.Tenant, rProfile.Category, rProfile.Subject), rProfile, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPSharedGroups(groups []*utils.TPSharedGroups) (err error) {
	if len(groups) == 0 {
		return nil
	}
	for _, group := range groups {
		iDB.db.Set(utils.TBLTPSharedGroups, utils.ConcatenatedKey(utils.TBLTPSharedGroups, group.TPid, group.ID), group, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActions(acts []*utils.TPActions) (err error) {
	if len(acts) == 0 {
		return nil
	}
	for _, action := range acts {
		iDB.db.Set(utils.TBLTPActions, utils.ConcatenatedKey(utils.TBLTPActions, action.TPid, action.ID), action, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActionPlans(aPlans []*utils.TPActionPlan) (err error) {
	if len(aPlans) == 0 {
		return nil
	}
	for _, aPlan := range aPlans {
		iDB.db.Set(utils.TBLTPActionPlans, utils.ConcatenatedKey(utils.TBLTPActionPlans, aPlan.TPid, aPlan.ID), aPlan, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActionTriggers(aTriggers []*utils.TPActionTriggers) (err error) {
	if len(aTriggers) == 0 {
		return nil
	}
	for _, aTrigger := range aTriggers {
		iDB.db.Set(utils.TBLTPActionTriggers, utils.ConcatenatedKey(utils.TBLTPActionTriggers, aTrigger.TPid, aTrigger.ID), aTrigger, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAccountActions(accActions []*utils.TPAccountActions) (err error) {
	if len(accActions) == 0 {
		return nil
	}
	for _, accAction := range accActions {
		iDB.db.Set(utils.TBLTPAccountActions, utils.ConcatenatedKey(utils.TBLTPAccountActions, accAction.TPid,
			accAction.LoadId, accAction.Tenant, accAction.Account), accAction, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPResources(resources []*utils.TPResourceProfile) (err error) {
	if len(resources) == 0 {
		return nil
	}
	for _, resource := range resources {
		iDB.db.Set(utils.TBLTPResources, utils.ConcatenatedKey(utils.TBLTPResources, resource.TPid, resource.Tenant, resource.ID), resource, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPStats(stats []*utils.TPStatProfile) (err error) {
	if len(stats) == 0 {
		return nil
	}
	for _, stat := range stats {
		iDB.db.Set(utils.TBLTPStats, utils.ConcatenatedKey(utils.TBLTPStats, stat.TPid, stat.Tenant, stat.ID), stat, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPThresholds(thresholds []*utils.TPThresholdProfile) (err error) {
	if len(thresholds) == 0 {
		return nil
	}

	for _, threshold := range thresholds {
		iDB.db.Set(utils.TBLTPThresholds, utils.ConcatenatedKey(utils.TBLTPThresholds, threshold.TPid, threshold.Tenant, threshold.ID), threshold, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	if len(filters) == 0 {
		return nil
	}

	for _, filter := range filters {
		iDB.db.Set(utils.TBLTPFilters, utils.ConcatenatedKey(utils.TBLTPFilters, filter.TPid, filter.Tenant, filter.ID), filter, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPSuppliers(suppliers []*utils.TPSupplierProfile) (err error) {
	if len(suppliers) == 0 {
		return nil
	}
	for _, supplier := range suppliers {
		iDB.db.Set(utils.TBLTPSuppliers, utils.ConcatenatedKey(utils.TBLTPSuppliers, supplier.TPid, supplier.Tenant, supplier.ID), supplier, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	if len(attributes) == 0 {
		return nil
	}

	for _, attribute := range attributes {
		iDB.db.Set(utils.TBLTPAttributes, utils.ConcatenatedKey(utils.TBLTPAttributes, attribute.TPid, attribute.Tenant, attribute.ID), attribute, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPChargers(cpps []*utils.TPChargerProfile) (err error) {
	if len(cpps) == 0 {
		return nil
	}

	for _, cpp := range cpps {
		iDB.db.Set(utils.TBLTPChargers, utils.ConcatenatedKey(utils.TBLTPChargers, cpp.TPid, cpp.Tenant, cpp.ID), cpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherProfiles(dpps []*utils.TPDispatcherProfile) (err error) {
	if len(dpps) == 0 {
		return nil
	}

	for _, dpp := range dpps {
		iDB.db.Set(utils.TBLTPDispatchers, utils.ConcatenatedKey(utils.TBLTPDispatchers, dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherHosts(dpps []*utils.TPDispatcherHost) (err error) {
	if len(dpps) == 0 {
		return nil
	}
	for _, dpp := range dpps {
		iDB.db.Set(utils.TBLTPDispatcherHosts, utils.ConcatenatedKey(utils.TBLTPDispatcherHosts, dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

//implement CdrStorage interface
func (iDB *InternalDB) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	return utils.ErrNotImplemented
}
func (iDB *InternalDB) RemoveSMCost(smc *SMCost) (err error) {
	return utils.ErrNotImplemented
}
func (iDB *InternalDB) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	return utils.ErrNotImplemented
}
func (iDB *InternalDB) GetCDRs(filter *utils.CDRsFilter, remove bool) (cdrs []*CDR, count int64, err error) {
	return nil, 0, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetSMCosts(cgrid, runid, originHost, originIDPrfx string) (smCosts []*SMCost, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetSMCost(smCost *SMCost) error {
	if smCost.CostDetails == nil {
		return nil
	}
	result, err := iDB.ms.Marshal(smCost)
	if err != nil {
		return err
	}
	iDB.db.Set(utils.SessionCostsTBL, utils.LOG_CALL_COST_PREFIX+smCost.CostSource+smCost.RunID+"_"+smCost.CGRID, result, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return err
}
