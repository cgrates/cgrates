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
	"strings"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cdrs"
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
	cfg    *config.CGRConfig
	dm     *DataDBService
	storDB *StorDBService

	filterSChan chan *engine.FilterS
	server      *cores.Server

	cdrS     *cdrs.CDRServer
	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager

	stopChan chan struct{}
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (cdrS *CDRServer) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if cdrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, cdrS.filterSChan); err != nil {
		return
	}

	var datadb *engine.DataManager
	if datadb, err = cdrS.dm.WaitForDM(ctx); err != nil {
		return
	}

	storDBChan := make(chan engine.StorDB, 1)
	cdrS.stopChan = make(chan struct{})
	cdrS.storDB.RegisterSyncChan(storDBChan)

	cdrS.Lock()
	defer cdrS.Unlock()

	cdrS.cdrS = cdrs.NewCDRServer(cdrS.cfg, datadb, filterS, cdrS.connMgr, storDBChan)
	go cdrS.cdrS.ListenAndServe(cdrS.stopChan)
	runtime.Gosched()
	utils.Logger.Info("Registering CDRS RPC service.")
	srv, err := birpc.NewServiceWithMethodsRename(cdrS, utils.CDRsV1, true, func(oldFn string) (newFn string) {
		return strings.TrimPrefix(oldFn, utils.V1Prfx)
	})
	if err != nil {
		return err
	}
	if !cdrS.cfg.DispatcherSCfg().Enabled {
		cdrS.server.RpcRegister(srv)
	}
	cdrS.connChan <- cdrS.anz.GetInternalCodec(srv, utils.CDRServer) // Signal that cdrS is operational
	return
}

// Reload handles the change of config
func (cdrService *CDRServer) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (cdrService *CDRServer) Shutdown() (err error) {
	cdrService.Lock()
	close(cdrService.stopChan)
	cdrService.cdrS = nil
	<-cdrService.connChan
	cdrService.Unlock()
	cdrService.server.RpcUnregisterName(utils.CDRsV1)
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
