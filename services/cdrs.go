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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *commonlisteners.CommonListenerS, internalCDRServerChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &CDRService{
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

// CDRService implements Service interface
type CDRService struct {
	sync.RWMutex
	cfg    *config.CGRConfig
	dm     *DataDBService
	storDB *StorDBService

	filterSChan chan *engine.FilterS
	server      *commonlisteners.CommonListenerS

	cdrS     *cdrs.CDRServer
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager

	stopChan chan struct{}
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (cdrSrv *CDRService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if cdrSrv.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, cdrSrv.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = cdrSrv.dm.WaitForDM(ctx); err != nil {
		return
	}
	if err = cdrSrv.anz.WaitForAnalyzerS(ctx); err != nil {
		return
	}

	storDBChan := make(chan engine.StorDB, 1)
	cdrSrv.stopChan = make(chan struct{})
	cdrSrv.storDB.RegisterSyncChan(storDBChan)

	cdrSrv.Lock()
	defer cdrSrv.Unlock()

	cdrSrv.cdrS = cdrs.NewCDRServer(cdrSrv.cfg, datadb, filterS, cdrSrv.connMgr, storDBChan)
	go cdrSrv.cdrS.ListenAndServe(cdrSrv.stopChan)
	runtime.Gosched()
	utils.Logger.Info("Registering CDRS RPC service.")
	srv, err := engine.NewServiceWithPing(cdrSrv.cdrS, utils.CDRsV1, utils.V1Prfx)
	if err != nil {
		return err
	}
	if !cdrSrv.cfg.DispatcherSCfg().Enabled {
		cdrSrv.server.RpcRegister(srv)
	}
	cdrSrv.connChan <- cdrSrv.anz.GetInternalCodec(srv, utils.CDRServer) // Signal that cdrS is operational
	return
}

// Reload handles the change of config
func (cdrService *CDRService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (cdrService *CDRService) Shutdown() (err error) {
	cdrService.Lock()
	close(cdrService.stopChan)
	cdrService.cdrS = nil
	<-cdrService.connChan
	cdrService.Unlock()
	cdrService.server.RpcUnregisterName(utils.CDRsV1)
	return
}

// IsRunning returns if the service is running
func (cdrService *CDRService) IsRunning() bool {
	cdrService.RLock()
	defer cdrService.RUnlock()
	return cdrService != nil && cdrService.cdrS != nil
}

// ServiceName returns the service name
func (cdrService *CDRService) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cdrService *CDRService) ShouldRun() bool {
	return cdrService.cfg.CdrsCfg().Enabled
}
