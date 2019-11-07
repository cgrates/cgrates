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
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

var (
	filterIndexesPrefixMap = utils.StringMap{
		utils.AttributeFilterIndexes:  true,
		utils.ResourceFilterIndexes:   true,
		utils.StatFilterIndexes:       true,
		utils.ThresholdFilterIndexes:  true,
		utils.SupplierFilterIndexes:   true,
		utils.ChargerFilterIndexes:    true,
		utils.DispatcherFilterIndexes: true,
	}
	loadCachePrefixMap = utils.StringMap{
		utils.DESTINATION_PREFIX:         true,
		utils.REVERSE_DESTINATION_PREFIX: true,
		utils.RATING_PLAN_PREFIX:         true,
		utils.RATING_PROFILE_PREFIX:      true,
		utils.ACTION_PREFIX:              true,
		utils.ACTION_PLAN_PREFIX:         true,
		utils.ACTION_TRIGGER_PREFIX:      true,
		utils.SHARED_GROUP_PREFIX:        true,
		utils.StatQueuePrefix:            true,
		utils.StatQueueProfilePrefix:     true,
		utils.ThresholdPrefix:            true,
		utils.ThresholdProfilePrefix:     true,
		utils.FilterPrefix:               true,
		utils.SupplierProfilePrefix:      true,
		utils.AttributeProfilePrefix:     true,
		utils.ChargerProfilePrefix:       true,
		utils.DispatcherProfilePrefix:    true,
		utils.DispatcherHostPrefix:       true,
	}
	cachePrefixMap = utils.StringMap{
		utils.DESTINATION_PREFIX:         true,
		utils.REVERSE_DESTINATION_PREFIX: true,
		utils.RATING_PLAN_PREFIX:         true,
		utils.RATING_PROFILE_PREFIX:      true,
		utils.ACTION_PREFIX:              true,
		utils.ACTION_PLAN_PREFIX:         true,
		utils.AccountActionPlansPrefix:   true,
		utils.ACTION_TRIGGER_PREFIX:      true,
		utils.SHARED_GROUP_PREFIX:        true,
		utils.ResourceProfilesPrefix:     true,
		utils.TimingsPrefix:              true,
		utils.ResourcesPrefix:            true,
		utils.StatQueuePrefix:            true,
		utils.StatQueueProfilePrefix:     true,
		utils.ThresholdPrefix:            true,
		utils.ThresholdProfilePrefix:     true,
		utils.FilterPrefix:               true,
		utils.SupplierProfilePrefix:      true,
		utils.AttributeProfilePrefix:     true,
		utils.ChargerProfilePrefix:       true,
		utils.DispatcherProfilePrefix:    true,
		utils.DispatcherHostPrefix:       true,
		utils.AttributeFilterIndexes:     true,
		utils.ResourceFilterIndexes:      true,
		utils.StatFilterIndexes:          true,
		utils.ThresholdFilterIndexes:     true,
		utils.SupplierFilterIndexes:      true,
		utils.ChargerFilterIndexes:       true,
		utils.DispatcherFilterIndexes:    true,
	}
)

// NewDataManager returns a new DataManager
func NewDataManager(dataDB DataDB, cacheCfg config.CacheCfg, rmtDataDBs, rplDataDBs []*DataManager) *DataManager {
	return &DataManager{
		dataDB:     dataDB,
		cacheCfg:   cacheCfg,
		rmtDataDBs: rmtDataDBs,
		rplDataDBs: rplDataDBs,
	}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB     DataDB
	rmtDataDBs []*DataManager
	rplDataDBs []*DataManager
	cacheCfg   config.CacheCfg
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	if dm != nil {
		return dm.dataDB
	}
	return nil
}

