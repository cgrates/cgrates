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
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/servmanager"
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

// waitForServicesToReachState ensures each service reaches the desired state, with the timeout applied individually per service.
// Returns a map of service names to their instances or an error if any service fails to reach its state within its timeout window.
func waitForServicesToReachState(state string, serviceIDs []string, indexer *servmanager.ServiceRegistry, timeout time.Duration,
) (map[string]servmanager.Service, error) {
	services := make(map[string]servmanager.Service, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		srv, err := waitForServiceState(state, serviceID, indexer, timeout)
		if err != nil {
			return nil, err
		}
		services[srv.ServiceName()] = srv

	}
	return services, nil
}

// waitForServiceState waits up to timeout duration for a service to reach the specified state.
// Returns the service instance or an error if the timeout is exceeded.
func waitForServiceState(state, serviceID string, indexer *servmanager.ServiceRegistry, timeout time.Duration,
) (servmanager.Service, error) {
	srv := indexer.Lookup(serviceID)
	select {
	case <-srv.StateChan(state):
		return srv, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timed out waiting for service %q state %q", serviceID, state)
	}
}
