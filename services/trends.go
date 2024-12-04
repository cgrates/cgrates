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
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewTrendsService returns the TrendS Service
func NewTrendService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS, internalTrendSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anzChan chan *AnalyzerService,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &TrendService{
		connChan:    internalTrendSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		clSChan:     clSChan,
		connMgr:     connMgr,
		anzChan:     anzChan,
		srvDep:      srvDep,
		filterSChan: filterSChan,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

type TrendService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	anzChan     chan *AnalyzerService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS

	trs *engine.TrendS
	cl  *commonlisteners.CommonListenerS

	connChan chan birpc.ClientConnector
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig
	srvDep   map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (trs *TrendService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if trs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	trs.srvDep[utils.DataDB].Add(1)
	trs.cl = <-trs.clSChan
	trs.clSChan <- trs.cl
	if err = trs.cacheS.WaitToPrecache(ctx,
		utils.CacheTrendProfiles,
		utils.CacheTrends,
	); err != nil {
		return err
	}
	var datadb *engine.DataManager
	if datadb, err = trs.dm.WaitForDM(ctx); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, trs.filterSChan); err != nil {
		return
	}
	anz := <-trs.anzChan
	trs.anzChan <- anz

	trs.Lock()
	defer trs.Unlock()
	trs.trs = engine.NewTrendService(datadb, trs.cfg, filterS, trs.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.TrendS))
	if err := trs.trs.StartTrendS(ctx); err != nil {
		return err
	}
	srv, err := engine.NewService(trs.trs)
	if err != nil {
		return err
	}
	if !trs.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			trs.cl.RpcRegister(s)
		}
	}
	trs.intRPCconn = anz.GetInternalCodec(srv, utils.Trends)
	trs.connChan <- trs.intRPCconn
	close(trs.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the change of config
func (trs *TrendService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	trs.Lock()
	trs.trs.Reload(ctx)
	trs.Unlock()
	return
}

// Shutdown stops the service
func (trs *TrendService) Shutdown() (err error) {
	defer trs.srvDep[utils.DataDB].Done()
	trs.Lock()
	defer trs.Unlock()
	trs.trs.StopTrendS()
	trs.trs = nil
	<-trs.connChan
	trs.cl.RpcUnregisterName(utils.TrendSv1)
	return
}

// IsRunning returns if the service is running
func (trs *TrendService) IsRunning() bool {
	return trs.trs != nil
}

// ServiceName returns the service name
func (trs *TrendService) ServiceName() string {
	return utils.TrendS
}

// ShouldRun returns if the service should be running
func (trs *TrendService) ShouldRun() bool {
	return trs.cfg.TrendSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (trs *TrendService) StateChan(stateID string) chan struct{} {
	return trs.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (trs *TrendService) IntRPCConn() birpc.ClientConnector {
	return trs.intRPCconn
}
