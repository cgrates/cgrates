// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package registrarc

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
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
	if len(dhS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0 {
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
			if len(dhS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0 {
				rTm = time.NewTimer(dhS.cfg.RegistrarCCfg().RPC.RefreshInterval)
				dhS.registerRPCHosts()
			}
		case <-stopChan:
			if len(dhS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0 {
				rTm.Stop()
			}
			return
		case <-rTm.C:
			dhS.registerRPCHosts()
			rTm.Reset(dhS.cfg.RegistrarCCfg().RPC.RefreshInterval)
		}
	}
}

// Shutdown is called to shutdown the service
func (dhS *RegistrarCService) Shutdown() {
	if len(dhS.cfg.RegistrarCCfg().RPC.RegistrarSConns) != 0 {
		unregisterHosts(dhS.connMgr, dhS.cfg.RegistrarCCfg().RPC,
			dhS.cfg.GeneralCfg().DefaultTenant, utils.RegistrarSv1UnregisterRPCHosts)
	}
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
			if err := dhS.connMgr.Call(context.TODO(), []string{connID}, utils.RegistrarSv1RegisterRPCHosts, args, &rply); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Unable to set the hosts to the conn with ID <%s> because : %s",
					utils.RegistrarC, connID, err))
				continue
			}
		}
	}
}

func unregisterHosts(connMgr *engine.ConnManager, regCfg *config.RegistrarCCfg, dTnt, method string) {
	var rply string
	for _, connID := range regCfg.RegistrarSConns {
		for tnt, hostCfgs := range regCfg.Hosts {
			if tnt == utils.MetaDefault {
				tnt = dTnt
			}
			if err := connMgr.Call(context.TODO(), []string{connID}, method, NewUnregisterArgs(tnt, hostCfgs), &rply); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> Unable to unregister the hosts with tenant<%s> to the conn with ID <%s> because : %s",
					utils.RegistrarC, tnt, connID, err))
			}
		}
	}
}
