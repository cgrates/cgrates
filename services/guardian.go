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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewGuardianService instantiates a new GuardianService.
func NewGuardianService(cfg *config.CGRConfig) *GuardianService {
	return &GuardianService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// GuardianService implements Service interface.
type GuardianService struct {
	mu         sync.RWMutex
	cfg        *config.CGRConfig
	cl         *commonlisteners.CommonListenerS
	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start handles the service start.
func (s *GuardianService) Start(_ chan struct{}, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.AnalyzerS,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	s.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	s.mu.Lock()
	defer s.mu.Unlock()

	svcs, _ := engine.NewServiceWithName(guardian.Guardian, utils.GuardianS, true)
	if !s.cfg.DispatcherSCfg().Enabled {
		for _, svc := range svcs {
			s.cl.RpcRegister(svc)
		}
	}
	s.intRPCconn = anz.GetInternalCodec(svcs, utils.GuardianS)
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *GuardianService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *GuardianService) Shutdown(_ *servmanager.ServiceRegistry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cl.RpcUnregisterName(utils.GuardianSv1)
	close(s.StateChan(utils.StateServiceDOWN))
	return nil
}

// ServiceName returns the service name
func (s *GuardianService) ServiceName() string {
	return utils.GuardianS
}

// ShouldRun returns if the service should be running.
func (s *GuardianService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (s *GuardianService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (s *GuardianService) IntRPCConn() birpc.ClientConnector {
	return s.intRPCconn
}
