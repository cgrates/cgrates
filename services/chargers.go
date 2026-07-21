// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewChargerService returns the Charger Service
func NewChargerService(cfg *config.CGRConfig) *ChargerService {
	return &ChargerService{
		cfg: cfg,
	}
}

// ChargerService implements Service interface
type ChargerService struct {
	mu   sync.RWMutex
	cfg  *config.CGRConfig
	chrS *chargers.ChargerS
}

// Start should handle the service start
func (chrS *ChargerService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		chrS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheChargerProfiles,
		utils.CacheChargerFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	chrS.mu.Lock()
	defer chrS.mu.Unlock()
	chrS.chrS = chargers.NewChargerService(dbs.DataManager(), fs.FilterS(), chrS.cfg, cms.ConnManager())
	srv, err := newRPCService(apis.NewChargerSv1(chrS.chrS), utils.ChargerSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ChargerS, srv)
	return nil
}

// Reload handles the change of config
func (chrS *ChargerService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (chrS *ChargerService) Shutdown(registry *servmanager.Registry) (err error) {
	chrS.mu.Lock()
	defer chrS.mu.Unlock()
	chrS.chrS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ChargerSv1)
	return
}

// ServiceName returns the service name
func (chrS *ChargerService) ServiceName() string {
	return utils.ChargerS
}

// ShouldRun returns if the service should be running
func (chrS *ChargerService) ShouldRun() bool {
	return chrS.cfg.ChargerSCfg().Enabled
}
