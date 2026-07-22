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

// NewJanusAgent returns the Janus Agent
func NewJanusAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *cores.Server, connMgr *engine.ConnManager, caps *engine.Caps,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &JanusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		caps:        caps,
		srvDep:      srvDep,
	}
}

// JanusAgent implements Service interface
type JanusAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	server      *cores.Server
	jA          *agents.JanusAgent

	// we can realy stop the JanusAgent so keep a flag
	// if we registerd the jandlers
	started bool
	connMgr *engine.ConnManager
	caps    *engine.Caps
	srvDep  map[string]*sync.WaitGroup
}

// Start should jandle the sercive start
func (ja *JanusAgent) Start() (err error) {
	filterS := <-ja.filterSChan
	ja.filterSChan <- filterS

	ja.Lock()
	if ja.started {
		ja.Unlock()
		return utils.ErrServiceAlreadyRunning
	}
	ja.jA, err = agents.NewJanusAgent(ja.cfg, ja.connMgr, filterS, ja.caps)
	if err != nil {
		return
	}
	if err = ja.jA.Connect(); err != nil {
		return
	}

	ja.server.RegisterHttpFunc("POST "+ja.cfg.JanusAgentCfg().URL, ja.jA.CreateSession)
	ja.server.RegisterHttpFunc("OPTIONS "+ja.cfg.JanusAgentCfg().URL, ja.jA.CORSOptions)
	ja.server.RegisterHttpFunc(fmt.Sprintf("OPTIONS %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), ja.jA.SessionKeepalive)
	ja.server.RegisterHttpFunc(fmt.Sprintf("OPTIONS %s/{sessionID}/", ja.cfg.JanusAgentCfg().URL), ja.jA.CORSOptions)
	ja.server.RegisterHttpFunc(fmt.Sprintf("GET %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), ja.jA.PollSession)
	ja.server.RegisterHttpFunc(fmt.Sprintf("POST %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), ja.jA.AttachPlugin)
	ja.server.RegisterHttpFunc(fmt.Sprintf("POST %s/{sessionID}/{handleID}", ja.cfg.JanusAgentCfg().URL), ja.jA.HandlePlugin)

	ja.started = true
	ja.Unlock()
	utils.Logger.Info(fmt.Sprintf("<%s> successfully started.", utils.JanusAgent))
	return
}

// Reload jandles the change of config
func (ja *JanusAgent) Reload() (err error) {
	return // no reload
}

// Shutdown stops the service
func (ja *JanusAgent) Shutdown() (err error) {
	ja.Lock()
	err = ja.jA.Shutdown()
	ja.started = false
	ja.Unlock()
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ja *JanusAgent) IsRunning() bool {
	ja.RLock()
	defer ja.RUnlock()
	return ja.started
}

// ServiceName returns the service name
func (ja *JanusAgent) ServiceName() string {
	return utils.JanusAgent
}

// ShouldRun returns if the service should be running
func (ja *JanusAgent) ShouldRun() bool {
	return ja.cfg.JanusAgentCfg().Enabled
}
