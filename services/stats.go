// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/stats"
	"github.com/cgrates/cgrates/utils"
)

// NewStatService returns the Stat Service
func NewStatService(cfg *config.CGRConfig) *StatService {
	return &StatService{
		cfg: cfg,
	}
}

// StatService implements Service interface
type StatService struct {
	cfg *config.CGRConfig
	sts *stats.StatS
}

// Start should handle the service start
func (s *StatService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
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
		utils.CacheStatQueueProfiles,
		utils.CacheStatQueues,
		utils.CacheStatFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	ss := stats.NewStatService(s.cfg, dbs.DataManager(), cacheS.CacheS(), fs.FilterS(), cms.ConnManager())
	srv, err := newRPCService(apis.NewStatSv1(ss), utils.StatSv1)
	if err != nil {
		return err
	}
	ss.StartLoop(context.TODO())
	s.sts = ss
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.StatS, srv)
	return nil
}

// Reload handles the change of config
func (s *StatService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	s.sts.Reload(context.TODO())
	return nil
}

// Shutdown stops the service
func (s *StatService) Shutdown(registry *servmanager.Registry) error {
	s.sts.Shutdown(context.TODO())
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.StatSv1)
	return nil
}

// ServiceName returns the service name
func (s *StatService) ServiceName() string {
	return utils.StatS
}

// ShouldRun returns if the service should be running
func (s *StatService) ShouldRun() bool {
	return s.cfg.StatSCfg().Enabled
}
