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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewCDRServer returns the CDR Server
func NewCDRServer() servmanager.Service {
	return &CDRServer{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// CDRServer implements Service interface
type CDRServer struct {
	sync.RWMutex
	cdrS     *engine.CDRServer
	rpcv1    *v1.CDRsV1
	rpcv2    *v2.CDRsV2
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (cdrS *CDRServer) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if cdrS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	var ralConn, attrSConn, thresholdSConn, statsConn, chargerSConn rpcclient.RpcClientConnection

	chargerSConn, err = sp.GetConnection(utils.ChargerS, sp.GetConfig().CdrsCfg().CDRSChargerSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ChargerS, err.Error()))
		return
	}
	ralConn, err = sp.GetConnection(utils.ResponderS, sp.GetConfig().CdrsCfg().CDRSRaterConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.RALService, err.Error()))
		return
	}
	attrSConn, err = sp.GetConnection(utils.AttributeS, sp.GetConfig().CdrsCfg().CDRSAttributeSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.AttributeS, err.Error()))
		return
	}
	thresholdSConn, err = sp.GetConnection(utils.ThresholdS, sp.GetConfig().CdrsCfg().CDRSThresholdSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.ThresholdS, err.Error()))
		return
	}
	statsConn, err = sp.GetConnection(utils.StatS, sp.GetConfig().CdrsCfg().CDRSStatSConns)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to %s: %s",
			utils.StatS, err.Error()))
		return
	}
	cdrS.Lock()
	defer cdrS.Unlock()
	cdrS.cdrS = engine.NewCDRServer(sp.GetConfig(), sp.GetCDRStorage(), sp.GetDM(),
		ralConn, attrSConn,
		thresholdSConn, statsConn, chargerSConn, sp.GetFilterS())
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrS.cdrS.RegisterHandlersToServer(sp.GetServer())
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrS.rpcv1 = v1.NewCDRsV1(cdrS.cdrS)
	cdrS.rpcv2 = &v2.CDRsV2{CDRsV1: *cdrS.rpcv1}
	sp.GetServer().RpcRegister(cdrS.rpcv1)
	sp.GetServer().RpcRegister(cdrS.rpcv2)
	// Make the cdr server available for internal communication
	sp.GetServer().RpcRegister(cdrS.cdrS) // register CdrServer for internal usage (TODO: refactor this)
	cdrS.connChan <- cdrS.cdrS            // Signal that cdrS is operational
	return
}

// GetIntenternalChan returns the internal connection chanel
func (cdrS *CDRServer) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return cdrS.connChan
}

// Reload handles the change of config
func (cdrS *CDRServer) Reload(sp servmanager.ServiceProvider) (err error) {
	return
}

// Shutdown stops the service
func (cdrS *CDRServer) Shutdown() (err error) {
	cdrS.cdrS = nil
	cdrS.rpcv1 = nil
	cdrS.rpcv2 = nil
	<-cdrS.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (cdrS *CDRServer) GetRPCInterface() interface{} {
	return cdrS.cdrS
}

// IsRunning returns if the service is running
func (cdrS *CDRServer) IsRunning() bool {
	return cdrS != nil && cdrS.cdrS != nil
}

// ServiceName returns the service name
func (cdrS *CDRServer) ServiceName() string {
	return utils.CDRServer
}
