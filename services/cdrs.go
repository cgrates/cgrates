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
	"runtime"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *CDRService {
	return &CDRService{
		cfg:       cfg,
		connMgr:   connMgr,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// CDRService implements Service interface
type CDRService struct {
	sync.RWMutex

	cdrS *cdrs.CDRServer
	cl   *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector // expose API methods over internal connection
	stateDeps  *StateDependencies    // channel subscriptions for state changes
}

// Start should handle the sercive start
func (cs *CDRService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
			utils.StorDB,
		},
		registry, cs.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cs.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)
	sdbs := srvDeps[utils.StorDB].(*StorDBService)

	cs.Lock()
	defer cs.Unlock()

	cs.cdrS = cdrs.NewCDRServer(cs.cfg, dbs.DataManager(), fs.FilterS(), cs.connMgr, sdbs.DB())
	runtime.Gosched()
	srv, err := engine.NewServiceWithPing(cs.cdrS, utils.CDRsV1, utils.V1Prfx)
	if err != nil {
		return err
	}
	if !cs.cfg.DispatcherSCfg().Enabled {
		cs.cl.RpcRegister(srv)
	}

	cs.intRPCconn = anz.GetInternalCodec(srv, utils.CDRServer)
	return
}

// Reload handles the change of config
func (cs *CDRService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return
}

// Shutdown stops the service
func (cs *CDRService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	cs.Lock()
	cs.cdrS = nil
	cs.Unlock()
	cs.cl.RpcUnregisterName(utils.CDRsV1)
	return
}

// ServiceName returns the service name
func (cs *CDRService) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cs *CDRService) ShouldRun() bool {
	return cs.cfg.CdrsCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (cs *CDRService) StateChan(stateID string) chan struct{} {
	return cs.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (cs *CDRService) IntRPCConn() birpc.ClientConnector {
	return cs.intRPCconn
}
