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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"
)

// NewConnection returns a new connection
func NewConnection(cfg *config.CGRConfig, serviceConnChan, dispatcherSChan chan rpcclient.RpcClientConnection, conns []*config.RemoteHost) (rpcclient.RpcClientConnection, error) {
	if len(conns) == 0 {
		return nil, nil
	}
	internalChan := serviceConnChan
	if cfg.DispatcherSCfg().Enabled {
		internalChan = dispatcherSChan
	}
	return engine.NewRPCPool(rpcclient.POOL_FIRST,
		cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		conns, internalChan, false)
}
