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

// UpdateFilterIndexes will update the indexes for every reference of a filter that exists in a profile.
// Every profile that contains the filters from oldFltr will be updated with the new values for newFltr.
// oldFltr and newFltr has the same tenant and ID.
func UpdateFilterIndexes(dm *DataManager, oldFltr *Filter, newFltr *Filter) (err error) {
	return nil
}

// addReverseFilterIndexForFilter will add a reference for the filter in reverse filter indexes
func addReverseFilterIndexForFilter(dm *DataManager, idxItmType, ctx, tnt,
	itemID string, filterIDs []string) (err error) {
	for _, fltrID := range filterIDs {
		if strings.HasPrefix(fltrID, utils.Meta) { // we do not reverse for inline filters
			continue
		}

		tntFltrID := utils.ConcatenatedKey(tnt, fltrID)
		refID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntFltrID)
		var indexes map[string]utils.StringMap
		if indexes, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ReverseFilterIndexes], tntFltrID,
			utils.EmptyString, nil); err != nil {

			if err != utils.ErrNotFound {
				guardian.Guardian.UnguardIDs(refID)
				return
			}
			err = nil
			indexes = map[string]utils.StringMap{
				idxItmType: make(map[string]bool), // not found in database any reverse, we declare them to add in the next steps
			}
		}
		indexes[idxItmType] = map[string]bool{
			itemID: true,
		}
		// it is removed in StoreIndexes
		/* // remove the old reference from cache in case
		for idxKeyItmType := range indexes {
			Cache.Remove(utils.CacheReverseFilterIndexes, utils.ConcatenatedKey(tntCtx, idxKeyItmType),
				true, utils.NonTransactional)
		} */
		indexerKey := utils.ConcatenatedKey(tnt, fltrID)
		if ctx != utils.EmptyString {
			indexerKey = utils.ConcatenatedKey(tnt, ctx)
		}
		fltrIndexer := NewFilterIndexer(dm, utils.ReverseFilterIndexes, indexerKey)
		fltrIndexer.indexes = indexes
		if err = fltrIndexer.StoreIndexes(true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			return
		}
		guardian.Guardian.UnguardIDs(refID)
	}
	return
}

// removeReverseFilterIndexForFilter will remove a reference for the filter in reverse filter indexes
func removeReverseFilterIndexForFilter(dm *DataManager, idxItmType, ctx, tnt, itemID string, filterIDs []string) (err error) {
	for _, fltrID := range filterIDs {
		if strings.HasPrefix(fltrID, utils.Meta) { // we do not reverse for inline filters
			continue
		}
		tntCtx := utils.ConcatenatedKey(tnt, fltrID)
		refID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntCtx)
		var indexes map[string]utils.StringMap
		if indexes, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ReverseFilterIndexes], tntCtx,
			utils.EmptyString, nil); err != nil {
			if err != utils.ErrNotFound {
				guardian.Guardian.UnguardIDs(refID)
				return
			}
			err = nil
			continue // already removed
		}

		delete(indexes[idxItmType], itemID) // delete index from map

		indexerKey := tnt
		if ctx != utils.EmptyString {
			indexerKey = utils.ConcatenatedKey(tnt, ctx)
		}
		fltrIndexer := NewFilterIndexer(dm, utils.ReverseFilterIndexes, indexerKey)
		if err = fltrIndexer.StoreIndexes(true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			return
		}
		guardian.Guardian.UnguardIDs(refID)
	}
	return
}
