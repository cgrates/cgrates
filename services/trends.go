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

// NewTrendsService returns the TrendS Service
func NewTrendService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) *TrendService {
	return &TrendService{
		cfg:        cfg,
		connMgr:    connMgr,
		srvDep:     srvDep,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

type TrendService struct {
	sync.RWMutex

	trs *engine.TrendS
	cl  *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig
	srvDep  map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
func (trs *TrendService) Start(shutdown chan struct{}) (err error) {
	if trs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	trs.srvDep[utils.DataDB].Add(1)

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		trs.srvIndexer, trs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	trs.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheTrendProfiles,
		utils.CacheTrends); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	trs.Lock()
	defer trs.Unlock()
	trs.trs = engine.NewTrendService(dbs.DataManager(), trs.cfg, fs.FilterS(), trs.connMgr)
	if err := trs.trs.StartTrendS(context.TODO()); err != nil {
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
	close(trs.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the change of config
func (trs *TrendService) Reload(_ chan struct{}) (err error) {
	trs.Lock()
	trs.trs.Reload(context.TODO())
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
