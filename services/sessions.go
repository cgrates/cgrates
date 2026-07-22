// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig, dm *DataDBService,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) *SessionService {
	return &SessionService{
		connChan: internalChan,
		cfg:      cfg,
		dm:       dm,
		server:   server,
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
	stopChan chan struct{}

	sm       *sessions.SessionS
	connChan chan birpc.ClientConnector

	connMgr *engine.ConnManager
	anz     *AnalyzerService
	srvDep  map[string]*sync.WaitGroup
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

	// Pass internal connection
	srv, err := engine.NewService(v1.NewSessionSv1(smg.sm))
	if err != nil {
		return err
	}
	smg.sm.PopulateCtx(context.WithClient(context.TODO(), srv))
	// Restore previuos sessions backup and start backup looping
	if smg.cfg.SessionSCfg().BackupInterval != 0 {
		if err := smg.sm.RestoreAndBackupSessions(smg.stopChan); err != nil {
			return err
		}
	}

	//start sync session in a separate gorutine
	go smg.sm.SyncSessions(smg.stopChan)
	smg.connChan <- smg.anz.GetInternalCodec(srv, utils.SessionS)
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
	if smg.cfg.ListenCfg().BiJSONListen != "" || smg.cfg.ListenCfg().BiGobListen != "" {
		smg.server.BiRPCRegisterName(utils.SessionSv1, srv)
	}
	return nil
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
		return err
	}
	if smg.cfg.ListenCfg().BiJSONListen != "" || smg.cfg.ListenCfg().BiGobListen != "" {
		_ = smg.server.BiRPCUnregisterName(utils.SessionSv1)
	}
	smg.sm = nil
	<-smg.connChan
	return nil
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
