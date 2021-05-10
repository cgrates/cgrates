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
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

var (
	FilterIndexTypes = utils.NewStringSet([]string{utils.MetaPrefix, utils.MetaString, utils.MetaSuffix})
)

// newFilterIndex will get the index from DataManager if is not found it will create it
// is used to update the mentioned index
func newFilterIndex(apiCtx *context.Context, dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string, newFlt *Filter) (indexes map[string]utils.StringSet, err error) {
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	indexes = make(map[string]utils.StringSet)
	if len(filterIDs) == 0 { // in case of None
		idxKey := utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny)
		var rcvIndx map[string]utils.StringSet
		if rcvIndx, err = dm.GetIndexes(apiCtx, idxItmType, tntCtx,
			idxKey,
			true, false); err != nil {
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
	// in case of more filters we parse each filter rule and only for supported index types
	// we try to get them from Cache/DataDB or if not found in this location we create them here
	for _, fltrID := range filterIDs {
		var fltr *Filter
		if newFlt != nil && newFlt.Tenant == tnt && newFlt.ID == fltrID {
			fltr = newFlt
		} else if fltr, err = dm.GetFilter(apiCtx, tnt, fltrID,
			true, false, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = fmt.Errorf("broken reference to filter: %+v for itemType: %+v and ID: %+v",
					fltrID, idxItmType, itemID)
			}
			return
		}
		for _, flt := range fltr.Rules {
			if !FilterIndexTypes.Has(flt.Type) {
				continue
			}

			isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
			for _, fldVal := range flt.Values {
				var idxKey string
				if isDyn {
					if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) { // do not index if both the element and the value is dynamic
						continue
					}
					idxKey = utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal)
				} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
					idxKey = utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element)
				} else {
					// do not index not dynamic filters
					continue
				}
				var rcvIndx map[string]utils.StringSet
				// only read from cache in case if we do not find the index to not cache the negative response
				if rcvIndx, err = dm.GetIndexes(apiCtx, idxItmType, tntCtx,
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
func addItemToFilterIndex(apiCtx *context.Context, dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	// early lock to be sure that until we do not write back the indexes
	// another goroutine can't create new indexes
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
	defer guardian.Guardian.UnguardIDs(refID)

	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, filterIDs, nil); err != nil {
		return
	}
	// in case we have a profile with only non indexable filters(e.g. only *gt)
	if len(indexes) == 0 {
		return
	}

	for indxKey, index := range indexes {
		index.Add(itemID)
		// remove from cache in order to corectly update the index
		if err = Cache.Remove(apiCtx, idxItmType, utils.ConcatenatedKey(tntCtx, indxKey), true, utils.NonTransactional); err != nil {
			return
		}
	}
	return dm.SetIndexes(apiCtx, idxItmType, tntCtx, indexes, true, utils.NonTransactional)
}

// removeItemFromFilterIndex will remove the itemID from the existing/created index and set it in the DataDB
func removeItemFromFilterIndex(apiCtx *context.Context, dm *DataManager, idxItmType, tnt, ctx, itemID string, filterIDs []string) (err error) {
	tntCtx := tnt
	if ctx != utils.EmptyString {
		tntCtx = utils.ConcatenatedKey(tnt, ctx)
	}
	// early lock to be sure that until we do not write back the indexes
	// another goroutine can't create new indexes
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
	defer guardian.Guardian.UnguardIDs(refID)

	var indexes map[string]utils.StringSet
	if indexes, err = newFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, filterIDs, nil); err != nil {
		return
	}
	if len(indexes) == 0 { // in case we have a profile with only non indexable filters(e.g. only *gt)
		return
	}
	for idxKey, index := range indexes {
		index.Remove(itemID)
		if index.Size() == 0 { // empty index set it with nil for cache
			indexes[idxKey] = nil // this will not be set in DB(handled by driver)
		}
		// remove from cache in order to corectly update the index
		if err = Cache.Remove(apiCtx, idxItmType, utils.ConcatenatedKey(tntCtx, idxKey), true, utils.NonTransactional); err != nil {
			return
		}
	}
	return dm.SetIndexes(apiCtx, idxItmType, tntCtx, indexes, true, utils.NonTransactional)
}