func (dm *DataManager) LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs,
	aaPlIDs, atrgIDs, sgIDs, rpIDs, resIDs, stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs,
	splPrflIDs, alsPrfIDs, cppIDs, dppIDs, dphIDs []string) (err error) {
	if dm.DataDB().GetStorageType() == utils.MAPSTOR {
		if dm.cacheCfg == nil {
			return
		}
		for k, cacheCfg := range dm.cacheCfg {
			k = utils.CacheInstanceToPrefix[k] // alias into prefixes understood by storage
			if loadCachePrefixMap.HasKey(k) && cacheCfg.Precache {
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
			utils.DispatcherProfilePrefix:    dppIDs,
			utils.DispatcherHostPrefix:       dphIDs,
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
	if !cachePrefixMap.HasKey(prfx) {
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
		case utils.DispatcherProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetDispatcherProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.DispatcherHostPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetDispatcherHost(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.AttributeFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheAttributeFilterIndexes, dataID)
		case utils.ResourceFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheResourceFilterIndexes, dataID)
		case utils.StatFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheStatFilterIndexes, dataID)
		case utils.ThresholdFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheThresholdFilterIndexes, dataID)
		case utils.SupplierFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheSupplierFilterIndexes, dataID)
		case utils.ChargerFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheChargerFilterIndexes, dataID)
		case utils.DispatcherFilterIndexes:
			err = dm.MatchFilterIndexFromKey(utils.CacheDispatcherFilterIndexes, dataID)
		case utils.LoadIDPrefix:
			_, err = dm.GetItemLoadIDs(utils.EmptyString, true)
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Remove(utils.CachePrefixToInstance[prfx], dataID,
					cacheCommit(utils.NonTransactional), utils.NonTransactional)
				err = nil
			} else {
				return utils.NewCGRError(utils.DataManager,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("error <%s> querying DataManager for category: <%s>, dataID: <%s>", err.Error(), prfx, dataID))
			}
		}
	}
	return
}

func (dm *DataManager) GetAccount(id string) (acc *Account, err error) {
	acc, err = dm.dataDB.GetAccountDrv(id)
	if err != nil {
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if acc, rmtErr = rmtDM.dataDB.GetAccountDrv(id); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			return nil, err
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if ssq, rmtErr = rmtDM.dataDB.GetStoredStatQueueDrv(tenant, id); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheStatQueues, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.dataDB.SetStoredStatQueueDrv(ssq); err != nil {
				return
			}
		}
	}
	return
}

