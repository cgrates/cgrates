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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func NewDataManager(dataDB DataDB) *DataManager {
	return &DataManager{dataDB: dataDB,
		cacheCfg: config.CgrConfig().CacheCfg()}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB   DataDB
	cacheCfg config.CacheCfg
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	if dm != nil {
		return dm.dataDB
	}
	return nil
}

func (dm *DataManager) LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs,
	aaPlIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rpIDs, resIDs,
	stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs, splPrflIDs, alsPrfIDs, cppIDs []string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if dm.cacheCfg == nil {
			return
		}
		for k, cacheCfg := range dm.cacheCfg {
			k = utils.CacheInstanceToPrefix[k] // alias into prefixes understood by storage
			if utils.IsSliceMember([]string{utils.DESTINATION_PREFIX, utils.REVERSE_DESTINATION_PREFIX,
				utils.RATING_PLAN_PREFIX, utils.RATING_PROFILE_PREFIX,
				utils.ACTION_PREFIX, utils.ACTION_PLAN_PREFIX, utils.ACTION_TRIGGER_PREFIX,
				utils.SHARED_GROUP_PREFIX, utils.ALIASES_PREFIX, utils.REVERSE_ALIASES_PREFIX, utils.StatQueuePrefix,
				utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
				utils.FilterPrefix, utils.SupplierProfilePrefix, utils.AttributeProfilePrefix, utils.ChargerProfilePrefix}, k) && cacheCfg.Precache {
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
			utils.DERIVEDCHARGERS_PREFIX:     dcIDs,
			utils.ALIASES_PREFIX:             alsIDs,
			utils.REVERSE_ALIASES_PREFIX:     rvAlsIDs,
			utils.ResourceProfilesPrefix:     rpIDs,
			utils.ResourcesPrefix:            resIDs,
			utils.StatQueuePrefix:            stqIDs,
			utils.StatQueueProfilePrefix:     stqpIDs,
			utils.ThresholdPrefix:            thIDs,
			utils.ThresholdProfilePrefix:     thpIDs,
			utils.FilterPrefix:               fltrIDs,
			utils.SupplierProfilePrefix:      splPrflIDs,
			utils.AttributeProfilePrefix:     alsPrfIDs,
			utils.ChargerProfilePrefix:       cppIDs,
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
	transID := Cache.BeginTransaction()
	Cache.Clear([]string{utils.CachePrefixToInstance[prefix]})
	keyList, err := dm.DataDB().GetKeysForPrefix(prefix)
	if err != nil {
		Cache.RollbackTransaction(transID)
		return err
	}
	switch prefix {
	case utils.RATING_PLAN_PREFIX:
		for _, key := range keyList {
			_, err := dm.GetRatingPlan(key[len(utils.RATING_PLAN_PREFIX):], true, transID)
			if err != nil {
				Cache.RollbackTransaction(transID)
				return err
			}
		}
	default:
		Cache.RollbackTransaction(transID)
		return utils.ErrInvalidKey
	}
	Cache.CommitTransaction(transID)
	return nil
}

func (dm *DataManager) CacheDataFromDB(prfx string, ids []string, mustBeCached bool) (err error) {
	if !utils.IsSliceMember([]string{
		utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.RATING_PLAN_PREFIX,
		utils.RATING_PROFILE_PREFIX,
		utils.ACTION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.ACTION_TRIGGER_PREFIX,
		utils.SHARED_GROUP_PREFIX,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX,
		utils.ResourceProfilesPrefix,
		utils.TimingsPrefix,
		utils.ResourcesPrefix,
		utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix,
		utils.ThresholdPrefix,
		utils.ThresholdProfilePrefix,
		utils.FilterPrefix,
		utils.SupplierProfilePrefix,
		utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix}, prfx) {
		return utils.NewCGRError(utils.DataManager,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if ids == nil {
		keyIDs, err := dm.DataDB().GetKeysForPrefix(prfx)
		if err != nil {
			return utils.NewCGRError(utils.DataManager,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("DataManager error <%s> querying keys for prefix: <%s>", err.Error(), prfx))
		}
		for _, keyID := range keyIDs {
			if mustBeCached { // Only consider loading ids which are already in cache
				if _, hasIt := Cache.Get(utils.CachePrefixToInstance[prfx], keyID[len(prfx):]); !hasIt {
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
			if _, hasIt := Cache.Get(utils.CachePrefixToInstance[prfx], dataID); !hasIt { // only cache if previously there
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
		case utils.ALIASES_PREFIX:
			_, err = dm.DataDB().GetAlias(dataID, true, utils.NonTransactional)
		case utils.REVERSE_ALIASES_PREFIX:
			_, err = dm.DataDB().GetReverseAlias(dataID, true, utils.NonTransactional)
		case utils.ResourceProfilesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetResourceProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ResourcesPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetResource(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.StatQueueProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetStatQueueProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.StatQueuePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetStatQueue(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.TimingsPrefix:
			_, err = dm.GetTiming(dataID, true, utils.NonTransactional)
		case utils.ThresholdProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetThresholdProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ThresholdPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetThreshold(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.FilterPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetFilter(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.SupplierProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetSupplierProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.AttributeProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetAttributeProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ChargerProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetChargerProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		}
		if err != nil {
			return utils.NewCGRError(utils.DataManager,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error <%s> querying DataManager for category: <%s>, dataID: <%s>", err.Error(), prfx, dataID))
		}
	}
	return
}

// GetStatQueue retrieves a StatQueue from dataDB
// handles caching and deserialization of metrics
func (dm *DataManager) GetStatQueue(tenant, id string,
	cacheRead, cacheWrite bool, transactionID string) (sq *StatQueue, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheStatQueues, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueue), nil
		}
	}
	ssq, err := dm.dataDB.GetStoredStatQueueDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheStatQueues, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if sq, err = ssq.AsStatQueue(dm.dataDB.Marshaler()); err != nil {
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheStatQueues, tntID, sq, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

// SetStatQueue converts to StoredStatQueue and stores the result in dataDB
func (dm *DataManager) SetStatQueue(sq *StatQueue) (err error) {
	ssq, err := NewStoredStatQueue(sq, dm.dataDB.Marshaler())
	if err != nil {
		return err
	}
	if err = dm.dataDB.SetStoredStatQueueDrv(ssq); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.StatQueuePrefix, []string{sq.TenantID()}, true)
}

// RemoveStatQueue removes the StoredStatQueue and clears the cache for StatQueue
func (dm *DataManager) RemoveStatQueue(tenant, id string, transactionID string) (err error) {
	if err = dm.dataDB.RemStoredStatQueueDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

// GetFilter returns
func (dm *DataManager) GetFilter(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (fltr *Filter, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheFilters, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Filter), nil
		}
	}
	if strings.HasPrefix(id, utils.Meta) {
		fltr, err = NewFilterFromInline(tenant, id)
	} else {
		fltr, err = dm.DataDB().GetFilterDrv(tenant, id)
	}
	if err != nil {
		if err == utils.ErrNotFound {
			if cacheWrite {
				Cache.Set(utils.CacheFilters, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheFilters, tntID, fltr, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetFilter(fltr *Filter) (err error) {
	if err = dm.DataDB().SetFilterDrv(fltr); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.FilterPrefix, []string{fltr.TenantID()}, true)
}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveFilterDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetThreshold(tenant, id string,
	cacheRead, cacheWrite bool, transactionID string) (th *Threshold, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheThresholds, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Threshold), nil
		}
	}
	th, err = dm.dataDB.GetThresholdDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheThresholds, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheThresholds, tntID, th, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetThreshold(th *Threshold) (err error) {
	if err = dm.DataDB().SetThresholdDrv(th); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.ThresholdPrefix, []string{th.TenantID()}, true); err != nil {
		return
	}
	return
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetThresholdProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (th *ThresholdProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheThresholdProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ThresholdProfile), nil
		}
	}
	th, err = dm.dataDB.GetThresholdProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheThresholdProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheThresholdProfiles, tntID, th, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetThresholdProfile(th *ThresholdProfile, withIndex bool) (err error) {
	oldTh, err := dm.GetThresholdProfile(th.Tenant, th.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetThresholdProfileDrv(th); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.ThresholdProfilePrefix,
		[]string{th.TenantID()}, true); err != nil {
		return
	}
	if withIndex {
		if oldTh != nil {
			var needsRemove bool
			for _, fltrID := range oldTh.FilterIDs {
				if !utils.IsSliceMember(th.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.ThresholdProfilePrefix,
					th.Tenant).RemoveItemFromIndex(th.Tenant, th.ID, oldTh.FilterIDs); err != nil {
					return
				}
			}
		}
		return createAndIndex(utils.ThresholdProfilePrefix, th.Tenant, utils.EmptyString, th.ID, th.FilterIDs, dm)
	}
	return
}

func (dm *DataManager) RemoveThresholdProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	oldTh, err := dm.GetThresholdProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemThresholdProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		return NewFilterIndexer(dm,
			utils.ThresholdProfilePrefix, tenant).RemoveItemFromIndex(tenant, id, oldTh.FilterIDs)
	}
	return
}

func (dm *DataManager) GetStatQueueProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (sqp *StatQueueProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheStatQueueProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueueProfile), nil
		}
	}
	sqp, err = dm.dataDB.GetStatQueueProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheStatQueueProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheStatQueueProfiles, tntID, sqp, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetStatQueueProfile(sqp *StatQueueProfile, withIndex bool) (err error) {
	oldSts, err := dm.GetStatQueueProfile(sqp.Tenant, sqp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetStatQueueProfileDrv(sqp); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.StatQueueProfilePrefix,
		[]string{sqp.TenantID()}, true); err != nil {
		return
	}
	if withIndex {
		if oldSts != nil {
			var needsRemove bool
			for _, fltrID := range oldSts.FilterIDs {
				if !utils.IsSliceMember(sqp.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.StatQueueProfilePrefix,
					sqp.Tenant).RemoveItemFromIndex(sqp.Tenant, sqp.ID, oldSts.FilterIDs); err != nil {
					return
				}
			}
		}
		return createAndIndex(utils.StatQueueProfilePrefix, sqp.Tenant, utils.EmptyString, sqp.ID, sqp.FilterIDs, dm)
	}
	return
}

func (dm *DataManager) RemoveStatQueueProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	oldSts, err := dm.GetStatQueueProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemStatQueueProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		return NewFilterIndexer(dm, utils.StatQueueProfilePrefix, tenant).RemoveItemFromIndex(tenant, id, oldSts.FilterIDs)
	}
	return
}

func (dm *DataManager) GetTiming(id string, skipCache bool,
	transactionID string) (t *utils.TPTiming, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheTimings, id); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.TPTiming), nil
		}
	}
	t, err = dm.dataDB.GetTimingDrv(id)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheTimings, id, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheTimings, id, t, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetTiming(t *utils.TPTiming) (err error) {
	if err = dm.DataDB().SetTimingDrv(t); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.TimingsPrefix, []string{t.ID}, true)
}

