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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAPIerSv1Service returns the APIerSv1 Service
func NewAdminSv1Service(cfg *config.CGRConfig,
	dm *DataDBService, storDB *StorDBService,
	filterSChan chan *engine.FilterS, clSChan chan *commonlisteners.CommonListenerS,
	connMgr *engine.ConnManager, anzChan chan *AnalyzerService,
	srvIndexer *servmanager.ServiceIndexer) servmanager.Service {
	return &AdminSv1Service{
		cfg:         cfg,
		dm:          dm,
		storDB:      storDB,
		filterSChan: filterSChan,
		clSChan:     clSChan,
		connMgr:     connMgr,
		anzChan:     anzChan,
		srvIndexer:  srvIndexer,
		stateDeps:   NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// AdminSv1Service implements Service interface
type AdminSv1Service struct {
	sync.RWMutex

	clSChan     chan *commonlisteners.CommonListenerS
	dm          *DataDBService
	storDB      *StorDBService
	anzChan     chan *AnalyzerService
	filterSChan chan *engine.FilterS

	api *apis.AdminSv1
	cl  *commonlisteners.CommonListenerS

	stopChan chan struct{}
	connMgr  *engine.ConnManager
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector       // RPC connector with internal APIs
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (apiService *AdminSv1Service) Start(ctx *context.Context, _ context.CancelFunc) (err error) {
	if apiService.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	apiService.cl = <-apiService.clSChan
	apiService.clSChan <- apiService.cl
	var filterS *engine.FilterS
	if filterS, err = waitForFilterS(ctx, apiService.filterSChan); err != nil {
		return
	}
	var datadb *engine.DataManager
	if datadb, err = apiService.dm.WaitForDM(ctx); err != nil {
		return
	}
	anz := <-apiService.anzChan
	apiService.anzChan <- anz

	storDBChan := make(chan engine.StorDB, 1)
	apiService.stopChan = make(chan struct{})
	apiService.storDB.RegisterSyncChan(storDBChan)

	apiService.Lock()
	defer apiService.Unlock()

	apiService.api = apis.NewAdminSv1(apiService.cfg, datadb, apiService.connMgr, filterS, storDBChan)

	// go apiService.api.ListenAndServe(apiService.stopChan)
	// runtime.Gosched()
	srv, _ := engine.NewService(apiService.api)
	// srv, _ := birpc.NewService(apiService.api, "", false)

	if !apiService.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			apiService.cl.RpcRegister(s)
		}
		rpl, _ := engine.NewService(apis.NewReplicatorSv1(datadb, apiService.api))
		for _, s := range rpl {
			apiService.cl.RpcRegister(s)
		}
	}

	//backwards compatible
	apiService.intRPCconn = anz.GetInternalCodec(srv, utils.AdminSv1)
	close(apiService.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (apiService *AdminSv1Service) Reload(*context.Context, context.CancelFunc) (err error) {
	return
}

// Shutdown stops the service
func (apiService *AdminSv1Service) Shutdown() (err error) {
	apiService.Lock()
	// close(apiService.stopChan)
	apiService.api = nil
	apiService.cl.RpcUnregisterName(utils.AdminSv1)
	apiService.Unlock()
	return
}

// IsRunning returns if the service is running
func (apiService *AdminSv1Service) IsRunning() bool {
	apiService.RLock()
	defer apiService.RUnlock()
	return apiService.api != nil
}

// ServiceName returns the service name
func (apiService *AdminSv1Service) ServiceName() string {
	return utils.AdminS
}

// ShouldRun returns if the service should be running
func (apiService *AdminSv1Service) ShouldRun() bool {
	return apiService.cfg.AdminSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (apiService *AdminSv1Service) StateChan(stateID string) chan struct{} {
	return apiService.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (apiService *AdminSv1Service) IntRPCConn() birpc.ClientConnector {
	return apiService.intRPCconn
}
