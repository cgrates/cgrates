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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// internalDBCacheCfg indexes the internal DataDB partitions
func newInternalDBCfg(itemsCacheCfg map[string]*config.ItemOpt, isDataDB bool) map[string]*ltcache.CacheConfig {
	if isDataDB {
		return map[string]*ltcache.CacheConfig{
			utils.CacheDestinations: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheDestinations].Limit,
				TTL:       itemsCacheCfg[utils.CacheDestinations].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheDestinations].StaticTTL,
			},
			utils.CacheReverseDestinations: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheReverseDestinations].Limit,
				TTL:       itemsCacheCfg[utils.CacheReverseDestinations].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheReverseDestinations].StaticTTL,
			},
			utils.CacheActions: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheActions].Limit,
				TTL:       itemsCacheCfg[utils.CacheActions].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheActions].StaticTTL,
			},
			utils.CacheActionPlans: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheActionPlans].Limit,
				TTL:       itemsCacheCfg[utils.CacheActionPlans].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheActionPlans].StaticTTL,
			},
			utils.CacheAccountActionPlans: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheAccountActionPlans].Limit,
				TTL:       itemsCacheCfg[utils.CacheAccountActionPlans].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheAccountActionPlans].StaticTTL,
			},
			utils.CacheActionTriggers: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheActionTriggers].Limit,
				TTL:       itemsCacheCfg[utils.CacheActionTriggers].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheActionTriggers].StaticTTL,
			},
			utils.CacheRatingPlans: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheRatingPlans].Limit,
				TTL:       itemsCacheCfg[utils.CacheRatingPlans].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheRatingPlans].StaticTTL,
			},
			utils.CacheRatingProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheRatingProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheRatingProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheRatingProfiles].StaticTTL,
			},
			utils.CacheAccounts: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheAccounts].Limit,
				TTL:       itemsCacheCfg[utils.CacheAccounts].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheAccounts].StaticTTL,
			},
			utils.CacheSharedGroups: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheSharedGroups].Limit,
				TTL:       itemsCacheCfg[utils.CacheSharedGroups].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheSharedGroups].StaticTTL,
			},

			utils.CacheTimings: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheTimings].Limit,
				TTL:       itemsCacheCfg[utils.CacheTimings].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheTimings].StaticTTL,
			},
			utils.CacheFilters: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheFilters].Limit,
				TTL:       itemsCacheCfg[utils.CacheFilters].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheFilters].StaticTTL,
			},
			utils.CacheResourceProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheResourceProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheResourceProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheResourceProfiles].StaticTTL,
			},
			utils.CacheResourceFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheResources: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheResources].Limit,
				TTL:       itemsCacheCfg[utils.CacheResources].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheResources].StaticTTL,
			},
			utils.CacheStatFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheStatQueueProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheStatQueueProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheStatQueueProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheStatQueueProfiles].StaticTTL,
			},
			utils.CacheStatQueues: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheStatQueues].Limit,
				TTL:       itemsCacheCfg[utils.CacheStatQueues].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheStatQueues].StaticTTL,
			},
			utils.CacheThresholdFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheThresholdProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheThresholdProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheThresholdProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheThresholdProfiles].StaticTTL,
			},
			utils.CacheThresholds: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheThresholds].Limit,
				TTL:       itemsCacheCfg[utils.CacheThresholds].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheThresholds].StaticTTL,
			},
			utils.CacheSupplierFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheSupplierProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheSupplierProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheSupplierProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheSupplierProfiles].StaticTTL,
			},
			utils.CacheChargerFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheChargerProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheChargerProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheChargerProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheChargerProfiles].StaticTTL,
			},
			utils.CacheAttributeFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheAttributeProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheAttributeProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheAttributeProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheAttributeProfiles].StaticTTL,
			},
			utils.CacheDispatcherFilterIndexes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.MetaFilterIndexes].Limit,
				TTL:       itemsCacheCfg[utils.MetaFilterIndexes].TTL,
				StaticTTL: itemsCacheCfg[utils.MetaFilterIndexes].StaticTTL,
			},
			utils.CacheDispatcherProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheDispatcherProfiles].Limit,
				TTL:       itemsCacheCfg[utils.CacheDispatcherProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheDispatcherProfiles].StaticTTL,
			},
			utils.CacheDispatcherHosts: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheDispatcherHosts].Limit,
				TTL:       itemsCacheCfg[utils.CacheDispatcherHosts].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheDispatcherHosts].StaticTTL,
			},
			utils.CacheLoadIDs: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CacheLoadIDs].Limit,
				TTL:       itemsCacheCfg[utils.CacheLoadIDs].TTL,
				StaticTTL: itemsCacheCfg[utils.CacheLoadIDs].StaticTTL,
			},
		}
	} else {
		return map[string]*ltcache.CacheConfig{
			utils.TBLVersions: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLVersions].Limit,
				TTL:       itemsCacheCfg[utils.TBLVersions].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLVersions].StaticTTL,
			},
			utils.TBLTPTimings: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPTimings].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPTimings].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPTimings].StaticTTL,
			},
			utils.TBLTPDestinations: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPDestinations].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPDestinations].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPDestinations].StaticTTL,
			},
			utils.TBLTPRates: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPRates].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPRates].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPRates].StaticTTL,
			},
			utils.TBLTPDestinationRates: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPDestinationRates].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPDestinationRates].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPDestinationRates].StaticTTL,
			},
			utils.TBLTPRatingPlans: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPRatingPlans].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPRatingPlans].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPRatingPlans].StaticTTL,
			},
			utils.TBLTPRateProfiles: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPRateProfiles].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPRateProfiles].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPRateProfiles].StaticTTL,
			},
			utils.TBLTPSharedGroups: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPSharedGroups].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPSharedGroups].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPSharedGroups].StaticTTL,
			},
			utils.TBLTPActions: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPActions].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPActions].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPActions].StaticTTL,
			},
			utils.TBLTPActionTriggers: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPActionTriggers].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPActionTriggers].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPActionTriggers].StaticTTL,
			},
			utils.TBLTPAccountActions: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPAccountActions].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPAccountActions].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPAccountActions].StaticTTL,
			},
			utils.TBLTPResources: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPResources].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPResources].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPResources].StaticTTL,
			},
			utils.TBLTPStats: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPStats].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPStats].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPStats].StaticTTL,
			},
			utils.TBLTPThresholds: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPThresholds].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPThresholds].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPThresholds].StaticTTL,
			},
			utils.TBLTPFilters: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPFilters].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPFilters].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPFilters].StaticTTL,
			},
			utils.SessionCostsTBL: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.SessionCostsTBL].Limit,
				TTL:       itemsCacheCfg[utils.SessionCostsTBL].TTL,
				StaticTTL: itemsCacheCfg[utils.SessionCostsTBL].StaticTTL,
			},
			utils.TBLTPActionPlans: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPActionPlans].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPActionPlans].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPActionPlans].StaticTTL,
			},
			utils.TBLTPSuppliers: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPSuppliers].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPSuppliers].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPSuppliers].StaticTTL,
			},
			utils.TBLTPAttributes: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPAttributes].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPAttributes].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPAttributes].StaticTTL,
			},
			utils.TBLTPChargers: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPChargers].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPChargers].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPChargers].StaticTTL,
			},
			utils.TBLTPDispatchers: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPDispatchers].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPDispatchers].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPDispatchers].StaticTTL,
			},
			utils.TBLTPDispatcherHosts: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.TBLTPDispatcherHosts].Limit,
				TTL:       itemsCacheCfg[utils.TBLTPDispatcherHosts].TTL,
				StaticTTL: itemsCacheCfg[utils.TBLTPDispatcherHosts].StaticTTL,
			},
			utils.CDRsTBL: &ltcache.CacheConfig{
				MaxItems:  itemsCacheCfg[utils.CDRsTBL].Limit,
				TTL:       itemsCacheCfg[utils.CDRsTBL].TTL,
				StaticTTL: itemsCacheCfg[utils.CDRsTBL].StaticTTL,
			},
		}
	}
}

