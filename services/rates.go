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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewRateService constructs RateService
func NewRateService(cfg *config.CGRConfig) *RateService {
	return &RateService{
		cfg:       cfg,
		rldChan:   make(chan struct{}),
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// RateService is the service structure for RateS
type RateService struct {
	sync.RWMutex

	rateS *rates.RateS
	cl    *commonlisteners.CommonListenerS

	rldChan  chan struct{}
	stopChan chan struct{}
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// ServiceName returns the service name
func (rs *RateService) ServiceName() string {
	return utils.RateS
}

// ShouldRun returns if the service should be running
func (rs *RateService) ShouldRun() (should bool) {
	return rs.cfg.RateSCfg().Enabled
}

// Reload handles the change of config
func (rs *RateService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (_ error) {
	rs.rldChan <- struct{}{}
	return
}

// Shutdown stops the service
func (rs *RateService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	rs.Lock()
	defer rs.Unlock()
	close(rs.stopChan)
	rs.rateS = nil
	rs.cl.RpcUnregisterName(utils.RateSv1)
	return
}

// Start should handle the service start
func (rs *RateService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		registry, rs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	rs.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRateProfiles,
		utils.CacheRateProfilesFilterIndexes,
		utils.CacheRateFilterIndexes); err != nil {
		return err
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	rs.Lock()
	rs.rateS = rates.NewRateS(rs.cfg, fs.FilterS(), dbs.DataManager())
	rs.Unlock()

	rs.stopChan = make(chan struct{})
	go rs.rateS.ListenAndServe(rs.stopChan, rs.rldChan)

	srv, err := engine.NewServiceWithPing(rs.rateS, utils.RateSv1, utils.V1Prfx)
	if err != nil {
		return err
	}
	// srv, _ := birpc.NewService(apis.NewRateSv1(rs.rateS), "", false)
	if !rs.cfg.DispatcherSCfg().Enabled {
		rs.cl.RpcRegister(srv)
	}

	rs.intRPCconn = anz.GetInternalCodec(srv, utils.RateS)
	return
}

// StateChan returns signaling channel of specific state
func (rs *RateService) StateChan(stateID string) chan struct{} {
	return rs.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (rs *RateService) IntRPCConn() birpc.ClientConnector {
	return rs.intRPCconn
}
