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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"

	v1 "github.com/cgrates/cgrates/apier/v1"
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

	sm       *sessions.SessionS
	rpc      *v1.SMGenericV1
	connChan chan birpc.ClientConnector

	// in order to stop the bircp server if necesary
	birpcEnabled bool
	connMgr      *engine.ConnManager
	anz          *AnalyzerService
	srvDep       map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (smg *SessionService) Start() error {
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
	//start sync session in a separate gorutine
	smg.stopChan = make(chan struct{})
	go smg.sm.ListenAndServe(smg.stopChan)
	// Pass internal connection
	srv, err := engine.NewServiceWithName(v1.NewSessionSv1(smg.sm), "", false)
	if err != nil {
		return err
	}
	smg.connChan <- smg.anz.GetInternalCodec(srv, utils.SessionS)
	// Register RPC handler
	smg.rpc = v1.NewSMGenericV1(smg.sm)

	if !smg.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			smg.server.RpcRegister(s)
		}

		var legacySrv engine.IntService
		legacySrv, err := engine.NewService(smg.rpc)
		if err != nil {
			return err
		}
		for _, s := range legacySrv {
			smg.server.RpcRegister(s)
		}
	}
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBijson != "" {
		smg.birpcEnabled = true
		for n, s := range srv {
			smg.server.BiRPCRegisterName(n, s)
		}
		// run this in it's own goroutine
		go smg.start()
	}
	return nil
}

func (smg *SessionService) start() (err error) {
	if err := smg.server.ServeBiRPC(smg.cfg.SessionSCfg().ListenBijson,
		smg.cfg.SessionSCfg().ListenBigob, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
		smg.Lock()
		smg.birpcEnabled = false
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
	if smg.birpcEnabled {
		smg.server.StopBiRPC()
		smg.birpcEnabled = false
	}
	smg.sm = nil
	smg.rpc = nil
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
