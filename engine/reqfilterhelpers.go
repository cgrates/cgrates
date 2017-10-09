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
// helper on top of dataDB.MatchReqFilterIndex, adding utils.NOT_AVAILABLE to list of fields queried
// executes a number of $(len(fields) + 1) queries to dataDB so the size of event influences the speed of return
func matchingItemIDsForEvent(ev map[string]interface{}, dm *DataManager, dbIdxKey string) (itemIDs utils.StringMap, err error) {
	itemIDs = make(utils.StringMap)
	for fldName, fieldValIf := range ev {
		fldVal, canCast := utils.CastFieldIfToString(fieldValIf)
		if !canCast {
			return nil, fmt.Errorf("Cannot cast field: %s into string", fldName)
		}
		dbItemIDs, err := dm.DataDB().MatchReqFilterIndex(dbIdxKey, fldName, fldVal)
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
	dbItemIDs, err := dm.DataDB().MatchReqFilterIndex(dbIdxKey, utils.NOT_AVAILABLE, utils.NOT_AVAILABLE) // add unindexed itemIDs to be checked
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
