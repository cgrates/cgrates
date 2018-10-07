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
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// global package variable
var Guardian = &GuardianLocker{locksMap: make(map[string]*itemLock)}

type itemLock struct {
	lk  chan struct{}
	cnt int64
}

// GuardianLocker is an optimized locking system per locking key
type GuardianLocker struct {
	locksMap   map[string]*itemLock
	sync.Mutex // protects the map
}

func (gl *GuardianLocker) lockItem(itmID string) {
	gl.Lock()
	itmLock, exists := gl.locksMap[itmID]
	if !exists {
		itmLock = &itemLock{lk: make(chan struct{}, 1)}
		gl.locksMap[itmID] = itmLock
		itmLock.lk <- struct{}{}
	}
	itmLock.cnt++
	select {
	case <-itmLock.lk:
		gl.Unlock()
		return
	default: // move further so we can unlock
	}
	gl.Unlock()
	<-itmLock.lk
}

func (gl *GuardianLocker) unlockItem(itmID string) {
	gl.Lock()
	itmLock, exists := gl.locksMap[itmID]
	if !exists {
		gl.Unlock()
		return
	}
	itmLock.cnt--
	if itmLock.cnt == 0 {
		delete(gl.locksMap, itmID)
	}
	gl.Unlock()
	itmLock.lk <- struct{}{}
}

// Guard executes the handler between locks
func (gl *GuardianLocker) Guard(handler func() (interface{}, error), timeout time.Duration, lockIDs ...string) (reply interface{}, err error) {
	for _, lockID := range lockIDs {
		gl.lockItem(lockID)
	}
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
	for _, lockID := range lockIDs {
		gl.unlockItem(lockID)
	}
	return
}

// GuardTimed aquires a lock for duration
func (gl *GuardianLocker) GuardIDs(timeout time.Duration, lockIDs ...string) {
	for _, lockID := range lockIDs {
		gl.lockItem(lockID)
	}
	if timeout != 0 {
		go func(timeout time.Duration, lockIDs ...string) {
			time.Sleep(timeout)
			utils.Logger.Warning(fmt.Sprintf("<Guardian> WARNING: force timing-out locks: %+v", lockIDs))
			gl.UnguardIDs(lockIDs...)
		}(timeout, lockIDs...)
	}
	return
}

// UnguardTimed attempts to unlock a set of locks based on their locksUUID
func (gl *GuardianLocker) UnguardIDs(lockIDs ...string) {
	for _, lockID := range lockIDs {
		gl.unlockItem(lockID)
	}
	return
}
