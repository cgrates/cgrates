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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewConfigService instantiates a new ConfigService.
func NewConfigService(cfg *config.CGRConfig, srvIndexer *servmanager.ServiceIndexer) *ConfigService {
	return &ConfigService{
		cfg:        cfg,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// ConfigService implements Service interface.
type ConfigService struct {
	mu         sync.RWMutex
	cfg        *config.CGRConfig
	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start handles the service start.
func (s *ConfigService) Start(_ chan struct{}) error {
	cls := s.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.GuardianS, utils.CommonListenerS, utils.StateServiceUP)
	}
	anz := s.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), s.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.GuardianS, utils.AnalyzerS, utils.StateServiceUP)
	}

	srv, _ := engine.NewServiceWithName(s.cfg, utils.ConfigS, true)
	if !s.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cls.CLS().RpcRegister(s)
		}
	}
	s.intRPCconn = anz.GetInternalCodec(srv, utils.ConfigSv1)
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *ConfigService) Reload(_ chan struct{}) error {
	return nil
}

// Shutdown stops the service.
func (s *ConfigService) Shutdown() error {
	return nil
}

// IsRunning returns whether the service is running or not.
func (s *ConfigService) IsRunning() bool {
	return CheckServiceState(s.ServiceName(), utils.StateServiceUP, s.srvIndexer)
}

// ServiceName returns the service name
func (s *ConfigService) ServiceName() string {
	return utils.ConfigS
}

// ShouldRun returns if the service should be running.
func (s *ConfigService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (s *ConfigService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (s *ConfigService) IntRPCConn() birpc.ClientConnector {
	return s.intRPCconn
}
