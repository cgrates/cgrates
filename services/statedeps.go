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

package services

import (
	"sync"
)

// NewStateDependencies constructs a StateDependencies struct
func NewStateDependencies(servStates []string) (stDeps *StateDependencies) {
	stDeps = &StateDependencies{stateDeps: make(map[string]chan struct{})}
	for _, stateID := range servStates {
		stDeps.stateDeps[stateID] = make(chan struct{})
	}
	return
}

// StateDependencies enhances a service with state dependencies management
type StateDependencies struct {
	stateDeps    map[string]chan struct{} // listeners for various states of the service
	stateDepsMux sync.RWMutex             // protects stateDeps
}

// RegisterStateDependency will be called by a service interested by specific stateID of the service
func (sDs *StateDependencies) StateChan(stateID string) (retChan chan struct{}) {
	sDs.stateDepsMux.RLock()
	retChan = sDs.stateDeps[stateID]
	sDs.stateDepsMux.RUnlock()
	return
}
