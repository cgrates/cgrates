// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// NewLoaderService returns the Loader Service
func NewLoaderService(cfg *config.CGRConfig, preloadIDs []string) *LoaderService {
	return &LoaderService{
		cfg:        cfg,
		stopChan:   make(chan struct{}),
		preloadIDs: preloadIDs,
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	mu         sync.RWMutex
	cfg        *config.CGRConfig
	ldrs       *loaders.LoaderS
	preloadIDs []string
	stopChan   chan struct{}
}

// Start should handle the service start
func (s *LoaderService) Start(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ldrs = loaders.NewLoaderS(s.cfg, dbs, fs, cms.ConnManager())
	if !s.ldrs.Enabled() {
		return nil
	}

	var reply string
	for _, loaderID := range s.preloadIDs {
		if err = s.ldrs.V1Run(context.TODO(),
			&loaders.ArgsProcessFolder{
				APIOpts: map[string]any{
					utils.MetaForceLock:   true,
					utils.MetaStopOnError: true,
				}, LoaderID: loaderID,
			}, &reply); err != nil {
			return fmt.Errorf("could not preload loader with ID %q: %v", loaderID, err)
		}
	}

	if err := s.ldrs.ListenAndServe(s.stopChan); err != nil {
		return err
	}
	srv, err := newRPCService(apis.NewLoaderSv1(s.ldrs), utils.LoaderSv1)
	if err != nil {
		return err
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.LoaderS, srv)
	return nil
}

// Reload handles the change of config
func (s *LoaderService) Reload(shutdown *utils.SyncedChan, registry *servmanager.Registry) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.FilterS,
			utils.DB,
		},
		s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DB].(*DBService).DataManager()
	close(s.stopChan)
	s.stopChan = make(chan struct{})

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.ldrs.Reload(dbs, fs, cms)
	return s.ldrs.ListenAndServe(s.stopChan)
}

// Shutdown stops the service
func (s *LoaderService) Shutdown(registry *servmanager.Registry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.stopChan)
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.RpcUnregisterName(utils.LoaderSv1)
	return nil
}

// ServiceName returns the service name
func (s *LoaderService) ServiceName() string {
	return utils.LoaderS
}

// ShouldRun returns if the service should be running
func (s *LoaderService) ShouldRun() bool {
	return s.cfg.LoaderCfg().Enabled()
}
