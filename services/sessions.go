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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// NewSessionService returns the Session Service
func NewSessionService(cfg *config.CGRConfig) *SessionService {
	return &SessionService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// SessionService implements Service interface
type SessionService struct {
	mu           sync.RWMutex
	sm           *sessions.SessionS
	bircpEnabled bool // to stop birpc server if needed
	stopChan     chan struct{}
	cfg          *config.CGRConfig
	stateDeps    *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (smg *SessionService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
		},
		registry, smg.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DataDB].(*DataDBService).DataManager()

	smg.mu.Lock()
	defer smg.mu.Unlock()

	smg.sm = sessions.NewSessionS(smg.cfg, dbs, fs, cms.ConnManager())
	//start sync session in a separate goroutine
	smg.stopChan = make(chan struct{})
	go smg.sm.ListenAndServe(smg.stopChan)
	// Pass internal connection via BiRPCClient

	// Register RPC handler
	srv, _ := engine.NewServiceWithName(smg.sm, utils.SessionS, true) // methods with multiple options
	// srv, _ := birpc.NewService(apis.NewSessionSv1(smg.sm), utils.EmptyString, false) // methods with multiple options
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	// Register BiRpc handlers
	if smg.cfg.SessionSCfg().ListenBijson != utils.EmptyString {
		smg.bircpEnabled = true
		for n, s := range srv {
			cl.BiRPCRegisterName(n, s)
		}
		// run this in it's own goroutine
		go smg.start(shutdown, cl)
	}
	cms.AddInternalConn(utils.SessionS, srv)
	return
}

func (smg *SessionService) start(shutdown *utils.SyncedChan, cl *commonlisteners.CommonListenerS) (err error) {
	if err := cl.ServeBiRPC(smg.cfg.SessionSCfg().ListenBijson,
		smg.cfg.SessionSCfg().ListenBigob, smg.sm.OnBiJSONConnect, smg.sm.OnBiJSONDisconnect); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> serve BiRPC error: %s!", utils.SessionS, err))
		smg.mu.Lock()
		smg.bircpEnabled = false
		smg.mu.Unlock()
		shutdown.CloseOnce()
	}
	return
}

// Reload handles the change of config
func (smg *SessionService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (smg *SessionService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
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

// StateChan returns signaling channel of specific state
func (smg *SessionService) StateChan(stateID string) chan struct{} {
	return smg.stateDeps.StateChan(stateID)
}
