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

	"github.com/cgrates/cgrates/analyzers"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, server *cores.Server,
	filterSChan chan *engine.FilterS, shdChan *utils.SyncedChan,
	internalAnalyzerSChan chan rpcclient.ClientConnector,
	srvDep map[string]*sync.WaitGroup) *AnalyzerService {
	return &AnalyzerService{
		connChan:    internalAnalyzerSChan,
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		srvDep:      srvDep,
	}
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	server      *cores.Server
	filterSChan chan *engine.FilterS
	stopChan    chan struct{}
	shdChan     *utils.SyncedChan

	anz      *analyzers.AnalyzerService
	rpc      *v1.AnalyzerSv1
	connChan chan rpcclient.ClientConnector
	srvDep   map[string]*sync.WaitGroup
}

// Start should handle the sercive start
func (anz *AnalyzerService) Start() (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	anz.Lock()
	defer anz.Unlock()
	if anz.anz, err = analyzers.NewAnalyzerService(anz.cfg); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		return
	}
	anz.stopChan = make(chan struct{})
	go func(a *analyzers.AnalyzerService) {
		if err := a.ListenAndServe(anz.stopChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			anz.shdChan.CloseOnce()
		}
		return
	}(anz.anz)
	anz.server.SetAnalyzer(anz.anz)
	anz.rpc = v1.NewAnalyzerSv1(anz.anz)
	go anz.start()
	return
}

func (anz *AnalyzerService) start() {
	var fS *engine.FilterS
	select {
	case <-anz.stopChan:
		return
	case fS = <-anz.filterSChan:
		if !anz.IsRunning() {
			return
		}
		anz.Lock()
		defer anz.Unlock()
		anz.filterSChan <- fS
		anz.anz.SetFilterS(fS)
	}
	if !anz.cfg.DispatcherSCfg().Enabled {
		anz.server.RpcRegister(anz.rpc)
	}
	anz.connChan <- anz.rpc
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
	close(anz.stopChan)
	anz.server.SetAnalyzer(nil)
	anz.anz.Shutdown()
	anz.anz = nil
	anz.rpc = nil
	<-anz.connChan
	anz.Unlock()
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

// GetAnalyzerS returns the analyzer object
func (anz *AnalyzerService) GetAnalyzerS() *analyzers.AnalyzerService {
	return anz.anz
}

// GetInternalCodec returns the connection wrapped in analyzer connector
func (anz *AnalyzerService) GetInternalCodec(c rpcclient.ClientConnector, to string) rpcclient.ClientConnector {
	if !anz.IsRunning() {
		return c
	}
	return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, utils.EmptyString, to)
}

// GetInternalCodec returns the connection wrapped in analyzer connector
func (anz *AnalyzerService) GetInternalBiRPCCodec(c rpcclient.BiRPCConector, to string) rpcclient.BiRPCConector {
	if !anz.IsRunning() {
		return c
	}
	return anz.anz.NewAnalyzerBiRPCConnector(c, rpcclient.BiRPCInternal, utils.EmptyString, to)
}
