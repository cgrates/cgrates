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

// NewRankingService returns the RankingS Service
func NewRankingService(cfg *config.CGRConfig, dm *DataDBService,
	filterSChan chan *engine.FilterS,
	clSChan chan *commonlisteners.CommonListenerS,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &RankingService{
		cfg:         cfg,
		dm:          dm,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

type RankingService struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	filterSChan chan *engine.FilterS

	ran *engine.RankingS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (ran *RankingService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if ran.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	ran.srvDep[utils.DataDB].Add(1)
	ran.cl = <-ran.clSChan
	ran.clSChan <- ran.cl
	cacheS := ran.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), ran.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RankingS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(ctx,
		utils.CacheRankingProfiles,
		utils.CacheRankings,
	); err != nil {
		return err
	}
	var datadb *engine.DataManager
	if datadb, err = ran.dm.WaitForDM(ctx); err != nil {
		return
	}
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, ran.filterSChan); err != nil {
		return
	}
	anz := ran.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), ran.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RankingS, utils.AnalyzerS, utils.StateServiceUP)
	}

	ran.Lock()
	defer ran.Unlock()
	ran.ran = engine.NewRankingS(datadb, ran.connMgr, filterS, ran.cfg)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem",
		utils.CoreS, utils.RankingS))
	if err := ran.ran.StartRankingS(ctx); err != nil {
		return err
	}
	srv, err := engine.NewService(ran.ran)
	if err != nil {
		return err
	}
	if !ran.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			ran.cl.RpcRegister(s)
		}
	}
	ran.intRPCconn = anz.GetInternalCodec(srv, utils.RankingS)
	close(ran.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the change of config
func (ran *RankingService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	ran.Lock()
	ran.ran.Reload(ctx)
	ran.Unlock()
	return
}

// Shutdown stops the service
func (ran *RankingService) Shutdown() (err error) {
	defer ran.srvDep[utils.DataDB].Done()
	ran.Lock()
	defer ran.Unlock()
	ran.ran.StopRankingS()
	ran.ran = nil
	ran.cl.RpcUnregisterName(utils.RankingSv1)
	return
}

// IsRunning returns if the service is running
func (ran *RankingService) IsRunning() bool {
	return ran.ran != nil
}

// ServiceName returns the service name
func (ran *RankingService) ServiceName() string {
	return utils.RankingS
}

// ShouldRun returns if the service should be running
func (ran *RankingService) ShouldRun() bool {
	return ran.cfg.RankingSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (ran *RankingService) StateChan(stateID string) chan struct{} {
	return ran.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (ran *RankingService) IntRPCConn() birpc.ClientConnector {
	return ran.intRPCconn
}
