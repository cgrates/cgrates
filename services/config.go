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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewConfigService instantiates a new ConfigService.
func NewConfigService(cfg *config.CGRConfig) *ConfigService {
	return &ConfigService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ConfigService implements Service interface.
type ConfigService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	cl        *commonlisteners.CommonListenerS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *ConfigService) Start(_ chan struct{}, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	s.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)

	svcs, _ := engine.NewServiceWithName(s.cfg, utils.ConfigS, true)
	if !s.cfg.DispatcherSCfg().Enabled {
		for _, svc := range svcs {
			s.cl.RpcRegister(svc)
		}
	}
	cms.AddInternalConn(utils.ConfigS, svcs)
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *ConfigService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
	return nil
}

// Shutdown stops the service.
func (s *ConfigService) Shutdown(_ *servmanager.ServiceRegistry) error {
	s.cl.RpcUnregisterName(utils.ConfigSv1)
	close(s.StateChan(utils.StateServiceDOWN))
	return nil
}

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
