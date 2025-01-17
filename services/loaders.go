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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
		stateDeps:  NewStateDependencies([]string{utils.StateServiceUP, utils.StateServiceDOWN}),
	}
}

// LoaderService implements Service interface
type LoaderService struct {
	sync.RWMutex
	cfg        *config.CGRConfig
	ldrs       *loaders.LoaderS
	cl         *commonlisteners.CommonListenerS
	preloadIDs []string
	stopChan   chan struct{}
	stateDeps  *StateDependencies // channel subscriptions for state changes
}

// Start should handle the service start
func (s *LoaderService) Start(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	s.cl = srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DataDB].(*DataDBService).DataManager()

	s.Lock()
	defer s.Unlock()

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
	srv, _ := engine.NewService(s.ldrs)
	// srv, _ := birpc.NewService(apis.NewLoaderSv1(ldrs.ldrs), "", false)
	for _, svc := range srv {
		s.cl.RpcRegister(svc)
	}
	cms.AddInternalConn(utils.LoaderS, srv)
	return nil
}

// Reload handles the change of config
func (s *LoaderService) Reload(_ *utils.SyncedChan, registry *servmanager.ServiceRegistry) error {
	srvDeps, err := WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.ConnManager,
			utils.FilterS,
			utils.DataDB,
		},
		registry, s.cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	cms := srvDeps[utils.ConnManager].(*ConnManagerService).ConnManager()
	fs := srvDeps[utils.FilterS].(*FilterService).FilterS()
	dbs := srvDeps[utils.DataDB].(*DataDBService).DataManager()
	close(s.stopChan)
	s.stopChan = make(chan struct{})

	s.RLock()
	defer s.RUnlock()

	s.ldrs.Reload(dbs, fs, cms)
	return s.ldrs.ListenAndServe(s.stopChan)
}

// Shutdown stops the service
func (s *LoaderService) Shutdown(_ *servmanager.ServiceRegistry) (_ error) {
	s.Lock()
	s.ldrs = nil
	close(s.stopChan)
	s.cl.RpcUnregisterName(utils.LoaderSv1)
	s.Unlock()
	return
}

// ServiceName returns the service name
func (s *LoaderService) ServiceName() string {
	return utils.LoaderS
}

// ShouldRun returns if the service should be running
func (s *LoaderService) ShouldRun() bool {
	return s.cfg.LoaderCfg().Enabled()
}

// GetLoaderS returns the initialized LoaderService
func (s *LoaderService) GetLoaderS() *loaders.LoaderS {
	return s.ldrs
}

// StateChan returns signaling channel of specific state
func (s *LoaderService) StateChan(stateID string) chan struct{} {
	return s.stateDeps.StateChan(stateID)
}
