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

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig, connMgr *engine.ConnManager) *CacheService {
	return &CacheService{
		cfg:       cfg,
		connMgr:   connMgr,
		cacheCh:   make(chan *engine.CacheS, 1),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	mu sync.Mutex
	cl *commonlisteners.CommonListenerS

	cacheCh chan *engine.CacheS
	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (cS *CacheService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.DataDB,
			utils.AnalyzerS,
			utils.CoreS,
		},
		registry, cS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)
	cs := srvDeps[utils.CoreS].(*CoreService)

	cS.mu.Lock()
	defer cS.mu.Unlock()

	engine.Cache = engine.NewCacheS(cS.cfg, dbs.DataManager(), cS.connMgr, cs.CoreS().CapsStats)
	go engine.Cache.Precache(shutdown)

	cS.cacheCh <- engine.Cache

	srv, _ := engine.NewService(engine.Cache)
	// srv, _ := birpc.NewService(apis.NewCacheSv1(engine.Cache), "", false)
	if !cS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cS.cl.RpcRegister(s)
		}
	}
	cS.intRPCconn = anz.GetInternalCodec(srv, utils.CacheS)
	return
}

// Reload handles the change of config
func (cS *CacheService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (_ error) {
	return
}

// Shutdown stops the service
func (cS *CacheService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	cS.cl.RpcUnregisterName(utils.CacheSv1)
	return
}

// ServiceName returns the service name
func (cS *CacheService) ServiceName() string {
	return utils.CacheS
}

// ShouldRun returns if the service should be running
func (cS *CacheService) ShouldRun() bool {
	return true
}

// GetDMChan returns the CacheS chanel
func (cS *CacheService) GetCacheSChan() chan *engine.CacheS {
	return cS.cacheCh
}

func (cS *CacheService) WaitToPrecache(shutdown chan struct{}, cacheIDs ...string) (err error) {
	var cacheS *engine.CacheS
	select {
	case <-shutdown:
		return
	case cacheS = <-cS.cacheCh:
		cS.cacheCh <- cacheS
	}
	for _, cacheID := range cacheIDs {
		select {
		case <-shutdown:
			return
		case <-cacheS.GetPrecacheChannel(cacheID):
		}
	}
	return
}

// StateChan returns signaling channel of specific state
func (cS *CacheService) StateChan(stateID string) chan struct{} {
	return cS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (cS *CacheService) IntRPCConn() birpc.ClientConnector {
	return cS.intRPCconn
}
