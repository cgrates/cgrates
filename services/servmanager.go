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
	registry *servmanager.Registry, shutdown *utils.SyncedChan) error {
	srvDeps, err := registry.WaitForServices(shutdown, utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		}, cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return err
	}
	srv, err := newRPCService(apis.NewServiceManagerV1(srvMngr), utils.ServiceManagerV1)
	if err != nil {
		return err
	}

	cms := srvDeps[utils.ConnManager].(*ConnManagerService)
	cl := srvDeps[utils.CommonListenerS].(*CommonListenerService).CLS()
	if err := cl.RpcRegister(srv); err != nil {
		return err
	}
	cms.AddInternalConn(utils.ServiceManager, srv)
	return nil
}
