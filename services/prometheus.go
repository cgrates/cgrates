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
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewPrometheusAgent returns the Prometheus Agent
func NewPrometheusAgent(cfg *config.CGRConfig, cm *engine.ConnManager, server *cores.Server,
	srvDep map[string]*sync.WaitGroup) *PrometheusAgent {
	return &PrometheusAgent{
		cfg:    cfg,
		cm:     cm,
		server: server,
		srvDep: srvDep,
	}
}

// PrometheusAgent implements Agent interface
type PrometheusAgent struct {
	mu     sync.RWMutex
	cfg    *config.CGRConfig
	cm     *engine.ConnManager
	server *cores.Server
	srvDep map[string]*sync.WaitGroup

	pa *agents.PrometheusAgent
}

// Start should handle the sercive start
func (s *PrometheusAgent) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pa = agents.NewPrometheusAgent(s.cfg, s.cm)
	s.server.RegisterHttpHandler(s.cfg.PrometheusAgentCfg().Path, s.pa)
	return nil
}

// Reload handles configuration changes.
func (s *PrometheusAgent) Reload() error {
	return nil
}

// Shutdown stops the service.
func (s *PrometheusAgent) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pa = nil
	return nil
}

// ServiceName returns the service name.
func (s *PrometheusAgent) ServiceName() string {
	return utils.PrometheusAgent
}

// ShouldRun returns if the service should be running.
func (s *PrometheusAgent) ShouldRun() bool {
	return s.cfg.PrometheusAgentCfg().Enabled
}

// IsRunning checks whether the service is running.
func (s *PrometheusAgent) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pa != nil
}
