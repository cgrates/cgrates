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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRadiusAgent returns the Radius Agent
func NewRadiusAgent() servmanager.Service {
	return new(RadiusAgent)
}

// RadiusAgent implements Agent interface
type RadiusAgent struct {
	sync.RWMutex
	rad *agents.RadiusAgent
}

// Start should handle the sercive start
func (rad *RadiusAgent) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if rad.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	rad.Lock()
	defer rad.Unlock()
	var smgConn rpcclient.RpcClientConnection
	utils.Logger.Info("Starting Radius agent")
	if smgConn, err = sp.NewConnection(utils.SessionS, sp.GetConfig().RadiusAgentCfg().SessionSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.RadiusAgent, utils.SessionS, err.Error()))
		return
	}

	if rad.rad, err = agents.NewRadiusAgent(sp.GetConfig(), sp.GetFilterS(), smgConn); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		return
	}

	go func() {
		if err = rad.rad.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.RadiusAgent, err.Error()))
		}
		sp.GetExitChan() <- true
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
func (rad *RadiusAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (rad *RadiusAgent) Reload(sp servmanager.ServiceProvider) (err error) {
	var smgConn rpcclient.RpcClientConnection
	if smgConn, err = sp.NewConnection(utils.SessionS, sp.GetConfig().RadiusAgentCfg().SessionSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.RadiusAgent, utils.SessionS, err.Error()))
		return
	}
	rad.Lock()
	defer rad.Unlock()
	rad.rad.SetSessionSConnection(smgConn)
	return // partial reload
}

// Shutdown stops the service
func (rad *RadiusAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// GetRPCInterface returns the interface to register for server
func (rad *RadiusAgent) GetRPCInterface() interface{} {
	return rad.rad
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
