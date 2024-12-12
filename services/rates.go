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
func NewRateService(cfg *config.CGRConfig,
	srvIndexer *servmanager.ServiceIndexer) *RateService {
	return &RateService{
		cfg:        cfg,
		rldChan:    make(chan struct{}),
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
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

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// ServiceName returns the service name
func (rs *RateService) ServiceName() string {
	return utils.RateS
}

// ShouldRun returns if the service should be running
func (rs *RateService) ShouldRun() (should bool) {
	return rs.cfg.RateSCfg().Enabled
}

// IsRunning returns if the service is running
func (rs *RateService) IsRunning() bool {
	rs.RLock()
	defer rs.RUnlock()
	return rs.rateS != nil
}

// Reload handles the change of config
func (rs *RateService) Reload(_ chan struct{}) (_ error) {
	rs.rldChan <- struct{}{}
	return
}

// Shutdown stops the service
func (rs *RateService) Shutdown() (err error) {
	rs.Lock()
	defer rs.Unlock()
	close(rs.stopChan)
	rs.rateS.Shutdown() //we don't verify the error because shutdown never returns an err
	rs.rateS = nil
	rs.cl.RpcUnregisterName(utils.RateSv1)
	return
}

// Start should handle the service start
func (rs *RateService) Start(shutdown chan struct{}) (err error) {
	if rs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cls := rs.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), rs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RateS, utils.CommonListenerS, utils.StateServiceUP)
	}
	rs.cl = cls.CLS()
	cacheS := rs.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), rs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RateS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheRateProfiles,
		utils.CacheRateProfilesFilterIndexes,
		utils.CacheRateFilterIndexes); err != nil {
		return
	}
	fs := rs.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), rs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RateS, utils.FilterS, utils.StateServiceUP)
	}
	dbs := rs.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), rs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RateS, utils.DataDB, utils.StateServiceUP)
	}
	anz := rs.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), rs.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.RateS, utils.AnalyzerS, utils.StateServiceUP)
	}

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
	close(rs.stateDeps.StateChan(utils.StateServiceUP))
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
