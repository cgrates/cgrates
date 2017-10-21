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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewDataManager(dataDB DataDB) *DataManager {
	return &DataManager{dataDB: dataDB, cacheCfg: config.CgrConfig().CacheConfig}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB   DataDB
	cacheCfg config.CacheConfig
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	return dm.dataDB
}

func (dm *DataManager) LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aaPlIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rpIDs, resIDs []string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if dm.cacheCfg == nil {
			return
		}
		for k, cacheCfg := range dm.cacheCfg {
			k = utils.CacheInstanceToPrefix[k] // alias into prefixes understood by storage
			if utils.IsSliceMember([]string{utils.DESTINATION_PREFIX, utils.REVERSE_DESTINATION_PREFIX,
				utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX, utils.LCR_PREFIX, utils.CDR_STATS_PREFIX,
				utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACTION_TRIGGER_PREFIX,
				utils.SHARED_GROUP_PREFIX, utils.ALIASES_PREFIX, utils.REVERSE_ALIASES_PREFIX}, k) && cacheCfg.Precache {
				if err := dm.PreloadCacheForPrefix(k); err != nil && err != utils.ErrInvalidKey {
					return err
				}
			}
		}
		return
	} else {
		for key, ids := range map[string][]string{
			utils.DESTINATION_PREFIX:         dstIDs,
			utils.REVERSE_DESTINATION_PREFIX: rvDstIDs,
			utils.RATING_PLAN_PREFIX:         rplIDs,
			utils.RATING_PROFILE_PREFIX:      rpfIDs,
			utils.ACTION_PREFIX:              actIDs,
			utils.ACTION_PLAN_PREFIX:         aplIDs,
			utils.AccountActionPlansPrefix:   aaPlIDs,
			utils.ACTION_TRIGGER_PREFIX:      atrgIDs,
			utils.SHARED_GROUP_PREFIX:        sgIDs,
			utils.LCR_PREFIX:                 lcrIDs,
			utils.DERIVEDCHARGERS_PREFIX:     dcIDs,
			utils.ALIASES_PREFIX:             alsIDs,
			utils.REVERSE_ALIASES_PREFIX:     rvAlsIDs,
			utils.ResourceProfilesPrefix:     rpIDs,
			utils.ResourcesPrefix:            resIDs,
		} {
			if err = dm.CacheDataFromDB(key, ids, false); err != nil {
				return
			}
		}
	}
	return
}

//Used for MapStorage
func (dm *DataManager) PreloadCacheForPrefix(prefix string) error {
	transID := cache.BeginTransaction()
	cache.RemPrefixKey(prefix, false, transID)
	keyList, err := dm.DataDB().GetKeysForPrefix(prefix)
	if err != nil {
		cache.RollbackTransaction(transID)
		return err
	}
	switch prefix {
	case utils.RATING_PLAN_PREFIX:
		for _, key := range keyList {
			_, err := dm.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true, transID)
			if err != nil {
				cache.RollbackTransaction(transID)
				return err
			}
		}
	default:
		cache.RollbackTransaction(transID)
		return utils.ErrInvalidKey
	}
	cache.CommitTransaction(transID)
	return nil
}

