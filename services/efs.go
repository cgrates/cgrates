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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// ExportFailoverService is the service structure for ExportFailover
type ExportFailoverService struct {
	mu        sync.Mutex
	efS       *efs.EfS
	srv       *birpc.Service
	stopChan  chan struct{}
	cfg       *config.CGRConfig
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// NewExportFailoverService is the constructor for the TpeService
func NewExportFailoverService(cfg *config.CGRConfig) *ExportFailoverService {
	return &ExportFailoverService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// Start should handle the service start
func (s *ExportFailoverService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.efS = efs.NewEfs(s.cfg, cms.ConnManager())
	s.stopChan = make(chan struct{})
	s.srv, _ = engine.NewServiceWithPing(s.efS, utils.EfSv1, utils.V1Prfx)
	cl.RpcRegister(s.srv)
	cms.AddInternalConn(utils.EFs, s.srv)
	return
}

// Reload handles the change of config
func (s *ExportFailoverService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (s *ExportFailoverService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.srv = nil
	close(s.stopChan)

	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.EfSv1)
	// NEXT SHOULD EXPORT ALL THE SHUTDOWN LOGGERS TO WRITE
	return
}

// ShouldRun returns if the service should be running
func (s *ExportFailoverService) ShouldRun() bool {
	return s.cfg.EFsCfg().Enabled
}

// ServiceName returns the service name
func (s *ExportFailoverService) ServiceName() string {
	return utils.EFs
}

// StateChan returns signaling channel of specific state
func (s *ExportFailoverService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
