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
	Guardian.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Errorf("Possible memleak for key: %s", key)
		}
	}
	Guardian.Unlock()
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
	Guardian.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
	Guardian.Unlock()
}

func TestGuardianGuardIDs(t *testing.T) {

	//lock with 3 keys
	lockIDs := []string{"test1", "test2", "test3"}
	// make sure the keys are not in guardian before lock
	Guardian.Lock()
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locksMap[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	Guardian.Unlock()

	// lock 3 items
	tStart := time.Now()
	lockDur := 2 * time.Millisecond
	Guardian.GuardIDs(lockDur, lockIDs...)
	Guardian.Lock()
	for _, lockID := range lockIDs {
		if itmLock, hasKey := Guardian.locksMap[lockID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lockID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected itmLock found: %+v", itmLock)
		}
	}
	Guardian.Unlock()
	secLockDur := time.Duration(1 * time.Millisecond)

	// second lock to test counter
	go Guardian.GuardIDs(secLockDur, lockIDs[1:]...)
	time.Sleep(20 * time.Microsecond) // give time for goroutine to lock

	// check if counters were properly increased
	Guardian.Lock()
	lkID := lockIDs[0]
	eCnt := int64(1)
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[1]
	eCnt = int64(2)
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[2]
	eCnt = int64(1) // we did not manage to increase it yet since it did not pass first lock
	if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	Guardian.Unlock()

	time.Sleep(lockDur + secLockDur + 10*time.Millisecond) // give time to unlock before proceeding

	// make sure all counters were removed
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locksMap[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}

	// test lock  without timer
	Guardian.GuardIDs(0, lockIDs...)
	if totalLockDur := time.Now().Sub(tStart); totalLockDur < lockDur {
		t.Errorf("Lock duration too small")
	}
	time.Sleep(time.Duration(30) * time.Millisecond)

	// making sure the items stay locked
	Guardian.Lock()
	if len(Guardian.locksMap) != 3 {
		t.Errorf("locksMap should be have 3 elements, have: %+v", Guardian.locksMap)
	}
	for _, lkID := range lockIDs {
		if itmLock, hasKey := Guardian.locksMap[lkID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lkID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
		}
	}
	Guardian.Unlock()

	Guardian.UnguardIDs(lockIDs...)
	time.Sleep(time.Duration(50) * time.Millisecond)

	// make sure items were unlocked
	Guardian.Lock()
	if len(Guardian.locksMap) != 0 {
		t.Errorf("locksMap should have 0 elements, has: %+v", Guardian.locksMap)
	}
	Guardian.Unlock()
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
