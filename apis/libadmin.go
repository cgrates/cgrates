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

package apis

import (
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// CallCache caching the item based on cacheopt
// visible in APIerSv2
func (admS *AdminSv1) CallCache(ctx *context.Context, cacheopt string, tnt, cacheID, itemID string,
	filters *[]string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{}
	switch utils.FirstNonEmpty(cacheopt, admS.cfg.GeneralCfg().DefaultCaching) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
		var argCache map[string][]string
		if argCache, err = admS.composeArgsReload(ctx, tnt, cacheID, itemID, filters); err != nil {
			return
		}
		args = utils.NewAttrReloadCacheWithOptsFromMap(argCache, tnt, opts)
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
		var argCache map[string][]string
		if argCache, err = admS.composeArgsReload(ctx, tnt, cacheID, itemID, filters); err != nil {
			return
		}
		args = utils.NewAttrReloadCacheWithOptsFromMap(argCache, tnt, opts)
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
		var argCache map[string][]string
		if argCache, err = admS.composeArgsReload(ctx, tnt, cacheID, itemID, filters); err != nil {
			return
		}
		args = utils.NewAttrReloadCacheWithOptsFromMap(argCache, tnt, opts)
	case utils.MetaClear:
		cacheIDs := make([]string, 1, 3)
		cacheIDs[0] = cacheID
		// do not send a EmptyString if the item doesn't have indexes
		if cIdx, has := utils.CacheInstanceToCacheIndex[cacheID]; has {
			cacheIDs = append(cacheIDs, cIdx)
		}
		switch cacheID { // add the items to the cache reload
		case utils.CacheThresholdProfiles:
			cacheIDs = append(cacheIDs, utils.CacheThresholds)
		case utils.CacheResourceProfiles:
			cacheIDs = append(cacheIDs, utils.CacheResources)
		case utils.CacheStatQueueProfiles:
			cacheIDs = append(cacheIDs, utils.CacheStatQueues)
		}
		method = utils.CacheSv1Clear
		args = &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   tnt,
			CacheIDs: cacheIDs,
			APIOpts:  opts,
		}

	}
	return admS.connMgr.Call(ctx, admS.cfg.AdminSCfg().CachesConns,
		method, args, &reply)
}

// composeArgsReload add the ItemID to AttrReloadCache
// for a specific CacheID
func (admS *AdminSv1) composeArgsReload(ctx *context.Context, tnt, cacheID, itemID string, filterIDs *[]string) (argCache map[string][]string, err error) {
	argCache = map[string][]string{cacheID: {itemID}}
	switch cacheID { // add the items to the cache reload
	case utils.CacheThresholdProfiles:
		argCache[utils.CacheThresholds] = []string{itemID}
	case utils.CacheResourceProfiles:
		argCache[utils.CacheResources] = []string{itemID}
	case utils.CacheStatQueueProfiles:
		argCache[utils.CacheStatQueues] = []string{itemID}
	}
	if filterIDs == nil { // in case we remove a profile we do not need to reload the indexes
		return
	}
	// populate the indexes
	idxCacheID := utils.CacheInstanceToCacheIndex[cacheID]
	if len(*filterIDs) == 0 { // in case we do not have any filters reload the *none filter indexes
		indxID := utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny)
		argCache[idxCacheID] = []string{utils.ConcatenatedKey(tnt, indxID)}
		return
	}
	indxIDs := make([]string, 0, len(*filterIDs))
	for _, id := range *filterIDs {
		var fltr *engine.Filter
		if fltr, err = admS.dm.GetFilter(ctx, tnt, id, true, true, utils.NonTransactional); err != nil {
			return
		}
		for _, flt := range fltr.Rules {
			if !engine.FilterIndexTypes.Has(flt.Type) ||
				engine.IsDynamicDPPath(flt.Element) {
				continue
			}
			isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
			for _, fldVal := range flt.Values {
				if engine.IsDynamicDPPath(fldVal) {
					continue
				}
				if isDyn {
					if !strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
						indxIDs = append(indxIDs, utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal))
					}
				} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
					indxIDs = append(indxIDs, utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element))
				}
			}
		}
	}
	argCache[idxCacheID] = make([]string, len(indxIDs))
	for i, indxID := range indxIDs {
		argCache[idxCacheID][i] = utils.ConcatenatedKey(tnt, indxID)
	}
	return
}

