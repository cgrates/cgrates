// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig) *SessionService {
	return &SessionService{
		cfg: cfg,
	}
}

// SessionService implements Service interface
type SessionService struct {
	mu           sync.RWMutex
	sm           *sessions.SessionS
	bircpEnabled bool // to stop birpc server if needed
	stopChan     chan struct{}
	cfg          *config.CGRConfig
}

// Start should handle the service start
func (smg *SessionService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		smg.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()
	cacheS := srvDeps[utils.CacheS].(*CacheService)

	smg.mu.Lock()
	defer smg.mu.Unlock()

	smg.sm = sessions.NewSessionS(smg.cfg, dbs, cacheS.CacheS(), fs, cms.ConnManager())
	//start sync session in a separate goroutine
	smg.stopChan = make(chan struct{})
	go smg.sm.ListenAndServe(smg.stopChan)
	// Pass internal connection via BiRPCClient

	// Register RPC handler
	srv, err := newRPCService(apis.NewSessionSv1(smg.sm), utils.SessionSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBiJSON != utils.EmptyString {
		smg.bircpEnabled = true
		cl.BiRPCRegisterName(srv.Name, srv)
		// run this in it's own goroutine
		go smg.start(shutdown, cl)
	}
	cms.AddInternalConn(utils.SessionS, srv)
	return nil
}

func (smg *SessionService) start(shutdown *utils.SyncedChan, cl *commonlisteners.CommonListenerS) (err error) {
	if err := cl.ServeBiRPC(smg.cfg.SessionSCfg().ListenBiJSON,
		smg.cfg.SessionSCfg().ListenBiGob, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
		smg.mu.Lock()
		smg.bircpEnabled = false
		smg.mu.Unlock()
		shutdown.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (smg *SessionService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown(registry *servmanager.Registry) (err error) {
	smg.mu.Lock()
	defer smg.mu.Unlock()
	close(smg.stopChan)
	if err = smg.sm.Shutdown(); err != nil {
		return
	}
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	if smg.bircpEnabled {
		cl.StopBiRPC()
		smg.bircpEnabled = false
	}
	smg.sm = nil
	cl.RpcUnregisterName(utils.SessionSv1)
	// smg.server.BiRPCUnregisterName(utils.SessionSv1)
	return
}

// ServiceName returns the service name
func (smg *SessionService) ServiceName() string {
	return utils.SessionS
}

// ShouldRun returns if the service should be running
func (smg *SessionService) ShouldRun() bool {
	return smg.cfg.SessionSCfg().Enabled
}
