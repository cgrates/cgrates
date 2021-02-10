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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent(cfg *config.CGRConfig,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &KamailioAgent{
		cfg:     cfg,
		shdChan: shdChan,
		connMgr: connMgr,
		srvDep:  srvDep,
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	shdChan *utils.SyncedChan

	kam     *agents.KamailioAgent
	connMgr *engine.ConnManager
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start() (err error) {
	if kam.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	kam.Lock()
	defer kam.Unlock()

	kam.kam = agents.NewKamailioAgent(kam.cfg.KamAgentCfg(), kam.connMgr,
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone))

	go func(k *agents.KamailioAgent) {
		if err = k.Connect(); err != nil &&
			!strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.shdChan.CloseOnce()
		}
	}(kam.kam)
	return
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload() (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam.Reload()
	go func(k *agents.KamailioAgent) {
		if err = k.Connect(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.shdChan.CloseOnce()
		}
	}(kam.kam)
	return
}

// Shutdown stops the service
func (kam *KamailioAgent) Shutdown() (err error) {
	kam.Lock()
	defer kam.Unlock()
	if err = kam.kam.Shutdown(); err != nil {
		return
	}
	kam.kam = nil
	return
}

// IsRunning returns if the service is running
func (kam *KamailioAgent) IsRunning() bool {
	kam.RLock()
	defer kam.RUnlock()
	return kam != nil && kam.kam != nil
}

// ServiceName returns the service name
func (kam *KamailioAgent) ServiceName() string {
	return utils.KamailioAgent
}

// ShouldRun returns if the service should be running
func (kam *KamailioAgent) ShouldRun() bool {
	return kam.cfg.KamAgentCfg().Enabled
}
