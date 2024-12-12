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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewJanusAgent returns the Janus Agent
func NewJanusAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *JanusAgent {
	return &JanusAgent{
		cfg:        cfg,
		connMgr:    connMgr,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// JanusAgent implements Service interface
type JanusAgent struct {
	sync.RWMutex

	jA *agents.JanusAgent

	// we can realy stop the JanusAgent so keep a flag
	// if we registerd the jandlers
	started bool

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should jandle the sercive start
func (ja *JanusAgent) Start(_ chan struct{}) (err error) {
	cls := ja.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), ja.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.JanusAgent, utils.CommonListenerS, utils.StateServiceUP)
	}
	cl := cls.CLS()
	fs := ja.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), ja.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.JanusAgent, utils.FilterS, utils.StateServiceUP)
	}

	ja.Lock()
	if ja.started {
		ja.Unlock()
		return utils.ErrServiceAlreadyRunning
	}
	ja.jA, err = agents.NewJanusAgent(ja.cfg, ja.connMgr, fs.FilterS())
	if err != nil {
		return
	}
	if err = ja.jA.Connect(); err != nil {
		return
	}

	cl.RegisterHttpHandler("POST "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CreateSession))
	cl.RegisterHttpHandler("OPTIONS "+ja.cfg.JanusAgentCfg().URL, http.HandlerFunc(ja.jA.CORSOptions))
	cl.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.SessionKeepalive))
	cl.RegisterHttpHandler(fmt.Sprintf("OPTIONS %s/{sessionID}/", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.CORSOptions))
	cl.RegisterHttpHandler(fmt.Sprintf("GET %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.PollSession))
	cl.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.AttachPlugin))
	cl.RegisterHttpHandler(fmt.Sprintf("POST %s/{sessionID}/{handleID}", ja.cfg.JanusAgentCfg().URL), http.HandlerFunc(ja.jA.HandlePlugin))

	ja.started = true
	ja.Unlock()
	close(ja.stateDeps.StateChan(utils.StateServiceUP))
	utils.Logger.Info(fmt.Sprintf("<%s> successfully started.", utils.JanusAgent))
	return
}

// Reload jandles the change of config
func (ja *JanusAgent) Reload(_ chan struct{}) (err error) {
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

// StateChan returns signaling channel of specific state
func (ja *JanusAgent) StateChan(stateID string) chan struct{} {
	return ja.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (ja *JanusAgent) IntRPCConn() birpc.ClientConnector {
	return ja.intRPCconn
}
