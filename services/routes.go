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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRouteService returns the Route Service
func NewRouteService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *RouteService {
	return &RouteService{
		cfg:       cfg,
		connMgr:   connMgr,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// RouteService implements Service interface
type RouteService struct {
	sync.RWMutex

	routeS *engine.RouteS
	cl     *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (routeS *RouteService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, routeS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	routeS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRouteProfiles,
		utils.CacheRouteFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS = engine.NewRouteService(dbs.DataManager(), fs.FilterS(), routeS.cfg, routeS.connMgr)
	srv, _ := engine.NewService(routeS.routeS)
	// srv, _ := birpc.NewService(apis.NewRouteSv1(routeS.routeS), "", false)
	if !routeS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			routeS.cl.RpcRegister(s)
		}
	}
	routeS.intRPCconn = anz.GetInternalCodec(srv, utils.RouteS)
	return
}

// Reload handles the change of config
func (routeS *RouteService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (routeS *RouteService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	routeS.Lock()
	defer routeS.Unlock()
	routeS.routeS = nil
	routeS.cl.RpcUnregisterName(utils.RouteSv1)
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

// IntRPCConn returns the internal connection used by RPCClient
func (routeS *RouteService) IntRPCConn() birpc.ClientConnector {
	return routeS.intRPCconn
}
