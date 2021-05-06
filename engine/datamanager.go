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
		utils.AttributeFilterIndexes:  {},
		utils.ResourceFilterIndexes:   {},
		utils.StatFilterIndexes:       {},
		utils.ThresholdFilterIndexes:  {},
		utils.RouteFilterIndexes:      {},
		utils.ChargerFilterIndexes:    {},
		utils.DispatcherFilterIndexes: {},
		utils.ActionPlanIndexes:       {},
		utils.FilterIndexPrfx:         {},
	}
	cachePrefixMap = utils.StringSet{
		utils.DestinationPrefix:        {},
		utils.ReverseDestinationPrefix: {},
		utils.RatingPlanPrefix:         {},
		utils.RatingProfilePrefix:      {},
		utils.ActionPrefix:             {},
		utils.ActionPlanPrefix:         {},
		utils.AccountActionPlansPrefix: {},
		utils.ActionTriggerPrefix:      {},
		utils.SharedGroupPrefix:        {},
		utils.ResourceProfilesPrefix:   {},
		utils.TimingsPrefix:            {},
		utils.ResourcesPrefix:          {},
		utils.StatQueuePrefix:          {},
		utils.StatQueueProfilePrefix:   {},
		utils.ThresholdPrefix:          {},
		utils.ThresholdProfilePrefix:   {},
		utils.FilterPrefix:             {},
		utils.RouteProfilePrefix:       {},
		utils.AttributeProfilePrefix:   {},
		utils.ChargerProfilePrefix:     {},
		utils.DispatcherProfilePrefix:  {},
		utils.DispatcherHostPrefix:     {},
		utils.AttributeFilterIndexes:   {},
		utils.ResourceFilterIndexes:    {},
		utils.StatFilterIndexes:        {},
		utils.ThresholdFilterIndexes:   {},
		utils.RouteFilterIndexes:       {},
		utils.ChargerFilterIndexes:     {},
		utils.DispatcherFilterIndexes:  {},
		utils.FilterIndexPrfx:          {},
		utils.MetaAPIBan:               {}, // not realy a prefix as this is not stored in DB
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
	} else if len(ids) != 0 && ids[0] == utils.MetaAny {
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
		case utils.DestinationPrefix:
			_, err = dm.GetDestination(dataID, false, true, utils.NonTransactional)
		case utils.ReverseDestinationPrefix:
			_, err = dm.GetReverseDestination(dataID, false, true, utils.NonTransactional)
		case utils.RatingPlanPrefix:
			_, err = dm.GetRatingPlan(dataID, true, utils.NonTransactional)
		case utils.RatingProfilePrefix:
			_, err = dm.GetRatingProfile(dataID, true, utils.NonTransactional)
		case utils.ActionPrefix:
			_, err = dm.GetActions(dataID, true, utils.NonTransactional)
		case utils.ActionPlanPrefix:
			_, err = dm.GetActionPlan(dataID, true, utils.NonTransactional)
		case utils.AccountActionPlansPrefix:
			_, err = dm.GetAccountActionPlans(dataID, true, utils.NonTransactional)
		case utils.ActionTriggerPrefix:
			_, err = dm.GetActionTriggers(dataID, true, utils.NonTransactional)
		case utils.SharedGroupPrefix:
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

func (dm *DataManager) RebuildReverseForPrefix(prefix string) (err error) {
	switch prefix {
	case utils.ReverseDestinationPrefix:
		if err = dm.dataDB.RemoveKeysForPrefix(prefix); err != nil {
			return
		}
		var keys []string
		if keys, err = dm.dataDB.GetKeysForPrefix(utils.DestinationPrefix); err != nil {
			return
		}
		for _, key := range keys {
			var dest *Destination
			if dest, err = dm.GetDestination(key[len(utils.DestinationPrefix):], false, true, utils.NonTransactional); err != nil {
				return err
			}
			if err = dm.SetReverseDestination(dest.Id, dest.Prefixes, utils.NonTransactional); err != nil {
				return err
			}
		}
	case utils.AccountActionPlansPrefix:
		if err = dm.dataDB.RemoveKeysForPrefix(prefix); err != nil {
			return
		}
		var keys []string
		if keys, err = dm.dataDB.GetKeysForPrefix(utils.ActionPlanPrefix); err != nil {
			return
		}
		for _, key := range keys {
			var apl *ActionPlan
			if apl, err = dm.GetActionPlan(key[len(utils.ActionPlanPrefix):], true, utils.NonTransactional); err != nil {
				return err
			}
			for acntID := range apl.AccountIDs {
				if err = dm.SetAccountActionPlans(acntID, []string{apl.Id}, false); err != nil {
					return err
				}
			}
		}
	default:
		return utils.ErrInvalidKey
	}
	return
}

func (dm *DataManager) GetDestination(key string, cacheRead, cacheWrite bool, transactionID string) (dest *Destination, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheDestinations, key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Destination), nil
		}
	}
	dest, err = dm.dataDB.GetDestinationDrv(key, transactionID)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetDestination, &utils.StringWithAPIOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &dest); err == nil {
				err = dm.dataDB.SetDestinationDrv(dest, utils.NonTransactional)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheDestinations, key, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return
		}
	}
	if cacheWrite {
		if err := Cache.Set(utils.CacheDestinations, key, dest, nil,
			cacheCommit(transactionID), transactionID); err != nil {
			return nil, err
		}
	}
	return
}

