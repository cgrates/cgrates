// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package servmanager

import (
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type mockService struct {
	name      string
	start     func(*mockService) error
	reload    func(*mockService) error
	shutdown  func(*mockService) error
	isRunning bool
	shouldRun bool
}

func (m *mockService) Start() error {
	return m.start(m)
}
func (m *mockService) Reload() error {
	return m.reload(m)
}
func (m *mockService) Shutdown() error {
	return m.shutdown(m)
}
func (m *mockService) IsRunning() bool {
	return m.isRunning
}
func (m *mockService) ShouldRun() bool {
	return m.shouldRun
}
func (m *mockService) ServiceName() string {
	return m.name
}

// This test aims to highlight an issue related to services, where the waitgroup counter goes
// up before we even try to start the service. So, if the start fails, the counter doesn't go
// back down. This messes up the graceful shutdown of the cgr-engine.
func TestStartServicesDeadlock(t *testing.T) {
	t.Skip("Skipping this test until we start working on the service implementation")
	cfg := config.NewDefaultCGRConfig()
	shdWg := new(sync.WaitGroup)
	shdChan := utils.NewSyncedChan()
	srvManager := NewServiceManager(cfg, shdChan, shdWg, nil)

	// The service will fail to start due to the utils.ErrNotImplemented error.
	mockSrv := &mockService{
		name: "mockService",
		start: func(m *mockService) error {
			return utils.ErrNotImplemented
		},
		reload: func(m *mockService) error {
			return nil
		},
		shutdown: func(m *mockService) error {
			m.isRunning = false
			return nil
		},
		isRunning: false,
		shouldRun: true,
	}

	srvManager.AddServices(mockSrv)
	err := srvManager.StartServices()
	if err != nil {
		t.Fatal(err)
	}
	<-shdChan.Done()
	done := make(chan struct{})
	go func() {
		shdWg.Wait()
		close(done)
	}()
	select {
	case <-done:
		if mockSrv.isRunning {
			t.Error("expected mock service to be stopped")
		}
	case <-time.After(2 * time.Millisecond):
		t.Error("waitgroup shdWg's counter is not 0")
	}
}
