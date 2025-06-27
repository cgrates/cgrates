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
	"slices"
	"strings"

	"github.com/cgrates/baningo"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

var (
	filterIndexesPrefixMap = utils.StringSet{
		utils.AttributeFilterIndexes:        {},
		utils.ResourceFilterIndexes:         {},
		utils.IPFilterIndexes:               {},
		utils.StatFilterIndexes:             {},
		utils.ThresholdFilterIndexes:        {},
		utils.RouteFilterIndexes:            {},
		utils.ChargerFilterIndexes:          {},
		utils.RateProfilesFilterIndexPrfx:   {},
		utils.ActionProfilesFilterIndexPrfx: {},
		utils.RateFilterIndexPrfx:           {},
		utils.ActionPlanIndexes:             {},
		utils.FilterIndexPrfx:               {},
		utils.AccountFilterIndexPrfx:        {},
	}
	cachePrefixMap = utils.StringSet{
		utils.ResourceProfilesPrefix:        {},
		utils.ResourcesPrefix:               {},
		utils.IPProfilesPrefix:              {},
		utils.IPAllocationsPrefix:           {},
		utils.StatQueuePrefix:               {},
		utils.StatQueueProfilePrefix:        {},
		utils.ThresholdPrefix:               {},
		utils.ThresholdProfilePrefix:        {},
		utils.TrendPrefix:                   {},
		utils.TrendProfilePrefix:            {},
		utils.RankingProfilePrefix:          {},
		utils.RankingPrefix:                 {},
		utils.FilterPrefix:                  {},
		utils.RouteProfilePrefix:            {},
		utils.AttributeProfilePrefix:        {},
		utils.ChargerProfilePrefix:          {},
		utils.AccountFilterIndexPrfx:        {},
		utils.AccountPrefix:                 {},
		utils.RateProfilePrefix:             {},
		utils.ActionProfilePrefix:           {},
		utils.AttributeFilterIndexes:        {},
		utils.ResourceFilterIndexes:         {},
		utils.IPFilterIndexes:               {},
		utils.StatFilterIndexes:             {},
		utils.ThresholdFilterIndexes:        {},
		utils.RouteFilterIndexes:            {},
		utils.ChargerFilterIndexes:          {},
		utils.RateProfilesFilterIndexPrfx:   {},
		utils.ActionProfilesFilterIndexPrfx: {},
		utils.RateFilterIndexPrfx:           {},
		utils.FilterIndexPrfx:               {},
		utils.MetaAPIBan:                    {}, // not realy a prefix as this is not stored in DB
	}
)

// NewDataManager returns a new DataManager
func NewDataManager(dataDB DataDB, cfg *config.CGRConfig, connMgr *ConnManager) *DataManager {
	ms, _ := utils.NewMarshaler(cfg.GeneralCfg().DBDataEncoding)
	return &DataManager{
		dataDB:  dataDB,
		cfg:     cfg,
		connMgr: connMgr,
		ms:      ms,
	}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB  DataDB
	cfg     *config.CGRConfig
	connMgr *ConnManager
	ms      utils.Marshaler
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	if dm != nil {
		return dm.dataDB
	}
	return nil
}

