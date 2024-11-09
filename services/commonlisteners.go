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
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/utils"
)

// NewCommonListenerService instantiates a new CommonListenerService.
func NewCommonListenerService(cfg *config.CGRConfig, caps *engine.Caps, srvDep map[string]*sync.WaitGroup) *CommonListenerService {
	return &CommonListenerService{
		cfg:    cfg,
		caps:   caps,
		clsCh:  make(chan *commonlisteners.CommonListenerS, 1),
		srvDep: srvDep,
	}
}

// CommonListenerService implements Service interface.
type CommonListenerService struct {
	mu sync.RWMutex

	cls *commonlisteners.CommonListenerS

	clsCh  chan *commonlisteners.CommonListenerS
	caps   *engine.Caps
	cfg    *config.CGRConfig
	srvDep map[string]*sync.WaitGroup
}

// Start handles the service start.
func (cl *CommonListenerService) Start(*context.Context, context.CancelFunc) error {
	if cl.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.cls = commonlisteners.NewCommonListenerS(cl.caps)
	cl.clsCh <- cl.cls
	if len(cl.cfg.HTTPCfg().RegistrarSURL) != 0 {
		cl.cls.RegisterHTTPFunc(cl.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if cl.cfg.ConfigSCfg().Enabled {
		cl.cls.RegisterHTTPFunc(cl.cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	return nil
}

// Reload handles the config changes.
func (cl *CommonListenerService) Reload(*context.Context, context.CancelFunc) error {
	return nil
}

// Shutdown stops the service.
func (cl *CommonListenerService) Shutdown() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.cls = nil
	<-cl.clsCh
	return nil
}

// IsRunning returns whether the service is running or not.
func (cl *CommonListenerService) IsRunning() bool {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.cls != nil
}

// ServiceName returns the service name
func (cl *CommonListenerService) ServiceName() string {
	return utils.CommonListenerS
}

// ShouldRun returns if the service should be running.
func (cl *CommonListenerService) ShouldRun() bool {
	return true
}

// WaitForCLS waits for the CommonListenerS structure to be initialized.
func (cl *CommonListenerService) WaitForCLS(ctx *context.Context) (*commonlisteners.CommonListenerS, error) {
	cl.mu.RLock()
	clsCh := cl.clsCh
	cl.mu.RUnlock()
	var cls *commonlisteners.CommonListenerS
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case cls = <-clsCh:
		clsCh <- cls
	}
	return cls, nil
}