// updatedIndexes will compare the old filtersIDs with the new ones and only update the filters indexes that are added/removed
// idxItmType - the index object type(e.g.*attribute_filter_indexes, *rate_filter_indexes, *threshold_filter_indexes)
// tnt - the tenant of the object
// ctx - the rate profile id for rate from RateProfile(sub indexes); for all the rest the ctx is ""(AttributePrf and DispatcherPrf have a separate function)
// itemID - the object id
// oldFilterIds - the filtersIDs that the old object had; this is optional if the object did not exist
// newFilterIDs - the filtersIDs for the object that will be set
// useCtx - in case of subindexes(e.g. Rate from RateProfiles) need to add the ctx to the itemID when reverse filter indexes are set
// 			used when updating the filters
func updatedIndexes(apiCtx *context.Context, dm *DataManager, idxItmType, tnt, ctx, itemID string, oldFilterIds *[]string, newFilterIDs []string, useCtx bool) (err error) {
	itmCtx := itemID
	if useCtx {
		itmCtx = utils.ConcatenatedKey(itemID, ctx)
	}
	if oldFilterIds == nil { // nothing to remove so just create the new indexes
		if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itmCtx, newFilterIDs); err != nil {
			return
		}
		return addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, newFilterIDs)
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
		if err = removeIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itmCtx, oldFilterIDs); err != nil {
			return
		}
		if err = removeItemFromFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, oldFilterIDs); err != nil {
			return
		}
	}

	if len(newFilterIDs) != 0 || newFltrs.Size() == 0 {
		// has some indexes to add or
		// the old profile has filters but the new one does not so add the *none index
		if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itmCtx, newFilterIDs); err != nil {
			return
		}
		if err = addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
			return
		}
	}
	return
}

// updatedIndexesWithContexts will compare the old contexts with the new ones and only update what is needed
// this is used by the profiles that have context(e.g. AttributeProfile)
// idxItmType - the index object type(e.g.*attribute_filter_indexes, *rate_filter_indexes, *threshold_filter_indexes)
// tnt - the tenant of the object
// itemID - the object id
// oldContexts -  the old contexts/subsystems for profile; this is optional if the object did not exist
// oldFilterIds - the filtersIDs that the old object had; this is optional if the object did not exist
// newContexts -  the new contexts/subsystems for profile that will be set
// newFilterIDs - the filtersIDs for the object that will be set
func updatedIndexesWithContexts(apiCtx *context.Context, dm *DataManager, idxItmType, tnt, itemID string,
	oldContexts, oldFilterIDs *[]string, newContexts, newFilterIDs []string) (err error) {
	if oldContexts == nil { // new profile add all indexes
		if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, newFilterIDs); err != nil {
			return
		}
		for _, ctx := range newContexts {
			if err = addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
				return
			}
		}
		return
	}

	oldCtx := utils.NewStringSet(*oldContexts)
	newCtx := utils.NewStringSet(newContexts)

	// split the contexts in three categories
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

	// remove all indexes for the old contexs
	if oldFilterIDs != nil {
		if len(updateContexts) == 0 {
			if err = removeIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, *oldFilterIDs); err != nil {
				return
			}
		}
		for _, ctx := range removeContexts {
			if err = removeItemFromFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, *oldFilterIDs); err != nil {
				return
			}
		}
	}
	// update the indexes for the contexts tha were not removed
	// in a similar way we do for the profile that do not have contexs
	if len(updateContexts) != 0 {
		if oldFilterIDs == nil { // nothing to remove so just create the new indexes
			if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, newFilterIDs); err != nil {
				return
			}
			for _, ctx := range updateContexts {
				if err = addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
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
				if err = removeIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, removeFilterIDs); err != nil {
					return
				}
				for _, ctx := range updateContexts {
					if err = removeItemFromFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, removeFilterIDs); err != nil {
						return
					}
				}
			}

			if len(addFilterIDs) != 0 || newFltrs.Size() == 0 {
				// has some indexes to add or
				// the old profile has filters but the new one does not so add the *none index
				if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, addFilterIDs); err != nil {
					return
				}
				for _, ctx := range updateContexts {
					if err = addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, addFilterIDs); err != nil {
						return
					}
				}
			}
		}
	} else if err = addIndexFiltersItem(apiCtx, dm, idxItmType, tnt, itemID, newFilterIDs); err != nil {
		return
	}

	// add indexes for new contexts
	for _, ctx := range addContexts {
		if err = addItemToFilterIndex(apiCtx, dm, idxItmType, tnt, ctx, itemID, newFilterIDs); err != nil {
			return
		}
	}
	return
}

// splitFilterIndex splits the cache key so it can be used to recache the indexes
func splitFilterIndex(tntCtxIdxKey string) (tntCtx, idxKey string, err error) {
	splt := utils.SplitConcatenatedKey(tntCtxIdxKey) // tntCtx:filterType:fieldName:fieldVal
	lsplt := len(splt)
	if lsplt < 4 {
		err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", tntCtxIdxKey)
		return
	}
	tntCtx = utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
	idxKey = utils.ConcatenatedKey(splt[lsplt-3:]...)
	return
}

