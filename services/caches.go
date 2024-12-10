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
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig, dm *DataDBService, connMgr *engine.ConnManager,
	cores *CoreService,
	srvIndexer *servmanager.ServiceIndexer) *CacheService {
	return &CacheService{
		cfg:        cfg,
		cores:      cores,
		dm:         dm,
		connMgr:    connMgr,
		cacheCh:    make(chan *engine.CacheS, 1),
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	cores *CoreService
	dm    *DataDBService

	cl *commonlisteners.CommonListenerS

	cacheCh chan *engine.CacheS
	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (cS *CacheService) Start(ctx *context.Context, shtDw context.CancelFunc) (err error) {
	cls := cS.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), cS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.CacheS, utils.CommonListenerS, utils.StateServiceUP)
	}
	cS.cl = cls.CLS()
	var dm *engine.DataManager
	if dm, err = cS.dm.WaitForDM(ctx); err != nil {
		return
	}

	anz := cS.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), cS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.CacheS, utils.AnalyzerS, utils.StateServiceUP)
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
	cS.intRPCconn = anz.GetInternalCodec(srv, utils.CacheS)
	close(cS.stateDeps.StateChan(utils.StateServiceUP))
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

// StateChan returns signaling channel of specific state
func (cS *CacheService) StateChan(stateID string) chan struct{} {
	return cS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (cS *CacheService) IntRPCConn() birpc.ClientConnector {
	return cS.intRPCconn
}
