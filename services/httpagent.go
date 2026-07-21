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

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig) *HTTPAgent {
	return &HTTPAgent{
		cfg: cfg,
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig

	// we can realy stop the HTTPAgent so keep a flag
	// if we registerd the handlers
	started bool
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.CapS,
		},
		ha.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	caps := srvDeps[utils.CapS].(*CapService).Caps()
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	ha.mu.Lock()
	defer ha.mu.Unlock()

	ha.started = true
	for _, agntCfg := range ha.cfg.HTTPAgentCfg() {
		cl.RegisterHttpHandler(agntCfg.URL,
			agents.NewHTTPAgent(ha.cfg, cm, agntCfg.Conns, cacheS.CacheS(),
				fs, ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload, agntCfg.ReplyPayload,
				agntCfg.RequestProcessors, caps))
	}
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown(_ *servmanager.Registry) (err error) {
	ha.mu.Lock()
	ha.started = false
	ha.mu.Unlock()
	return // no shutdown for the momment
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}

// ShouldRun returns if the service should be running
func (ha *HTTPAgent) ShouldRun() bool {
	return len(ha.cfg.HTTPAgentCfg()) != 0
}
