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
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

// NewTPeService is the constructor for the TpeService
func NewTPeService(cfg *config.CGRConfig, connMgr *engine.ConnManager, dm *DataDBService,
	server *commonlisteners.CommonListenerS, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &TPeService{
		cfg:     cfg,
		srvDep:  srvDep,
		dm:      dm,
		connMgr: connMgr,
		server:  server,
	}
}

// TypeService implements Service interface
type TPeService struct {
	sync.RWMutex

	cfg      *config.CGRConfig
	server   *commonlisteners.CommonListenerS
	connMgr  *engine.ConnManager
	tpes     *tpes.TPeS
	dm       *DataDBService
	srv      *birpc.Service
	stopChan chan struct{}

	srvDep map[string]*sync.WaitGroup
}

// Start should handle the service start
func (tpSrv *TPeService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	var datadb *engine.DataManager
	if datadb, err = tpSrv.dm.WaitForDM(ctx); err != nil {
		return
	}

	tpSrv.tpes = tpes.NewTPeS(tpSrv.cfg, datadb, tpSrv.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.TPeS))
	tpSrv.stopChan = make(chan struct{})
	tpSrv.srv, _ = birpc.NewService(apis.NewTPeSv1(tpSrv.tpes), utils.EmptyString, false)
	tpSrv.server.RpcRegister(tpSrv.srv)
	return
}

// Reload handles the change of config
func (tpSrv *TPeService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (tpSrv *TPeService) Shutdown() (err error) {
	tpSrv.srv = nil
	close(tpSrv.stopChan)
	utils.Logger.Info(fmt.Sprintf("<%s> stopped <%s> subsystem", utils.CoreS, utils.TPeS))
	return
}

// IsRunning returns if the service is running
func (tpSrv *TPeService) IsRunning() bool {
	tpSrv.Lock()
	defer tpSrv.Unlock()
	return tpSrv != nil && tpSrv.tpes != nil
}

// ServiceName returns the service name
func (tpSrv *TPeService) ServiceName() string {
	return utils.TPeS
}

// ShouldRun returns if the service should be running
func (tpSrv *TPeService) ShouldRun() bool {
	return tpSrv.cfg.TpeSCfg().Enabled
}
