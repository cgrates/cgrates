// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, profiler *cores.Profiler) *CoreService {
	return &CoreService{
		cfg:      cfg,
		profiler: profiler,
	}
}

// CoreService implements Service interface
type CoreService struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	cS       *cores.CoreS
	profiler *cores.Profiler
	stopChan chan struct{}
}

// Start should handle the service start
func (s *CoreService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CapS,
			utils.CommonListenerS,
			utils.ConnManager,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopChan = make(chan struct{})
	s.cS = cores.NewCoreService(s.cfg, caps, s.profiler, s.stopChan, shutdown, cl)
	srv, err := newRPCService(apis.NewCoreSv1(s.cS), utils.CoreSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.CoreS, srv)
	return nil
}

// Reload handles the change of config
func (s *CoreService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown unregisters CoreS. Process profiling is stopped by the engine.
func (s *CoreService) Shutdown(registry *servmanager.Registry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.stopChan)
	s.cS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.CoreSv1)
	return nil
}

// ServiceName returns the service name
func (s *CoreService) ServiceName() string {
	return utils.CoreS
}

// ShouldRun returns if the service should be running
func (s *CoreService) ShouldRun() bool {
	return true
}

// CoreS returns the CoreS object.
func (s *CoreService) CoreS() *cores.CoreS {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cS
}
