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

package v1

import (
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// CallCache caching the item based on cacheopt
// visible in APIerSv2
func (apierSv1 *APIerSv1) CallCache(cacheopt string, tnt, cacheID, itemID, groupID string,
	filters *[]string, contexts []string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{}
	switch utils.FirstNonEmpty(cacheopt, apierSv1.Config.GeneralCfg().DefaultCaching) {
	case utils.MetaNone:
		return
	case utils.MetaReload:
		method = utils.CacheSv1ReloadCache
		var argCache map[string][]string
		if argCache, err = apierSv1.composeArgsReload(tnt, cacheID, itemID, filters, contexts); err != nil {
			return
		}
		args = utils.NewAttrReloadCacheWithOptsFromMap(argCache, tnt, opts)
	case utils.MetaLoad:
		method = utils.CacheSv1LoadCache
		var argCache map[string][]string
		if argCache, err = apierSv1.composeArgsReload(tnt, cacheID, itemID, filters, contexts); err != nil {
			return
		}
		args = utils.NewAttrReloadCacheWithOptsFromMap(argCache, tnt, opts)
	case utils.MetaRemove:
		if groupID != utils.EmptyString {
			method = utils.CacheSv1RemoveGroup
			args = &utils.ArgsGetGroupWithAPIOpts{
				Tenant:  tnt,
				APIOpts: opts,
				ArgsGetGroup: utils.ArgsGetGroup{
					CacheID: cacheID,
					GroupID: groupID,
				},
			}
			break
		}
		method = utils.CacheSv1RemoveItems
		var argCache map[string][]string
		if argCache, err = apierSv1.composeArgsReload(tnt, cacheID, itemID, filters, contexts); err != nil {
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
	return apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		method, args, &reply)
}

// composeArgsReload add the ItemID to AttrReloadCache
// for a specific CacheID
func (apierSv1 *APIerSv1) composeArgsReload(tnt, cacheID, itemID string, filterIDs *[]string, contexts []string) (argCache map[string][]string, err error) {
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
		if cacheID != utils.CacheAttributeProfiles &&
			cacheID != utils.CacheDispatcherProfiles {
			argCache[idxCacheID] = []string{utils.ConcatenatedKey(tnt, indxID)}
			return
		}
		argCache[idxCacheID] = make([]string, len(contexts))
		for i, ctx := range contexts {
			argCache[idxCacheID][i] = utils.ConcatenatedKey(tnt, ctx, indxID)
		}
		return
	}
	indxIDs := make([]string, 0, len(*filterIDs))
	for _, id := range *filterIDs {
		var fltr *engine.Filter
		if fltr, err = apierSv1.DataManager.GetFilter(tnt, id, true, true, utils.NonTransactional); err != nil {
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
	if cacheID != utils.CacheAttributeProfiles &&
		cacheID != utils.CacheDispatcherProfiles {
		argCache[idxCacheID] = make([]string, len(indxIDs))
		for i, indxID := range indxIDs {
			argCache[idxCacheID][i] = utils.ConcatenatedKey(tnt, indxID)
		}
		return
	}

	argCache[idxCacheID] = make([]string, 0, len(indxIDs)*len(indxIDs))
	for _, ctx := range contexts {
		for _, indxID := range indxIDs {
			argCache[idxCacheID] = append(argCache[idxCacheID], utils.ConcatenatedKey(tnt, ctx, indxID))
		}
	}
	return
}

// callCacheForIndexes will only call CacheClear because don't have access at ItemID
func (apierSv1 *APIerSv1) callCacheForRemoveIndexes(cacheopt string, tnt, cacheID string,
	itemIDs []string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{} = utils.NewAttrReloadCacheWithOptsFromMap(map[string][]string{cacheID: itemIDs}, tnt, opts)
	switch utils.FirstNonEmpty(cacheopt, apierSv1.Config.GeneralCfg().DefaultCaching) {
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
	return apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		method, args, &reply)
}

func (apierSv1 *APIerSv1) callCacheForComputeIndexes(cacheopt, tnt string,
	cacheItems map[string][]string, opts map[string]interface{}) (err error) {
	var reply, method string
	var args interface{} = utils.NewAttrReloadCacheWithOptsFromMap(cacheItems, tnt, opts)
	switch utils.FirstNonEmpty(cacheopt, apierSv1.Config.GeneralCfg().DefaultCaching) {
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
	return apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		method, args, &reply)
}

// callCacheRevDestinations used for reverse destination, loadIDs and indexes replication
func (apierSv1 *APIerSv1) callCacheMultiple(cacheopt, tnt, cacheID string, itemIDs []string, opts map[string]interface{}) (err error) {
	if len(itemIDs) == 0 {
		return
	}
	var reply, method string
	var args interface{}
	switch utils.FirstNonEmpty(cacheopt, apierSv1.Config.GeneralCfg().DefaultCaching) {
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
	return apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		method, args, &reply)
}

func composeCacheArgsForFilter(dm *engine.DataManager, fltr *engine.Filter, tnt, tntID string, args map[string][]string) (_ map[string][]string, err error) {
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
	if rcvIndx, err = dm.GetIndexes(utils.CacheReverseFilterIndexes, tntID,
		utils.EmptyString, true, true); err != nil && err != utils.ErrNotFound { // error when geting the revers
		return
	}
	if err == utils.ErrNotFound || len(rcvIndx) == 0 { // no reverse index for this filter
		return args, nil
	}

	for k, ids := range rcvIndx {
		switch k {
		default:
			for _, indx := range indxIDs {
				args[k] = append(args[k], utils.ConcatenatedKey(tnt, indx))
			}
		case utils.CacheAttributeFilterIndexes: // this is slow
			for attrID := range ids {
				var attr *engine.AttributeProfile
				if attr, err = dm.GetAttributeProfile(tnt, attrID, true, true, utils.NonTransactional); err != nil {
					return
				}
				for _, ctx := range attr.Contexts {
					for _, indx := range indxIDs {
						args[utils.CacheAttributeFilterIndexes] = append(args[utils.CacheAttributeFilterIndexes], utils.ConcatenatedKey(tnt, ctx, indx))
					}
				}
			}
		case utils.CacheDispatcherFilterIndexes: // this is slow
			for attrID := range ids {
				var attr *engine.DispatcherProfile
				if attr, err = dm.GetDispatcherProfile(tnt, attrID, true, true, utils.NonTransactional); err != nil {
					return
				}
				for _, ctx := range attr.Subsystems {
					for _, indx := range indxIDs {
						args[utils.CacheDispatcherFilterIndexes] = append(args[utils.CacheDispatcherFilterIndexes], utils.ConcatenatedKey(tnt, ctx, indx))
					}
				}
			}
		}
	}
	return args, nil
}

// callCacheForFilter will call the cache for filter
func callCacheForFilter(connMgr *engine.ConnManager, cacheConns []string, cacheopt, dftCache, tnt string,
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
	return connMgr.Call(cacheConns, nil, method, args, &reply)
}
