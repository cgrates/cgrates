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
// helper on top of dataDB.MatchReqFilterIndex, adding utils.NOT_AVAILABLE to list of fields queried
// executes a number of $(len(fields) + 1) queries to dataDB so the size of event influences the speed of return
func matchingItemIDsForEvent(ev map[string]interface{}, fieldIDs []string,
	dm *DataManager, dbIdxKey, filterType string) (itemIDs utils.StringMap, err error) {
	if len(fieldIDs) == 0 {
		fieldIDs = make([]string, len(ev))
		i := 0
		for fldID := range ev {
			fieldIDs[i] = fldID
			i += 1
		}
	}
	itemIDs = make(utils.StringMap)
	for _, fldName := range fieldIDs {
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
		dbItemIDs, err := dm.MatchFilterIndex(dbIdxKey, filterType, fldName, fldVal)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		for itemID := range dbItemIDs {
			if _, hasIt := itemIDs[itemID]; !hasIt { // Add it to list if not already there
				itemIDs[itemID] = dbItemIDs[itemID]
			}
		}
	}
	dbItemIDs, err := dm.MatchFilterIndex(dbIdxKey, utils.MetaDefault, utils.NOT_AVAILABLE, utils.NOT_AVAILABLE) // add unindexed itemIDs to be checked
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