func (dm *DataManager) CacheDataFromDB(prfx string, ids []string, mustBeCached bool) (err error) {
	if !utils.IsSliceMember([]string{utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.RATING_PLAN_PREFIX,
		utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.ACTION_TRIGGER_PREFIX,
		utils.SHARED_GROUP_PREFIX,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.LCR_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX,
		utils.ResourceProfilesPrefix,
		utils.TimingsPrefix,
		utils.ResourcesPrefix}, prfx) {
		return utils.NewCGRError(utils.MONGO,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if ids == nil {
		keyIDs, err := dm.DataDB().GetKeysForPrefix(prfx)
		if err != nil {
			return utils.NewCGRError(utils.MONGO,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("DataManager error <%s> querying keys for prefix: <%s>", err.Error(), prfx))
		}
		for _, keyID := range keyIDs {
			if mustBeCached { // Only consider loading ids which are already in cache
				if _, hasIt := cache.Get(keyID); !hasIt {
					continue
				}
			}
			ids = append(ids, keyID[len(prfx):])
		}
		var nrItems int
		if cCfg, has := dm.cacheCfg[utils.CachePrefixToInstance[prfx]]; has {
			nrItems = cCfg.Limit
		}
		if nrItems > 0 && nrItems < len(ids) { // More ids than cache config allows it, limit here
			ids = ids[:nrItems]
		}
	}
	for _, dataID := range ids {
		if mustBeCached {
			if _, hasIt := cache.Get(prfx + dataID); !hasIt { // only cache if previously there
				continue
			}
		}
		switch prfx {
		case utils.DESTINATION_PREFIX:
			_, err = dm.DataDB().GetDestination(dataID, true, utils.NonTransactional)
		case utils.REVERSE_DESTINATION_PREFIX:
			_, err = dm.DataDB().GetReverseDestination(dataID, true, utils.NonTransactional)
		case utils.RATING_PLAN_PREFIX:
			_, err = dm.GetRatingPlan(dataID, true, utils.NonTransactional)
		case utils.RATING_PROFILE_PREFIX:
			_, err = dm.GetRatingProfile(dataID, true, utils.NonTransactional)
		case utils.ACTION_PREFIX:
			_, err = dm.GetActions(dataID, true, utils.NonTransactional)
		case utils.ACTION_PLAN_PREFIX:
			_, err = dm.DataDB().GetActionPlan(dataID, true, utils.NonTransactional)
		case utils.AccountActionPlansPrefix:
			_, err = dm.DataDB().GetAccountActionPlans(dataID, true, utils.NonTransactional)
		case utils.ACTION_TRIGGER_PREFIX:
			_, err = dm.GetActionTriggers(dataID, true, utils.NonTransactional)
		case utils.SHARED_GROUP_PREFIX:
			_, err = dm.GetSharedGroup(dataID, true, utils.NonTransactional)
		case utils.DERIVEDCHARGERS_PREFIX:
			_, err = dm.GetDerivedChargers(dataID, true, utils.NonTransactional)
		case utils.LCR_PREFIX:
			_, err = dm.GetLCR(dataID, true, utils.NonTransactional)
		case utils.ALIASES_PREFIX:
			_, err = dm.DataDB().GetAlias(dataID, true, utils.NonTransactional)
		case utils.REVERSE_ALIASES_PREFIX:
			_, err = dm.DataDB().GetReverseAlias(dataID, true, utils.NonTransactional)
		case utils.ResourceProfilesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetResourceProfile(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.ResourcesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetResource(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.TimingsPrefix:
			_, err = dm.GetTiming(dataID, true, utils.NonTransactional)
		case utils.ThresholdProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetThresholdProfile(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.ThresholdPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetThreshold(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		case utils.FilterPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetFilter(tntID.Tenant, tntID.ID, true, utils.NonTransactional)
		}
		if err != nil {
			return utils.NewCGRError(utils.MONGO,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error <%s> querying mongo for category: <%s>, dataID: <%s>", err.Error(), prfx, dataID))
		}
	}
	return
}

// GetStatQueue retrieves a StatQueue from dataDB
// handles caching and deserialization of metrics
func (dm *DataManager) GetStatQueue(tenant, id string, skipCache bool, transactionID string) (sq *StatQueue, err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueue), nil
		}
	}
	ssq, err := dm.dataDB.GetStoredStatQueue(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if sq, err = ssq.AsStatQueue(dm.dataDB.Marshaler()); err != nil {
		return nil, err
	}
	cache.Set(key, sq, cacheCommit(transactionID), transactionID)
	return
}

// SetStatQueue converts to StoredStatQueue and stores the result in dataDB
func (dm *DataManager) SetStatQueue(sq *StatQueue) (err error) {
	ssq, err := NewStoredStatQueue(sq, dm.dataDB.Marshaler())
	if err != nil {
		return err
	}
	return dm.dataDB.SetStoredStatQueue(ssq)
}

// RemStatQueue removes the StoredStatQueue and clears the cache for StatQueue
func (dm *DataManager) RemStatQueue(tenant, id string, transactionID string) (err error) {
	if err = dm.dataDB.RemStoredStatQueue(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id), cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetFilter(tenant, id string, skipCache bool, transactionID string) (fltr *Filter, err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Filter), nil
		}
	}
	fltr, err = dm.dataDB.GetFilterDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, fltr, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetFilter(fltr *Filter) (err error) {
	return dm.DataDB().SetFilterDrv(fltr)
}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveFilterDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.FilterPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetThreshold(tenant, id string, skipCache bool, transactionID string) (th *Threshold, err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Threshold), nil
		}
	}
	th, err = dm.dataDB.GetThresholdDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, th, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetThreshold(th *Threshold) (err error) {
	return dm.DataDB().SetThresholdDrv(th)
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetThresholdProfile(tenant, id string, skipCache bool, transactionID string) (th *ThresholdProfile, err error) {
	key := utils.ThresholdProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ThresholdProfile), nil
		}
	}
	th, err = dm.dataDB.GetThresholdProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, th, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetThresholdProfile(th *ThresholdProfile) (err error) {
	return dm.DataDB().SetThresholdProfileDrv(th)
}

func (dm *DataManager) RemoveThresholdProfile(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemThresholdProfileDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.ThresholdProfilePrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetStatQueueProfile(tenant, id string, skipCache bool, transactionID string) (sqp *StatQueueProfile, err error) {
	key := utils.StatQueueProfilePrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueueProfile), nil
		}
	}
	sqp, err = dm.dataDB.GetStatQueueProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, sqp, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetStatQueueProfile(sqp *StatQueueProfile) (err error) {
	return dm.DataDB().SetStatQueueProfileDrv(sqp)
}

