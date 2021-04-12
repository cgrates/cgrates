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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// Guardian is the global package variable
var Guardian = &GuardianLocker{
	locks: make(map[string]*itemLock),
	refs:  make(map[string][]string)}

type itemLock struct {
	lk  chan struct{}
	cnt int64
}

// GuardianLocker is an optimized locking system per locking key
type GuardianLocker struct {
	locks   map[string]*itemLock
	lkMux   sync.Mutex          // protects the locks
	refs    map[string][]string // used in case of remote locks
	refsMux sync.RWMutex        // protects the map
}

func (gl *GuardianLocker) lockItem(itmID string) {
	if itmID == "" {
		return
	}
	gl.lkMux.Lock()
	itmLock, exists := gl.locks[itmID]
	if !exists {
		itmLock = &itemLock{lk: make(chan struct{}, 1)}
		gl.locks[itmID] = itmLock
		itmLock.lk <- struct{}{}
	}
	itmLock.cnt++
	select {
	case <-itmLock.lk:
		gl.lkMux.Unlock()
		return
	default: // move further so we can unlock
	}
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
func (gl *GuardianLocker) lockWithReference(refID string, lkIDs []string) string {
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
	gl.refs[refID] = lkIDs
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
	lkIDs, has := gl.refs[refID]
	if !has {
		gl.refsMux.Unlock()
		gl.unlockItem(refID)
		return
	}
	delete(gl.refs, refID)
	gl.refsMux.Unlock()
	for _, lk := range lkIDs {
		gl.unlockItem(lk)
	}
	gl.unlockItem(refID)
	return
}

// Guard executes the handler between locks
func (gl *GuardianLocker) Guard(ctx *context.Context, handler func(*context.Context) (interface{}, error), timeout time.Duration, lockIDs ...string) (reply interface{}, err error) {
	for _, lockID := range lockIDs {
		gl.lockItem(lockID)
	}
	rplyChan := make(chan interface{})
	errChan := make(chan error)
	if timeout > 0 { // wait with timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel() // stop the context at the end of function
	}
	go func(rplyChan chan interface{}, errChan chan error) {
		// execute
		if rply, err := handler(ctx); err != nil {
			errChan <- err
		} else {
			rplyChan <- rply
		}
	}(rplyChan, errChan)

	select {
	case err = <-errChan:
	case reply = <-rplyChan:
	case <-ctx.Done(): // ignore context error but log it
		utils.Logger.Warning(fmt.Sprintf("<Guardian> force timing-out locks: <%+v> because: <%s> ", lockIDs, ctx.Err()))
	}
	for _, lockID := range lockIDs {
		gl.unlockItem(lockID)
	}
	return
}

// GuardIDs aquires a lock for duration
// returns the reference ID for the lock group aquired
func (gl *GuardianLocker) GuardIDs(refID string, timeout time.Duration, lkIDs ...string) (retRefID string) {
	retRefID = gl.lockWithReference(refID, lkIDs)
	if timeout != 0 && retRefID != "" {
		go func() {
			time.Sleep(timeout)
			lkIDs := gl.unlockWithReference(retRefID)
			if len(lkIDs) != 0 {
				utils.Logger.Warning(fmt.Sprintf("<Guardian> force timing-out locks: %+v", lkIDs))
			}
		}()
	}
	return
}

// UnguardIDs attempts to unlock a set of locks based on their reference ID received on lock
func (gl *GuardianLocker) UnguardIDs(refID string) (lkIDs []string) {
	if refID == "" {
		return
	}
	return gl.unlockWithReference(refID)
}
