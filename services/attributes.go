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
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig,
	dspS *DispatcherService,
	sIndxr *servmanager.ServiceIndexer) *AttributeService {
	return &AttributeService{
		cfg:        cfg,
		dspS:       dspS,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
		srvIndexer: sIndxr,
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	sync.RWMutex

	dspS *DispatcherService

	attrS *engine.AttributeS
	cl    *commonlisteners.CommonListenerS
	rpc   *apis.AttributeSv1 // useful on restart

	cfg *config.CGRConfig

	intRPCconn birpc.ClientConnector       // expose API methods over internal connection
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies
}

// Start should handle the service start
func (attrS *AttributeService) Start(shutdown chan struct{}) (err error) {
	if attrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.CacheS,
			utils.FilterS,
			utils.DataDB,
			utils.AnalyzerS,
		},
		attrS.srvIndexer, attrS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	attrS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cacheS := srvDeps[utils.CacheS].(*CacheService)
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAttributeProfiles,
		utils.CacheAttributeFilterIndexes); err != nil {
		return
	}
	fs := srvDeps[utils.FilterS].(*FilterService)
	dbs := srvDeps[utils.DataDB].(*DataDBService)
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	attrS.Lock()
	defer attrS.Unlock()
	attrS.attrS = engine.NewAttributeService(dbs.DataManager(), fs.FilterS(), attrS.cfg)
	attrS.rpc = apis.NewAttributeSv1(attrS.attrS)
	srv, _ := engine.NewService(attrS.rpc)
	// srv, _ := birpc.NewService(attrS.rpc, "", false)
	if !attrS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			attrS.cl.RpcRegister(s)
		}
	}
	dspShtdChan := attrS.dspS.RegisterShutdownChan(attrS.ServiceName())
	go func() {
		for {
			if _, closed := <-dspShtdChan; closed {
				return
			}
			if attrS.IsRunning() {
				attrS.cl.RpcRegister(srv)
			}

		}
	}()

	attrS.intRPCconn = anz.GetInternalCodec(srv, utils.AttributeS)
	close(attrS.stateDeps.StateChan(utils.StateServiceUP)) // inform listeners about the service reaching UP state
	return
}

// Reload handles the change of config
func (attrS *AttributeService) Reload(_ chan struct{}) (err error) {
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown() (err error) {
	attrS.Lock()
	attrS.attrS = nil
	attrS.rpc = nil
	attrS.cl.RpcUnregisterName(utils.AttributeSv1)
	attrS.dspS.UnregisterShutdownChan(attrS.ServiceName())
	attrS.Unlock()
	return
}

// IsRunning returns if the service is running
func (attrS *AttributeService) IsRunning() bool {
	attrS.RLock()
	defer attrS.RUnlock()
	return attrS.attrS != nil
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}

// ShouldRun returns if the service should be running
func (attrS *AttributeService) ShouldRun() bool {
	return attrS.cfg.AttributeSCfg().Enabled
}

// StateChan returns signaling channel of specific state
func (attrS *AttributeService) StateChan(stateID string) chan struct{} {
	return attrS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (attrS *AttributeService) IntRPCConn() birpc.ClientConnector {
	return attrS.intRPCconn
}
