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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewChargerService returns the Charger Service
func NewChargerService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *CacheService, filterSChan chan *engine.FilterS, server *cores.Server,
	internalChargerSChan chan birpc.ClientConnector, connMgr *engine.ConnManager,
	anz *AnalyzerService, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &ChargerService{
		connChan:    internalChargerSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
		anz:         anz,
		srvDep:      srvDep,
	}
}

// ChargerService implements Service interface
type ChargerService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *CacheService
	filterSChan chan *engine.FilterS
	server      *cores.Server
	connMgr     *engine.ConnManager

	chrS     *engine.ChargerService
	connChan chan birpc.ClientConnector
	anz      *AnalyzerService
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (chrS *ChargerService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if chrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	if err = chrS.cacheS.WaitToPrecache(ctx,
		utils.CacheChargerProfiles,
		utils.CacheChargerFilterIndexes); err != nil {
		return
	}

	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, chrS.filterSChan); err != nil {
		return
	}

	var datadb *engine.DataManager
	if datadb, err = chrS.dm.WaitForDM(ctx); err != nil {
		return
	}

	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS = engine.NewChargerService(datadb, filterS, chrS.cfg, chrS.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ChargerS))
	srv, _ := birpc.NewService(apis.NewChargerSv1(chrS.chrS), "", false)
	if !chrS.cfg.DispatcherSCfg().Enabled {
		chrS.server.RpcRegister(srv)
	}
	chrS.connChan <- chrS.anz.GetInternalCodec(srv, utils.ChargerS)
	return
}

// Reload handles the change of config
func (chrS *ChargerService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (chrS *ChargerService) Shutdown() (err error) {
	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS.Shutdown()
	chrS.chrS = nil
	<-chrS.connChan
	chrS.server.RpcUnregisterName(utils.ChargerSv1)
	return
}

// IsRunning returns if the service is running
func (chrS *ChargerService) IsRunning() bool {
	chrS.RLock()
	defer chrS.RUnlock()
	return chrS != nil && chrS.chrS != nil
}

// ServiceName returns the service name
func (chrS *ChargerService) ServiceName() string {
	return utils.ChargerS
}

// ShouldRun returns if the service should be running
func (chrS *ChargerService) ShouldRun() bool {
	return chrS.cfg.ChargerSCfg().Enabled
}
