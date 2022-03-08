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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewCoreService returns the Core Service
func NewCoreService(cfg *config.CGRConfig, caps *engine.Caps, server *cores.Server,
	internalCoreSChan chan birpc.ClientConnector, anz *AnalyzerService,
	fileCpu io.Closer, fileMEM string, stopMemPrf chan struct{},
	shdWg *sync.WaitGroup, srvDep map[string]*sync.WaitGroup) *CoreService {
	return &CoreService{
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
		csCh:       make(chan *cores.CoreS, 1),
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
	fileCpu    io.Closer
	fileMem    string
	cS         *cores.CoreS
	connChan   chan birpc.ClientConnector
	anz        *AnalyzerService
	srvDep     map[string]*sync.WaitGroup
	csCh       chan *cores.CoreS
}

// Start should handle the service start
func (cS *CoreService) Start(_ *context.Context, shtDw context.CancelFunc) (_ error) {
	if cS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	cS.Lock()
	defer cS.Unlock()
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CoreS))
	cS.stopChan = make(chan struct{})
	cS.cS = cores.NewCoreService(cS.cfg, cS.caps, cS.fileCpu, cS.fileMem, cS.stopChan, cS.stopMemPrf, cS.shdWg, shtDw)
	cS.csCh <- cS.cS
	srv, _ := engine.NewService(cS.cS)
	// srv, _ := birpc.NewService(apis.NewCoreSv1(cS.cS), utils.EmptyString, false)
	if !cS.cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cS.server.RpcRegister(s)
		}
	}
	cS.connChan <- cS.anz.GetInternalCodec(srv, utils.CoreS)
	return
}

// Reload handles the change of config
func (cS *CoreService) Reload(*context.Context, context.CancelFunc) error {
	return nil
}

// Shutdown stops the service
func (cS *CoreService) Shutdown() (_ error) {
	cS.Lock()
	defer cS.Unlock()
	cS.cS.Shutdown()
	close(cS.stopChan)
	cS.cS = nil
	<-cS.connChan
	<-cS.csCh
	cS.server.RpcUnregisterName(utils.CoreSv1)
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
func (cS *CoreService) WaitForCoreS(ctx *context.Context) (cs *cores.CoreS, err error) {
	cS.RLock()
	cSCh := cS.csCh
	cS.RUnlock()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case cs = <-cSCh:
		cSCh <- cs
	}
	return
}
