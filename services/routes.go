// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/routes"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig) *RouteService {
	return &RouteService{
		cfg: cfg,
	}
}

// RouteService implements Service interface
type RouteService struct {
	mu     sync.RWMutex
	cfg    *config.CGRConfig
	routeS *routes.RouteS
}

// Start should handle the sercive start
func (routeS *RouteService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) (err error) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DB,
		},
		routeS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRouteProfiles,
		utils.CacheRouteFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DB].(*DBService)

	routeS.mu.Lock()
	defer routeS.mu.Unlock()
	routeS.routeS = routes.NewRouteService(dbs.DataManager(), fs.FilterS(), routeS.cfg, cms.ConnManager())
	srv, err := newRPCService(apis.NewRouteSv1(routeS.routeS), utils.RouteSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.RouteS, srv)
	return
}

// Reload handles the change of config
func (routeS *RouteService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown(registry *servmanager.Registry) (err error) {
	routeS.mu.Lock()
	defer routeS.mu.Unlock()
	routeS.routeS = nil
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.RouteSv1)
	return
}

// ServiceName returns the service name
func (routeS *RouteService) ServiceName() string {
	return utils.RouteS
}

// ShouldRun returns if the service should be running
func (routeS *RouteService) ShouldRun() bool {
	return routeS.cfg.RouteSCfg().Enabled
}
