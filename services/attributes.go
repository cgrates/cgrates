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
	if utils.StructChanTimeout(
		attrS.srvIndexer.GetService(utils.CommonListenerS).StateChan(utils.StateServiceUP),
		attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.CommonListenerS, utils.StateServiceUP)
	}
	cls := attrS.srvIndexer.GetService(utils.CommonListenerS).(*CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.CommonListenerS, utils.StateServiceUP)
	}
	attrS.cl = cls.CLS()
	cacheS := attrS.srvIndexer.GetService(utils.CacheS).(*CacheService)
	if utils.StructChanTimeout(cacheS.StateChan(utils.StateServiceUP), attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.CacheS, utils.StateServiceUP)
	}
	if err = cacheS.WaitToPrecache(shutdown,
		utils.CacheAttributeProfiles,
		utils.CacheAttributeFilterIndexes); err != nil {
		return
	}
	fs := attrS.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.FilterS, utils.StateServiceUP)
	}
	dbs := attrS.srvIndexer.GetService(utils.DataDB).(*DataDBService)
	if utils.StructChanTimeout(dbs.StateChan(utils.StateServiceUP), attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.DataDB, utils.StateServiceUP)
	}
	anz := attrS.srvIndexer.GetService(utils.AnalyzerS).(*AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), attrS.cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.AttributeS, utils.AnalyzerS, utils.StateServiceUP)
	}

	attrS.Lock()
	defer attrS.Unlock()
	attrS.attrS = engine.NewAttributeService(dbs.DataManager(), fs.FilterS(), attrS.cfg)
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AttributeS))
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
	attrS.attrS.Shutdown()
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
