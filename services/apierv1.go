// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"runtime"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAPIerSv1Service returns the APIerSv1 Service
func NewAPIerSv1Service(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *cores.Server,
	schedService *SchedulerService,
	responderService *ResponderService,
	internalAPIerSv1Chan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *APIerSv1Service {
	return &APIerSv1Service{
		connChan:         internalAPIerSv1Chan,
		cfg:              cfg,
		dm:               dm,
		storDB:           storDB,
		filterSChan:      filterSChan,
		server:           server,
		schedService:     schedService,
		responderService: responderService,
		connMgr:          connMgr,
		APIerSv1Chan:     make(chan *v1.APIerSv1, 1),
		anz:              anz,
		srvDep:           srvDep,
	}
}

// APIerSv1Service implements Service interface
type APIerSv1Service struct {
	sync.RWMutex
	cfg              *config.CGRConfig
	dm               *DataDBService
	storDB           *StorDBService
	filterSChan      chan *engine.FilterS
	server           *cores.Server
	schedService     *SchedulerService
	responderService *ResponderService
	connMgr          *engine.ConnManager

	api      *v1.APIerSv1
	connChan chan birpc.ClientConnector

	stopChan chan struct{}

	APIerSv1Chan chan *v1.APIerSv1
	anz          *AnalyzerService
	srvDep       map[string]*sync.WaitGroup
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (apiService *APIerSv1Service) Start() error {
	if apiService.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-apiService.filterSChan
	apiService.filterSChan <- filterS
	dbchan := apiService.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	storDBChan := make(chan engine.StorDB, 1)
	apiService.stopChan = make(chan struct{})
	apiService.storDB.RegisterSyncChan(storDBChan)
	stordb := <-storDBChan

	respChan := make(chan *engine.Responder, 1)
	apiService.responderService.RegisterSyncChan(apiService.ServiceName(), respChan)
	apiService.Lock()
	defer apiService.Unlock()

	apiService.api = &v1.APIerSv1{
		DataManager:      datadb,
		CdrDb:            stordb,
		StorDb:           stordb,
		Config:           apiService.cfg,
		SchedulerService: apiService.schedService,
		FilterS:          filterS,
		ConnMgr:          apiService.connMgr,
		StorDBChan:       storDBChan,

		Responder:     apiService.responderService.GetResponder(), // if already started use it
		ResponderChan: respChan,                                   // if not wait in listenAndServe
	}

	go apiService.api.ListenAndServe(apiService.stopChan)
	runtime.Gosched()

	srv, err := engine.NewService(apiService.api)
	if err != nil {
		return err
	}
	if !apiService.cfg.DispatcherSCfg().Enabled {
		apiService.server.RpcRegister(srv)

		//backwards compatible
		legacySrv, err := engine.NewServiceWithName(apiService.api, utils.ApierV1, true)
		if err != nil {
			return err
		}
		apiService.server.RpcRegister(legacySrv)

		rplSrv, err := engine.NewService(v1.NewReplicatorSv1(datadb, apiService.api))
		if err != nil {
			return err
		}
		apiService.server.RpcRegister(rplSrv)
	}

	apiService.connChan <- apiService.anz.GetInternalCodec(srv, utils.APIerSv1)

	apiService.APIerSv1Chan <- apiService.api
	return nil
}

// Reload handles the change of config
func (apiService *APIerSv1Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (apiService *APIerSv1Service) Shutdown() (err error) {
	apiService.Lock()
	close(apiService.stopChan)
	apiService.api = nil
	<-apiService.connChan
	apiService.responderService.UnregisterSyncChan(apiService.ServiceName())
	apiService.Unlock()
	return
}

// IsRunning returns if the service is running
func (apiService *APIerSv1Service) IsRunning() bool {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.api != nil
}

// ServiceName returns the service name
func (apiService *APIerSv1Service) ServiceName() string {
	return utils.APIerSv1
}

// GetAPIerSv1 returns the APIerSv1
func (apiService *APIerSv1Service) GetAPIerSv1() *v1.APIerSv1 {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.api
}

// ShouldRun returns if the service should be running
func (apiService *APIerSv1Service) ShouldRun() bool {
	return apiService.cfg.ApierCfg().Enabled
}

// GetAPIerSv1Chan returns the DataManager chanel
func (apiService *APIerSv1Service) GetAPIerSv1Chan() chan *v1.APIerSv1 {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.APIerSv1Chan
}