func (dm *DataManager) RemoveStatQueueProfile(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemStatQueueProfileDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.StatQueueProfilePrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetTiming(id string, skipCache bool, transactionID string) (t *utils.TPTiming, err error) {
	key := utils.TimingsPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.TPTiming), nil
		}
	}
	t, err = dm.dataDB.GetTimingDrv(id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, t, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetTiming(t *utils.TPTiming) (err error) {
	return dm.DataDB().SetTimingDrv(t)
}

func (dm *DataManager) RemoveTiming(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveTimingDrv(id); err != nil {
		return
	}
	cache.RemKey(utils.TimingsPrefix+id, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetResource(tenant, id string, skipCache bool, transactionID string) (rs *Resource, err error) {
	key := utils.ResourcesPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Resource), nil
		}
	}
	rs, err = dm.dataDB.GetResourceDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, rs, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetResource(rs *Resource) (err error) {
	return dm.DataDB().SetResourceDrv(rs)
}

func (dm *DataManager) RemoveResource(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveResourceDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.ResourcesPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetResourceProfile(tenant, id string, skipCache bool, transactionID string) (rp *ResourceProfile, err error) {
	key := utils.ResourceProfilesPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ResourceProfile), nil
		}
	}
	rp, err = dm.dataDB.GetResourceProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, rp, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetResourceProfile(rp *ResourceProfile) (err error) {
	return dm.DataDB().SetResourceProfileDrv(rp)
}

