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

// NewApierV1Service returns the ApierV1 Service
func NewApierV1Service() *ApierV1Service {
	return &ApierV1Service{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// ApierV1Service implements Service interface
type ApierV1Service struct {
	sync.RWMutex
	api      *v1.ApierV1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (api *ApierV1Service) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if api.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	api.Lock()
	defer api.Unlock()

	// create cache connection
	var cacheSrpc, schedulerSrpc, attributeSrpc rpcclient.RpcClientConnection
	if cacheSrpc, err = sp.NewConnection(utils.CacheS, sp.GetConfig().ApierCfg().CachesConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.CacheS, err.Error()))
		return
	}

	// create scheduler connection
	if schedulerSrpc, err = sp.NewConnection(utils.SchedulerS, sp.GetConfig().ApierCfg().SchedulerConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.SchedulerS, err.Error()))
		return
	}

	// create scheduler connection
	if attributeSrpc, err = sp.NewConnection(utils.AttributeS, sp.GetConfig().ApierCfg().AttributeSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.AttributeS, err.Error()))
		return
	}

	// get scheduler
	var schS *SchedulerService
	if sch, has := sp.GetService(utils.SchedulerS); !has {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
			utils.ApierV1, utils.SchedulerS))
		err = utils.ErrNotFound
		return
	} else if sc, canCast := sch.(*SchedulerService); !canCast {
		utils.Logger.Err(fmt.Sprintf("<%s> Wrong type(%T) for subsystem <%s>",
			utils.ApierV1, sch, utils.SchedulerS))
		err = utils.ErrNotFound // return another error
		return
	} else {
		schS = sc
	}

	// get responder
	var responder *engine.Responder
	if rsp, has := sp.GetService(utils.ResponderS); !has {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
			utils.ApierV1, utils.ResponderS))
		err = utils.ErrNotFound
		return
	} else if rp, canCast := rsp.(*ResponderService); !canCast {
		utils.Logger.Err(fmt.Sprintf("<%s> Wrong type(%T) for subsystem <%s>",
			utils.ApierV1, rsp, utils.ResponderS))
		err = utils.ErrNotFound // return another error
		return
	} else {
		responder = rp.GetResponder()
	}

	api.api = &v1.ApierV1{
		StorDb:      sp.GetLoadStorage(),
		DataManager: sp.GetDM(),
		CdrDb:       sp.GetCDRStorage(),
		Config:      sp.GetConfig(),
		Responder:   responder,
		Scheduler:   schS,
		HTTPPoster: engine.NewHTTPPoster(sp.GetConfig().GeneralCfg().HttpSkipTlsVerify,
			sp.GetConfig().GeneralCfg().ReplyTimeout),
		FilterS:    sp.GetFilterS(),
		CacheS:     cacheSrpc,
		SchedulerS: schedulerSrpc,
		AttributeS: attributeSrpc}

	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(api.api)
	}

	utils.RegisterRpcParams("", &v1.CDRsV1{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", api.api)

	api.connChan <- api.api

	return
}

// GetIntenternalChan returns the internal connection chanel
func (api *ApierV1Service) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return api.connChan
}

// Reload handles the change of config
func (api *ApierV1Service) Reload(sp servmanager.ServiceProvider) (err error) {
	var cacheSrpc, schedulerSrpc, attributeSrpc rpcclient.RpcClientConnection
	if cacheSrpc, err = sp.NewConnection(utils.CacheS, sp.GetConfig().ApierCfg().CachesConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.CacheS, err.Error()))
		return
	}

	// create scheduler connection
	if schedulerSrpc, err = sp.NewConnection(utils.SchedulerS, sp.GetConfig().ApierCfg().SchedulerConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.SchedulerS, err.Error()))
		return
	}

	// create scheduler connection
	if attributeSrpc, err = sp.NewConnection(utils.AttributeS, sp.GetConfig().ApierCfg().AttributeSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.AttributeS, err.Error()))
		return
	}
	api.Lock()
	api.api.SetAttributeSConnection(attributeSrpc)
	api.api.SetCacheSConnection(cacheSrpc)
	api.api.SetSchedulerSConnection(schedulerSrpc)
	api.Unlock()
	return
}

// Shutdown stops the service
func (api *ApierV1Service) Shutdown() (err error) {
	api.Lock()
	defer api.Unlock()
	api.api = nil
	<-api.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (api *ApierV1Service) GetRPCInterface() interface{} {
	return api.api
}

// IsRunning returns if the service is running
func (api *ApierV1Service) IsRunning() bool {
	api.RLock()
	defer api.RUnlock()
	return api != nil && api.api != nil
}

// ServiceName returns the service name
func (api *ApierV1Service) ServiceName() string {
	return utils.ApierV1
}

// GetApierV1 returns the apierV1
func (api *ApierV1Service) GetApierV1() *v1.ApierV1 {
	api.RLock()
	defer api.RUnlock()
	return api.api
}