// ComputeIndexes gets the indexes from the DB and ensure that the items are indexed
// getFilters returns a list of filters IDs for the given profile id
func ComputeIndexes(cntxt *context.Context, dm *DataManager, tnt, ctx, idxItmType string, IDs *[]string,
	transactionID string, getFilters func(tnt, id, ctx string) (*[]string, error), newFltr *Filter) (indexes utils.StringSet, err error) {
	indexes = make(utils.StringSet)
	var profilesIDs []string
	if IDs == nil { // get all items
		var ids []string
		if ids, err = dm.DataDB().GetKeysForPrefix(cntxt, utils.CacheIndexesToPrefix[idxItmType]); err != nil {
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
	// early lock to be sure that until we do not write back the indexes
	// another goroutine can't create new indexes
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
	defer guardian.Guardian.UnguardIDs(refID)
	for _, id := range profilesIDs {
		var filterIDs *[]string
		if filterIDs, err = getFilters(tnt, id, ctx); err != nil {
			return
		}
		if filterIDs == nil {
			continue
		}
		var index map[string]utils.StringSet
		if index, err = newFilterIndex(cntxt, dm, idxItmType,
			tnt, ctx, id, *filterIDs, newFltr); err != nil {
			return
		}
		// ensure that the item is in the index set
		for key, idx := range index {
			idx.Add(id)
			indexes.Add(utils.ConcatenatedKey(tntCtx, key))
		}
		if err = dm.SetIndexes(cntxt, idxItmType, tntCtx, index, cacheCommit(transactionID), transactionID); err != nil {
			return
		}
	}
	return
}

// addIndexFiltersItem will add a reference for the items in the reverse filter index
func addIndexFiltersItem(ctx *context.Context, dm *DataManager, idxItmType, tnt, itemID string, filterIDs []string) (err error) {
	for _, ID := range filterIDs {
		if strings.HasPrefix(ID, utils.Meta) { // skip inline
			continue
		}
		tntCtx := utils.ConcatenatedKey(tnt, ID)
		refID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntCtx)
		var indexes map[string]utils.StringSet
		if indexes, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, tntCtx,
			idxItmType, true, false); err != nil {
			if err != utils.ErrNotFound {
				guardian.Guardian.UnguardIDs(refID)
				return
			}
			err = nil
			indexes = map[string]utils.StringSet{
				idxItmType: make(utils.StringSet), // create an empty index if is not found in DB in case we add them later
			}
		}
		indexes[idxItmType].Add(itemID)
		for indxKey := range indexes {
			if err = Cache.Remove(ctx, utils.CacheReverseFilterIndexes, utils.ConcatenatedKey(tntCtx, indxKey), true, utils.NonTransactional); err != nil {
				guardian.Guardian.UnguardIDs(refID)
				return
			}
		}
		if err = dm.SetIndexes(ctx, utils.CacheReverseFilterIndexes, tntCtx, indexes, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			return
		}
		guardian.Guardian.UnguardIDs(refID)
	}
	return
}

// removeIndexFiltersItem will removes a reference for the items in the reverse filter index
func removeIndexFiltersItem(ctx *context.Context, dm *DataManager, idxItmType, tnt, itemID string, filterIDs []string) (err error) {
	for _, ID := range filterIDs {
		if strings.HasPrefix(ID, utils.Meta) { // skip inline
			continue
		}
		tntCtx := utils.ConcatenatedKey(tnt, ID)
		refID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntCtx)
		var indexes map[string]utils.StringSet
		if indexes, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, tntCtx,
			idxItmType, true, false); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			if err != utils.ErrNotFound {
				return
			}
			err = nil
			continue // it is already removed
		}
		indexes[idxItmType].Remove(itemID)

		for indxKey := range indexes {
			if err = Cache.Remove(ctx, utils.CacheReverseFilterIndexes, utils.ConcatenatedKey(tntCtx, indxKey), true, utils.NonTransactional); err != nil {
				guardian.Guardian.UnguardIDs(refID)
				return
			}
		}
		if err = dm.SetIndexes(ctx, utils.CacheReverseFilterIndexes, tntCtx, indexes, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			return
		}
		guardian.Guardian.UnguardIDs(refID)
	}
	return
}