type InternalDB struct {
	tasks               []*Task
	db                  *ltcache.TransCache
	mu                  sync.RWMutex
	stringIndexedFields []string
	prefixIndexedFields []string
	indexedFieldsMutex  sync.RWMutex   // used for reload
	cnter               *utils.Counter // used for OrderID for cdr
	ms                  Marshaler
}

// NewInternalDB constructs an InternalDB
func NewInternalDB(stringIndexedFields, prefixIndexedFields []string,
	isDataDB bool, itemsCacheCfg map[string]*config.ItemOpt) (iDB *InternalDB) {
	ms, _ := NewMarshaler(config.CgrConfig().GeneralCfg().DBDataEncoding)
	iDB = &InternalDB{
		db:                  ltcache.NewTransCache(newInternalDBCfg(itemsCacheCfg, isDataDB)),
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields,
		cnter:               utils.NewCounter(time.Now().UnixNano(), 0),
		ms:                  ms,
	}
	return
}

// SetStringIndexedFields set the stringIndexedFields, used at StorDB reload (is thread safe)
func (iDB *InternalDB) SetStringIndexedFields(stringIndexedFields []string) {
	iDB.indexedFieldsMutex.Lock()
	iDB.stringIndexedFields = stringIndexedFields
	iDB.indexedFieldsMutex.Unlock()
}

