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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService,
	server *utils.Server, chrsChan, respChan, resChan, thsChan, stsChan,
	supChan, attrsChan, cdrsChan, dispatcherChan chan rpcclient.RpcClientConnection,
	exitChan chan bool) servmanager.Service {
	return &SessionService{
		connChan:       make(chan rpcclient.RpcClientConnection, 1),
		cfg:            cfg,
		dm:             dm,
		server:         server,
		chrsChan:       chrsChan,
		respChan:       respChan,
		resChan:        resChan,
		thsChan:        thsChan,
		stsChan:        stsChan,
		supChan:        supChan,
		attrsChan:      attrsChan,
		cdrsChan:       cdrsChan,
		dispatcherChan: dispatcherChan,
		exitChan:       exitChan,
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex
	cfg            *config.CGRConfig
	dm             *DataDBService
	server         *utils.Server
	chrsChan       chan rpcclient.RpcClientConnection
	respChan       chan rpcclient.RpcClientConnection
	resChan        chan rpcclient.RpcClientConnection
	thsChan        chan rpcclient.RpcClientConnection
	stsChan        chan rpcclient.RpcClientConnection
	supChan        chan rpcclient.RpcClientConnection
	attrsChan      chan rpcclient.RpcClientConnection
	cdrsChan       chan rpcclient.RpcClientConnection
	dispatcherChan chan rpcclient.RpcClientConnection
	exitChan       chan bool

	sm       *sessions.SessionS
	rpc      *v1.SMGenericV1
	rpcv1    *v1.SessionSv1
	connChan chan rpcclient.RpcClientConnection

	// in order to stop the bircp server if necesary
	bircpEnabled bool
}

// Start should handle the sercive start
func (smg *SessionService) Start() (err error) {
	if smg.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	smg.Lock()
	defer smg.Unlock()
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrConns, cdrsConn, chargerSConn rpcclient.RpcClientConnection

	if chargerSConn, err = NewConnection(smg.cfg, smg.chrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ChargerSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ChargerS, err.Error()))
		return
	}

	if ralsConns, err = NewConnection(smg.cfg, smg.respChan, smg.dispatcherChan, smg.cfg.SessionSCfg().RALsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResponderS, err.Error()))
		return
	}

	if resSConns, err = NewConnection(smg.cfg, smg.resChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ResSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResourceS, err.Error()))
		return
	}

	if threshSConns, err = NewConnection(smg.cfg, smg.thsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ThreshSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ThresholdS, err.Error()))
		return
	}

	if statSConns, err = NewConnection(smg.cfg, smg.stsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().StatSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.StatS, err.Error()))
		return
	}

	if suplSConns, err = NewConnection(smg.cfg, smg.supChan, smg.dispatcherChan, smg.cfg.SessionSCfg().SupplSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.SupplierS, err.Error()))
		return
	}

	if attrConns, err = NewConnection(smg.cfg, smg.attrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().AttrSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.AttributeS, err.Error()))
		return
	}

	if cdrsConn, err = NewConnection(smg.cfg, smg.cdrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().CDRsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.CDRServer, err.Error()))
		return
	}

	sReplConns, err := sessions.NewSReplConns(smg.cfg.SessionSCfg().ReplicationConns,
		smg.cfg.GeneralCfg().Reconnects, smg.cfg.GeneralCfg().ConnectTimeout,
		smg.cfg.GeneralCfg().ReplyTimeout)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to SMGReplicationConnection error: <%s>",
			utils.SessionS, err.Error()))
		return
	}

	smg.sm = sessions.NewSessionS(smg.cfg, ralsConns, resSConns, threshSConns,
		statSConns, suplSConns, attrConns, cdrsConn, chargerSConn,
		sReplConns, smg.dm.GetDM(), smg.cfg.GeneralCfg().DefaultTimezone)
	//start sync session in a separate gorutine
	go func(sm *sessions.SessionS) {
		if err = sm.ListenAndServe(smg.exitChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.SessionS, err))
		}
	}(smg.sm)
	// Pass internal connection via BiRPCClient
	smg.connChan <- smg.sm
	// Register RPC handler
	smg.rpc = v1.NewSMGenericV1(smg.sm)
	smg.server.RpcRegister(smg.rpc)

	smg.rpcv1 = v1.NewSessionSv1(smg.sm) // methods with multiple options
	if !smg.cfg.DispatcherSCfg().Enabled {
		smg.server.RpcRegister(smg.rpcv1)
	}
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBijson != "" {
		smg.bircpEnabled = true
		for method, handler := range smg.rpc.Handlers() {
			smg.server.BiRPCRegisterName(method, handler)
		}
		for method, handler := range smg.rpcv1.Handlers() {
			smg.server.BiRPCRegisterName(method, handler)
		}
		// run this in it's own gorutine
		go func() {
			if err := smg.server.ServeBiJSON(smg.cfg.SessionSCfg().ListenBijson, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
				smg.Lock()
				smg.bircpEnabled = false
				smg.Unlock()
			}
		}()
	}
	return
}

// GetIntenternalChan returns the internal connection chanel
func (smg *SessionService) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return smg.connChan
}

// Reload handles the change of config
func (smg *SessionService) Reload() (err error) {
	var ralsConns, resSConns, threshSConns, statSConns, suplSConns, attrConns, cdrsConn, chargerSConn rpcclient.RpcClientConnection

	if chargerSConn, err = NewConnection(smg.cfg, smg.chrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ChargerSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ChargerS, err.Error()))
		return
	}

	if ralsConns, err = NewConnection(smg.cfg, smg.respChan, smg.dispatcherChan, smg.cfg.SessionSCfg().RALsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResponderS, err.Error()))
		return
	}

	if resSConns, err = NewConnection(smg.cfg, smg.resChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ResSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ResourceS, err.Error()))
		return
	}

	if threshSConns, err = NewConnection(smg.cfg, smg.thsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().ThreshSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.ThresholdS, err.Error()))
		return
	}

	if statSConns, err = NewConnection(smg.cfg, smg.stsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().StatSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.StatS, err.Error()))
		return
	}

	if suplSConns, err = NewConnection(smg.cfg, smg.supChan, smg.dispatcherChan, smg.cfg.SessionSCfg().SupplSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.SupplierS, err.Error()))
		return
	}

	if attrConns, err = NewConnection(smg.cfg, smg.attrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().AttrSConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.AttributeS, err.Error()))
		return
	}

	if cdrsConn, err = NewConnection(smg.cfg, smg.cdrsChan, smg.dispatcherChan, smg.cfg.SessionSCfg().CDRsConns); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
			utils.SessionS, utils.CDRServer, err.Error()))
		return
	}

	sReplConns, err := sessions.NewSReplConns(smg.cfg.SessionSCfg().ReplicationConns,
		smg.cfg.GeneralCfg().Reconnects, smg.cfg.GeneralCfg().ConnectTimeout,
		smg.cfg.GeneralCfg().ReplyTimeout)
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
	if smg.bircpEnabled {
		smg.server.StopBiRPC()
		smg.bircpEnabled = false
	}
	smg.sm = nil
	smg.rpc = nil
	smg.rpcv1 = nil
	<-smg.connChan
	return
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

// ShouldRun returns if the service should be running
func (smg *SessionService) ShouldRun() bool {
	return smg.cfg.SessionSCfg().Enabled
}
