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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// MatchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.GetIndexes, adding utils.MetaAny to list of fields queried
func MatchingItemIDsForEvent(ctx *context.Context, ev utils.MapStorage, stringFldIDs, prefixFldIDs, suffixFldIDs *[]string,
	dm *DataManager, cacheID, itemIDPrefix string, indexedSelects, nestedFields bool) (itemIDs utils.StringSet, err error) {
	itemIDs = make(utils.StringSet)
	var allFieldIDs []string
	if indexedSelects && (stringFldIDs == nil || prefixFldIDs == nil || suffixFldIDs == nil) {
		allFieldIDs = ev.GetKeys(nestedFields, 2, utils.EmptyString)
	}
	// Guard will protect the function with automatic locking
	lockID := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	guardian.Guardian.Guard(ctx, func(ctx *context.Context) (_ error) {
		if !indexedSelects {
			var keysWithID []string
			if keysWithID, err = dm.DataDB().GetKeysForPrefix(ctx, utils.CacheIndexesToPrefix[cacheID]); err != nil {
				return
			}
			var sliceIDs []string
			for _, id := range keysWithID {
				sliceIDs = append(sliceIDs, utils.SplitConcatenatedKey(id)[1])
			}
			itemIDs = utils.NewStringSet(sliceIDs)
			return
		}
		stringFieldVals := map[string]string{utils.MetaAny: utils.MetaAny}                                 // cache here field string values, start with default one
		filterIndexTypes := []string{utils.MetaString, utils.MetaPrefix, utils.MetaSuffix, utils.MetaNone} // the MetaNone is used for all items that do not have filters
		for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs, suffixFldIDs, {utils.MetaAny}} {  // same routine for both string and prefix filter types
			if fieldIDs == nil {
				fieldIDs = &allFieldIDs
			}
			for _, fldName := range *fieldIDs {
				var fieldValIf interface{}
				fieldValIf, err = ev.FieldAsInterface(utils.SplitPath(fldName, utils.NestingSep[0], -1))
				if err != nil && filterIndexTypes[i] != utils.MetaNone {
					continue
				}
				if _, cached := stringFieldVals[fldName]; !cached {
					stringFieldVals[fldName] = utils.IfaceAsString(fieldValIf)
				}
				fldVal := stringFieldVals[fldName]
				fldVals := []string{fldVal}
				// default is only one fieldValue checked
				if filterIndexTypes[i] == utils.MetaPrefix {
					fldVals = utils.SplitPrefix(fldVal, 1) // all prefixes till last digit
				} else if filterIndexTypes[i] == utils.MetaSuffix {
					fldVals = utils.SplitSuffix(fldVal) // all suffix till first digit
				}
				var dbItemIDs utils.StringSet // list of items matched in DB
				for _, val := range fldVals {
					var dbIndexes map[string]utils.StringSet // list of items matched in DB
					key := utils.ConcatenatedKey(filterIndexTypes[i], fldName, val)
					if dbIndexes, err = dm.GetIndexes(ctx, cacheID, itemIDPrefix, key, utils.NonTransactional, true, true); err != nil {
						if err == utils.ErrNotFound {
							err = nil
							continue
						}
						return
					}
					dbItemIDs = dbIndexes[key]
					break // we got at least one answer back, longest prefix wins
				}
				for itemID := range dbItemIDs {
					if _, hasIt := itemIDs[itemID]; !hasIt { // Add it to list if not already there
						itemIDs[itemID] = dbItemIDs[itemID]
					}
				}
			}
		}
		return
	},
		config.CgrConfig().GeneralCfg().LockingTimeout, lockID)
	if err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

// Weight returns weight of the first matching DynamicWeight
func WeightFromDynamics(ctx *context.Context, dWs []*utils.DynamicWeight,
	fltrS *FilterS, tnt string, ev utils.DataProvider) (wg float64, err error) {
	for _, dW := range dWs {
		var pass bool
		if pass, err = fltrS.Pass(ctx, tnt, dW.FilterIDs, ev); err != nil {
			return
		} else if pass {
			return dW.Weight, nil
		}
	}
	return 0.0, nil
}
