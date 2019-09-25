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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSessionService returns the Session Service
func NewSessionService() servmanager.Service {
	return &SessionService{
		connChan: make(chan rpcclient.RpcClientConnection, 1),
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex
	sm       *sessions.SessionS
	rpc      *v1.SMGenericV1
	rpcv1    *v1.SessionSv1
	connChan chan rpcclient.RpcClientConnection
}

// Start should handle the sercive start
func (smg *SessionService) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if smg.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	smg.Lock()
	defer smg.Unlock()
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrConns, cdrsConn, chargerSConn rpcclient.RpcClientConnection

	if chargerSConn, err = sp.GetConnection(utils.ChargerS, sp.GetConfig().SessionSCfg().ChargerSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ChargerS, err.Error()))
		return
	}

	if ralsConns, err = sp.GetConnection(utils.ResponderS, sp.GetConfig().SessionSCfg().RALsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResponderS, err.Error()))
		return
	}

	if resSConns, err = sp.GetConnection(utils.ResourceS, sp.GetConfig().SessionSCfg().ResSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResourceS, err.Error()))
		return
	}

	if threshSConns, err = sp.GetConnection(utils.ThresholdS, sp.GetConfig().SessionSCfg().ThreshSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ThresholdS, err.Error()))
		return
	}

	if statSConns, err = sp.GetConnection(utils.StatS, sp.GetConfig().SessionSCfg().StatSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.StatS, err.Error()))
		return
	}

	if suplSConns, err = sp.GetConnection(utils.SupplierS, sp.GetConfig().SessionSCfg().SupplSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.SupplierS, err.Error()))
		return
	}

	if attrConns, err = sp.GetConnection(utils.AttributeS, sp.GetConfig().SessionSCfg().AttrSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.AttributeS, err.Error()))
		return
	}

	if cdrsConn, err = sp.GetConnection(utils.CDRServer, sp.GetConfig().SessionSCfg().CDRsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.CDRServer, err.Error()))
		return
	}

	sReplConns, err := sessions.NewSReplConns(sp.GetConfig().SessionSCfg().SessionReplicationConns,
		sp.GetConfig().GeneralCfg().Reconnects, sp.GetConfig().GeneralCfg().ConnectTimeout,
		sp.GetConfig().GeneralCfg().ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>",
			utils.SessionS, err.Error()))
		return
	}

	smg.sm = sessions.NewSessionS(sp.GetConfig(), ralsConns, resSConns, threshSConns,
		statSConns, suplSConns, attrConns, cdrsConn, chargerSConn,
		sReplConns, sp.GetDM(), sp.GetConfig().GeneralCfg().DefaultTimezone)
	//start sync session in a separate gorutine
	go func(sm *sessions.SessionS) {
		if err = sm.ListenAndServe(sp.GetExitChan()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.SessionS, err))
		}
	}(smg.sm)
	// Pass internal connection via BiRPCClient
	smg.connChan <- smg.sm
	// Register RPC handler
	smg.rpc = v1.NewSMGenericV1(smg.sm)
	sp.GetServer().RpcRegister(smg.rpc)

	smg.rpcv1 = v1.NewSessionSv1(smg.sm) // methods with multiple options
	if !sp.GetConfig().DispatcherSCfg().Enabled {
		sp.GetServer().RpcRegister(smg.rpcv1)
	}
	// Register BiRpc handlers
	if sp.GetConfig().SessionSCfg().ListenBijson != "" {
		for method, handler := range smg.rpc.Handlers() {
			sp.GetServer().BiRPCRegisterName(method, handler)
		}
		for method, handler := range smg.rpcv1.Handlers() {
			sp.GetServer().BiRPCRegisterName(method, handler)
		}
		// run this in it's own gorutine
		go sp.GetServer().ServeBiJSON(sp.GetConfig().SessionSCfg().ListenBijson, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect)
	}
	return
}

// GetIntenternalChan returns the internal connection chanel
func (smg *SessionService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return smg.connChan
}

// Reload handles the change of config
func (smg *SessionService) Reload(sp servmanager.ServiceProvider) (err error) {
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrConns, cdrsConn, chargerSConn rpcclient.RpcClientConnection

	if chargerSConn, err = sp.GetConnection(utils.ChargerS, sp.GetConfig().SessionSCfg().ChargerSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ChargerS, err.Error()))
		return
	}

	if ralsConns, err = sp.GetConnection(utils.ResponderS, sp.GetConfig().SessionSCfg().RALsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResponderS, err.Error()))
		return
	}

	if resSConns, err = sp.GetConnection(utils.ResourceS, sp.GetConfig().SessionSCfg().ResSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResourceS, err.Error()))
		return
	}

	if threshSConns, err = sp.GetConnection(utils.ThresholdS, sp.GetConfig().SessionSCfg().ThreshSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ThresholdS, err.Error()))
		return
	}

	if statSConns, err = sp.GetConnection(utils.StatS, sp.GetConfig().SessionSCfg().StatSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.StatS, err.Error()))
		return
	}

	if suplSConns, err = sp.GetConnection(utils.SupplierS, sp.GetConfig().SessionSCfg().SupplSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.SupplierS, err.Error()))
		return
	}

	if attrConns, err = sp.GetConnection(utils.AttributeS, sp.GetConfig().SessionSCfg().AttrSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.AttributeS, err.Error()))
		return
	}

	if cdrsConn, err = sp.GetConnection(utils.CDRServer, sp.GetConfig().SessionSCfg().CDRsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.CDRServer, err.Error()))
		return
	}

	sReplConns, err := sessions.NewSReplConns(sp.GetConfig().SessionSCfg().SessionReplicationConns,
		sp.GetConfig().GeneralCfg().Reconnects, sp.GetConfig().GeneralCfg().ConnectTimeout,
		sp.GetConfig().GeneralCfg().ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>",
			utils.SessionS, err.Error()))
		return
	}
	smg.Lock()
	smg.sm.SetAttributeSConnection(attrConns)
	smg.sm.SetChargerSConnection(chargerSConn)
	smg.sm.SetRALsConnection(ralsConns)
	smg.sm.SetResourceSConnection(resSConns)
	smg.sm.SetThresholSConnection(threshSConns)
	smg.sm.SetStatSConnection(statSConns)
	smg.sm.SetSupplierSConnection(suplSConns)
	smg.sm.SetCDRSConnection(cdrsConn)
	smg.sm.SetReplicationConnections(sReplConns)
	smg.Unlock()
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown() (err error) {
	smg.Lock()
	defer smg.Unlock()
	if err = smg.sm.Shutdown(); err != nil {
		return
	}
	smg.sm = nil
	smg.rpc = nil
	smg.rpcv1 = nil
	<-smg.connChan
	return
}

// GetRPCInterface returns the interface to register for server
func (smg *SessionService) GetRPCInterface() interface{} {
	return smg.sm
}

// IsRunning returns if the service is running
func (smg *SessionService) IsRunning() bool {
	smg.RLock()
	defer smg.RUnlock()
	return smg != nil && smg.sm != nil
}

// ServiceName returns the service name
func (smg *SessionService) ServiceName() string {
	return utils.SessionS
}
