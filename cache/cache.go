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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const (
	PREFIX_LEN = 4
	ADD        = "ADD"
	REM        = "REM"
	REM_PREFIX = "PRF"
)

var (
	cache    cacheStore
	cacheMux sync.RWMutex
	cfg      *config.CacheConfig
	// transaction stuff
	transactionBuffer map[string][]*transactionItem // Queue tasks based on transactionID
	transBufMux       sync.Mutex                    // Protects the transactionBuffer
	transactionMux    sync.Mutex                    // Queue transactions on commit
)

type transactionItem struct {
	verb  string      // action which will be executed on cache
	key   string      // item key
	value interface{} // item value
}

func init() {
	NewCache(nil)
}

func NewCache(cacheCfg *config.CacheConfig) {
	cfg = cacheCfg
	cache = newLruStore()
	transactionBuffer = make(map[string][]*transactionItem) // map[transactionID][]*transactionItem
}

func BeginTransaction() string {
	transID := utils.GenUUID()
	transBufMux.Lock()
	transactionBuffer[transID] = make([]*transactionItem, 0)
	transBufMux.Unlock()
	return transID
}

func RollbackTransaction(transID string) {
	transBufMux.Lock()
	delete(transactionBuffer, transID)
	transBufMux.Unlock()
}

func CommitTransaction(transID string) {
	transactionMux.Lock()
	transBufMux.Lock()
	// apply all transactioned items in one shot
	cacheMux.Lock()
	for _, item := range transactionBuffer[transID] {
		switch item.verb {
		case REM:
			RemKey(item.key, true, transID)
		case REM_PREFIX:
			RemPrefixKey(item.key, true, transID)
		case ADD:
			Set(item.key, item.value, true, transID)
		}
	}
	cacheMux.Unlock()
	delete(transactionBuffer, transID)
	transBufMux.Unlock()
	transactionMux.Unlock()
}

// The function to be used to cache a key/value pair when expiration is not needed
func Set(key string, value interface{}, commit bool, transID string) {
	if commit {
		if transID == "" { // Lock locally
			cacheMux.Lock()
			defer cacheMux.Unlock()
		}
		cache.Put(key, value)
		//log.Println("ADD: ", key)
	} else {
		transBufMux.Lock()
		transactionBuffer[transID] = append(transactionBuffer[transID], &transactionItem{verb: ADD, key: key, value: value})
		transBufMux.Unlock()
	}
}

func RemKey(key string, commit bool, transID string) {
	if commit {
		if transID == "" { // Lock per operation not transaction
			cacheMux.Lock()
			defer cacheMux.Unlock()
		}
		cache.Delete(key)
	} else {
		transBufMux.Lock()
		transactionBuffer[transID] = append(transactionBuffer[transID], &transactionItem{verb: REM, key: key})
		transBufMux.Unlock()
	}
}

func RemPrefixKey(prefix string, commit bool, transID string) {
	if commit {
		if transID == "" { // Lock locally
			cacheMux.Lock()
			defer cacheMux.Unlock()
		}
		cache.DeletePrefix(prefix)
	} else {
		transBufMux.Lock()
		transactionBuffer[transID] = append(transactionBuffer[transID], &transactionItem{verb: REM_PREFIX, key: prefix})
		transBufMux.Unlock()
	}
}

// Delete all keys from cache
func Flush() {
	cacheMux.Lock()
	cache = newLruStore()
	cacheMux.Unlock()
}

// The function to extract a value for a key that never expire
func Get(key string) (interface{}, bool) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.Get(key)
}

func CountEntries(prefix string) (result int) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.CountEntriesForPrefix(prefix)
}

func GetEntriesKeys(prefix string) (keys []string) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.GetKeysForPrefix(prefix)
}
