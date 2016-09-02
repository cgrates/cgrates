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
	return &itemLock{keyID: keyID, lk: new(sync.Mutex)}
}

// itemLock represents one lock with key autodestroy
type itemLock struct {
	keyID string // store it so we know what to destroy
	lk    *sync.Mutex
	cnt   int
}

// lock() keeps also record of running jobs on same item
func (il *itemLock) lock() {
	il.lk.Lock()
	il.cnt++
}

// unlock() executes combined lock with autoremoving lock from Guardian
func (il *itemLock) unlock() {
	il.cnt--
	if il.cnt == 0 {
		Guardian.Lock()
		delete(Guardian.locksMap, il.keyID)
		Guardian.Unlock()
	}
	il.lk.Unlock()
}

// GuardianLock is an optimized locking system per locking key
type GuardianLock struct {
	locksMap map[string]*itemLock
	sync.Mutex
}

func (guard *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var itmLocks []*itemLock // will need to lock all of them before proceeding with our task
	guard.Lock()
	for _, name := range names {
		var itmLock *itemLock
		itmLock, exists := Guardian.locksMap[name]
		if !exists {
			itmLock = newItemLock(name)
			Guardian.locksMap[name] = itmLock
		}
		itmLocks = append(itmLocks, itmLock)
	}
	guard.Unlock()

	for _, itmLock := range itmLocks {
		itmLock.lock()
	}

	handlerDone := make(chan struct{})
	go func(chan struct{}) {
		// execute
		reply, err = handler()
		handlerDone <- struct{}{}
	}(handlerDone)

	if timeout > 0 { // wait with timeout
		select {
		case <-handlerDone:
		case <-time.After(timeout):
		}
	} else { // a bit dangerous but wait till handler finishes
		<-handlerDone
	}
	// release
	for _, itmLock := range itmLocks {
		itmLock.unlock()
	}
	return
}
