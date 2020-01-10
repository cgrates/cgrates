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

	"github.com/cgrates/cgrates/engine"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService,
	server *utils.Server, internalChan chan rpcclient.ClientConnector,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &SessionService{
		connChan: internalChan,
		cfg:      cfg,
		dm:       dm,
		server:   server,
		exitChan: exitChan,
		connMgr:  connMgr,
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	dm       *DataDBService
	server   *utils.Server
	exitChan chan bool

	sm       *sessions.SessionS
	rpc      *v1.SMGenericV1
	rpcv1    *v1.SessionSv1
	connChan chan rpcclient.ClientConnector

	// in order to stop the bircp server if necesary
	bircpEnabled bool
	connMgr      *engine.ConnManager
}

// Start should handle the sercive start
func (smg *SessionService) Start() (err error) {
	if smg.IsRunning() {
		return fmt.Errorf("service aleady running")
	}
	var datadb *engine.DataManager
	if smg.dm.IsRunning() {
		dbchan := smg.dm.GetDMChan()
		datadb = <-dbchan
		dbchan <- datadb
	}
	smg.Lock()
	defer smg.Unlock()

	smg.sm = sessions.NewSessionS(smg.cfg, datadb, smg.connMgr)
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

	smg.rpcv1 = v1.NewSessionSv1(smg.sm) // methods with multiple options
	if !smg.cfg.DispatcherSCfg().Enabled {
		smg.server.RpcRegister(smg.rpc)
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
func (smg *SessionService) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	return smg.connChan
}

// Reload handles the change of config
func (smg *SessionService) Reload() (err error) {
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
