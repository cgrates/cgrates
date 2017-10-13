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
	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/utils"
)

func NewDataManager(dataDB DataDB) *DataManager {
	return &DataManager{dataDB: dataDB}
}

// DataManager is the data storage manager for CGRateS
// transparently manages data retrieval, further serialization and caching
type DataManager struct {
	dataDB DataDB
}

// DataDB exports access to dataDB
func (dm *DataManager) DataDB() DataDB {
	return dm.dataDB
}

// GetStatQueue retrieves a StatQueue from dataDB
// handles caching and deserialization of metrics
func (dm *DataManager) GetStatQueue(tenant, id string, skipCache bool, transactionID string) (sq *StatQueue, err error) {
	key := utils.StatQueuePrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*StatQueue), nil
		}
	}
	ssq, err := dm.dataDB.GetStoredStatQueue(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	if sq, err = ssq.AsStatQueue(dm.dataDB.Marshaler()); err != nil {
		return nil, err
	}
	cache.Set(key, sq, cacheCommit(transactionID), transactionID)
	return
}

// SetStatQueue converts to StoredStatQueue and stores the result in dataDB
func (dm *DataManager) SetStatQueue(sq *StatQueue) (err error) {
	ssq, err := NewStoredStatQueue(sq, dm.dataDB.Marshaler())
	if err != nil {
		return err
	}
	return dm.dataDB.SetStoredStatQueue(ssq)
}

// RemStatQueue removes the StoredStatQueue and clears the cache for StatQueue
func (dm *DataManager) RemStatQueue(tenant, id string, transactionID string) (err error) {
	if err = dm.dataDB.RemStoredStatQueue(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.StatQueuePrefix+utils.ConcatenatedKey(tenant, id), cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetFilter(tenant, id string, skipCache bool, transactionID string) (fltr *Filter, err error) {
	key := utils.FilterPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Filter), nil
		}
	}
	fltr, err = dm.dataDB.GetFilterDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, fltr, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetFilter(fltr *Filter) (err error) {
	return dm.DataDB().SetFilterDrv(fltr)
}

func (dm *DataManager) RemoveFilter(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveFilterDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.FilterPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) GetThreshold(tenant, id string, skipCache bool, transactionID string) (th *Threshold, err error) {
	key := utils.ThresholdPrefix + utils.ConcatenatedKey(tenant, id)
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Threshold), nil
		}
	}
	th, err = dm.dataDB.GetThresholdDrv(tenant, id)
	if err != nil {
		if err == utils.ErrNotFound {
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, th, cacheCommit(transactionID), transactionID)
	return
}

func (dm *DataManager) SetThreshold(th *Threshold) (err error) {
	return dm.DataDB().SetThresholdDrv(th)
}

func (dm *DataManager) RemoveThreshold(tenant, id, transactionID string) (err error) {
	if err = dm.DataDB().RemoveThresholdDrv(tenant, id); err != nil {
		return
	}
	cache.RemKey(utils.ThresholdPrefix+utils.ConcatenatedKey(tenant, id),
		cacheCommit(transactionID), transactionID)
	return
}
