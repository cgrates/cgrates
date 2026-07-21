// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRateService constructs RateService
func NewRateService(cfg *config.CGRConfig) *RateService {
	return &RateService{
		cfg: cfg,
	}
}

// RateService is the service structure for RateS
type RateService struct {
	mu    sync.RWMutex
	cfg   *config.CGRConfig
	rateS *rates.RateS
}

// Start should handle the service start
func (rs *RateService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		rs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRateProfiles,
		utils.CacheRateProfilesFilterIndexes,
		utils.CacheRateFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()

	rs.mu.Lock()
	rs.rateS = rates.NewRateS(rs.cfg, fs, dbs)
	rs.mu.Unlock()

	srv, err := newRPCService(apis.NewRateSv1(rs.rateS), utils.RateSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.RateS, srv)
	return
}

// Reload handles the change of config
func (rs *RateService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	return nil
}

// Shutdown stops the service
func (rs *RateService) Shutdown(registry *servmanager.Registry) (err error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.rateS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.RateSv1)
	return
}

// ServiceName returns the service name
func (rs *RateService) ServiceName() string {
	return utils.RateS
}

// ShouldRun returns if the service should be running
func (rs *RateService) ShouldRun() (should bool) {
	return rs.cfg.RateSCfg().Enabled
}
