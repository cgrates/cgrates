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
	"sync"

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

	efS *efs.EfS
	cl  *commonlisteners.CommonListenerS
	srv *birpc.Service

	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// NewExportFailoverService is the constructor for the TpeService
func NewExportFailoverService(cfg *config.CGRConfig, connMgr *engine.ConnManager) *ExportFailoverService {
	return &ExportFailoverService{
		cfg:       cfg,
		connMgr:   connMgr,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// Start should handle the service start
func (efServ *ExportFailoverService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	cls, err := WaitForServiceState(utils.StateServiceUP, utils.CommonListenerS, registry,
		efServ.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	efServ.cl = cls.(*CommonListenerService).CLS()

	efServ.Lock()
	defer efServ.Unlock()

	efServ.efS = efs.NewEfs(efServ.cfg, efServ.connMgr)
	efServ.stopChan = make(chan struct{})
	efServ.srv, _ = engine.NewServiceWithPing(efServ.efS, utils.EfSv1, utils.V1Prfx)
	efServ.cl.RpcRegister(efServ.srv)
	return
}

// Reload handles the change of config
func (efServ *ExportFailoverService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (efServ *ExportFailoverService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	efServ.srv = nil
	close(efServ.stopChan)
	// NEXT SHOULD EXPORT ALL THE SHUTDOWN LOGGERS TO WRITE
	return
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
