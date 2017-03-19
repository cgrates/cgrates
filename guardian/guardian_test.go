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
	"testing"
	"time"
)

func delayHandler() (interface{}, error) {
	time.Sleep(100 * time.Millisecond)
	return nil, nil
}

// Forks 3 groups of workers and makes sure that the time for execution is the one we expect for all 15 goroutines (with 100ms )
func TestGuardianMultipleKeys(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian.Guard(delayHandler, 0, key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*100) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur || execTime > mustExecDur+time.Duration(20*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian.RLock()
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
	Guardian.RUnlock()
}

func TestGuardianTimeout(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian.Guard(delayHandler, time.Duration(10*time.Millisecond), key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*10) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur || execTime > mustExecDur+time.Duration(20*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian.RLock()
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
	Guardian.RUnlock()
}

func TestGuardianGuardIDs(t *testing.T) {
	lockIDs := []string{"test1", "test2", "test3"}
	Guardian.RLock()
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locksMap[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	Guardian.RUnlock()
	tStart := time.Now()
	lockDur := 2 * time.Millisecond
	Guardian.GuardIDs(lockDur, lockIDs...)
	Guardian.RLock()
	for _, lockID := range lockIDs {
		if itmLock, hasKey := Guardian.locksMap[lockID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lockID)
		} else if atomic.LoadInt64(&itmLock.cnt) != 1 {
			t.Errorf("Unexpected itmLock found: %+v", itmLock)
		}
	}
	Guardian.RUnlock()
	go Guardian.GuardIDs(time.Duration(1*time.Millisecond), lockIDs[1:]...) // to test counter
	time.Sleep(20 * time.Microsecond)                                       // give time for goroutine to lock
	Guardian.RLock()
	lkID := lockIDs[0]
	eCnt := int64(1)
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if cnt := atomic.LoadInt64(&itmLock.cnt); cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", cnt, lkID)
	}
	lkID = lockIDs[1]
	eCnt = int64(2)
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if cnt := atomic.LoadInt64(&itmLock.cnt); cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", cnt, lkID)
	}
	lkID = lockIDs[2]
	eCnt = int64(2)
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if cnt := atomic.LoadInt64(&itmLock.cnt); cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", cnt, lkID)
	}
	Guardian.RUnlock()
	Guardian.GuardIDs(0, lockIDs...)
	if totalLockDur := time.Now().Sub(tStart); totalLockDur < lockDur {
		t.Errorf("Lock duration too small")
	}
	time.Sleep(time.Duration(30) * time.Millisecond)
	Guardian.RLock()
	if len(Guardian.locksMap) != 3 {
		t.Errorf("locksMap should be have 3 elements, have: %+v", Guardian.locksMap)
	}
	for _, lkID := range lockIDs {
		if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lkID)
		} else if cnt := atomic.LoadInt64(&itmLock.cnt); cnt != 1 {
			t.Errorf("Unexpected counter: %d for itmLock with id %s", cnt, lkID)
		}
	}
	Guardian.RUnlock()
	Guardian.UnguardIDs(lockIDs...)
	time.Sleep(time.Duration(50) * time.Millisecond)
	Guardian.RLock()
	if len(Guardian.locksMap) != 0 {
		t.Errorf("locksMap should have 0 elements, has: %+v", Guardian.locksMap)
	}
	Guardian.RUnlock()
}

func BenchmarkGuard(b *testing.B) {
	for i := 0; i < 100; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "2")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}

}

func BenchmarkGuardian(b *testing.B) {
	for i := 0; i < 100; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}
}
