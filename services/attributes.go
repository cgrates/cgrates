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

// NewAttributeService returns the Attribute Service
func NewAttributeService() servmanager.Service {
	return &AttributeService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	sync.RWMutex
	attrS    *engine.AttributeService
	rpc      *v1.AttributeSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (attrS *AttributeService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if attrS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	if waitCache {
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheAttributeProfiles)
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheAttributeFilterIndexes)
	}

	attrS.Lock()
	defer attrS.Unlock()
	attrS.attrS, err = engine.NewAttributeService(sp.GetDM(),
		sp.GetFilterS(), sp.GetConfig())
	if err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.AttributeS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.ServiceManager, utils.AttributeS))
	attrS.rpc = v1.NewAttributeSv1(attrS.attrS)
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(attrS.rpc)
	}
	attrS.connChan <- attrS.rpc
	return
}

// GetIntenternalChan returns the internal connection chanel
func (attrS *AttributeService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return attrS.connChan
}

// Reload handles the change of config
func (attrS *AttributeService) Reload(sp servmanager.ServiceProvider) (err error) {
	return
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown() (err error) {
	attrS.Lock()
	defer attrS.Unlock()
	if err = attrS.attrS.Shutdown(); err != nil {
		return
	}
	attrS.attrS = nil
	attrS.rpc = nil
	<-attrS.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (attrS *AttributeService) GetRPCInterface() interface{} {
	return attrS.rpc
}

// IsRunning returns if the service is running
func (attrS *AttributeService) IsRunning() bool {
	attrS.RLock()
	defer attrS.RUnlock()
	return attrS != nil && attrS.attrS != nil
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}
