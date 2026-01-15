/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

type BiRPCSessionSFuncs struct {
	sessionSOnBiJSONConnect    func(c birpc.ClientConnector) //store OnBiJSONConnect
	sessionSOnBiJSONDisconnect func(c birpc.ClientConnector) //store OnBiJSONDisconnect
}

// Returns a new BiRPCSessionSFuncs struct
func NewSessionSBiJSONFuncs(onConn, onDisconn func(c birpc.ClientConnector)) *BiRPCSessionSFuncs {
	return &BiRPCSessionSFuncs{
		sessionSOnBiJSONConnect:    onConn,
		sessionSOnBiJSONDisconnect: onDisconn,
	}
}

// GetSessionSOnBiJSONFuncs returns sessionSOnBiJSONConnect and sessionSOnBiJSONDisconnect (1 time use)
func (smg *SessionService) GetSessionSOnBiJSONFuncs() (onConn, onDisconn func(c birpc.ClientConnector)) {
	<-smg.biRPCFuncsDone // Make sure funcs are initialized in the structure before getting them
	return smg.biRPCFuncs.sessionSOnBiJSONConnect, smg.biRPCFuncs.sessionSOnBiJSONDisconnect
}

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *SessionService {
	return &SessionService{
		connChan:       internalChan,
		cfg:            cfg,
		dm:             dm,
		server:         server,
		connMgr:        connMgr,
		anz:            anz,
		srvDep:         srvDep,
		biRPCFuncsDone: make(chan struct{}),
	}
}

// SessionService implements Service interface
type SessionService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	dm       *DataDBService
	server   *cores.Server
	stopChan chan struct{}

	sm       *sessions.SessionS
	connChan chan birpc.ClientConnector

	// in order to stop the bircp server if necesary
	birpcEnabled   bool
	biRPCFuncs     *BiRPCSessionSFuncs
	biRPCFuncsDone chan struct{} // marks when biRPCFuncs are initialized
	connMgr        *engine.ConnManager
	anz            *AnalyzerService
	srvDep         map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (smg *SessionService) Start() error {
	if smg.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	smg.srvDep[utils.DataDB].Add(1) // DataDB will wait for session service to close before closing
	var datadb *engine.DataManager
	if smg.dm.ShouldRun() {
		dbchan := smg.dm.GetDMChan()
		datadb = <-dbchan
		dbchan <- datadb
	}
	smg.Lock()
	defer smg.Unlock()

	smg.sm = sessions.NewSessionS(smg.cfg, datadb, smg.connMgr)
	smg.stopChan = make(chan struct{})

	// Restore previuos sessions backup and start backup looping
	if smg.cfg.SessionSCfg().BackupInterval != 0 {
		if err := smg.sm.RestoreAndBackupSessions(smg.stopChan); err != nil {
			return err
		}
	}

	//start sync session in a separate gorutine
	go smg.sm.SyncSessions(smg.stopChan)
	// Pass internal connection
	srv, err := engine.NewService(v1.NewSessionSv1(smg.sm))
	if err != nil {
		return err
	}
	smg.connChan <- smg.anz.GetInternalCodec(srv, utils.SessionS)
	smg.sm.PopulateCtx(context.WithClient(context.TODO(), srv))
	if !smg.cfg.DispatcherSCfg().Enabled {
		smg.server.RpcRegister(srv)

		// maintain backwards compatibility
		legacySrv, err := engine.NewService(v1.NewSMGenericV1(smg.sm))
		if err != nil {
			return err
		}
		smg.server.RpcRegister(legacySrv)
	}
	// Register BiRpc handlers
	if smg.cfg.ListenCfg().BiJSONListen != utils.EmptyString {
		smg.birpcEnabled = true
		smg.server.BiRPCRegisterName(utils.SessionSv1, srv)
		smg.biRPCFuncs = NewSessionSBiJSONFuncs(smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect)
		smg.biRPCFuncsDone <- struct{}{}
	}
	return nil
}

func (smg *SessionService) DisableBiRPC() {
	if smg == nil {
		return
	}
	smg.Lock()
	smg.birpcEnabled = false
	smg.Unlock()
}

// Reload handles the change of config
func (smg *SessionService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown() (err error) {
	defer smg.srvDep[utils.DataDB].Done() // signal DataDB when session service finishes shutting down
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
	<-smg.connChan
	return
}

// IsRunning returns if the service is running
func (smg *SessionService) IsRunning() bool {
	smg.RLock()
	defer smg.RUnlock()
	return smg.sm != nil
}

// ServiceName returns the service name
func (smg *SessionService) ServiceName() string {
	return utils.SessionS
}

// ShouldRun returns if the service should be running
func (smg *SessionService) ShouldRun() bool {
	return smg.cfg.SessionSCfg().Enabled
}
