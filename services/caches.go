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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig, dm *DataDBService, connMgr *engine.ConnManager,
	clSChan chan *commonlisteners.CommonListenerS, internalChan chan birpc.ClientConnector,
	anz *AnalyzerService, // dspS *DispatcherService,
	cores *CoreService,
	srvDep map[string]*sync.WaitGroup) *CacheService {
	return &CacheService{
		cfg:     cfg,
		srvDep:  srvDep,
		anz:     anz,
		cores:   cores,
		clSChan: clSChan,
		dm:      dm,
		connMgr: connMgr,
		rpc:     internalChan,
		cacheCh: make(chan *engine.CacheS, 1),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	anz     *AnalyzerService
	cores   *CoreService
	clSChan chan *commonlisteners.CommonListenerS
	dm      *DataDBService

	cl *commonlisteners.CommonListenerS

	cacheCh chan *engine.CacheS
	rpc     chan birpc.ClientConnector
	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (cS *CacheService) Start(ctx *context.Context, shtDw context.CancelFunc) (err error) {
	cS.cl = <-cS.clSChan
	cS.clSChan <- cS.cl
	var dm *engine.DataManager
	if dm, err = cS.dm.WaitForDM(ctx); err != nil {
		return
	}
	if err = cS.anz.WaitForAnalyzerS(ctx); err != nil {
		return
	}
	var cs *cores.CoreS
	if cs, err = cS.cores.WaitForCoreS(ctx); err != nil {
		return
	}
	engine.Cache = engine.NewCacheS(cS.cfg, dm, cS.connMgr, cs.CapsStats)
	go engine.Cache.Precache(ctx, shtDw)

	cS.cacheCh <- engine.Cache

	srv, _ := engine.NewService(engine.Cache)
	// srv, _ := birpc.NewService(apis.NewCacheSv1(engine.Cache), "", false)
	if !cS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cS.cl.RpcRegister(s)
		}
	}
	cS.rpc <- cS.anz.GetInternalCodec(srv, utils.CacheS)
	return
}

// Reload handles the change of config
func (cS *CacheService) Reload(*context.Context, context.CancelFunc) (_ error) {
	return
}

// Shutdown stops the service
func (cS *CacheService) Shutdown() (_ error) {
	cS.cl.RpcUnregisterName(utils.CacheSv1)
	return
}

// IsRunning returns if the service is running
func (cS *CacheService) IsRunning() bool {
	return true
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

func (cS *CacheService) WaitToPrecache(ctx *context.Context, cacheIDs ...string) (err error) {
	var cacheS *engine.CacheS
	select {
	case <-ctx.Done():
		return ctx.Err()
	case cacheS = <-cS.cacheCh:
		cS.cacheCh <- cacheS
	}
	for _, cacheID := range cacheIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cacheS.GetPrecacheChannel(cacheID):
		}
	}
	return
}
