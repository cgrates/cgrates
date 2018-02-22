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
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
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
	cfg      config.CacheConfig

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
	// ToDo: revert to nil config as soon as we handle cacheInstances properly
	dfCfg, _ := config.NewDefaultCGRConfig()
	NewCache(dfCfg.CacheCfg())
}

func NewCache(cacheCfg config.CacheConfig) {
	cfg = cacheCfg
	cache = newLRUTTL(cacheCfg)
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
		cache.Put(key, value, nil)
		//log.Println("ADD: ", key)
	} else {
		transBufMux.Lock()
		transactionBuffer[transID] = append(transactionBuffer[transID], &transactionItem{verb: ADD, key: key, value: value})
		transBufMux.Unlock()
	}
}

// RemKey removes a specific item from cache
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

// RemPrefixKey removes a complete category of data out of cache
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
	cache.Clear()
	cacheMux.Unlock()
}

// The function to extract a value for a key that never expire
func Get(key string) (interface{}, bool) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.Get(key)
}

func GetCloned(key string) (cln interface{}, err error) {
	return guardian.Guardian.Guard(func() (cln interface{}, err error) {
		cacheMux.RLock()
		origVal, hasIt := cache.Get(key)
		cacheMux.RUnlock()
		if !hasIt {
			return nil, utils.NewCGRError(utils.Cache,
				utils.NotFoundCaps, utils.ItemNotFound,
				fmt.Sprintf("item with key <%s> was not found in <%s>", key, cln))
		} else if origVal == nil {
			return nil, nil
		}
		if _, canClone := origVal.(utils.Cloner); !canClone {
			return nil, utils.NewCGRError(utils.Cache,
				utils.NotCloneableCaps, utils.ItemNotCloneable,
				fmt.Sprintf("item with key <%s> is not following cloner interface", key))
		}
		retVals := reflect.ValueOf(origVal).MethodByName("Clone").Call(nil) // Call Clone method on the object
		errIf := retVals[1].Interface()
		var notNil bool
		if err, notNil = errIf.(error); notNil {
			return
		}
		return retVals[0].Interface(), nil
	}, time.Duration(2*time.Second), utils.Cache+"GetClone"+key)
}

func CountEntries(prefix string) (result int) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.CountEntriesForPrefix(prefix)
}

func GetEntryKeys(prefix string) (keys []string) {
	cacheMux.RLock()
	defer cacheMux.RUnlock()
	return cache.GetKeysForPrefix(prefix)
}
