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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewGuardianService instantiates a new GuardianService.
func NewGuardianService(cfg *config.CGRConfig, srvIndexer *servmanager.ServiceIndexer) *GuardianService {
	return &GuardianService{
		cfg:        cfg,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// GuardianService implements Service interface.
type GuardianService struct {
	mu         sync.RWMutex
	cfg        *config.CGRConfig
	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start handles the service start.
func (s *GuardianService) Start(_ chan struct{}) error {
	cls := s.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.GuardianS, utils.CommonListenerS, utils.StateServiceUP)
	}
	anz := s.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.GuardianS, utils.AnalyzerS, utils.StateServiceUP)
	}
	srv, _ := engine.NewServiceWithName(guardian.Guardian, utils.GuardianS, true)
	if !s.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cls.CLS().RpcRegister(s)
		}
	}
	s.intRPCconn = anz.GetInternalCodec(srv, utils.GuardianS)
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *GuardianService) Reload(_ chan struct{}) error {
	return nil
}

// Shutdown stops the service.
func (s *GuardianService) Shutdown() error {
	return nil
}

// IsRunning returns whether the service is running or not.
func (s *GuardianService) IsRunning() bool {
	return true
}

// ServiceName returns the service name
func (s *GuardianService) ServiceName() string {
	return utils.FilterS
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
