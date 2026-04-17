/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package servmanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// serviceEntry holds the Service together with its current state and locks.
type serviceEntry struct {
	svc Service

	stateMu sync.Mutex    // guards state and change
	state   string        // UP or DOWN
	change  chan struct{} // closed on state change to wake waiters

	lifecycleMu sync.Mutex // held for the full Start/Reload/Shutdown call
}

// Registry tracks registered services and their state.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*serviceEntry
}

// NewRegistry returns an initialized registry.
func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]*serviceEntry),
	}
}

// Register adds services to the registry. Duplicates are ignored.
func (r *Registry) Register(svcs ...Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, svc := range svcs {
		id := svc.ServiceName()
		if _, ok := r.entries[id]; ok {
			continue
		}
		r.entries[id] = &serviceEntry{
			svc:    svc,
			state:  utils.StateServiceDOWN,
			change: make(chan struct{}),
		}
	}
}

// Lookup returns the Service for id, or nil if not found.
func (r *Registry) Lookup(id string) Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	if !ok {
		return nil
	}
	return e.svc
}

// List returns all registered services.
func (r *Registry) List() []Service {
	r.mu.RLock()
	defer r.mu.RUnlock()
	svcs := make([]Service, 0, len(r.entries))
	for _, e := range r.entries {
		svcs = append(svcs, e.svc)
	}
	return svcs
}

// entry returns the serviceEntry for id, or nil if unknown.
func (r *Registry) entry(id string) *serviceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.entries[id]
}

// State returns the current state of a service, or "" if unknown.
func (r *Registry) State(id string) string {
	e := r.entry(id)
	if e == nil {
		return ""
	}
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	return e.state
}

// SetState updates state and unblocks anyone waiting on it.
func (r *Registry) SetState(id, state string) error {
	if state != utils.StateServiceUP && state != utils.StateServiceDOWN {
		return fmt.Errorf("invalid service state: %q", state)
	}
	e := r.entry(id)
	if e == nil {
		return fmt.Errorf("unknown service %q", id)
	}
	e.stateMu.Lock()
	defer e.stateMu.Unlock()
	if e.state == state {
		return nil
	}
	e.state = state
	close(e.change) // wake every waiter at once
	e.change = make(chan struct{})
	return nil
}

// waitForState blocks until the service reaches target, shutdown gets
// triggered, or it times out. Pass nil shutdown from Shutdown methods where
// shutdown has already been triggered.
func (r *Registry) waitForState(shutdown *utils.SyncedChan, id, target string,
	timeout time.Duration) error {
	e := r.entry(id)
	if e == nil {
		return fmt.Errorf("unknown service %q", id)
	}
	var shutdownCh <-chan struct{}
	if shutdown != nil {
		shutdownCh = shutdown.Done()
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		e.stateMu.Lock()
		if e.state == target {
			e.stateMu.Unlock()
			return nil
		}
		ch := e.change
		e.stateMu.Unlock()
		select {
		case <-ch: // fires on any state change, not just the expected one
		case <-shutdownCh:
			return fmt.Errorf("shutdown while waiting for %q state %q", id, target)
		case <-timer.C:
			return fmt.Errorf("timed out waiting for %q state %q", id, target)
		}
	}
}

// WaitForService blocks until the service reaches target and returns
// the service instance. Pass nil shutdown from Shutdown methods where
// shutdown has already been triggered.
func (r *Registry) WaitForService(shutdown *utils.SyncedChan, id, target string,
	timeout time.Duration) (Service, error) {
	if err := r.waitForState(shutdown, id, target, timeout); err != nil {
		return nil, err
	}
	return r.Lookup(id), nil
}

// WaitForServices waits for all ids to reach target within timeout.
func (r *Registry) WaitForServices(shutdown *utils.SyncedChan, target string,
	ids []string, timeout time.Duration) (map[string]Service, error) {
	services := make(map[string]Service, len(ids))
	deadline := time.Now().Add(timeout)
	for _, id := range ids {
		if err := r.waitForState(shutdown, id, target, time.Until(deadline)); err != nil {
			return nil, err
		}
		services[id] = r.Lookup(id)
	}
	return services, nil
}

// LockService locks the lifecycle mutex for id and returns the unlock
// function. Panics if id is not registered.
func (r *Registry) LockService(id string) (unlock func()) {
	e := r.entry(id)
	e.lifecycleMu.Lock()
	return e.lifecycleMu.Unlock
}
