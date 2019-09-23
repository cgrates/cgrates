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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewRalService returns the Ral Service
func NewRalService(sp servmanager.ServiceProvider) servmanager.Service {
	apiv1 := NewApierV1Service()
	apiv2 := NewApierV2Service(apiv1)
	resp := NewResponderService()
	sp.AddService(apiv1, apiv2, resp)
	return &RalService{
		apiv1:     apiv1,
		apiv2:     apiv2,
		responder: resp,
		connChan:  make(chan rpcclient.RpcClientConnection, 1),
	}
}

// RalService implements Service interface
type RalService struct {
	sync.RWMutex
	rals      *v1.RALsV1
	apiv1     *ApierV1Service
	apiv2     *ApierV2Service
	responder *ResponderService
	connChan  chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (rals *RalService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if rals.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	rals.Lock()
	defer rals.Unlock()

	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheDestinations)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheReverseDestinations)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheRatingPlans)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheRatingProfiles)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheActions)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheActionPlans)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheAccountActionPlans)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheActionTriggers)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheSharedGroups)
	<-sp.GetCacheS().GetPrecacheChannel(utils.CacheTimings)

	if err = rals.responder.Start(sp, waitCache); err != nil {
		return
	}

	if err = rals.apiv1.Start(sp, waitCache); err != nil {
		return
	}

	if err = rals.apiv2.Start(sp, waitCache); err != nil {
		return
	}

	rals.rals = v1.NewRALsV1()

	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(rals.rals)
	}

	utils.RegisterRpcParams(utils.RALsV1, rals.rals)

	rals.connChan <- rals.rals
	return
}

// GetIntenternalChan returns the internal connection chanel
func (rals *RalService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return rals.connChan
}

// Reload handles the change of config
func (rals *RalService) Reload(sp servmanager.ServiceProvider) (err error) {
	if err = rals.apiv1.Reload(sp); err != nil {
		return
	}
	if err = rals.apiv2.Reload(sp); err != nil {
		return
	}
	if err = rals.responder.Reload(sp); err != nil {
		return
	}
	return
}

// Shutdown stops the service
func (rals *RalService) Shutdown() (err error) {
	rals.Lock()
	defer rals.Unlock()
	if err = rals.apiv1.Shutdown(); err != nil {
		return
	}
	if err = rals.apiv2.Shutdown(); err != nil {
		return
	}
	if err = rals.responder.Shutdown(); err != nil {
		return
	}
	rals.rals = nil
	<-rals.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (rals *RalService) GetRPCInterface() interface{} {
	return rals.rals
}

// IsRunning returns if the service is running
func (rals *RalService) IsRunning() bool {
	rals.RLock()
	defer rals.RUnlock()
	return rals != nil && rals.rals != nil
}

// ServiceName returns the service name
func (rals *RalService) ServiceName() string {
	return utils.RALService
}
