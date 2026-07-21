// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"os"
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, fileCPU *os.File, shdWg *sync.WaitGroup) *CoreService {
	return &CoreService{
		shdWg:   shdWg,
		cfg:     cfg,
		fileCPU: fileCPU,
	}
}

// CoreService implements Service interface
type CoreService struct {
	mu       sync.RWMutex
	cfg      *config.CGRConfig
	cS       *cores.CoreS
	fileCPU  *os.File
	stopChan chan struct{}
	shdWg    *sync.WaitGroup
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
	s.cS = cores.NewCoreService(s.cfg, caps, s.fileCPU, s.stopChan, s.shdWg, shutdown, cl)
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

// Shutdown stops the service
func (s *CoreService) Shutdown(registry *servmanager.Registry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cS.Shutdown()
	close(s.stopChan)
	s.cS.StopCPUProfiling()
	s.cS.StopMemoryProfiling()
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