func (dm *DataManager) SetDestination(dest *Destination, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.SetDestinationDrv(dest, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DestinationPrefix, dest.Id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetDestination,
			&DestinationWithAPIOpts{
				Destination: dest,
				Tenant:      config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveDestination(destID string, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}

	var oldDst *Destination
	if oldDst, err = dm.GetDestination(destID, true, false,
		transactionID); err != nil && err != utils.ErrNotFound {
		return
	}

	if err = dm.dataDB.RemoveDestinationDrv(destID, transactionID); err != nil {
		return
	}
	if err = Cache.Remove(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID); err != nil {
		return
	}
	if oldDst == nil {
		return utils.ErrNotFound
	}
	for _, prfx := range oldDst.Prefixes {
		if err = dm.dataDB.RemoveReverseDestinationDrv(destID, prfx, transactionID); err != nil {
			return
		}
		dm.GetReverseDestination(prfx, false, true, transactionID) // it will recache the destination
	}

	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDestinations]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DestinationPrefix, destID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveDestination,
			&utils.StringWithAPIOpts{
				Arg:    destID,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) SetReverseDestination(destID string, prefixes []string, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.SetReverseDestinationDrv(destID, prefixes, transactionID); err != nil {
		return
	}
	if config.CgrConfig().DataDbCfg().Items[utils.MetaReverseDestinations].Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DestinationPrefix, destID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetReverseDestination,
			&DestinationWithAPIOpts{Destination: &Destination{Id: destID, Prefixes: prefixes}})
	}
	return
}

func (dm *DataManager) GetReverseDestination(prefix string,
	cacheRead, cacheWrite bool, transactionID string) (ids []string, err error) {
	if dm == nil {
		err = utils.ErrNoDatabaseConn
		return
	}
	if cacheRead {
		if x, ok := Cache.Get(utils.CacheReverseDestinations, prefix); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
	}
	ids, err = dm.dataDB.GetReverseDestinationDrv(prefix, transactionID)
	if err != nil {
		if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaReverseDestinations]; err == utils.ErrNotFound && itm.Remote {
			if err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil,
				utils.ReplicatorSv1GetReverseDestination, &utils.StringWithAPIOpts{
					Arg:    prefix,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &ids); err == nil {
				err = dm.dataDB.SetReverseDestinationDrv(prefix, ids, transactionID)
			}
		}
		if err != nil {
			err = utils.CastRPCErr(err)
			if err == utils.ErrNotFound && cacheWrite {
				if errCh := Cache.Set(utils.CacheReverseDestinations, prefix, nil, nil,
					cacheCommit(transactionID), transactionID); errCh != nil {
					return nil, errCh
				}
			}
			return
		}
	}
	if cacheWrite {
		if errCh := Cache.Set(utils.CacheReverseDestinations, prefix, ids, nil,
			cacheCommit(transactionID), transactionID); errCh != nil {
			return nil, errCh
		}
	}
	return
}

