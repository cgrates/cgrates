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
	"github.com/cgrates/cgrates/apis"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan, connMgr *engine.ConnManager,
	anz *AnalyzerService, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &SessionService{
		connChan: internalChan,
		cfg:      cfg,
		dm:       dm,
		server:   server,
		shdChan:  shdChan,
		connMgr:  connMgr,
		anz:      anz,
		srvDep:   srvDep,
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	dm       *DataDBService
	server   *cores.Server
	shdChan  *utils.SyncedChan
	stopChan chan struct{}

	sm *sessions.SessionS
	//rpc      *apis.SMGenericV1
	rpcv1    *apis.SessionSv1
	connChan chan birpc.ClientConnector

	// in order to stop the bircp server if necesary
	bircpEnabled bool
	connMgr      *engine.ConnManager
	anz          *AnalyzerService
	srvDep       map[string]*sync.WaitGroup
}

// Start should handle the service start
func (smg *SessionService) Start() (err error) {
	if smg.IsRunning() {
		return utils.ErrServiceAlreadyRunning
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
	//start sync session in a separate goroutine
	smg.stopChan = make(chan struct{})
	go smg.sm.ListenAndServe(smg.stopChan)
	// Pass internal connection via BiRPCClient

	// Register RPC handler
	smg.rpcv1 = apis.NewSessionSv1(smg.sm) // methods with multiple options
	srv, _ := birpc.NewService(smg.rpcv1, utils.EmptyString, false)
	//smg.rpc = v1.NewSMGenericV1(smg.sm, smg.caps)
	smg.rpcv1 = apis.NewSessionSv1(smg.sm) // methods with multiple options
	if !smg.cfg.DispatcherSCfg().Enabled {
		smg.server.RpcRegister(srv)
	}
	smg.connChan <- smg.anz.GetInternalCodec(srv, utils.SessionS)
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBijson != utils.EmptyString {
		smg.bircpEnabled = true
		// smg.server.BiRPCRegisterName("SMGenericV1", smg.rpc)
		smg.server.BiRPCRegisterName(utils.SessionSv1, smg.rpcv1)
		// run this in it's own goroutine
		go smg.start()
	}
	return
}

func (smg *SessionService) start() (err error) {
	if err := smg.server.ServeBiRPC(smg.cfg.SessionSCfg().ListenBijson,
		smg.cfg.SessionSCfg().ListenBigob, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
		smg.Lock()
		smg.bircpEnabled = false
		smg.Unlock()
		smg.shdChan.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (smg *SessionService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown() (err error) {
	smg.Lock()
	defer smg.Unlock()
	close(smg.stopChan)
	if err = smg.sm.Shutdown(); err != nil {
		return
	}
	if smg.bircpEnabled {
		smg.server.StopBiRPC()
		smg.bircpEnabled = false
	}
	smg.sm = nil
	// smg.rpc = nil
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
