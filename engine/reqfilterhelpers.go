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

	"github.com/cgrates/cgrates/utils"
)

// matchingItemIDsForEvent returns the list of item IDs matching fieldName/fieldValue for an event
// fieldIDs limits the fields which are checked against indexes
// helper on top of dataDB.MatchReqFilterIndex, adding utils.ANY to list of fields queried
func matchingItemIDsForEvent(ev map[string]interface{}, stringFldIDs, prefixFldIDs *[]string,
	dm *DataManager, dbIdxKey string) (itemIDs utils.StringMap, err error) {
	itemIDs = make(utils.StringMap)
	allFieldIDs := make([]string, len(ev))
	i := 0
	for fldID := range ev {
		allFieldIDs[i] = fldID
		i += 1
	}
	filterIndexTypes := []string{MetaString, MetaPrefix}
	for i, fieldIDs := range []*[]string{stringFldIDs, prefixFldIDs} { // same routine for both string and prefix filter types
		if fieldIDs == nil {
			fieldIDs = &allFieldIDs
		}
		for _, fldName := range *fieldIDs {
			fieldValIf, has := ev[fldName]
			if !has {
				continue
			}
			fldVal, canCast := utils.CastFieldIfToString(fieldValIf)
			if !canCast {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> cannot cast field: %s into string", utils.FilterS, fldName))
				continue
			}
			fldVals := []string{fldVal} // default is only one fieldValue checked
			if filterIndexTypes[i] == MetaPrefix {
				fldVals = utils.SplitPrefix(fldVal, 1) // all prefixes till last digit
			}
			var dbItemIDs utils.StringMap // list of items matched in DB
			for _, val := range fldVals {
				dbItemIDs, err = dm.MatchFilterIndex(dbIdxKey, filterIndexTypes[i], fldName, val)
				if err != nil {
					if err == utils.ErrNotFound {
						continue
					}
					return nil, err
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
	dbItemIDs, err := dm.MatchFilterIndex(dbIdxKey, utils.MetaDefault, utils.ANY, utils.ANY) // add unindexed itemIDs to be checked
	if err != nil {
		if err != utils.ErrNotFound {
			return nil, err
		}
		err = nil // not found is ignored
	}
	for itemID := range dbItemIDs {
		if _, hasIt := itemIDs[itemID]; !hasIt {
			itemIDs[itemID] = dbItemIDs[itemID]
		}
	}
	return
}
