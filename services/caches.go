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

	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig) *CacheService {
	return &CacheService{
		cfg:       cfg,
		cacheCh:   make(chan *engine.CacheS, 1),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	mu        sync.Mutex
	cfg       *config.CGRConfig
	cl        *commonlisteners.CommonListenerS
	cacheCh   chan *engine.CacheS
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start should handle the sercive start
func (cS *CacheService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.DataDB,
			utils.ConnManager,
			utils.CoreS,
		},
		registry, cS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cs := srvDeps[utils.CoreS].(*CoreService)

	cS.mu.Lock()
	defer cS.mu.Unlock()

	engine.Cache = engine.NewCacheS(cS.cfg, dbs.DataManager(), cms.ConnManager(), cs.CoreS().CapsStats)
	go engine.Cache.Precache(shutdown)

	cS.cacheCh <- engine.Cache

	srv, _ := engine.NewService(engine.Cache)
	// srv, _ := birpc.NewService(apis.NewCacheSv1(engine.Cache), "", false)
	for _, s := range srv {
		cS.cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.CacheS, srv)
	return
}

// Reload handles the change of config
func (cS *CacheService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (_ error) {
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

func (cS *CacheService) WaitToPrecache(shutdown *utils.SyncedChan, cacheIDs ...string) (err error) {
	var cacheS *engine.CacheS
	select {
	case <-shutdown.Done():
		return
	case cacheS = <-cS.cacheCh:
		cS.cacheCh <- cacheS
	}
	for _, cacheID := range cacheIDs {
		select {
		case <-shutdown.Done():
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
