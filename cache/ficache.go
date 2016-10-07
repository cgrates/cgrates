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
package cache

import (
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type filtersIndexer interface {
	Index(itemIDs []string) (map[string]map[string]utils.StringMap, error) // Index items based on ID, nil for all
}

func newFiCache() *fiCache {
	return &fiCache{cache: make(map[string]map[string]map[string]utils.StringMap)}
}

// FiCache is a cache handling filter indexing for various services
type fiCache struct {
	cache        map[string]map[string]map[string]utils.StringMap // map[serviceID]map[fieldName]map[fieldValue]utils.StringMap[itemID]
	sync.RWMutex                                                  // protects cache
	indexers     map[string]filtersIndexer
	indxrsMux    sync.RWMutex // protects indexers
}

func (fiCh *fiCache) registerIndexer(id string, idxr filtersIndexer) { // Not protected but also should
	fiCh.indxrsMux.Lock()
	fiCh.indexers[id] = idxr
	fiCh.indxrsMux.Unlock()
}

func (fiCh *fiCache) indexForIndexer(indxrID string, itemIDs []string) error {
	fiCh.indxrsMux.RLock()
	idxr, hasIt := fiCh.indexers[indxrID]
	fiCh.indxrsMux.RUnlock()
	if !hasIt {
		return utils.ErrNotFound
	}
	newIndxMp, err := idxr.Index(itemIDs)
	if err != nil {
		return err
	}
	fiCh.Lock()
	if _, hasIt := fiCh.indexers[indxrID]; !hasIt {
		fiCh.cache[indxrID] = newIndxMp
		return nil
	}
	// Merge old index cache with new one
	for fldNameKey, mpFldName := range newIndxMp {
		if _, hasIt := fiCh.cache[indxrID][fldNameKey]; !hasIt {
			fiCh.cache[indxrID][fldNameKey] = mpFldName
		} else {
			for fldValKey, strMap := range mpFldName {
				if _, hasIt := fiCh.cache[indxrID][fldNameKey][fldValKey]; !hasIt {
					fiCh.cache[indxrID][fldNameKey][fldValKey] = strMap
				} else {
					for resIDKey := range strMap {
						fiCh.cache[indxrID][fldNameKey][fldValKey][resIDKey] = true
					}
				}
			}
		}
	}
	fiCh.Unlock()
	return nil
}

// Empty the cache for specific indexers
func (fiCh *fiCache) flushForIndexers(indxrIDs []string) {
	fiCh.Lock()
	defer fiCh.Unlock()
	if indxrIDs == nil {
		fiCh.cache = make(map[string]map[string]map[string]utils.StringMap)
		return
	}
	for _, indxID := range indxrIDs {
		delete(fiCh.cache, indxID)
	}
}
