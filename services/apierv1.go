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
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewApierV1Service returns the ApierV1 Service
func NewApierV1Service(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *utils.Server,
	schedService *SchedulerService,
	responderService *ResponderService,
	internalAPIerV1Chan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager) *ApierV1Service {
	return &ApierV1Service{
		connChan:         internalAPIerV1Chan,
		cfg:              cfg,
		dm:               dm,
		storDB:           storDB,
		filterSChan:      filterSChan,
		server:           server,
		schedService:     schedService,
		responderService: responderService,
		connMgr:          connMgr,
		apierV1Chan:      make(chan *v1.ApierV1, 1),
	}
}

// ApierV1Service implements Service interface
type ApierV1Service struct {
	sync.RWMutex
	cfg              *config.CGRConfig
	dm               *DataDBService
	storDB           *StorDBService
	filterSChan      chan *engine.FilterS
	server           *utils.Server
	schedService     *SchedulerService
	responderService *ResponderService
	connMgr          *engine.ConnManager

	api      *v1.ApierV1
	connChan chan rpcclient.ClientConnector

	syncStop chan struct{}

	apierV1Chan chan *v1.ApierV1
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (apiService *ApierV1Service) Start() (err error) {
	if apiService.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	filterS := <-apiService.filterSChan
	apiService.filterSChan <- filterS
	dbchan := apiService.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	apiService.Lock()
	defer apiService.Unlock()

	storDBChan := make(chan engine.StorDB, 1)
	apiService.syncStop = make(chan struct{})
	apiService.storDB.RegisterSyncChan(storDBChan)
	stordb := <-storDBChan

	apiService.api = &v1.ApierV1{
		DataManager:      datadb,
		CdrDb:            stordb,
		StorDb:           stordb,
		Config:           apiService.cfg,
		Responder:        apiService.responderService.GetResponder(),
		SchedulerService: apiService.schedService,
		FilterS:          filterS,
		ConnMgr:          apiService.connMgr,
		StorDBChan:       storDBChan,
	}

	go func(api *v1.ApierV1, stopChan chan struct{}) {
		if err := api.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.CDRServer, err.Error()))
			// erS.exitChan <- true
		}
	}(apiService.api, apiService.syncStop)
	time.Sleep(1)

	if !apiService.cfg.DispatcherSCfg().Enabled {
		apiService.server.RpcRegister(apiService.api)
		apiService.server.RpcRegister(v1.NewReplicatorSv1(datadb))
	}

	utils.RegisterRpcParams("", &v1.CDRsV1{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", apiService.api)

	apiService.connChan <- apiService.api

	apiService.apierV1Chan <- apiService.api
	return
}

// GetIntenternalChan returns the internal connection chanel
func (apiService *ApierV1Service) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return apiService.connChan
}

// Reload handles the change of config
func (apiService *ApierV1Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (apiService *ApierV1Service) Shutdown() (err error) {
	apiService.Lock()
	close(apiService.syncStop)
	apiService.api = nil
	<-apiService.connChan
	apiService.Unlock()
	return
}

// IsRunning returns if the service is running
func (apiService *ApierV1Service) IsRunning() bool {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService != nil && apiService.api != nil
}

// ServiceName returns the service name
func (apiService *ApierV1Service) ServiceName() string {
	return utils.ApierV1
}

// GetApierV1 returns the apierV1
func (apiService *ApierV1Service) GetApierV1() *v1.ApierV1 {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.api
}

// ShouldRun returns if the service should be running
func (apiService *ApierV1Service) ShouldRun() bool {
	return apiService.cfg.RalsCfg().Enabled
}

// GetDMChan returns the DataManager chanel
func (apiService *ApierV1Service) GetApierV1Chan() chan *v1.ApierV1 {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.apierV1Chan
}
