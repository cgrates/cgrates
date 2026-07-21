// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/thresholds"
	"github.com/cgrates/cgrates/utils"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig) *ThresholdService {
	return &ThresholdService{
		cfg: cfg,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	cfg  *config.CGRConfig
	thrs *thresholds.ThresholdS
}

// Start should handle the service start
func (s *ThresholdService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheThresholdFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	ts := thresholds.NewThresholdService(s.cfg, dbs.DataManager(), cacheS.CacheS(), fs.FilterS(), cms.ConnManager())
	srv, err := newRPCService(apis.NewThresholdSv1(ts), utils.ThresholdSv1)
	if err != nil {
		return err
	}
	ts.StartLoop(context.TODO())
	s.thrs = ts
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ThresholdS, srv)
	return nil
}

// Reload handles the change of config
func (s *ThresholdService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	s.thrs.Reload(context.TODO())
	return nil
}

// Shutdown stops the service
func (s *ThresholdService) Shutdown(registry *servmanager.Registry) error {
	s.thrs.Shutdown(context.TODO())
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ThresholdSv1)
	return nil
}

// ServiceName returns the service name
func (s *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}

// ShouldRun returns if the service should be running
func (s *ThresholdService) ShouldRun() bool {
	return s.cfg.ThresholdSCfg().Enabled
}
