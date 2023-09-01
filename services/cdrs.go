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
	"runtime"
	"sync"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *cores.Server, internalCDRServerChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &CDRServer{
		connChan:    internalCDRServerChan,
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

// CDRServer implements Service interface
type CDRServer struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	storDB      *StorDBService
	filterSChan chan *engine.FilterS
	server      *cores.Server

	cdrS     *engine.CDRServer
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager

	stopChan chan struct{}
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (cdrService *CDRServer) Start() error {
	if cdrService.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	filterS := <-cdrService.filterSChan
	cdrService.filterSChan <- filterS
	dbchan := cdrService.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	storDBChan := make(chan engine.StorDB, 1)
	cdrService.stopChan = make(chan struct{})
	cdrService.storDB.RegisterSyncChan(storDBChan)

	cdrService.Lock()
	defer cdrService.Unlock()

	cdrService.cdrS = engine.NewCDRServer(cdrService.cfg, storDBChan, datadb, filterS, cdrService.connMgr)
	go cdrService.cdrS.ListenAndServe(cdrService.stopChan)
	runtime.Gosched()
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrService.cdrS.RegisterHandlersToServer(cdrService.server)
	utils.Logger.Info("Registering CDRS RPC service.")

	cdrsV1 := v1.NewCDRsV1(cdrService.cdrS)
	cdrsV2 := &v2.CDRsV2{CDRsV1: *cdrsV1}
	srv, err := engine.NewService(cdrsV1)
	if err != nil {
		return err
	}
	cdrsV2Srv, err := birpc.NewService(cdrsV2, "", false)
	if err != nil {
		return err
	}
	engine.RegisterPingMethod(cdrsV2Srv.Methods)
	srv[utils.CDRsV2] = cdrsV2Srv
	if !cdrService.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cdrService.server.RpcRegister(s)
		}
	}
	cdrService.connChan <- cdrService.anz.GetInternalCodec(srv, utils.CDRServer) // Signal that cdrS is operational
	return nil
}

// Reload handles the change of config
func (cdrService *CDRServer) Reload() (err error) {
	return
}

// Shutdown stops the service
func (cdrService *CDRServer) Shutdown() (err error) {
	cdrService.Lock()
	close(cdrService.stopChan)
	cdrService.cdrS = nil
	<-cdrService.connChan
	cdrService.Unlock()
	return
}

// IsRunning returns if the service is running
func (cdrService *CDRServer) IsRunning() bool {
	cdrService.RLock()
	defer cdrService.RUnlock()
	return cdrService != nil && cdrService.cdrS != nil
}

// ServiceName returns the service name
func (cdrService *CDRServer) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cdrService *CDRServer) ShouldRun() bool {
	return cdrService.cfg.CdrsCfg().Enabled
}
