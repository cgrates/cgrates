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
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func delayHandler2() error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Forks 3 groups of workers and makes sure that the time for execution is the one we expect for all 15 goroutines (with 100ms )
func TestGuardian2MultipleKeys(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian2.Guard(delayHandler2, 0, key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*100) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur ||
		execTime > mustExecDur+100*time.Millisecond {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian2.lkMux.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian2.locks[key]; hasKey {
			t.Errorf("Possible memleak for key: %s", key)
		}
	}
	Guardian2.lkMux.Unlock()
}

func TestGuardian2Timeout(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian2.Guard(delayHandler2, 10*time.Millisecond, key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*10) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur ||
		execTime > mustExecDur+100*time.Millisecond {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian2.lkMux.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian2.locks[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
	Guardian2.lkMux.Unlock()
}

func TestGuardian2GuardIDs(t *testing.T) {

	//lock with 3 keys
	lockIDs := []string{"test1", "test2", "test3"}
	// make sure the keys are not in guardian before lock
	Guardian2.lkMux.Lock()
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian2.locks[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	Guardian2.lkMux.Unlock()
	// lock 3 items
	tStart := time.Now()
	lockDur := 2 * time.Millisecond
	Guardian2.GuardIDs("", lockDur, lockIDs...)
	Guardian2.lkMux.Lock()
	for _, lockID := range lockIDs {
		if itmLock, hasKey := Guardian2.locks[lockID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lockID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected itmLock found: %+v", itmLock)
		}
	}
	Guardian2.lkMux.Unlock()
	secLockDur := time.Millisecond
	// second lock to test counter
	go Guardian2.GuardIDs("", secLockDur, lockIDs[1:]...)
	time.Sleep(30 * time.Microsecond) // give time for goroutine to lock
	// check if counters were properly increased
	Guardian2.lkMux.Lock()
	lkID := lockIDs[0]
	eCnt := int64(1)
	if itmLock, hasKey := Guardian2.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[1]
	eCnt = int64(2)
	if itmLock, hasKey := Guardian2.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[2]
	eCnt = int64(1) // we did not manage to increase it yet since it did not pass first lock
	if itmLock, hasKey := Guardian2.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	Guardian2.lkMux.Unlock()
	time.Sleep(lockDur + secLockDur + 50*time.Millisecond) // give time to unlock before proceeding

	// make sure all counters were removed
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian2.locks[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	// test lock  without timer
	refID := Guardian2.GuardIDs("", 0, lockIDs...)

	if totalLockDur := time.Now().Sub(tStart); totalLockDur < lockDur {
		t.Errorf("Lock duration too small")
	}
	time.Sleep(30 * time.Millisecond)
	// making sure the items stay locked
	Guardian2.lkMux.Lock()
	if len(Guardian2.locks) != 3 {
		t.Errorf("locks should have 3 elements, have: %+v", Guardian2.locks)
	}
	for _, lkID := range lockIDs {
		if itmLock, hasKey := Guardian2.locks[lkID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lkID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
		}
	}
	Guardian2.lkMux.Unlock()
	Guardian2.UnguardIDs(refID)
	// make sure items were unlocked
	Guardian2.lkMux.Lock()
	if len(Guardian2.locks) != 0 {
		t.Errorf("locks should have 0 elements, has: %+v", Guardian2.locks)
	}
	Guardian2.lkMux.Unlock()
}

// TestGuardian2GuardIDsConcurrent executes GuardIDs concurrently
func TestGuardian2GuardIDsConcurrent(t *testing.T) {
	maxIter := 500
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	refID := utils.GenUUID()
	for i := 0; i < maxIter; i++ {
		sg.Add(1)
		go func() {
			if retRefID := Guardian2.GuardIDs(refID, 0, keys...); retRefID != "" {
				if lkIDs := Guardian2.UnguardIDs(refID); !reflect.DeepEqual(keys, lkIDs) {
					t.Errorf("expecting: %+v, received: %+v", keys, lkIDs)
				}
			}
			sg.Done()
		}()
	}
	sg.Wait()

	Guardian2.lkMux.Lock()
	if len(Guardian2.locks) != 0 {
		t.Errorf("Possible memleak for locks: %+v", Guardian2.locks)
	}
	Guardian2.lkMux.Unlock()
	Guardian2.refsMux.Lock()
	if len(Guardian2.refs) != 0 {
		t.Errorf("Possible memleak for refs: %+v", Guardian2.refs)
	}
	Guardian2.refsMux.Unlock()
}

func TestGuardian2GuardIDsTimeoutConcurrent(t *testing.T) {
	maxIter := 50
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	refID := utils.GenUUID()
	for i := 0; i < maxIter; i++ {
		sg.Add(1)
		go func() {
			Guardian2.GuardIDs(refID, time.Microsecond, keys...)
			sg.Done()
		}()
	}
	sg.Wait()
	time.Sleep(10 * time.Millisecond)
	Guardian2.lkMux.Lock()
	if len(Guardian2.locks) != 0 {
		t.Errorf("Possible memleak for locks: %+v", Guardian2.locks)
	}
	Guardian2.lkMux.Unlock()
	Guardian2.refsMux.Lock()
	if len(Guardian2.refs) != 0 {
		t.Errorf("Possible memleak for refs: %+v", Guardian2.refs)
	}
	Guardian2.refsMux.Unlock()
}

// BenchmarkGuard-8      	  200000	     13759 ns/op
func BenchmarkGuard2(b *testing.B) {
	wg := new(sync.WaitGroup)
	wg.Add(b.N * 3)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		go func() {
			Guardian2.Guard(func() error {
				time.Sleep(time.Microsecond)
				return nil
			}, 0, "1")
			wg.Done()
		}()
		go func() {
			Guardian2.Guard(func() error {
				time.Sleep(time.Microsecond)
				return nil
			}, 0, "2")
			wg.Done()
		}()
		go func() {
			Guardian2.Guard(func() error {
				time.Sleep(time.Microsecond)
				return nil
			}, 0, "1")
			wg.Done()
		}()
	}
	wg.Wait()
}

// BenchmarkGuardian-8   	 1000000	      5794 ns/op
func BenchmarkGuardian2(b *testing.B) {
	wg := new(sync.WaitGroup)
	wg.Add(b.N)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		go func(n int) {
			Guardian2.Guard(func() error {
				time.Sleep(time.Microsecond)
				return nil
			}, 0, strconv.Itoa(n))
			wg.Done()
		}(n)
	}
	wg.Wait()
}

// BenchmarkGuardIDs-8   	 1000000	      8732 ns/op
func BenchmarkGuardIDs2(b *testing.B) {
	wg := new(sync.WaitGroup)
	wg.Add(b.N)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		go func(i int) {
			if refID := Guardian2.GuardIDs("", 0, strconv.Itoa(i)); refID != "" {
				time.Sleep(time.Microsecond)
				Guardian2.UnguardIDs(refID)
			}
			wg.Done()
		}(n)
	}
	wg.Wait()
}

func TestGuardian2LockItemUnlockItem(t *testing.T) {
	//for coverage purposes
	itemID := utils.EmptyString
	Guardian2.lockItem(itemID)
	Guardian2.unlockItem(itemID)
	if itemID != utils.EmptyString {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.EmptyString, itemID)
	}
}

func TestGuardian2LockUnlockWithReference(t *testing.T) {
	//for coverage purposes
	refID := utils.EmptyString
	Guardian2.lockWithReference(refID, 0)
	Guardian2.unlockWithReference(refID)
	if refID != utils.EmptyString {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.EmptyString, refID)
	}
}

func TestGuardian2GuardUnguardIDs(t *testing.T) {
	//for coverage purposes
	refID := utils.EmptyString
	lkIDs := []string{"test1", "test2", "test3"}
	Guardian2.GuardIDs(refID, time.Second, lkIDs...)
	Guardian2.UnguardIDs(refID)
	if refID != utils.EmptyString {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.EmptyString, refID)
	}
}

func TestGuardian2GuardUnguardIDsCase2(t *testing.T) {
	//for coverage purposes
	lkIDs := []string{"test1", "test2", "test3"}
	err := Guardian2.Guard(func() error {
		return utils.ErrNotFound
	}, 10*time.Millisecond, lkIDs...)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}
