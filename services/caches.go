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
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewCacheService .
func NewCacheService(cfg *config.CGRConfig, dm *DataDBService,
	server *cores.Server, internalChan chan birpc.ClientConnector,
	anz *AnalyzerService, // dspS *DispatcherService,
	cores *CoreService,
	srvDep map[string]*sync.WaitGroup) *CacheService {
	return &CacheService{
		cfg:     cfg,
		srvDep:  srvDep,
		anz:     anz,
		cores:   cores,
		server:  server,
		dm:      dm,
		rpc:     internalChan,
		cacheCh: make(chan *engine.CacheS, 1),
	}
}

// CacheService implements Agent interface
type CacheService struct {
	cfg    *config.CGRConfig
	anz    *AnalyzerService
	cores  *CoreService
	server *cores.Server
	dm     *DataDBService
	rpc    chan birpc.ClientConnector
	srvDep map[string]*sync.WaitGroup

	cacheCh chan *engine.CacheS
}

// Start should handle the sercive start
func (cS *CacheService) Start(ctx *context.Context, shtDw context.CancelFunc) (err error) {
	var dm *engine.DataManager
	if dm, err = cS.dm.WaitForDM(ctx); err != nil {
		return
	}
	var cs *cores.CoreService
	if cs, err = cS.cores.WaitForCoreS(ctx); err != nil {
		return
	}
	engine.Cache = engine.NewCacheS(cS.cfg, dm, cs.CapsStats)
	go engine.Cache.Precache(ctx, shtDw)

	cS.cacheCh <- engine.Cache

	chSv1, _ := birpc.NewService(apis.NewCacheSv1(engine.Cache), "", false)
	if !cS.cfg.DispatcherSCfg().Enabled {
		cS.server.RpcRegister(chSv1)
	}
	cS.rpc <- cS.anz.GetInternalCodec(chSv1, utils.CacheS)
	return
}

// Reload handles the change of config
func (cS *CacheService) Reload(*context.Context, context.CancelFunc) (_ error) {
	return
}

// Shutdown stops the service
func (cS *CacheService) Shutdown() (_ error) {
	cS.server.RpcUnregisterName(utils.CacheSv1)
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
	return

	for _, cacheID := range cacheIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cacheS.GetPrecacheChannel(cacheID):
		}
	}
	return
}
