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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// newFilterIndex will get the index from DataManager if is not found it will create it
func newFilterIndex(dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (indexes map[string]utils.StringSet, err error) {
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
			err = nil
			indexes[idxKey] = make(utils.StringSet) // create an empty index if is not found in DB in case we add them later
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
					idxKey, true, false); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					err = nil
					indexes[idxKey] = make(utils.StringSet) // create an empty index if is not found in DB in case we add them later
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

// addItemToFilterIndex will add the itemID to the existing/created index and set it in the DataDB
func addItemToFilterIndex(dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(dm, idxItmType, tnt, ctx, itemID, filterIDs); err != nil {
		return
	}
	if len(indexes) == 0 { // in case we have a profile with only non indexable filters(e.g. only *gt)
		return
	}
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	refID := guardian.Guardian.GuardIDs("",
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
	defer guardian.Guardian.UnguardIDs(refID)

	for _, index := range indexes {
		index.Add(itemID)
	}

	if err = dm.SetIndexes(idxItmType, tntCtx, indexes, true, utils.NonTransactional); err != nil {
		return
	}
	for indxKey := range indexes {
		if err = Cache.Remove(idxItmType, utils.ConcatenatedKey(tntCtx, indxKey), true, utils.NonTransactional); err != nil {
			return
		}
	}
	return
}

// addItemToFilterIndex will remove the itemID from the existing/created index and set it in the DataDB
func removeItemFromFilterIndex(dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(dm, idxItmType, tnt, ctx, itemID, filterIDs); err != nil {
		return
	}
	if len(indexes) == 0 { // in case we have a profile with only non indexable filters(e.g. only *gt)
		return
	}

	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	refID := guardian.Guardian.GuardIDs("",
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
	defer guardian.Guardian.UnguardIDs(refID)

	for idxKey, index := range indexes {
		index.Remove(itemID)
		if index.Size() == 0 { // empty index set it with nil for cache
			indexes[idxKey] = nil // this will not be set in DB(handled by driver)
		}
	}

	if err = dm.SetIndexes(idxItmType, tntCtx, indexes, true, utils.NonTransactional); err != nil {
		return
	}
	for indxKey := range indexes {
		if err = Cache.Remove(idxItmType, utils.ConcatenatedKey(tntCtx, indxKey), true, utils.NonTransactional); err != nil {
			return
		}
	}
	return
}

// updatedIndexes will compare the old filtersIDs with the new ones and only uptdate the filters indexes that are added/removed
func updatedIndexes(dm *DataManager, idxItmType, tnt, ctx, itemID string, oldFilterIds *[]string, newFilterIDs []string) (err error) {
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
			newFilterIDs = append(newFilterIDs, fltrID)
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

func updatedIndexesWithContexts(dm *DataManager, idxItmType, tnt, itemID string,
	oldContexts, oldFilterIDs *[]string, newContexts, newFilterIDs []string) (err error) {
	if oldContexts == nil { // new profile add all indexes
		for _, ctx := range newContexts {
			if err = addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
				return
			}
		}
		return
	}

	oldCtx := utils.NewStringSet(*oldContexts)
	newCtx := utils.NewStringSet(newContexts)

	removeContexts := make([]string, 0, len(*oldContexts))
	addContexts := make([]string, 0, len(newContexts))
	updateContexts := make([]string, 0, len(newContexts))

	for ctx := range oldCtx {
		if !newCtx.Has(ctx) { // append only if the index needs to be removed
			removeContexts = append(removeContexts, ctx)
		} else {
			updateContexts = append(updateContexts, ctx)
		}
	}

	for ctx := range newCtx {
		if !oldCtx.Has(ctx) { // append only if the index needs to be added
			addContexts = append(addContexts, ctx)
		}
	}

	if oldFilterIDs != nil {
		for _, ctx := range removeContexts {
			if err = removeItemFromFilterIndex(dm, idxItmType, tnt, ctx, itemID, *oldFilterIDs); err != nil {
				return
			}
		}
	}
	if len(updateContexts) != 0 {
		if oldFilterIDs == nil { // nothing to remove so just create the new indexes
			for _, ctx := range updateContexts {
				if err = addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
					return
				}
			}
		} else if len(*oldFilterIDs) != 0 || len(newFilterIDs) != 0 { // nothing to update
			// check what indexes needs to be updated
			oldFltrs := utils.NewStringSet(*oldFilterIDs)
			newFltrs := utils.NewStringSet(newFilterIDs)

			removeFilterIDs := make([]string, 0, len(*oldFilterIDs))
			addFilterIDs := make([]string, 0, len(newFilterIDs))

			for fltrID := range oldFltrs {
				if !newFltrs.Has(fltrID) { // append only if the index needs to be removed
					removeFilterIDs = append(removeFilterIDs, fltrID)
				}
			}

			for fltrID := range newFltrs {
				if !oldFltrs.Has(fltrID) { // append only if the index needs to be added
					addFilterIDs = append(addFilterIDs, fltrID)
				}
			}

			if len(removeFilterIDs) != 0 || oldFltrs.Size() == 0 {
				// has some indexes to remove or
				// the old profile doesn't have filters but the new one has so remove the *none index
				for _, ctx := range updateContexts {
					if err = removeItemFromFilterIndex(dm, idxItmType, tnt, ctx, itemID, removeFilterIDs); err != nil {
						return
					}
				}
			}

			if len(addFilterIDs) != 0 || newFltrs.Size() == 0 {
				// has some indexes to add or
				// the old profile has filters but the new one does not so add the *none index
				for _, ctx := range updateContexts {
					if err = addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, addFilterIDs); err != nil {
						return
					}
				}
			}
		}
	}

	for _, ctx := range addContexts {
		if err = addItemToFilterIndex(dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
			return
		}
	}
	return
}

func splitFilterIndex(tntCtxIdxKey string) (tntCtx, idxKey string, err error) {
	splt := utils.SplitConcatenatedKey(tntCtxIdxKey) // tntCtx:filterType:fieldName:fieldVal
	lsplt := len(splt)
	if lsplt < 4 {
		err = fmt.Errorf("WRONG_IDX_KEY_FORMAT")
		return
	}
	tntCtx = utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
	idxKey = utils.ConcatenatedKey(splt[lsplt-3:]...)
	return
}

// ComputeIndexes gets the indexes from tha DB and ensure that the items are indexed
// getFilters returns a list of filters IDs for the given profile id
func ComputeIndexes(dm *DataManager, tnt, ctx, idxItmType string, IDs *[]string,
	transactionID string, getFilters func(tnt, id, ctx string) (*[]string, error)) (err error) {
	var profilesIDs []string
	if IDs == nil { // get all items
		var ids []string
		if ids, err = dm.DataDB().GetKeysForPrefix(utils.CacheIndexesToPrefix[idxItmType]); err != nil {
			return
		}
		for _, id := range ids {
			profilesIDs = append(profilesIDs, utils.SplitConcatenatedKey(id)[1])
		}
	} else {
		profilesIDs = *IDs
	}
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	for _, id := range profilesIDs {
		var filterIDs *[]string
		if filterIDs, err = getFilters(tnt, id, ctx); err != nil {
			return
		}
		if filterIDs == nil {
			continue
		}
		var index map[string]utils.StringSet
		if index, err = newFilterIndex(dm, idxItmType,
			tnt, ctx, id, *filterIDs); err != nil {
			return
		}
		for _, idx := range index {
			idx.Add(id)
		}
		if err = dm.SetIndexes(idxItmType, tntCtx, index, cacheCommit(transactionID), transactionID); err != nil {
			return
		}
	}
	return
}
