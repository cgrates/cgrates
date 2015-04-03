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
)

var AccLock *AccountLock

func init() {
	AccLock = NewAccountLock()
}

type AccountLock struct {
	queue map[string]chan bool
	mu    sync.Mutex
}

func NewAccountLock() *AccountLock {
	return &AccountLock{queue: make(map[string]chan bool)}
}

func (cm *AccountLock) GuardCallCost(handler func() (*CallCost, error), name string) (reply *CallCost, err error) {
	cm.mu.Lock()
	lock, exists := AccLock.queue[name]
	if !exists {
		lock = make(chan bool, 1)
		AccLock.queue[name] = lock
	}
	lock <- true
	cm.mu.Unlock()
	reply, err = handler()
	<-lock
	return
}

func (cm *AccountLock) Guard(handler func() (float64, error), names ...string) (reply float64, err error) {
	cm.mu.Lock()
	for _, name := range names {
		lock, exists := AccLock.queue[name]
		if !exists {
			lock = make(chan bool, 1)
			AccLock.queue[name] = lock
		}
		lock <- true
	}
	cm.mu.Unlock()
	reply, err = handler()
	for _, name := range names {
		lock := AccLock.queue[name]
		<-lock
	}
	return
}
