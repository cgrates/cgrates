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

	"github.com/cgrates/cgrates/servmanager"
)

// NewServiceIndexer constructs a ServiceIndexer
func NewServiceIndexer() *ServiceIndexer {
	return &ServiceIndexer{srvS: make(map[string]servmanager.Service)}
}

// ServiceIndexer implements servmanager.Service indexing in a thread safe way
type ServiceIndexer struct {
	mux sync.RWMutex

	srvS map[string]servmanager.Service // servmanager.Services indexed by ID
}

// Getservmanager.Service returns one servmanager.Service or nil
func (sI *ServiceIndexer) GetService(srvID string) servmanager.Service {
	sI.mux.RLock()
	defer sI.mux.RUnlock()
	return sI.srvS[srvID]
}

// Addservmanager.Service adds a servmanager.Service based on it's id to the index
func (sI *ServiceIndexer) AddService(srvID string, srv servmanager.Service) {
	sI.mux.Lock()
	sI.srvS[srvID] = srv
	sI.mux.Unlock()
}

// Remservmanager.Service will remove a servmanager.Service based on it's ID
func (sI *ServiceIndexer) RemService(srvID string) {
	sI.mux.Lock()
	defer sI.mux.Unlock()
	delete(sI.srvS, srvID)
}

// Getservmanager.Services returns the list of servmanager.Services indexed
func (sI *ServiceIndexer) GetServices() []servmanager.Service {
	sI.mux.RLock()
	defer sI.mux.RUnlock()
	srvs := make([]servmanager.Service, 0, len(sI.srvS))
	for _, s := range sI.srvS {
		srvs = append(srvs, s)
	}
	return srvs
}
