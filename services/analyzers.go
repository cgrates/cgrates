// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/analyzers"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, server *cores.Server,
	filterSChan chan *engine.FilterS, shdChan *utils.SyncedChan,
	internalAnalyzerSChan chan birpc.ClientConnector,
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
	connChan chan birpc.ClientConnector
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
	}(anz.anz)
	anz.server.SetAnalyzer(anz.anz)
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
	srv, err := engine.NewService(v1.NewAnalyzerSv1(anz.anz))
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to initialize service, error: <%s>",
			utils.AnalyzerS, err.Error()))
		return
	}
	if !anz.cfg.DispatcherSCfg().Enabled {
		anz.server.RpcRegister(srv)
	}
	anz.connChan <- srv
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
	<-anz.connChan
	anz.Unlock()
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

// GetAnalyzerS returns the analyzer object
func (anz *AnalyzerService) GetAnalyzerS() *analyzers.AnalyzerService {
	return anz.anz
}

// GetInternalCodec returns the connection wrapped in analyzer connector
func (anz *AnalyzerService) GetInternalCodec(c birpc.ClientConnector, to string) birpc.ClientConnector {
	if !anz.IsRunning() {
		return c
	}
	return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, utils.EmptyString, to)
}