// callCacheForIndexes will only call CacheClear because don't have access at ItemID
func (admS *AdminSv1) callCacheForRemoveIndexes(ctx *context.Context, cacheopt string, tnt, cacheID string,
	itemIDs []string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{} = utils.NewAttrReloadCacheWithOptsFromMap(map[string][]string{cacheID: itemIDs}, tnt, opts)
	switch utils.FirstNonEmpty(cacheopt, admS.cfg.GeneralCfg().DefaultCaching) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
	case utils.MetaClear:
		method = utils.CacheSv1Clear
		args = &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   tnt,
			CacheIDs: []string{cacheID},
			APIOpts:  opts,
		}
	}
	return admS.connMgr.Call(ctx, admS.cfg.AdminSCfg().CachesConns,
		method, args, &reply)
}

func (admS *AdminSv1) callCacheForComputeIndexes(ctx *context.Context, cacheopt, tnt string,
	cacheItems map[string][]string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{} = utils.NewAttrReloadCacheWithOptsFromMap(cacheItems, tnt, opts)
	switch utils.FirstNonEmpty(cacheopt, admS.cfg.GeneralCfg().DefaultCaching) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
	case utils.MetaClear:
		method = utils.CacheSv1Clear
		cacheIDs := make([]string, 0, len(cacheItems))
		for idx := range cacheItems {
			cacheIDs = append(cacheIDs, idx)
		}
		args = &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   tnt,
			CacheIDs: cacheIDs,
			APIOpts:  opts,
		}
	}
	return admS.connMgr.Call(ctx, admS.cfg.AdminSCfg().CachesConns,
		method, args, &reply)
}

/*
// callCacheRevDestinations used for reverse destination, loadIDs and indexes replication
func (apierSv1 *AdminS) callCacheMultiple(cacheopt, tnt, cacheID string, itemIDs []string, opts map[string]interface{}) (err error) {
	if len(itemIDs) == 0 {
		return
	}
	var reply, method string
	var args interface{}
	switch utils.FirstNonEmpty(cacheopt, apierSv1.cfg.GeneralCfg().DefaultCaching) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
		args = utils.NewAttrReloadCacheWithOptsFromMap(map[string][]string{cacheID: itemIDs}, tnt, opts)
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
		args = utils.NewAttrReloadCacheWithOptsFromMap(map[string][]string{cacheID: itemIDs}, tnt, opts)
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
		args = utils.NewAttrReloadCacheWithOptsFromMap(map[string][]string{cacheID: itemIDs}, tnt, opts)
	case utils.MetaClear:
		method = utils.CacheSv1Clear
		args = &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   tnt,
			CacheIDs: []string{cacheID},
			APIOpts:  opts,
		}
	}
	return apierSv1.ConnMgr.Call(context.TODO(), apierSv1.cfg.ApierCfg().CachesConns,
		method, args, &reply)
}
*/

func composeCacheArgsForFilter(dm *engine.DataManager, ctx *context.Context, fltr *engine.Filter, tnt, tntID string, args map[string][]string) (_ map[string][]string, err error) {
	indxIDs := make([]string, 0, len(fltr.Rules))
	for _, flt := range fltr.Rules {
		if !engine.FilterIndexTypes.Has(flt.Type) ||
			engine.IsDynamicDPPath(flt.Element) {
			continue
		}
		isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
		for _, fldVal := range flt.Values {
			if engine.IsDynamicDPPath(fldVal) {
				continue
			}
			if isDyn {
				if !strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
					indxIDs = append(indxIDs, utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal))
				}
			} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
				indxIDs = append(indxIDs, utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element))
			}
		}
	}
	if len(indxIDs) == 0 { // no index
		return args, nil
	}

	var rcvIndx map[string]utils.StringSet
	if rcvIndx, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, tntID,
		utils.EmptyString, true, true); err != nil && err != utils.ErrNotFound { // error when geting the revers
		return
	}
	if err == utils.ErrNotFound || len(rcvIndx) == 0 { // no reverse index for this filter
		return args, nil
	}

	for k := range rcvIndx {
		for _, indx := range indxIDs {
			args[k] = append(args[k], utils.ConcatenatedKey(tnt, indx))
		}
	}
	return args, nil
}

// callCacheForFilter will call the cache for filter
func callCacheForFilter(connMgr *engine.ConnManager, cacheConns []string, ctx *context.Context, cacheopt, dftCache, tnt string,
	argC map[string][]string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{} = utils.NewAttrReloadCacheWithOptsFromMap(argC, tnt, opts)
	switch utils.FirstNonEmpty(cacheopt, dftCache) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
	case utils.MetaRemove:
		method = utils.CacheSv1RemoveItems
	case utils.MetaClear:
		cacheIDs := make([]string, 0, len(argC))
		for k := range argC {
			cacheIDs = append(cacheIDs, k)
		}
		method = utils.CacheSv1Clear
		args = &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   tnt,
			CacheIDs: cacheIDs,
			APIOpts:  opts,
		}
	}
	return connMgr.Call(ctx, cacheConns, method, args, &reply)
}