func (dm *DataManager) UpdateReverseDestination(oldDest, newDest *Destination,
	transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if oldDest == nil {
		return dm.dataDB.SetReverseDestinationDrv(newDest.Id, newDest.Prefixes, transactionID)
	}

	cCommit := cacheCommit(transactionID)
	var addedPrefixes []string
	for _, oldPrefix := range oldDest.Prefixes {
		var found bool
		for _, newPrefix := range newDest.Prefixes {
			if oldPrefix == newPrefix {
				found = true
				break
			}
		}
		if !found {
			if err = dm.dataDB.RemoveReverseDestinationDrv(newDest.Id, oldPrefix, transactionID); err != nil {
				return
			}
			if err = Cache.Remove(utils.CacheReverseDestinations, oldPrefix,
				cCommit, transactionID); err != nil {
				return
			}
		}
	}

	for _, newPrefix := range newDest.Prefixes {
		var found bool
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
	return dm.SetReverseDestination(newDest.Id, addedPrefixes, transactionID)
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
				utils.ReplicatorSv1GetAccount, &utils.StringWithAPIOpts{
					Arg:    id,
					Tenant: tenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.SetAccountDrv(acc); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AccountPrefix, acc.ID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetAccount,
			&AccountWithAPIOpts{
				Account: acc,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					utils.EmptyString, utils.EmptyString)}) // the account doesn't have cache
	}
	return
}

