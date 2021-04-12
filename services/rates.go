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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/cores"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	//"github.com/cgrates/cgrates/apier/v1"
)

// NewRateService constructs RateService
func NewRateService(cfg *config.CGRConfig,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	dmS *DataDBService, server *cores.Server,
	intConnChan chan birpc.ClientConnector, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &RateService{
		cfg:         cfg,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		dmS:         dmS,
		server:      server,
		intConnChan: intConnChan,
		rldChan:     make(chan struct{}),
		anz:         anz,
		srvDep:      srvDep,
	}
}

// RateService is the service structure for RateS
type RateService struct {
	sync.RWMutex

	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	dmS         *DataDBService
	cacheS      *engine.CacheS
	server      *cores.Server

	rldChan  chan struct{}
	stopChan chan struct{}

	rateS       *rates.RateS
	rpc         *v1.RateSv1
	intConnChan chan birpc.ClientConnector
	anz         *AnalyzerService
	srvDep      map[string]*sync.WaitGroup
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
func (rs *RateService) Reload() (err error) {
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
	<-rs.intConnChan
	return
}

// Start should handle the service start
func (rs *RateService) Start() (err error) {
	if rs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-rs.cacheS.GetPrecacheChannel(utils.CacheRateProfiles)
	<-rs.cacheS.GetPrecacheChannel(utils.CacheRateProfilesFilterIndexes)
	<-rs.cacheS.GetPrecacheChannel(utils.CacheRateFilterIndexes)

	fltrS := <-rs.filterSChan
	rs.filterSChan <- fltrS

	dbchan := rs.dmS.GetDMChan()
	dm := <-dbchan
	dbchan <- dm
	rs.Lock()
	rs.rateS = rates.NewRateS(rs.cfg, fltrS, dm)
	rs.Unlock()

	rs.stopChan = make(chan struct{})
	go rs.rateS.ListenAndServe(rs.stopChan, rs.rldChan)

	rs.rpc = v1.NewRateSv1(rs.rateS)
	if !rs.cfg.DispatcherSCfg().Enabled {
		rs.server.RpcRegister(rs.rpc)
	}

	rs.intConnChan <- rs.anz.GetInternalCodec(rs.rpc, utils.RateS)
	return
}
