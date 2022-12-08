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

	"github.com/cgrates/cgrates/utils"
)

func ComputeChargerIndexes(dm *DataManager, tenant string, cppIDs *[]string,
	transactionID string) (cppIndexes *FilterIndexer, err error) {
	var chargerIDs []string
	//var cppIndexes *FilterIndexer
	if cppIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			chargerIDs = append(chargerIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		cppIndexes = NewFilterIndexer(dm, utils.ChargerProfilePrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ChargerProfilePrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			cppIndexes = NewFilterIndexer(dm, utils.ChargerProfilePrefix, tenant)
		} else {
			cppIndexes = NewFilterIndexerWithIndexes(dm, utils.ChargerProfilePrefix, tenant, oldIDx)
		}
		chargerIDs = *cppIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range chargerIDs {
		cpp, err := dm.GetChargerProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(cpp.FilterIDs))
		for i, fltrID := range cpp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: cpp.Tenant,
					ID:     cpp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, cpp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for charger: %+v",
						fltrID, cpp)
				}
				return nil, err
			}
			cppIndexes.IndexTPFilter(FilterToTPFilter(fltr), cpp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := cppIndexes.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := cppIndexes.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return cppIndexes, nil
}
