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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	//"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/rpcclient"
)

// NewRateService constructs RateService
func NewRateService(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *utils.Server, exitChan chan bool,
	intConnChan chan rpcclient.ClientConnector) servmanager.Service {
	return &RateService{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		exitChan:    exitChan,
		intConnChan: intConnChan,
		rldChan:     make(chan struct{}),
	}
}

// RateService is the service structure for RateS
type RateService struct {
	sync.RWMutex

	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	server      *utils.Server
	exitChan    chan bool
	intConnChan chan rpcclient.ClientConnector
	rldChan     chan struct{}

	rateS *rates.RateS
	//rpc *v1.EventExporterSv1
}

// GetIntenternalChan is deprecated and it will be removed shortly
func (rs *RateService) GetIntenternalChan() (conn chan rpcclient.ClientConnector) {
	panic("deprecated method")
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
	if err = rs.rateS.Shutdown(); err != nil {
		return
	}
	rs.rateS = nil
	<-rs.intConnChan
	return
}

// Start should handle the service start
func (rs *RateService) Start() (err error) {
	if rs.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	fltrS := <-rs.filterSChan
	rs.filterSChan <- fltrS

	rs.Lock()
	rs.rateS = rates.NewRateS(rs.cfg, fltrS)
	rs.Unlock()
	/*rs.rpc = v1.NewEventExporterSv1(es.eeS)
	if !rs.cfg.DispatcherSCfg().Enabled {
		rs.server.RpcRegister(es.rpc)
	}
	*/
	rs.intConnChan <- rs.rateS
	return rs.rateS.ListenAndServe(rs.exitChan, rs.rldChan)
}
