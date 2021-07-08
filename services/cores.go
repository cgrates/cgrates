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
	"io"
	"sync"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, server *cores.Server,
	internalCoreSChan chan rpcclient.ClientConnector, anz *AnalyzerService,
	fileCpu io.Closer, fileMEM string, shdWg *sync.WaitGroup, stopMemPrf chan struct{},
	shdChan *utils.SyncedChan, srvDep map[string]*sync.WaitGroup) *CoreService {
	return &CoreService{
		shdChan:    shdChan,
		shdWg:      shdWg,
		stopMemPrf: stopMemPrf,
		connChan:   internalCoreSChan,
		cfg:        cfg,
		caps:       caps,
		fileCpu:    fileCpu,
		fileMem:    fileMEM,
		server:     server,
		anz:        anz,
		srvDep:     srvDep,
	}
}

// CoreService implements Service interface
type CoreService struct {
	sync.RWMutex
	cfg        *config.CGRConfig
	server     *cores.Server
	caps       *engine.Caps
	stopChan   chan struct{}
	shdWg      *sync.WaitGroup
	stopMemPrf chan struct{}
	shdChan    *utils.SyncedChan
	fileCpu    io.Closer
	fileMem    string
	cS         *cores.CoreService
	rpc        *v1.CoreSv1
	connChan   chan rpcclient.ClientConnector
	anz        *AnalyzerService
	srvDep     map[string]*sync.WaitGroup
}

// Start should handle the service start
func (cS *CoreService) Start() (err error) {
	if cS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cS.Lock()
	defer cS.Unlock()
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CoreS))
	cS.stopChan = make(chan struct{})
	cS.cS = cores.NewCoreService(cS.cfg, cS.caps, cS.fileCpu, cS.fileMem, cS.stopChan, cS.shdWg, cS.stopMemPrf, cS.shdChan)
	cS.rpc = v1.NewCoreSv1(cS.cS)
	if !cS.cfg.DispatcherSCfg().Enabled {
		cS.server.RpcRegister(cS.rpc)
	}
	cS.connChan <- cS.anz.GetInternalCodec(cS.rpc, utils.CoreS)
	return
}

// Reload handles the change of config
func (cS *CoreService) Reload() (err error) {
	return
}

// Shutdown stops the service
func (cS *CoreService) Shutdown() (err error) {
	cS.Lock()
	defer cS.Unlock()
	cS.cS.Shutdown()
	close(cS.stopChan)
	cS.cS = nil
	cS.rpc = nil
	<-cS.connChan
	return
}

// IsRunning returns if the service is running
func (cS *CoreService) IsRunning() bool {
	cS.RLock()
	defer cS.RUnlock()
	return cS != nil && cS.cS != nil
}

// ServiceName returns the service name
func (cS *CoreService) ServiceName() string {
	return utils.CoreS
}

// ShouldRun returns if the service should be running
func (cS *CoreService) ShouldRun() bool {
	return true
}

// GetCoreS returns the coreS
func (cS *CoreService) GetCoreS() *cores.CoreService {
	cS.RLock()
	defer cS.RUnlock()
	return cS.cS
}
