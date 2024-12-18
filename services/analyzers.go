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
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig,
	srvIndexer *servmanager.ServiceIndexer) *AnalyzerService {
	anz := &AnalyzerService{
		cfg:        cfg,
		srvIndexer: srvIndexer,
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP}),
	}

	// Wait for AnalyzerService only when it should run.
	if !anz.ShouldRun() {
		close(anz.StateChan(utils.StateServiceUP))
	}

	return anz
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex

	anz *analyzers.AnalyzerS
	cl  *commonlisteners.CommonListenerS

	cancelFunc context.CancelFunc
	cfg        *config.CGRConfig

	intRPCconn birpc.ClientConnector       // share the API object implementing API calls for internal
	srvIndexer *servmanager.ServiceIndexer // access directly services from here
	stateDeps  *StateDependencies          // channel subscriptions for state changes

}

// Start should handle the sercive start
func (anz *AnalyzerService) Start(shutdown chan struct{}) (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cls, err := waitForServiceState(utils.StateServiceUP, utils.CommonListenerS, anz.srvIndexer,
		anz.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	anz.cl = cls.(*CommonListenerService).CLS()

	anz.Lock()
	defer anz.Unlock()
	if anz.anz, err = analyzers.NewAnalyzerS(anz.cfg); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		return
	}
	anzCtx, cancel := context.WithCancel(context.TODO())
	anz.cancelFunc = cancel
	go func(a *analyzers.AnalyzerS) {
		if err := a.ListenAndServe(anzCtx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			close(shutdown)
		}
	}(anz.anz)
	anz.cl.SetAnalyzer(anz.anz)
	go anz.start()
	close(anz.stateDeps.StateChan(utils.StateServiceUP))
	return
}

func (anz *AnalyzerService) start() {
	fs := anz.srvIndexer.GetService(utils.FilterS).(*FilterService)
	if utils.StructChanTimeout(fs.StateChan(utils.StateServiceUP), anz.cfg.GeneralCfg().ConnectTimeout) {
		return
		// return utils.NewServiceStateTimeoutError(utils.AnalyzerS, utils.FilterS, utils.StateServiceUP)
	}

	if !anz.IsRunning() {
		return
	}
	anz.Lock()
	anz.anz.SetFilterS(fs.FilterS())

	srv, _ := engine.NewService(anz.anz)
	// srv, _ := birpc.NewService(apis.NewAnalyzerSv1(anz.anz), "", false)
	if !anz.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			anz.cl.RpcRegister(s)
		}
	}
	anz.Unlock()
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload(_ chan struct{}) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
	anz.cancelFunc()
	anz.cl.SetAnalyzer(nil)
	anz.anz.Shutdown()
	anz.anz = nil
	anz.Unlock()
	anz.cl.RpcUnregisterName(utils.AnalyzerSv1)
	return
}

// IsRunning returns if the service is running
func (anz *AnalyzerService) IsRunning() bool {
	anz.RLock()
	defer anz.RUnlock()
	return anz.anz != nil
}

// ServiceName returns the service name
func (anz *AnalyzerService) ServiceName() string {
	return utils.AnalyzerS
}

// ShouldRun returns if the service should be running
func (anz *AnalyzerService) ShouldRun() bool {
	return anz.cfg.AnalyzerSCfg().Enabled
}

// GetInternalCodec returns the connection wrapped in analyzer connector
func (anz *AnalyzerService) GetInternalCodec(c birpc.ClientConnector, to string) birpc.ClientConnector {
	if !anz.IsRunning() {
		return c
	}
	return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, utils.EmptyString, to)
}

// StateChan returns signaling channel of specific state
func (anz *AnalyzerService) StateChan(stateID string) chan struct{} {
	return anz.stateDeps.StateChan(stateID)
}

// IntRPCConn returns the internal connection used by RPCClient
func (anz *AnalyzerService) IntRPCConn() birpc.ClientConnector {
	return anz.intRPCconn
}
