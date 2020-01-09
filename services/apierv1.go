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

	syncStop   chan struct{}
	storDBChan chan engine.StorDB
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (api *ApierV1Service) Start() (err error) {
	if api.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	filterS := <-api.filterSChan
	api.filterSChan <- filterS
	dbchan := api.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	api.Lock()
	defer api.Unlock()

	api.storDBChan = make(chan engine.StorDB, 1)
	api.syncStop = make(chan struct{})
	api.storDB.RegisterSyncChan(api.storDBChan)
	stordb := <-api.storDBChan

	api.api = &v1.ApierV1{
		DataManager:      datadb,
		CdrDb:            stordb,
		StorDb:           stordb,
		Config:           api.cfg,
		Responder:        api.responderService.GetResponder(),
		SchedulerService: api.schedService,
		HTTPPoster: engine.NewHTTPPoster(api.cfg.GeneralCfg().HttpSkipTlsVerify,
			api.cfg.GeneralCfg().ReplyTimeout),
		FilterS: filterS,
		ConnMgr: api.connMgr,
	}

	if !api.cfg.DispatcherSCfg().Enabled {
		api.server.RpcRegister(api.api)
		api.server.RpcRegister(v1.NewReplicatorSv1(datadb))
	}

	utils.RegisterRpcParams("", &v1.CDRsV1{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", api.api)

	api.connChan <- api.api
	go api.sync()

	return
}

// GetIntenternalChan returns the internal connection chanel
func (api *ApierV1Service) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return api.connChan
}

// Reload handles the change of config
func (api *ApierV1Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (api *ApierV1Service) Shutdown() (err error) {
	api.Lock()
	close(api.syncStop)
	api.api = nil
	<-api.connChan
	api.Unlock()
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

// sync handles stordb sync
func (api *ApierV1Service) sync() {
	for {
		select {
		case <-api.syncStop:
			return
		case stordb, ok := <-api.storDBChan:
			if !ok { // the chanel was closed by the shutdown of stordbService
				return
			}
			api.Lock()
			if api.api != nil {
				api.api.SetStorDB(stordb)
			}
			api.Unlock()
		}
	}
}
