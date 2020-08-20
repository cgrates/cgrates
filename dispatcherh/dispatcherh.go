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

package dispatcherh

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewDispatcherHService constructs a DispatcherHService
func NewDispatcherHService(cfg *config.CGRConfig,
	connMgr *engine.ConnManager) (*DispatcherHostsService, error) {
	return &DispatcherHostsService{
		cfg:     cfg,
		connMgr: connMgr,
		stop:    make(chan struct{}),
	}, nil
}

// DispatcherHostsService  is the service handling dispatching towards internal components
// designed to handle automatic partitioning and failover
type DispatcherHostsService struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
	stop    chan struct{}
}

// ListenAndServe will initialize the service
func (dhS *DispatcherHostsService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info("Starting DispatcherH service")
	for {
		if err = dhS.registerHosts(); err != nil {
			return
		}
		select {
		case <-dhS.stop:
			return
		case e := <-exitChan:
			exitChan <- e // put back for the others listening for shutdown request
			return
		case <-time.After(dhS.cfg.DispatcherHCfg().RegisterInterval):
		}
	}
}

// Shutdown is called to shutdown the service
func (dhS *DispatcherHostsService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.DispatcherH))
	dhS.unregisterHosts()
	close(dhS.stop)
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.DispatcherH))
	return nil
}

func (dhS *DispatcherHostsService) registerHosts() (err error) {
	dHs := make([]*engine.DispatcherHost, len(dhS.cfg.DispatcherHCfg().HostIDs))
	for i, hID := range dhS.cfg.DispatcherHCfg().HostIDs {
		tntID := utils.NewTenantID(hID)
		dHs[i] = &engine.DispatcherHost{
			ID:     tntID.ID,
			Tenant: tntID.Tenant,
			Conns:  make([]*config.RemoteHost, 1),
		}
	}
	for _, connID := range dhS.cfg.DispatcherHCfg().DispatchersConns {
		connCfg := dhS.cfg.RPCConns()[connID]
		var conn *config.RemoteHost
		if conn, err = getConnCfg(dhS.cfg, dhS.cfg.DispatcherHCfg().RegisterTransport, connCfg.Conns[0]); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unable to get the connection for<%s> because : %s",
				utils.DispatcherH, connID, err))
			continue
		}
		for _, dh := range dHs {
			dh.Conns[0] = conn
		}
		var rply string
		if err := dhS.connMgr.Call([]string{connID}, nil, utils.DispatcherHv1RegisterHosts, dHs, &rply); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unable to set the hosts to the conn with ID <%s> because : %s",
				utils.DispatcherH, connID, err))
			continue
		} else if rply != utils.OK {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unexpected reply recieved when setting the hosts: %s",
				utils.DispatcherH, rply))
			continue
		}
	}
	return
}

func (dhS *DispatcherHostsService) unregisterHosts() {
	var rply string
	for _, connID := range dhS.cfg.DispatcherHCfg().DispatchersConns {
		if err := dhS.connMgr.Call([]string{connID}, nil, utils.DispatcherHv1UnregisterHosts, dhS.cfg.DispatcherHCfg().HostIDs, &rply); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unable to set the hosts to the conn with ID <%s> because : %s",
				utils.DispatcherH, connID, err))
			continue
		} else if rply != utils.OK {
			utils.Logger.Warning(fmt.Sprintf("<%s> Unexpected reply recieved when setting the hosts: %s",
				utils.DispatcherH, rply))
			continue
		}
	}
}

func (dhS *DispatcherHostsService) Call(_ string, _, _ interface{}) error {
	return utils.ErrNotImplemented
}