func (dm *DataManager) RemoveResourceProfile(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveResourceProfileDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.ResourceProfilesPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetActionTriggers(id string, skipCache bool, transactionID string) (attrs ActionTriggers, err error) {
	key := utils.ACTION_TRIGGER_PREFIX + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(ActionTriggers), nil
		}
	}
	attrs, err = dm.dataDB.GetActionTriggersDrv(id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, attrs, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) RemoveActionTriggers(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return
	}
	cache.RemKey(utils.ACTION_TRIGGER_PREFIX+id, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetLCR(id string, skipCache bool, transactionID string) (lcr *LCR, err error) {
	key := utils.LCR_PREFIX + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*LCR), nil
		}
	}
	lcr, err = dm.DataDB().GetLCRDrv(id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, lcr, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetActionTriggers(key string, attr ActionTriggers, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetActionTriggersDrv(key, attr); err != nil {
			cache.RemKey(utils.ACTION_TRIGGER_PREFIX+key, cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetActionTriggersDrv(key, attr)
	}
}

func (dm *DataManager) GetSharedGroup(key string, skipCache bool, transactionID string) (sg *SharedGroup, err error) {
	cachekey := utils.SHARED_GROUP_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cachekey); ok {
			if x != nil {
				return x.(*SharedGroup), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	sg, err = dm.DataDB().GetSharedGroupDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(cachekey, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(cachekey, sg, cacheCommit(transactionID), transactionID)
	return

}

func (dm *DataManager) SetSharedGroup(sg *SharedGroup, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetSharedGroupDrv(sg); err != nil {
			cache.RemKey(utils.SHARED_GROUP_PREFIX+sg.Id, cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetSharedGroupDrv(sg)
	}
}

func (dm *DataManager) SetLCR(lcr *LCR, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetLCRDrv(lcr); err != nil {
			cache.RemKey(utils.LCR_PREFIX+lcr.GetId(), cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetLCRDrv(lcr)
	}
}

func (dm *DataManager) GetDerivedChargers(key string, skipCache bool, transactionID string) (dcs *utils.DerivedChargers, err error) {
	cacheKey := utils.DERIVEDCHARGERS_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cacheKey); ok {
			if x != nil {
				return x.(*utils.DerivedChargers), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	dcs, err = dm.DataDB().GetDerivedChargersDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(cacheKey, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(cacheKey, dcs, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetActions(key string, skipCache bool, transactionID string) (as Actions, err error) {
	cachekey := utils.ACTION_PREFIX + key
	if !skipCache {
		if x, err := cache.GetCloned(cachekey); err != nil {
			if err.Error() != utils.ItemNotFound {
				return nil, err
			}
		} else if x == nil {
			return nil, utils.ErrNotFound
		} else {
			return x.(Actions), nil
		}
	}
	as, err = dm.DataDB().GetActionsDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(cachekey, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(cachekey, as, cacheCommit(transactionID), transactionID)
	return

}

func (dm *DataManager) SetActions(key string, as Actions, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetActionsDrv(key, as); err != nil {
			cache.RemKey(utils.ACTION_PREFIX+key, cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetActionsDrv(key, as)
	}
}

func (dm *DataManager) RemoveActions(key, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionsDrv(key); err != nil {
		return
	}
	cache.RemKey(utils.ACTION_PREFIX+key, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetRatingPlan(key string, skipCache bool, transactionID string) (rp *RatingPlan, err error) {
	cachekey := utils.RATING_PLAN_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cachekey); ok {
			if x != nil {
				return x.(*RatingPlan), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	rp, err = dm.DataDB().GetRatingPlanDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(cachekey, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(cachekey, rp, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetRatingPlanDrv(rp); err != nil {
			cache.RemKey(utils.RATING_PLAN_PREFIX+rp.Id, cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetRatingPlanDrv(rp)
	}
}

func (dm *DataManager) GetRatingProfile(key string, skipCache bool, transactionID string) (rpf *RatingProfile, err error) {
	cachekey := utils.RATING_PROFILE_PREFIX + key
	if !skipCache {
		if x, ok := cache.Get(cachekey); ok {
			if x != nil {
				return x.(*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	rpf, err = dm.DataDB().GetRatingProfileDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(cachekey, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(cachekey, rpf, cacheCommit(transactionID), transactionID)
	return

}

func (dm *DataManager) SetRatingProfile(rpf *RatingProfile, transactionID string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if err = dm.DataDB().SetRatingProfileDrv(rpf); err != nil {
			cache.RemKey(utils.RATING_PROFILE_PREFIX+rpf.Id, cacheCommit(transactionID), transactionID)
		}
		return
	} else {
		return dm.DataDB().SetRatingProfileDrv(rpf)
	}
}
