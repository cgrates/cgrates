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

package servmanager

import (
	"sync"
)

// NewServiceIndexer constructs a ServiceIndexer
func NewServiceIndexer() *ServiceIndexer {
	return &ServiceIndexer{srvS: make(map[string]Service)}
}

// ServiceIndexer implements service indexing in a thread safe way
type ServiceIndexer struct {
	mux sync.RWMutex

	srvS map[string]Service // services indexed by ID
}

// GetService returns one service or nil
func (sI *ServiceIndexer) GetService(srvID string) Service {
	sI.mux.RLock()
	defer sI.mux.RUnlock()
	return sI.srvS[srvID]
}

// AddService adds a service based on it's id to the index
func (sI *ServiceIndexer) AddService(srvID string, srv Service) {
	sI.mux.Lock()
	sI.srvS[srvID] = srv
	sI.mux.Unlock()
}

// RemoveService will remove a service based on it's ID
func (sI *ServiceIndexer) RemoveService(srvID string) {
	sI.mux.Lock()
	defer sI.mux.Unlock()
	delete(sI.srvS, srvID)
}

// GetServices returns the list of services indexed
func (sI *ServiceIndexer) GetServices() []Service {
	sI.mux.RLock()
	defer sI.mux.RUnlock()
	srvs := make([]Service, 0, len(sI.srvS))
	for _, s := range sI.srvS {
		srvs = append(srvs, s)
	}
	return srvs
}
