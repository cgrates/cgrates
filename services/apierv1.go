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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewApierV1Service returns the ApierV1 Service
func NewApierV1Service(cfg *config.CGRConfig, dm *DataDBService,
	cdrStorage engine.CdrStorage, loadStorage engine.LoadStorage,
	filterSChan chan *engine.FilterS,
	server *utils.Server, cacheSChan, schedChan, attrsChan,
	dispatcherChan chan rpcclient.RpcClientConnection,
	schedService *SchedulerService,
	responderService *ResponderService) *ApierV1Service {
	return &ApierV1Service{
		connChan:         make(chan rpcclient.RpcClientConnection, 1),
		cfg:              cfg,
		dm:               dm,
		cdrStorage:       cdrStorage,
		loadStorage:      loadStorage,
		filterSChan:      filterSChan,
		server:           server,
		cacheSChan:       cacheSChan,
		schedChan:        schedChan,
		attrsChan:        attrsChan,
		dispatcherChan:   dispatcherChan,
		schedService:     schedService,
		responderService: responderService,
	}
}

// ApierV1Service implements Service interface
type ApierV1Service struct {
	sync.RWMutex
	cfg              *config.CGRConfig
	dm               *DataDBService
	cdrStorage       engine.CdrStorage
	loadStorage      engine.LoadStorage
	filterSChan      chan *engine.FilterS
	server           *utils.Server
	cacheSChan       chan rpcclient.RpcClientConnection
	schedChan        chan rpcclient.RpcClientConnection
	attrsChan        chan rpcclient.RpcClientConnection
	dispatcherChan   chan rpcclient.RpcClientConnection
	schedService     *SchedulerService
	responderService *ResponderService

	api      *v1.ApierV1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (api *ApierV1Service) Start() (err error) {
	if api.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	filterS := <-api.filterSChan
	api.filterSChan <- filterS

	api.Lock()
	defer api.Unlock()

	// create cache connection
	var cacheSrpc, schedulerSrpc, attributeSrpc rpcclient.RpcClientConnection
	if cacheSrpc, err = NewConnection(api.cfg, api.cacheSChan, api.dispatcherChan, api.cfg.ApierCfg().CachesConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.CacheS, err.Error()))
		return
	}

	// create scheduler connection
	if schedulerSrpc, err = NewConnection(api.cfg, api.schedChan, api.dispatcherChan, api.cfg.ApierCfg().SchedulerConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.SchedulerS, err.Error()))
		return
	}

	// create scheduler connection
	if attributeSrpc, err = NewConnection(api.cfg, api.attrsChan, api.dispatcherChan, api.cfg.ApierCfg().AttributeSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.AttributeS, err.Error()))
		return
	}

	api.api = &v1.ApierV1{
		StorDb:           api.loadStorage,
		DataManager:      api.dm.GetDM(),
		CdrDb:            api.cdrStorage,
		Config:           api.cfg,
		Responder:        api.responderService.GetResponder(),
		SchedulerService: api.schedService,
		HTTPPoster: engine.NewHTTPPoster(api.cfg.GeneralCfg().HttpSkipTlsVerify,
			api.cfg.GeneralCfg().ReplyTimeout),
		FilterS:    filterS,
		CacheS:     cacheSrpc,
		SchedulerS: schedulerSrpc,
		AttributeS: attributeSrpc}

	if !api.cfg.DispatcherSCfg().Enabled {
		api.server.RpcRegister(api.api)
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
func (api *ApierV1Service) Reload() (err error) {
	var cacheSrpc, schedulerSrpc, attributeSrpc rpcclient.RpcClientConnection
	if cacheSrpc, err = NewConnection(api.cfg, api.cacheSChan, api.dispatcherChan, api.cfg.ApierCfg().CachesConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.CacheS, err.Error()))
		return
	}

	// create scheduler connection
	if schedulerSrpc, err = NewConnection(api.cfg, api.schedChan, api.dispatcherChan, api.cfg.ApierCfg().SchedulerConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s, error: %s",
			utils.ApierV1, utils.SchedulerS, err.Error()))
		return
	}

	// create scheduler connection
	if attributeSrpc, err = NewConnection(api.cfg, api.attrsChan, api.dispatcherChan, api.cfg.ApierCfg().AttributeSConns); err != nil {
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

// ShouldRun returns if the service should be running
func (api *ApierV1Service) ShouldRun() bool {
	return api.cfg.RalsCfg().Enabled
}
