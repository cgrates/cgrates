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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// ExportFailoverService is the service structure for ExportFailover
type ExportFailoverService struct {
	sync.Mutex

	clSChan chan *commonlisteners.CommonListenerS

	efS *efs.EfS
	cl  *commonlisteners.CommonListenerS
	srv *birpc.Service

	stopChan    chan struct{}
	intConnChan chan birpc.ClientConnector
	connMgr     *engine.ConnManager
	cfg         *config.CGRConfig
	srvDep      map[string]*sync.WaitGroup

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// NewExportFailoverService is the constructor for the TpeService
func NewExportFailoverService(cfg *config.CGRConfig, connMgr *engine.ConnManager,
	intConnChan chan birpc.ClientConnector,
	clSChan chan *commonlisteners.CommonListenerS,
	srvDep map[string]*sync.WaitGroup,
	srvIndexer *servmanager.ServiceIndexer) *ExportFailoverService {
	return &ExportFailoverService{
		cfg:         cfg,
		clSChan:     clSChan,
		connMgr:     connMgr,
		intConnChan: intConnChan,
		srvDep:      srvDep,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// Start should handle the service start
func (efServ *ExportFailoverService) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if efServ.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	efServ.cl = <-efServ.clSChan
	efServ.clSChan <- efServ.cl
	efServ.Lock()
	efServ.efS = efs.NewEfs(efServ.cfg, efServ.connMgr)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.EFs))
	efServ.stopChan = make(chan struct{})
	efServ.srv, _ = engine.NewServiceWithPing(efServ.efS, utils.EfSv1, utils.V1Prfx)
	efServ.cl.RpcRegister(efServ.srv)
	efServ.Unlock()
	close(efServ.stateDeps.StateChan(utils.StateServiceUP))
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

// StateChan returns signaling channel of specific state
func (efServ *ExportFailoverService) StateChan(stateID string) chan struct{} {
	return efServ.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (efServ *ExportFailoverService) IntRPCConn() birpc.ClientConnector {
	return efServ.intRPCconn
}
