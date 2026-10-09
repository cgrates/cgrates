// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService returns the Analyzer Service
func NewAnalyzerService(cfg *config.CGRConfig) *AnalyzerService {
	anz := &AnalyzerService{
		cfg: cfg,
	}
	return anz
}

// AnalyzerService implements Service interface
type AnalyzerService struct {
	mu         sync.RWMutex
	cfg        *config.CGRConfig
	anz        *analyzers.AnalyzerS
	cancelFunc context.CancelFunc
}

// Start should handle the sercive start
func (anz *AnalyzerService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	cls, err := registry.WaitForService(shutdown, utils.CommonListenerS, utils.StateServiceUP,
		anz.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := cls.(*CommonListenerService).CLS()

	anz.mu.Lock()
	defer anz.mu.Unlock()
	a, err := analyzers.NewAnalyzerS(anz.cfg)
	if err != nil {
		return err
	}
	anz.anz = a
	anzCtx, cancel := context.WithCancel(context.TODO())
	anz.cancelFunc = cancel
	go func() {
		if err := a.ListenAndServe(anzCtx); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
			shutdown.CloseOnce()
		}
	}()
	cl.SetAnalyzer(a)
	go anz.registerRPC(shutdown, registry, cl, a)
	return nil
}

func (anz *AnalyzerService) registerRPC(shutdown *utils.SyncedChan, registry *servmanager.Registry,
	cl *commonlisteners.CommonListenerS, a *analyzers.AnalyzerS) {
	fs, err := registry.WaitForService(shutdown, utils.FilterS, utils.StateServiceUP,
		anz.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	anz.mu.Lock()
	defer anz.mu.Unlock()
	if anz.anz != a {
		return
	}
	a.SetFilterS(fs.(*FilterService).FilterS())

	srv, err := newRPCService(apis.NewAnalyzerSv1(a), utils.AnalyzerSv1)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error registering RPC service: %s", utils.AnalyzerS, err))
		shutdown.CloseOnce()
		return
	}
	_ = cl.RpcRegister(srv)
}

// Reload handles the change of config
func (anz *AnalyzerService) Reload(_ *utils.SyncedChan, _ *servmanager.Registry) (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (anz *AnalyzerService) Shutdown(registry *servmanager.Registry) (err error) {
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	anz.mu.Lock()
	anz.cancelFunc()
	cl.SetAnalyzer(nil)
	anz.anz.Shutdown()
	anz.anz = nil
	anz.mu.Unlock()
	cl.RpcUnregisterName(utils.AnalyzerSv1)
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

// GetInternalCodec wraps the provided ClientConnector when an analyzer exists.
// Otherwise, it returns the original connector unchanged.
func (anz *AnalyzerService) GetInternalCodec(c birpc.ClientConnector, to string) birpc.ClientConnector {
	anz.mu.RLock()
	defer anz.mu.RUnlock()
	if anz.anz == nil {
		return c
	}
	return anz.anz.NewAnalyzerConnector(c, utils.MetaInternal, "", to)
}
