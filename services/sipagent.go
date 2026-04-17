/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewSIPAgent returns the sip Agent
func NewSIPAgent(cfg *config.CGRConfig) *SIPAgent {
	return &SIPAgent{
		cfg: cfg,
	}
}

// SIPAgent implements Agent interface
type SIPAgent struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	sip       *agents.SIPAgent
	oldListen string
}

// Start should handle the sercive start
func (sip *SIPAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.FilterS,
			utils.CapS,
		},
		sip.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cm := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	caps := srvDeps[utils.CapS].(*CapService).Caps()

	sip.mu.Lock()
	defer sip.mu.Unlock()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip, err = agents.NewSIPAgent(cm, sip.cfg, fs, caps)
	if err != nil {
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
func (sip *SIPAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	if sip.oldListen == sip.cfg.SIPAgentCfg().Listen {
		return
	}
	sip.mu.Lock()
	sip.sip.Shutdown()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip.InitStopChan()
	sip.mu.Unlock()
	go sip.listenAndServe(shutdown)
	return
}

// Shutdown stops the service
func (sip *SIPAgent) Shutdown(_ *servmanager.Registry) (err error) {
	sip.mu.Lock()
	defer sip.mu.Unlock()
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
