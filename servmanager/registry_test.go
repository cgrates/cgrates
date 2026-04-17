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

package servmanager_test

import (
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

type mockService struct {
	name      string
	shouldRun bool
}

func (m *mockService) Start(*utils.SyncedChan, *servmanager.Registry) error  { return nil }
func (m *mockService) Reload(*utils.SyncedChan, *servmanager.Registry) error { return nil }
func (m *mockService) Shutdown(*servmanager.Registry) error                  { return nil }
func (m *mockService) ShouldRun() bool                                       { return m.shouldRun }
func (m *mockService) ServiceName() string                                   { return m.name }

func newTestRegistry(svcs ...servmanager.Service) *servmanager.Registry {
	r := servmanager.NewRegistry()
	r.Register(svcs...)
	return r
}

func TestRegistryState(t *testing.T) {
	r := newTestRegistry(&mockService{
		name:      "svc",
		shouldRun: true,
	})

	if got := r.State("svc"); got != utils.StateServiceDOWN {
		t.Fatalf("default state = %q, want %q", got, utils.StateServiceDOWN)
	}
	if got := r.State("missing"); got != "" {
		t.Fatalf("State(unknown) = %q, want empty", got)
	}
	if err := r.SetState("svc", utils.StateServiceUP); err != nil {
		t.Fatalf("SetState: %v", err)
	}
	if got := r.State("svc"); got != utils.StateServiceUP {
		t.Fatalf("state after SetState = %q, want %q", got, utils.StateServiceUP)
	}
}

func TestRegistryWaitForService(t *testing.T) {
	r := newTestRegistry(&mockService{
		name:      "svc",
		shouldRun: true,
	})
	shutdown := utils.NewSyncedChan()

	type result struct {
		svc servmanager.Service
		err error
	}
	done := make(chan result, 1)
	go func() {
		svc, err := r.WaitForService(shutdown, "svc", utils.StateServiceUP, 2*time.Second)
		done <- result{svc, err}
	}()

	time.Sleep(20 * time.Millisecond)
	if err := r.SetState("svc", utils.StateServiceUP); err != nil {
		t.Fatalf("SetState: %v", err)
	}

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("WaitForService: %v", res.err)
		}
		if res.svc == nil {
			t.Fatal("WaitForService returned nil service")
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForService did not wake up")
	}

	// times out if the state never changes
	_, err := r.WaitForService(shutdown, "svc", utils.StateServiceDOWN, 30*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRegistryLockService(t *testing.T) {
	r := newTestRegistry(&mockService{
		name:      "svc",
		shouldRun: true,
	})

	unlock := r.LockService("svc")

	// second lock must block until the first unlocks
	unblocked := make(chan struct{})
	go func() {
		unlock2 := r.LockService("svc")
		close(unblocked)
		unlock2()
	}()

	select {
	case <-unblocked:
		t.Fatal("second lock didn't block")
	case <-time.After(20 * time.Millisecond):
	}

	unlock()

	select {
	case <-unblocked:
	case <-time.After(time.Second):
		t.Fatal("second lock didn't unblock after unlock")
	}
}

func TestRegistryConcurrent(t *testing.T) { // best run with -race flag
	r := newTestRegistry(&mockService{
		name:      "svc",
		shouldRun: true,
	})
	shutdown := utils.NewSyncedChan()
	defer shutdown.CloseOnce()

	var wg sync.WaitGroup

	for range 8 {
		wg.Go(func() {
			for range 100 {
				_ = r.State("svc")
			}
		})
	}

	for i := range 4 {
		wg.Go(func() {
			for j := range 100 {
				target := utils.StateServiceUP
				if (i+j)%2 == 0 {
					target = utils.StateServiceDOWN
				}
				_ = r.SetState("svc", target)
			}
		})
	}

	for range 4 {
		wg.Go(func() {
			for range 10 {
				_, _ = r.WaitForService(shutdown, "svc", utils.StateServiceUP, time.Millisecond)
			}
		})
	}

	wg.Wait()
}
