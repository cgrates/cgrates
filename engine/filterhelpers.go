// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// MatchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.MatchFilterIndex, adding utils.ANY to list of fields queried
func MatchingItemIDsForEvent(ev map[string]any, stringFldIDs, prefixFldIDs *[]string,
	dm *DataManager, cacheID, itemIDPrefix string, indexedSelects, nestedFields bool) (itemIDs utils.StringMap, err error) {
	itemIDs = make(utils.StringMap)
	var allFieldIDs []string
	navEv := utils.MapStorage(ev)
	if indexedSelects && (stringFldIDs == nil || prefixFldIDs == nil) {
		allFieldIDs = navEv.GetKeys(nestedFields)
	}
	// Guard will protect the function with automatic locking
	lockID := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	guardian.Guardian.Guard(func() (gRes any, gErr error) {
		if !indexedSelects {
			var keysWithID []string
			if keysWithID, err = dm.DataDB().GetKeysForPrefix(utils.CacheIndexesToPrefix[cacheID]); err != nil {
				return
			}
			var sliceIDs []string
			for _, id := range keysWithID {
				sliceIDs = append(sliceIDs, strings.Split(id, ":")[1])
			}
			itemIDs = utils.StringMapFromSlice(sliceIDs)
			return
		}
		stringFieldVals := map[string]string{utils.ANY: utils.ANY}                        // cache here field string values, start with default one
		filterIndexTypes := []string{utils.MetaString, utils.MetaPrefix, utils.META_NONE} // the META_NONE is used for all items that do not have filters
		for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs, {utils.ANY}} {   // same routine for both string and prefix filter types
			if fieldIDs == nil {
				fieldIDs = &allFieldIDs
			}
			for _, fldName := range *fieldIDs {
				fieldValIf, err := navEv.FieldAsInterface(strings.Split(fldName, utils.NestingSep))
				if err != nil && filterIndexTypes[i] != utils.META_NONE {
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
				}
				if fldName != utils.META_ANY {
					fldName = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + fldName
				}
				var dbItemIDs utils.StringMap // list of items matched in DB
				for _, val := range fldVals {
					if dbItemIDs, err = dm.MatchFilterIndex(cacheID, itemIDPrefix, filterIndexTypes[i], fldName, val); err != nil {
						if err == utils.ErrNotFound {
							err = nil
							continue
						}
						return
					}
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
	}, config.CgrConfig().GeneralCfg().LockingTimeout, lockID)
	if len(itemIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}
