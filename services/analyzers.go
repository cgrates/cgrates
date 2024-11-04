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
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, server *commonlisteners.CommonListenerS,
	filterSChan chan *engine.FilterS,
	internalAnalyzerSChan chan birpc.ClientConnector,
	srvDep map[string]*sync.WaitGroup) *AnalyzerService {
	return &AnalyzerService{
		connChan:    internalAnalyzerSChan,
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		srvDep:      srvDep,
	}
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	server      *commonlisteners.CommonListenerS
	filterSChan chan *engine.FilterS
	ctx         *context.Context
	cancelFunc  context.CancelFunc

	anz      *analyzers.AnalyzerS
	connChan chan birpc.ClientConnector
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (anz *AnalyzerService) Start(ctx *context.Context, shtDwn context.CancelFunc) (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	anz.Lock()
	defer anz.Unlock()
	if anz.anz, err = analyzers.NewAnalyzerService(anz.cfg); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		return
	}
	anz.ctx, anz.cancelFunc = context.WithCancel(ctx)
	go func(a *analyzers.AnalyzerS) {
		if err := a.ListenAndServe(anz.ctx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			shtDwn()
		}
	}(anz.anz)
	anz.server.SetAnalyzer(anz.anz)
	go anz.start(ctx)
	return
}

func (anz *AnalyzerService) start(ctx *context.Context) {
	fS, err := waitForFilterS(ctx, anz.filterSChan)
	if err != nil {
		anz.connChan <- nil
		return
	}

	if !anz.IsRunning() {
		return
	}
	anz.Lock()
	anz.anz.SetFilterS(fS)

	srv, _ := engine.NewService(anz.anz)
	// srv, _ := birpc.NewService(apis.NewAnalyzerSv1(anz.anz), "", false)
	if !anz.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			anz.server.RpcRegister(s)
		}
	}
	anz.Unlock()
	anz.connChan <- anz.GetInternalCodec(srv, utils.AnalyzerS)
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload(*context.Context, context.CancelFunc) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
	anz.cancelFunc()
	anz.server.SetAnalyzer(nil)
	anz.anz.Shutdown()
	anz.anz = nil
	<-anz.connChan
	anz.Unlock()
	anz.server.RpcUnregisterName(utils.AnalyzerSv1)
	return
}

// IsRunning returns if the service is running
func (anz *AnalyzerService) IsRunning() bool {
	anz.RLock()
	defer anz.RUnlock()
	return anz != nil && anz.anz != nil
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
