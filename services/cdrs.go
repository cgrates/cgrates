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
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *utils.Server, internalCDRServerChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager) servmanager.Service {
	return &CDRServer{
		connChan:    internalCDRServerChan,
		cfg:         cfg,
		dm:          dm,
		storDB:      storDB,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
	}
}

// CDRServer implements Service interface
type CDRServer struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	storDB      *StorDBService
	filterSChan chan *engine.FilterS
	server      *utils.Server

	cdrS     *engine.CDRServer
	rpcv1    *v1.CDRsV1
	rpcv2    *v2.CDRsV2
	connChan chan rpcclient.ClientConnector
	connMgr  *engine.ConnManager
}

// Start should handle the sercive start
func (cdrS *CDRServer) Start() (err error) {
	if cdrS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	filterS := <-cdrS.filterSChan
	cdrS.filterSChan <- filterS

	cdrS.Lock()
	defer cdrS.Unlock()
	cdrS.cdrS = engine.NewCDRServer(cdrS.cfg, cdrS.storDB.GetDM(), cdrS.dm.GetDM(),
		filterS, cdrS.connMgr)
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrS.cdrS.RegisterHandlersToServer(cdrS.server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrS.rpcv1 = v1.NewCDRsV1(cdrS.cdrS)
	cdrS.rpcv2 = &v2.CDRsV2{CDRsV1: *cdrS.rpcv1}
	cdrS.server.RpcRegister(cdrS.rpcv1)
	cdrS.server.RpcRegister(cdrS.rpcv2)
	// Make the cdr server available for internal communication
	cdrS.server.RpcRegister(cdrS.cdrS) // register CdrServer for internal usage (TODO: refactor this)
	cdrS.connChan <- cdrS.cdrS         // Signal that cdrS is operational
	return
}

// GetIntenternalChan returns the internal connection chanel
func (cdrS *CDRServer) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return cdrS.connChan
}

// Reload handles the change of config
func (cdrS *CDRServer) Reload() (err error) {

	cdrS.Lock()
	if cdrS.storDB.WasReconnected() { // rewrite the connection if was changed
		cdrS.cdrS.SetStorDB(cdrS.storDB.GetDM())
	}
	cdrS.Unlock()
	return
}

// Shutdown stops the service
func (cdrS *CDRServer) Shutdown() (err error) {
	cdrS.Lock()
	cdrS.cdrS = nil
	cdrS.rpcv1 = nil
	cdrS.rpcv2 = nil
	<-cdrS.connChan
	cdrS.Unlock()
	return
}

// IsRunning returns if the service is running
func (cdrS *CDRServer) IsRunning() bool {
	cdrS.RLock()
	defer cdrS.RUnlock()
	return cdrS != nil && cdrS.cdrS != nil
}

// ServiceName returns the service name
func (cdrS *CDRServer) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cdrS *CDRServer) ShouldRun() bool {
	return cdrS.cfg.CdrsCfg().Enabled
}
