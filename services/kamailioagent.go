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
	"strings"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent(cfg *config.CGRConfig) *KamailioAgent {
	return &KamailioAgent{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	kam       *agents.KamailioAgent
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	cms, err := WaitForServiceState(utils.StateServiceUP, utils.ConnManager, registry, kam.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}

	kam.Lock()
	defer kam.Unlock()

	kam.kam = agents.NewKamailioAgent(kam.cfg.KamAgentCfg(), cms.(*ConnManagerService).ConnManager(),
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone))

	go kam.connect(kam.kam, shutdown)
	return
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload(shutdown *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam.Reload()
	go kam.connect(kam.kam, shutdown)
	return
}

func (kam *KamailioAgent) connect(k *agents.KamailioAgent, shutdown *utils.SyncedChan) (err error) {
	if err = k.Connect(); err != nil {
		if !strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
			if !strings.Contains(err.Error(), "KamEvapi") {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			}
			shutdown.CloseOnce()
		}
	}
	return
}

// Shutdown stops the service
func (kam *KamailioAgent) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	kam.Lock()
	defer kam.Unlock()
	err = kam.kam.Shutdown()
	kam.kam = nil
	return
}

// ServiceName returns the service name
func (kam *KamailioAgent) ServiceName() string {
	return utils.KamailioAgent
}

// ShouldRun returns if the service should be running
func (kam *KamailioAgent) ShouldRun() bool {
	return kam.cfg.KamAgentCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (kam *KamailioAgent) StateChan(stateID string) chan struct{} {
	return kam.stateDeps.StateChan(stateID)
}
