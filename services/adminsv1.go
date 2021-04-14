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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAPIerSv1Service returns the APIerSv1 Service
func NewAdminSv1Service(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *cores.Server,
	internalAPIerSv1Chan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &AdminSv1Service{
		connChan:    internalAPIerSv1Chan,
		cfg:         cfg,
		dm:          dm,
		storDB:      storDB,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// APIerSv1Service implements Service interface
type AdminSv1Service struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	storDB      *StorDBService
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	api      *apis.AdminSv1
	connChan chan birpc.ClientConnector

	stopChan chan struct{}

	anz    *AnalyzerService
	srvDep map[string]*sync.WaitGroup
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (apiService *AdminSv1Service) Start() (err error) {
	if apiService.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	// filterS := <-apiService.filterSChan
	// apiService.filterSChan <- filterS
	dbchan := apiService.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	// apiService.stopChan = make(chan struct{})
	// storDBChan := make(chan engine.StorDB, 1)
	// apiService.storDB.RegisterSyncChan(storDBChan)
	// stordb := <-storDBChan

	apiService.Lock()
	defer apiService.Unlock()

	apiService.api = apis.NewAdminSv1(apiService.cfg, datadb, apiService.connMgr)

	// go apiService.api.ListenAndServe(apiService.stopChan)
	// runtime.Gosched()
	srv, _ := birpc.NewService(apiService.api, "", false)

	if !apiService.cfg.DispatcherSCfg().Enabled {
		apiService.server.RpcRegister(srv)
	}

	//backwards compatible
	apiService.connChan <- apiService.anz.GetInternalCodec(srv, srv.Name)

	return
}

// Reload handles the change of config
func (apiService *AdminSv1Service) Reload() (err error) {
	return
}

// Shutdown stops the service
func (apiService *AdminSv1Service) Shutdown() (err error) {
	apiService.Lock()
	// close(apiService.stopChan)
	apiService.api = nil
	<-apiService.connChan
	apiService.server.RpcUnregisterName(utils.AdminSv1)
	apiService.Unlock()
	return
}

// IsRunning returns if the service is running
func (apiService *AdminSv1Service) IsRunning() bool {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.api != nil
}

// ServiceName returns the service name
func (apiService *AdminSv1Service) ServiceName() string {
	return utils.AdminS
}

// ShouldRun returns if the service should be running
func (apiService *AdminSv1Service) ShouldRun() bool {
	return apiService.cfg.AdminSCfg().Enabled
}
