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

// NewThresholdService returns the Threshold Service
func NewThresholdService() servmanager.Service {
	return &ThresholdService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	thrs     *engine.ThresholdService
	rpc      *v1.ThresholdSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if thrs.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	if waitCache {
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheThresholdProfiles)
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheThresholds)
		<-sp.GetCacheS().GetPrecacheChannel(utils.CacheThresholdFilterIndexes)
	}

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs, err = engine.NewThresholdService(sp.GetDM(), sp.GetConfig(), sp.GetFilterS())
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.ThresholdS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop()
	thrs.rpc = v1.NewThresholdSv1(thrs.thrs)
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(thrs.rpc)
	}
	thrs.connChan <- thrs.rpc
	return
}

// GetIntenternalChan returns the internal connection chanel
func (thrs *ThresholdService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return thrs.connChan
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload(sp servmanager.ServiceProvider) (err error) {
	thrs.Lock()
	thrs.thrs.Reload()
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown() (err error) {
	thrs.Lock()
	defer thrs.Unlock()
	if err = thrs.thrs.Shutdown(); err != nil {
		return
	}
	thrs.thrs = nil
	thrs.rpc = nil
	<-thrs.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (thrs *ThresholdService) GetRPCInterface() interface{} {
	return thrs.rpc
}

// IsRunning returns if the service is running
func (thrs *ThresholdService) IsRunning() bool {
	thrs.RLock()
	defer thrs.RUnlock()
	return thrs != nil && thrs.thrs != nil
}

// ServiceName returns the service name
func (thrs *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}
