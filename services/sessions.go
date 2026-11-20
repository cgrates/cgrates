// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"context"
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
	birpcEnabled bool
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

	sm := sessions.NewSessionS(smg.cfg, dbs, cacheS.CacheS(), fs, cms.ConnManager())
	if err := sm.RestoreBackups(context.TODO()); err != nil {
		return err
	}
	srv, err := newRPCService(apis.NewSessionSv1(sm), utils.SessionSv1)
	if err != nil {
		return err
	}
	smg.sm = sm
	sm.StartBackups()
	smg.stopChan = make(chan struct{})
	go sm.ListenAndServe(smg.stopChan)
	cl.RpcRegister(srv)
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBiJSON != utils.EmptyString {
		smg.birpcEnabled = true
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
		smg.birpcEnabled = false
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
func (smg *SessionService) Shutdown(registry *servmanager.Registry) error {
	smg.mu.Lock()
	defer smg.mu.Unlock()
	close(smg.stopChan)
	err := smg.sm.Shutdown()
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	if smg.birpcEnabled {
		cl.StopBiRPC()
		smg.birpcEnabled = false
	}
	smg.sm = nil
	cl.RpcUnregisterName(utils.SessionSv1)
	return err
}

// ServiceName returns the service name
func (smg *SessionService) ServiceName() string {
	return utils.SessionS
}

// ShouldRun returns if the service should be running
func (smg *SessionService) ShouldRun() bool {
	return smg.cfg.SessionSCfg().Enabled
}
