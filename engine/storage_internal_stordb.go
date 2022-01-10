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

// GetTpIds implements LoadReader interface
func (iDB *InternalDB) GetTpIds(colName string) (ids []string, err error) {
	tpIDs := make(utils.StringSet)
	if colName == utils.EmptyString { // if colName is empty we need to parse all partitions
		for _, conNm := range utils.CacheStorDBPartitions { // iterate through all columns
			for _, key := range iDB.db.GetItemIDs(conNm, utils.EmptyString) {
				tpIDs.Add(strings.Split(key, utils.InInFieldSep)[0])
			}
		}
	} else {
		for _, key := range iDB.db.GetItemIDs(utils.CacheStorDBPartitions[colName], utils.EmptyString) {
			tpIDs.Add(strings.Split(key, utils.InInFieldSep)[0])
		}
	}
	return tpIDs.AsSlice(), nil
}

func (iDB *InternalDB) GetTpTableIds(tpid, table string, distinct []string,
	filters map[string]string, paginator *utils.PaginatorWithSearch) (ids []string, err error) {
	fullIDs := iDB.db.GetItemIDs(utils.CacheStorDBPartitions[table], tpid)
	idSet := make(utils.StringSet)
	for _, fullID := range fullIDs {
		idSet.Add(fullID[len(tpid)+1:])
	}
	ids = idSet.AsSlice()
	return
}

func (iDB *InternalDB) GetTPResources(tpid, tenant, id string) (resources []*utils.TPResourceProfile, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPResources, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPResources, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPStats, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPStats, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPThresholds, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPThresholds, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPFilters, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPFilters, id)
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

func (iDB *InternalDB) GetTPRoutes(tpid, tenant, id string) (supps []*utils.TPRouteProfile, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPRoutes, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPRoutes, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		supps = append(supps, x.(*utils.TPRouteProfile))

	}
	if len(supps) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPAttributes(tpid, tenant, id string) (attrs []*utils.TPAttributeProfile, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPAttributes, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPAttributes, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPChargers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPChargers, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPDispatchers, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPDispatchers, id)
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
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPDispatcherHosts, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPDispatcherHosts, id)
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

