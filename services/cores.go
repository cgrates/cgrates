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
	"os"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps,
	fileCPU *os.File, shdWg *sync.WaitGroup,
	srvIndexer *servmanager.ServiceRegistry) *CoreService {
	return &CoreService{
		shdWg:      shdWg,
		cfg:        cfg,
		caps:       caps,
		fileCPU:    fileCPU,
		csCh:       make(chan *cores.CoreS, 1),
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}
}

// CoreService implements Service interface
type CoreService struct {
	mu sync.RWMutex

	cS *cores.CoreS
	cl *commonlisteners.CommonListenerS

	fileCPU  *os.File
	caps     *engine.Caps
	csCh     chan *cores.CoreS
	stopChan chan struct{}
	shdWg    *sync.WaitGroup
	cfg      *config.CGRConfig

	intRPCconn birpc.ClientConnector        // expose API methods over internal connection
	srvIndexer *servmanager.ServiceRegistry // access directly services from here
	stateDeps  *StateDependencies           // channel subscriptions for state changes
}

// Start should handle the service start
func (cS *CoreService) Start(shutdown chan struct{}) error {
	if cS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	srvDeps, err := waitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.AnalyzerS,
		},
		cS.srvIndexer, cS.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cS.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	anz := srvDeps[utils.AnalyzerS].(*AnalyzerService)

	cS.mu.Lock()
	defer cS.mu.Unlock()
	cS.stopChan = make(chan struct{})
	cS.cS = cores.NewCoreService(cS.cfg, cS.caps, cS.fileCPU, cS.stopChan, cS.shdWg, shutdown)
	cS.csCh <- cS.cS
	srv, err := engine.NewService(cS.cS)
	if err != nil {
		return err
	}
	if !cS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cS.cl.RpcRegister(s)
		}
	}

	cS.intRPCconn = anz.GetInternalCodec(srv, utils.CoreS)
	close(cS.stateDeps.StateChan(utils.StateServiceUP))
	return nil
}

// Reload handles the change of config
func (cS *CoreService) Reload(_ chan struct{}) error {
	return nil
}

// Shutdown stops the service
func (cS *CoreService) Shutdown() error {
	cS.mu.Lock()
	defer cS.mu.Unlock()
	cS.cS.Shutdown()
	close(cS.stopChan)
	cS.cS.StopCPUProfiling()
	cS.cS.StopMemoryProfiling()
	cS.cS = nil
	<-cS.csCh
	cS.cl.RpcUnregisterName(utils.CoreSv1)
	return nil
}

// IsRunning returns if the service is running
func (cS *CoreService) IsRunning() bool {
	cS.mu.RLock()
	defer cS.mu.RUnlock()
	return cS.cS != nil
}

// ServiceName returns the service name
func (cS *CoreService) ServiceName() string {
	return utils.CoreS
}

// ShouldRun returns if the service should be running
func (cS *CoreService) ShouldRun() bool {
	return true
}

// StateChan returns signaling channel of specific state
func (cS *CoreService) StateChan(stateID string) chan struct{} {
	return cS.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (cS *CoreService) IntRPCConn() birpc.ClientConnector {
	return cS.intRPCconn
}

// CoreS returns the CoreS object.
func (cS *CoreService) CoreS() *cores.CoreS {
	cS.mu.RLock()
	defer cS.mu.RUnlock()
	return cS.cS
}
