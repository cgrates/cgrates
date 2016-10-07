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
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur || execTime > mustExecDur+time.Duration(10*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
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
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur || execTime > mustExecDur+time.Duration(10*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	for _, key := range keys {
		if _, hasKey := Guardian.locksMap[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
}

func TestGuardianGuardIDs(t *testing.T) {
	lockIDs := []string{"test1", "test2", "test3"}
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locksMap[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	tStart := time.Now()
	lockDur := 2 * time.Millisecond
	Guardian.GuardIDs(lockDur, lockIDs...)
	for _, lockID := range lockIDs {
		if itmLock, hasKey := Guardian.locksMap[lockID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lockID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected itmLock found: %+v", itmLock)
		}
	}
	go Guardian.GuardIDs(time.Duration(1*time.Millisecond), lockIDs[1:]...) // to test counter
	time.Sleep(20 * time.Microsecond)                                       // give time for goroutine to lock
	if itmLock, hasKey := Guardian.locksMap["test1"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test1")
	} else if itmLock.cnt != 1 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	}
	if itmLock, hasKey := Guardian.locksMap["test2"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test2")
	} else if itmLock.cnt != 2 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	}
	if itmLock, hasKey := Guardian.locksMap["test3"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test3")
	} else if itmLock.cnt != 2 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	}
	Guardian.GuardIDs(0, lockIDs...)
	if totalLockDur := time.Now().Sub(tStart); totalLockDur < lockDur {
		t.Errorf("Lock duration too small")
	}
	//time.Sleep(1000 * time.Microsecond)
	if len(Guardian.locksMap) != 3 {
		t.Errorf("locksMap should be have 3 elements, have: %+v", Guardian.locksMap)
	} else if itmLock, hasKey := Guardian.locksMap["test1"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test1")
	} else if itmLock.cnt != 1 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	} else if itmLock, hasKey := Guardian.locksMap["test2"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test2")
	} else if itmLock.cnt != 1 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	} else if itmLock, hasKey := Guardian.locksMap["test3"]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", "test2")
	} else if itmLock.cnt != 1 {
		t.Errorf("Unexpected itmLock found: %+v", itmLock)
	}
	Guardian.UnguardIDs(lockIDs...)
	if len(Guardian.locksMap) != 0 {
		t.Errorf("locksMap should be have 0 elements, have: %+v", Guardian.locksMap)
	}
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
