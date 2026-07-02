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
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig) *CacheService {
	return &CacheService{
		cfg:     cfg,
		cacheCh: make(chan *engine.CacheS, 1),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	mu      sync.Mutex
	cfg     *config.CGRConfig
	cacheCh chan *engine.CacheS
}

// Start should handle the sercive start
func (cS *CacheService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.DB,
			utils.ConnManager,
			utils.CoreS,
		},
		cS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	dbs := srvDeps[utils.DB].(*DBService)
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cs := srvDeps[utils.CoreS].(*CoreService)

	cS.mu.Lock()
	defer cS.mu.Unlock()

	cache := engine.NewCacheS(cS.cfg, dbs.DataManager(), cms.ConnManager(), cs.CoreS().CapsStats)
	dbs.DataManager().SetCache(cache)
	cms.ConnManager().SetCache(cache)
	if capsStats := cs.CoreS().CapsStats; capsStats != nil {
		capsStats.SetCache(cache)
	}
	go cache.Precache(shutdown)

	cS.cacheCh <- cache

	srv, _ := engine.NewService(cache)
	// srv, _ := birpc.NewService(apis.NewCacheSv1(cache), "", false)
	for _, s := range srv {
		cl.RpcRegister(s)
	}
	cms.AddInternalConn(utils.CacheS, srv)
	if len(cS.cfg.HTTPCfg().RegistrarSURL) != 0 {
		cl.RegisterHTTPFunc(cS.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar(cS.cfg, cache))
	}
	return nil
}

// Reload handles the change of config
func (cS *CacheService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (_ error) {
	return
}

// Shutdown stops the service
func (cS *CacheService) Shutdown(registry *servmanager.Registry) (_ error) {
	cS.mu.Lock()
	defer cS.mu.Unlock()
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.CacheSv1)
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

// CacheS returns the cache, blocking until Start has built it.
func (cS *CacheService) CacheS() *engine.CacheS {
	c := <-cS.cacheCh
	cS.cacheCh <- c
	return c
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
