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
	"time"

	"github.com/cgrates/baningo"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	filterIndexesPrefixMap = utils.StringSet{
		utils.AttributeFilterIndexes:      {},
		utils.ResourceFilterIndexes:       {},
		utils.StatFilterIndexes:           {},
		utils.ThresholdFilterIndexes:      {},
		utils.RouteFilterIndexes:          {},
		utils.ChargerFilterIndexes:        {},
		utils.DispatcherFilterIndexes:     {},
		utils.RateProfilesFilterIndexPrfx: {},
		utils.RateFilterIndexPrfx:         {},
		utils.ActionPlanIndexes:           {},
		utils.FilterIndexPrfx:             {},
	}
	cachePrefixMap = utils.StringSet{
		utils.DESTINATION_PREFIX:          {},
		utils.REVERSE_DESTINATION_PREFIX:  {},
		utils.RATING_PLAN_PREFIX:          {},
		utils.RATING_PROFILE_PREFIX:       {},
		utils.ACTION_PREFIX:               {},
		utils.ACTION_PLAN_PREFIX:          {},
		utils.AccountActionPlansPrefix:    {},
		utils.ACTION_TRIGGER_PREFIX:       {},
		utils.SHARED_GROUP_PREFIX:         {},
		utils.ResourceProfilesPrefix:      {},
		utils.TimingsPrefix:               {},
		utils.ResourcesPrefix:             {},
		utils.StatQueuePrefix:             {},
		utils.StatQueueProfilePrefix:      {},
		utils.ThresholdPrefix:             {},
		utils.ThresholdProfilePrefix:      {},
		utils.FilterPrefix:                {},
		utils.RouteProfilePrefix:          {},
		utils.AttributeProfilePrefix:      {},
		utils.ChargerProfilePrefix:        {},
		utils.DispatcherProfilePrefix:     {},
		utils.DispatcherHostPrefix:        {},
		utils.RateProfilePrefix:           {},
		utils.AttributeFilterIndexes:      {},
		utils.ResourceFilterIndexes:       {},
		utils.StatFilterIndexes:           {},
		utils.ThresholdFilterIndexes:      {},
		utils.RouteFilterIndexes:          {},
		utils.ChargerFilterIndexes:        {},
		utils.DispatcherFilterIndexes:     {},
		utils.RateProfilesFilterIndexPrfx: {},
		utils.RateFilterIndexPrfx:         {},
		utils.FilterIndexPrfx:             {},
		utils.MetaAPIBan:                  {}, // not realy a prefix as this is not stored in DB
	}
)

// NewDataManager returns a new DataManager
func NewDataManager(dataDB DataDB, cacheCfg *config.CacheCfg, connMgr *ConnManager) *DataManager {
	ms, _ := NewMarshaler(config.CgrConfig().GeneralCfg().DBDataEncoding)
	return &DataManager{
		dataDB:   dataDB,
		cacheCfg: cacheCfg,
		connMgr:  connMgr,
		ms:       ms,
	}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB   DataDB
	cacheCfg *config.CacheCfg
	connMgr  *ConnManager
	ms       Marshaler
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	if dm != nil {
		return dm.dataDB
	}
	return nil
}

func (dm *DataManager) LoadDataDBCache(attr map[string][]string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if dm.DataDB().GetStorageType() == utils.INTERNAL {
		return // all the data is in cache already
	}
	for key, ids := range attr {
		if err = dm.CacheDataFromDB(key, ids, false); err != nil {
			return
		}
	}
	return
}

func (dm *DataManager) CacheDataFromDB(prfx string, ids []string, mustBeCached bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if !cachePrefixMap.Has(prfx) {
		return utils.NewCGRError(utils.DataManager,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if dm.cacheCfg.Partitions[utils.CachePrefixToInstance[prfx]].Limit == 0 {
		return
	}
	if prfx == utils.MetaAPIBan { // no need for ids in this case
		ids = []string{utils.EmptyString}
	} else if ids == nil {
		if mustBeCached {
			ids = Cache.GetItemIDs(utils.CachePrefixToInstance[prfx], utils.EmptyString)
		} else {
			if ids, err = dm.DataDB().GetKeysForPrefix(prfx); err != nil {
				return utils.NewCGRError(utils.DataManager,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("DataManager error <%s> querying keys for prefix: <%s>", err.Error(), prfx))
			}
			if cCfg, has := dm.cacheCfg.Partitions[utils.CachePrefixToInstance[prfx]]; has &&
				cCfg.Limit >= 0 &&
				cCfg.Limit < len(ids) {
				ids = ids[:cCfg.Limit]
			}
			for i := range ids {
				ids[i] = strings.TrimPrefix(ids[i], prfx)
			}
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
			_, err = dm.GetDestination(dataID, true, utils.NonTransactional)
		case utils.REVERSE_DESTINATION_PREFIX:
			_, err = dm.GetReverseDestination(dataID, true, utils.NonTransactional)
		case utils.RATING_PLAN_PREFIX:
			_, err = dm.GetRatingPlan(dataID, true, utils.NonTransactional)
		case utils.RATING_PROFILE_PREFIX:
			_, err = dm.GetRatingProfile(dataID, true, utils.NonTransactional)
		case utils.ACTION_PREFIX:
			_, err = dm.GetActions(dataID, true, utils.NonTransactional)
		case utils.ACTION_PLAN_PREFIX:
			_, err = dm.GetActionPlan(dataID, true, utils.NonTransactional)
		case utils.AccountActionPlansPrefix:
			_, err = dm.GetAccountActionPlans(dataID, true, utils.NonTransactional)
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
		case utils.RouteProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetRouteProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
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
		case utils.RateProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetRateProfile(tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.AttributeFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, idxKey, false, true)
		case utils.ResourceFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheResourceFilterIndexes, tntCtx, idxKey, false, true)
		case utils.StatFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheStatFilterIndexes, tntCtx, idxKey, false, true)
		case utils.ThresholdFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheThresholdFilterIndexes, tntCtx, idxKey, false, true)
		case utils.RouteFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheRouteFilterIndexes, tntCtx, idxKey, false, true)
		case utils.ChargerFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheChargerFilterIndexes, tntCtx, idxKey, false, true)
		case utils.DispatcherFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheDispatcherFilterIndexes, tntCtx, idxKey, false, true)
		case utils.RateProfilesFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheRateProfilesFilterIndexes, tntCtx, idxKey, false, true)
		case utils.RateFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(utils.CacheRateFilterIndexes, tntCtx, idxKey, false, true)
		case utils.FilterIndexPrfx:
			idx := strings.LastIndexByte(dataID, utils.InInFieldSep[0])
			if idx < 0 {
				err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", dataID)
				return
			}
			_, err = dm.GetIndexes(utils.CacheReverseFilterIndexes, dataID[:idx], dataID[idx+1:], false, true)
		case utils.LoadIDPrefix:
			_, err = dm.GetItemLoadIDs(utils.EmptyString, true)
		case utils.MetaAPIBan:
			_, err = dm.GetAPIBan(utils.EmptyString, config.CgrConfig().APIBanCfg().Keys, false, false, true)
		}
		if err != nil {
			if err != utils.ErrNotFound {
				return utils.NewCGRError(utils.DataManager,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("error <%s> querying DataManager for category: <%s>, dataID: <%s>", err.Error(), prfx, dataID))
			}
			if err = Cache.Remove(utils.CachePrefixToInstance[prfx], dataID,
				cacheCommit(utils.NonTransactional), utils.NonTransactional); err != nil {
				return
			}
		}
	}
	return
}

