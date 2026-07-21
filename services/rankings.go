// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/rankings"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRankingService returns the RankingS Service
func NewRankingService(cfg *config.CGRConfig) *RankingService {
	return &RankingService{
		cfg: cfg,
	}
}

type RankingService struct {
	mu  sync.RWMutex
	cfg *config.CGRConfig
	ran *rankings.RankingS
}

// Start should handle the sercive start
func (ran *RankingService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		ran.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRankingProfiles,
		utils.CacheRankings); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	ran.mu.Lock()
	defer ran.mu.Unlock()
	ran.ran = rankings.NewRankingS(dbs.DataManager(), cacheS.CacheS(), cms.ConnManager(), fs.FilterS(), ran.cfg)
	if err := ran.ran.StartRankingS(context.TODO()); err != nil {
		return err
	}
	srv, err := newRPCService(apis.NewRankingSv1(ran.ran), utils.RankingSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.RankingS, srv)
	return nil
}

// Reload handles the change of config
func (ran *RankingService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	ran.mu.Lock()
	ran.ran.Reload(context.TODO())
	ran.mu.Unlock()
	return
}

// Shutdown stops the service
func (ran *RankingService) Shutdown(registry *servmanager.Registry) (err error) {
	ran.mu.Lock()
	defer ran.mu.Unlock()
	ran.ran.StopRankingS()
	ran.ran = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.RankingSv1)
	return
}

// ServiceName returns the service name
func (ran *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (ran *RankingService) ShouldRun() bool {
	return ran.cfg.RankingSCfg().Enabled
}
