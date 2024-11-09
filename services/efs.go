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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ExportFailoverService is the service structure for ExportFailover
type ExportFailoverService struct {
	sync.Mutex

	cls *CommonListenerService

	efS *efs.EfS
	cl  *commonlisteners.CommonListenerS
	srv *birpc.Service

	stopChan    chan struct{}
	intConnChan chan birpc.ClientConnector
	connMgr     *engine.ConnManager
	cfg         *config.CGRConfig
	srvDep      map[string]*sync.WaitGroup
}

// NewExportFailoverService is the constructor for the TpeService
func NewExportFailoverService(cfg *config.CGRConfig, connMgr *engine.ConnManager,
	intConnChan chan birpc.ClientConnector,
	cls *CommonListenerService, srvDep map[string]*sync.WaitGroup) *ExportFailoverService {
	return &ExportFailoverService{
		cfg:         cfg,
		cls:         cls,
		connMgr:     connMgr,
		intConnChan: intConnChan,
		srvDep:      srvDep,
	}
}

// Start should handle the service start
func (efServ *ExportFailoverService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if efServ.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	if efServ.cl, err = efServ.cls.WaitForCLS(ctx); err != nil {
		return err
	}
	efServ.Lock()
	efServ.efS = efs.NewEfs(efServ.cfg, efServ.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.EFs))
	efServ.stopChan = make(chan struct{})
	efServ.srv, _ = engine.NewServiceWithPing(efServ.efS, utils.EfSv1, utils.V1Prfx)
	efServ.cl.RpcRegister(efServ.srv)
	efServ.Unlock()
	return
}

// Reload handles the change of config
func (efServ *ExportFailoverService) Reload(ctx *context.Context, _ context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (efServ *ExportFailoverService) Shutdown() (err error) {
	efServ.srv = nil
	close(efServ.stopChan)
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.EFs))
	// NEXT SHOULD EXPORT ALL THE SHUTDOWN LOGGERS TO WRITE
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.EFs))
	return
}

// IsRunning returns if the service is running
func (efServ *ExportFailoverService) IsRunning() bool {
	efServ.Lock()
	defer efServ.Unlock()
	return efServ.efS != nil
}

// ShouldRun returns if the service should be running
func (efServ *ExportFailoverService) ShouldRun() bool {
	return efServ.cfg.EFsCfg().Enabled
}

// ServiceName returns the service name
func (efServ *ExportFailoverService) ServiceName() string {
	return utils.EFs
}
