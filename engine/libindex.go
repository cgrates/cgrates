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

var (
	filterIndexType = utils.StringMap{
		utils.MetaString: true,
		utils.MetaPrefix: true}
)

// UpdateFilterIndexes will update the indexes for every reference of a filter that exists in a profile.
// Every profile that contains the filters from oldFltr will be updated with the new values for newFltr.
// oldFltr and newFltr has the same tenant and ID.
func UpdateFilterIndexes(dm *DataManager, tnt string, oldFltr *Filter, newFltr *Filter) (err error) {
	// we need the rules in roder to compute the new indexes
	oldRules := utils.StringMap{}    // rules from old filters
	newRules := utils.StringMap{}    // rules for new filters
	removeRules := utils.StringMap{} // the difference from newRules and oldRules that are needed to be removed
	// first we check the rules from the new filter
	for _, fltr := range newFltr.Rules {
		if !filterIndexType.HasKey(fltr.Type) { // we do not consider other types, just *string and *prefix
			continue
		}
		isElementDyn := strings.HasPrefix(fltr.Element, utils.DynamicDataPrefix)
		for _, value := range fltr.Values {
			var idxKey string
			if isElementDyn {
				// we do not index element:value both of dynamic types e.g. *string:~*req.Account:~*req.Destination
				if strings.HasPrefix(value, utils.DynamicDataPrefix) {
					continue
				}
				idxKey = utils.ConcatenatedKey(fltr.Type, fltr.Element, value)
			} else if strings.HasPrefix(value, utils.DynamicDataPrefix) {
				idxKey = utils.ConcatenatedKey(fltr.Type, value, fltr.Element)
			} else {
				continue // none of the element or value are dynamic, so we do not index
			}
			newRules[idxKey] = true
		}
	}
	// now we check the rules from the old filter
	// compare the new rules and old rules and check what rules needs to be removed
	for _, fltr := range oldFltr.Rules {
		if !filterIndexType.HasKey(fltr.Type) { // we do not consider other types, just *string and *prefix
			continue
		}
		isElementDyn := strings.HasPrefix(fltr.Element, utils.DynamicDataPrefix)
		for _, value := range fltr.Values {
			var idxKey string
			if isElementDyn {
				// we do not index element:value both of dynamic types e.g. *string:~*req.Account:~*req.Destination
				if strings.HasPrefix(value, utils.DynamicDataPrefix) {
					continue
				}
				idxKey = utils.ConcatenatedKey(fltr.Type, fltr.Element, value)
			} else if strings.HasPrefix(value, utils.DynamicDataPrefix) {
				idxKey = utils.ConcatenatedKey(fltr.Type, value, fltr.Element)
			} else {
				continue // none of the element or value are dynamic, so we do not index
			}
			if !newRules.HasKey(idxKey) {
				removeRules[idxKey] = true
			} else {
				oldRules[idxKey] = true
			}
		}
	}

	needsRebuild := len(removeRules) != 0 // nothing to remove
	if !needsRebuild {                    //check if we added something in remove rules by checking the difference betweend remove rules and old rules
		for key := range newRules {
			if needsRebuild = !oldRules.HasKey(key); needsRebuild {
				break
			}
		}
		if !needsRebuild {
			return // nothing to change
		}
	}

	tntFltrID := utils.ConcatenatedKey(newFltr.Tenant, newFltr.ID)
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, utils.CacheReverseFilterIndexes+tntFltrID)
	defer guardian.Guardian.UnguardIDs(refID)
	var rcvIndexes map[string]utils.StringMap
	// get all the reverse indexes for the specific filter from db
	if rcvIndexes, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ReverseFilterIndexes], tntFltrID,
		utils.EmptyString, nil); err != nil {
		if err != utils.ErrNotFound {
			return //
		}
		err = nil // if the error is NOT_FOUND, it means that no indexes were found for this filter, so no need to update
		return
	}
	removeIndexKeys := removeRules.Slice()

	for idxItmType, index := range rcvIndexes {
		switch idxItmType {
		case utils.CacheChargerFilterIndexes:
			// remove the indexes from this filter for this partition
			if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheChargerProfiles,
				tnt, removeIndexKeys, index); err != nil {
				return
			}
			// we removed the old reverse indexes, now we have to compute the new ones
			chargerIDs := index.Slice()
			if _, err = ComputeChargerIndexes(dm, newFltr.Tenant, &chargerIDs,
				utils.NonTransactional); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeFilterIndexesForFilter removes the itemID for the index keys
// used to remove the old indexes when a filter is updated
func removeFilterIndexesForFilter(dm *DataManager, idxItmType, cacheItmType, tnt string,
	removeIndexKeys []string, itemIDs utils.StringMap) (err error) {
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tnt)
	defer guardian.Guardian.UnguardIDs(refID)
	for _, idxKey := range removeIndexKeys { // delete old filters indexes for this item
		var remIndx map[string]utils.StringMap
		if remIndx, err = dm.GetFilterIndexes(idxItmType, tnt,
			utils.EmptyString, nil); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			err = nil
			continue
		}
		for idx := range itemIDs {
			delete(remIndx[idxKey], idx)
		}

		fltrIndexer := NewFilterIndexer(dm, utils.CacheInstanceToPrefix[cacheItmType], tnt)
		fltrIndexer.indexes = remIndx
		if err = fltrIndexer.StoreIndexes(true, utils.NonTransactional); err != nil {
			return
		}
	}
	return
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
		indexes[idxItmType].Copy(map[string]bool{
			itemID: true,
		})
		indexerKey := utils.ConcatenatedKey(tnt, fltrID)
		if ctx != utils.EmptyString {
			indexerKey = utils.ConcatenatedKey(tnt, ctx)
		}
		fltrIndexer := NewFilterIndexer(dm, utils.ReverseFilterIndexes, indexerKey)
		fltrIndexer.indexes = indexes
		if err = fltrIndexer.StoreIndexes(true, utils.NonTransactional); err != nil { // it will remove from cache the old ones
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
			continue // already removed
		}
		delete(indexes[idxItmType], itemID) // delete index from map

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
