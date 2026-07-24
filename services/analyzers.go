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
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig, server *utils.Server, exitChan chan bool,
	internalAnalyzerSChan chan birpc.ClientConnector) servmanager.Service {
	return &AnalyzerService{
		connChan: internalAnalyzerSChan,
		cfg:      cfg,
		server:   server,
		exitChan: exitChan,
	}
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	server   *utils.Server
	exitChan chan bool

	anz      *analyzers.AnalyzerService
	rpc      *v1.AnalyzerSv1
	connChan chan birpc.ClientConnector
}

// Start should handle the sercive start
func (anz *AnalyzerService) Start() (err error) {
	if anz.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	if anz.anz, err = analyzers.NewAnalyzerService(); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		anz.exitChan <- true
		return
	}
	go func() {
		if err := anz.anz.ListenAndServe(anz.exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
		}
		anz.anz.Shutdown()
		anz.exitChan <- true
		return
	}()
	anz.rpc = v1.NewAnalyzerSv1(anz.anz)
	if !anz.cfg.DispatcherSCfg().Enabled {
		anz.server.RpcRegister(anz.rpc)
	}
	anz.connChan <- anz.rpc

	return
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload() (err error) {
	return // for the moment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown() (err error) {
	anz.Lock()
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
