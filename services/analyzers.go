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
func NewAnalyzerService(cfg *config.CGRConfig) *AnalyzerService {
	anz := &AnalyzerService{
		cfg:       cfg,
		stateDeps: NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
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

	intRPCconn birpc.ClientConnector // share the API object implementing API calls for internal
	stateDeps  *StateDependencies    // channel subscriptions for state changes

}

// Start should handle the sercive start
func (anz *AnalyzerService) Start(shutdown *utils.SyncedChan, registry *servmanager.ServiceRegistry) (err error) {
	cls, err := waitForServiceState(utils.StateServiceUP, utils.CommonListenerS, registry,
		anz.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	anz.cl = cls.(*CommonListenerService).CLS()

	anz.Lock()
	defer anz.Unlock()
	if anz.anz, err = analyzers.NewAnalyzerS(anz.cfg); err != nil {
		return
	}
	anzCtx, cancel := context.WithCancel(context.TODO())
	anz.cancelFunc = cancel
	go func(a *analyzers.AnalyzerS) {
		if err := a.ListenAndServe(anzCtx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			shutdown.CloseOnce()
		}
	}(anz.anz)
	anz.cl.SetAnalyzer(anz.anz)
	go anz.start(registry)
	return
}

func (anz *AnalyzerService) start(registry *servmanager.ServiceRegistry) {
	fs, err := waitForServiceState(utils.StateServiceUP, utils.FilterS, registry,
		anz.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	anz.Lock()
	anz.anz.SetFilterS(fs.(*FilterService).FilterS())

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
func (anz *AnalyzerService) Reload(_ *utils.SyncedChan, _ *servmanager.ServiceRegistry) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown(_ *servmanager.ServiceRegistry) (err error) {
	anz.Lock()
	anz.cancelFunc()
	anz.cl.SetAnalyzer(nil)
	anz.anz.Shutdown()
	anz.anz = nil
	anz.Unlock()
	anz.cl.RpcUnregisterName(utils.AnalyzerSv1)
	return
}

// ServiceName returns the service name
func (anz *AnalyzerService) ServiceName() string {
	return utils.AnalyzerS
}

// ShouldRun returns if the service should be running
func (anz *AnalyzerService) ShouldRun() bool {
	return anz.cfg.AnalyzerSCfg().Enabled
}

// GetInternalCodec wraps the provided ClientConnector in an analyzer connector
// if the analyzer service should run. Otherwise, it returns the original connector
// unchanged.
func (anz *AnalyzerService) GetInternalCodec(c birpc.ClientConnector, to string) birpc.ClientConnector {
	if !anz.ShouldRun() {
		// It's enough to check the result of ShouldRun as other
		// services calling GetInternalCodec had already waited for
		// AnalyzerService to be initiated/started.
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
