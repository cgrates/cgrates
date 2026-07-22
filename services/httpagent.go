// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *cores.Server, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &HTTPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	server      *cores.Server

	// we can realy stop the HTTPAgent so keep a flag
	// if we registerd the handlers
	started bool
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start() (err error) {
	if ha.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-ha.filterSChan
	ha.filterSChan <- filterS

	ha.Lock()
	ha.started = true
	utils.Logger.Info(fmt.Sprintf("<%s> successfully started HTTPAgent", utils.HTTPAgent))
	for _, agntCfg := range ha.cfg.HTTPAgentCfg() {
		ha.server.RegisterHttpHandler(agntCfg.URL,
			agents.NewHTTPAgent(ha.connMgr,
				agntCfg.SessionSConns, agntCfg.StatSConns, agntCfg.ThresholdSConns,
				filterS, ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors, ha.caps))
	}
	ha.Unlock()
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload() (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown() (err error) {
	ha.Lock()
	ha.started = false
	ha.Unlock()
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ha *HTTPAgent) IsRunning() bool {
	ha.RLock()
	defer ha.RUnlock()
	return ha.started
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}

// ShouldRun returns if the service should be running
func (ha *HTTPAgent) ShouldRun() bool {
	return len(ha.cfg.HTTPAgentCfg()) != 0
}
