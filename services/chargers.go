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

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewChargerService returns the Charger Service
func NewChargerService() servmanager.Service {
	return &ChargerService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// ChargerService implements Service interface
type ChargerService struct {
	sync.RWMutex
	chrS     *engine.ChargerService
	rpc      *v1.ChargerSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (chrS *ChargerService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if chrS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}
	if waitCache {
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheChargerProfiles)
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheChargerFilterIndexes)
	}
	var attrSConn rpcclient.RpcClientConnection
	if attrSConn, err = sp.GetConnection(utils.AttributeS, sp.GetConfig().ChargerSCfg().AttributeSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.ChargerS, utils.AttributeS, err.Error()))
		return
	}
	chrS.Lock()
	defer chrS.Unlock()
	if chrS.chrS, err = engine.NewChargerService(sp.GetDM(), sp.GetFilterS(), attrSConn, sp.GetConfig()); err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.ChargerS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ChargerS))
	cSv1 := v1.NewChargerSv1(chrS.chrS)
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(cSv1)
	}
	chrS.connChan <- cSv1
	return
}

// GetIntenternalChan returns the internal connection chanel
func (chrS *ChargerService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return chrS.connChan
}

// Reload handles the change of config
func (chrS *ChargerService) Reload(sp servmanager.ServiceProvider) (err error) {
	var attrSConn rpcclient.RpcClientConnection
	if attrSConn, err = sp.GetConnection(utils.AttributeS, sp.GetConfig().ChargerSCfg().AttributeSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.ChargerS, utils.AttributeS, err.Error()))
		return
	}
	chrS.Lock()
	chrS.chrS.SetAttributeConnection(attrSConn)
	chrS.Unlock()
	return
}

// Shutdown stops the service
func (chrS *ChargerService) Shutdown() (err error) {
	chrS.Lock()
	defer chrS.Unlock()
	if err = chrS.chrS.Shutdown(); err != nil {
		return
	}
	chrS.chrS = nil
	chrS.rpc = nil
	<-chrS.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (chrS *ChargerService) GetRPCInterface() interface{} {
	return chrS.rpc
}

// IsRunning returns if the service is running
func (chrS *ChargerService) IsRunning() bool {
	chrS.RLock()
	defer chrS.RUnlock()
	return chrS != nil && chrS.chrS != nil
}

// ServiceName returns the service name
func (chrS *ChargerService) ServiceName() string {
	return utils.ChargerS
}
