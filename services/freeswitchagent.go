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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewFreeswitchAgent returns the Freeswitch Agent
func NewFreeswitchAgent(cfg *config.CGRConfig,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &FreeswitchAgent{
		cfg:      cfg,
		exitChan: exitChan,
		connMgr:  connMgr,
	}
}

// FreeswitchAgent implements Agent interface
type FreeswitchAgent struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	exitChan chan bool

	fS      *agents.FSsessions
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (fS *FreeswitchAgent) Start() (err error) {
	if fS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	fS.Lock()
	defer fS.Unlock()

	fS.fS = agents.NewFSsessions(fS.cfg.FsAgentCfg(), fS.cfg.GeneralCfg().DefaultTimezone, fS.connMgr)

	go func() {
		if err := fS.fS.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			fS.exitChan <- true // stop the engine here
		}
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
// no chanel for FreeswitchAgent
func (fS *FreeswitchAgent) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return nil
}

// Reload handles the change of config
func (fS *FreeswitchAgent) Reload() (err error) {
	if err = fS.Shutdown(); err != nil {
		return
	}
	fS.Lock()
	defer fS.Unlock()
	fS.fS.Reload()
	go func() {
		if err := fS.fS.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			fS.exitChan <- true // stop the engine here
		}
	}()
	return
}

// Shutdown stops the service
func (fS *FreeswitchAgent) Shutdown() (err error) {
	fS.Lock()
	defer fS.Unlock()
	if err = fS.fS.Shutdown(); err != nil {
		return
	}
	fS.fS = nil
	return
}

// IsRunning returns if the service is running
func (fS *FreeswitchAgent) IsRunning() bool {
	fS.RLock()
	defer fS.RUnlock()
	return fS != nil && fS.fS != nil
}

// ServiceName returns the service name
func (fS *FreeswitchAgent) ServiceName() string {
	return utils.FreeSWITCHAgent
}

// ShouldRun returns if the service should be running
func (fS *FreeswitchAgent) ShouldRun() bool {
	return fS.cfg.FsAgentCfg().Enabled
}
