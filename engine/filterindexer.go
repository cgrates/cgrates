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

func NewFilterIndexer(dm *DataManager, itemType, dbKeySuffix string) *FilterIndexer {
	return &FilterIndexer{dm: dm, itemType: itemType, dbKeySuffix: dbKeySuffix,
		indexes:          make(map[string]utils.StringMap),
		reveseIndex:      make(map[string]utils.StringMap),
		chngdIndxKeys:    make(utils.StringMap),
		chngdRevIndxKeys: make(utils.StringMap)}
}

// FilterIndexer is a centralized indexer for all data sources using RequestFilter
// retrieves and stores it's data from/to dataDB
// not thread safe, meant to be used as logic within other code blocks
type FilterIndexer struct {
	indexes          map[string]utils.StringMap // map[fieldName:fieldValue]utils.StringMap[itemID]
	reveseIndex      map[string]utils.StringMap // map[itemID]utils.StringMap[fieldName:fieldValue]
	dm               *DataManager
	itemType         string
	dbKeySuffix      string          // get/store the result from/into this key
	chngdIndxKeys    utils.StringMap // keep record of the changed fieldName:fieldValue pair so we can re-cache wisely
	chngdRevIndxKeys utils.StringMap // keep record of the changed itemID so we can re-cache wisely
}

// ChangedKeys returns the changed keys from original indexes so we can reload wisely
func (rfi *FilterIndexer) ChangedKeys(reverse bool) utils.StringMap {
	if reverse {
		return rfi.chngdRevIndxKeys
	}
	return rfi.chngdIndxKeys
}

// IndexTPFilter parses reqFltrs, adding itemID in the indexes and marks the changed keys in chngdIndxKeys
func (rfi *FilterIndexer) IndexTPFilter(tpFltr *utils.TPFilterProfile, itemID string) {
	for _, fltr := range tpFltr.Filters {
		switch fltr.Type {
		case MetaString:
			for _, fldVal := range fltr.Values {
				concatKey := utils.ConcatenatedKey(fltr.Type, fltr.FieldName, fldVal)
				if _, hasIt := rfi.indexes[concatKey]; !hasIt {
					rfi.indexes[concatKey] = make(utils.StringMap)
				}
				rfi.indexes[concatKey][itemID] = true
				if _, hasIt := rfi.reveseIndex[itemID]; !hasIt {
					rfi.reveseIndex[itemID] = make(utils.StringMap)
				}
				rfi.reveseIndex[itemID][concatKey] = true
				rfi.chngdIndxKeys[concatKey] = true
			}
			rfi.chngdRevIndxKeys[itemID] = true
		case MetaPrefix:
			for _, fldVal := range fltr.Values {
				concatKey := utils.ConcatenatedKey(fltr.Type, fltr.FieldName, fldVal)
				if _, hasIt := rfi.indexes[concatKey]; !hasIt {
					rfi.indexes[concatKey] = make(utils.StringMap)
				}
				rfi.indexes[concatKey][itemID] = true
				if _, hasIt := rfi.reveseIndex[itemID]; !hasIt {
					rfi.reveseIndex[itemID] = make(utils.StringMap)
				}
				rfi.reveseIndex[itemID][concatKey] = true
				rfi.chngdIndxKeys[concatKey] = true
			}
			rfi.chngdRevIndxKeys[itemID] = true
		default:
			concatKey := utils.ConcatenatedKey(utils.MetaDefault, utils.ANY, utils.ANY)
			if _, hasIt := rfi.indexes[concatKey]; !hasIt {
				rfi.indexes[concatKey] = make(utils.StringMap)
			}
			if _, hasIt := rfi.reveseIndex[itemID]; !hasIt {
				rfi.reveseIndex[itemID] = make(utils.StringMap)
			}
			rfi.reveseIndex[itemID][concatKey] = true
			rfi.indexes[concatKey][itemID] = true // Fields without real field index will be located in map[*any:*any][rl.ID]
		}
	}

	return
}

func (rfi *FilterIndexer) cacheRemItemType() { // ToDo: tune here by removing per item
	switch rfi.itemType {

	case utils.ThresholdProfilePrefix:
		Cache.Clear([]string{utils.CacheThresholdFilterIndexes, utils.CacheThresholdFilterRevIndexes})

	case utils.ResourceProfilesPrefix:
		Cache.Clear([]string{utils.CacheResourceFilterIndexes, utils.CacheResourceFilterRevIndexes})

	case utils.StatQueueProfilePrefix:
		Cache.Clear([]string{utils.CacheStatFilterIndexes, utils.CacheStatFilterRevIndexes})

	case utils.SupplierProfilePrefix:
		Cache.Clear([]string{utils.CacheSupplierFilterIndexes, utils.CacheSupplierFilterRevIndexes})

	case utils.AttributeProfilePrefix:
		Cache.Clear([]string{utils.CacheAttributeFilterIndexes, utils.CacheAttributeFilterRevIndexes})

	}
}

// StoreIndexes handles storing the indexes to dataDB
func (rfi *FilterIndexer) StoreIndexes(commit bool, transactionID string) (err error) {
	if err = rfi.dm.SetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		rfi.indexes, commit, transactionID); err != nil {
		return
	}
	if err = rfi.dm.SetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi.itemType], rfi.dbKeySuffix,
		rfi.reveseIndex, commit, transactionID); err != nil {
		return
	}
	rfi.cacheRemItemType()
	return
}

//Populate the FilterIndexer.reveseIndex for specifil itemID
func (rfi *FilterIndexer) loadItemReverseIndex(filterType, itemID string) (err error) {
	rcvReveseIdx, err := rfi.dm.GetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi.itemType], rfi.dbKeySuffix,
		map[string]string{itemID: ""})
	if err != nil {
		return err
	}
	for _, val2 := range rcvReveseIdx {
		if _, has := rfi.reveseIndex[itemID]; !has {
			rfi.reveseIndex[itemID] = make(utils.StringMap)
		}
		rfi.reveseIndex[itemID] = val2
	}
	return err
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
func (rfi *FilterIndexer) RemoveItemFromIndex(itemID string) (err error) {
	if err = rfi.loadItemReverseIndex(MetaString, itemID); err != nil {
		return err
	}
	for key, _ := range rfi.reveseIndex[itemID] {
		kSplt := strings.Split(key, utils.CONCATENATED_KEY_SEP)
		if len(kSplt) != 3 {
			return fmt.Errorf("Malformed key in db: %s", key)
		}
		if err = rfi.loadFldNameFldValIndex(kSplt[0], kSplt[1], kSplt[2]); err != nil {
			return err
		}
	}
	for _, itmMp := range rfi.indexes {
		for range itmMp {
			if _, has := itmMp[itemID]; has {
				delete(itmMp, itemID) //Force deleting in driver
			}
		}
	}
	rfi.reveseIndex[itemID] = make(utils.StringMap) //Force deleting in driver
	if err = rfi.dm.SetFilterIndexes(
		utils.PrefixToIndexCache[rfi.itemType], rfi.dbKeySuffix,
		rfi.indexes, false, utils.NonTransactional); err != nil {
		return
	}
	if err = rfi.dm.SetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi.itemType], rfi.dbKeySuffix,
		rfi.reveseIndex, false, utils.NonTransactional); err != nil {
		return
	}
	return
}
