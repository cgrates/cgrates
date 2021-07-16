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

// Guardian is the global package variable
var Guardian = &GuardianLocker{
	locks: make(map[string]*itemLock),
	refs:  make(map[string]*refObj)}

type itemLock struct {
	lk  chan struct{} //better with  mutex
	cnt int64
}

type refObj struct {
	refs []string
	tm   *time.Timer
}

// GuardianLocker is an optimized locking system per locking key
type GuardianLocker struct {
	locks   map[string]*itemLock
	lkMux   sync.Mutex         // protects the locks
	refs    map[string]*refObj // used in case of remote locks
	refsMux sync.RWMutex       // protects the map
}

func (gl *GuardianLocker) lockItem(itmID string) {
	if itmID == "" {
		return
	}
	gl.lkMux.Lock()
	itmLock, exists := gl.locks[itmID]
	if !exists {
		gl.locks[itmID] = &itemLock{lk: make(chan struct{}, 1), cnt: 1}
		gl.lkMux.Unlock()
		return
	}
	itmLock.cnt++
	gl.lkMux.Unlock()
	<-itmLock.lk
}

func (gl *GuardianLocker) unlockItem(itmID string) {
	gl.lkMux.Lock()
	itmLock, exists := gl.locks[itmID]
	if !exists {
		gl.lkMux.Unlock()
		return
	}
	itmLock.cnt--
	if itmLock.cnt == 0 {
		delete(gl.locks, itmID)
	}
	gl.lkMux.Unlock()
	itmLock.lk <- struct{}{}
}

// lockWithReference will perform locks and also generate a lock reference for it (so it can be used when remotely locking)
func (gl *GuardianLocker) lockWithReference(refID string, timeout time.Duration, lkIDs ...string) string {
	var refEmpty bool
	if refID == "" {
		refEmpty = true
		refID = utils.GenUUID()
	}
	gl.lockItem(refID) // make sure we only process one simultaneous refID at the time, otherwise checking already used refID is not reliable
	gl.refsMux.Lock()
	if !refEmpty {
		if _, has := gl.refs[refID]; has {
			gl.refsMux.Unlock()
			gl.unlockItem(refID)
			return "" // no locking was done
		}
	}
	var tm *time.Timer
	if timeout != 0 {
		tm = time.AfterFunc(timeout, func() {
			if lkIDs := gl.unlockWithReference(refID); len(lkIDs) != 0 {
				utils.Logger.Warning(fmt.Sprintf("<Guardian> force timing-out locks: %+v", lkIDs))
			}
		})
	}
	gl.refs[refID] = &refObj{
		refs: lkIDs,
		tm:   tm,
	}
	gl.refsMux.Unlock()
	// execute the real locks
	for _, lk := range lkIDs {
		gl.lockItem(lk)
	}
	gl.unlockItem(refID)
	return refID
}

// unlockWithReference will unlock based on the reference ID
func (gl *GuardianLocker) unlockWithReference(refID string) (lkIDs []string) {
	gl.lockItem(refID)
	gl.refsMux.Lock()
	ref, has := gl.refs[refID]
	if !has {
		gl.refsMux.Unlock()
		gl.unlockItem(refID)
		return
	}
	if ref.tm != nil {
		ref.tm.Stop()
	}
	delete(gl.refs, refID)
	gl.refsMux.Unlock()
	lkIDs = ref.refs
	for _, lk := range lkIDs {
		gl.unlockItem(lk)
	}
	gl.unlockItem(refID)
	return
}

// Guard executes the handler between locks
func (gl *GuardianLocker) Guard(handler func() error, timeout time.Duration, lockIDs ...string) (err error) { // do we need the interface here as a reply?
	for _, lockID := range lockIDs {
		gl.lockItem(lockID)
	}
	errChan := make(chan error, 1)
	go func() {
		errChan <- handler()
	}()
	if timeout > 0 { // wait with timeout
		tm := time.NewTimer(timeout)
		select {
		case err = <-errChan:
			close(errChan)
			tm.Stop()
		case <-tm.C:
			utils.Logger.Warning(fmt.Sprintf("<Guardian> force timing-out locks: %+v", lockIDs))
		}
	} else { // a bit dangerous but wait till handler finishes
		err = <-errChan
		close(errChan)
	}
	for _, lockID := range lockIDs {
		gl.unlockItem(lockID)
	}
	return
}

// GuardIDs aquires a lock for duration
// returns the reference ID for the lock group aquired
func (gl *GuardianLocker) GuardIDs(refID string, timeout time.Duration, lkIDs ...string) string {
	return gl.lockWithReference(refID, timeout, lkIDs...)
}

// UnguardIDs attempts to unlock a set of locks based on their reference ID received on lock
func (gl *GuardianLocker) UnguardIDs(refID string) (_ []string) {
	if refID == "" {
		return
	}
	return gl.unlockWithReference(refID)
}