func (dm *DataManager) GetDestination(key string, skipCache bool, transactionID string) (dest *Destination, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	dest, err = dm.dataDB.GetDestinationDrv(key, skipCache, transactionID)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetDestination, &utils.StringWithOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &dest); err == nil {
				err = dm.dataDB.SetDestinationDrv(dest, utils.NonTransactional)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetDestination(dest *Destination, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.SetDestinationDrv(dest, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetDestination,
			&DestinationWithOpts{
				Destination: dest,
				Tenant:      config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveDestination(destID string, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.RemoveDestinationDrv(destID, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1RemoveDestination,
			&utils.StringWithOpts{
				Arg:    destID,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

func (dm *DataManager) SetReverseDestination(dest *Destination, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.SetReverseDestinationDrv(dest, transactionID); err != nil {
		return
	}
	if config.CgrConfig().DataDbCfg().Items[utils.MetaReverseDestinations].Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetReverseDestination, dest, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) GetReverseDestination(prefix string,
	skipCache bool, transactionID string) (ids []string, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ids, err = dm.dataDB.GetReverseDestinationDrv(prefix, skipCache, transactionID)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaReverseDestinations]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetReverseDestination, &utils.StringWithOpts{
					Arg:    prefix,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &ids); err == nil {
				// need to discuss
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			return nil, err
		}
	}
	return
}

func (dm *DataManager) UpdateReverseDestination(oldDest, newDest *Destination,
	transactionID string) error {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	return dm.dataDB.UpdateReverseDestinationDrv(oldDest, newDest, transactionID)
}

func (dm *DataManager) GetAccount(id string) (acc *Account, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	acc, err = dm.dataDB.GetAccountDrv(id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccounts]; err == utils.ErrNotFound &&
			itm.Remote {
			splt := utils.SplitConcatenatedKey(id)
			tenant := utils.FirstNonEmpty(splt[0], config.CgrConfig().GeneralCfg().DefaultTenant)
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetAccount, &utils.StringWithOpts{
					Arg:    id,
					Tenant: tenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &acc); err == nil {
				err = dm.dataDB.SetAccountDrv(acc)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetAccount(acc *Account) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.SetAccountDrv(acc); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetAccount,
			&AccountWithOpts{
				Account: acc,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveAccount(id string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.RemoveAccountDrv(id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveAccount,
			&utils.StringWithOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	sq, err = dm.dataDB.GetStatQueueDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetStatQueue,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &sq); err == nil {
				var ssq *StoredStatQueue
				if dm.dataDB.GetStorageType() != utils.MetaInternal {
					// in case of internal we don't marshal
					if ssq, err = NewStoredStatQueue(sq, dm.ms); err != nil {
						return nil, err
					}
				}
				err = dm.dataDB.SetStatQueueDrv(ssq, sq)
			}
		}
		if err != nil {
			if err = utils.CastRPCErr(err); err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheStatQueues, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheStatQueues, tntID, sq, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

// SetStatQueue converts to StoredStatQueue and stores the result in dataDB
func (dm *DataManager) SetStatQueue(sq *StatQueue, metrics []*MetricWithFilters,
	minItems int, ttl *time.Duration, queueLength int, simpleSet bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if !simpleSet {
		tnt := sq.Tenant // save the tenant
		id := sq.ID      // save the ID from the initial StatQueue
		// handle metrics for statsQueue
		sq, err = dm.GetStatQueue(tnt, id, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return
		}
		if err == utils.ErrNotFound {
			sq = &StatQueue{Tenant: tnt, ID: id, SQMetrics: make(map[string]StatMetric)}
			// if the statQueue didn't exists simply initiate all the metrics
			for _, metric := range metrics {
				var stsMetric StatMetric
				if stsMetric, err = NewStatMetric(metric.MetricID,
					minItems,
					metric.FilterIDs); err != nil {
					return
				}
				sq.SQMetrics[metric.MetricID] = stsMetric
			}
		} else {
			for sqMetricID := range sq.SQMetrics {
				// we consider that the metric needs to be removed
				needsRemove := true
				for _, metric := range metrics {
					// in case we found the metric in the metrics define by the user we leave it
					if sqMetricID == metric.MetricID {
						needsRemove = false
						break
					}
					if _, has := sq.SQMetrics[metric.MetricID]; !has {
						var stsMetric StatMetric
						if stsMetric, err = NewStatMetric(metric.MetricID,
							minItems,
							metric.FilterIDs); err != nil {
							return
						}
						sq.SQMetrics[metric.MetricID] = stsMetric
					}
				}
				if needsRemove {
					delete(sq.SQMetrics, sqMetricID)
				}
			}
			// if the user define a statQueue with an existing metric check if we need to update it based on queue length
			sq.ttl = ttl
			if _, err = sq.remExpired(); err != nil {
				return
			}
			if len(sq.SQItems) > queueLength {
				for i := 0; i < queueLength-len(sq.SQItems); i++ {
					item := sq.SQItems[0]
					if err = sq.remEventWithID(item.EventID); err != nil {
						return
					}
					sq.SQItems = sq.SQItems[1:]
				}
			}
		}
	}

	var ssq *StoredStatQueue
	if dm.dataDB.GetStorageType() != utils.MetaInternal ||
		config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues].Replicate {
		// in case of internal we don't marshal
		if ssq, err = NewStoredStatQueue(sq, dm.ms); err != nil {
			return
		}
	}
	if err = dm.dataDB.SetStatQueueDrv(ssq, sq); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetStatQueue,
			&StoredStatQueueWithOpts{
				StoredStatQueue: ssq,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

// RemoveStatQueue removes the StoredStatQueue
func (dm *DataManager) RemoveStatQueue(tenant, id string, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.RemStatQueueDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveStatQueue,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

// GetFilter returns a filter based on the given ID
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
		if fltr, err = NewFilterFromInline(tenant, id); err != nil {
			return
		}
	} else if dm == nil { // in case we want the filter from dataDB but the connection to dataDB a optional (e.g. SessionS)
		err = utils.ErrNoDatabaseConn
		return
	} else {
		fltr, err = dm.DataDB().GetFilterDrv(tenant, id)
		if err != nil {
			if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaFilters]; err == utils.ErrNotFound && itm.Remote {
				if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetFilter,
					&utils.TenantIDWithOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						Opts: map[string]interface{}{
							utils.OptsAPIKey:  itm.APIKey,
							utils.OptsRouteID: itm.RouteID,
						}}, &fltr); err == nil {
					err = dm.dataDB.SetFilterDrv(fltr)
				}
			}
			if err != nil {
				err = utils.CastRPCErr(err)
				if err == utils.ErrNotFound && cacheWrite {
					if errCh := Cache.Set(utils.CacheFilters, tntID, nil, nil,
						cacheCommit(transactionID), transactionID); errCh != nil {
						return nil, errCh
					}
				}
				return
			}
		}
		if err = fltr.Compile(); err != nil { // only compile the value when we get the filter from DB or from remote0
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheFilters, tntID, fltr, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetFilter(fltr *Filter, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	var oldFlt *Filter
	if oldFlt, err = dm.GetFilter(fltr.Tenant, fltr.ID, true, false,
		utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetFilterDrv(fltr); err != nil {
		return
	}
	if withIndex {
		if err = UpdateFilterIndex(dm, oldFlt, fltr); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaFilters]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetFilter,
			&FilterWithOpts{
				Filter: fltr,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	var oldFlt *Filter
	if oldFlt, err = dm.GetFilter(tenant, id, true, false,
		utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		return err
	}
	var tntCtx string
	if withIndex {
		tntCtx = utils.ConcatenatedKey(tenant, id)
		var rcvIndx map[string]utils.StringSet
		if rcvIndx, err = dm.GetIndexes(utils.CacheReverseFilterIndexes, tntCtx,
			utils.EmptyString, true, true); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			err = nil // no index for this filter so  no remove needed from index side
		} else {
			return fmt.Errorf("cannot remove filter <%s> because will broken the reference to following items: %s",
				tntCtx, utils.ToJSON(rcvIndx))
		}
	}
	if err = dm.DataDB().RemoveFilterDrv(tenant, id); err != nil {
		return
	}
	if oldFlt == nil {
		return utils.ErrNotFound
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaFilters]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveFilter,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	th, err = dm.dataDB.GetThresholdDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholds]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetThreshold, &utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &th); err == nil {
				err = dm.dataDB.SetThresholdDrv(th)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheThresholds, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheThresholds, tntID, th, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetThreshold(th *Threshold, snooze time.Duration, simpleSet bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if !simpleSet {
		tnt := th.Tenant // save the tenant
		id := th.ID      // save the ID from the initial Threshold
		th, err = dm.GetThreshold(tnt, id, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return
		}
		if err == utils.ErrNotFound {
			th = &Threshold{Tenant: tnt, ID: id, Hits: 0}
		} else {
			if th.tPrfl == nil {
				if th.tPrfl, err = dm.GetThresholdProfile(th.Tenant, th.ID, true, false, utils.NonTransactional); err != nil {
					return
				}
			}
			th.Snooze = th.Snooze.Add(-th.tPrfl.MinSleep).Add(snooze)
		}
	}
	if err = dm.DataDB().SetThresholdDrv(th); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholds]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetThreshold,
			&ThresholdWithOpts{
				Threshold: th,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholds]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1RemoveThreshold,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	th, err = dm.dataDB.GetThresholdProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholdProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetThresholdProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &th); err == nil {
				err = dm.dataDB.SetThresholdProfileDrv(th)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheThresholdProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheThresholdProfiles, tntID, th, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetThresholdProfile(th *ThresholdProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(th.Tenant, th.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, th.TenantID())
		}
	}
	oldTh, err := dm.GetThresholdProfile(th.Tenant, th.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetThresholdProfileDrv(th); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldTh != nil {
			oldFiltersIDs = &oldTh.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheThresholdFilterIndexes, th.Tenant,
			utils.EmptyString, th.ID, oldFiltersIDs, th.FilterIDs); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholdProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetThresholdProfile,
			&ThresholdProfileWithOpts{
				ThresholdProfile: th,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveThresholdProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheThresholdFilterIndexes, tenant, id, oldTh.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheThresholdFilterIndexes,
			tenant, utils.EmptyString, id, oldTh.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholdProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1RemoveThresholdProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	sqp, err = dm.dataDB.GetStatQueueProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueueProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetStatQueueProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &sqp); err == nil {
				err = dm.dataDB.SetStatQueueProfileDrv(sqp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheStatQueueProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheStatQueueProfiles, tntID, sqp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetStatQueueProfile(sqp *StatQueueProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(sqp.Tenant, sqp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, sqp.TenantID())
		}
	}
	oldSts, err := dm.GetStatQueueProfile(sqp.Tenant, sqp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetStatQueueProfileDrv(sqp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldSts != nil {
			oldFiltersIDs = &oldSts.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheStatFilterIndexes, sqp.Tenant,
			utils.EmptyString, sqp.ID, oldFiltersIDs, sqp.FilterIDs); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueueProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetStatQueueProfile,
			&StatQueueProfileWithOpts{
				StatQueueProfile: sqp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveStatQueueProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheStatFilterIndexes, tenant, id, oldSts.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheStatFilterIndexes,
			tenant, utils.EmptyString, id, oldSts.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueueProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1RemoveStatQueueProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	t, err = dm.dataDB.GetTimingDrv(id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaTimings]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetTiming,
				&utils.StringWithOpts{
					Arg:    id,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &t); err == nil {
				err = dm.dataDB.SetTimingDrv(t)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheTimings, id, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheTimings, id, t, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

func (dm *DataManager) SetTiming(t *utils.TPTiming) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetTimingDrv(t); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.TimingsPrefix, []string{t.ID}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaTimings]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetTiming,
			&utils.TPTimingWithOpts{
				TPTiming: t,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveTiming(id, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveTimingDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheTimings, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if config.CgrConfig().DataDbCfg().Items[utils.MetaTimings].Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveTiming, id, &reply)
	}
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rs, err = dm.dataDB.GetResourceDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResources]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetResource,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &rs); err == nil {
				err = dm.dataDB.SetResourceDrv(rs)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheResources, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheResources, tntID, rs, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetResource(rs *Resource, ttl *time.Duration, usageLimit float64, simpleSet bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if !simpleSet {
		// do stuff
		tnt := rs.Tenant // save the tenant
		id := rs.ID      // save the ID from the initial StatQueue
		// handle metrics for statsQueue
		rs, err = dm.GetResource(tnt, id, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return
		}
		if err == utils.ErrNotFound {
			rs = &Resource{Tenant: tnt, ID: id, Usages: make(map[string]*ResourceUsage)}
			// if the resource didn't exists simply initiate the Usages
		} else {
			rs.ttl = ttl
			rs.removeExpiredUnits()
			for rsUsage := range rs.Usages {
				if rs.totalUsage() > usageLimit {
					if err = rs.clearUsage(rsUsage); err != nil {
						return
					}
				} else {
					break
				}
			}
		}
	}
	if err = dm.DataDB().SetResourceDrv(rs); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResources]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetResource,
			&ResourceWithOpts{
				Resource: rs,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveResource(tenant, id, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveResourceDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResources]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveResource,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rp, err = dm.dataDB.GetResourceProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResourceProfile]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetResourceProfile, &utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &rp); err == nil {
				err = dm.dataDB.SetResourceProfileDrv(rp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheResourceProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheResourceProfiles, tntID, rp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetResourceProfile(rp *ResourceProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(rp.Tenant, rp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, rp.TenantID())
		}
	}
	oldRes, err := dm.GetResourceProfile(rp.Tenant, rp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetResourceProfileDrv(rp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRes != nil {
			oldFiltersIDs = &oldRes.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheResourceFilterIndexes, rp.Tenant,
			utils.EmptyString, rp.ID, oldFiltersIDs, rp.FilterIDs); err != nil {
			return err
		}
		Cache.Clear([]string{utils.CacheEventResources})
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResourceProfile]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetResourceProfile,
			&ResourceProfileWithOpts{
				ResourceProfile: rp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveResourceProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheResourceFilterIndexes, tenant, id, oldRes.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheResourceFilterIndexes,
			tenant, utils.EmptyString, id, oldRes.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResourceProfile]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveResourceProfile, &utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	attrs, err = dm.dataDB.GetActionTriggersDrv(id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionTriggers]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetActionTriggers,
				&utils.StringWithOpts{
					Arg:    id,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, attrs); err == nil {
				err = dm.dataDB.SetActionTriggersDrv(id, attrs)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheActionTriggers, id, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheActionTriggers, id, attrs, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

func (dm *DataManager) RemoveActionTriggers(id, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheActionTriggers, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionTriggers]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveActionTriggers,
			&utils.StringWithOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

//SetActionTriggersArgWithOpts is used to send the key and the ActionTriggers to Replicator
type SetActionTriggersArgWithOpts struct {
	Key    string
	Attrs  ActionTriggers
	Tenant string
	Opts   map[string]interface{}
}

func (dm *DataManager) SetActionTriggers(key string, attr ActionTriggers,
	transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetActionTriggersDrv(key, attr); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.ACTION_TRIGGER_PREFIX, []string{key}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionTriggers]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1SetActionTriggers,
			&SetActionTriggersArgWithOpts{
				Attrs: attr, Key: key,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	sg, err = dm.DataDB().GetSharedGroupDrv(key)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaSharedGroups]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetSharedGroup, &utils.StringWithOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &sg); err == nil {
				err = dm.dataDB.SetSharedGroupDrv(sg)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheSharedGroups, key, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheSharedGroups, key, sg, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

func (dm *DataManager) SetSharedGroup(sg *SharedGroup,
	transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetSharedGroupDrv(sg); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.SHARED_GROUP_PREFIX,
		[]string{sg.Id}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaSharedGroups]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetSharedGroup,
			&SharedGroupWithOpts{
				SharedGroup: sg,
				Tenant:      config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveSharedGroup(id, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveSharedGroupDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheSharedGroups, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaSharedGroups]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveSharedGroup,
			&utils.StringWithOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	as, err = dm.DataDB().GetActionsDrv(key)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActions]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetActions, &utils.StringWithOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &as); err == nil {
				err = dm.dataDB.SetActionsDrv(key, as)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheActions, key, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheActions, key, as, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

//SetActionsArgsWithOpts is used to send the key and the Actions to replicator
type SetActionsArgsWithOpts struct {
	Key    string
	Acs    Actions
	Tenant string
	Opts   map[string]interface{}
}

func (dm *DataManager) SetActions(key string, as Actions, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetActionsDrv(key, as); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.ACTION_PREFIX, []string{key}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActions]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil, utils.ReplicatorSv1SetActions,
			&SetActionsArgsWithOpts{
				Key: key, Acs: as,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveActions(key, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveActionsDrv(key); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheActions, key,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActions]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveActions, &utils.StringWithOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

func (dm *DataManager) GetActionPlan(key string, skipCache bool, transactionID string) (ats *ActionPlan, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ats, err = dm.dataDB.GetActionPlanDrv(key, skipCache, transactionID)
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; err == utils.ErrNotFound && itm.Remote {
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
			utils.ReplicatorSv1GetActionPlan, &utils.StringWithOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &ats); err == nil {
			err = dm.dataDB.SetActionPlanDrv(key, ats, true, utils.NonTransactional)
		}
	}
	if err != nil {
		err = utils.CastRPCErr(err)
		return nil, err
	}
	return
}

// SetActionPlanArgWithOpts is used in replicatorV1 for dispatcher
type SetActionPlanArgWithOpts struct {
	Key       string
	Ats       *ActionPlan
	Overwrite bool
	Tenant    string
	Opts      map[string]interface{}
}

func (dm *DataManager) SetActionPlan(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.SetActionPlanDrv(key, ats, overwrite, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetActionPlan, &SetActionPlanArgWithOpts{
				Key:       key,
				Ats:       ats,
				Overwrite: overwrite,
				Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) GetAllActionPlans() (ats map[string]*ActionPlan, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ats, err = dm.dataDB.GetAllActionPlansDrv()
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; ((err == nil && len(ats) == 0) || err == utils.ErrNotFound) && itm.Remote {
		err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
			utils.ReplicatorSv1GetAllActionPlans,
			&utils.StringWithOpts{
				Arg:    utils.EmptyString,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &ats)
	}
	if err != nil {
		err = utils.CastRPCErr(err)
		return nil, err
	}
	return
}

func (dm *DataManager) RemoveActionPlan(key string, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.RemoveActionPlanDrv(key, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveActionPlan,
			&utils.StringWithOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}
func (dm *DataManager) GetAccountActionPlans(acntID string,
	skipCache bool, transactionID string) (apIDs []string, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	apIDs, err = dm.dataDB.GetAccountActionPlansDrv(acntID, skipCache, transactionID)
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccountActionPlans]; ((err == nil && len(apIDs) == 0) || err == utils.ErrNotFound) && itm.Remote {
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
			utils.ReplicatorSv1GetAccountActionPlans,
			&utils.StringWithOpts{
				Arg:    acntID,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &apIDs); err == nil {
			err = dm.dataDB.SetAccountActionPlansDrv(acntID, apIDs, true)
		}
	}
	if err != nil {
		err = utils.CastRPCErr(err)
		return nil, err
	}
	return
}

//SetAccountActionPlansArgWithOpts is used to send the key and the Actions to replicator
type SetAccountActionPlansArgWithOpts struct {
	AcntID    string
	AplIDs    []string
	Overwrite bool
	Tenant    string
	Opts      map[string]interface{}
}

func (dm *DataManager) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.SetAccountActionPlansDrv(acntID, aPlIDs, overwrite); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccountActionPlans]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetAccountActionPlans, &SetAccountActionPlansArgWithOpts{
				AcntID:    acntID,
				AplIDs:    aPlIDs,
				Overwrite: overwrite,
				Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

// RemAccountActionPlansArgsWithOpts is used in replicatorV1 for dispatcher
type RemAccountActionPlansArgsWithOpts struct {
	AcntID string
	ApIDs  []string
	Tenant string
	Opts   map[string]interface{}
}

func (dm *DataManager) RemAccountActionPlans(acntID string, apIDs []string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.dataDB.RemAccountActionPlansDrv(acntID, apIDs); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccountActionPlans]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemAccountActionPlans,
			&RemAccountActionPlansArgsWithOpts{
				AcntID: acntID, ApIDs: apIDs,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rp, err = dm.DataDB().GetRatingPlanDrv(key)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingPlans]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetRatingPlan,
				&utils.StringWithOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &rp); err == nil {
				err = dm.dataDB.SetRatingPlanDrv(rp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheRatingPlans, key, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheRatingPlans, key, rp, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

func (dm *DataManager) SetRatingPlan(rp *RatingPlan, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetRatingPlanDrv(rp); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.RATING_PLAN_PREFIX, []string{rp.Id}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingPlans]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRatingPlan,
			&RatingPlanWithOpts{
				RatingPlan: rp,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveRatingPlan(key string, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveRatingPlanDrv(key); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheRatingPlans, key,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingPlans]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveRatingPlan,
			&utils.StringWithOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

// GetRatingProfile returns the RatingProfile for the key
func (dm *DataManager) GetRatingProfile(key string, skipCache bool,
	transactionID string) (rpf *RatingProfile, err error) {
	if !skipCache {
		for _, cacheRP := range []string{utils.CacheRatingProfilesTmp, utils.CacheRatingProfiles} {
			if x, ok := Cache.Get(cacheRP, key); ok {
				if x != nil {
					return x.(*RatingProfile), nil
				}
				return nil, utils.ErrNotFound
			}
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rpf, err = dm.DataDB().GetRatingProfileDrv(key)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetRatingProfile,
				&utils.StringWithOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &rpf); err == nil {
				err = dm.dataDB.SetRatingProfileDrv(rpf)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheRatingProfiles, key, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if errCh := Cache.Set(utils.CacheRatingProfiles, key, rpf, nil,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return nil, errCh
	}
	return
}

func (dm *DataManager) SetRatingProfile(rpf *RatingProfile,
	transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetRatingProfileDrv(rpf); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.RATING_PROFILE_PREFIX, []string{rpf.Id}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRatingProfile,
			&RatingProfileWithOpts{
				RatingProfile: rpf,
				Tenant:        config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveRatingProfile(key string,
	transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveRatingProfileDrv(key); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheRatingProfiles, key,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveRatingProfile,
			&utils.StringWithOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

func (dm *DataManager) HasData(category, subject, tenant string) (has bool, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	return dm.DataDB().HasDataDrv(category, subject, tenant)
}

func (dm *DataManager) GetRouteProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rpp *RouteProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRouteProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RouteProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rpp, err = dm.dataDB.GetRouteProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRouteProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetRouteProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &rpp); err == nil {
				err = dm.dataDB.SetRouteProfileDrv(rpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheRouteProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	// populate cache will compute specific config parameters
	if err = rpp.Compile(); err != nil {
		return nil, err
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheRouteProfiles, tntID, rpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetRouteProfile(rpp *RouteProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(rpp.Tenant, rpp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, rpp.TenantID())
		}
	}
	oldRpp, err := dm.GetRouteProfile(rpp.Tenant, rpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetRouteProfileDrv(rpp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheRouteFilterIndexes, rpp.Tenant,
			utils.EmptyString, rpp.ID, oldFiltersIDs, rpp.FilterIDs); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRouteProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRouteProfile,
			&RouteProfileWithOpts{
				RouteProfile: rpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveRouteProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	oldRpp, err := dm.GetRouteProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveRouteProfileDrv(tenant, id); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(dm, utils.CacheRouteFilterIndexes, tenant, id, oldRpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheRouteFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRouteProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveRouteProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

// GetAttributeProfile returns the AttributeProfile with the given id
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
	if strings.HasPrefix(id, utils.Meta) {
		attrPrfl, err = NewAttributeFromInline(tenant, id)
		return // do not set inline attributes in cache it breaks the interanal db matching
	} else if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	} else {
		if attrPrfl, err = dm.dataDB.GetAttributeProfileDrv(tenant, id); err != nil {
			if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAttributeProfiles]; err == utils.ErrNotFound && itm.Remote {
				if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
					utils.ReplicatorSv1GetAttributeProfile,
					&utils.TenantIDWithOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						Opts: map[string]interface{}{
							utils.OptsAPIKey:  itm.APIKey,
							utils.OptsRouteID: itm.RouteID,
						}}, &attrPrfl); err == nil {
					err = dm.dataDB.SetAttributeProfileDrv(attrPrfl)
				}
			}
			if err != nil {
				err = utils.CastRPCErr(err)
				if err == utils.ErrNotFound && cacheWrite {
					if errCh := Cache.Set(utils.CacheAttributeProfiles, tntID, nil, nil,
						cacheCommit(transactionID), transactionID); errCh != nil {
						return nil, errCh
					}

				}
				return nil, err
			}
		}
		if err = attrPrfl.Compile(); err != nil {
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheAttributeProfiles, tntID, attrPrfl, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetAttributeProfile(ap *AttributeProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(ap.Tenant, ap.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, ap.TenantID())
		}
	}
	oldAP, err := dm.GetAttributeProfile(ap.Tenant, ap.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetAttributeProfileDrv(ap); err != nil {
		return err
	}
	if withIndex {
		var oldContexes *[]string
		var oldFiltersIDs *[]string
		if oldAP != nil {
			oldContexes = &oldAP.Contexts
			oldFiltersIDs = &oldAP.FilterIDs
		}
		if err = updatedIndexesWithContexts(dm, utils.CacheAttributeFilterIndexes, ap.Tenant, ap.ID,
			oldContexes, oldFiltersIDs, ap.Contexts, ap.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAttributeProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetAttributeProfile,
			&AttributeProfileWithOpts{
				AttributeProfile: ap,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveAttributeProfile(tenant, id string, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheAttributeFilterIndexes, tenant, id, oldAttr.FilterIDs); err != nil {
			return
		}
		for _, context := range oldAttr.Contexts {
			if err = removeItemFromFilterIndex(dm, utils.CacheAttributeFilterIndexes,
				tenant, context, id, oldAttr.FilterIDs); err != nil {
				return
			}
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAttributeProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveAttributeProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	cpp, err = dm.dataDB.GetChargerProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaChargerProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetChargerProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &cpp); err == nil {
				err = dm.dataDB.SetChargerProfileDrv(cpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheChargerProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheChargerProfiles, tntID, cpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetChargerProfile(cpp *ChargerProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(cpp.Tenant, cpp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, cpp.TenantID())
		}
	}
	oldCpp, err := dm.GetChargerProfile(cpp.Tenant, cpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetChargerProfileDrv(cpp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldCpp != nil {
			oldFiltersIDs = &oldCpp.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheChargerFilterIndexes, cpp.Tenant,
			utils.EmptyString, cpp.ID, oldFiltersIDs, cpp.FilterIDs); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaChargerProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetChargerProfile,
			&ChargerProfileWithOpts{
				ChargerProfile: cpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveChargerProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheChargerFilterIndexes, tenant, id, oldCpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheChargerFilterIndexes,
			tenant, utils.EmptyString, id, oldCpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaChargerProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveChargerProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	dpp, err = dm.dataDB.GetDispatcherProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetDispatcherProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &dpp); err == nil {
				err = dm.dataDB.SetDispatcherProfileDrv(dpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheDispatcherProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheDispatcherProfiles, tntID, dpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetDispatcherProfile(dpp *DispatcherProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(dpp.Tenant, dpp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, dpp.TenantID())
		}
	}
	oldDpp, err := dm.GetDispatcherProfile(dpp.Tenant, dpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetDispatcherProfileDrv(dpp); err != nil {
		return err
	}
	if withIndex {
		var oldContexes *[]string
		var oldFiltersIDs *[]string
		if oldDpp != nil {
			oldContexes = &oldDpp.Subsystems
			oldFiltersIDs = &oldDpp.FilterIDs
		}
		if err = updatedIndexesWithContexts(dm, utils.CacheDispatcherFilterIndexes, dpp.Tenant, dpp.ID,
			oldContexes, oldFiltersIDs, dpp.Subsystems, dpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetDispatcherProfile,
			&DispatcherProfileWithOpts{
				DispatcherProfile: dpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveDispatcherProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
		if err = removeIndexFiltersItem(dm, utils.CacheDispatcherFilterIndexes, tenant, id, oldDpp.FilterIDs); err != nil {
			return
		}
		for _, ctx := range oldDpp.Subsystems {
			if err = removeItemFromFilterIndex(dm, utils.CacheDispatcherFilterIndexes,
				tenant, ctx, id, oldDpp.FilterIDs); err != nil {
				return
			}
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveDispatcherProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
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
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	dH, err = dm.dataDB.GetDispatcherHostDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherHosts]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetDispatcherHost,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &dH); err == nil {
				err = dm.dataDB.SetDispatcherHostDrv(dH)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheDispatcherHosts, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if err = Cache.Set(utils.CacheDispatcherHosts, tntID, dH, nil,
			cacheCommit(transactionID), transactionID); err != nil {
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetDispatcherHost(dpp *DispatcherHost) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetDispatcherHostDrv(dpp); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherHosts]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetDispatcherHost,
			&DispatcherHostWithOpts{
				DispatcherHost: dpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveDispatcherHost(tenant, id string,
	transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
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
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherHosts]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveDispatcherHost,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

func (dm *DataManager) GetItemLoadIDs(itemIDPrefix string, cacheWrite bool) (loadIDs map[string]int64, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	loadIDs, err = dm.DataDB().GetItemLoadIDsDrv(itemIDPrefix)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaLoadIDs]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetItemLoadIDs,
				&utils.StringWithOpts{
					Arg:    itemIDPrefix,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &loadIDs); err == nil {
				err = dm.dataDB.SetLoadIDsDrv(loadIDs)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				for key := range loadIDs {
					if errCh := Cache.Set(utils.CacheLoadIDs, key, nil, nil,
						cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
						return nil, errCh
					}
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		for key, val := range loadIDs {
			if errCh := Cache.Set(utils.CacheLoadIDs, key, val, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
				return nil, errCh
			}
		}
	}
	return
}

func (dm *DataManager) SetLoadIDs(loadIDs map[string]int64) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetLoadIDsDrv(loadIDs); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaLoadIDs]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetLoadIDs,
			&utils.LoadIDsWithOpts{
				LoadIDs: loadIDs,
				Tenant:  config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) GetRateProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rpp *RateProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRateProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*RateProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rpp, err = dm.dataDB.GetRateProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRateProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetRateProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &rpp); err == nil {
				rpp.Sort()
				err = dm.dataDB.SetRateProfileDrv(rpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheRateProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if err = rpp.Compile(); err != nil {
		return nil, err
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheRateProfiles, tntID, rpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetRateProfile(rpp *RateProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(rpp.Tenant, rpp.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, rpp.TenantID())
		}
	}
	oldRpp, err := dm.GetRateProfile(rpp.Tenant, rpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	rpp.Sort()
	if err = dm.DataDB().SetRateProfileDrv(rpp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheRateProfilesFilterIndexes, rpp.Tenant,
			utils.EmptyString, rpp.ID, oldFiltersIDs, rpp.FilterIDs); err != nil {
			return err
		}
		// remove indexes for old rates
		if oldRpp != nil {
			for key, rate := range oldRpp.Rates {
				if _, has := rpp.Rates[key]; has {
					continue
				}
				if err = removeItemFromFilterIndex(dm, utils.CacheRateFilterIndexes,
					rpp.Tenant, rpp.ID, key, rate.FilterIDs); err != nil {
					return
				}
			}
		}
		// create index for each rate
		for key, rate := range rpp.Rates {
			var oldRateFiltersIDs *[]string
			if oldRpp != nil {
				if oldRate, has := oldRpp.Rates[key]; has {
					oldRateFiltersIDs = &oldRate.FilterIDs
				}
			}
			// when we create the indexes for rates we use RateProfile ID as context
			if err := updatedIndexes(dm, utils.CacheRateFilterIndexes, rpp.Tenant,
				rpp.ID, key, oldRateFiltersIDs, rate.FilterIDs); err != nil {
				return err
			}
		}

	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRateProfile,
			&RateProfileWithOpts{
				RateProfile: rpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveRateProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	oldRpp, err := dm.GetRateProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveRateProfileDrv(tenant, id); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		for key, rate := range oldRpp.Rates {
			if err = removeItemFromFilterIndex(dm, utils.CacheRateFilterIndexes,
				oldRpp.Tenant, oldRpp.ID, key, rate.FilterIDs); err != nil {
				return
			}
		}
		if err = removeItemFromFilterIndex(dm, utils.CacheRateProfilesFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveRateProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

func (dm *DataManager) RemoveRateProfileRates(tenant, id string, rateIDs []string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	oldRpp, err := dm.GetRateProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil {
		return err
	}
	if len(rateIDs) == 0 {
		if withIndex {
			for key, rate := range oldRpp.Rates {
				if err = removeItemFromFilterIndex(dm, utils.CacheRateFilterIndexes,
					tenant, id, key, rate.FilterIDs); err != nil {
					return
				}
			}
		}
		oldRpp.Rates = map[string]*Rate{}
	} else {
		for _, rateID := range rateIDs {
			if _, has := oldRpp.Rates[rateID]; !has {
				continue
			}
			if withIndex {

				if err = removeItemFromFilterIndex(dm, utils.CacheRateFilterIndexes,
					tenant, id, rateID, oldRpp.Rates[rateID].FilterIDs); err != nil {
					return
				}
			}
			delete(oldRpp.Rates, rateID)
		}
	}
	if err = dm.DataDB().SetRateProfileDrv(oldRpp); err != nil {
		return err
	}

	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRateProfile,
			&RateProfileWithOpts{
				RateProfile: oldRpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) SetRateProfileRates(rpp *RateProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		for _, rate := range rpp.Rates {
			if brokenReference := dm.checkFilters(rpp.Tenant, rate.FilterIDs); len(brokenReference) != 0 {
				// if we get a broken filter do not update the rates
				return fmt.Errorf("broken reference to filter: %+v for rate with ID: %+v",
					brokenReference, rate.ID)
			}
		}
	}
	oldRpp, err := dm.GetRateProfile(rpp.Tenant, rpp.ID, true, false, utils.NonTransactional)
	if err != nil {
		return err
	}
	// create index for each rate
	for key, rate := range rpp.Rates {
		if withIndex {
			var oldRateFiltersIDs *[]string
			if oldRate, has := oldRpp.Rates[key]; has {
				oldRateFiltersIDs = &oldRate.FilterIDs
			}
			// when we create the indexes for rates we use RateProfile ID as context
			if err := updatedIndexes(dm, utils.CacheRateFilterIndexes, rpp.Tenant,
				rpp.ID, key, oldRateFiltersIDs, rate.FilterIDs); err != nil {
				return err
			}
		}
		oldRpp.Rates[key] = rate
	}

	if err = dm.DataDB().SetRateProfileDrv(oldRpp); err != nil {
		return err
	}

	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetRateProfile,
			&RateProfileWithOpts{
				RateProfile: oldRpp,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) GetActionProfile(tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (ap *ActionProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheActionProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ActionProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ap, err = dm.dataDB.GetActionProfileDrv(tenant, id)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetActionProfile,
				&utils.TenantIDWithOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					}}, &ap); err == nil {
				err = dm.dataDB.SetActionProfileDrv(ap)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheActionProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheActionProfiles, tntID, ap, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetActionProfile(ap *ActionProfile, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if withIndex {
		if brokenReference := dm.checkFilters(ap.Tenant, ap.FilterIDs); len(brokenReference) != 0 {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("broken reference to filter: %+v for item with ID: %+v",
				brokenReference, ap.TenantID())
		}
	}
	oldRpp, err := dm.GetActionProfile(ap.Tenant, ap.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetActionProfileDrv(ap); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(dm, utils.CacheActionProfilesFilterIndexes, ap.Tenant,
			utils.EmptyString, ap.ID, oldFiltersIDs, ap.FilterIDs); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionProfiles]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetActionProfile,
			&ActionProfileWithOpts{
				ActionProfile: ap,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply); err != nil {
			err = utils.CastRPCErr(err)
			return
		}
	}
	return
}

func (dm *DataManager) RemoveActionProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	oldRpp, err := dm.GetActionProfile(tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveActionProfileDrv(tenant, id); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeItemFromFilterIndex(dm, utils.CacheActionProfilesFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionProfiles]; itm.Replicate {
		var reply string
		dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveActionProfile,
			&utils.TenantIDWithOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				}}, &reply)
	}
	return
}

// Reconnect reconnects to the DB when the config was changed
func (dm *DataManager) Reconnect(marshaller string, newcfg *config.DataDbCfg) (err error) {
	d, err := NewDataDBConn(newcfg.DataDbType, newcfg.DataDbHost, newcfg.DataDbPort, newcfg.DataDbName,
		newcfg.DataDbUser, newcfg.DataDbPass, marshaller, newcfg.Opts)
	if err != nil {
		return
	}
	// ToDo: consider locking
	dm.dataDB.Close()
	dm.dataDB = d
	return
}

func (dm *DataManager) GetIndexes(idxItmType, tntCtx, idxKey string,
	cacheRead, cacheWrite bool) (indexes map[string]utils.StringSet, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}

	if cacheRead && idxKey != utils.EmptyString { // do not check cache if we want all the indexes
		if x, ok := Cache.Get(idxItmType, utils.ConcatenatedKey(tntCtx, idxKey)); ok { // Attempt to find in cache first
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return map[string]utils.StringSet{
				idxKey: x.(utils.StringSet),
			}, nil
		}
	}
	if indexes, err = dm.DataDB().GetIndexesDrv(idxItmType, tntCtx, idxKey); err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetIndexes,
				&utils.GetIndexesArg{
					IdxItmType: idxItmType,
					TntCtx:     tntCtx,
					IdxKey:     idxKey,
					Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
					Opts: map[string]interface{}{
						utils.OptsAPIKey:  itm.APIKey,
						utils.OptsRouteID: itm.RouteID,
					},
				}, &indexes); err == nil {
				err = dm.dataDB.SetIndexesDrv(idxItmType, tntCtx, indexes, true, utils.NonTransactional)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound {
				if cacheWrite && idxKey != utils.EmptyString {
					if errCh := Cache.Set(idxItmType, utils.ConcatenatedKey(tntCtx, idxKey), nil, []string{tntCtx},
						true, utils.NonTransactional); errCh != nil {
						return nil, errCh
					}
				}
			}
			return nil, err
		}
	}

	if cacheWrite {
		for k, v := range indexes {
			if err = Cache.Set(idxItmType, utils.ConcatenatedKey(tntCtx, k), v, []string{tntCtx},
				true, utils.NonTransactional); err != nil {
				return nil, err
			}
		}
	}
	return
}

func (dm *DataManager) SetIndexes(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().SetIndexesDrv(idxItmType, tntCtx,
		indexes, commit, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1SetIndexes,
			&utils.SetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				Indexes:    indexes,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &reply); err != nil {
			err = utils.CastRPCErr(err)
		}
	}
	return
}

func (dm *DataManager) RemoveIndexes(idxItmType, tntCtx, idxKey string) (err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if err = dm.DataDB().RemoveIndexesDrv(idxItmType, tntCtx, idxKey); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes]; itm.Replicate {
		var reply string
		if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RplConns, nil,
			utils.ReplicatorSv1RemoveIndexes,
			&utils.GetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				IdxKey:     idxKey,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				Opts: map[string]interface{}{
					utils.OptsAPIKey:  itm.APIKey,
					utils.OptsRouteID: itm.RouteID,
				},
			}, &reply); err != nil {
			err = utils.CastRPCErr(err)
		}
	}
	return
}

func (dm *DataManager) GetAPIBan(ip string, apiKeys []string, single, cacheRead, cacheWrite bool) (banned bool, err error) {
	if cacheRead {
		if x, ok := Cache.Get(utils.MetaAPIBan, ip); ok && x != nil { // Attempt to find in cache first
			return x.(bool), nil
		}
	}
	if single {
		if banned, err = baningo.CheckIP(ip, apiKeys...); err != nil {
			return
		}
		if cacheWrite {
			if err = Cache.Set(utils.MetaAPIBan, ip, banned, nil, true, utils.NonTransactional); err != nil {
				return false, err
			}
		}
		return
	}
	var bannedIPs []string
	if bannedIPs, err = baningo.GetBannedIPs(apiKeys...); err != nil {
		return
	}
	for _, bannedIP := range bannedIPs {
		if bannedIP == ip {
			banned = true
		}
		if cacheWrite {
			if err = Cache.Set(utils.MetaAPIBan, bannedIP, true, nil, true, utils.NonTransactional); err != nil {
				return false, err
			}
		}
	}
	if len(ip) != 0 && !banned && cacheWrite {
		if err = Cache.Set(utils.MetaAPIBan, ip, false, nil, true, utils.NonTransactional); err != nil {
			return false, err
		}
	}
	return
}

// checkFilters returns the id of the first Filter that is not valid
// it should be called after the dm nil check
func (dm *DataManager) checkFilters(tenant string, ids []string) (brokenReference string) {
	for _, id := range ids {
		// in case of inline filter we try to build them
		// if they are not correct it should fail here not in indexes
		if strings.HasPrefix(id, utils.Meta) {
			if _, err := NewFilterFromInline(tenant, id); err != nil {
				return id
			}
		} else if x, has := Cache.Get(utils.CacheFilters, // because the method HasDataDrv doesn't use cache
			utils.ConcatenatedKey(tenant, id)); has && x == nil { // check to see if filter is already in cache
			return id
		} else if has, err := dm.DataDB().HasDataDrv(utils.FilterPrefix, // check in local DB if we have the filter
			id, tenant); err != nil || !has {
			// in case we can not find it localy try to find it in the remote DB
			if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaFilters]; err == utils.ErrNotFound && itm.Remote {
				var fltr *Filter
				err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetFilter,
					&utils.TenantIDWithOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						Opts: map[string]interface{}{
							utils.OptsAPIKey:  itm.APIKey,
							utils.OptsRouteID: itm.RouteID,
						}}, &fltr)
				has = fltr == nil
			}
			// not in local DB and not in remote DB
			if err != nil || !has {
				return id
			}
		}
	}
	return
}
