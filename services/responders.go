// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewResponderService returns the Resonder Service
func NewResponderService(cfg *config.CGRConfig, server *cores.Server,
	internalRALsChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan, anz *AnalyzerService,
	srvDep map[string]*sync.WaitGroup,
	filterSCh chan *engine.FilterS) *ResponderService {
	return &ResponderService{
		connChan:  internalRALsChan,
		cfg:       cfg,
		server:    server,
		shdChan:   shdChan,
		anz:       anz,
		srvDep:    srvDep,
		filterSCh: filterSCh,
		syncChans: make(map[string]chan *engine.Responder),
	}
}

// ResponderService implements Service interface
// this service is manged by the RALs as a component
type ResponderService struct {
	sync.RWMutex
	cfg     *config.CGRConfig
	server  *cores.Server
	shdChan *utils.SyncedChan

	resp      *engine.Responder
	connChan  chan birpc.ClientConnector
	anz       *AnalyzerService
	srvDep    map[string]*sync.WaitGroup
	syncChans map[string]chan *engine.Responder

	filterSCh chan *engine.FilterS
}

// Start should handle the sercive start
// For this service the start should be called from RAL Service
func (resp *ResponderService) Start() error {
	if resp.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	filterS := <-resp.filterSCh
	resp.filterSCh <- filterS

	resp.Lock()
	defer resp.Unlock()
	resp.resp = &engine.Responder{
		FilterS:          filterS,
		ShdChan:          resp.shdChan,
		MaxComputedUsage: resp.cfg.RalsCfg().MaxComputedUsage,
	}

	srv, err := engine.NewService(resp.resp)
	if err != nil {
		return err
	}
	if !resp.cfg.DispatcherSCfg().Enabled {
		resp.server.RpcRegister(srv)
	}

	resp.connChan <- resp.anz.GetInternalCodec(srv, utils.ResponderS) // Rater done
	resp.sync()
	return nil
}

// Reload handles the change of config
func (resp *ResponderService) Reload() (err error) {
	resp.Lock()
	resp.resp.SetMaxComputedUsage(resp.cfg.RalsCfg().MaxComputedUsage)
	resp.Unlock()
	return
}

// Shutdown stops the service
func (resp *ResponderService) Shutdown() (err error) {
	resp.Lock()
	resp.resp = nil
	<-resp.connChan
	for _, c := range resp.syncChans {
		c <- nil // just tell the services that responder is nil
	}
	resp.Unlock()
	return
}

// IsRunning returns if the service is running
func (resp *ResponderService) IsRunning() bool {
	resp.RLock()
	defer resp.RUnlock()
	return resp.isRunning()
}

func (resp *ResponderService) isRunning() bool {
	return resp != nil && resp.resp != nil
}

// ServiceName returns the service name
func (resp *ResponderService) ServiceName() string {
	return utils.ResponderS
}

// GetResponder returns the responder created
func (resp *ResponderService) GetResponder() *engine.Responder {
	resp.RLock()
	defer resp.RUnlock()
	return resp.resp
}

// ShouldRun returns if the service should be running
func (resp *ResponderService) ShouldRun() bool {
	return resp.cfg.RalsCfg().Enabled
}

// RegisterSyncChan used by dependent subsystems to register a channel to reload only the responder(thread safe)
func (resp *ResponderService) RegisterSyncChan(srv string, c chan *engine.Responder) {
	resp.Lock()
	resp.syncChans[srv] = c
	if resp.isRunning() {
		c <- resp.resp
	}
	resp.Unlock()
}

// UnregisterSyncChan used by dependent subsystems to unregister a channel
func (resp *ResponderService) UnregisterSyncChan(srv string) {
	resp.Lock()
	c, has := resp.syncChans[srv]
	if has {
		close(c)
		delete(resp.syncChans, srv)
	}
	resp.Unlock()
}

// sync sends the responder over syncChansv (not thread safe)
func (resp *ResponderService) sync() {
	if resp.isRunning() {
		for _, c := range resp.syncChans {
			c <- resp.resp
		}
	}
}
