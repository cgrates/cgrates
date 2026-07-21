// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/apis"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewActionService returns the Action Service
func NewActionService(cfg *config.CGRConfig) *ActionService {
	return &ActionService{cfg: cfg}
}

// ActionService implements Service interface
type ActionService struct {
	mu   sync.RWMutex
	cfg  *config.CGRConfig
	acts *actions.ActionS
}

// Start should handle the service start
func (acts *ActionService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
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
		utils.CacheActionProfiles,
		utils.CacheActionProfilesFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()

	acts.mu.Lock()
	defer acts.mu.Unlock()
	acts.acts = actions.NewActionS(acts.cfg, cacheS.CacheS(), fs, dbs, cms.ConnManager())
	srv, err := newRPCService(apis.NewActionSv1(acts.acts), utils.ActionSv1)
	if err != nil {
		return
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ActionS, srv)
	return
}

// Reload handles the change of config
func (acts *ActionService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service
func (acts *ActionService) Shutdown(registry *servmanager.Registry) (err error) {
	acts.mu.Lock()
	defer acts.mu.Unlock()
	acts.acts.Shutdown()
	acts.acts = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ActionSv1)
	return
}

// ServiceName returns the service name
func (acts *ActionService) ServiceName() string {
	return utils.ActionS
}

// ShouldRun returns if the service should be running
func (acts *ActionService) ShouldRun() bool {
	return acts.cfg.ActionSCfg().Enabled
}
