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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewChargerService returns the Charger Service
func NewChargerService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager,
	srvIndexer *servmanager.ServiceIndexer) *ChargerService {
	return &ChargerService{
		cfg:        cfg,
		connMgr:    connMgr,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// ChargerService implements Service interface
type ChargerService struct {
	sync.RWMutex

	chrS *engine.ChargerS
	cl   *commonlisteners.CommonListenerS

	connMgr *engine.ConnManager
	cfg     *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes
}

// Start should handle the service start
func (chrS *ChargerService) Start(shutdown chan struct{}) (err error) {
	if chrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cls := chrS.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), chrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ChargerS, utils.CommonListenerS, utils.StateServiceUP)
	}
	chrS.cl = cls.CLS()
	cacheS := chrS.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), chrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ChargerS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheChargerProfiles,
		utils.CacheChargerFilterIndexes); err != nil {
		return
	}
	fs := chrS.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), chrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ChargerS, utils.FilterS, utils.StateServiceUP)
	}
	dbs := chrS.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), chrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ChargerS, utils.DataDB, utils.StateServiceUP)
	}
	anz := chrS.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), chrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.ChargerS, utils.AnalyzerS, utils.StateServiceUP)
	}

	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS = engine.NewChargerService(dbs.DataManager(), fs.FilterS(), chrS.cfg, chrS.connMgr)
	srv, _ := engine.NewService(chrS.chrS)
	// srv, _ := birpc.NewService(apis.NewChargerSv1(chrS.chrS), "", false)
	if !chrS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			chrS.cl.RpcRegister(s)
		}
	}

	chrS.intRPCconn = anz.GetInternalCodec(srv, utils.ChargerS)
	close(chrS.stateDeps.StateChan(utils.StateServiceUP))
	return
}

// Reload handles the change of config
func (chrS *ChargerService) Reload(_ chan struct{}) (err error) {
	return
}

// Shutdown stops the service
func (chrS *ChargerService) Shutdown() (err error) {
	chrS.Lock()
	defer chrS.Unlock()
	chrS.chrS = nil
	chrS.cl.RpcUnregisterName(utils.ChargerSv1)
	return
}

// IsRunning returns if the service is running
func (chrS *ChargerService) IsRunning() bool {
	chrS.RLock()
	defer chrS.RUnlock()
	return chrS.chrS != nil
}

// ServiceName returns the service name
func (chrS *ChargerService) ServiceName() string {
	return utils.ChargerS
}

// ShouldRun returns if the service should be running
func (chrS *ChargerService) ShouldRun() bool {
	return chrS.cfg.ChargerSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (chrS *ChargerService) StateChan(stateID string) chan struct{} {
	return chrS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (chrS *ChargerService) IntRPCConn() birpc.ClientConnector {
	return chrS.intRPCconn
}
