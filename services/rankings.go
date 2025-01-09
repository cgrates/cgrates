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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRankingService returns the RankingS Service
func NewRankingService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup) *RankingService {
	return &RankingService{
		cfg:       cfg,
		connMgr:   connMgr,
		srvDep:    srvDep,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

type RankingService struct {
	sync.RWMutex

	ran *engine.RankingS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (ran *RankingService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) (err error) {
	ran.srvDep[utils.DataDB].Add(1)

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, ran.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	ran.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRankingProfiles,
		utils.CacheRankings); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	ran.Lock()
	defer ran.Unlock()
	ran.ran = engine.NewRankingS(dbs.DataManager(), ran.connMgr, fs.FilterS(), ran.cfg)
	if err := ran.ran.StartRankingS(context.TODO()); err != nil {
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
	return nil
}

// Reload handles the change of config
func (ran *RankingService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) (err error) {
	ran.Lock()
	ran.ran.Reload(context.TODO())
	ran.Unlock()
	return
}

// Shutdown stops the service
func (ran *RankingService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	defer ran.srvDep[utils.DataDB].Done()
	ran.Lock()
	defer ran.Unlock()
	ran.ran.StopRankingS()
	ran.ran = nil
	ran.cl.RpcUnregisterName(utils.RankingSv1)
	return
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
