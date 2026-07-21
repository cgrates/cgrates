// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package servmanager_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// blockingService blocks Start on a channel so tests can control when it
// completes.
type blockingService struct {
	name        string
	shouldRun   bool
	started     atomic.Int64
	stopped     atomic.Int64
	resumeStart chan error
}

func newBlockingService(name string) *blockingService {
	return &blockingService{name: name, shouldRun: true}
}

func (s *blockingService) Start(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	s.started.Add(1)
	if s.resumeStart == nil {
		return nil
	}
	return <-s.resumeStart
}

func (s *blockingService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

func (s *blockingService) Shutdown(_ *servmanager.Registry) error {
	s.stopped.Add(1)
	return nil
}

func (s *blockingService) ShouldRun() bool     { return s.shouldRun }
func (s *blockingService) ServiceName() string { return s.name }

func newTestManager(t *testing.T, svcs ...servmanager.Service) (
	*servmanager.ServiceManager, *servmanager.Registry, *sync.WaitGroup,
) {
	t.Helper()
	cfg := config.NewDefaultCGRConfig()
	shdWg := &sync.WaitGroup{}
	registry := servmanager.NewRegistry()
	m := servmanager.NewServiceManager(shdWg, cfg, registry, svcs)
	return m, registry, shdWg
}

func waitForState(t *testing.T, r *servmanager.Registry, id, target string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if r.State(id) == target {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("service %q never reached state %q", id, target)
}

func waitShdWg(t *testing.T, shdWg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		shdWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shdWg.Wait did not return")
	}
}

func TestLifecycleSuccess(t *testing.T) {
	svc := newBlockingService("svc")
	m, registry, shdWg := newTestManager(t, svc)
	shutdown := utils.NewSyncedChan()
	defer shutdown.CloseOnce()

	m.StartServices(shutdown)
	waitForState(t, registry, "svc", utils.StateServiceUP)

	m.ShutdownServices()
	waitForState(t, registry, "svc", utils.StateServiceDOWN)

	if got := svc.stopped.Load(); got != 1 {
		t.Fatalf("stopped %d times, want 1", got)
	}
	waitShdWg(t, shdWg)
}

func TestShutdownDuringStart(t *testing.T) {
	svc := newBlockingService("svc")
	svc.resumeStart = make(chan error, 1)

	m, registry, shdWg := newTestManager(t, svc)
	shutdown := utils.NewSyncedChan()
	defer shutdown.CloseOnce()

	m.StartServices(shutdown)

	// wait for Start to be called (but stay blocking)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if svc.started.Load() == 1 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if svc.started.Load() != 1 {
		t.Fatal("Start was never called")
	}

	// shutdown must block until Start finishes
	shutdownDone := make(chan struct{})
	go func() {
		m.ShutdownServices()
		close(shutdownDone)
	}()

	time.Sleep(30 * time.Millisecond)
	select {
	case <-shutdownDone:
		t.Fatal("ShutdownServices returned while Start was still running")
	default:
	}

	// let Start finish, so Shutdown can continue
	svc.resumeStart <- nil

	select {
	case <-shutdownDone:
	case <-time.After(2 * time.Second):
		t.Fatal("ShutdownServices did not finish")
	}

	if got := svc.stopped.Load(); got != 1 {
		t.Fatalf("stopped %d times, want 1", got)
	}
	if state := registry.State("svc"); state != utils.StateServiceDOWN {
		t.Fatalf("final state = %q, want %q", state, utils.StateServiceDOWN)
	}
	waitShdWg(t, shdWg)
}

func TestStartFailure(t *testing.T) {
	svc := newBlockingService("svc")
	svc.resumeStart = make(chan error, 1)
	svc.resumeStart <- errors.New("oops")

	m, registry, shdWg := newTestManager(t, svc)
	shutdown := utils.NewSyncedChan()
	defer shutdown.CloseOnce()

	m.StartServices(shutdown)
	waitShdWg(t, shdWg)

	if state := registry.State("svc"); state != utils.StateServiceDOWN {
		t.Fatalf("state = %q after Start failure, want %q", state, utils.StateServiceDOWN)
	}

	m.ShutdownServices()

	if got := svc.stopped.Load(); got != 0 {
		t.Fatalf("Shutdown called %d times after Start failure, want 0", got)
	}
}