func (dm *DataManager) RemoveTiming(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveTimingDrv(id); err != nil {
		return
	}
	Cache.Remove(utils.CacheTimings, id,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetResource(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rs *Resource, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheResources, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Resource), nil
		}
	}
	rs, err = dm.dataDB.GetResourceDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheResources, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheResources, tntID, rs, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetResource(rs *Resource) (err error) {
	if err = dm.DataDB().SetResourceDrv(rs); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(
		utils.CacheInstanceToPrefix[utils.CacheResources],
		[]string{rs.TenantID()}, true); err != nil {
		return
	}
	return
}

func (dm *DataManager) RemoveResource(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveResourceDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheResources, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetResourceProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rp *ResourceProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheResourceProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ResourceProfile), nil
		}
	}
	rp, err = dm.dataDB.GetResourceProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheResourceProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheResourceProfiles, tntID, rp, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetResourceProfile(rp *ResourceProfile, withIndex bool) (err error) {
	oldRes, err := dm.GetResourceProfile(rp.Tenant, rp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetResourceProfileDrv(rp); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.ResourceProfilesPrefix,
		[]string{rp.TenantID()}, true); err != nil {
		return
	}
	//to be implemented in tests
	if withIndex {
		if oldRes != nil {
			var needsRemove bool
			for _, fltrID := range oldRes.FilterIDs {
				if !utils.IsSliceMember(rp.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.ResourceProfilesPrefix,
					rp.Tenant).RemoveItemFromIndex(rp.Tenant, rp.ID, oldRes.FilterIDs); err != nil {
					return
				}
			}
		}
		if err = createAndIndex(utils.ResourceProfilesPrefix, rp.Tenant, utils.EmptyString, rp.ID, rp.FilterIDs, dm); err != nil {
			return
		}
		Cache.Clear([]string{utils.CacheEventResources})
	}
	return
}

