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

func NewReqFilterIndexer(dataDB AccountingStorage, dbKey string) (*ReqFilterIndexer, error) {
	indexes, err := dataDB.GetReqFilterIndexes(dbKey)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}
	if indexes == nil {
		indexes = make(map[string]map[string]utils.StringMap)
	}
	return &ReqFilterIndexer{dataDB: dataDB, dbKey: dbKey, indexes: indexes, chngdIndxKeys: make(utils.StringMap)}, nil
}

// ReqFilterIndexer is a centralized indexer for all data sources using RequestFilter
// retrieves and stores it's data from/to dataDB
// not thread safe, meant to be used as logic within other code blocks
type ReqFilterIndexer struct {
	indexes       map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue]utils.StringMap[resourceID]
	dataDB        AccountingStorage
	dbKey         string          // get/store the result from/into this key
	chngdIndxKeys utils.StringMap // keep record of the changed fieldName:fieldValue pair so we can re-cache wisely
}

// ChangedKeys returns the changed keys from original indexes so we can reload wisely
func (rfi *ReqFilterIndexer) ChangedKeys() utils.StringMap {
	return rfi.chngdIndxKeys
}

// IndexFilters parses reqFltrs, adding itemID in the indexes and marks the changed keys in chngdIndxKeys
func (rfi *ReqFilterIndexer) IndexFilters(itemID string, reqFltrs []*RequestFilter) {
	var hasMetaString bool
	for _, fltr := range reqFltrs {
		if fltr.Type != MetaString {
			continue
		}
		hasMetaString = true // Mark that we found at least one metatring so we don't index globally
		if _, hastIt := rfi.indexes[fltr.FieldName]; !hastIt {
			rfi.indexes[fltr.FieldName] = make(map[string]utils.StringMap)
		}
		for _, fldVal := range fltr.Values {
			if _, hasIt := rfi.indexes[fltr.FieldName][fldVal]; !hasIt {
				rfi.indexes[fltr.FieldName][fldVal] = make(utils.StringMap)
			}
			rfi.indexes[fltr.FieldName][fldVal][itemID] = true
			rfi.chngdIndxKeys[utils.ConcatenatedKey(fltr.FieldName, fldVal)] = true
		}
	}
	if !hasMetaString {
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE] = make(map[string]utils.StringMap)
		}
		if _, hasIt := rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE]; !hasIt {
			rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE] = make(utils.StringMap)
		}
		rfi.indexes[utils.NOT_AVAILABLE][utils.NOT_AVAILABLE][itemID] = true // Fields without real field index will be located in map[NOT_AVAILABLE][NOT_AVAILABLE][rl.ID]
	}
	return
}

// StoreIndexes handles storing the indexes to dataDB
func (rfi *ReqFilterIndexer) StoreIndexes() error {
	return rfi.dataDB.SetReqFilterIndexes(rfi.dbKey, rfi.indexes)
}
