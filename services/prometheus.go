// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
