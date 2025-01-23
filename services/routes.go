/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package services

import (
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig) *RouteService {
	return &RouteService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// RouteService implements Service interface
type RouteService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	routeS    *engine.RouteS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (routeS *RouteService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
		},
		registry, routeS.cfg.GeneralCfg().ConnectTimeout)
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
	dbs := srvDeps[utils.DataDB].(*DataDBService)

	routeS.mu.Lock()
	defer routeS.mu.Unlock()
	routeS.routeS = engine.NewRouteService(dbs.DataManager(), fs.FilterS(), routeS.cfg, cms.ConnManager())
	srv, _ := engine.NewService(routeS.routeS)
	// srv, _ := birpc.NewService(apis.NewRouteSv1(routeS.routeS), "", false)
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.RouteS, srv)
	return
}

// Reload handles the change of config
func (routeS *RouteService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown(registry *servmanager.ServiceRegistry) (err error) {
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

// StateChan returns signaling channel of specific state
func (routeS *RouteService) StateChan(stateID string) chan struct{} {
	return routeS.stateDeps.StateChan(stateID)
}