func (dm *DataManager) CacheDataFromDB(ctx *context.Context, prfx string, ids []string, mustBeCached bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if !cachePrefixMap.Has(prfx) {
		return utils.NewCGRError(utils.DataManager,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedCachePrefix,
			fmt.Sprintf("prefix <%s> is not a supported cache prefix", prfx))
	}
	if dm.cfg.CacheCfg().Partitions[utils.CachePrefixToInstance[prfx]].Limit == 0 {
		return
	}
	// *apiban and *dispatchers are not stored in database
	if prfx == utils.MetaAPIBan || prfx == utils.MetaSentryPeer { // no need for ids in this case
		ids = []string{utils.EmptyString}
	} else if len(ids) != 0 && ids[0] == utils.MetaAny {
		if mustBeCached {
			ids = Cache.GetItemIDs(utils.CachePrefixToInstance[prfx], utils.EmptyString)
		} else {
			if ids, err = dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
				return utils.NewCGRError(utils.DataManager,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("DataManager error <%s> querying keys for prefix: <%s>", err.Error(), prfx))
			}
			if cCfg, has := dm.cfg.CacheCfg().Partitions[utils.CachePrefixToInstance[prfx]]; has &&
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
		if mustBeCached &&
			!Cache.HasItem(utils.CachePrefixToInstance[prfx], dataID) { // only cache if previously there
			continue
		}
		switch prfx {
		case utils.ResourceProfilesPrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, utils.ResourceProfileLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetResourceProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.ResourcesPrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, utils.ResourceLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetResource(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.IPProfilesPrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, utils.IPProfileLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetIPProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.IPAllocationsPrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, utils.IPAllocationsLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetIPAllocations(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.StatQueueProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, statQueueProfileLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetStatQueueProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.StatQueuePrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, statQueueLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetStatQueue(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.ThresholdProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, thresholdProfileLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetThresholdProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.RankingProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, utils.RankingProfileLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetRankingProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.TrendProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetTrendProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ThresholdPrefix:
			tntID := utils.NewTenantID(dataID)
			lkID := guardian.Guardian.GuardIDs("", dm.cfg.GeneralCfg().LockingTimeout, thresholdLockKey(tntID.Tenant, tntID.ID))
			_, err = dm.GetThreshold(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
			guardian.Guardian.UnguardIDs(lkID)
		case utils.FilterPrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetFilter(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.RouteProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetRouteProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.AttributeProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetAttributeProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ChargerProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetChargerProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.RateProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetRateProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.ActionProfilePrefix:
			tntID := utils.NewTenantID(dataID)
			_, err = dm.GetActionProfile(ctx, tntID.Tenant, tntID.ID, false, true, utils.NonTransactional)
		case utils.AttributeFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheAttributeFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.ResourceFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheResourceFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.IPFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheIPFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.StatFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheStatFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.ThresholdFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheThresholdFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.RouteFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheRouteFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.ChargerFilterIndexes:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheChargerFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.RateProfilesFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheRateProfilesFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.RateFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheRateFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.ActionProfilesFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheActionProfilesFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.AccountFilterIndexPrfx:
			var tntCtx, idxKey string
			if tntCtx, idxKey, err = splitFilterIndex(dataID); err != nil {
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheAccountsFilterIndexes, tntCtx, idxKey, utils.NonTransactional, false, true)
		case utils.FilterIndexPrfx:
			idx := strings.LastIndexByte(dataID, utils.InInFieldSep[0])
			if idx < 0 {
				err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", dataID)
				return
			}
			_, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, dataID[:idx], dataID[idx+1:], utils.NonTransactional, false, true)
		case utils.LoadIDPrefix:
			_, err = dm.GetItemLoadIDs(ctx, utils.EmptyString, true)
		case utils.MetaAPIBan:
			_, err = GetAPIBan(ctx, utils.EmptyString, dm.cfg.APIBanCfg().Keys, false, false, true)
		}
		if err != nil {
			if err != utils.ErrNotFound &&
				err != utils.ErrDSPProfileNotFound &&
				err != utils.ErrDSPHostNotFound {
				return utils.NewCGRError(utils.DataManager,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("error <%s> querying DataManager for category: <%s>, dataID: <%s>", err.Error(), prfx, dataID))
			}
			err = nil
			// if err = Cache.Remove(ctx, utils.CachePrefixToInstance[prfx], dataID,
			// cacheCommit(utils.NonTransactional), utils.NonTransactional); err != nil {
			// return
			// }
		}
	}
	return
}

// GetFilter returns a filter based on the given ID
func (dm *DataManager) GetFilter(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
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
		fltr, err = dm.DataDB().GetFilterDrv(ctx, tenant, id)
		if err != nil {
			if itm := dm.cfg.DataDbCfg().Items[utils.MetaFilters]; err == utils.ErrNotFound && itm.Remote {
				if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns, utils.ReplicatorSv1GetFilter,
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
								dm.cfg.GeneralCfg().NodeID)),
					}, &fltr); err == nil {
					err = dm.dataDB.SetFilterDrv(ctx, fltr)
				}
			}
			if err != nil {
				err = utils.CastRPCErr(err)
				if err == utils.ErrNotFound && cacheWrite {
					if errCh := Cache.Set(ctx, utils.CacheFilters, tntID, nil, nil,
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
		if errCh := Cache.Set(ctx, utils.CacheFilters, tntID, fltr, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetFilter(ctx *context.Context, fltr *Filter, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	var oldFlt *Filter
	if oldFlt, err = dm.GetFilter(ctx, fltr.Tenant, fltr.ID, true, false,
		utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetFilterDrv(ctx, fltr); err != nil {
		return
	}
	if withIndex {
		if err = UpdateFilterIndex(ctx, dm, oldFlt, fltr); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaFilters]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.FilterPrefix, fltr.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetFilter,
			&FilterWithAPIOpts{
				Filter: fltr,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveFilter(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	var oldFlt *Filter
	if oldFlt, err = dm.GetFilter(ctx, tenant, id, true, false,
		utils.NonTransactional); err != nil && err != utils.ErrNotFound {
		return err
	}
	var tntCtx string
	if withIndex {
		tntCtx = utils.ConcatenatedKey(tenant, id)
		var rcvIndx map[string]utils.StringSet
		if rcvIndx, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, tntCtx,
			utils.EmptyString, utils.NonTransactional, true, true); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			err = nil // no index for this filter so  no remove needed from index side
		} else {
			return fmt.Errorf("cannot remove filter <%s> because will broken the reference to following items: %s",
				tntCtx, utils.ToJSON(rcvIndx))
		}
	}
	if err = dm.DataDB().RemoveFilterDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldFlt == nil {
		return utils.ErrNotFound
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaFilters]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.FilterPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveFilter,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetThreshold(ctx *context.Context, tenant, id string,
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
	th, err = dm.dataDB.GetThresholdDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholds]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetThreshold, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &th); err == nil {
				err = dm.dataDB.SetThresholdDrv(ctx, th)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheThresholds, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheThresholds, tntID, th, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetThreshold(ctx *context.Context, th *Threshold) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetThresholdDrv(ctx, th); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholds]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ThresholdPrefix, th.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetThreshold,
			&ThresholdWithAPIOpts{
				Threshold: th,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveThreshold(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveThresholdDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholds]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ThresholdPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveThreshold,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetThresholdProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
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
	th, err = dm.dataDB.GetThresholdProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholdProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetThresholdProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &th); err == nil {
				err = dm.dataDB.SetThresholdProfileDrv(ctx, th)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheThresholdProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheThresholdProfiles, tntID, th, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetThresholdProfile(ctx *context.Context, th *ThresholdProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, th.Tenant, th.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, th.TenantID())
		}
	}
	oldTh, err := dm.GetThresholdProfile(ctx, th.Tenant, th.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetThresholdProfileDrv(ctx, th); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldTh != nil {
			oldFiltersIDs = &oldTh.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheThresholdFilterIndexes, th.Tenant,
			utils.EmptyString, th.ID, oldFiltersIDs, th.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholdProfiles]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ThresholdProfilePrefix, th.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetThresholdProfile,
			&ThresholdProfileWithAPIOpts{
				ThresholdProfile: th,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}

	if oldTh == nil || // create the threshold if it didn't exist before
		oldTh.MaxHits != th.MaxHits ||
		oldTh.MinHits != th.MinHits ||
		oldTh.MinSleep != th.MinSleep { // reset the threshold if the profile changed this fields
		err = dm.SetThreshold(ctx, &Threshold{
			Tenant: th.Tenant,
			ID:     th.ID,
			Hits:   0,
		})
	} else if _, errTh := dm.GetThreshold(ctx, th.Tenant, th.ID, // do not try to get the threshold if the configuration changed
		true, false, utils.NonTransactional); errTh == utils.ErrNotFound { // the threshold does not exist
		err = dm.SetThreshold(ctx, &Threshold{
			Tenant: th.Tenant,
			ID:     th.ID,
			Hits:   0,
		})
	}
	return
}

func (dm *DataManager) RemoveThresholdProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldTh, err := dm.GetThresholdProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemThresholdProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldTh == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheThresholdFilterIndexes, tenant, id, oldTh.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheThresholdFilterIndexes,
			tenant, utils.EmptyString, id, oldTh.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaThresholdProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ThresholdProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return dm.RemoveThreshold(ctx, tenant, id) // remove the threshold
}

// GetStatQueue retrieves a StatQueue from dataDB
// handles caching and deserialization of metrics
func (dm *DataManager) GetStatQueue(ctx *context.Context, tenant, id string,
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
	sq, err = dm.dataDB.GetStatQueueDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueues]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns, utils.ReplicatorSv1GetStatQueue,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &sq); err == nil {
				var ssq *StoredStatQueue
				if dm.dataDB.GetStorageType() != utils.MetaInternal {
					// in case of internal we don't marshal
					if ssq, err = NewStoredStatQueue(sq, dm.ms); err != nil {
						return nil, err
					}
				}
				err = dm.dataDB.SetStatQueueDrv(ctx, ssq, sq)
			}
		}
		if err != nil {
			if err = utils.CastRPCErr(err); err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheStatQueues, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheStatQueues, tntID, sq, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

// SetStatQueue converts to StoredStatQueue and stores the result in dataDB
func (dm *DataManager) SetStatQueue(ctx *context.Context, sq *StatQueue) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	var ssq *StoredStatQueue
	if dm.dataDB.GetStorageType() != utils.MetaInternal {
		// in case of internal we don't marshal
		if ssq, err = NewStoredStatQueue(sq, dm.ms); err != nil {
			return
		}
	}
	if err = dm.dataDB.SetStatQueueDrv(ctx, ssq, sq); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.StatQueuePrefix, sq.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetStatQueue,
			&StatQueueWithAPIOpts{
				StatQueue: sq,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// RemoveStatQueue removes the StoredStatQueue
func (dm *DataManager) RemoveStatQueue(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.RemStatQueueDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.StatQueuePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveStatQueue,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetStatQueueProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
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
	sqp, err = dm.dataDB.GetStatQueueProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetStatQueueProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &sqp); err == nil {
				err = dm.dataDB.SetStatQueueProfileDrv(ctx, sqp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheStatQueueProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheStatQueueProfiles, tntID, sqp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetStatQueueProfile(ctx *context.Context, sqp *StatQueueProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, sqp.Tenant, sqp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, sqp.TenantID())
		}
	}
	oldSts, err := dm.GetStatQueueProfile(ctx, sqp.Tenant, sqp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetStatQueueProfileDrv(ctx, sqp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldSts != nil {
			oldFiltersIDs = &oldSts.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheStatFilterIndexes, sqp.Tenant,
			utils.EmptyString, sqp.ID, oldFiltersIDs, sqp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.StatQueueProfilePrefix, sqp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetStatQueueProfile,
			&StatQueueProfileWithAPIOpts{
				StatQueueProfile: sqp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}
	if oldSts == nil || // create the stats queue if it didn't exist before
		oldSts.QueueLength != sqp.QueueLength ||
		oldSts.TTL != sqp.TTL ||
		oldSts.MinItems != sqp.MinItems ||
		oldSts.Stored != sqp.Stored && oldSts.Stored { // reset the stats queue if the profile changed this fields
		guardian.Guardian.Guard(ctx, func(ctx *context.Context) (_ error) { // we change the queue so lock it
			var sq *StatQueue
			if sq, err = NewStatQueue(sqp.Tenant, sqp.ID, sqp.Metrics,
				uint64(sqp.MinItems)); err != nil {
				return
			}
			err = dm.SetStatQueue(ctx, sq)
			return
		}, dm.cfg.GeneralCfg().LockingTimeout, utils.StatQueuePrefix+sqp.TenantID())
	} else {
		guardian.Guardian.Guard(ctx, func(ctx *context.Context) (_ error) { // we change the queue so lock it
			oSq, errRs := dm.GetStatQueue(ctx, sqp.Tenant, sqp.ID, // do not try to get the stats queue if the configuration changed
				true, false, utils.NonTransactional)
			if errRs == utils.ErrNotFound { // the stats queue does not exist
				var sq *StatQueue
				if sq, err = NewStatQueue(sqp.Tenant, sqp.ID, sqp.Metrics,
					uint64(sqp.MinItems)); err != nil {
					return
				}
				err = dm.SetStatQueue(ctx, sq)
				return
			} else if errRs != nil {
				return
			}
			// update the metrics if needed
			cMetricIDs := utils.StringSet{}
			for _, metric := range sqp.Metrics { // add missing metrics and recreate the old metrics that changed
				cMetricIDs.Add(metric.MetricID)
				if oSqMetric, has := oSq.SQMetrics[metric.MetricID]; !has ||
					!slices.Equal(oSqMetric.GetFilterIDs(), metric.FilterIDs) { // recreate it if the filter changed
					if oSq.SQMetrics[metric.MetricID], err = NewStatMetric(metric.MetricID,
						uint64(sqp.MinItems), metric.FilterIDs); err != nil {
						return
					}
				}
			}
			for sqMetricID := range oSq.SQMetrics { // remove the old metrics
				if !cMetricIDs.Has(sqMetricID) {
					delete(oSq.SQMetrics, sqMetricID)
				}
			}
			if sqp.Stored { // already changed the value in cache
				err = dm.SetStatQueue(ctx, oSq) // only set it in DB if Stored is true
			}
			return
		}, dm.cfg.GeneralCfg().LockingTimeout, utils.StatQueuePrefix+sqp.TenantID())
	}
	return
}

func (dm *DataManager) RemoveStatQueueProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldSts, err := dm.GetStatQueueProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemStatQueueProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldSts == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheStatFilterIndexes, tenant, id, oldSts.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheStatFilterIndexes,
			tenant, utils.EmptyString, id, oldSts.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaStatQueueProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.StatQueueProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveStatQueueProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return dm.RemoveStatQueue(ctx, tenant, id)
}

// GetTrend retrieves a Trend from dataDB
func (dm *DataManager) GetTrend(ctx *context.Context, tenant, id string,
	cacheRead, cacheWrite bool, transactionID string) (tr *utils.Trend, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)

	if cacheRead {
		if x, ok := Cache.Get(utils.CacheTrends, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.Trend), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if tr, err = dm.dataDB.GetTrendDrv(ctx, tenant, id); err != nil {
		if err != utils.ErrNotFound { // database error
			return
		}
		// ErrNotFound
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaTrends]; itm.Remote {
			if err = dm.connMgr.Call(context.TODO(), dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetTrend, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &tr); err != nil {
				err = utils.CastRPCErr(err)
				if err != utils.ErrNotFound { // RPC error
					return
				}
			} else if err = dm.dataDB.SetTrendDrv(ctx, tr); err != nil {
				return
			}
		}
		// have Trend or ErrNotFound
		if err == utils.ErrNotFound {
			if cacheWrite {
				if errCache := Cache.Set(ctx, utils.CacheTrends, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCache != nil {
					return nil, errCache
				}
			}
			return
		}
	}
	if err = tr.Uncompress(dm.ms); err != nil {
		return nil, err
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheTrends, tntID, tr, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}

	return
}

// SetTrend stores Trend in dataDB
func (dm *DataManager) SetTrend(ctx *context.Context, tr *utils.Trend) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if dm.dataDB.GetStorageType() != utils.MetaInternal {
		if tr, err = tr.Compress(dm.ms, dm.cfg.TrendSCfg().StoreUncompressedLimit); err != nil {
			return
		}
	}
	if err = dm.DataDB().SetTrendDrv(ctx, tr); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaTrends]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.TrendPrefix, tr.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetTrend,
			&utils.TrendWithAPIOpts{
				Trend: tr,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}
	return
}

// RemoveTrend removes the stored Trend
func (dm *DataManager) RemoveTrend(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveTrendDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaTrends]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.TrendPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveTrend,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetTrendProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (trp *utils.TrendProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheTrendProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.TrendProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	trp, err = dm.dataDB.GetTrendProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaTrendProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetTrendProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &trp); err == nil {
				err = dm.dataDB.SetTrendProfileDrv(ctx, trp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheTrendProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheTrendProfiles, tntID, trp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) GetTrendProfileIDs(ctx *context.Context, tenants []string) (tps map[string][]string, err error) {
	prfx := utils.TrendProfilePrefix
	var keys []string
	if len(tenants) == 0 {
		keys, err = dm.dataDB.GetKeysForPrefix(ctx, prfx)
		if err != nil {
			return
		}
	} else {
		for _, tenant := range tenants {
			var tntkeys []string
			tntPrfx := prfx + tenant + utils.ConcatenatedKeySep
			tntkeys, err = dm.dataDB.GetKeysForPrefix(ctx, tntPrfx)
			if err != nil {
				return
			}
			keys = append(keys, tntkeys...)
		}
	}
	// if len(keys) == 0 {
	// 	return nil, utils.ErrNotFound
	// }

	tps = make(map[string][]string)
	for _, key := range keys {
		indx := strings.Index(key, utils.ConcatenatedKeySep)
		tenant := key[len(utils.TrendProfilePrefix):indx]
		id := key[indx+1:]
		tps[tenant] = append(tps[tenant], id)
	}
	return
}

func (dm *DataManager) SetTrendProfile(ctx *context.Context, trp *utils.TrendProfile) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldTrd, err := dm.GetTrendProfile(ctx, trp.Tenant, trp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetTrendProfileDrv(ctx, trp); err != nil {
		return err
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaTrendProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.TrendProfilePrefix, trp.TenantID(),
			utils.ReplicatorSv1SetTrendProfile,
			&utils.TrendProfileWithAPIOpts{
				TrendProfile: trp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	if oldTrd == nil ||
		oldTrd.QueueLength != trp.QueueLength ||
		oldTrd.Schedule != trp.Schedule {
		if err = dm.SetTrend(ctx, utils.NewTrendFromProfile(trp)); err != nil {
			return
		}
	}
	return
}

func (dm *DataManager) RemoveTrendProfile(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldTrs, err := dm.GetTrendProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}

	if err = dm.DataDB().RemTrendProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldTrs == nil {
		return utils.ErrNotFound
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankingProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.TrendProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveTrendProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return dm.RemoveTrend(ctx, tenant, id)
}

func (dm *DataManager) GetRankingProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool, transactionID string) (rgp *utils.RankingProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRankingProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.RankingProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rgp, err = dm.dataDB.GetRankingProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankingProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(context.TODO(), dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetRankingProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rgp); err == nil {
				err = dm.dataDB.SetRankingProfileDrv(ctx, rgp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheRankingProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheRankingProfiles, tntID, rgp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) GetRankingProfileIDs(ctx *context.Context, tenants []string) (rns map[string][]string, err error) {
	prfx := utils.RankingProfilePrefix
	var keys []string
	if len(tenants) == 0 {
		keys, err = dm.dataDB.GetKeysForPrefix(ctx, prfx)
		if err != nil {
			return
		}
	} else {
		for _, tenant := range tenants {
			var tntkeys []string
			tntPrfx := prfx + tenant + utils.ConcatenatedKeySep
			tntkeys, err = dm.dataDB.GetKeysForPrefix(ctx, tntPrfx)
			if err != nil {
				return
			}
			keys = append(keys, tntkeys...)
		}
	}
	// if len(keys) == 0 {
	// 	return nil, utils.ErrNotFound
	// }
	rns = make(map[string][]string)
	for _, key := range keys {
		indx := strings.Index(key, utils.ConcatenatedKeySep)
		tenant := key[len(utils.RankingProfilePrefix):indx]
		id := key[indx+1:]
		rns[tenant] = append(rns[tenant], id)
	}
	return
}

func (dm *DataManager) SetRankingProfile(ctx *context.Context, rnp *utils.RankingProfile) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldRnk, err := dm.GetRankingProfile(ctx, rnp.Tenant, rnp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetRankingProfileDrv(ctx, rnp); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankingProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RankingProfilePrefix, rnp.TenantID(),
			utils.ReplicatorSv1SetRankingProfile,
			&utils.RankingProfileWithAPIOpts{
				RankingProfile: rnp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	if oldRnk == nil || oldRnk.Sorting != rnp.Sorting ||
		oldRnk.Schedule != rnp.Schedule {
		if err = dm.SetRanking(ctx, utils.NewRankingFromProfile(rnp)); err != nil {
			return
		}
	}
	return
}

func (dm *DataManager) RemoveRankingProfile(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldSgs, err := dm.GetRankingProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemRankingProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldSgs == nil {
		return utils.ErrNotFound
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankingProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RankingProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRankingProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}
func (dm *DataManager) GetRanking(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool, transactionID string) (rn *utils.Ranking, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRankings, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.Ranking), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if rn, err = dm.dataDB.GetRankingDrv(ctx, tenant, id); err != nil {
		if err != utils.ErrNotFound { // database error
			return
		}
		// ErrNotFound
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankings]; itm.Remote {
			if err = dm.connMgr.Call(context.TODO(), dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetRanking, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rn); err == nil {
				err = dm.dataDB.SetRankingDrv(ctx, rn)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheRankings, tntID, nil, nil, cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheRankings, tntID, rn, nil, cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

// SetRanking stores Ranking in dataDB
func (dm *DataManager) SetRanking(ctx *context.Context, rn *utils.Ranking) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetRankingDrv(ctx, rn); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankings]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RankingPrefix, rn.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRanking,
			&utils.RankingWithAPIOpts{
				Ranking: rn,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}
	return
}

// RemoveRanking removes the stored Ranking
func (dm *DataManager) RemoveRanking(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveRankingDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRankings]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RankingPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRanking,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetResource(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rs *utils.Resource, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheResources, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.Resource), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rs, err = dm.dataDB.GetResourceDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaResources]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetResource,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rs); err == nil {
				err = dm.dataDB.SetResourceDrv(ctx, rs)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheResources, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheResources, tntID, rs, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetResource(ctx *context.Context, rs *utils.Resource) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetResourceDrv(ctx, rs); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaResources]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ResourcesPrefix, rs.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetResource,
			&utils.ResourceWithAPIOpts{
				Resource: rs,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveResource(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveResourceDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaResources]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ResourcesPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveResource,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetResourceProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rp *utils.ResourceProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheResourceProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.ResourceProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rp, err = dm.dataDB.GetResourceProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaResourceProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetResourceProfile, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rp); err == nil {
				err = dm.dataDB.SetResourceProfileDrv(ctx, rp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheResourceProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheResourceProfiles, tntID, rp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetResourceProfile(ctx *context.Context, rp *utils.ResourceProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, rp.Tenant, rp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, rp.TenantID())
		}
	}
	oldRes, err := dm.GetResourceProfile(ctx, rp.Tenant, rp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetResourceProfileDrv(ctx, rp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRes != nil {
			oldFiltersIDs = &oldRes.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheResourceFilterIndexes, rp.Tenant,
			utils.EmptyString, rp.ID, oldFiltersIDs, rp.FilterIDs, false); err != nil {
			return err
		}
		Cache.Clear([]string{utils.CacheEventResources})
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaResourceProfiles]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ResourceProfilesPrefix, rp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetResourceProfile,
			&utils.ResourceProfileWithAPIOpts{
				ResourceProfile: rp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}
	if oldRes == nil || // create the resource if it didn't exist before
		oldRes.UsageTTL != rp.UsageTTL ||
		oldRes.Limit != rp.Limit ||
		oldRes.Stored != rp.Stored && oldRes.Stored { // reset the resource if the profile changed this fields
		err = dm.SetResource(ctx, &utils.Resource{
			Tenant: rp.Tenant,
			ID:     rp.ID,
			Usages: make(map[string]*utils.ResourceUsage),
		})
	} else if _, errRs := dm.GetResource(ctx, rp.Tenant, rp.ID, // do not try to get the resource if the configuration changed
		true, false, utils.NonTransactional); errRs == utils.ErrNotFound { // the resource does not exist
		err = dm.SetResource(ctx, &utils.Resource{
			Tenant: rp.Tenant,
			ID:     rp.ID,
			Usages: make(map[string]*utils.ResourceUsage),
		})
	}
	return
}

func (dm *DataManager) RemoveResourceProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldRes, err := dm.GetResourceProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveResourceProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldRes == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheResourceFilterIndexes, tenant, id, oldRes.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheResourceFilterIndexes,
			tenant, utils.EmptyString, id, oldRes.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaResourceProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ResourceProfilesPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveResourceProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return dm.RemoveResource(ctx, tenant, id)
}

func (dm *DataManager) GetIPAllocations(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (ip *utils.IPAllocations, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheIPAllocations, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.IPAllocations), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ip, err = dm.dataDB.GetIPAllocationsDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPAllocations]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetIPAllocations,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &ip); err == nil {
				err = dm.dataDB.SetIPAllocationsDrv(ctx, ip)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheIPAllocations, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheIPAllocations, tntID, ip, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetIPAllocations(ctx *context.Context, ip *utils.IPAllocations) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetIPAllocationsDrv(ctx, ip); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPAllocations]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.IPAllocationsPrefix, ip.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetIPAllocations,
			&utils.IPAllocationsWithAPIOpts{
				IPAllocations: ip,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveIPAllocations(ctx *context.Context, tenant, id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveIPAllocationsDrv(ctx, tenant, id); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPAllocations]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.IPAllocationsPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveIPAllocations,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetIPProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (ipp *utils.IPProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheIPProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.IPProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ipp, err = dm.dataDB.GetIPProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetIPProfile, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &ipp); err == nil {
				err = dm.dataDB.SetIPProfileDrv(ctx, ipp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheIPProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheIPProfiles, tntID, ipp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetIPProfile(ctx *context.Context, ipp *utils.IPProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, ipp.Tenant, ipp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, ipp.TenantID())
		}
	}
	oldIPP, err := dm.GetIPProfile(ctx, ipp.Tenant, ipp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetIPProfileDrv(ctx, ipp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldIPP != nil {
			oldFiltersIDs = &oldIPP.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheIPFilterIndexes, ipp.Tenant,
			utils.EmptyString, ipp.ID, oldFiltersIDs, ipp.FilterIDs, false); err != nil {
			return err
		}
		Cache.Clear([]string{utils.CacheEventIPs})
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPProfiles]; itm.Replicate {
		if err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.IPProfilesPrefix, ipp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetIPProfile,
			&utils.IPProfileWithAPIOpts{
				IPProfile: ipp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)}); err != nil {
			return
		}
	}
	if oldIPP == nil || // create the resource if it didn't exist before
		oldIPP.TTL != ipp.TTL ||
		oldIPP.Stored != ipp.Stored && oldIPP.Stored { // reset the resource if the profile changed this fields
		err = dm.SetIPAllocations(ctx, &utils.IPAllocations{
			Tenant:      ipp.Tenant,
			ID:          ipp.ID,
			Allocations: make(map[string]*utils.PoolAllocation),
		})
	} else if _, errRs := dm.GetIPAllocations(ctx, ipp.Tenant, ipp.ID, // do not try to get the resource if the configuration changed
		true, false, utils.NonTransactional); errRs == utils.ErrNotFound { // the resource does not exist
		err = dm.SetIPAllocations(ctx, &utils.IPAllocations{
			Tenant:      ipp.Tenant,
			ID:          ipp.ID,
			Allocations: make(map[string]*utils.PoolAllocation),
		})
	}
	return
}

func (dm *DataManager) RemoveIPProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldIPP, err := dm.GetIPProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveIPProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldIPP == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheIPFilterIndexes, tenant, id, oldIPP.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheIPFilterIndexes,
			tenant, utils.EmptyString, id, oldIPP.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaIPProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.IPProfilesPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return dm.RemoveIPAllocations(ctx, tenant, id)
}

func (dm *DataManager) HasData(category, subject, tenant string) (has bool, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	return dm.DataDB().HasDataDrv(context.TODO(), category, subject, tenant)
}

func (dm *DataManager) GetRouteProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rpp *utils.RouteProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRouteProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.RouteProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rpp, err = dm.dataDB.GetRouteProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaRouteProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns, utils.ReplicatorSv1GetRouteProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rpp); err == nil {
				err = dm.dataDB.SetRouteProfileDrv(ctx, rpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheRouteProfiles, tntID, nil, nil,
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
		if errCh := Cache.Set(ctx, utils.CacheRouteProfiles, tntID, rpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetRouteProfile(ctx *context.Context, rpp *utils.RouteProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, rpp.Tenant, rpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, rpp.TenantID())
		}
	}
	oldRpp, err := dm.GetRouteProfile(ctx, rpp.Tenant, rpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetRouteProfileDrv(ctx, rpp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheRouteFilterIndexes, rpp.Tenant,
			utils.EmptyString, rpp.ID, oldFiltersIDs, rpp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRouteProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RouteProfilePrefix, rpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRouteProfile,
			&utils.RouteProfileWithAPIOpts{
				RouteProfile: rpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRouteProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldRpp, err := dm.GetRouteProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveRouteProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheRouteFilterIndexes, tenant, id, oldRpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheRouteFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRouteProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RouteProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRouteProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// GetAttributeProfile returns the AttributeProfile with the given id
func (dm *DataManager) GetAttributeProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (attrPrfl *utils.AttributeProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheAttributeProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.AttributeProfile), nil
		}
	}
	if strings.HasPrefix(id, utils.Meta) {
		attrPrfl, err = utils.NewAttributeFromInline(tenant, id)
		return // do not set inline attributes in cache it breaks the interanal db matching
	} else if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	} else {
		if attrPrfl, err = dm.dataDB.GetAttributeProfileDrv(ctx, tenant, id); err != nil {
			if itm := dm.cfg.DataDbCfg().Items[utils.MetaAttributeProfiles]; err == utils.ErrNotFound && itm.Remote {
				if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
					utils.ReplicatorSv1GetAttributeProfile,
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
								dm.cfg.GeneralCfg().NodeID)),
					}, &attrPrfl); err == nil {
					err = dm.dataDB.SetAttributeProfileDrv(ctx, attrPrfl)
				}
			}
			if err != nil {
				err = utils.CastRPCErr(err)
				if err == utils.ErrNotFound && cacheWrite {
					if errCh := Cache.Set(ctx, utils.CacheAttributeProfiles, tntID, nil, nil,
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
		if errCh := Cache.Set(ctx, utils.CacheAttributeProfiles, tntID, attrPrfl, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetAttributeProfile(ctx *context.Context, ap *utils.AttributeProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, ap.Tenant, ap.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, ap.TenantID())
		}
	}
	oldAP, err := dm.GetAttributeProfile(ctx, ap.Tenant, ap.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	for _, attribute := range ap.Attributes {
		if attribute.Type == utils.MetaPassword {
			password := attribute.Value.GetRule()
			if password, err = utils.ComputeHash(password); err != nil {
				return
			}
			if attribute.Value, err = utils.NewRSRParsers(password, utils.RSRSep); err != nil {
				return
			}
			attribute.Type = utils.MetaConstant
		}
	}
	if err = dm.DataDB().SetAttributeProfileDrv(ctx, ap); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldAP != nil {
			oldFiltersIDs = &oldAP.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheAttributeFilterIndexes, ap.Tenant,
			utils.EmptyString, ap.ID, oldFiltersIDs, ap.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaAttributeProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.AttributeProfilePrefix, ap.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetAttributeProfile,
			&utils.AttributeProfileWithAPIOpts{
				AttributeProfile: ap,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveAttributeProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldAttr, err := dm.GetAttributeProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveAttributeProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldAttr == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheAttributeFilterIndexes, tenant, id, oldAttr.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheAttributeFilterIndexes,
			tenant, utils.EmptyString, id, oldAttr.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaAttributeProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.AttributeProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetChargerProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (cpp *utils.ChargerProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheChargerProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.ChargerProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	cpp, err = dm.dataDB.GetChargerProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaChargerProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetChargerProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &cpp); err == nil {
				err = dm.dataDB.SetChargerProfileDrv(ctx, cpp)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheChargerProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheChargerProfiles, tntID, cpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetChargerProfile(ctx *context.Context, cpp *utils.ChargerProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, cpp.Tenant, cpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, cpp.TenantID())
		}
	}
	oldCpp, err := dm.GetChargerProfile(ctx, cpp.Tenant, cpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetChargerProfileDrv(ctx, cpp); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldCpp != nil {
			oldFiltersIDs = &oldCpp.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheChargerFilterIndexes, cpp.Tenant,
			utils.EmptyString, cpp.ID, oldFiltersIDs, cpp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaChargerProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ChargerProfilePrefix, cpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetChargerProfile,
			&utils.ChargerProfileWithAPIOpts{
				ChargerProfile: cpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveChargerProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldCpp, err := dm.GetChargerProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveChargerProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldCpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheChargerFilterIndexes, tenant, id, oldCpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheChargerFilterIndexes,
			tenant, utils.EmptyString, id, oldCpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaChargerProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ChargerProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveChargerProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetItemLoadIDs(ctx *context.Context, itemIDPrefix string, cacheWrite bool) (loadIDs map[string]int64, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	loadIDs, err = dm.DataDB().GetItemLoadIDsDrv(ctx, itemIDPrefix)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaLoadIDs]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetItemLoadIDs,
				&utils.StringWithAPIOpts{
					Arg:    itemIDPrefix,
					Tenant: dm.cfg.GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &loadIDs); err == nil {
				err = dm.dataDB.SetLoadIDsDrv(ctx, loadIDs)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				for key := range loadIDs {
					if errCh := Cache.Set(ctx, utils.CacheLoadIDs, key, nil, nil,
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
			if errCh := Cache.Set(ctx, utils.CacheLoadIDs, key, val, nil,
				cacheCommit(utils.NonTransactional), utils.NonTransactional); errCh != nil {
				return nil, errCh
			}
		}
	}
	return
}

// SetLoadIDs sets the loadIDs in the DB
func (dm *DataManager) SetLoadIDs(ctx *context.Context, loadIDs map[string]int64) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetLoadIDsDrv(ctx, loadIDs); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaLoadIDs]; itm.Replicate {
		objIDs := make([]string, 0, len(loadIDs))
		for k := range loadIDs {
			objIDs = append(objIDs, k)
		}
		err = replicateMultipleIDs(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.LoadIDPrefix, objIDs, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetLoadIDs,
			&utils.LoadIDsWithAPIOpts{
				LoadIDs: loadIDs,
				Tenant:  dm.cfg.GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetRateProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (rpp *utils.RateProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheRateProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.RateProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	rpp, err = dm.dataDB.GetRateProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaRateProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetRateProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &rpp); err == nil {
				rpp.Sort()
				err = dm.dataDB.SetRateProfileDrv(ctx, rpp, false)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheRateProfiles, tntID, nil, nil,
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
		if errCh := Cache.Set(ctx, utils.CacheRateProfiles, tntID, rpp, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) GetRateProfileRates(ctx *context.Context, args *utils.ArgsSubItemIDs, needIDs bool) (rateIDs []string, rates []*utils.Rate, err error) {
	if dm == nil {
		return nil, nil, utils.ErrNoDatabaseConn
	}
	return dm.DataDB().GetRateProfileRatesDrv(ctx, args.Tenant, args.ProfileID, args.ItemsPrefix, needIDs)
}

func (dm *DataManager) SetRateProfile(ctx *context.Context, rpp *utils.RateProfile, optOverwrite, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	// check if the filters are valid, this can be inline or filter object
	if len(rpp.FilterIDs) != 0 {
		if err := dm.checkFilters(ctx, rpp.Tenant, rpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, rpp.TenantID())
		}
		for _, rate := range rpp.Rates {
			if err := dm.checkFilters(ctx, rpp.Tenant, rate.FilterIDs); err != nil {
				// if we get a broken filter do not update the rates
				return fmt.Errorf("%+s for item with ID: %+v",
					err, rate.ID)
			}
		}
	}
	rpp.Sort()
	// get the old RateProfile in case of updating fields
	oldRpp, err := dm.GetRateProfile(ctx, rpp.Tenant, rpp.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}

	if withIndex {
		// remove indexes for old rates if the rates does not exist in new  rate profile
		var oldRpFltrs *[]string
		if oldRpp != nil {
			oldRpFltrs = &oldRpp.FilterIDs
		}
		// create index for our profile
		if err := updatedIndexes(ctx, dm, utils.CacheRateProfilesFilterIndexes, rpp.Tenant,
			utils.EmptyString, rpp.ID, oldRpFltrs, rpp.FilterIDs, false); err != nil {
			return err
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
			if err := updatedIndexes(ctx, dm, utils.CacheRateFilterIndexes, rpp.Tenant,
				rpp.ID, key, oldRateFiltersIDs, rate.FilterIDs, true); err != nil {
				return err
			}
		}
	}
	// if not overwriting, we will add the rates in case the profile is already in database, also the fields of the profile are changed too in case of the same tenantID
	if err = dm.DataDB().SetRateProfileDrv(ctx, rpp, optOverwrite); err != nil {
		return err
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RateProfilePrefix, rpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRateProfile,
			&utils.RateProfileWithAPIOpts{
				RateProfile: rpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRateProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	var oldRpp *utils.RateProfile
	oldRpp, err = dm.GetRateProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return
	}
	if err = dm.DataDB().RemoveRateProfileDrv(ctx, tenant, id, nil); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		for key, rate := range oldRpp.Rates {
			if err = removeIndexFiltersItem(ctx, dm, utils.CacheRateFilterIndexes, tenant, utils.ConcatenatedKey(key, oldRpp.ID), rate.FilterIDs); err != nil {
				return
			}
			if err = removeItemFromFilterIndex(ctx, dm, utils.CacheRateFilterIndexes,
				oldRpp.Tenant, oldRpp.ID, key, rate.FilterIDs); err != nil {
				return
			}
		}
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheRateProfilesFilterIndexes, tenant, id, oldRpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheRateProfilesFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RateProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRateProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRateProfileRates(ctx *context.Context, tenant, id string, rateIDs *[]string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldRpp, err := dm.GetRateProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil {
		return err
	}
	if rateIDs == nil {
		if withIndex {
			for key, rate := range oldRpp.Rates {
				if err = removeItemFromFilterIndex(ctx, dm, utils.CacheRateFilterIndexes,
					tenant, id, key, rate.FilterIDs); err != nil {
					return
				}
			}
		}
		oldRpp.Rates = map[string]*utils.Rate{}
	} else {
		for _, rateID := range *rateIDs {
			if _, has := oldRpp.Rates[rateID]; !has {
				continue
			}
			if withIndex {
				for key, rate := range oldRpp.Rates {
					if err = removeIndexFiltersItem(ctx, dm, utils.CacheRateFilterIndexes,
						tenant, utils.ConcatenatedKey(key, oldRpp.ID), rate.FilterIDs); err != nil {
						return
					}
					if err = removeItemFromFilterIndex(ctx, dm, utils.CacheRateFilterIndexes,
						tenant, id, rateID, rate.FilterIDs); err != nil {
						return
					}
				}
			}
			delete(oldRpp.Rates, rateID)
		}
	}
	if err = dm.DataDB().RemoveRateProfileDrv(ctx, tenant, id, rateIDs); err != nil {
		return
	}

	if itm := dm.cfg.DataDbCfg().Items[utils.MetaRateProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.RateProfilePrefix, oldRpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRateProfile,
			&utils.RateProfileWithAPIOpts{
				RateProfile: oldRpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) GetActionProfile(ctx *context.Context, tenant, id string, cacheRead, cacheWrite bool,
	transactionID string) (ap *utils.ActionProfile, err error) {
	tntID := utils.ConcatenatedKey(tenant, id)
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheActionProfiles, tntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*utils.ActionProfile), nil
		}
	}
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ap, err = dm.dataDB.GetActionProfileDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaActionProfiles]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetActionProfile,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &ap); err == nil {
				err = dm.dataDB.SetActionProfileDrv(ctx, ap)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(ctx, utils.CacheActionProfiles, tntID, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}

			}
			return nil, err
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(ctx, utils.CacheActionProfiles, tntID, ap, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) SetActionProfile(ctx *context.Context, ap *utils.ActionProfile, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, ap.Tenant, ap.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, ap.TenantID())
		}
	}
	oldRpp, err := dm.GetActionProfile(ctx, ap.Tenant, ap.ID, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetActionProfileDrv(ctx, ap); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheActionProfilesFilterIndexes, ap.Tenant,
			utils.EmptyString, ap.ID, oldFiltersIDs, ap.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaActionProfiles]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ActionProfilePrefix, ap.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetActionProfile,
			&utils.ActionProfileWithAPIOpts{
				ActionProfile: ap,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveActionProfile(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldAct, err := dm.GetActionProfile(ctx, tenant, id, true, false, utils.NonTransactional)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveActionProfileDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldAct == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheActionProfilesFilterIndexes, tenant, id, oldAct.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheActionProfilesFilterIndexes,
			tenant, utils.EmptyString, id, oldAct.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaActionProfiles]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.ActionProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveActionProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// Reconnect reconnects to the DB when the config was changed
func (dm *DataManager) Reconnect(d DataDB) {
	// ToDo: consider locking
	dm.dataDB.Close()
	dm.dataDB = d
}

func (dm *DataManager) GetIndexes(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string,
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
	if indexes, err = dm.DataDB().GetIndexesDrv(ctx, idxItmType, tntCtx, idxKey, transactionID); err != nil {
		if itm := dm.cfg.DataDbCfg().Items[idxItmType]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetIndexes,
				&utils.GetIndexesArg{
					IdxItmType: idxItmType,
					TntCtx:     tntCtx,
					IdxKey:     idxKey,
					Tenant:     dm.cfg.GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &indexes); err == nil {
				err = dm.dataDB.SetIndexesDrv(ctx, idxItmType, tntCtx, indexes, true, utils.NonTransactional)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite &&
				idxKey != utils.EmptyString {
				if errCh := Cache.Set(ctx, idxItmType, utils.ConcatenatedKey(tntCtx, idxKey), nil, []string{tntCtx},
					true, utils.NonTransactional); errCh != nil {
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	if cacheWrite {
		for k, v := range indexes {
			if err = Cache.Set(ctx, idxItmType, utils.ConcatenatedKey(tntCtx, k), v, []string{tntCtx},
				true, utils.NonTransactional); err != nil {
				return nil, err
			}
		}
	}
	return
}

func (dm *DataManager) SetIndexes(ctx *context.Context, idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetIndexesDrv(ctx, idxItmType, tntCtx,
		indexes, commit, transactionID); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[idxItmType]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.CacheInstanceToPrefix[idxItmType], tntCtx, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetIndexes,
			&utils.SetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				Indexes:    indexes,
				Tenant:     dm.cfg.GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveIndexes(ctx *context.Context, idxItmType, tntCtx, idxKey string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveIndexesDrv(ctx, idxItmType, tntCtx, idxKey); err != nil {
		return
	}
	if itm := dm.cfg.DataDbCfg().Items[idxItmType]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.CacheInstanceToPrefix[idxItmType], tntCtx, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveIndexes,
			&utils.GetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				IdxKey:     idxKey,
				Tenant:     dm.cfg.GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func GetAPIBan(ctx *context.Context, ip string, apiKeys []string, single, cacheRead, cacheWrite bool) (banned bool, err error) {
	if cacheRead {
		if x, ok := Cache.Get(utils.MetaAPIBan, ip); ok && x != nil { // Attempt to find in cache first
			return x.(bool), nil
		}
	}
	if single {
		if banned, err = baningo.CheckIP(ctx, ip, apiKeys...); err != nil {
			return
		}
		if cacheWrite {
			if err = Cache.Set(ctx, utils.MetaAPIBan, ip, banned, nil, true, utils.NonTransactional); err != nil {
				return false, err
			}
		}
		return
	}
	var bannedIPs []string
	if bannedIPs, err = baningo.GetBannedIPs(ctx, apiKeys...); err != nil {
		return
	}
	for _, bannedIP := range bannedIPs {
		if bannedIP == ip {
			banned = true
		}
		if cacheWrite {
			if err = Cache.Set(ctx, utils.MetaAPIBan, bannedIP, true, nil, true, utils.NonTransactional); err != nil {
				return false, err
			}
		}
	}
	if len(ip) != 0 && !banned && cacheWrite {
		if err = Cache.Set(ctx, utils.MetaAPIBan, ip, false, nil, true, utils.NonTransactional); err != nil {
			return false, err
		}
	}
	return
}

// checkFilters returns the id of the first Filter that is not valid
// it should be called after the dm nil check
func (dm *DataManager) checkFilters(ctx *context.Context, tenant string, ids []string) (err error) {
	for _, id := range ids {
		// in case of inline filter we try to build them
		// if they are not correct it should fail here not in indexes
		if strings.HasPrefix(id, utils.Meta) {
			if fltr, err := NewFilterFromInline(tenant, id); err != nil {
				return fmt.Errorf("broken reference to filter: <%s>", id)
			} else if err := CheckFilter(fltr); err != nil {
				return err
			}
		} else if x, has := Cache.Get(utils.CacheFilters, // because the method HasDataDrv doesn't use cache
			utils.ConcatenatedKey(tenant, id)); has && x == nil { // check to see if filter is already in cache
			return fmt.Errorf("broken reference to filter: <%s>", id)
		} else if has, err := dm.DataDB().HasDataDrv(ctx, utils.FilterPrefix, // check in local DB if we have the filter
			id, tenant); err != nil || !has {
			// in case we can not find it localy try to find it in the remote DB
			if itm := dm.cfg.DataDbCfg().Items[utils.MetaFilters]; err == utils.ErrNotFound && itm.Remote {
				var fltr *Filter
				err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns, utils.ReplicatorSv1GetFilter,
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
								dm.cfg.GeneralCfg().NodeID)),
					}, &fltr)
				has = fltr == nil
			}
			// not in local DB and not in remote DB
			if err != nil || !has {
				return fmt.Errorf("broken reference to filter: <%s>", id)
			}
		}
	}
	return
}

func (dm *DataManager) GetAccount(ctx *context.Context, tenant, id string) (ap *utils.Account, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	ap, err = dm.dataDB.GetAccountDrv(ctx, tenant, id)
	if err != nil {
		if itm := dm.cfg.DataDbCfg().Items[utils.MetaAccounts]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(ctx, dm.cfg.DataDbCfg().RmtConns,
				utils.ReplicatorSv1GetAccount,
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(dm.cfg.DataDbCfg().RmtConnID,
							dm.cfg.GeneralCfg().NodeID)),
				}, &ap); err == nil {
				err = dm.dataDB.SetAccountDrv(ctx, ap)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetAccount(ctx *context.Context, ap *utils.Account, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err := dm.checkFilters(ctx, ap.Tenant, ap.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, ap.TenantID())
		}
	}
	oldRpp, err := dm.GetAccount(ctx, ap.Tenant, ap.ID)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().SetAccountDrv(ctx, ap); err != nil {
		return err
	}
	if withIndex {
		var oldFiltersIDs *[]string
		if oldRpp != nil {
			oldFiltersIDs = &oldRpp.FilterIDs
		}
		if err := updatedIndexes(ctx, dm, utils.CacheAccountsFilterIndexes, ap.Tenant,
			utils.EmptyString, ap.ID, oldFiltersIDs, ap.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		err = replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.AccountPrefix, ap.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: ap,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveAccount(ctx *context.Context, tenant, id string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	oldRpp, err := dm.GetAccount(ctx, tenant, id)
	if err != nil && err != utils.ErrNotFound {
		return err
	}
	if err = dm.DataDB().RemoveAccountDrv(ctx, tenant, id); err != nil {
		return
	}
	if oldRpp == nil {
		return utils.ErrNotFound
	}
	if withIndex {
		if err = removeIndexFiltersItem(ctx, dm, utils.CacheAccountsFilterIndexes, tenant, id, oldRpp.FilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(ctx, dm, utils.CacheAccountsFilterIndexes,
			tenant, utils.EmptyString, id, oldRpp.FilterIDs); err != nil {
			return
		}
	}
	if itm := dm.cfg.DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		replicate(ctx, dm.connMgr, dm.cfg.DataDbCfg().RplConns,
			dm.cfg.DataDbCfg().RplFiltered,
			utils.AccountPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					dm.cfg.DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}
