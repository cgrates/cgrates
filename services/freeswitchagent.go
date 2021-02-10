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
)

// NewFreeswitchAgent returns the Freeswitch Agent
func NewFreeswitchAgent(cfg *config.CGRConfig,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &FreeswitchAgent{
		cfg:     cfg,
		shdChan: shdChan,
		connMgr: connMgr,
		srvDep:  srvDep,
	}
}

// FreeswitchAgent implements Agent interface
type FreeswitchAgent struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	shdChan *utils.SyncedChan

	fS      *agents.FSsessions
	connMgr *engine.ConnManager
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (fS *FreeswitchAgent) Start() (err error) {
	if fS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	fS.Lock()
	defer fS.Unlock()

	fS.fS = agents.NewFSsessions(fS.cfg.FsAgentCfg(), fS.cfg.GeneralCfg().DefaultTimezone, fS.connMgr)

	go func(f *agents.FSsessions) {
		if err := f.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			fS.shdChan.CloseOnce() // stop the engine here
		}
	}(fS.fS)
	return
}

// Reload handles the change of config
func (fS *FreeswitchAgent) Reload() (err error) {
	fS.Lock()
	defer fS.Unlock()
	if err = fS.fS.Shutdown(); err != nil {
		return
	}
	fS.fS.Reload()
	go func(f *agents.FSsessions) {
		if err := fS.fS.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			fS.shdChan.CloseOnce() // stop the engine here
		}
	}(fS.fS)
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
