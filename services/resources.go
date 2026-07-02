/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
