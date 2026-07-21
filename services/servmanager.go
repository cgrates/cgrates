// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func RegisterServiceManagerV1(cfg *config.CGRConfig, srvMngr *servmanager.ServiceManager,
	registry *servmanager.Registry, shutdown *utils.SyncedChan) {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		}, cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	srv, err := newRPCService(apis.NewServiceManagerV1(srvMngr), utils.ServiceManagerV1)
	if err != nil {
		return
	}
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ServiceManager, srv)
}