// RemoveStatQueue removes the StoredStatQueue
func (dm *DataManager) RemoveStatQueue(tenant, id string, transactionID string) (err error) {
	if err = dm.dataDB.RemStoredStatQueueDrv(tenant, id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveStatQueue(tenant, id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if fltr, rmtErr = rmtDM.GetFilter(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheFilters, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}

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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetFilter(fltr); err != nil {
				return
			}
		}
	}
	return

}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveFilterDrv(tenant, id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveFilter(tenant, id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if th, rmtErr = rmtDM.GetThreshold(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheThresholds, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetThreshold(th); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveThreshold(tenant, id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if th, rmtErr = rmtDM.GetThresholdProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheThresholdProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
		if err := createAndIndex(utils.ThresholdProfilePrefix, th.Tenant,
			utils.EmptyString, th.ID, th.FilterIDs, dm); err != nil {
			return err
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetThresholdProfile(th, withIndex); err != nil {
				return
			}
		}
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
	if oldTh == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = NewFilterIndexer(dm, utils.ThresholdProfilePrefix,
			tenant).RemoveItemFromIndex(tenant, id, oldTh.FilterIDs); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveThresholdProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if sqp, rmtErr = rmtDM.GetStatQueueProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheStatQueueProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
		if err = createAndIndex(utils.StatQueueProfilePrefix, sqp.Tenant,
			utils.EmptyString, sqp.ID, sqp.FilterIDs, dm); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetStatQueueProfile(sqp, withIndex); err != nil {
				return
			}
		}
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
	if oldSts == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = NewFilterIndexer(dm, utils.StatQueueProfilePrefix,
			tenant).RemoveItemFromIndex(tenant, id, oldSts.FilterIDs); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveStatQueueProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if t, rmtErr = rmtDM.GetTiming(id, false,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheTimings, id, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
	}
	Cache.Set(utils.CacheTimings, id, t, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetTiming(t *utils.TPTiming) (err error) {
	if err = dm.DataDB().SetTimingDrv(t); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.TimingsPrefix, []string{t.ID}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetTimingDrv(t); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveTiming(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveTimingDrv(id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveTiming(id,
				utils.NonTransactional); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if rs, rmtErr = rmtDM.GetResource(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheResources, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetResource(rs); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveResource(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveResourceDrv(tenant, id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveResource(tenant, id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if rp, rmtErr = rmtDM.GetResourceProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheResourceProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetResourceProfile(rp, withIndex); err != nil {
				return
			}
		}
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
	if oldRes == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = NewFilterIndexer(dm, utils.ResourceProfilesPrefix,
			tenant).RemoveItemFromIndex(tenant, id, oldRes.FilterIDs); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveResourceProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if attrs, rmtErr = rmtDM.GetActionTriggers(id, true,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheActionTriggers, id, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
			return nil, err
		}
	}
	Cache.Set(utils.CacheActionTriggers, id, attrs, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) RemoveActionTriggers(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveActionTriggers(id,
				utils.NonTransactional); err != nil {
				return
			}
		}
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
	if err = dm.CacheDataFromDB(utils.ACTION_TRIGGER_PREFIX, []string{key}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetActionTriggersDrv(key, attr); err != nil {
				return
			}
		}
	}
	return
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if sg, rmtErr = rmtDM.GetSharedGroup(key, true,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheSharedGroups, key, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
			return nil, err
		}
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
	if err = dm.CacheDataFromDB(utils.SHARED_GROUP_PREFIX,
		[]string{sg.Id}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetSharedGroupDrv(sg); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveSharedGroup(id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveSharedGroupDrv(id); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveSharedGroup(id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
	Cache.Remove(utils.CacheSharedGroups, id,
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if as, rmtErr = rmtDM.GetActions(key, true,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheActions, key, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
			return nil, err
		}
	}
	Cache.Set(utils.CacheActions, key, as, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetActions(key string, as Actions, transactionID string) (err error) {
	if err = dm.DataDB().SetActionsDrv(key, as); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.ACTION_PREFIX, []string{key}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetActionsDrv(key, as); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveActions(key, transactionID string) (err error) {
	if err = dm.DataDB().RemoveActionsDrv(key); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveActions(key,
				utils.NonTransactional); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if rp, rmtErr = rmtDM.GetRatingPlan(key, true,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheRatingPlans, key, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
			return nil, err
		}
	}
	Cache.Set(utils.CacheRatingPlans, key, rp, nil,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	if err = dm.DataDB().SetRatingPlanDrv(rp); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.RATING_PLAN_PREFIX, []string{rp.Id}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetRatingPlanDrv(rp); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveRatingPlan(key string, transactionID string) (err error) {
	if err = dm.DataDB().RemoveRatingPlanDrv(key); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveRatingPlan(key,
				utils.NonTransactional); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if rpf, rmtErr = rmtDM.GetRatingProfile(key, true,
					utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(utils.CacheRatingProfiles, key, nil, nil,
					cacheCommit(transactionID), transactionID)
			}
			return nil, err
		}
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
	if err = dm.CacheDataFromDB(utils.RATING_PROFILE_PREFIX, []string{rpf.Id}, true); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetRatingProfileDrv(rpf); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveRatingProfile(key string,
	transactionID string) (err error) {
	if err = dm.DataDB().RemoveRatingProfileDrv(key); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveRatingProfile(key,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
	Cache.Remove(utils.CacheRatingProfiles, key,
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) HasData(category, subject, tenant string) (has bool, err error) {
	return dm.DataDB().HasDataDrv(category, subject, tenant)
}

func (dm *DataManager) GetFilterIndexes(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	if indexes, err = dm.DataDB().GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType, fldNameVal); err != nil {
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if indexes, rmtErr = rmtDM.GetFilterIndexes(cacheID, itemIDPrefix,
					filterType, fldNameVal); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetFilterIndexes(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	if err = dm.DataDB().SetFilterIndexesDrv(cacheID, itemIDPrefix,
		indexes, commit, transactionID); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.DataDB().SetFilterIndexesDrv(cacheID, itemIDPrefix,
				indexes, commit, transactionID); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveFilterIndexes(cacheID, itemIDPrefix string) (err error) {
	if err = dm.DataDB().RemoveFilterIndexesDrv(cacheID, itemIDPrefix); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveFilterIndexes(cacheID,
				itemIDPrefix); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) MatchFilterIndexFromKey(cacheID, key string) (err error) {
	splt := utils.SplitConcatenatedKey(key) // prefix:filterType:fieldName:fieldVal
	lsplt := len(splt)
	if lsplt < 4 {
		return utils.ErrNotFound
	}
	fieldVal := splt[lsplt-1]
	fieldName := splt[lsplt-2]
	filterType := splt[lsplt-3]
	itemIDPrefix := utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
	_, err = dm.MatchFilterIndex(cacheID, itemIDPrefix, filterType, fieldName, fieldVal)
	return
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if itemIDs, rmtErr = rmtDM.MatchFilterIndex(cacheID, itemIDPrefix,
					filterType, fieldName, fieldVal); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound {
				Cache.Set(cacheID, fieldValKey, nil, nil,
					true, utils.NonTransactional)

			}
			return nil, err
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if supp, rmtErr = rmtDM.GetSupplierProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheSupplierProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
	}
	// populate cache will compute specific config parameters
	if err = supp.Compile(); err != nil {
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
		if err = createAndIndex(utils.SupplierProfilePrefix, supp.Tenant,
			utils.EmptyString, supp.ID, supp.FilterIDs, dm); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetSupplierProfile(supp, withIndex); err != nil {
				return
			}
		}
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
	if oldSupp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = NewFilterIndexer(dm, utils.SupplierProfilePrefix,
			tenant).RemoveItemFromIndex(tenant, id, oldSupp.FilterIDs); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveSupplierProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if attrPrfl, rmtErr = rmtDM.GetAttributeProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheAttributeProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetAttributeProfile(ap, withIndex); err != nil {
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
	if oldAttr == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		for _, context := range oldAttr.Contexts {
			if err = NewFilterIndexer(dm, utils.AttributeProfilePrefix,
				utils.ConcatenatedKey(tenant, context)).RemoveItemFromIndex(tenant, id, oldAttr.FilterIDs); err != nil {
				return
			}
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveAttributeProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
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
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if cpp, rmtErr = rmtDM.GetChargerProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheChargerProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
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
	if withIndex {
		if oldCpp != nil {
			var needsRemove bool
			for _, fltrID := range oldCpp.FilterIDs {
				if !utils.IsSliceMember(cpp.FilterIDs, fltrID) {
					needsRemove = true
				}
			}
			if needsRemove {
				if err = NewFilterIndexer(dm, utils.ChargerProfilePrefix,
					cpp.Tenant).RemoveItemFromIndex(cpp.Tenant, cpp.ID, oldCpp.FilterIDs); err != nil {
					return
				}
			}
		}
		if err = createAndIndex(utils.ChargerProfilePrefix, cpp.Tenant,
			utils.EmptyString, cpp.ID, cpp.FilterIDs, dm); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetChargerProfile(cpp, withIndex); err != nil {
				return
			}
		}
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
	if oldCpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = NewFilterIndexer(dm, utils.ChargerProfilePrefix,
			tenant).RemoveItemFromIndex(tenant, id, oldCpp.FilterIDs); err != nil {
			return
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveChargerProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) GetDispatcherProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (dpp *DispatcherProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheDispatcherProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*DispatcherProfile), nil
		}
	}
	dpp, err = dm.dataDB.GetDispatcherProfileDrv(tenant, id)
	if err != nil {
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if dpp, rmtErr = rmtDM.GetDispatcherProfile(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheDispatcherProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
	}
	if cacheWrite {
		Cache.Set(utils.CacheDispatcherProfiles, tntID, dpp, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetDispatcherProfile(dpp *DispatcherProfile, withIndex bool) (err error) {
	oldDpp, err := dm.GetDispatcherProfile(dpp.Tenant, dpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetDispatcherProfileDrv(dpp); err != nil {
		return err
	}
	if withIndex {
		if oldDpp != nil {
			for _, ctx := range oldDpp.Subsystems {
				var needsRemove bool
				if !utils.IsSliceMember(dpp.Subsystems, ctx) {
					needsRemove = true
				} else {
					for _, fltrID := range oldDpp.FilterIDs {
						if !utils.IsSliceMember(dpp.FilterIDs, fltrID) {
							needsRemove = true
						}
					}
				}
				if needsRemove {
					if err = NewFilterIndexer(dm, utils.DispatcherProfilePrefix,
						utils.ConcatenatedKey(dpp.Tenant, ctx)).RemoveItemFromIndex(dpp.Tenant, dpp.ID, oldDpp.FilterIDs); err != nil {
						return
					}
				}
			}
		}
		for _, ctx := range dpp.Subsystems {
			if err = createAndIndex(utils.DispatcherProfilePrefix, dpp.Tenant, ctx, dpp.ID, dpp.FilterIDs, dm); err != nil {
				return
			}
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetDispatcherProfile(dpp, withIndex); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) RemoveDispatcherProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	oldDpp, err := dm.GetDispatcherProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveDispatcherProfileDrv(tenant, id); err != nil {
		return
	}
	if oldDpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		for _, ctx := range oldDpp.Subsystems {
			if err = NewFilterIndexer(dm, utils.DispatcherProfilePrefix,
				utils.ConcatenatedKey(tenant, ctx)).RemoveItemFromIndex(tenant, id, oldDpp.FilterIDs); err != nil {
				return
			}
		}
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveDispatcherProfile(tenant, id,
				utils.NonTransactional, withIndex); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) GetDispatcherHost(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (dH *DispatcherHost, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheDispatcherHosts, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*DispatcherHost), nil
		}
	}
	dH, err = dm.dataDB.GetDispatcherHostDrv(tenant, id)
	if err != nil {
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if dH, rmtErr = rmtDM.GetDispatcherHost(tenant, id, false,
					false, utils.NonTransactional); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				Cache.Set(utils.CacheDispatcherHosts, tntID, nil, nil,
					cacheCommit(transactionID), transactionID)

			}
			return nil, err
		}
	}
	if cacheWrite {
		cfg := config.CgrConfig()
		if dH.rpcConn, err = NewRPCPool(
			rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			dH.Conns, IntRPC.GetInternalChanel(), false); err != nil {
			return nil, err
		}
		Cache.Set(utils.CacheDispatcherHosts, tntID, dH, nil,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (dm *DataManager) SetDispatcherHost(dpp *DispatcherHost) (err error) {
	if err = dm.DataDB().SetDispatcherHostDrv(dpp); err != nil {
		if len(dm.rplDataDBs) != 0 {
			for _, rplDM := range dm.rplDataDBs {
				if err = rplDM.SetDispatcherHost(dpp); err != nil {
					return
				}
			}
		}
	}
	return
}

func (dm *DataManager) RemoveDispatcherHost(tenant, id string,
	transactionID string) (err error) {
	oldDpp, err := dm.GetDispatcherHost(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveDispatcherHostDrv(tenant, id); err != nil {
		return
	}
	if oldDpp == nil {
		return utils.ErrNotFound
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.RemoveDispatcherHost(tenant, id,
				utils.NonTransactional); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) GetItemLoadIDs(itemIDPrefix string, cacheWrite bool) (loadIDs map[string]int64, err error) {
	loadIDs, err = dm.DataDB().GetItemLoadIDsDrv(itemIDPrefix)
	if err != nil {
		if len(dm.rmtDataDBs) != 0 {
			var rmtErr error
			for _, rmtDM := range dm.rmtDataDBs {
				if loadIDs, rmtErr = rmtDM.GetItemLoadIDs(itemIDPrefix, false); rmtErr == nil {
					break
				}
			}
			err = rmtErr
		}
		if err != nil {
			if err == utils.ErrNotFound && cacheWrite {
				for key, _ := range loadIDs {
					Cache.Set(utils.CacheLoadIDs, key, nil, nil,
						cacheCommit(utils.NonTransactional), utils.NonTransactional)
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		for key, val := range loadIDs {
			Cache.Set(utils.CacheLoadIDs, key, val, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
		}
	}
	return
}

func (dm *DataManager) SetLoadIDs(loadIDs map[string]int64) (err error) {
	if err = dm.DataDB().SetLoadIDsDrv(loadIDs); err != nil {
		return
	}
	if len(dm.rplDataDBs) != 0 {
		for _, rplDM := range dm.rplDataDBs {
			if err = rplDM.SetLoadIDs(loadIDs); err != nil {
				return
			}
		}
	}
	return
}

// Reconnect recconnects to the DB when the config was changes
func (dm *DataManager) Reconnect(marshaler string, newcfg *config.DataDbCfg) (err error) {
	var d DataDB
	switch newcfg.DataDbType {
	case utils.REDIS:
		var dbNb int
		dbNb, err = strconv.Atoi(newcfg.DataDbName)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return
		}
		host := newcfg.DataDbHost
		if newcfg.DataDbPort != "" && strings.Index(host, ":") == -1 {
			host += ":" + newcfg.DataDbPort
		}
		d, err = NewRedisStorage(host, dbNb, newcfg.DataDbPass, marshaler, utils.REDIS_MAX_CONNS, newcfg.DataDbSentinelName)
	case utils.MONGO:
		d, err = NewMongoStorage(newcfg.DataDbHost, newcfg.DataDbPort, newcfg.DataDbName,
			newcfg.DataDbUser, newcfg.DataDbPass, utils.DataDB, nil, true)
	case utils.INTERNAL:
		if marshaler == utils.JSON {
			d = NewInternalDBJson(nil, nil)
		} else {
			d = NewInternalDB(nil, nil)
		}
	default:
		err = fmt.Errorf("unknown db '%s' valid options are '%s' or '%s or '%s'",
			newcfg.DataDbType, utils.REDIS, utils.MONGO, utils.INTERNAL)
	}
	if err != nil {
		return
	}
	// ToDo: consider locking
	dm.dataDB.Close()
	dm.dataDB = d
	return
}
