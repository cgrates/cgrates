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

// composeArgsReload add the ItemID to AttrReloadCache
// for a specific CacheID
func (apierSv1 *APIerSv1) composeArgsReload(tnt, cacheID, itemID string, filterIDs *[]string, contexts []string, opts map[string]interface{}) (rpl utils.AttrReloadCacheWithOpts, err error) {
	rpl = utils.AttrReloadCacheWithOpts{
		TenantArg: utils.TenantArg{Tenant: tnt},
		ArgsCache: map[string][]string{
			utils.CacheInstanceToArg[cacheID]: {itemID},
		},
		Opts: opts,
	}
	if filterIDs == nil { // in case we remove a profile we do not need to reload the indexes
		return
	}
	// popultate the indexes
	idxCacheID := utils.CacheInstanceToCacheIndex[cacheID]
	if len(*filterIDs) == 0 { // in case we do not have any filters reload the *none filter indexes
		indxID := utils.ConcatenatedKey(utils.META_NONE, utils.META_ANY, utils.META_ANY)
		if cacheID != utils.CacheAttributeProfiles &&
			cacheID != utils.CacheDispatcherProfiles {
			rpl.ArgsCache[idxCacheID] = []string{utils.ConcatenatedKey(tnt, indxID)}
			return
		}
		rpl.ArgsCache[idxCacheID] = make([]string, len(contexts))
		for i, ctx := range contexts {
			rpl.ArgsCache[idxCacheID][i] = utils.ConcatenatedKey(tnt, ctx, indxID)
		}
	}
	indxIDs := make([]string, 0, len(*filterIDs))
	for _, id := range *filterIDs {
		var fltr *engine.Filter
		if fltr, err = apierSv1.DataManager.GetFilter(tnt, id, true, true, utils.NonTransactional); err != nil {
			return
		}
		for _, flt := range fltr.Rules {
			if !engine.FilterIndexTypes.Has(flt.Type) {
				continue
			}
			isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
			for _, fldVal := range flt.Values {
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
		rpl.ArgsCache[idxCacheID] = make([]string, len(indxIDs))
		for i, indxID := range indxIDs {
			rpl.ArgsCache[idxCacheID][i] = utils.ConcatenatedKey(tnt, indxID)
		}
		return
	}

	rpl.ArgsCache[idxCacheID] = make([]string, 0, len(indxIDs)*len(indxIDs))
	for _, ctx := range contexts {
		for _, indxID := range indxIDs {
			rpl.ArgsCache[idxCacheID] = append(rpl.ArgsCache[idxCacheID], utils.ConcatenatedKey(tnt, ctx, indxID))
		}
	}
	return
}
