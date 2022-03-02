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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

// NewTpeService is the constructor for the TpeService
func NewTpeService(cfg *config.CGRConfig, connMgr *engine.ConnManager,
	server *cores.Server, srvDep map[string]*sync.WaitGroup) servmanager.Service {
	return &TpeService{
		cfg:     cfg,
		srvDep:  srvDep,
		connMgr: connMgr,
		server:  server,
	}
}

// TypeService implements Service interface
type TpeService struct {
	sync.RWMutex

	cfg      *config.CGRConfig
	server   *cores.Server
	connMgr  *engine.ConnManager
	tpes     *tpes.TpeS
	srv      *birpc.Service
	stopChan chan struct{}

	srvDep map[string]*sync.WaitGroup
}

// Start should handle the service start
func (tpSrv *TpeService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	tpSrv.tpes = tpes.NewTpeS(tpSrv.cfg, tpSrv.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.TpeS))
	tpSrv.stopChan = make(chan struct{})
	tpSrv.srv, _ = birpc.NewService(apis.NewTpeSv1(tpSrv.tpes), utils.EmptyString, false)
	tpSrv.server.RpcRegister(tpSrv.srv)
	return
}

// Reload handles the change of config
func (tpSrv *TpeService) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (tpSrv *TpeService) Shutdown() (err error) {
	tpSrv.srv = nil
	close(tpSrv.stopChan)
	return
}

// IsRunning returns if the service is running
func (tpSrv *TpeService) IsRunning() bool {
	tpSrv.Lock()
	defer tpSrv.Unlock()
	return tpSrv != nil && tpSrv.tpes != nil
}

// ServiceName returns the service name
func (tpSrv *TpeService) ServiceName() string {
	return utils.TpeS
}

// ShouldRun returns if the service should be running
func (tpSrv *TpeService) ShouldRun() bool {
	return tpSrv.cfg.TpeSCfg().Enabled
}
