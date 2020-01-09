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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// MatchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.MatchFilterIndex, adding utils.ANY to list of fields queried
func MatchingItemIDsForEvent(ev map[string]interface{}, stringFldIDs, prefixFldIDs *[]string,
	dm *DataManager, cacheID, itemIDPrefix string, indexedSelects, nestedFields bool) (itemIDs utils.StringMap, err error) {
	itemIDs = make(utils.StringMap)

	// Guard will protect the function with automatic locking
	lockID := utils.CacheInstanceToPrefix[cacheID] + itemIDPrefix
	guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
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
		allFieldIDs := make([]string, len(ev))
		i := 0
		for fldID := range ev {
			allFieldIDs[i] = fldID
			i++
		}
		stringFieldVals := map[string]string{utils.ANY: utils.ANY}                               // cache here field string values, start with default one
		filterIndexTypes := []string{utils.MetaString, utils.MetaPrefix, utils.META_NONE}        // the META_NONE is used for all items that do not have filters
		for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs, &[]string{utils.ANY}} { // same routine for both string and prefix filter types
			if fieldIDs == nil {
				fieldIDs = &allFieldIDs
			}
			for _, fldName := range *fieldIDs {
				fieldValIf, has := ev[fldName]
				if !has && filterIndexTypes[i] != utils.META_NONE {
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
