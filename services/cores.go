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
	"os"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, fileCPU *os.File, shdWg *sync.WaitGroup) *CoreService {
	return &CoreService{
		shdWg:     shdWg,
		cfg:       cfg,
		fileCPU:   fileCPU,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CoreService implements Service interface
type CoreService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	cS        *cores.CoreS
	fileCPU   *os.File
	stopChan  chan struct{}
	shdWg     *sync.WaitGroup
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (s *CoreService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CapS,
			utils.CommonListenerS,
			utils.ConnManager,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopChan = make(chan struct{})
	s.cS = cores.NewCoreService(s.cfg, caps, s.fileCPU, s.stopChan, s.shdWg, shutdown)
	srv, err := engine.NewService(s.cS)
	if err != nil {
		return err
	}
	for _, svc := range srv {
		cl.RpcRegister(svc)
	}
	cms.AddInternalConn(utils.CoreS, srv)
	return nil
}

// Reload handles the change of config
func (s *CoreService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service
func (s *CoreService) Shutdown(registry *servmanager.ServiceRegistry) error {
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

// StateChan returns signaling channel of specific state
func (s *CoreService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// CoreS returns the CoreS object.
func (s *CoreService) CoreS() *cores.CoreS {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cS
}
