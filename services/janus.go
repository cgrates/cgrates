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
	"fmt"
	"net/http"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewJanusAgent returns the Janus Agent
func NewJanusAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *cores.Server, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &JanusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
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
	srvDep  map[string]*sync.WaitGroup
}

// Start should jandle the sercive start
func (ja *JanusAgent) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	filterS := <-ja.filterSChan
	ja.filterSChan <- filterS

	ja.Lock()
	if ja.started {
		ja.Unlock()
		return utils.ErrServiceAlreadyRunning
	}
	ja.jA, err = agents.NewJanusAgent(ja.cfg, ja.connMgr, filterS)
	if err != nil {
		return
	}
	if err = ja.jA.Connect(); err != nil {
		return
	}

	ja.server.RegisterHttpHandler("POST "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CreateSession))
	ja.server.RegisterHttpHandler("OPTIONS "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CORSOptions))
	ja.server.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.SessionKeepalive))
	ja.server.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}/", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.CORSOptions))
	ja.server.RegisterHttpHandler(fmt.Sprintf("GET %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.PollSession))
	ja.server.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.AttachPlugin))
	ja.server.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}/{handleID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.HandlePlugin))

	ja.started = true
	ja.Unlock()
	utils.Logger.Info(fmt.Sprintf("<%s> successfully started.", utils.JanusAgent))
	return
}

// Reload jandles the change of config
func (ja *JanusAgent) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
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
