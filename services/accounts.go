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
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/accounts"
	"github.com/cgrates/cgrates/commonlisteners"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAccountService returns the Account Service
func NewAccountService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *AccountService {
	return &AccountService{
		cfg:       cfg,
		connMgr:   connMgr,
		rldChan:   make(chan struct{}, 1),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// AccountService implements Service interface
type AccountService struct {
	sync.RWMutex

	acts *accounts.AccountS
	cl   *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the service start
func (acts *AccountService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, acts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	acts.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAccounts,
		utils.CacheAccountsFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	acts.Lock()
	defer acts.Unlock()
	acts.acts = accounts.NewAccountS(acts.cfg, fs.FilterS(), acts.connMgr, dbs.DataManager())
	acts.stopChan = make(chan struct{})
	go acts.acts.ListenAndServe(acts.stopChan, acts.rldChan)
	srv, err := engine.NewServiceWithPing(acts.acts, utils.AccountSv1, utils.V1Prfx)
	if err != nil {
		return err
	}

	if !acts.cfg.DispatcherSCfg().Enabled {
		acts.cl.RpcRegister(srv)
	}

	acts.intRPCconn = anz.GetInternalCodec(srv, utils.AccountS)
	return
}

// Reload handles the change of config
func (acts *AccountService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	acts.rldChan <- struct{}{}
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (acts *AccountService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	acts.Lock()
	close(acts.stopChan)
	acts.acts = nil
	acts.Unlock()
	acts.cl.RpcUnregisterName(utils.AccountSv1)
	return
}

// ServiceName returns the service name
func (acts *AccountService) ServiceName() string {
	return utils.AccountS
}

// ShouldRun returns if the service should be running
func (acts *AccountService) ShouldRun() bool {
	return acts.cfg.AccountSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (acts *AccountService) StateChan(stateID string) chan struct{} {
	return acts.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (acts *AccountService) IntRPCConn() birpc.ClientConnector {
	return acts.intRPCconn
}
