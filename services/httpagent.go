// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *utils.Server, connMgr *engine.ConnManager) servmanager.Service {
	return &HTTPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	server      *utils.Server

	ha      *agents.HTTPAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start() (err error) {
	if ha.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-ha.filterSChan
	ha.filterSChan <- filterS

	ha.Lock()
	defer ha.Unlock()
	utils.Logger.Info("Starting HTTP agent")
	for _, agntCfg := range ha.cfg.HttpAgentCfg() {
		ha.server.RegisterHttpHandler(agntCfg.Url,
			agents.NewHTTPAgent(ha.connMgr, agntCfg.SessionSConns, filterS,
				ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload() (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown() (err error) {
	return // no shutdown for the moment
}

// IsRunning returns if the service is running
func (ha *HTTPAgent) IsRunning() bool {
	ha.RLock()
	defer ha.RUnlock()
	return ha != nil && ha.ha != nil
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}

// ShouldRun returns if the service should be running
func (ha *HTTPAgent) ShouldRun() bool {
	return len(ha.cfg.HttpAgentCfg()) != 0
}
