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

func NewFilterIndexer(dm *DataManager, itemType, dbKeySuffix string) *FilterIndexer {
	return &FilterIndexer{dm: dm, itemType: itemType, dbKeySuffix: dbKeySuffix,
		indexes:       make(map[string]utils.StringMap),
		chngdIndxKeys: make(utils.StringMap)}
}

// FilterIndexer is a centralized indexer for all data sources using RequestFilter
// retrieves and stores it's data from/to dataDB
// not thread safe, meant to be used as logic within other code blocks
type FilterIndexer struct {
	indexes       map[string]utils.StringMap // map[fieldName:fieldValue]utils.StringMap[itemID]
	dm            *DataManager
	itemType      string
	dbKeySuffix   string          // get/store the result from/into this key
	chngdIndxKeys utils.StringMap // keep record of the changed fieldName:fieldValue pair so we can re-cache wisely
}

// IndexTPFilter parses reqFltrs, adding itemID in the indexes and marks the changed keys in chngdIndxKeys
func (rfi *FilterIndexer) IndexTPFilter(tpFltr *utils.TPFilterProfile, itemID string) {
	for _, fltr := range tpFltr.Filters {
		switch fltr.Type {
		case utils.MetaString:
			for _, fldVal := range fltr.Values {
				concatKey := utils.ConcatenatedKey(fltr.Type, fltr.Element, fldVal)
				if _, hasIt := rfi.indexes[concatKey]; !hasIt {
					rfi.indexes[concatKey] = make(utils.StringMap)
				}
				rfi.indexes[concatKey][itemID] = true
				rfi.chngdIndxKeys[concatKey] = true
			}
		case utils.MetaPrefix:
			for _, fldVal := range fltr.Values {
				concatKey := utils.ConcatenatedKey(fltr.Type, fltr.Element, fldVal)
				if _, hasIt := rfi.indexes[concatKey]; !hasIt {
					rfi.indexes[concatKey] = make(utils.StringMap)
				}
				rfi.indexes[concatKey][itemID] = true
				rfi.chngdIndxKeys[concatKey] = true
			}
		case utils.META_NONE:
			concatKey := utils.ConcatenatedKey(utils.META_NONE, utils.ANY, utils.ANY)
			if _, hasIt := rfi.indexes[concatKey]; !hasIt {
				rfi.indexes[concatKey] = make(utils.StringMap)
			}
			rfi.indexes[concatKey][itemID] = true
			rfi.chngdIndxKeys[concatKey] = true
		}
	}
	return
}

func (rfi *FilterIndexer) cacheRemItemType() { // ToDo: tune here by removing per item
	switch rfi.itemType {

	case utils.ThresholdProfilePrefix:
		Cache.Clear([]string{utils.CacheThresholdFilterIndexes})

	case utils.ResourceProfilesPrefix:
		Cache.Clear([]string{utils.CacheResourceFilterIndexes})

	case utils.StatQueueProfilePrefix:
		Cache.Clear([]string{utils.CacheStatFilterIndexes})

	case utils.RouteProfilePrefix:
		Cache.Clear([]string{utils.CacheRouteFilterIndexes})

	case utils.AttributeProfilePrefix:
		Cache.Clear([]string{utils.CacheAttributeFilterIndexes})

	case utils.ChargerProfilePrefix:
		Cache.Clear([]string{utils.CacheChargerFilterIndexes})

	case utils.DispatcherProfilePrefix:
		Cache.Clear([]string{utils.CacheDispatcherFilterIndexes})

	case utils.RateProfilePrefix:
		Cache.Clear([]string{utils.CacheRateProfilesFilterIndexes})

	case utils.RatePrefix:
		Cache.Clear([]string{utils.CacheRateFilterIndexes})
	}
}

// StoreIndexes handles storing the indexes to dataDB
func (rfi *FilterIndexer) StoreIndexes(commit bool, transactionID string) (err error) {
	lockID := utils.CacheInstanceToPrefix[utils.PrefixToIndexCache[rfi.itemType]] + rfi.dbKeySuffix
	refID := guardian.Guardian.GuardIDs("",
		config.CgrConfig().GeneralCfg().LockingTimeout, lockID)
	defer guardian.Guardian.UnguardIDs(refID)
	if err = rfi.dm.SetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		rfi.indexes, commit, transactionID); err != nil {
		return
	}
	rfi.cacheRemItemType()
	return
}

//Populate FilterIndexer.indexes with specific fieldName:fieldValue , item
func (rfi *FilterIndexer) loadFldNameFldValIndex(filterType, fldName, fldVal string) error {
	rcvIdx, err := rfi.dm.GetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix, filterType,
		map[string]string{fldName: fldVal})
	if err != nil {
		return err
	}
	for fldName, nameValMp := range rcvIdx {
		if _, has := rfi.indexes[fldName]; !has {
			rfi.indexes[fldName] = make(utils.StringMap)
		}
		rfi.indexes[fldName] = nameValMp
	}
	return nil
}

