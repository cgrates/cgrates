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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewKamailioAgent returns the Kamailio Agent
func NewKamailioAgent() servmanager.Service {
	return new(KamailioAgent)
}

// KamailioAgent implements Agent interface
type KamailioAgent struct {
	sync.RWMutex
	kam *agents.KamailioAgent
}

// Start should handle the sercive start
func (kam *KamailioAgent) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if kam.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	kam.Lock()
	defer kam.Unlock()
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting Kamailio agent")
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().KamAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.KamailioAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.NewConnection(utils.SessionS, sp.GetConfig().KamAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.KamailioAgent, utils.SessionS, err.Error()))
			return
		}
	}
	kam.kam = agents.NewKamailioAgent(sp.GetConfig().KamAgentCfg(), sS,
		utils.FirstNonEmpty(sp.GetConfig().KamAgentCfg().Timezone, sp.GetConfig().GeneralCfg().DefaultTimezone))
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
			sp.GetExitChan() <- true
		}
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
func (kam *KamailioAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (kam *KamailioAgent) Reload(sp servmanager.ServiceProvider) (err error) {
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().KamAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.KamailioAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.NewConnection(utils.SessionS, sp.GetConfig().FsAgentCfg().SessionSConns); err != nil {
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
			sp.GetExitChan() <- true
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

// GetRPCInterface returns the interface to register for server
func (kam *KamailioAgent) GetRPCInterface() interface{} {
	return kam.kam
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
