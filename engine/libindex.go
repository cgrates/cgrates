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

/*
import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

//createAndIndex create indexes for an item
func createAndIndex2(tnt, ctx, itmType, itmID string, filterIDs []string) (indx map[string]utils.StringSet, err error) {
	indexerKey := tnt
	if ctx != "" {
		indexerKey = utils.ConcatenatedKey(tnt, ctx)
	}
	indexer := NewFilterIndexer(dm, itmType, indexerKey)
	fltrIDs := make([]string, len(filterIDs))
	for i, fltrID := range filterIDs {
		fltrIDs[i] = fltrID
	}
	if len(fltrIDs) == 0 {
		fltrIDs = []string{utils.META_NONE}
	}
	for _, fltrID := range fltrIDs {
		var fltr *Filter
		if fltrID == utils.META_NONE {
			fltr = &Filter{
				Tenant: tnt,
				ID:     itmID,
				Rules: []*FilterRule{
					{
						Type:    utils.META_NONE,
						Element: utils.META_ANY,
						Values:  []string{utils.META_ANY},
					},
				},
			}
		} else if fltr, err = dm.GetFilter(tnt, fltrID,
			true, false, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = fmt.Errorf("broken reference to filter: %+v for itemType: %+v and ID: %+v",
					fltrID, itmType, itmID)
			}
			return
		}
		for _, flt := range fltr.Rules {
			var fldType, fldName string
			var fldVals []string
			if utils.SliceHasMember([]string{utils.META_NONE, utils.MetaPrefix, utils.MetaString}, flt.Type) {
				fldType, fldName = flt.Type, flt.Element
				fldVals = flt.Values
			}
			for _, fldVal := range fldVals {
				if err = indexer.loadFldNameFldValIndex(fldType,
					fldName, fldVal); err != nil && err != utils.ErrNotFound {
					return err
				}
			}
		}
		indexer.IndexTPFilter(FilterToTPFilter(fltr), itmID)
	}
	return indexer.StoreIndexes(true, utils.NonTransactional)
}
*/
