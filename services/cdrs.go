// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	srvV1, err := engine.NewService(cdrsV1)
	if err != nil {
		return err
	}
	srvV2, err := engine.NewService(&v2.CDRsV2{CDRsV1: *cdrsV1})
	if err != nil {
		return err
	}
	if !cdrService.cfg.DispatcherSCfg().Enabled {
		cdrService.server.RpcRegister(srvV1)
		cdrService.server.RpcRegister(srvV2)
	}
	intSrv := engine.IntService{
		utils.CDRsV1: srvV1,
		utils.CDRsV2: srvV2,
	}
	cdrService.connChan <- cdrService.anz.GetInternalCodec(intSrv, utils.CDRServer) // Signal that cdrS is operational
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
	return cdrService.cdrS != nil
}

// ServiceName returns the service name
func (cdrService *CDRServer) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cdrService *CDRServer) ShouldRun() bool {
	return cdrService.cfg.CdrsCfg().Enabled
}
