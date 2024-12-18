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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewConnManagerService instantiates a new ConnManagerService.
func NewConnManagerService(cfg *config.CGRConfig) *ConnManagerService {
	return &ConnManagerService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// ConnManagerService implements Service interface.
type ConnManagerService struct {
	mu        sync.RWMutex
	cfg       *config.CGRConfig
	connMgr   *engine.ConnManager
	anz       *AnalyzerService
	stateDeps *StateDependencies // channel subscriptions for state changes
}

// Start handles the service start.
func (s *ConnManagerService) Start(shutdown chan struct{}, registry *servmanager.ServiceRegistry) error {
	s.anz = registry.Lookup(utils.AnalyzerS).(*AnalyzerService)
	if s.anz.ShouldRun() { // wait for AnalyzerS only if it should run
		if _, err := WaitForServiceState(utils.StateServiceInit, utils.AnalyzerS, registry,
			s.cfg.GeneralCfg().ConnectTimeout); err != nil {
			return err
		}
	}
	s.connMgr = engine.NewConnManager(s.cfg)
	close(s.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the config changes.
func (s *ConnManagerService) Reload(_ chan struct{}, _ *servmanager.ServiceRegistry) error {
	s.connMgr.Reload()
	return nil
}

// Shutdown stops the service.
func (s *ConnManagerService) Shutdown(_ *servmanager.ServiceRegistry) error {
	s.connMgr = nil
	engine.SetConnManager(nil)
	close(s.stateDeps.StateChan(utils.StateServiceDOWN))
	return nil
}

// ServiceName returns the service name
func (s *ConnManagerService) ServiceName() string {
	return utils.ConnManager
}

// ShouldRun returns if the service should be running.
func (s *ConnManagerService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (s *ConnManagerService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}

// ConnManager returns the ConnManager object.
func (s *ConnManagerService) ConnManager() *engine.ConnManager {
	return s.connMgr
}

// AddInternalConn registers direct internal RPC access for a service.
// TODO: Add function to remove internal conns (useful for shutdown).
func (s *ConnManagerService) AddInternalConn(svcName string, receiver birpc.ClientConnector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	route, exists := serviceMethods[svcName]
	if !exists {
		return
	}
	rpcIntChan := make(chan birpc.ClientConnector, 1)
	s.connMgr.AddInternalConn(route.internalPath, route.receiver, rpcIntChan)
	if route.biRPCPath != "" {
		s.connMgr.AddInternalConn(route.biRPCPath, route.receiver, rpcIntChan)
	}
	rpcIntChan <- s.anz.GetInternalCodec(receiver, svcName)
}

// internalRoute defines how a service's methods can be accessed internally within the system.
type internalRoute struct {
	receiver     string // method receiver name (e.g. "ChargerSv1")
	internalPath string // internal API path
	biRPCPath    string // bidirectional API path, if supported
}

var serviceMethods = map[string]internalRoute{
	utils.AnalyzerS: {
		receiver:     utils.AnalyzerSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzerS),
	},
	utils.AdminS: {
		receiver:     utils.AdminSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS),
	},
	utils.AttributeS: {
		receiver:     utils.AttributeSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes),
	},
	utils.CacheS: {
		receiver:     utils.CacheSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches),
	},
	utils.CDRs: {
		receiver:     utils.CDRsV1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs),
	},
	utils.ChargerS: {
		receiver:     utils.ChargerSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers),
	},
	utils.GuardianS: {
		receiver:     utils.GuardianSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian),
	},
	utils.LoaderS: {
		receiver:     utils.LoaderSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders),
	},
	utils.ResourceS: {
		receiver:     utils.ResourceSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources),
	},
	utils.SessionS: {
		receiver:     utils.SessionSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS),
		biRPCPath:    utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS),
	},
	utils.StatS: {
		receiver:     utils.StatSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats),
	},
	utils.RankingS: {
		receiver:     utils.RankingSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings),
	},
	utils.TrendS: {
		receiver:     utils.TrendSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends),
	},
	utils.RouteS: {
		receiver:     utils.RouteSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes),
	},
	utils.ThresholdS: {
		receiver:     utils.ThresholdSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds),
	},
	utils.ServiceManagerS: {
		receiver:     utils.ServiceManagerV1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager),
	},
	utils.ConfigS: {
		receiver:     utils.ConfigSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig),
	},
	utils.CoreS: {
		receiver:     utils.CoreSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore),
	},
	utils.EEs: {
		receiver:     utils.EeSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs),
	},
	utils.RateS: {
		receiver:     utils.RateSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates),
	},
	utils.DispatcherS: {
		receiver:     utils.DispatcherSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers),
	},
	utils.AccountS: {
		receiver:     utils.AccountSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts),
	},
	utils.ActionS: {
		receiver:     utils.ActionSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions),
	},
	utils.TPeS: {
		receiver:     utils.TPeSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTpes),
	},
	utils.EFs: {
		receiver:     utils.EfSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs),
	},
	utils.ERs: {
		receiver:     utils.ErSv1,
		internalPath: utils.ConcatenatedKey(utils.MetaInternal, utils.MetaERs),
	},
}
