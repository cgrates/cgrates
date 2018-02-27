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

package utils

import (
	"sync"
	"time"
)

type ResponseCacheItem struct {
	Value interface{}
	Err   error
}

type ResponseCache struct {
	ttl       time.Duration
	cache     map[string]*ResponseCacheItem
	semaphore map[string]chan bool // used for waiting till the first goroutine processes the response
	mu        sync.RWMutex
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		ttl:       ttl,
		cache:     make(map[string]*ResponseCacheItem),
		semaphore: make(map[string]chan bool),
		mu:        sync.RWMutex{},
	}
}

func (rc *ResponseCache) Cache(key string, item *ResponseCacheItem) {
	if rc.ttl == 0 {
		return
	}
	rc.mu.Lock()
	rc.cache[key] = item
	if _, found := rc.semaphore[key]; found {
		close(rc.semaphore[key])  // send release signal
		delete(rc.semaphore, key) // delete key
	}
	rc.mu.Unlock()
	go func() {
		time.Sleep(rc.ttl)
		rc.mu.Lock()
		delete(rc.cache, key)
		rc.mu.Unlock()
	}()
}

func (rc *ResponseCache) Get(key string) (*ResponseCacheItem, error) {
	if rc.ttl == 0 {
		return nil, ErrNotImplemented
	}
	rc.mu.RLock()
	item, ok := rc.cache[key]
	rc.mu.RUnlock()
	if ok {
		return item, nil
	}
	rc.wait(key) // wait for other goroutine processsing this key
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	item, ok = rc.cache[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item, nil
}

func (rc *ResponseCache) wait(key string) {
	rc.mu.RLock()
	lockChan, found := rc.semaphore[key]
	rc.mu.RUnlock()
	if found {
		<-lockChan
	} else {
		rc.mu.Lock()
		rc.semaphore[key] = make(chan bool)
		rc.mu.Unlock()
	}
}
