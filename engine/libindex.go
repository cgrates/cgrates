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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

var (
	filterIndexType = utils.StringMap{
		utils.MetaString: true,
		utils.MetaPrefix: true,
	}
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
		if !strings.HasPrefix(idxItmType, utils.Meta) {
			idxItmType = strings.Split(idxItmType, utils.CONCATENATED_KEY_SEP)[0]
		}
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
		case utils.CacheThresholdFilterIndexes:
			// remove the indexes from this filter for this partition
			if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheThresholdProfiles,
				tnt, removeIndexKeys, index); err != nil {
				return
			}
			// we removed the old reverse indexes, now we have to compute the new ones
			thresholdIDs := index.Slice()
			if _, err = ComputeThresholdIndexes(dm, newFltr.Tenant, &thresholdIDs,
				utils.NonTransactional); err != nil {
				return err
			}
		case utils.CacheResourceFilterIndexes:
			// remove the indexes from this filter for this partition
			if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheResourceProfiles,
				tnt, removeIndexKeys, index); err != nil {
				return
			}
			// we removed the old reverse indexes, now we have to compute the new ones
			resourceIDs := index.Slice()
			if _, err = ComputeResourceIndexes(dm, newFltr.Tenant, &resourceIDs,
				utils.NonTransactional); err != nil {
				return err
			}
		case utils.CacheSupplierFilterIndexes:
			// remove the indexes from this filter for this partition
			if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheSupplierProfiles,
				tnt, removeIndexKeys, index); err != nil {
				return
			}
			// we removed the old reverse indexes, now we have to compute the new ones
			supplierIDs := index.Slice()
			if _, err = ComputeSupplierIndexes(dm, newFltr.Tenant, &supplierIDs,
				utils.NonTransactional); err != nil {
				return err
			}
		case utils.CacheStatFilterIndexes:

			// remove the indexes from this filter for this partition
			if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheStatQueueProfiles,
				tnt, removeIndexKeys, index); err != nil {
				return
			}
			// we removed the old reverse indexes, now we have to compute the new ones
			statQueueIDs := index.Slice()
			if _, err = ComputeStatIndexes(dm, newFltr.Tenant, &statQueueIDs,
				utils.NonTransactional); err != nil {
				return err
			}
		case utils.CacheAttributeFilterIndexes:
			attributeIDs := index.Slice()
			for _, attrID := range attributeIDs {
				var ap *AttributeProfile
				if ap, err = dm.GetAttributeProfile(newFltr.Tenant, attrID,
					true, false, utils.NonTransactional); err != nil {
					return
				}
				for _, ctx := range ap.Contexts {
					tntCtx := utils.ConcatenatedKey(newFltr.Tenant, ctx)
					if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheAttributeProfiles,
						tntCtx, // remove the indexes for the filter
						removeIndexKeys, index); err != nil {
						return
					}
					if _, err = ComputeAttributeIndexes(dm, newFltr.Tenant, ctx, &[]string{attrID},
						utils.NonTransactional); err != nil {
						return err
					}
				}
			}
		case utils.CacheDispatcherFilterIndexes:
			dispatcherIDs := index.Slice()
			for _, dspID := range dispatcherIDs {
				var dpp *DispatcherProfile
				if dpp, err = dm.GetDispatcherProfile(newFltr.Tenant, dspID,
					true, false, utils.NonTransactional); err != nil {
					return
				}
				for _, subsys := range dpp.Subsystems {
					tntSubsys := utils.ConcatenatedKey(newFltr.Tenant, subsys)
					if err = removeFilterIndexesForFilter(dm, idxItmType, utils.CacheDispatcherProfiles,
						tntSubsys, // remove the indexes for the filter
						removeIndexKeys, index); err != nil {
						return
					}
					if _, err = ComputeDispatcherIndexes(dm, newFltr.Tenant, subsys, &[]string{dspID},
						utils.NonTransactional); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// removeFilterIndexesForFilter removes itemIDs from the specified filter index keys.
// Used to update the index map when a filter is modified.
func removeFilterIndexesForFilter(dm *DataManager, idxItmType, cacheItmType, tnt string,
	removeIndexKeys []string, itemIDs utils.StringMap) error {
	if len(removeIndexKeys) == 0 {
		return nil // no indexes to remove
	}
	refID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout, idxItmType+tnt)
	defer guardian.Guardian.UnguardIDs(refID)

	// Retrieve current filter fltrIdx.
	fltrIdx, err := dm.GetFilterIndexes(idxItmType, tnt,
		utils.EmptyString, nil)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil // nothing to remove
		}
		return err
	}

	// Remove itemIDs from the specified index keys.
	for _, idxKey := range removeIndexKeys {
		for itemID := range itemIDs {
			delete(fltrIdx[idxKey], itemID)
		}
	}

	// Store the updated indexes.
	fltrIndexer := NewFilterIndexer(dm, utils.CacheInstanceToPrefix[cacheItmType], tnt)
	fltrIndexer.indexes = fltrIdx
	return fltrIndexer.StoreIndexes(true, utils.NonTransactional)
}

// addReverseFilterIndexForFilter will add a reference for the filter in reverse filter indexes
func addReverseFilterIndexForFilter(dm *DataManager, idxItmType, tnt,
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
		if indexes[idxItmType] == nil {
			indexes[idxItmType] = make(utils.StringMap)
		}
		indexes[idxItmType].Copy(map[string]bool{
			itemID: true,
		})
		indexerKey := utils.ConcatenatedKey(tnt, fltrID)
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
func removeReverseFilterIndexForFilter(dm *DataManager, idxItmType, tnt, itemID string, filterIDs []string) (err error) {
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
			guardian.Guardian.UnguardIDs(refID)
			if err != utils.ErrNotFound {
				return
			}
			err = nil
			continue // already removed
		}

		delete(indexes[idxItmType], itemID) // delete index from map

		indexerKey := utils.ConcatenatedKey(tnt, fltrID)
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
