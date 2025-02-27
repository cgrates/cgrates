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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewFreeswitchAgent returns the Freeswitch Agent
func NewFreeswitchAgent(cfg *config.CGRConfig) *FreeswitchAgent {
	return &FreeswitchAgent{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// FreeswitchAgent implements Agent interface
type FreeswitchAgent struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	fS        *agents.FSsessions
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (fS *FreeswitchAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.FilterS,
		},
		registry, fS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService)

	fS.Lock()
	defer fS.Unlock()

	fS.fS, err = agents.NewFSsessions(fs.cfg, fs.FilterS(), fS.cfg.GeneralCfg().DefaultTimezone, cms.ConnManager())
	if err != nil {
		return
	}
	go fS.connect(shutdown)
	return
}

// Reload handles the change of config
func (fS *FreeswitchAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	fS.Lock()
	defer fS.Unlock()
	if err = fS.fS.Shutdown(); err != nil {
		return
	}
	fS.fS.Reload()
	go fS.connect(shutdown)
	return
}

func (fS *FreeswitchAgent) connect(shutdown *utils.SyncedChan) {
	if err := fS.fS.Connect(); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
		shutdown.CloseOnce() // stop the engine here
	}
	return
}

// Shutdown stops the service
func (fS *FreeswitchAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	fS.Lock()
	defer fS.Unlock()
	err = fS.fS.Shutdown()
	fS.fS = nil
	return
}

// ServiceName returns the service name
func (fS *FreeswitchAgent) ServiceName() string {
	return utils.FreeSWITCHAgent
}

// ShouldRun returns if the service should be running
func (fS *FreeswitchAgent) ShouldRun() bool {
	return fS.cfg.FsAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (fS *FreeswitchAgent) StateChan(stateID string) chan struct{} {
	return fS.stateDeps.StateChan(stateID)
}
