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

package registrarc

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewRegistrarCService constructs a DispatcherHService
func NewRegistrarCService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) *RegistrarCService {
	return &RegistrarCService{
		cfg:     cfg,
		connMgr: connMgr,
	}
}

// RegistrarCService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type RegistrarCService struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
}

// ListenAndServe will initialize the service
func (dhS *RegistrarCService) ListenAndServe(stopChan, rldChan <-chan struct{}) {
	dTm, rTm := &time.Timer{}, &time.Timer{}
	var dTmStarted, rTmStarted bool
	if dTmStarted = dhS.cfg.RegistrarCCfg().Dispatcher.Enabled; dTmStarted {
		dTm = time.NewTimer(dhS.cfg.RegistrarCCfg().Dispatcher.RefreshInterval)
		dhS.registerDispHosts()
	}
	if rTmStarted = dhS.cfg.RegistrarCCfg().RPC.Enabled; rTmStarted {
		rTm = time.NewTimer(dhS.cfg.RegistrarCCfg().RPC.RefreshInterval)
		dhS.registerRPCHosts()
	}
	for {
		select {
		case <-rldChan:
			if rTmStarted {
				rTm.Stop()
			}
			if dTmStarted {
				dTm.Stop()
			}
			if dTmStarted = dhS.cfg.RegistrarCCfg().Dispatcher.Enabled; dTmStarted {
				dTm = time.NewTimer(dhS.cfg.RegistrarCCfg().Dispatcher.RefreshInterval)
				dhS.registerDispHosts()
			}
			if rTmStarted = dhS.cfg.RegistrarCCfg().RPC.Enabled; rTmStarted {
				rTm = time.NewTimer(dhS.cfg.RegistrarCCfg().RPC.RefreshInterval)
				dhS.registerRPCHosts()
			}
		case <-stopChan:
			if dhS.cfg.RegistrarCCfg().Dispatcher.Enabled {
				dTm.Stop()
			}
			if dhS.cfg.RegistrarCCfg().RPC.Enabled {
				rTm.Stop()
			}
			return
		case <-dTm.C:
			dhS.registerDispHosts()
			dTm.Reset(dhS.cfg.RegistrarCCfg().Dispatcher.RefreshInterval)
		case <-rTm.C:
			dhS.registerRPCHosts()
			rTm.Reset(dhS.cfg.RegistrarCCfg().RPC.RefreshInterval)
		}
	}
}

// Shutdown is called to shutdown the service
func (dhS *RegistrarCService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.RegistrarC))
	if dhS.cfg.RegistrarCCfg().Dispatcher.Enabled {
		unregisterHosts(dhS.connMgr, dhS.cfg.RegistrarCCfg().Dispatcher,
			dhS.cfg.GeneralCfg().DefaultTenant, utils.RegistrarSv1UnregisterDispatcherHosts)
	}
	if dhS.cfg.RegistrarCCfg().RPC.Enabled {
		unregisterHosts(dhS.connMgr, dhS.cfg.RegistrarCCfg().RPC,
			dhS.cfg.GeneralCfg().DefaultTenant, utils.RegistrarSv1UnregisterRPCHosts)
	}
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.RegistrarC))
}

func (dhS *RegistrarCService) registerDispHosts() {
	for _, connID := range dhS.cfg.RegistrarCCfg().Dispatcher.RegistrarSConns {
		for tnt, hostCfgs := range dhS.cfg.RegistrarCCfg().Dispatcher.Hosts {
			if tnt == utils.MetaDefault {
				tnt = dhS.cfg.GeneralCfg().DefaultTenant
			}
			args, err := NewRegisterArgs(dhS.cfg, tnt, hostCfgs)
			if err != nil {
				continue
			}
			var rply string
			if err := dhS.connMgr.Call([]string{connID}, nil, utils.RegistrarSv1RegisterDispatcherHosts, args, &rply); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Unable to set the hosts to the conn with ID <%s> because : %s",
					utils.RegistrarC, connID, err))
				continue
			}
		}
	}
	return
}

func (dhS *RegistrarCService) registerRPCHosts() {
	for _, connID := range dhS.cfg.RegistrarCCfg().RPC.RegistrarSConns {
		for tnt, hostCfgs := range dhS.cfg.RegistrarCCfg().RPC.Hosts {
			if tnt == utils.MetaDefault {
				tnt = dhS.cfg.GeneralCfg().DefaultTenant
			}
			args, err := NewRegisterArgs(dhS.cfg, tnt, hostCfgs)
			if err != nil {
				continue
			}
			var rply string
			if err := dhS.connMgr.Call([]string{connID}, nil, utils.RegistrarSv1RegisterRPCHosts, args, &rply); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Unable to set the hosts to the conn with ID <%s> because : %s",
					utils.RegistrarC, connID, err))
				continue
			}
		}
	}
	return
}

func unregisterHosts(connMgr *engine.ConnManager, regCfg *config.RegistrarCCfg, dTnt, method string) {
	var rply string
	for _, connID := range regCfg.RegistrarSConns {
		for tnt, hostCfgs := range regCfg.Hosts {
			if tnt == utils.MetaDefault {
				tnt = dTnt
			}
			if err := connMgr.Call([]string{connID}, nil, method, NewUnregisterArgs(tnt, hostCfgs), &rply); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Unable to unregister the hosts with tenant<%s> to the conn with ID <%s> because : %s",
					utils.RegistrarC, tnt, connID, err))
			}
		}
	}
}
