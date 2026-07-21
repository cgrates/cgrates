// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		cfg: cfg,
	}
}

// PrometheusAgent implements the Service interface.
type PrometheusAgent struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
}

// Start handles the service start.
func (s *PrometheusAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fS := srvDeps[utils.FilterS].(*FilterService).FilterS()

	s.mu.Lock()
	defer s.mu.Unlock()

	pa := agents.NewPrometheusAgent(s.cfg, cm, fS)
	cl.RegisterHttpHandler(s.cfg.PrometheusAgentCfg().Path, pa)
	return
}

// Reload handles configuration changes.
func (s *PrometheusAgent) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service.
func (s *PrometheusAgent) Shutdown(_ *servmanager.Registry) (err error) {
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