func (iDB *InternalDB) GetTPRateProfiles(tpid, tenant, id string) (tpPrfs []*utils.TPRateProfile, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPRateProfiles, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPRateProfiles, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		tpPrfs = append(tpPrfs, x.(*utils.TPRateProfile))
	}
	if len(tpPrfs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPActionProfiles(tpid, tenant, id string) (tpPrfs []*utils.TPActionProfile, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPActionProfiles, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPActionProfiles, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		tpPrfs = append(tpPrfs, x.(*utils.TPActionProfile))
	}
	if len(tpPrfs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetTPAccounts(tpid, tenant, id string) (tpPrfs []*utils.TPAccount, err error) {
	key := tpid
	if tenant != utils.EmptyString {
		key += utils.ConcatenatedKeySep + tenant
	}
	if id != utils.EmptyString {
		key += utils.ConcatenatedKeySep + id
	}
	ids := iDB.db.GetItemIDs(utils.CacheTBLTPAccounts, key)
	for _, id := range ids {
		x, ok := iDB.db.Get(utils.CacheTBLTPAccounts, id)
		if !ok || x == nil {
			return nil, utils.ErrNotFound
		}
		tpPrfs = append(tpPrfs, x.(*utils.TPAccount))
	}
	if len(tpPrfs) == 0 {
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
		if tag, has := args["tag"]; has {
			key += utils.ConcatenatedKeySep + tag
		} else if id, has := args["id"]; has {
			key += utils.ConcatenatedKeySep + args["tenant"] +
				utils.ConcatenatedKeySep + id
		}
	}
	ids := iDB.db.GetItemIDs(utils.CacheStorDBPartitions[table], key)
	for _, id := range ids {
		iDB.db.Remove(utils.CacheStorDBPartitions[table], id,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPResources(resources []*utils.TPResourceProfile) (err error) {
	if len(resources) == 0 {
		return nil
	}
	for _, resource := range resources {
		iDB.db.Set(utils.CacheTBLTPResources, utils.ConcatenatedKey(resource.TPid, resource.Tenant, resource.ID), resource, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPStats(stats []*utils.TPStatProfile) (err error) {
	if len(stats) == 0 {
		return nil
	}
	for _, stat := range stats {
		iDB.db.Set(utils.CacheTBLTPStats, utils.ConcatenatedKey(stat.TPid, stat.Tenant, stat.ID), stat, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPThresholds(thresholds []*utils.TPThresholdProfile) (err error) {
	if len(thresholds) == 0 {
		return nil
	}

	for _, threshold := range thresholds {
		iDB.db.Set(utils.CacheTBLTPThresholds, utils.ConcatenatedKey(threshold.TPid, threshold.Tenant, threshold.ID), threshold, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPFilters(filters []*utils.TPFilterProfile) (err error) {
	if len(filters) == 0 {
		return nil
	}

	for _, filter := range filters {
		iDB.db.Set(utils.CacheTBLTPFilters, utils.ConcatenatedKey(filter.TPid, filter.Tenant, filter.ID), filter, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRoutes(routes []*utils.TPRouteProfile) (err error) {
	if len(routes) == 0 {
		return nil
	}
	for _, route := range routes {
		iDB.db.Set(utils.CacheTBLTPRoutes, utils.ConcatenatedKey(route.TPid, route.Tenant, route.ID), route, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAttributes(attributes []*utils.TPAttributeProfile) (err error) {
	if len(attributes) == 0 {
		return nil
	}

	for _, attribute := range attributes {
		iDB.db.Set(utils.CacheTBLTPAttributes, utils.ConcatenatedKey(attribute.TPid, attribute.Tenant, attribute.ID), attribute, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPChargers(cpps []*utils.TPChargerProfile) (err error) {
	if len(cpps) == 0 {
		return nil
	}

	for _, cpp := range cpps {
		iDB.db.Set(utils.CacheTBLTPChargers, utils.ConcatenatedKey(cpp.TPid, cpp.Tenant, cpp.ID), cpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherProfiles(dpps []*utils.TPDispatcherProfile) (err error) {
	if len(dpps) == 0 {
		return nil
	}

	for _, dpp := range dpps {
		iDB.db.Set(utils.CacheTBLTPDispatchers, utils.ConcatenatedKey(dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
func (iDB *InternalDB) SetTPDispatcherHosts(dpps []*utils.TPDispatcherHost) (err error) {
	if len(dpps) == 0 {
		return nil
	}
	for _, dpp := range dpps {
		iDB.db.Set(utils.CacheTBLTPDispatcherHosts, utils.ConcatenatedKey(dpp.TPid, dpp.Tenant, dpp.ID), dpp, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPRateProfiles(tpPrfs []*utils.TPRateProfile) (err error) {
	if len(tpPrfs) == 0 {
		return nil
	}
	for _, tpPrf := range tpPrfs {
		iDB.db.Set(utils.CacheTBLTPRateProfiles, utils.ConcatenatedKey(tpPrf.TPid, tpPrf.Tenant, tpPrf.ID), tpPrf, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPActionProfiles(tpPrfs []*utils.TPActionProfile) (err error) {
	if len(tpPrfs) == 0 {
		return nil
	}
	for _, tpPrf := range tpPrfs {
		iDB.db.Set(utils.CacheTBLTPActionProfiles, utils.ConcatenatedKey(tpPrf.TPid, tpPrf.Tenant, tpPrf.ID), tpPrf, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) SetTPAccounts(tpPrfs []*utils.TPAccount) (err error) {
	if len(tpPrfs) == 0 {
		return nil
	}
	for _, tpPrf := range tpPrfs {
		iDB.db.Set(utils.CacheTBLTPAccounts, utils.ConcatenatedKey(tpPrf.TPid, tpPrf.Tenant, tpPrf.ID), tpPrf, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}
