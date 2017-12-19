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
	"github.com/cgrates/cgrates/utils"
)

func NewReqFilterIndexer(dm *DataManager, itemType, dbKeySuffix string) *ReqFilterIndexer {
	return &ReqFilterIndexer{dm: dm, itemType: itemType, dbKeySuffix: dbKeySuffix,
		indexes:          make(map[string]map[string]utils.StringMap),
		reveseIndex:      make(map[string]map[string]utils.StringMap),
		chngdIndxKeys:    make(utils.StringMap),
		chngdRevIndxKeys: make(utils.StringMap)}
}

// ReqFilterIndexer is a centralized indexer for all data sources using RequestFilter
// retrieves and stores it's data from/to dataDB
// not thread safe, meant to be used as logic within other code blocks
type ReqFilterIndexer struct {
	indexes          map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue]utils.StringMap[itemID]
	reveseIndex      map[string]map[string]utils.StringMap // map[itemID]map[fieldName]utils.StringMap[fieldValue]
	dm               *DataManager
	itemType         string
	dbKeySuffix      string          // get/store the result from/into this key
	chngdIndxKeys    utils.StringMap // keep record of the changed fieldName:fieldValue pair so we can re-cache wisely
	chngdRevIndxKeys utils.StringMap // keep record of the changed itemID:fieldName pair so we can re-cache wisely
}

// ChangedKeys returns the changed keys from original indexes so we can reload wisely
func (rfi *ReqFilterIndexer) ChangedKeys(reverse bool) utils.StringMap {
	if reverse {
		return rfi.chngdRevIndxKeys
	}
	return rfi.chngdIndxKeys
}

// IndexFilters parses reqFltrs, adding itemID in the indexes and marks the changed keys in chngdIndxKeys
func (rfi *ReqFilterIndexer) IndexFilters(itemID string, reqFltrs *Filter) {
	var hasMetaString bool
	if _, hasIt := rfi.reveseIndex[itemID]; !hasIt {
		rfi.reveseIndex[itemID] = make(map[string]utils.StringMap)
	}
	for _, fltr := range reqFltrs.RequestFilters {
		if fltr.Type != MetaString {
			continue
		}
		hasMetaString = true // Mark that we found at least one metatring so we don't index globally
		if _, hastIt := rfi.indexes[fltr.FieldName]; !hastIt {
			rfi.indexes[fltr.FieldName] = make(map[string]utils.StringMap)
		}
		if _, hastIt := rfi.reveseIndex[itemID][fltr.FieldName]; !hastIt {
			rfi.reveseIndex[itemID][fltr.FieldName] = make(utils.StringMap)
		}
		for _, fldVal := range fltr.Values {
			if _, hasIt := rfi.indexes[fltr.FieldName][fldVal]; !hasIt {
				rfi.indexes[fltr.FieldName][fldVal] = make(utils.StringMap)
			}
			rfi.indexes[fltr.FieldName][fldVal][itemID] = true
			rfi.reveseIndex[itemID][fltr.FieldName][fldVal] = true
			rfi.chngdIndxKeys[utils.ConcatenatedKey(fltr.FieldName, fldVal)] = true
		}
		rfi.chngdRevIndxKeys[utils.ConcatenatedKey(itemID, fltr.FieldName)] = true
	}
	if !hasMetaString {
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE] = make(map[string]utils.StringMap)
		}
		if _, hastIt := rfi.reveseIndex[itemID][utils.NOT_AVAILABLE]; !hastIt {
			rfi.reveseIndex[itemID][utils.NOT_AVAILABLE] = make(utils.StringMap)
		}
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = make(utils.StringMap)
		}
		rfi.reveseIndex[itemID][utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = true
		rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE][itemID] = true // Fields without real field index will be located in map[NOT_AVAILABLE][NOT_AVAILABLE][rl.ID]
	}
	return
}

// IndexTPFilter parses reqFltrs, adding itemID in the indexes and marks the changed keys in chngdIndxKeys
func (rfi *ReqFilterIndexer) IndexTPFilter(tpFltr *utils.TPFilterProfile, itemID string) {
	var hasMetaString bool
	if _, hasIt := rfi.reveseIndex[itemID]; !hasIt {
		rfi.reveseIndex[itemID] = make(map[string]utils.StringMap)
	}
	for _, fltr := range tpFltr.Filters {
		if fltr.Type != MetaString {
			continue
		}
		hasMetaString = true // Mark that we found at least one metatring so we don't index globally
		if _, hastIt := rfi.indexes[fltr.FieldName]; !hastIt {
			rfi.indexes[fltr.FieldName] = make(map[string]utils.StringMap)
		}
		if _, hastIt := rfi.reveseIndex[itemID][fltr.FieldName]; !hastIt {
			rfi.reveseIndex[itemID][fltr.FieldName] = make(utils.StringMap)
		}
		for _, fldVal := range fltr.Values {
			if _, hasIt := rfi.indexes[fltr.FieldName][fldVal]; !hasIt {
				rfi.indexes[fltr.FieldName][fldVal] = make(utils.StringMap)
			}
			rfi.indexes[fltr.FieldName][fldVal][itemID] = true
			rfi.reveseIndex[itemID][fltr.FieldName][fldVal] = true
			rfi.chngdIndxKeys[utils.ConcatenatedKey(fltr.FieldName, fldVal)] = true
		}
		rfi.chngdRevIndxKeys[utils.ConcatenatedKey(itemID, fltr.FieldName)] = true
	}
	if !hasMetaString {
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE] = make(map[string]utils.StringMap)
		}
		if _, hastIt := rfi.reveseIndex[itemID][utils.NOT_AVAILABLE]; !hastIt {
			rfi.reveseIndex[itemID][utils.NOT_AVAILABLE] = make(utils.StringMap)
		}
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = make(utils.StringMap)
		}
		rfi.reveseIndex[itemID][utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = true
		rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE][itemID] = true // Fields without real field index will be located in map[NOT_AVAILABLE][NOT_AVAILABLE][rl.ID]
	}
	return
}

