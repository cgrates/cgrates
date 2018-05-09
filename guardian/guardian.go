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

package guardian

import (
	"sync"
	"sync/atomic"
	"time"
)

// global package variable
var Guardian = &GuardianLock{locksMap: make(map[string]*itemLock)}

func newItemLock(keyID string) *itemLock {
	return &itemLock{keyID: keyID}
}

// itemLock represents one lock with key autodestroy
type itemLock struct {
	keyID string // store it so we know what to destroy
	cnt   int64
	sync.Mutex
}

// lock() executes combined lock with increasing counter
func (il *itemLock) lock() {
	atomic.AddInt64(&il.cnt, 1)
	il.Lock()
}

// unlock() executes combined lock with autoremoving lock from Guardian
func (il *itemLock) unlock() {
	atomic.AddInt64(&il.cnt, -1)
	cnt := atomic.LoadInt64(&il.cnt)
	if cnt < 0 { // already unlocked
		return
	}
	if cnt == 0 { // last lock in the queue
		Guardian.Lock()
		delete(Guardian.locksMap, il.keyID)
		Guardian.Unlock()
	}
	il.Unlock()
}

type itemLocks []*itemLock

func (ils itemLocks) lock() {
	for _, itmLock := range ils {
		itmLock.lock()
	}
}

func (ils itemLocks) unlock() {
	for _, itmLock := range ils {
		itmLock.unlock()
	}
}

// GuardianLock is an optimized locking system per locking key
type GuardianLock struct {
	locksMap     map[string]*itemLock
	sync.RWMutex // protects the maps
}

// lockItems locks a set of lockIDs
// returning the lock structs so they can be later unlocked
func (guard *GuardianLock) lockItems(lockIDs []string) (itmLocks itemLocks) {
	guard.Lock()
	for _, lockID := range lockIDs {
		var itmLock *itemLock
		itmLock, exists := guard.locksMap[lockID]
		if !exists {
			itmLock = newItemLock(lockID)
			guard.locksMap[lockID] = itmLock
		}
		itmLocks = append(itmLocks, itmLock)
	}
	guard.Unlock()

	itmLocks.lock()
	return
}

// Guard executes the handler between locks
func (guard *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, lockIDs ...string) (reply interface{}, err error) {
	itmLocks := guard.lockItems(lockIDs)

	rplyChan := make(chan interface{})
	errChan := make(chan error)
	go func(rplyChan chan interface{}, errChan chan error) {
		// execute
		if rply, err := handler(); err != nil {
			errChan <- err
		} else {
			rplyChan <- rply
		}
	}(rplyChan, errChan)

	if timeout > 0 { // wait with timeout
		select {
		case err = <-errChan:
		case reply = <-rplyChan:
		case <-time.After(timeout):
		}
	} else { // a bit dangerous but wait till handler finishes
		select {
		case err = <-errChan:
		case reply = <-rplyChan:
		}
	}

	itmLocks.unlock()
	return
}

// GuardTimed aquires a lock for duration
func (guard *GuardianLock) GuardIDs(timeout time.Duration, lockIDs ...string) {
	guard.lockItems(lockIDs)
	if timeout != 0 {
		go func(timeout time.Duration, lockIDs ...string) {
			time.Sleep(timeout)
			guard.UnguardIDs(lockIDs...)
		}(timeout, lockIDs...)
	}
	return
}

// UnguardTimed attempts to unlock a set of locks based on their locksUUID
func (guard *GuardianLock) UnguardIDs(lockIDs ...string) {
	var itmLocks itemLocks
	guard.RLock()
	for _, lockID := range lockIDs {
		var itmLock *itemLock
		itmLock, exists := Guardian.locksMap[lockID]
		if exists {
			itmLocks = append(itmLocks, itmLock)
		}
	}
	guard.RUnlock()
	itmLocks.unlock()
	return
}
