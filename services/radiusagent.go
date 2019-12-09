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
	"github.com/cgrates/rpcclient"
)

// NewRadiusAgent returns the Radius Agent
func NewRadiusAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	sSChan, dispatcherChan chan rpcclient.ClientConnector,
	exitChan chan bool) servmanager.Service {
	return &RadiusAgent{
		cfg:            cfg,
		filterSChan:    filterSChan,
		sSChan:         sSChan,
		dispatcherChan: dispatcherChan,
		exitChan:       exitChan,
	}
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	filterSChan    chan *engine.FilterS
	sSChan         chan rpcclient.ClientConnector
	dispatcherChan chan rpcclient.ClientConnector
	exitChan       chan bool

	rad *agents.RadiusAgent
}

// Start should handle the sercive start
func (rad *RadiusAgent) Start() (err error) {
	if rad.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	filterS := <-rad.filterSChan
	rad.filterSChan <- filterS

	rad.Lock()
	defer rad.Unlock()
	var smgConn rpcclient.ClientConnector
	utils.Logger.Info("Starting Radius agent")
	if smgConn, err = NewConnection(rad.cfg, rad.sSChan, rad.dispatcherChan, rad.cfg.RadiusAgentCfg().SessionSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.RadiusAgent, utils.SessionS, err.Error()))
		return
	}

	if rad.rad, err = agents.NewRadiusAgent(rad.cfg, filterS, smgConn); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		return
	}

	go func() {
		if err = rad.rad.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		}
		rad.exitChan <- true
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
func (rad *RadiusAgent) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return nil
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload() (err error) {
	var smgConn rpcclient.ClientConnector
	if smgConn, err = NewConnection(rad.cfg, rad.sSChan, rad.dispatcherChan, rad.cfg.RadiusAgentCfg().SessionSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.RadiusAgent, utils.SessionS, err.Error()))
		return
	}
	rad.Lock()
	rad.rad.SetSessionSConnection(smgConn)
	rad.Unlock()
	return // partial reload
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (rad *RadiusAgent) IsRunning() bool {
	rad.RLock()
	defer rad.RUnlock()
	return rad != nil && rad.rad != nil
}

// ServiceName returns the service name
func (rad *RadiusAgent) ServiceName() string {
	return utils.RadiusAgent
}

// ShouldRun returns if the service should be running
func (rad *RadiusAgent) ShouldRun() bool {
	return rad.cfg.RadiusAgentCfg().Enabled
}
