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
	"sync"
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
	cnt   int
	sync.Mutex
}

// unlock() executes combined lock with autoremoving lock from Guardian
func (il *itemLock) unlock() {
	il.cnt--
	if il.cnt == 0 { // last lock in the queue
		Guardian.Lock()
		delete(Guardian.locksMap, il.keyID)
		Guardian.Unlock()
	}
	il.Unlock()
}

// GuardianLock is an optimized locking system per locking key
type GuardianLock struct {
	locksMap   map[string]*itemLock
	sync.Mutex // protects the maps
}

// lockItems locks a set of lockIDs
// returning the lock structs so they can be later unlocked
func (guard *GuardianLock) lockItems(lockIDs []string) (itmLocks []*itemLock) {
	guard.Lock()
	for _, lockID := range lockIDs {
		var itmLock *itemLock
		itmLock, exists := Guardian.locksMap[lockID]
		if !exists {
			itmLock = newItemLock(lockID)
			Guardian.locksMap[lockID] = itmLock
		}
		itmLock.cnt++
		itmLocks = append(itmLocks, itmLock)
	}
	guard.Unlock()
	for _, itmLock := range itmLocks {
		itmLock.Lock()
	}
	return
}

// unlockItems will unlock the items provided
func (guard *GuardianLock) unlockItems(itmLocks []*itemLock) {
	for _, itmLock := range itmLocks {
		itmLock.unlock()
	}
}

func (guard *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, lockIDs ...string) (reply interface{}, err error) {
	itmLocks := guard.lockItems(lockIDs)
	defer guard.unlockItems(itmLocks)

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

	return
}

// GuardTimed aquires a lock for duration, returning an identifier which can be used later to cancel the lock or find it's status
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
// Returns false if locks were not found
func (guard *GuardianLock) UnguardIDs(lockIDs ...string) {
	var itmLocks []*itemLock
	guard.Lock()
	for _, lockID := range lockIDs {
		var itmLock *itemLock
		itmLock, exists := Guardian.locksMap[lockID]
		if exists {
			itmLocks = append(itmLocks, itmLock)
		}
	}
	guard.Unlock()
	guard.unlockItems(itmLocks)
	return
}