// UpdateFilterIndex  will update the indexes for the new Filter
// we do not care what is added
// exported for the migrator
func UpdateFilterIndex(ctx *context.Context, dm *DataManager, oldFlt, newFlt *Filter) (err error) {
	if oldFlt == nil { // no filter before so no index to update
		return // nothing to update
	}

	// split the rules so we can determine if we need to update the indexes
	oldRules := utils.StringSet{}
	newRules := utils.StringSet{}    // we only need to determine if we added new rules to rebuild
	removeRules := utils.StringSet{} // but we need to know what indexes to remove
	for _, flt := range newFlt.Rules {
		if !FilterIndexTypes.Has(flt.Type) {
			continue
		}
		isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
		for _, fldVal := range flt.Values {
			var idxKey string
			if isDyn {
				if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) { // do not index if both the element and the value is dynamic
					continue
				}
				idxKey = utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal)
			} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
				idxKey = utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element)
			} else {
				// do not index not dynamic filters
				continue
			}
			newRules.Add(idxKey)
		}
	}
	for _, flt := range oldFlt.Rules {
		if !FilterIndexTypes.Has(flt.Type) {
			continue
		}
		isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
		for _, fldVal := range flt.Values {
			var idxKey string
			if isDyn {
				if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) { // do not index if both the element and the value is dynamic
					continue
				}
				idxKey = utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal)
			} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
				idxKey = utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element)
			} else {
				// do not index not dynamic filters
				continue
			}
			if !newRules.Has(idxKey) {
				removeRules.Add(idxKey)
			} else {
				oldRules.Add(idxKey)
			}
		}
	}
	needsRebuild := removeRules.Size() != 0 // nothing to remove means nothing to rebuild
	if !needsRebuild {                      // so check if we added somrthing
		for key := range newRules {
			if needsRebuild = !oldRules.Has(key); needsRebuild {
				break
			}
		}
		if !needsRebuild { // if we did not remove or add we do not need to rebuild the indexes
			return
		}
	}

	tntID := newFlt.TenantID()
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntID)
	defer guardian.Guardian.UnguardIDs(refID)
	var rcvIndx map[string]utils.StringSet
	// get all reverse indexes from DB
	if rcvIndx, err = dm.GetIndexes(ctx, utils.CacheReverseFilterIndexes, tntID,
		utils.EmptyString, true, false); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		err = nil // no index for this filter so  no update needed
		return
	}
	removeIndexKeys := removeRules.AsSlice()

	// remove the old indexes and compute the new ones
	for idxItmType, indx := range rcvIndx {
		switch idxItmType {
		case utils.CacheThresholdFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, // remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					th, e := dm.GetThresholdProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(th.FilterIDs))
					for i, fltrID := range th.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheStatFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, // remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					sq, e := dm.GetStatQueueProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(sq.FilterIDs))
					for i, fltrID := range sq.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheResourceFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, // remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					rs, e := dm.GetResourceProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(rs.FilterIDs))
					for i, fltrID := range rs.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheRouteFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, // remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					rt, e := dm.GetRouteProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(rt.FilterIDs))
					for i, fltrID := range rt.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheChargerFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, // remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					ch, e := dm.GetChargerProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(ch.FilterIDs))
					for i, fltrID := range ch.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheAccountsFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, //remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					ap, e := dm.GetAccount(tnt, id)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(ap.FilterIDs))
					for i, fltrID := range ap.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheActionProfilesFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, //remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					acp, e := dm.GetActionProfile(tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(acp.FilterIDs))
					for i, fltrID := range acp.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		case utils.CacheRateProfilesFilterIndexes:
			if err = removeFilterIndexesForFilter(dm, idxItmType, newFlt.Tenant, //remove the indexes for the filter
				removeIndexKeys, indx); err != nil {
				return
			}
			idxSlice := indx.AsSlice()
			if _, err = ComputeIndexes(ctx, dm, newFlt.Tenant, utils.EmptyString, idxItmType, // compute all the indexes for afected items
				&idxSlice, utils.NonTransactional, func(tnt, id, ctx string) (*[]string, error) {
					rp, e := dm.GetRateProfile(context.TODO(), tnt, id, true, false, utils.NonTransactional)
					if e != nil {
						return nil, e
					}
					fltrIDs := make([]string, len(rp.FilterIDs))
					for i, fltrID := range rp.FilterIDs {
						fltrIDs[i] = fltrID
					}
					return &fltrIDs, nil
				}, newFlt); err != nil && err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}

		case utils.CacheRateFilterIndexes:
			itemIDs := make(map[string]utils.StringSet)
			for itemID := range indx {
				idSplit := strings.SplitN(itemID, utils.ConcatenatedKeySep, 2)
				if len(idSplit) < 2 {
					return errors.New("Expected to be 2 values")
				}
				if itemIDs[idSplit[1]] == nil {
					itemIDs[idSplit[1]] = make(utils.StringSet)
				}
				itemIDs[idSplit[1]].Add(idSplit[0])
			}
			for rpID, ids := range itemIDs {
				tntCtx := utils.ConcatenatedKey(newFlt.Tenant, rpID)
				if err = removeFilterIndexesForFilter(dm, idxItmType, tntCtx,
					removeIndexKeys, ids); err != nil {
					return
				}
				var rp *utils.RateProfile
				if rp, err = dm.GetRateProfile(context.TODO(), newFlt.Tenant, rpID, true, false, utils.NonTransactional); err != nil {
					return
				}
				for itemID := range ids {
					rate, has := rp.Rates[itemID]
					if !has {
						return utils.ErrNotFound
					}
					refID := guardian.Guardian.GuardIDs(utils.EmptyString,
						config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
					var updIdx map[string]utils.StringSet
					if updIdx, err = newFilterIndex(context.TODO(), dm, idxItmType,
						newFlt.Tenant, rpID, itemID, rate.FilterIDs, newFlt); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					for _, idx := range updIdx {
						idx.Add(itemID)
					}
					if err = dm.SetIndexes(context.TODO(), idxItmType, tntCtx,
						updIdx, false, utils.NonTransactional); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					guardian.Guardian.UnguardIDs(refID)
				}
			}
		case utils.CacheAttributeFilterIndexes:
			for itemID := range indx {
				var ap *AttributeProfile
				if ap, err = dm.GetAttributeProfile(context.TODO(), newFlt.Tenant, itemID,
					true, false, utils.NonTransactional); err != nil {
					return
				}
				for _, ctx := range ap.Contexts {
					tntCtx := utils.ConcatenatedKey(newFlt.Tenant, ctx)
					if err = removeFilterIndexesForFilter(dm, idxItmType,
						tntCtx, // remove the indexes for the filter
						removeIndexKeys, indx); err != nil {
						return
					}
					refID := guardian.Guardian.GuardIDs(utils.EmptyString,
						config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
					var updIdx map[string]utils.StringSet
					if updIdx, err = newFilterIndex(context.TODO(), dm, idxItmType,
						newFlt.Tenant, ctx, itemID, ap.FilterIDs, newFlt); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					for _, idx := range updIdx {
						idx.Add(itemID)
					}
					if err = dm.SetIndexes(context.TODO(), idxItmType, tntCtx,
						updIdx, false, utils.NonTransactional); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					guardian.Guardian.UnguardIDs(refID)
				}
			}
		case utils.CacheDispatcherFilterIndexes:
			for itemID := range indx {
				var dp *DispatcherProfile
				if dp, err = dm.GetDispatcherProfile(newFlt.Tenant, itemID,
					true, false, utils.NonTransactional); err != nil {
					return
				}
				for _, ctx := range dp.Subsystems {
					tntCtx := utils.ConcatenatedKey(newFlt.Tenant, ctx)
					if err = removeFilterIndexesForFilter(dm, idxItmType,
						tntCtx, // remove the indexes for the filter
						removeIndexKeys, indx); err != nil {
						return
					}
					refID := guardian.Guardian.GuardIDs(utils.EmptyString,
						config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tntCtx)
					var updIdx map[string]utils.StringSet
					if updIdx, err = newFilterIndex(context.TODO(), dm, idxItmType,
						newFlt.Tenant, ctx, itemID, dp.FilterIDs, newFlt); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					for _, idx := range updIdx {
						idx.Add(itemID)
					}
					if err = dm.SetIndexes(context.TODO(), idxItmType, tntCtx,
						updIdx, false, utils.NonTransactional); err != nil {
						guardian.Guardian.UnguardIDs(refID)
						return
					}
					guardian.Guardian.UnguardIDs(refID)
				}
			}
		}
	}
	return
}

// removeFilterIndexesForFilter removes the itemID for the index keys
// used to remove the old indexes when a filter is updated
func removeFilterIndexesForFilter(dm *DataManager, idxItmType, tnt string,
	removeIndexKeys []string, itemIDs utils.StringSet) (err error) {
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tnt)
	defer guardian.Guardian.UnguardIDs(refID)
	for _, idxKey := range removeIndexKeys { // delete old filters indexes for this item
		var remIndx map[string]utils.StringSet
		if remIndx, err = dm.GetIndexes(context.TODO(), idxItmType, tnt,
			idxKey, true, false); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			err = nil
			continue
		}
		for idx := range itemIDs {
			remIndx[idxKey].Remove(idx)
		}

		if err = dm.SetIndexes(context.TODO(), idxItmType, tnt, remIndx, true, utils.NonTransactional); err != nil {
			return
		}
	}
	return
}
