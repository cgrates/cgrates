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
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

//implement LoadReader interface
func (iDB *InternalDB) GetTpIds(colName string) (ids []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetTpTableIds(tpid, table string, distinct utils.TPDistinctIds,
	filters map[string]string, paginator *utils.PaginatorWithSearch) (ids []string, err error) {
	fullIDs := iDB.db.GetItemIDs(table, tpid)
	idSet := utils.NewStringSet(nil)
	for _, fullID := range fullIDs {
		switch table {
		// in case of account action <loadid:tenant:account> and rating profile <loadid:tenant:category:subject>
		// the retutned value may be only the loadID
		case utils.TBLTPAccountActions, utils.TBLTPRateProfiles:
			if len(distinct) == 1 { // special case when to return only the loadID
				sliceID := strings.Split(fullID[len(tpid)+1:], utils.CONCATENATED_KEY_SEP)
				idSet.Add(sliceID[0])
				continue
			}
		// in the rest of the cases we return all ID every time
		default:
		}
		idSet.Add(fullID[len(tpid)+1:])
	}
	ids = idSet.AsSlice()
	return
}

func (iDB *InternalDB) GetTPTimings(tpid, id string) (timings []*utils.ApierTPTiming, err error) {
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := filter.TPid

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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := filter.TPid

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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
	key := tpid
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
		iDB.db.Set(utils.TBLTPTimings, utils.ConcatenatedKey(timing.TPid, timing.ID), timing, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDestinations(dests []*utils.TPDestination) (err error) {
	if len(dests) == 0 {
		return nil
	}
	for _, destination := range dests {
		iDB.db.Set(utils.TBLTPDestinations, utils.ConcatenatedKey(destination.TPid, destination.ID), destination, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRates(rates []*utils.TPRate) (err error) {
	if len(rates) == 0 {
		return nil
	}
	for _, rate := range rates {
		iDB.db.Set(utils.TBLTPRates, utils.ConcatenatedKey(rate.TPid, rate.ID), rate, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPDestinationRates(dRates []*utils.TPDestinationRate) (err error) {
	if len(dRates) == 0 {
		return nil
	}
	for _, dRate := range dRates {
		iDB.db.Set(utils.TBLTPDestinationRates, utils.ConcatenatedKey(dRate.TPid, dRate.ID), dRate, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRatingPlans(ratingPlans []*utils.TPRatingPlan) (err error) {
	if len(ratingPlans) == 0 {
		return nil
	}
	for _, rPlan := range ratingPlans {
		iDB.db.Set(utils.TBLTPRatingPlans, utils.ConcatenatedKey(rPlan.TPid, rPlan.ID), rPlan, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRatingProfiles(ratingProfiles []*utils.TPRatingProfile) (err error) {
	if len(ratingProfiles) == 0 {
		return nil
	}
	for _, rProfile := range ratingProfiles {
		iDB.db.Set(utils.TBLTPRateProfiles, utils.ConcatenatedKey(rProfile.TPid,
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
		iDB.db.Set(utils.TBLTPSharedGroups, utils.ConcatenatedKey(group.TPid, group.ID), group, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActions(acts []*utils.TPActions) (err error) {
	if len(acts) == 0 {
		return nil
	}
	for _, action := range acts {
		iDB.db.Set(utils.TBLTPActions, utils.ConcatenatedKey(action.TPid, action.ID), action, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActionPlans(aPlans []*utils.TPActionPlan) (err error) {
	if len(aPlans) == 0 {
		return nil
	}
	for _, aPlan := range aPlans {
		iDB.db.Set(utils.TBLTPActionPlans, utils.ConcatenatedKey(aPlan.TPid, aPlan.ID), aPlan, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActionTriggers(aTriggers []*utils.TPActionTriggers) (err error) {
	if len(aTriggers) == 0 {
		return nil
	}
	for _, aTrigger := range aTriggers {
		iDB.db.Set(utils.TBLTPActionTriggers, utils.ConcatenatedKey(aTrigger.TPid, aTrigger.ID), aTrigger, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAccountActions(accActions []*utils.TPAccountActions) (err error) {
	if len(accActions) == 0 {
		return nil
	}
	for _, accAction := range accActions {
		iDB.db.Set(utils.TBLTPAccountActions, utils.ConcatenatedKey(accAction.TPid,
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
		iDB.db.Set(utils.TBLTPResources, utils.ConcatenatedKey(resource.TPid, resource.Tenant, resource.ID), resource, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPStats(stats []*utils.TPStatProfile) (err error) {
	if len(stats) == 0 {
		return nil
	}
	for _, stat := range stats {
		iDB.db.Set(utils.TBLTPStats, utils.ConcatenatedKey(stat.TPid, stat.Tenant, stat.ID), stat, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPThresholds(thresholds []*utils.TPThresholdProfile) (err error) {
	if len(thresholds) == 0 {
		return nil
	}

	for _, threshold := range thresholds {
		iDB.db.Set(utils.TBLTPThresholds, utils.ConcatenatedKey(threshold.TPid, threshold.Tenant, threshold.ID), threshold, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	if len(filters) == 0 {
		return nil
	}

	for _, filter := range filters {
		iDB.db.Set(utils.TBLTPFilters, utils.ConcatenatedKey(filter.TPid, filter.Tenant, filter.ID), filter, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPSuppliers(suppliers []*utils.TPSupplierProfile) (err error) {
	if len(suppliers) == 0 {
		return nil
	}
	for _, supplier := range suppliers {
		iDB.db.Set(utils.TBLTPSuppliers, utils.ConcatenatedKey(supplier.TPid, supplier.Tenant, supplier.ID), supplier, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	if len(attributes) == 0 {
		return nil
	}

	for _, attribute := range attributes {
		iDB.db.Set(utils.TBLTPAttributes, utils.ConcatenatedKey(attribute.TPid, attribute.Tenant, attribute.ID), attribute, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPChargers(cpps []*utils.TPChargerProfile) (err error) {
	if len(cpps) == 0 {
		return nil
	}

	for _, cpp := range cpps {
		iDB.db.Set(utils.TBLTPChargers, utils.ConcatenatedKey(cpp.TPid, cpp.Tenant, cpp.ID), cpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherProfiles(dpps []*utils.TPDispatcherProfile) (err error) {
	if len(dpps) == 0 {
		return nil
	}

	for _, dpp := range dpps {
		iDB.db.Set(utils.TBLTPDispatchers, utils.ConcatenatedKey(dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherHosts(dpps []*utils.TPDispatcherHost) (err error) {
	if len(dpps) == 0 {
		return nil
	}
	for _, dpp := range dpps {
		iDB.db.Set(utils.TBLTPDispatcherHosts, utils.ConcatenatedKey(dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

//implement CdrStorage interface
func (iDB *InternalDB) SetCDR(cdr *CDR, allowUpdate bool) (err error) {
	if cdr.OrderID == 0 {
		cdr.OrderID = iDB.cnter.Next()
	}
	cdrKey := utils.ConcatenatedKey(cdr.CGRID, cdr.RunID, cdr.OriginID)
	if !allowUpdate {
		if _, has := iDB.db.Get(utils.CDRsTBL, cdrKey); has {
			return utils.ErrExists
		}
	}
	idxs := utils.NewStringSet(nil)
	iDB.indexedFieldsMutex.RLock()
	if len(iDB.stringIndexedFields) == 0 && len(iDB.prefixIndexedFields) == 0 { // add default indexes
		idxs.Add(utils.ConcatenatedKey(utils.CGRID, cdr.CGRID))
		idxs.Add(utils.ConcatenatedKey(utils.RunID, cdr.RunID))
		idxs.Add(utils.ConcatenatedKey(utils.OriginHost, cdr.OriginHost))
		idxs.Add(utils.ConcatenatedKey(utils.Source, cdr.Source))
		idxs.Add(utils.ConcatenatedKey(utils.OriginID, cdr.OriginID))
		idxs.Add(utils.ConcatenatedKey(utils.ToR, cdr.ToR))
		idxs.Add(utils.ConcatenatedKey(utils.RequestType, cdr.RequestType))
		idxs.Add(utils.ConcatenatedKey(utils.Tenant, cdr.Tenant))
		idxs.Add(utils.ConcatenatedKey(utils.Category, cdr.Category))
		idxs.Add(utils.ConcatenatedKey(utils.Account, cdr.Account))
		idxs.Add(utils.ConcatenatedKey(utils.Subject, cdr.Subject))
		idxs.Add(utils.ConcatenatedKey(utils.Destination, cdr.Destination)) // include the whole Destination
		for i := len(cdr.Destination) - 1; i > 0; i-- {                     // add destination as prefix
			idxs.Add(utils.ConcatenatedKey(utils.Destination, cdr.Destination[:i]))
		}
	} else { // add user indexes
		mpCDR := cdr.AsMapStringIface()
		for _, cdrIdx := range iDB.stringIndexedFields {
			idxs.Add(utils.ConcatenatedKey(cdrIdx, utils.IfaceAsString(mpCDR[cdrIdx])))
		}
		for _, cdrIdx := range iDB.prefixIndexedFields {
			strVal := utils.IfaceAsString(mpCDR[cdrIdx])
			idxs.Add(utils.ConcatenatedKey(cdrIdx, strVal))
			for i := len(strVal) - 1; i > 0; i-- {
				idxs.Add(utils.ConcatenatedKey(cdrIdx, strVal[:i]))
			}
		}
	}
	iDB.indexedFieldsMutex.RUnlock()

	iDB.db.Set(utils.CDRsTBL, cdrKey, cdr, idxs.AsSlice(),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)

	return
}

func (iDB *InternalDB) RemoveSMCost(smc *SMCost) (err error) {
	iDB.db.Remove(utils.SessionCostsTBL, utils.ConcatenatedKey(smc.CGRID, smc.RunID, smc.OriginHost, smc.OriginID),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveSMCosts(qryFltr *utils.SMCostFilter) error {
	var smMpIDs utils.StringMap
	// Apply string filter
	for _, fltrSlc := range []struct {
		key string
		ids []string
	}{
		{utils.CGRID, qryFltr.CGRIDs},
		{utils.RunID, qryFltr.RunIDs},
		{utils.OriginID, qryFltr.OriginIDs},
		{utils.OriginHost, qryFltr.OriginHosts},
		{utils.CostSource, qryFltr.CostSources},
	} {
		if len(fltrSlc.ids) == 0 {
			continue
		}
		grpMpIDs := make(utils.StringMap)
		for _, id := range fltrSlc.ids {
			grpIDs := iDB.db.GetGroupItemIDs(utils.SessionCostsTBL, utils.ConcatenatedKey(fltrSlc.key, id))
			for _, id := range grpIDs {
				grpMpIDs[id] = true
			}
		}
		if len(grpMpIDs) == 0 {
			return utils.ErrNotFound
		}
		if smMpIDs == nil {
			smMpIDs = grpMpIDs
		} else {
			for id := range smMpIDs {
				if !grpMpIDs.HasKey(id) {
					delete(smMpIDs, id)
					if len(smMpIDs) == 0 {
						return utils.ErrNotFound
					}
				}
			}
		}
	}

	if smMpIDs == nil {
		smMpIDs = utils.StringMapFromSlice(iDB.db.GetItemIDs(utils.SessionCostsTBL, utils.EmptyString))
	}

	// check for Not filters
	for _, fltrSlc := range []struct {
		key string
		ids []string
	}{
		{utils.CGRID, qryFltr.NotCGRIDs},
		{utils.RunID, qryFltr.NotRunIDs},
		{utils.OriginID, qryFltr.NotOriginIDs},
		{utils.OriginHost, qryFltr.NotOriginHosts},
		{utils.CostSource, qryFltr.NotCostSources},
	} {
		if len(fltrSlc.ids) == 0 {
			continue
		}
		for _, id := range fltrSlc.ids {
			grpIDs := iDB.db.GetGroupItemIDs(utils.CDRsTBL, utils.ConcatenatedKey(fltrSlc.key, id))
			for _, id := range grpIDs {
				if smMpIDs.HasKey(id) {
					delete(smMpIDs, id)
					if len(smMpIDs) == 0 {
						return utils.ErrNotFound
					}
				}
			}
		}
	}

	if len(smMpIDs) == 0 {
		return utils.ErrNotFound
	}

	for key := range smMpIDs {
		iDB.db.Remove(utils.SessionCostsTBL, key,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return nil
}

// GetCDRs returns the CDRs from  DB based on given filters
func (iDB *InternalDB) GetCDRs(filter *utils.CDRsFilter, remove bool) (cdrs []*CDR, count int64, err error) {
	// filterPair used only for GetCDRs for internalDB
	type filterPair struct {
		key string
		ids []string
	}
	var pairSlice []filterPair
	var notPairSlice []filterPair
	// get indexed fields
	if len(iDB.stringIndexedFields) != 0 || len(iDB.prefixIndexedFields) != 0 {
		for _, cdrIdx := range iDB.stringIndexedFields {
			switch cdrIdx {
			case utils.CGRID:
				pairSlice = append(pairSlice, filterPair{utils.CGRID, filter.CGRIDs})
				notPairSlice = append(notPairSlice, filterPair{utils.CGRID, filter.NotCGRIDs})
				filter.NotCGRIDs = nil
				filter.CGRIDs = nil
			case utils.RunID:
				pairSlice = append(pairSlice, filterPair{utils.RunID, filter.RunIDs})
				notPairSlice = append(notPairSlice, filterPair{utils.RunID, filter.NotRunIDs})
				filter.NotRunIDs = nil
				filter.RunIDs = nil
			case utils.OriginID:
				pairSlice = append(pairSlice, filterPair{utils.OriginID, filter.OriginIDs})
				notPairSlice = append(notPairSlice, filterPair{utils.OriginID, filter.NotOriginIDs})
				filter.NotOriginIDs = nil
				filter.OriginIDs = nil
			case utils.OriginHost:
				pairSlice = append(pairSlice, filterPair{utils.OriginHost, filter.OriginHosts})
				notPairSlice = append(notPairSlice, filterPair{utils.OriginHost, filter.NotOriginHosts})
				filter.NotOriginHosts = nil
				filter.OriginHosts = nil
			case utils.Source:
				pairSlice = append(pairSlice, filterPair{utils.Source, filter.Sources})
				notPairSlice = append(notPairSlice, filterPair{utils.Source, filter.NotSources})
				filter.NotSources = nil
				filter.Sources = nil
			case utils.ToR:
				pairSlice = append(pairSlice, filterPair{utils.ToR, filter.ToRs})
				notPairSlice = append(notPairSlice, filterPair{utils.ToR, filter.NotToRs})
				filter.NotToRs = nil
				filter.ToRs = nil
			case utils.RequestType:
				pairSlice = append(pairSlice, filterPair{utils.RequestType, filter.RequestTypes})
				notPairSlice = append(notPairSlice, filterPair{utils.RequestType, filter.NotRequestTypes})
				filter.NotRequestTypes = nil
				filter.RequestTypes = nil
			case utils.Tenant:
				pairSlice = append(pairSlice, filterPair{utils.Tenant, filter.Tenants})
				notPairSlice = append(notPairSlice, filterPair{utils.Tenant, filter.NotTenants})
				filter.NotTenants = nil
				filter.Tenants = nil
			case utils.Category:
				pairSlice = append(pairSlice, filterPair{utils.Category, filter.Categories})
				notPairSlice = append(notPairSlice, filterPair{utils.Category, filter.NotCategories})
				filter.NotCategories = nil
				filter.Categories = nil
			case utils.Account:
				pairSlice = append(pairSlice, filterPair{utils.Account, filter.Accounts})
				notPairSlice = append(notPairSlice, filterPair{utils.Account, filter.NotAccounts})
				filter.NotAccounts = nil
				filter.Accounts = nil
			case utils.Subject:
				pairSlice = append(pairSlice, filterPair{utils.Subject, filter.Subjects})
				notPairSlice = append(notPairSlice, filterPair{utils.Subject, filter.NotSubjects})
				filter.NotSubjects = nil
				filter.Subjects = nil
			default:
				if val, has := filter.ExtraFields[cdrIdx]; has && val != utils.MetaExists { // if the filter value is *exist it should not be treated as a indexed field
					pairSlice = append(pairSlice, filterPair{cdrIdx, []string{val}})
					delete(filter.ExtraFields, cdrIdx)
				}
				if val, has := filter.NotExtraFields[cdrIdx]; has && val != utils.MetaExists { // if the filter value is *exist it should not be treated as a indexed field
					notPairSlice = append(notPairSlice, filterPair{cdrIdx, []string{val}})
					delete(filter.NotExtraFields, cdrIdx)
				}
			}
		}
		for _, cdrIdx := range iDB.prefixIndexedFields {
			switch cdrIdx {
			case utils.Destination:
				pairSlice = append(pairSlice, filterPair{utils.Destination, filter.DestinationPrefixes})
				notPairSlice = append(notPairSlice, filterPair{utils.Destination, filter.NotDestinationPrefixes})
				filter.DestinationPrefixes = nil
				filter.NotDestinationPrefixes = nil
			default:
				if val, has := filter.ExtraFields[cdrIdx]; has && val != utils.MetaExists { // if the filter value is *exist it should not be treated as a indexed field
					pairSlice = append(pairSlice, filterPair{cdrIdx, []string{val}})
					delete(filter.ExtraFields, cdrIdx)
				}
				if val, has := filter.NotExtraFields[cdrIdx]; has && val != utils.MetaExists { // if the filter value is *exist it should not be treated as a indexed field
					notPairSlice = append(notPairSlice, filterPair{cdrIdx, []string{val}})
					delete(filter.NotExtraFields, cdrIdx)
				}
			}
		}
	} else {
		pairSlice = []filterPair{
			{utils.CGRID, filter.CGRIDs},
			{utils.RunID, filter.RunIDs},
			{utils.OriginID, filter.OriginIDs},
			{utils.OriginHost, filter.OriginHosts},
			{utils.Source, filter.Sources},
			{utils.ToR, filter.ToRs},
			{utils.RequestType, filter.RequestTypes},
			{utils.Tenant, filter.Tenants},
			{utils.Category, filter.Categories},
			{utils.Account, filter.Accounts},
			{utils.Subject, filter.Subjects},
			{utils.Destination, filter.DestinationPrefixes},
		}
		notPairSlice = []filterPair{
			{utils.CGRID, filter.NotCGRIDs},
			{utils.RunID, filter.NotRunIDs},
			{utils.OriginID, filter.NotOriginIDs},
			{utils.OriginHost, filter.NotOriginHosts},
			{utils.Source, filter.NotSources},
			{utils.ToR, filter.NotToRs},
			{utils.RequestType, filter.NotRequestTypes},
			{utils.Tenant, filter.NotTenants},
			{utils.Category, filter.NotCategories},
			{utils.Account, filter.NotAccounts},
			{utils.Subject, filter.NotSubjects},
			{utils.Destination, filter.NotDestinationPrefixes},
		}
		filter.CGRIDs = nil
		filter.RunIDs = nil
		filter.OriginIDs = nil
		filter.OriginHosts = nil
		filter.Sources = nil
		filter.ToRs = nil
		filter.RequestTypes = nil
		filter.Tenants = nil
		filter.Categories = nil
		filter.Accounts = nil
		filter.Subjects = nil
		filter.DestinationPrefixes = nil
		filter.NotCGRIDs = nil
		filter.NotRunIDs = nil
		filter.NotOriginIDs = nil
		filter.NotOriginHosts = nil
		filter.NotSources = nil
		filter.NotToRs = nil
		filter.NotRequestTypes = nil
		filter.NotTenants = nil
		filter.NotCategories = nil
		filter.NotAccounts = nil
		filter.NotSubjects = nil
		filter.NotDestinationPrefixes = nil
	}

	// find indexed fields
	var cdrMpIDs utils.StringSet
	// Apply string filter
	for _, fltrSlc := range pairSlice {
		if len(fltrSlc.ids) == 0 {
			continue
		}
		grpMpIDs := make(utils.StringSet)
		for _, id := range fltrSlc.ids {
			grpMpIDs.AddSlice(iDB.db.GetGroupItemIDs(utils.CDRsTBL, utils.ConcatenatedKey(fltrSlc.key, id)))
		}
		if grpMpIDs.Size() == 0 {
			if filter.Count {
				return nil, 0, nil
			}
			return nil, 0, utils.ErrNotFound
		}
		if cdrMpIDs == nil {
			cdrMpIDs = grpMpIDs
			continue
		}
		cdrMpIDs.Intersect(grpMpIDs)
		if cdrMpIDs.Size() == 0 {
			if filter.Count {
				return nil, 0, nil
			}
			return nil, 0, utils.ErrNotFound
		}
	}
	if cdrMpIDs == nil {
		cdrMpIDs = utils.NewStringSet(iDB.db.GetItemIDs(utils.CDRsTBL, utils.EmptyString))
	}
	// check for Not filters
	for _, fltrSlc := range notPairSlice {
		if len(fltrSlc.ids) == 0 {
			continue
		}
		for _, id := range fltrSlc.ids {
			for _, id := range iDB.db.GetGroupItemIDs(utils.CDRsTBL, utils.ConcatenatedKey(fltrSlc.key, id)) {
				if !cdrMpIDs.Has(id) {
					continue
				}
				cdrMpIDs.Remove(id)
				if cdrMpIDs.Size() == 0 {
					if filter.Count {
						return nil, 0, nil
					}
					return nil, 0, utils.ErrNotFound
				}
			}
		}
	}

	if cdrMpIDs.Size() == 0 {
		if filter.Count {
			return nil, 0, nil
		}
		return nil, 0, utils.ErrNotFound
	}

	// check non indexed fields
	var minUsage time.Duration
	var maxUsage time.Duration
	if len(filter.MinUsage) != 0 {
		minUsage, err = utils.ParseDurationWithNanosecs(filter.MinUsage)
		if err != nil {
			return nil, 0, err
		}
	}
	if len(filter.MaxUsage) != 0 {
		maxUsage, err = utils.ParseDurationWithNanosecs(filter.MaxUsage)
		if err != nil {
			return nil, 0, err
		}
	}
	var offset int
	if filter.Paginator.Offset != nil {
		offset = *filter.Paginator.Offset
	}
	filter.Prepare()
	for key := range cdrMpIDs {
		x, ok := iDB.db.Get(utils.CDRsTBL, key)
		if !ok || x == nil {
			return nil, 0, utils.ErrNotFound
		}
		cdr := x.(*CDR)

		// default indexed filters
		if (len(filter.CGRIDs) > 0 && !utils.SliceHasMember(filter.CGRIDs, cdr.CGRID)) ||
			(len(filter.RunIDs) > 0 && !utils.SliceHasMember(filter.RunIDs, cdr.RunID)) ||
			(len(filter.OriginIDs) > 0 && !utils.SliceHasMember(filter.OriginIDs, cdr.OriginID)) ||
			(len(filter.OriginHosts) > 0 && !utils.SliceHasMember(filter.OriginHosts, cdr.OriginHost)) ||
			(len(filter.Sources) > 0 && !utils.SliceHasMember(filter.Sources, cdr.Source)) ||
			(len(filter.ToRs) > 0 && !utils.SliceHasMember(filter.ToRs, cdr.ToR)) ||
			(len(filter.RequestTypes) > 0 && !utils.SliceHasMember(filter.RequestTypes, cdr.RequestType)) ||
			(len(filter.Tenants) > 0 && !utils.SliceHasMember(filter.Tenants, cdr.Tenant)) ||
			(len(filter.Categories) > 0 && !utils.SliceHasMember(filter.Categories, cdr.Category)) ||
			(len(filter.Accounts) > 0 && !utils.SliceHasMember(filter.Accounts, cdr.Account)) ||
			(len(filter.Subjects) > 0 && !utils.SliceHasMember(filter.Subjects, cdr.Subject)) ||

			(len(filter.NotCGRIDs) > 0 && utils.SliceHasMember(filter.NotCGRIDs, cdr.CGRID)) ||
			(len(filter.NotRunIDs) > 0 && utils.SliceHasMember(filter.NotRunIDs, cdr.RunID)) ||
			(len(filter.NotOriginIDs) > 0 && utils.SliceHasMember(filter.NotOriginIDs, cdr.OriginID)) ||
			(len(filter.NotOriginHosts) > 0 && utils.SliceHasMember(filter.NotOriginHosts, cdr.OriginHost)) ||
			(len(filter.NotSources) > 0 && utils.SliceHasMember(filter.NotSources, cdr.Source)) ||
			(len(filter.NotToRs) > 0 && utils.SliceHasMember(filter.NotToRs, cdr.ToR)) ||
			(len(filter.NotRequestTypes) > 0 && utils.SliceHasMember(filter.NotRequestTypes, cdr.RequestType)) ||
			(len(filter.NotTenants) > 0 && utils.SliceHasMember(filter.NotTenants, cdr.Tenant)) ||
			(len(filter.NotCategories) > 0 && utils.SliceHasMember(filter.NotCategories, cdr.Category)) ||
			(len(filter.NotAccounts) > 0 && utils.SliceHasMember(filter.NotAccounts, cdr.Account)) ||
			(len(filter.NotSubjects) > 0 && utils.SliceHasMember(filter.NotSubjects, cdr.Subject)) ||

			(len(filter.Costs) > 0 && !utils.Float64SliceHasMember(filter.Costs, cdr.Cost)) ||
			(len(filter.NotCosts) > 0 && utils.Float64SliceHasMember(filter.NotCosts, cdr.Cost)) ||

			(len(filter.DestinationPrefixes) > 0 && !utils.HasPrefixSlice(filter.DestinationPrefixes, cdr.Destination)) ||
			(len(filter.NotDestinationPrefixes) > 0 && utils.HasPrefixSlice(filter.NotDestinationPrefixes, cdr.Destination)) ||

			(filter.OrderIDStart != nil && cdr.OrderID < *filter.OrderIDStart) ||
			(filter.OrderIDEnd != nil && cdr.OrderID >= *filter.OrderIDEnd) ||

			(filter.AnswerTimeStart != nil && !filter.AnswerTimeStart.IsZero() && cdr.AnswerTime.Before(*filter.AnswerTimeStart)) ||
			(filter.AnswerTimeEnd != nil && !filter.AnswerTimeEnd.IsZero() && cdr.AnswerTime.After(*filter.AnswerTimeEnd)) ||
			(filter.SetupTimeStart != nil && !filter.SetupTimeStart.IsZero() && cdr.SetupTime.Before(*filter.SetupTimeStart)) ||
			(filter.SetupTimeEnd != nil && !filter.SetupTimeEnd.IsZero() && cdr.SetupTime.Before(*filter.SetupTimeEnd)) ||

			(len(filter.MinUsage) != 0 && cdr.Usage < minUsage) ||
			(len(filter.MaxUsage) != 0 && cdr.Usage > maxUsage) {
			continue
		}

		// normal filters

		if filter.MinCost != nil {
			if filter.MaxCost == nil {
				if cdr.Cost < *filter.MinCost {
					continue
				}
			} else if *filter.MinCost == 0.0 && *filter.MaxCost == -1.0 { // Special case when we want to skip errors
				if cdr.Cost < 0 {
					continue
				}
			} else if cdr.Cost < *filter.MinCost || cdr.Cost >= *filter.MaxCost {
				continue
			}
		} else if filter.MaxCost != nil {
			if *filter.MaxCost == -1.0 { // Non-rated CDRs
				if cdr.Cost < 0 {
					continue
				}
			} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
				if cdr.Cost >= *filter.MaxCost {
					continue
				}
			}
		}
		if len(filter.ExtraFields) != 0 {
			passFilter := true
			for extFldID, extFldVal := range filter.ExtraFields {
				val, has := cdr.ExtraFields[extFldID]
				passFilter = val == extFldVal
				if extFldVal == utils.MetaExists {
					passFilter = has
				}
				if !passFilter {
					break
				}
			}
			if !passFilter {
				continue
			}
		}
		if len(filter.NotExtraFields) != 0 {
			passFilter := true
			for notExtFldID, notExtFldVal := range filter.NotExtraFields {
				val, has := cdr.ExtraFields[notExtFldID]
				passFilter = val != notExtFldVal
				if notExtFldVal == utils.MetaExists {
					passFilter = !has
				}
				if !passFilter {
					break
				}
			}
			if !passFilter {
				continue
			}
		}

		if filter.OrderBy == utils.EmptyString { // if do not have to order exit early
			if offset > 0 {
				offset--
				continue
			}
			if filter.Paginator.Limit != nil &&
				len(cdrs) >= *filter.Paginator.Limit {
				break
			}
		}
		//pass all filters and append to slice
		cdrs = append(cdrs, cdr)
	}
	if len(cdrs) <= offset { // if we have offset populated but not enough cdrs return
		if filter.Count {
			return nil, 0, nil
		}
		return nil, 0, utils.ErrNotFound
	}
	if filter.Count {
		cdrs = cdrs[offset:]
		if filter.Paginator.Limit != nil &&
			len(cdrs) >= *filter.Paginator.Limit {
			return nil, int64(*filter.Paginator.Limit), nil
		}
		return nil, int64(len(cdrs)), nil
	}
	if len(cdrs) == 0 {
		if remove {
			return nil, 0, nil
		}
		return nil, 0, utils.ErrNotFound
	}

	if filter.OrderBy != utils.EmptyString {
		separateVals := strings.Split(filter.OrderBy, utils.INFIELD_SEP)
		ascendent := !(len(separateVals) == 2 && separateVals[1] == "desc")
		switch separateVals[0] {
		case utils.OrderID:
			if ascendent {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].OrderID < cdrs[j].OrderID
				})
			} else {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].OrderID > cdrs[j].OrderID
				})
			}
		case utils.AnswerTime:
			if ascendent {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].AnswerTime.Before(cdrs[j].AnswerTime)
				})
			} else {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].AnswerTime.After(cdrs[j].AnswerTime)
				})
			}
		case utils.SetupTime:
			if ascendent {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].SetupTime.Before(cdrs[j].SetupTime)
				})
			} else {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].SetupTime.After(cdrs[j].SetupTime)
				})
			}
		case utils.Usage:
			if ascendent {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].Usage < cdrs[j].Usage
				})
			} else {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].Usage > cdrs[j].Usage
				})
			}
		case utils.Cost:
			if ascendent {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].Cost < cdrs[j].Cost
				})
			} else {
				sort.Slice(cdrs, func(i, j int) bool {
					return cdrs[i].Cost > cdrs[j].Cost
				})
			}
		default:
			return nil, 0, fmt.Errorf("Invalid value : %s", separateVals[0])
		}

		cdrs = cdrs[offset:]
		if filter.Paginator.Limit != nil {
			if len(cdrs) > *filter.Paginator.Limit {
				cdrs = cdrs[:*filter.Paginator.Limit]
			}
		}
	}
	if remove {
		for _, cdr := range cdrs {
			iDB.db.Remove(utils.CDRsTBL, utils.ConcatenatedKey(cdr.CGRID, cdr.RunID, cdr.OriginID),
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
		}
		return nil, 0, nil
	}
	return
}

func (iDB *InternalDB) GetSMCosts(cgrid, runid, originHost, originIDPrfx string) (smCosts []*SMCost, err error) {
	var smMpIDs utils.StringMap
	for _, fltrSlc := range []struct {
		key string
		id  string
	}{
		{utils.CGRID, cgrid},
		{utils.RunID, runid},
		{utils.OriginHost, originHost},
	} {
		if fltrSlc.id == utils.EmptyString {
			continue
		}
		grpMpIDs := make(utils.StringMap)

		grpIDs := iDB.db.GetGroupItemIDs(utils.SessionCostsTBL, utils.ConcatenatedKey(fltrSlc.key, fltrSlc.id))
		for _, id := range grpIDs {
			grpMpIDs[id] = true
		}

		if len(grpMpIDs) == 0 {
			return nil, utils.ErrNotFound
		}
		if smMpIDs == nil {
			smMpIDs = grpMpIDs
		} else {
			for id := range smMpIDs {
				if !grpMpIDs.HasKey(id) {
					delete(smMpIDs, id)
					if len(smMpIDs) == 0 {
						return nil, utils.ErrNotFound
					}
				}
			}
		}
	}
	if smMpIDs == nil {
		smMpIDs = utils.StringMapFromSlice(iDB.db.GetItemIDs(utils.SessionCostsTBL, utils.EmptyString))
	}
	if len(smMpIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	for key := range smMpIDs {
		x, ok := iDB.db.Get(utils.SessionCostsTBL, key)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		smCost := x.(*SMCost)
		if originIDPrfx != utils.EmptyString && !strings.HasPrefix(smCost.OriginID, originIDPrfx) {
			continue
		}
		smCosts = append(smCosts, smCost)
	}
	return
}

func (iDB *InternalDB) SetSMCost(smCost *SMCost) (err error) {
	if smCost.CostDetails == nil {
		return nil
	}
	idxs := utils.NewStringSet(nil)
	idxs.Add(utils.ConcatenatedKey(utils.CGRID, smCost.CGRID))
	idxs.Add(utils.ConcatenatedKey(utils.RunID, smCost.RunID))
	idxs.Add(utils.ConcatenatedKey(utils.OriginHost, smCost.OriginHost))
	idxs.Add(utils.ConcatenatedKey(utils.OriginID, smCost.OriginID))
	idxs.Add(utils.ConcatenatedKey(utils.CostSource, smCost.CostSource))
	iDB.db.Set(utils.SessionCostsTBL, utils.ConcatenatedKey(smCost.CGRID, smCost.RunID, smCost.OriginHost, smCost.OriginID), smCost, idxs.AsSlice(),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return err
}
