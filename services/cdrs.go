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
	server *utils.Server, internalCDRServerChan, chrsChan, respChan, attrsChan, thsChan, stsChan,
	dispatcherChan chan rpcclient.RpcClientConnection) servmanager.Service {
	return &CDRServer{
		connChan:       internalCDRServerChan,
		cfg:            cfg,
		dm:             dm,
		storDB:         storDB,
		filterSChan:    filterSChan,
		server:         server,
		chrsChan:       chrsChan,
		respChan:       respChan,
		attrsChan:      attrsChan,
		thsChan:        thsChan,
		stsChan:        stsChan,
		dispatcherChan: dispatcherChan,
	}
}

// CDRServer implements Service interface
type CDRServer struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	dm             *DataDBService
	storDB         *StorDBService
	filterSChan    chan *engine.FilterS
	server         *utils.Server
	chrsChan       chan rpcclient.RpcClientConnection
	respChan       chan rpcclient.RpcClientConnection
	attrsChan      chan rpcclient.RpcClientConnection
	thsChan        chan rpcclient.RpcClientConnection
	stsChan        chan rpcclient.RpcClientConnection
	dispatcherChan chan rpcclient.RpcClientConnection

	cdrS     *engine.CDRServer
	rpcv1    *v1.CDRsV1
	rpcv2    *v2.CDRsV2
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (cdrS *CDRServer) Start() (err error) {
	if cdrS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	filterS := <-cdrS.filterSChan
	cdrS.filterSChan <- filterS

	var ralConn, attrSConn, thresholdSConn, statsConn, chargerSConn rpcclient.RpcClientConnection

	chargerSConn, err = NewConnection(cdrS.cfg, cdrS.chrsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().ChargerSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ChargerS, err.Error()))
		return
	}
	ralConn, err = NewConnection(cdrS.cfg, cdrS.respChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().RaterConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.RALService, err.Error()))
		return
	}
	attrSConn, err = NewConnection(cdrS.cfg, cdrS.attrsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().AttributeSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.AttributeS, err.Error()))
		return
	}
	thresholdSConn, err = NewConnection(cdrS.cfg, cdrS.thsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().ThresholdSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ThresholdS, err.Error()))
		return
	}
	statsConn, err = NewConnection(cdrS.cfg, cdrS.stsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().StatSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.StatS, err.Error()))
		return
	}
	cdrS.Lock()
	defer cdrS.Unlock()
	cdrS.cdrS = engine.NewCDRServer(cdrS.cfg, cdrS.storDB.GetDM(), cdrS.dm.GetDM(),
		ralConn, attrSConn, thresholdSConn, statsConn, chargerSConn, filterS)
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
func (cdrS *CDRServer) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return cdrS.connChan
}

// Reload handles the change of config
func (cdrS *CDRServer) Reload() (err error) {
	var ralConn, attrSConn, thresholdSConn, statsConn, chargerSConn rpcclient.RpcClientConnection

	chargerSConn, err = NewConnection(cdrS.cfg, cdrS.chrsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().ChargerSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ChargerS, err.Error()))
		return
	}
	ralConn, err = NewConnection(cdrS.cfg, cdrS.respChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().RaterConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.RALService, err.Error()))
		return
	}
	attrSConn, err = NewConnection(cdrS.cfg, cdrS.attrsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().AttributeSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.AttributeS, err.Error()))
		return
	}
	thresholdSConn, err = NewConnection(cdrS.cfg, cdrS.thsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().ThresholdSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ThresholdS, err.Error()))
		return
	}
	statsConn, err = NewConnection(cdrS.cfg, cdrS.stsChan, cdrS.dispatcherChan, cdrS.cfg.CdrsCfg().StatSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.StatS, err.Error()))
		return
	}
	cdrS.Lock()
	if cdrS.storDB.WasReconnected() { // rewrite the connection if was changed
		cdrS.cdrS.SetStorDB(cdrS.storDB.GetDM())
	}
	cdrS.cdrS.SetRALsConnection(ralConn)
	cdrS.cdrS.SetAttributeSConnection(attrSConn)
	cdrS.cdrS.SetThresholSConnection(thresholdSConn)
	cdrS.cdrS.SetStatSConnection(statsConn)
	cdrS.cdrS.SetChargerSConnection(chargerSConn)
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
