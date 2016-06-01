/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
var Guardian = &GuardianLock{locksMap: make(map[string]chan bool)}

type GuardianLock struct {
	locksMap map[string]chan bool
	mu       sync.RWMutex
}

func (cm *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var locks []chan bool // take existing locks out of the mutex
	cm.mu.Lock()
	for _, name := range names {
		if lock, exists := Guardian.locksMap[name]; !exists {
			lock = make(chan bool, 1)
			Guardian.locksMap[name] = lock
			lock <- true
		} else {
			locks = append(locks, lock)
		}
	}
	cm.mu.Unlock()

	for _, lock := range locks {
		lock <- true
	}

	funcWaiter := make(chan bool)
	go func() {
		// execute
		reply, err = handler()
		funcWaiter <- true
	}()
	// wait with timeout
	if timeout > 0 {
		select {
		case <-funcWaiter:
		case <-time.After(timeout):
		}
	} else {
		<-funcWaiter
	}
	// release
	cm.mu.RLock()
	for _, name := range names {
		<-Guardian.locksMap[name]
	}
	cm.mu.RUnlock()
	return
}
