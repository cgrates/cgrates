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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent(cfg *config.CGRConfig, sSChan,
	dispatcherChan chan rpcclient.ClientConnector,
	exitChan chan bool) servmanager.Service {
	return &KamailioAgent{
		cfg:            cfg,
		sSChan:         sSChan,
		dispatcherChan: dispatcherChan,
		exitChan:       exitChan,
	}
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	sSChan         chan rpcclient.ClientConnector
	dispatcherChan chan rpcclient.ClientConnector
	exitChan       chan bool

	kam *agents.KamailioAgent
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start() (err error) {
	if kam.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	kam.Lock()
	defer kam.Unlock()
	var sS rpcclient.ClientConnector
	var sSInternal bool
	utils.Logger.Info("Starting Kamailio agent")
	if !kam.cfg.DispatcherSCfg().Enabled && kam.cfg.KamAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-kam.sSChan
		kam.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(kam.cfg, kam.sSChan, kam.dispatcherChan, kam.cfg.KamAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			return
		}
	}
	kam.kam = agents.NewKamailioAgent(kam.cfg.KamAgentCfg(), sS,
		utils.FirstNonEmpty(kam.cfg.KamAgentCfg().Timezone, kam.cfg.GeneralCfg().DefaultTimezone))
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(kam.kam)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			return
		}
	}
	go func() {
		if err = kam.kam.Connect(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.exitChan <- true
		}
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
func (kam *KamailioAgent) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return nil
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload() (err error) {
	var sS rpcclient.ClientConnector
	var sSInternal bool
	if !kam.cfg.DispatcherSCfg().Enabled && kam.cfg.KamAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-kam.sSChan
		kam.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(kam.cfg, kam.sSChan, kam.dispatcherChan, kam.cfg.KamAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			return
		}
	}
	if err = kam.Shutdown(); err != nil {
		return
	}
	kam.Lock()
	defer kam.Unlock()
	kam.kam.SetSessionSConnection(sS)
	kam.kam.Reload()
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(kam.kam)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			return
		}
	}
	go func() {
		if err = kam.kam.Connect(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") { // if closed by us do not log
				return
			}
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s", utils.KamailioAgent, err))
			kam.exitChan <- true
		}
	}()
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
