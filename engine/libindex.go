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

//createAndIndex create indexes for an item
func newFilterIndex(dm DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (indexes map[string]utils.StringSet, err error) {
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	indexes = make(map[string]utils.StringSet)
	if len(filterIDs) == 0 { // in case of None
		idxKey := utils.ConcatenatedKey(utils.META_NONE, utils.META_ANY, utils.META_ANY)
		var rcvIndx map[string]utils.StringSet
		if rcvIndx, err = dm.GetIndexes(idxItmType, tntCtx,
			idxKey,
			true, true); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			indexes[idxKey] = make(utils.StringSet) // create an empty index if is not found in DB in case we add them later
			err = nil
			return
		}
		for idxKey, idx := range rcvIndx { // parse the received indexes
			indexes[idxKey] = idx
		}
		return
	}
	for _, fltrID := range filterIDs {
		var fltr *Filter
		if fltr, err = dm.GetFilter(tnt, fltrID,
			true, false, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = fmt.Errorf("broken reference to filter: %+v for itemType: %+v and ID: %+v",
					fltrID, idxItmType, itemID)
			}
			return
		}
		for _, flt := range fltr.Rules {
			if !utils.SliceHasMember([]string{utils.MetaPrefix, utils.MetaString}, flt.Type) {
				continue
			}

			for _, fldVal := range flt.Values {
				idxKey := utils.ConcatenatedKey(flt.Type, flt.Element, fldVal)
				var rcvIndx map[string]utils.StringSet
				if rcvIndx, err = dm.GetIndexes(idxItmType, tntCtx,
					idxKey, true, true); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					indexes[idxKey] = make(utils.StringSet) // create an empty index if is not found in DB in case we add them later
					err = nil
					continue
				}
				for idxKey, idx := range rcvIndx { // parse the received indexes
					indexes[idxKey] = idx
				}
			}
		}
	}
	return
}

func addItemToFilterIndex(dm DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(dm, idxItmType, tnt, ctx, itemID, filterIDs); err != nil {
		return
	}

	for _, index := range indexes {
		index.Add(itemID)
	}
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	return dm.SetIndexes(idxItmType, tntCtx, indexes, true, utils.NonTransactional)
}

func removeItemFromFilterIndex(dm DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(dm, idxItmType, tnt, ctx, itemID, filterIDs); err != nil {
		return
	}

	for idxKey, index := range indexes {
		index.Remove(itemID)
		if index.Size() == 0 { // empty index set it with nil for cache
			indexes[idxKey] = nil // this will not be set in DB(handled by driver)
		}
	}
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	return dm.SetIndexes(idxItmType, tntCtx, indexes, true, utils.NonTransactional)
}

func updatedIndexes(dm DataManager, idxItmType, tnt, ctx, itemID string, oldFilterIds *[]string, newFilterIDs []string) (err error) {
	if oldFilterIds == nil { // nothing to remove so just create the new indexes
		return addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, newFilterIDs)
	}
	if len(*oldFilterIds) == 0 && len(newFilterIDs) == 0 { // nothing to update
		return
	}

	// check what indexes needs to be updated
	oldFltrs := utils.NewStringSet(*oldFilterIds)
	newFltrs := utils.NewStringSet(newFilterIDs)

	oldFilterIDs := make([]string, 0, len(*oldFilterIds))
	newFilterIDs = make([]string, 0, len(newFilterIDs))

	for fltrID := range oldFltrs {
		if !newFltrs.Has(fltrID) { // append only if the index needs to be removed
			oldFilterIDs = append(oldFilterIDs, fltrID)
		}
	}

	for fltrID := range newFltrs {
		if !oldFltrs.Has(fltrID) { // append only if the index needs to be added
			oldFilterIDs = append(newFilterIDs, fltrID)
		}
	}

	if len(oldFilterIDs) != 0 || oldFltrs.Size() == 0 {
		// has some indexes to remove or
		// the old profile doesn't have filters but the new one has so remove the *none index
		if err = removeItemFromFilterIndex(dm, idxItmType, tnt, ctx, itemID, oldFilterIDs); err != nil {
			return
		}
	}

	if len(newFilterIDs) != 0 || newFltrs.Size() == 0 {
		// has some indexes to add or
		// the old profile has filters but the new one does not so add the *none index
		if err = addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
			return
		}
	}
	return
}
