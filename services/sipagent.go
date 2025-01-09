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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewSIPAgent returns the sip Agent
func NewSIPAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *SIPAgent {
	return &SIPAgent{
		cfg:       cfg,
		connMgr:   connMgr,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// SIPAgent implements Agent interface
type SIPAgent struct {
	sync.RWMutex
	cfg *config.CGRConfig

	sip     *agents.SIPAgent
	connMgr *engine.ConnManager

	oldListen string

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (sip *SIPAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	fs, err := waitForServiceState(utils.StateServiceUP, utils.FilterS, registry,
		sip.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}

	sip.Lock()
	defer sip.Unlock()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip, err = agents.NewSIPAgent(sip.connMgr, sip.cfg, fs.(*FilterService).FilterS())
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.SIPAgent, err))
		return
	}
	go sip.listenAndServe(shutdown)
	return
}

func (sip *SIPAgent) listenAndServe(shutdown *utils.SyncedChan) {
	if err := sip.sip.ListenAndServe(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
		shutdown.CloseOnce() // stop the engine here
	}
}

// Reload handles the change of config
func (sip *SIPAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	if sip.oldListen == sip.cfg.SIPAgentCfg().Listen {
		return
	}
	sip.Lock()
	sip.sip.Shutdown()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip.InitStopChan()
	sip.Unlock()
	go sip.listenAndServe(shutdown)
	return
}

// Shutdown stops the service
func (sip *SIPAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	sip.Lock()
	defer sip.Unlock()
	sip.sip.Shutdown()
	sip.sip = nil
	return
}

// ServiceName returns the service name
func (sip *SIPAgent) ServiceName() string {
	return utils.SIPAgent
}

// ShouldRun returns if the service should be running
func (sip *SIPAgent) ShouldRun() bool {
	return sip.cfg.SIPAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (sip *SIPAgent) StateChan(stateID string) chan struct{} {
	return sip.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (sip *SIPAgent) IntRPCConn() birpc.ClientConnector {
	return sip.intRPCconn
}
