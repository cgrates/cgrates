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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewSIPAgent returns the sip Agent
func NewSIPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &SIPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     connMgr,
		srvDep:      srvDep,
	}
}

// SIPAgent implements Agent interface
type SIPAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	shdChan     *utils.SyncedChan

	sip     *agents.SIPAgent
	connMgr *engine.ConnManager
	srvDep  map[string]*sync.WaitGroup

	oldListen string
}

// Start should handle the sercive start
func (sip *SIPAgent) Start() (err error) {
	if sip.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-sip.filterSChan
	sip.filterSChan <- filterS

	sip.Lock()
	defer sip.Unlock()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip, err = agents.NewSIPAgent(sip.connMgr, sip.cfg, filterS)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.SIPAgent, err))
		return
	}
	go func() {
		if err = sip.sip.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
			sip.shdChan.CloseOnce() // stop the engine here
		}
	}()
	return
}

// Reload handles the change of config
func (sip *SIPAgent) Reload() (err error) {
	if sip.oldListen == sip.cfg.SIPAgentCfg().Listen {
		return
	}
	sip.Lock()
	sip.sip.Shutdown()
	sip.oldListen = sip.cfg.SIPAgentCfg().Listen
	sip.sip.InitStopChan()
	sip.Unlock()
	go func() {
		if err := sip.sip.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.SIPAgent, err.Error()))
			sip.shdChan.CloseOnce() // stop the engine here
		}
	}()
	return
}

// Shutdown stops the service
func (sip *SIPAgent) Shutdown() (err error) {
	sip.Lock()
	defer sip.Unlock()
	sip.sip.Shutdown()
	sip.sip = nil
	return
}

// IsRunning returns if the service is running
func (sip *SIPAgent) IsRunning() bool {
	sip.RLock()
	defer sip.RUnlock()
	return sip != nil && sip.sip != nil
}

// ServiceName returns the service name
func (sip *SIPAgent) ServiceName() string {
	return utils.SIPAgent
}

// ShouldRun returns if the service should be running
func (sip *SIPAgent) ShouldRun() bool {
	return sip.cfg.SIPAgentCfg().Enabled
}
