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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewPrometheusAgent returns the Prometheus Agent
func NewPrometheusAgent(cfg *config.CGRConfig) *PrometheusAgent {
	return &PrometheusAgent{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// PrometheusAgent implements Agent interface
type PrometheusAgent struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig

	stateDeps *StateDependencies
}

// Start should handle the sercive start
func (s *PrometheusAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()

	s.mu.Lock()
	defer s.mu.Unlock()

	pa := agents.NewPrometheusAgent(s.cfg, fs, cm, shutdown)
	cl.RegisterHttpHandler(s.cfg.PrometheusAgentCfg().Path, pa)
	return
}

// Reload handles configuration changes.
func (s *PrometheusAgent) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service.
func (s *PrometheusAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	return
}

// ServiceName returns the service name.
func (s *PrometheusAgent) ServiceName() string {
	return utils.PrometheusAgent
}

// ShouldRun returns if the service should be running.
func (s *PrometheusAgent) ShouldRun() bool {
	return s.cfg.PrometheusAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (s *PrometheusAgent) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