func (dm *DataManager) RemoveAccount(id string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.RemoveAccountDrv(id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccounts]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AccountPrefix, id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveAccount,
			&utils.StringWithAPIOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					utils.EmptyString, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &sq); err == nil {
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
		return utils.ErrNoDatabaseConn
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
			// if the statQueue didn't exists simply initiate all the metrics
			if sq, err = NewStatQueue(tnt, id, metrics, minItems); err != nil {
				return
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
				}
				if needsRemove {
					delete(sq.SQMetrics, sqMetricID)
				}
			}

			for _, metric := range metrics {
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
	if dm.dataDB.GetStorageType() != utils.MetaInternal {
		// in case of internal we don't marshal
		if ssq, err = NewStoredStatQueue(sq, dm.ms); err != nil {
			return
		}
	}
	if err = dm.dataDB.SetStatQueueDrv(ssq, sq); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.StatQueuePrefix, sq.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetStatQueue,
			&StatQueueWithAPIOpts{
				StatQueue: sq,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// RemoveStatQueue removes the StoredStatQueue
func (dm *DataManager) RemoveStatQueue(tenant, id string, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.RemStatQueueDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueues]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.StatQueuePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveStatQueue,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
								config.CgrConfig().GeneralCfg().NodeID)),
					}, &fltr); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if err = CheckFilter(fltr); err != nil {
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
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.FilterPrefix, fltr.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetFilter,
			&FilterWithAPIOpts{
				Filter: fltr,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.FilterPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveFilter,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				utils.ReplicatorSv1GetThreshold, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &th); err == nil {
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
		return utils.ErrNoDatabaseConn
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
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ThresholdPrefix, th.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetThreshold,
			&ThresholdWithAPIOpts{
				Threshold: th,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholds]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ThresholdPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveThreshold,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &th); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(th.Tenant, th.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, th.TenantID())
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
			utils.EmptyString, th.ID, oldFiltersIDs, th.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaThresholdProfiles]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ThresholdProfilePrefix, th.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetThresholdProfile,
			&ThresholdProfileWithAPIOpts{
				ThresholdProfile: th,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveThresholdProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ThresholdProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &sqp); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(sqp.Tenant, sqp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, sqp.TenantID())
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
			utils.EmptyString, sqp.ID, oldFiltersIDs, sqp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaStatQueueProfiles]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.StatQueueProfilePrefix, sqp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetStatQueueProfile,
			&StatQueueProfileWithAPIOpts{
				StatQueueProfile: sqp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveStatQueueProfile(tenant, id,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.StatQueueProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveStatQueueProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.StringWithAPIOpts{
					Arg:    id,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetTimingDrv(t); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.TimingsPrefix, []string{t.ID}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaTimings]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.TimingsPrefix, t.ID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetTiming,
			&utils.TPTimingWithAPIOpts{
				TPTiming: t,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveTiming(id, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveTimingDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheTimings, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if config.CgrConfig().DataDbCfg().Items[utils.MetaTimings].Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.TimingsPrefix, id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveTiming,
			id)
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &rs); err == nil {
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
		return utils.ErrNoDatabaseConn
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
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ResourcesPrefix, rs.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetResource,
			&ResourceWithAPIOpts{
				Resource: rs,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveResource(tenant, id, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveResourceDrv(tenant, id); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResources]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ResourcesPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveResource,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				utils.ReplicatorSv1GetResourceProfile, &utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &rp); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(rp.Tenant, rp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, rp.TenantID())
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
			utils.EmptyString, rp.ID, oldFiltersIDs, rp.FilterIDs, false); err != nil {
			return err
		}
		Cache.Clear([]string{utils.CacheEventResources})
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaResourceProfile]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ResourceProfilesPrefix, rp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetResourceProfile,
			&ResourceProfileWithAPIOpts{
				ResourceProfile: rp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveResourceProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ResourceProfilesPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveResourceProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.StringWithAPIOpts{
					Arg:    id,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheActionTriggers, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionTriggers]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionTriggerPrefix, id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveActionTriggers,
			&utils.StringWithAPIOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

//SetActionTriggersArgWithAPIOpts is used to send the key and the ActionTriggers to Replicator
type SetActionTriggersArgWithAPIOpts struct {
	Key     string
	Attrs   ActionTriggers
	Tenant  string
	APIOpts map[string]interface{}
}

func (dm *DataManager) SetActionTriggers(key string, attr ActionTriggers,
	transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetActionTriggersDrv(key, attr); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.ActionTriggerPrefix, []string{key}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionTriggers]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionTriggerPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetActionTriggers,
			&SetActionTriggersArgWithAPIOpts{
				Attrs:  attr,
				Key:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				utils.ReplicatorSv1GetSharedGroup, &utils.StringWithAPIOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetSharedGroupDrv(sg); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.SharedGroupPrefix,
		[]string{sg.Id}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaSharedGroups]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.SharedGroupPrefix, sg.Id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetSharedGroup,
			&SharedGroupWithAPIOpts{
				SharedGroup: sg,
				Tenant:      config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveSharedGroup(id, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveSharedGroupDrv(id); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheSharedGroups, id,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaSharedGroups]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.SharedGroupPrefix, id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveSharedGroup,
			&utils.StringWithAPIOpts{
				Arg:    id,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				utils.ReplicatorSv1GetActions, &utils.StringWithAPIOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
type SetActionsArgsWithAPIOpts struct {
	Key     string
	Acs     Actions
	Tenant  string
	APIOpts map[string]interface{}
}

func (dm *DataManager) SetActions(key string, as Actions, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetActionsDrv(key, as); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActions]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetActions,
			&SetActionsArgsWithAPIOpts{
				Key:    key,
				Acs:    as,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveActions(key, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveActionsDrv(key); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActions]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveActions,
			&utils.StringWithAPIOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
			utils.ReplicatorSv1GetActionPlan, &utils.StringWithAPIOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
					utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
						config.CgrConfig().GeneralCfg().NodeID)),
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

// SetActionPlanArgWithAPIOpts is used in replicatorV1 for dispatcher
type SetActionPlanArgWithAPIOpts struct {
	Key       string
	Ats       *ActionPlan
	Overwrite bool
	Tenant    string
	APIOpts   map[string]interface{}
}

func (dm *DataManager) SetActionPlan(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.SetActionPlanDrv(key, ats, overwrite, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionPlanPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetActionPlan,
			&SetActionPlanArgWithAPIOpts{
				Key:       key,
				Ats:       ats,
				Overwrite: overwrite,
				Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
			&utils.StringWithAPIOpts{
				Arg:    utils.EmptyString,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
					utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
						config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.RemoveActionPlanDrv(key, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaActionPlans]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ActionPlanPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveActionPlan,
			&utils.StringWithAPIOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
			&utils.StringWithAPIOpts{
				Arg:    acntID,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
					utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
						config.CgrConfig().GeneralCfg().NodeID)),
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

//SetAccountActionPlansArgWithAPIOpts is used to send the key and the Actions to replicator
type SetAccountActionPlansArgWithAPIOpts struct {
	AcntID    string
	AplIDs    []string
	Overwrite bool
	Tenant    string
	APIOpts   map[string]interface{}
}

func (dm *DataManager) SetAccountActionPlans(acntID string, aPlIDs []string, overwrite bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.SetAccountActionPlansDrv(acntID, aPlIDs, overwrite); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccountActionPlans]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AccountActionPlansPrefix, acntID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetAccountActionPlans,
			&SetAccountActionPlansArgWithAPIOpts{
				AcntID:    acntID,
				AplIDs:    aPlIDs,
				Overwrite: overwrite,
				Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// RemAccountActionPlansArgsWithAPIOpts is used in replicatorV1 for dispatcher
type RemAccountActionPlansArgsWithAPIOpts struct {
	AcntID  string
	ApIDs   []string
	Tenant  string
	APIOpts map[string]interface{}
}

func (dm *DataManager) RemAccountActionPlans(acntID string, apIDs []string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.dataDB.RemAccountActionPlansDrv(acntID, apIDs); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaAccountActionPlans]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AccountActionPlansPrefix, acntID, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemAccountActionPlans,
			&RemAccountActionPlansArgsWithAPIOpts{
				AcntID: acntID,
				ApIDs:  apIDs,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.StringWithAPIOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetRatingPlanDrv(rp); err != nil {
		return
	}
	if err = dm.CacheDataFromDB(utils.RatingPlanPrefix, []string{rp.Id}, true); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingPlans]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RatingPlanPrefix, rp.Id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRatingPlan,
			&RatingPlanWithAPIOpts{
				RatingPlan: rp,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRatingPlan(key string, transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveRatingPlanDrv(key); err != nil {
		return
	}
	if errCh := Cache.Remove(utils.CacheRatingPlans, key,
		cacheCommit(transactionID), transactionID); errCh != nil {
		return errCh
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingPlans]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RatingPlanPrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRatingPlan,
			&utils.StringWithAPIOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.StringWithAPIOpts{
					Arg:    key,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetRatingProfileDrv(rpf); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingProfiles]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RatingProfilePrefix, rpf.Id, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRatingProfile,
			&RatingProfileWithAPIOpts{
				RatingProfile: rpf,
				Tenant:        config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRatingProfile(key string,
	transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveRatingProfileDrv(key); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRatingProfiles]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RatingProfilePrefix, key, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRatingProfile,
			&utils.StringWithAPIOpts{
				Arg:    key,
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &rpp); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(rpp.Tenant, rpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, rpp.TenantID())
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
			utils.EmptyString, rpp.ID, oldFiltersIDs, rpp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaRouteProfiles]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RouteProfilePrefix, rpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetRouteProfile,
			&RouteProfileWithAPIOpts{
				RouteProfile: rpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveRouteProfile(tenant, id, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.RouteProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveRouteProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
								config.CgrConfig().GeneralCfg().NodeID)),
					}, &attrPrfl); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(ap.Tenant, ap.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, ap.TenantID())
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
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AttributeProfilePrefix, ap.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetAttributeProfile,
			&AttributeProfileWithAPIOpts{
				AttributeProfile: ap,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveAttributeProfile(tenant, id string, transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.AttributeProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &cpp); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(cpp.Tenant, cpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, cpp.TenantID())
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
			utils.EmptyString, cpp.ID, oldFiltersIDs, cpp.FilterIDs, false); err != nil {
			return err
		}
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaChargerProfiles]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ChargerProfilePrefix, cpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetChargerProfile,
			&ChargerProfileWithAPIOpts{
				ChargerProfile: cpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveChargerProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.ChargerProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveChargerProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &dpp); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if withIndex {
		if err = dm.checkFilters(dpp.Tenant, dpp.FilterIDs); err != nil {
			// if we get a broken filter do not set the profile
			return fmt.Errorf("%+s for item with ID: %+v",
				err, dpp.TenantID())
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
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DispatcherProfilePrefix, dpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetDispatcherProfile,
			&DispatcherProfileWithAPIOpts{
				DispatcherProfile: dpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveDispatcherProfile(tenant, id string,
	transactionID string, withIndex bool) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DispatcherProfilePrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveDispatcherProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.TenantIDWithAPIOpts{
					TenantID: &utils.TenantID{Tenant: tenant, ID: id},
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
				}, &dH); err == nil {
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetDispatcherHostDrv(dpp); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaDispatcherHosts]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DispatcherHostPrefix, dpp.TenantID(), // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetDispatcherHost,
			&DispatcherHostWithAPIOpts{
				DispatcherHost: dpp,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveDispatcherHost(tenant, id string,
	transactionID string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
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
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.DispatcherHostPrefix, utils.ConcatenatedKey(tenant, id), // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveDispatcherHost,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{Tenant: tenant, ID: id},
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
				&utils.StringWithAPIOpts{
					Arg:    itemIDPrefix,
					Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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

// SetLoadIDs sets the loadIDs in the DB
func (dm *DataManager) SetLoadIDs(loadIDs map[string]int64) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetLoadIDsDrv(loadIDs); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaLoadIDs]; itm.Replicate {
		objIDs := make([]string, 0, len(loadIDs))
		for k := range loadIDs {
			objIDs = append(objIDs, k)
		}
		err = replicateMultipleIDs(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.LoadIDPrefix, objIDs, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetLoadIDs,
			&utils.LoadIDsWithAPIOpts{
				LoadIDs: loadIDs,
				Tenant:  config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

// Reconnect reconnects to the DB when the config was changed
func (dm *DataManager) Reconnect(marshaller string, newcfg *config.DataDbCfg) (err error) {
	d, err := NewDataDBConn(newcfg.Type, newcfg.Host, newcfg.Port, newcfg.Name,
		newcfg.User, newcfg.Password, marshaller, newcfg.Opts)
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
					APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
						utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
							config.CgrConfig().GeneralCfg().NodeID)),
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
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().SetIndexesDrv(idxItmType, tntCtx,
		indexes, commit, transactionID); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes]; itm.Replicate {
		err = replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.CacheInstanceToPrefix[idxItmType], tntCtx, // this are used to get the host IDs from cache
			utils.ReplicatorSv1SetIndexes,
			&utils.SetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				Indexes:    indexes,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
	}
	return
}

func (dm *DataManager) RemoveIndexes(idxItmType, tntCtx, idxKey string) (err error) {
	if dm == nil {
		return utils.ErrNoDatabaseConn
	}
	if err = dm.DataDB().RemoveIndexesDrv(idxItmType, tntCtx, idxKey); err != nil {
		return
	}
	if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes]; itm.Replicate {
		replicate(dm.connMgr, config.CgrConfig().DataDbCfg().RplConns,
			config.CgrConfig().DataDbCfg().RplFiltered,
			utils.CacheInstanceToPrefix[idxItmType], tntCtx, // this are used to get the host IDs from cache
			utils.ReplicatorSv1RemoveIndexes,
			&utils.GetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				IdxKey:     idxKey,
				Tenant:     config.CgrConfig().GeneralCfg().DefaultTenant,
				APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID,
					config.CgrConfig().DataDbCfg().RplCache, utils.EmptyString)})
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
func (dm *DataManager) checkFilters(tenant string, ids []string) (err error) {
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
		} else if has, err := dm.DataDB().HasDataDrv(utils.FilterPrefix, // check in local DB if we have the filter
			id, tenant); err != nil || !has {
			// in case we can not find it localy try to find it in the remote DB
			if itm := config.CgrConfig().DataDbCfg().Items[utils.MetaFilters]; err == utils.ErrNotFound && itm.Remote {
				var fltr *Filter
				err = dm.connMgr.Call(config.CgrConfig().DataDbCfg().RmtConns, nil, utils.ReplicatorSv1GetFilter,
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{Tenant: tenant, ID: id},
						APIOpts: utils.GenerateDBItemOpts(itm.APIKey, itm.RouteID, utils.EmptyString,
							utils.FirstNonEmpty(config.CgrConfig().DataDbCfg().RmtConnID,
								config.CgrConfig().GeneralCfg().NodeID)),
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