// SetPrefixIndexedFields set the prefixIndexedFields, used at StorDB reload (is thread safe)
func (iDB *InternalDB) SetPrefixIndexedFields(prefixIndexedFields []string) {
	iDB.indexedFieldsMutex.Lock()
	iDB.prefixIndexedFields = prefixIndexedFields
	iDB.indexedFieldsMutex.Unlock()
}

func (iDB *InternalDB) Close() {}

func (iDB *InternalDB) Flush(_ string) error {
	iDB.db.Clear(nil)
	return nil
}

func (iDB *InternalDB) SelectDatabase(dbName string) (err error) {
	return nil
}

func (iDB *InternalDB) GetKeysForPrefix(prefix string) ([]string, error) {
	keyLen := len(utils.DESTINATION_PREFIX)
	if len(prefix) < keyLen {
		return nil, fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
	}
	category := prefix[:keyLen] // prefix length
	queryPrefix := prefix[keyLen:]
	ids := iDB.db.GetItemIDs(utils.CachePrefixToInstance[category], queryPrefix)
	for i := range ids {
		ids[i] = category + ids[i]
	}
	return ids, nil
}

func (iDB *InternalDB) RebuildReverseForPrefix(prefix string) (err error) {
	keys, err := iDB.GetKeysForPrefix(prefix)
	if err != nil {
		return err
	}
	for _, key := range keys {
		iDB.db.Remove(utils.CacheReverseDestinations, key,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		keys, err = iDB.GetKeysForPrefix(utils.DESTINATION_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			dest, err := iDB.GetDestinationDrv(key[len(utils.DESTINATION_PREFIX):], false, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err := iDB.SetReverseDestinationDrv(dest, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		return nil
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (iDB *InternalDB) RemoveReverseForPrefix(prefix string) (err error) {
	keys, err := iDB.GetKeysForPrefix(prefix)
	if err != nil {
		return err
	}
	for _, key := range keys {
		iDB.db.Remove(utils.CacheReverseDestinations, key,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	switch prefix {
	case utils.REVERSE_DESTINATION_PREFIX:
		keys, err = iDB.GetKeysForPrefix(utils.DESTINATION_PREFIX)
		if err != nil {
			return err
		}
		for _, key := range keys {
			dest, err := iDB.GetDestinationDrv(key[len(utils.DESTINATION_PREFIX):], false, utils.NonTransactional)
			if err != nil {
				return err
			}
			if err := iDB.RemoveDestinationDrv(dest.Id, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		return nil
	default:
		return utils.ErrInvalidKey
	}
	return nil
}

func (iDB *InternalDB) GetVersions(itm string) (vrs Versions, err error) {
	x, ok := iDB.db.Get(utils.TBLVersions, utils.Version)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	provVrs := x.(Versions)
	if itm != "" {
		if _, has := provVrs[itm]; !has {
			return nil, utils.ErrNotFound
		}
		return Versions{itm: provVrs[itm]}, nil
	}
	return provVrs, nil
}

func (iDB *InternalDB) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		if err = iDB.RemoveVersions(nil); err != nil {
			return err
		}
	}
	x, ok := iDB.db.Get(utils.TBLVersions, utils.Version)
	if !ok || x == nil {
		iDB.db.Set(utils.TBLVersions, utils.Version, vrs, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	provVrs := x.(Versions)
	for key, val := range vrs {
		provVrs[key] = val
	}
	iDB.db.Set(utils.TBLVersions, utils.Version, provVrs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) != 0 {
		var internalVersions Versions
		x, ok := iDB.db.Get(utils.TBLVersions, utils.Version)
		if !ok || x == nil {
			return utils.ErrNotFound
		}
		internalVersions = x.(Versions)
		for key := range vrs {
			delete(internalVersions, key)
		}
		iDB.db.Set(utils.TBLVersions, utils.Version, internalVersions, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return nil
	}
	iDB.db.Remove(utils.TBLVersions, utils.Version,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetStorageType() string {
	return utils.INTERNAL
}

func (iDB *InternalDB) IsDBEmpty() (resp bool, err error) {
	for cacheInstance, _ := range utils.CacheInstanceToPrefix {
		if len(iDB.db.GetItemIDs(cacheInstance, utils.EmptyString)) != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (iDB *InternalDB) HasDataDrv(category, subject, tenant string) (bool, error) {
	switch category {
	case utils.DESTINATION_PREFIX, utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACCOUNT_PREFIX:
		return iDB.db.HasItem(utils.CachePrefixToInstance[category], subject), nil
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.SupplierProfilePrefix, utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix, utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix:
		return iDB.db.HasItem(utils.CachePrefixToInstance[category], utils.ConcatenatedKey(tenant, subject)), nil
	}
	return false, errors.New("Unsupported HasData category")
}

func (iDB *InternalDB) GetRatingPlanDrv(id string) (rp *RatingPlan, err error) {
	x, ok := iDB.db.Get(utils.CacheRatingPlans, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RatingPlan), nil
}

func (iDB *InternalDB) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	iDB.db.Set(utils.CacheRatingPlans, rp.Id, rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRatingPlanDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheRatingPlans, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRatingProfileDrv(id string) (rp *RatingProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheRatingProfiles, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RatingProfile), nil
}

func (iDB *InternalDB) SetRatingProfileDrv(rp *RatingProfile) (err error) {
	iDB.db.Set(utils.CacheRatingProfiles, rp.Id, rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRatingProfileDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheRatingProfiles, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDestinationDrv(key string, skipCache bool, transactionID string) (dest *Destination, err error) {
	cCommit := cacheCommit(transactionID)

	if !skipCache {
		if x, ok := Cache.Get(utils.CacheDestinations, key); ok {
			if x != nil {
				return x.(*Destination), nil
			}
			return nil, utils.ErrNotFound
		}
	}

	x, ok := iDB.db.Get(utils.CacheDestinations, key)
	if !ok || x == nil {
		Cache.Set(utils.CacheDestinations, key, nil, nil, cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	dest = x.(*Destination)
	Cache.Set(utils.CacheDestinations, key, dest, nil, cCommit, transactionID)
	return
}

func (iDB *InternalDB) SetDestinationDrv(dest *Destination, transactionID string) (err error) {
	iDB.db.Set(utils.CacheDestinations, dest.Id, dest, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	Cache.Remove(utils.CacheDestinations, dest.Id,
		cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) RemoveDestinationDrv(destID string, transactionID string) (err error) {
	// get destination for prefix list
	d, err := iDB.GetDestinationDrv(destID, false, transactionID)
	if err != nil {
		return
	}
	iDB.db.Remove(utils.CacheDestinations, destID,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	Cache.Remove(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID)
	for _, prefix := range d.Prefixes {
		iDB.db.Remove(utils.CacheReverseDestinations, prefix,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		iDB.GetReverseDestinationDrv(prefix, true, transactionID) // it will recache the destination
	}
	return
}

func (iDB *InternalDB) SetReverseDestinationDrv(dest *Destination, transactionID string) (err error) {
	var mpRevDst utils.StringMap
	for _, p := range dest.Prefixes {
		if iDB.db.HasItem(utils.CacheReverseDestinations, p) {
			x, ok := iDB.db.Get(utils.CacheReverseDestinations, p)
			if !ok || x == nil {
				return utils.ErrNotFound
			}
			mpRevDst = x.(utils.StringMap)
		} else {
			mpRevDst = make(utils.StringMap)
		}
		mpRevDst[dest.Id] = true
		// for ReverseDestination we will use Groups
		iDB.db.Set(utils.CacheReverseDestinations, p, mpRevDst, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) GetReverseDestinationDrv(prefix string,
	skipCache bool, transactionID string) (ids []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheReverseDestinations, prefix); ok {
			if x != nil {
				return x.([]string), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	x, ok := iDB.db.Get(utils.CacheReverseDestinations, prefix)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	ids = x.(utils.StringMap).Slice()
	if len(ids) == 0 {
		Cache.Set(utils.CacheReverseDestinations, prefix, nil, nil,
			cacheCommit(transactionID), transactionID)
		return nil, utils.ErrNotFound
	}
	Cache.Set(utils.CacheReverseDestinations, prefix, ids, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) UpdateReverseDestinationDrv(oldDest, newDest *Destination,
	transactionID string) error {
	var obsoletePrefixes []string
	var mpRevDst utils.StringMap
	var addedPrefixes []string
	var found bool
	if oldDest == nil {
		oldDest = new(Destination) // so we can process prefixes
	}
	for _, oldPrefix := range oldDest.Prefixes {
		found = false
		for _, newPrefix := range newDest.Prefixes {
			if oldPrefix == newPrefix {
				found = true
				break
			}
		}
		if !found {
			obsoletePrefixes = append(obsoletePrefixes, oldPrefix)
		}
	}
	for _, newPrefix := range newDest.Prefixes {
		found = false
		for _, oldPrefix := range oldDest.Prefixes {
			if newPrefix == oldPrefix {
				found = true
				break
			}
		}
		if !found {
			addedPrefixes = append(addedPrefixes, newPrefix)
		}
	}
	// remove id for all obsolete prefixes
	cCommit := cacheCommit(transactionID)
	var err error
	for _, obsoletePrefix := range obsoletePrefixes {
		if iDB.db.HasItem(utils.CacheReverseDestinations, obsoletePrefix) {
			x, ok := iDB.db.Get(utils.CacheReverseDestinations, obsoletePrefix)
			if !ok || x == nil {
				return utils.ErrNotFound
			}
			mpRevDst = x.(utils.StringMap)
			if _, has := mpRevDst[oldDest.Id]; has {
				delete(mpRevDst, oldDest.Id)
			}
			// for ReverseDestination we will use Groups
			iDB.db.Set(utils.CacheReverseDestinations, obsoletePrefix, mpRevDst, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
		}

		Cache.Remove(utils.CacheReverseDestinations, obsoletePrefix,
			cCommit, transactionID)
	}
	// add the id to all new prefixes
	for _, addedPrefix := range addedPrefixes {
		if iDB.db.HasItem(utils.CacheReverseDestinations, addedPrefix) {
			x, ok := iDB.db.Get(utils.CacheReverseDestinations, addedPrefix)
			if !ok || x == nil {
				return utils.ErrNotFound
			}
			mpRevDst = x.(utils.StringMap)
		} else {
			mpRevDst = make(utils.StringMap)
		}
		mpRevDst[newDest.Id] = true
		// for ReverseDestination we will use Groups
		iDB.db.Set(utils.CacheReverseDestinations, addedPrefix, mpRevDst, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return err
}

func (iDB *InternalDB) GetActionsDrv(id string) (acts Actions, err error) {
	x, ok := iDB.db.Get(utils.CacheActions, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(Actions), err
}

func (iDB *InternalDB) SetActionsDrv(id string, acts Actions) (err error) {
	utils.Logger.Debug("ENTER IN SET ACTIONS DRV ??? ")
	iDB.db.Set(utils.CacheActions, id, acts, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	utils.Logger.Debug("Set With Success ")
	x, ok := iDB.db.Get(utils.CacheActions, id)
	if !ok || x == nil {
		utils.Logger.Debug("NOT FOUND ")
	}
	utils.Logger.Debug(fmt.Sprintf("ACTIONS : %+v", utils.ToJSON(x.(Actions))))
	return
}

func (iDB *InternalDB) RemoveActionsDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheActions, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetSharedGroupDrv(id string) (sh *SharedGroup, err error) {
	x, ok := iDB.db.Get(utils.CacheSharedGroups, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*SharedGroup).Clone(), nil
}

func (iDB *InternalDB) SetSharedGroupDrv(sh *SharedGroup) (err error) {
	iDB.db.Set(utils.CacheSharedGroups, sh.Id, sh, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveSharedGroupDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheSharedGroups, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionTriggersDrv(id string) (at ActionTriggers, err error) {
	x, ok := iDB.db.Get(utils.CacheActionTriggers, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(ActionTriggers).Clone(), nil
}

func (iDB *InternalDB) SetActionTriggersDrv(id string, at ActionTriggers) (err error) {
	iDB.db.Set(utils.CacheActionTriggers, id, at, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionTriggersDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheActionTriggers, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionPlanDrv(key string, skipCache bool,
	transactionID string) (ats *ActionPlan, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheActionPlans, key); ok {
			if x != nil {
				return x.(*ActionPlan), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	cCommit := cacheCommit(transactionID)
	x, ok := iDB.db.Get(utils.CacheActionPlans, key)
	if !ok || x == nil {
		Cache.Set(utils.CacheActionPlans, key, nil, nil,
			cCommit, transactionID)
		return nil, utils.ErrNotFound
	}
	ats = x.(*ActionPlan)
	Cache.Set(utils.CacheActionPlans, key, ats, nil,
		cCommit, transactionID)
	return
}

func (iDB *InternalDB) SetActionPlanDrv(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		iDB.db.Remove(utils.CacheActionPlans, key,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		Cache.Remove(utils.CacheActionPlans, key,
			cCommit, transactionID)
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := iDB.GetActionPlanDrv(key, true,
			transactionID); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}
	iDB.db.Set(utils.CacheActionPlans, key, ats, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionPlanDrv(key string, transactionID string) (err error) {
	iDB.db.Remove(utils.CacheActionPlans, key,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	Cache.Remove(utils.CacheActionPlans, key, cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
	keys, err := iDB.GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return nil, err
	}

	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		ap, err := iDB.GetActionPlanDrv(key[len(utils.ACTION_PLAN_PREFIX):], false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		ats[key[len(utils.ACTION_PLAN_PREFIX):]] = ap
	}
	return
}

func (iDB *InternalDB) GetAccountActionPlansDrv(acntID string,
	skipCache bool, transactionID string) (apIDs []string, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheAccountActionPlans, acntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	x, ok := iDB.db.Get(utils.CacheAccountActionPlans, acntID)
	if !ok || x == nil {
		Cache.Set(utils.CacheAccountActionPlans, acntID, nil, nil,
			cacheCommit(transactionID), transactionID)
		return nil, utils.ErrNotFound
	}
	apIDs = x.([]string)
	Cache.Set(utils.CacheAccountActionPlans, acntID, apIDs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) SetAccountActionPlansDrv(acntID string, apIDs []string, overwrite bool) (err error) {
	if !overwrite {
		if oldaPlIDs, err := iDB.GetAccountActionPlansDrv(acntID, true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return err
		} else {
			for _, oldAPid := range oldaPlIDs {
				if !utils.IsSliceMember(apIDs, oldAPid) {
					apIDs = append(apIDs, oldAPid)
				}
			}
		}
	}
	iDB.db.Set(utils.CacheAccountActionPlans, acntID, apIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemAccountActionPlansDrv(acntID string, apIDs []string) (err error) {
	if len(apIDs) == 0 {
		iDB.db.Remove(utils.CacheAccountActionPlans, acntID,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	oldaPlIDs, err := iDB.GetAccountActionPlansDrv(acntID, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	for i := 0; i < len(oldaPlIDs); {
		if utils.IsSliceMember(apIDs, oldaPlIDs[i]) {
			oldaPlIDs = append(oldaPlIDs[:i], oldaPlIDs[i+1:]...)
			continue
		}
		i++
	}
	if len(oldaPlIDs) == 0 {
		iDB.db.Remove(utils.CacheAccountActionPlans, acntID,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	iDB.db.Set(utils.CacheAccountActionPlans, acntID, oldaPlIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) PushTask(t *Task) (err error) {
	iDB.mu.Lock()
	iDB.tasks = append(iDB.tasks, t)
	iDB.mu.Unlock()
	return
}

func (iDB *InternalDB) PopTask() (t *Task, err error) {
	iDB.mu.Lock()
	if len(iDB.tasks) > 0 {
		t = iDB.tasks[0]
		iDB.tasks[0] = nil
		iDB.tasks = iDB.tasks[1:]
	} else {
		err = utils.ErrNotFound
	}
	iDB.mu.Unlock()
	return
}

func (iDB *InternalDB) GetAccountDrv(id string) (acc *Account, err error) {
	x, ok := iDB.db.Get(utils.CacheAccounts, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Account).Clone(), nil
}

func (iDB *InternalDB) SetAccountDrv(acc *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := iDB.GetAccountDrv(acc.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	acc.UpdateTime = time.Now()
	iDB.db.Set(utils.CacheAccounts, acc.ID, acc, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveAccountDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheAccounts, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceProfileDrv(tenant, id string) (rp *ResourceProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ResourceProfile), nil
}

func (iDB *InternalDB) SetResourceProfileDrv(rp *ResourceProfile) (err error) {
	iDB.db.Set(utils.CacheResourceProfiles, rp.TenantID(), rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	x, ok := iDB.db.Get(utils.CacheResources, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Resource), nil
}

func (iDB *InternalDB) SetResourceDrv(r *Resource) (err error) {
	iDB.db.Set(utils.CacheResources, r.TenantID(), r, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheResources, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetTimingDrv(id string) (tmg *utils.TPTiming, err error) {
	x, ok := iDB.db.Get(utils.CacheTimings, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*utils.TPTiming), nil
}

func (iDB *InternalDB) SetTimingDrv(timing *utils.TPTiming) (err error) {
	iDB.db.Set(utils.CacheTimings, timing.ID, timing, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveTimingDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheTimings, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, nil
}

func (iDB *InternalDB) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return nil
}

func (iDB *InternalDB) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	dbKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	x, ok := iDB.db.Get(cacheID, dbKey)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	if len(fldNameVal) != 0 {
		rcvidx := x.(map[string]utils.StringMap)
		indexes = make(map[string]utils.StringMap)
		for fldName, fldVal := range fldNameVal {
			if _, has := indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)]; !has {
				indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)] = make(utils.StringMap)
			}
			if len(rcvidx[utils.ConcatenatedKey(filterType, fldName, fldVal)]) != 0 {
				for key := range rcvidx[utils.ConcatenatedKey(filterType, fldName, fldVal)] {
					indexes[utils.ConcatenatedKey(filterType, fldName, fldVal)][key] = true
				}
			}
		}
		return
	} else {
		indexes = x.(map[string]utils.StringMap)
		if len(indexes) == 0 {
			return nil, utils.ErrNotFound
		}
	}
	return
}

func (iDB *InternalDB) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	originKey := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	dbKey := originKey
	if transactionID != "" {
		dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
	}
	if commit && transactionID != "" {
		x, _ := iDB.db.Get(cacheID, dbKey)
		iDB.db.Remove(cacheID, dbKey,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		iDB.db.Set(cacheID, originKey, x, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	var toBeDeleted []string
	toBeAdded := make(map[string]utils.StringMap)
	for key, strMp := range indexes {
		if len(strMp) == 0 { // remove with no more elements inside
			toBeDeleted = append(toBeDeleted, key)
			delete(indexes, key)
			continue
		}
		toBeAdded[key] = make(utils.StringMap)
		toBeAdded[key] = strMp
	}

	x, ok := iDB.db.Get(cacheID, dbKey)
	if !ok || x == nil {
		iDB.db.Set(cacheID, dbKey, toBeAdded, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return err
	}

	mp := x.(map[string]utils.StringMap)
	for _, key := range toBeDeleted {
		delete(mp, key)
	}
	for key, strMp := range toBeAdded {
		if _, has := mp[key]; !has {
			mp[key] = make(utils.StringMap)
		}
		mp[key] = strMp
	}
	iDB.db.Set(cacheID, dbKey, mp, nil,
		cacheCommit(transactionID), transactionID)
	return nil
}
func (iDB *InternalDB) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
	iDB.db.Remove(cacheID, utils.CacheInstanceToPrefix[cacheID]+itemIDPrefix,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) MatchFilterIndexDrv(cacheID, itemIDPrefix,
	filterType, fieldName, fieldVal string) (itemIDs utils.StringMap, err error) {

	x, ok := iDB.db.Get(cacheID, utils.CacheInstanceToPrefix[cacheID]+itemIDPrefix)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}

	indexes := x.(map[string]utils.StringMap)

	if _, hasIt := indexes[utils.ConcatenatedKey(filterType, fieldName, fieldVal)]; hasIt {
		itemIDs = indexes[utils.ConcatenatedKey(filterType, fieldName, fieldVal)]
	}
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

func (iDB *InternalDB) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueueProfile), nil

}
func (iDB *InternalDB) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	iDB.db.Set(utils.CacheStatQueueProfiles, sq.TenantID(), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemStatQueueProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	x, ok := iDB.db.Get(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueue), nil
}
func (iDB *InternalDB) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	if sq == nil {
		sq, err = ssq.AsStatQueue(iDB.ms)
		if err != nil {
			return
		}
	}
	iDB.db.Set(utils.CacheStatQueues, utils.ConcatenatedKey(sq.Tenant, sq.ID), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemStatQueueDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdProfileDrv(tenant, id string) (tp *ThresholdProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ThresholdProfile), nil
}

func (iDB *InternalDB) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	iDB.db.Set(utils.CacheThresholdProfiles, tp.TenantID(), tp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemThresholdProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdDrv(tenant, id string) (th *Threshold, err error) {
	x, ok := iDB.db.Get(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Threshold), nil
}

func (iDB *InternalDB) SetThresholdDrv(th *Threshold) (err error) {
	iDB.db.Set(utils.CacheThresholds, th.TenantID(), th, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveThresholdDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetFilterDrv(tenant, id string) (fltr *Filter, err error) {
	x, ok := iDB.db.Get(utils.CacheFilters, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Filter), nil

}

func (iDB *InternalDB) SetFilterDrv(fltr *Filter) (err error) {
	iDB.db.Set(utils.CacheFilters, fltr.TenantID(), fltr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveFilterDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetSupplierProfileDrv(tenant, id string) (spp *SupplierProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheSupplierProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*SupplierProfile), nil

}
func (iDB *InternalDB) SetSupplierProfileDrv(spp *SupplierProfile) (err error) {
	iDB.db.Set(utils.CacheSupplierProfiles, spp.TenantID(), spp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheSupplierProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetAttributeProfileDrv(tenant, id string) (attr *AttributeProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*AttributeProfile), nil
}
func (iDB *InternalDB) SetAttributeProfileDrv(attr *AttributeProfile) (err error) {
	iDB.db.Set(utils.CacheAttributeProfiles, attr.TenantID(), attr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetChargerProfileDrv(tenant, id string) (ch *ChargerProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ChargerProfile), nil
}
func (iDB *InternalDB) SetChargerProfileDrv(chr *ChargerProfile) (err error) {
	iDB.db.Set(utils.CacheChargerProfiles, chr.TenantID(), chr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveChargerProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetDispatcherProfileDrv(tenant, id string) (dpp *DispatcherProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherProfile), nil
}
func (iDB *InternalDB) SetDispatcherProfileDrv(dpp *DispatcherProfile) (err error) {
	iDB.db.Set(utils.CacheDispatcherProfiles, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	x, ok := iDB.db.Get(utils.CacheLoadIDs, utils.LoadIDs)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	loadIDs = x.(map[string]int64)
	if itemIDPrefix != utils.EmptyString {
		return map[string]int64{itemIDPrefix: loadIDs[itemIDPrefix]}, nil
	}
	return

}
func (iDB *InternalDB) SetLoadIDsDrv(loadIDs map[string]int64) (err error) {
	iDB.db.Set(utils.CacheLoadIDs, utils.LoadIDs, loadIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetDispatcherHostDrv(tenant, id string) (dpp *DispatcherHost, err error) {
	x, ok := iDB.db.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherHost), nil
}
func (iDB *InternalDB) SetDispatcherHostDrv(dpp *DispatcherHost) (err error) {
	iDB.db.Set(utils.CacheDispatcherHosts, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveLoadIDsDrv() (err error) {
	return utils.ErrNotImplemented
}