//RemoveItemFromIndex remove Indexes for a specific itemID
func (rfi *FilterIndexer) RemoveItemFromIndex(tenant, itemID string, oldFilters []string) (err error) {
	var filterIDs []string
	switch rfi.itemType {
	case utils.ThresholdProfilePrefix:
		th, err := rfi.dm.GetThresholdProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if th != nil {
			filterIDs = make([]string, len(th.FilterIDs))
			for i, fltrID := range th.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.AttributeProfilePrefix:
		attrPrf, err := rfi.dm.GetAttributeProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if attrPrf != nil {
			filterIDs = make([]string, len(attrPrf.FilterIDs))
			for i, fltrID := range attrPrf.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.ResourceProfilesPrefix:
		res, err := rfi.dm.GetResourceProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if res != nil {
			filterIDs = make([]string, len(res.FilterIDs))
			for i, fltrID := range res.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.StatQueueProfilePrefix:
		stq, err := rfi.dm.GetStatQueueProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if stq != nil {
			filterIDs = make([]string, len(stq.FilterIDs))
			for i, fltrID := range stq.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.RouteProfilePrefix:
		spp, err := rfi.dm.GetRouteProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if spp != nil {
			filterIDs = make([]string, len(spp.FilterIDs))
			for i, fltrID := range spp.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.ChargerProfilePrefix:
		cpp, err := rfi.dm.GetChargerProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if cpp != nil {
			filterIDs = make([]string, len(cpp.FilterIDs))
			for i, fltrID := range cpp.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.DispatcherProfilePrefix:
		dpp, err := rfi.dm.GetDispatcherProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if dpp != nil {
			filterIDs = make([]string, len(dpp.FilterIDs))
			for i, fltrID := range dpp.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.RateProfilePrefix:
		rpp, err := rfi.dm.GetRateProfile(tenant, itemID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if rpp != nil {
			filterIDs = make([]string, len(rpp.FilterIDs))
			for i, fltrID := range rpp.FilterIDs {
				filterIDs[i] = fltrID
			}
		}
	case utils.RatePrefix:
		composedIDs := utils.SplitConcatenatedKey(itemID)
		rppID, rateKey := composedIDs[0], composedIDs[1]
		rpp, err := rfi.dm.GetRateProfile(tenant, rppID, true, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			return err
		}
		if rpp != nil {
			if rate, has := rpp.Rates[rateKey]; has {
				filterIDs = make([]string, len(rate.FilterIDs))
				for i, fltrID := range rate.FilterIDs {
					filterIDs[i] = fltrID
				}
			}
		}
	default:
	}
	if len(filterIDs) == 0 {
		filterIDs = []string{utils.META_NONE}
	}
	for _, oldFltr := range oldFilters {
		filterIDs = append(filterIDs, oldFltr)
	}
	for _, fltrID := range filterIDs {
		var fltr *Filter
		if fltrID == utils.META_NONE {
			fltr = &Filter{
				Tenant: tenant,
				ID:     itemID,
				Rules: []*FilterRule{
					{
						Type:    utils.META_NONE,
						Element: utils.META_ANY,
						Values:  []string{utils.META_ANY},
					},
				},
			}
		} else if fltr, err = rfi.dm.GetFilter(tenant, fltrID,
			true, false, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = fmt.Errorf("broken reference to filter: %+v for itemType: %+v and ID: %+v",
					fltrID, rfi.itemType, itemID)
			}
			return err
		}
		for _, flt := range fltr.Rules {
			var fldType, fldName string
			var fldVals []string
			if utils.SliceHasMember([]string{utils.META_NONE, utils.MetaPrefix, utils.MetaString}, flt.Type) {
				fldType, fldName = flt.Type, flt.Element
				fldVals = flt.Values
			}
			for _, fldVal := range fldVals {
				if err = rfi.loadFldNameFldValIndex(fldType,
					fldName, fldVal); err != nil && err != utils.ErrNotFound {
					return err
				}
			}
		}
	}
	for _, itmMp := range rfi.indexes {
		if _, has := itmMp[itemID]; has {
			delete(itmMp, itemID) //Force deleting in driver
		}
	}
	return rfi.StoreIndexes(false, utils.NonTransactional)
}

//createAndIndex create indexes for an item
func createAndIndex(itmPrfx, tenant, context, itemID string, filterIDs []string, dm *DataManager) (err error) {
	indexerKey := tenant
	if context != "" {
		indexerKey = utils.ConcatenatedKey(tenant, context)
	}
	indexer := NewFilterIndexer(dm, itmPrfx, indexerKey)
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
				Tenant: tenant,
				ID:     itemID,
				Rules: []*FilterRule{
					{
						Type:    utils.META_NONE,
						Element: utils.META_ANY,
						Values:  []string{utils.META_ANY},
					},
				},
			}
		} else if fltr, err = dm.GetFilter(tenant, fltrID,
			true, false, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = fmt.Errorf("broken reference to filter: %+v for itemType: %+v and ID: %+v",
					fltrID, itmPrfx, itemID)
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
		indexer.IndexTPFilter(FilterToTPFilter(fltr), itemID)
	}
	return indexer.StoreIndexes(true, utils.NonTransactional)
}