// StoreIndexes handles storing the indexes to dataDB
func (rfi *ReqFilterIndexer) StoreIndexes(update bool) (err error) {
	if err = rfi.dm.SetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false),
		rfi.indexes, update); err != nil {
		return
	}
	return rfi.dm.SetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		rfi.reveseIndex, update)
}

//Populate the ReqFilterIndexer.reveseIndex for specifil itemID
func (rfi *ReqFilterIndexer) loadItemReverseIndex(itemID string) (err error) {
	rcvReveseIdx, err := rfi.dm.GetFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true),
		map[string]string{itemID: ""})
	if err != nil {
		return err
	}
	for key2, val2 := range rcvReveseIdx[itemID] {
		if _, has := rfi.reveseIndex[itemID]; !has {
			rfi.reveseIndex[itemID] = make(map[string]utils.StringMap)
		}
		if _, has := rfi.reveseIndex[itemID][key2]; !has {
			rfi.reveseIndex[itemID][key2] = make(utils.StringMap)
		}
		rfi.reveseIndex[itemID][key2] = val2
	}
	return err
}

//Populate ReqFilterIndexer.indexes with specific fieldName,fieldValue , item
func (rfi *ReqFilterIndexer) loadFldNameFldValIndex(fldName, fldVal string) error {
	rcvIdx, err := rfi.dm.GetFilterIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, false),
		map[string]string{fldName: fldVal})
	if err != nil {
		return err
	}
	for fldName, fldValMp := range rcvIdx {
		if _, has := rfi.indexes[fldName]; !has {
			rfi.indexes[fldName] = make(map[string]utils.StringMap)
		}
		for fldVal, itmMap := range fldValMp {
			rfi.indexes[fldName][fldVal] = itmMap
		}
	}
	return nil
}

//RemoveItemFromIndex remove Indexes for a specific itemID
func (rfi *ReqFilterIndexer) RemoveItemFromIndex(itemID string) (err error) {
	if err = rfi.loadItemReverseIndex(itemID); err != nil {
		return err
	}
	for fldName, fldValMp := range rfi.reveseIndex[itemID] {
		for fldVal := range fldValMp {
			if err = rfi.loadFldNameFldValIndex(fldName, fldVal); err != nil {
				return err
			}
		}
	}
	for _, fldValMp := range rfi.indexes {
		for _, itmIDMp := range fldValMp {
			if _, has := itmIDMp[itemID]; has {
				delete(itmIDMp, itemID)
			}
		}
	}
	if err = rfi.StoreIndexes(true); err != nil {
		return
	}
	if err = rfi.dm.RemoveFilterReverseIndexes(
		GetDBIndexKey(rfi.itemType, rfi.dbKeySuffix, true), itemID); err != nil {
		return
	}
	return
}

//GetDBIndexKey return the dbKey for an specific item
func GetDBIndexKey(itemType, dbKeySuffix string, reverse bool) (dbKey string) {
	var idxPrefix, rIdxPrefix string
	switch itemType {
	case utils.ThresholdProfilePrefix:
		idxPrefix = utils.ThresholdStringIndex
		rIdxPrefix = utils.ThresholdStringRevIndex
	case utils.ResourceProfilesPrefix:
		idxPrefix = utils.ResourceProfilesStringIndex
		rIdxPrefix = utils.ResourceProfilesStringRevIndex
	case utils.StatQueueProfilePrefix:
		idxPrefix = utils.StatQueuesStringIndex
		rIdxPrefix = utils.StatQueuesStringRevIndex
	case utils.SupplierProfilePrefix:
		idxPrefix = utils.SupplierProfilesStringIndex
		rIdxPrefix = utils.SupplierProfilesStringRevIndex
	case utils.AttributeProfilePrefix:
		idxPrefix = utils.AttributeProfilesStringIndex
		rIdxPrefix = utils.AttributeProfilesStringRevIndex
	}
	if reverse {
		return rIdxPrefix + dbKeySuffix
	}
	return idxPrefix + dbKeySuffix
}
