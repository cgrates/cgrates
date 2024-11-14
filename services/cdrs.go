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
	clSChan chan *commonlisteners.CommonListenerS, internalCDRServerChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anzChan chan *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &CDRService{
		connChan:    internalCDRServerChan,
		cfg:         cfg,
		dm:          dm,
		storDB:      storDB,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		anzChan:     anzChan,
		srvDep:      srvDep,
	}
}

// CDRService implements Service interface
type CDRService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	storDB      *StorDBService
	anzChan     chan *AnalyzerService
	filterSChan chan *engine.FilterS

	cdrS *cdrs.CDRServer
	cl   *commonlisteners.CommonListenerS

	connChan chan birpc.ClientConnector
	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (cs *CDRService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if cs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	cs.cl = <-cs.clSChan
	cs.clSChan <- cs.cl
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, cs.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = cs.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-cs.anzChan
	cs.anzChan <- anz

	storDBChan := make(chan engine.StorDB, 1)
	cs.stopChan = make(chan struct{})
	cs.storDB.RegisterSyncChan(storDBChan)

	cs.Lock()
	defer cs.Unlock()

	cs.cdrS = cdrs.NewCDRServer(cs.cfg, datadb, filterS, cs.connMgr, storDBChan)
	go cs.cdrS.ListenAndServe(cs.stopChan)
	runtime.Gosched()
	utils.Logger.Info("Registering CDRS RPC service.")
	srv, err := engine.NewServiceWithPing(cs.cdrS, utils.CDRsV1, utils.V1Prfx)
	if err != nil {
		return err
	}
	if !cs.cfg.DispatcherSCfg().Enabled {
		cs.cl.RpcRegister(srv)
	}
	cs.connChan <- anz.GetInternalCodec(srv, utils.CDRServer) // Signal that cdrS is operational
	return
}

// Reload handles the change of config
func (cs *CDRService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (cs *CDRService) Shutdown() (err error) {
	cs.Lock()
	close(cs.stopChan)
	cs.cdrS = nil
	<-cs.connChan
	cs.Unlock()
	cs.cl.RpcUnregisterName(utils.CDRsV1)
	return
}

// IsRunning returns if the service is running
func (cs *CDRService) IsRunning() bool {
	cs.RLock()
	defer cs.RUnlock()
	return cs.cdrS != nil
}

// ServiceName returns the service name
func (cs *CDRService) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cs *CDRService) ShouldRun() bool {
	return cs.cfg.CdrsCfg().Enabled
}
