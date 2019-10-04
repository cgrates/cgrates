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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDiameterAgent returns the Diameter Agent
func NewDiameterAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	sSChan, dispatcherChan chan rpcclient.RpcClientConnection,
	exitChan chan bool) servmanager.Service {
	return &DiameterAgent{
		cfg:            cfg,
		filterSChan:    filterSChan,
		sSChan:         sSChan,
		dispatcherChan: dispatcherChan,
		exitChan:       exitChan,
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	filterSChan    chan *engine.FilterS
	sSChan         chan rpcclient.RpcClientConnection
	dispatcherChan chan rpcclient.RpcClientConnection
	exitChan       chan bool

	da *agents.DiameterAgent
}

// Start should handle the sercive start
func (da *DiameterAgent) Start() (err error) {
	if da.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	filterS := <-da.filterSChan
	da.filterSChan <- filterS

	da.Lock()
	defer da.Unlock()
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting Diameter agent")
	if !da.cfg.DispatcherSCfg().Enabled && da.cfg.DiameterAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-da.sSChan
		da.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(da.cfg, da.sSChan, da.dispatcherChan, da.cfg.DiameterAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			return
		}
	}
	utils.Logger.Info("Starting CGRateS DiameterAgent service")

	da.da, err = agents.NewDiameterAgent(da.cfg, filterS, sS)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.DiameterAgent, err))
		return
	}
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(da.da)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			return
		}
	}

	go func() {
		if err = da.da.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
				utils.DiameterAgent, err))
		}
		da.exitChan <- true
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
func (da *DiameterAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (da *DiameterAgent) Reload() (err error) {
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	if !da.cfg.DispatcherSCfg().Enabled && da.cfg.DiameterAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		sSIntConn := <-da.sSChan
		da.sSChan <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = NewConnection(da.cfg, da.sSChan, da.dispatcherChan, da.cfg.DiameterAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			return
		}
	}
	da.Lock()
	defer da.Unlock()
	// da.da.SetSessionSConnection(sS)
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(da.da)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DiameterAgent, utils.SessionS, err.Error()))
			return
		}
	}
	return // partial reload
}

// Shutdown stops the service
func (da *DiameterAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (da *DiameterAgent) IsRunning() bool {
	da.RLock()
	defer da.RUnlock()
	return da != nil && da.da != nil
}

// ServiceName returns the service name
func (da *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (da *DiameterAgent) ShouldRun() bool {
	return da.cfg.DiameterAgentCfg().Enabled
}
