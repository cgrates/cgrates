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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewTrendsService returns the TrendS Service
func NewTrendService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS,
	server *cores.Server, internalTrendSChan chan birpc.ClientConnector,
	connMgr *engine.ConnManager, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &TrendService{
		connChan: internalTrendSChan,
		cfg:      cfg,
		dm:       dm,
		cacheS:   cacheS,
		server:   server,
		connMgr:  connMgr,
		anz:      anz,
		srvDep:   srvDep,
	}
}

type TrendService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *CacheService
	server      *cores.Server
	connMgr     *engine.ConnManager
	filterSChan chan *engine.FilterS
	connChan    chan birpc.ClientConnector
	trs         *engine.TrendS
	anz         *AnalyzerService
	srvDep      map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (tr *TrendService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if tr.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	tr.srvDep[utils.DataDB].Add(1)
	if err = tr.cacheS.WaitToPrecache(ctx, utils.CacheTrendProfiles); err != nil {
		return err
	}
	var datadb *engine.DataManager
	if datadb, err = tr.dm.WaitForDM(ctx); err != nil {
		return
	}

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, tr.filterSChan); err != nil {
		return
	}
	tr.trs = engine.NewTrendService(datadb, tr.cfg, filterS, tr.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.TrendS))
	srv, _ := engine.NewService(tr.trs)
	if !tr.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			tr.server.RpcRegister(s)
		}
	}
	tr.connChan <- tr.anz.GetInternalCodec(srv, utils.Trends)
	return nil
}

// Reload handles the change of config
func (tr *TrendService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (tr *TrendService) Shutdown() (err error) {
	defer tr.srvDep[utils.DataDB].Done()
	tr.Lock()
	defer tr.Unlock()
	<-tr.connChan
	tr.server.RpcUnregisterName(utils.TrendSv1)
	return
}

// IsRunning returns if the service is running
func (tr *TrendService) IsRunning() bool {
	tr.RLock()
	defer tr.RUnlock()
	return tr != nil && tr.trs != nil
}

// ServiceName returns the service name
func (tr *TrendService) ServiceName() string {
	return utils.TrendS
}

// ShouldRun returns if the service should be running
func (tr *TrendService) ShouldRun() bool {
	return tr.cfg.TrendSCfg().Enabled
}
