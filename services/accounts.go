// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/accounts"
	"github.com/cgrates/cgrates/apis"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAccountService returns the Account Service
func NewAccountService(cfg *config.CGRConfig) *AccountService {
	return &AccountService{cfg: cfg}
}

// AccountService implements Service interface
type AccountService struct {
	mu   sync.RWMutex
	cfg  *config.CGRConfig
	acts *accounts.AccountS
}

// Start should handle the service start
func (acts *AccountService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		acts.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAccounts,
		utils.CacheAccountsFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()

	acts.mu.Lock()
	defer acts.mu.Unlock()
	acts.acts = accounts.NewAccountS(acts.cfg, fs, cms.ConnManager(), dbs)
	srv, err := newRPCService(apis.NewAccountSv1(acts.acts), utils.AccountSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.AccountS, srv)
	return
}

// Reload handles the change of config
func (acts *AccountService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service
func (acts *AccountService) Shutdown(registry *servmanager.Registry) (err error) {
	acts.mu.Lock()
	defer acts.mu.Unlock()
	acts.acts = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.AccountSv1)
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