func (dm *DataManager) RemoveResourceProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	oldRes, err := dm.GetResourceProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveResourceProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		return NewFilterIndexer(dm, utils.ResourceProfilesPrefix, tenant).RemoveItemFromIndex(tenant, id, oldRes.FilterIDs)
	}
	return
}

func (dm *DataManager) GetActionTriggers(id string, skipCache bool,
	transactionID string) (attrs ActionTriggers, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheActionTriggers, id); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(ActionTriggers), nil
		}
	}
	attrs, err = dm.dataDB.GetActionTriggersDrv(id)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheActionTriggers, id, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheActionTriggers, id, attrs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) RemoveActionTriggers(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return
	}
	Cache.Remove(utils.CacheActionTriggers, id,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetActionTriggers(key string, attr ActionTriggers,
	transactionID string) (err error) {
	if err = dm.DataDB().SetActionTriggersDrv(key, attr); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.ACTION_TRIGGER_PREFIX, []string{key}, true)
}

func (dm *DataManager) GetSharedGroup(key string, skipCache bool,
	transactionID string) (sg *SharedGroup, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheSharedGroups, key); ok {
			if x != nil {
				return x.(*SharedGroup), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	sg, err = dm.DataDB().GetSharedGroupDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheSharedGroups, key, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheSharedGroups, key, sg, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetSharedGroup(sg *SharedGroup,
	transactionID string) (err error) {
	if err = dm.DataDB().SetSharedGroupDrv(sg); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.SHARED_GROUP_PREFIX,
		[]string{sg.Id}, true)
}

func (dm *DataManager) RemoveSharedGroup(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveSharedGroupDrv(id, transactionID); err != nil {
		return
	}
	Cache.Remove(utils.CacheSharedGroups, id,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetDerivedChargers(key string, skipCache bool,
	transactionID string) (dcs *utils.DerivedChargers, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheDerivedChargers, key); ok {
			if x != nil {
				return x.(*utils.DerivedChargers), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	dcs, err = dm.DataDB().GetDerivedChargersDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheDerivedChargers, key, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheDerivedChargers, key, dcs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) RemoveDerivedChargers(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveDerivedChargersDrv(id, transactionID); err != nil {
		return
	}
	Cache.Remove(utils.CacheDerivedChargers, id,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetActions(key string, skipCache bool, transactionID string) (as Actions, err error) {
	if !skipCache {
		if x, err := Cache.GetCloned(utils.CacheActions, key); err != nil {
			if err != ltcache.ErrNotFound {
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
			Cache.Set(utils.CacheActions, key, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheActions, key, as, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetActions(key string, as Actions, transactionID string) (err error) {
	if err = dm.DataDB().SetActionsDrv(key, as); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.ACTION_PREFIX, []string{key}, true)
}

func (dm *DataManager) RemoveActions(key, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionsDrv(key); err != nil {
		return
	}
	Cache.Remove(utils.CacheActions, key,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetRatingPlan(key string, skipCache bool,
	transactionID string) (rp *RatingPlan, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheRatingPlans, key); ok {
			if x != nil {
				return x.(*RatingPlan), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	rp, err = dm.DataDB().GetRatingPlanDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheRatingPlans, key, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheRatingPlans, key, rp, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	if err = dm.DataDB().SetRatingPlanDrv(rp); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.RATING_PLAN_PREFIX, []string{rp.Id}, true)
}

func (dm *DataManager) RemoveRatingPlan(key string, transactionID string) (err error) {
	if err = dm.DataDB().RemoveRatingPlanDrv(key); err != nil {
		return
	}
	Cache.Remove(utils.CacheRatingPlans, key,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetRatingProfile(key string, skipCache bool,
	transactionID string) (rpf *RatingProfile, err error) {
	if !skipCache {
		if x, ok := Cache.Get(utils.CacheRatingProfiles, key); ok {
			if x != nil {
				return x.(*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	rpf, err = dm.DataDB().GetRatingProfileDrv(key)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(utils.CacheRatingProfiles, key, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	Cache.Set(utils.CacheRatingProfiles, key, rpf, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetRatingProfile(rpf *RatingProfile,
	transactionID string) (err error) {
	if err = dm.DataDB().SetRatingProfileDrv(rpf); err != nil {
		return
	}
	return dm.CacheDataFromDB(utils.RATING_PROFILE_PREFIX, []string{rpf.Id}, true)
}

func (dm *DataManager) RemoveRatingProfile(key string,
	transactionID string) (err error) {
	if err = dm.DataDB().RemoveRatingProfileDrv(key); err != nil {
		return
	}
	Cache.Remove(utils.CacheRatingProfiles, key,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetUser(up *UserProfile) (err error) {
	return dm.DataDB().SetUserDrv(up)
}

func (dm *DataManager) GetUser(key string) (up *UserProfile, err error) {
	return dm.DataDB().GetUserDrv(key)
}

func (dm *DataManager) GetUsers() (result []*UserProfile, err error) {
	return dm.DataDB().GetUsersDrv()
}

func (dm *DataManager) RemoveUser(key string) error {
	return dm.DataDB().RemoveUserDrv(key)
}

func (dm *DataManager) GetSubscribers() (result map[string]*SubscriberData, err error) {
	return dm.DataDB().GetSubscribersDrv()
}

func (dm *DataManager) SetSubscriber(key string, sub *SubscriberData) (err error) {
	return dm.DataDB().SetSubscriberDrv(key, sub)
}

func (dm *DataManager) RemoveSubscriber(key string) (err error) {
	return dm.DataDB().RemoveSubscriberDrv(key)
}

func (dm *DataManager) HasData(category, subject, tenant string) (has bool, err error) {
	return dm.DataDB().HasDataDrv(category, subject, tenant)
}

func (dm *DataManager) GetFilterIndexes(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	return dm.DataDB().GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType, fldNameVal)
}

func (dm *DataManager) SetFilterIndexes(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	return dm.DataDB().SetFilterIndexesDrv(cacheID, itemIDPrefix, indexes, commit, transactionID)
}

func (dm *DataManager) RemoveFilterIndexes(cacheID, itemIDPrefix string) (err error) {
	return dm.DataDB().RemoveFilterIndexesDrv(cacheID, itemIDPrefix)
}

func (dm *DataManager) MatchFilterIndex(cacheID, itemIDPrefix,
	filterType, fieldName, fieldVal string) (itemIDs utils.StringMap, err error) {
	fieldValKey := utils.ConcatenatedKey(itemIDPrefix, filterType, fieldName, fieldVal)
	if x, ok := Cache.Get(cacheID, fieldValKey); ok { // Attempt to find in cache first
		if x == nil {
			return nil, utils.ErrNotFound
		}
		return x.(utils.StringMap), nil
	}
	// Not found in cache, check in DB
	itemIDs, err = dm.DataDB().MatchFilterIndexDrv(cacheID, itemIDPrefix, filterType, fieldName, fieldVal)
	if err != nil {
		if err == utils.ErrNotFound {
			Cache.Set(cacheID, fieldValKey, nil, nil,
				true, utils.NonTransactional)
		}
		return nil, err
	}
	Cache.Set(cacheID, fieldValKey, itemIDs, nil,
		true, utils.NonTransactional)
	return
}

func (dm *DataManager) GetSupplierProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (supp *SupplierProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheSupplierProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*SupplierProfile), nil
		}
	}
	supp, err = dm.dataDB.GetSupplierProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheSupplierProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheSupplierProfiles, tntID, supp, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetSupplierProfile(supp *SupplierProfile, withIndex bool) (err error) {
	oldSup, err := dm.GetSupplierProfile(supp.Tenant, supp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetSupplierProfileDrv(supp); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.SupplierProfilePrefix, []string{supp.TenantID()}, true); err != nil {
		return
	}
	if withIndex {
		if oldSup != nil {
			var needsRemove bool
			for _, fltrID := range oldSup.FilterIDs {
				if !utils.IsSliceMember(supp.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.SupplierProfilePrefix,
					supp.Tenant).RemoveItemFromIndex(supp.Tenant, supp.ID, oldSup.FilterIDs); err != nil {
					return
				}
			}
		}
		return createAndIndex(utils.SupplierProfilePrefix, supp.Tenant, utils.EmptyString, supp.ID, supp.FilterIDs, dm)
	}
	return
}

func (dm *DataManager) RemoveSupplierProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	oldSupp, err := dm.GetSupplierProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveSupplierProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheSupplierProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		return NewFilterIndexer(dm, utils.SupplierProfilePrefix, tenant).RemoveItemFromIndex(tenant, id, oldSupp.FilterIDs)
	}
	return
}

func (dm *DataManager) GetAttributeProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (attrPrfl *AttributeProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheAttributeProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*AttributeProfile), nil
		}
	}
	attrPrfl, err = dm.dataDB.GetAttributeProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheAttributeProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if err = attrPrfl.Compile(); err != nil {
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheAttributeProfiles, tntID, attrPrfl, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetAttributeProfile(ap *AttributeProfile, withIndex bool) (err error) {
	oldAP, err := dm.GetAttributeProfile(ap.Tenant, ap.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetAttributeProfileDrv(ap); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.AttributeProfilePrefix, []string{ap.TenantID()}, true); err != nil {
		return
	}
	if withIndex {
		if oldAP != nil {
			for _, ctx := range oldAP.Contexts {
				var needsRemove bool
				if !utils.IsSliceMember(ap.Contexts, ctx) {
					needsRemove = true
				} else {
					for _, fltrID := range oldAP.FilterIDs {
						if !utils.IsSliceMember(ap.FilterIDs, fltrID) {
							needsRemove = true
						}
					}
				}
				if needsRemove {
					if err = NewFilterIndexer(dm, utils.AttributeProfilePrefix,
						utils.ConcatenatedKey(ap.Tenant, ctx)).RemoveItemFromIndex(ap.Tenant, ap.ID, oldAP.FilterIDs); err != nil {
						return
					}
				}
			}
		}
		for _, ctx := range ap.Contexts {
			if err = createAndIndex(utils.AttributeProfilePrefix,
				ap.Tenant, ctx, ap.ID, ap.FilterIDs, dm); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveAttributeProfile(tenant, id string, transactionID string, withIndex bool) (err error) {
	oldAttr, err := dm.GetAttributeProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil {
		return err
	}
	if err = dm.DataDB().RemoveAttributeProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		for _, context := range oldAttr.Contexts {
			if err = NewFilterIndexer(dm, utils.AttributeProfilePrefix,
				utils.ConcatenatedKey(tenant, context)).RemoveItemFromIndex(tenant, id, oldAttr.FilterIDs); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) GetChargerProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (cpp *ChargerProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheChargerProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ChargerProfile), nil
		}
	}
	cpp, err = dm.dataDB.GetChargerProfileDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound && cacheWrite {
			Cache.Set(utils.CacheChargerProfiles, tntID, nil, nil,
				cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if cacheWrite {
		Cache.Set(utils.CacheChargerProfiles, tntID, cpp, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetChargerProfile(cpp *ChargerProfile, withIndex bool) (err error) {
	oldCpp, err := dm.GetChargerProfile(cpp.Tenant, cpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetChargerProfileDrv(cpp); err != nil {
		return err
	}
	if err = dm.CacheDataFromDB(utils.ChargerProfilePrefix, []string{cpp.TenantID()}, true); err != nil {
		return
	}
	if withIndex {
		if oldCpp != nil {
			var needsRemove bool
			for _, fltrID := range oldCpp.FilterIDs {
				if !utils.IsSliceMember(cpp.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.SupplierProfilePrefix,
					cpp.Tenant).RemoveItemFromIndex(cpp.Tenant, cpp.ID, oldCpp.FilterIDs); err != nil {
					return
				}
			}
		}
		return createAndIndex(utils.ChargerProfilePrefix, cpp.Tenant, utils.EmptyString, cpp.ID, cpp.FilterIDs, dm)
	}
	return
}

func (dm *DataManager) RemoveChargerProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	oldCpp, err := dm.GetChargerProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveChargerProfileDrv(tenant, id); err != nil {
		return
	}
	Cache.Remove(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	if withIndex {
		return NewFilterIndexer(dm, utils.ChargerProfilePrefix, tenant).RemoveItemFromIndex(tenant, id, oldCpp.FilterIDs)
	}
	return
}
