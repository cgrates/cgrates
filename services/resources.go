// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/resources"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewResourceService returns the Resource Service
func NewResourceService(cfg *config.CGRConfig) *ResourceService {
	return &ResourceService{
		cfg: cfg,
	}
}

// ResourceService implements Service interface
type ResourceService struct {
	cfg       *config.CGRConfig
	resources *resources.ResourceS
}

// Start should handle the service start
func (s *ResourceService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
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
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheResourceFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	rs := resources.NewResourceService(s.cfg, dbs.DataManager(), cacheS.CacheS(), fs.FilterS(), cms.ConnManager())
	srv, err := newRPCService(apis.NewResourceSv1(rs), utils.ResourceSv1)
	if err != nil {
		return err
	}
	rs.StartLoop(context.TODO())
	s.resources = rs
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ResourceS, srv)
	return nil
}

// Reload handles the change of config
func (s *ResourceService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) error {
	s.resources.Reload(context.TODO())
	return nil
}

// Shutdown stops the service
func (s *ResourceService) Shutdown(registry *servmanager.Registry) error {
	s.resources.Shutdown(context.TODO())
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.ResourceSv1)
	return nil
}

// ServiceName returns the service name
func (s *ResourceService) ServiceName() string {
	return utils.ResourceS
}

// ShouldRun returns if the service should be running
func (s *ResourceService) ShouldRun() bool {
	return s.cfg.ResourceSCfg().Enabled
}
