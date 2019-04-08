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

package engine

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewRPCPool(dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan rpcclient.RpcClientConnection, ttl time.Duration, lazyConnect bool) (*rpcclient.RpcClientPool, error) {
	var rpcClient *rpcclient.RpcClient
	var err error
	rpcPool := rpcclient.NewRpcClientPool(dispatchStrategy, replyTimeout)
	atLestOneConnected := false // If one connected we don't longer return errors
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.MetaInternal {
			rpcClient, err = rpcclient.NewRpcClient("", "", rpcConnCfg.TLS, keyPath, certPath, caPath, connAttempts,
				reconnects, connectTimeout, replyTimeout, rpcclient.INTERNAL_RPC, internalConnChan, lazyConnect)
		} else if utils.IsSliceMember([]string{utils.MetaJSONrpc, utils.MetaGOBrpc, ""}, rpcConnCfg.Transport) {
			codec := utils.GOB
			if rpcConnCfg.Transport != "" {
				codec = rpcConnCfg.Transport[1:] // Transport contains always * before codec understood by rpcclient
			}
			rpcClient, err = rpcclient.NewRpcClient("tcp", rpcConnCfg.Address, rpcConnCfg.TLS, keyPath, certPath, caPath,
				connAttempts, reconnects, connectTimeout, replyTimeout, codec, nil, lazyConnect)
		} else {
			return nil, fmt.Errorf("Unsupported transport: <%s>", rpcConnCfg.Transport)
		}
		if err == nil {
			atLestOneConnected = true
		}
		rpcPool.AddClient(rpcClient)
	}
	if atLestOneConnected {
		err = nil
	}
	return rpcPool, err
}

var IntRPC *rpcclient.RPCClientSet

func InitInternalRPC() {
	if !config.CgrConfig().DispatcherSCfg().Enabled {
		return
	}
	var subsystems []string
	subsystems = append(subsystems, utils.CacheSv1)
	subsystems = append(subsystems, utils.GuardianSv1)
	subsystems = append(subsystems, utils.SchedulerSv1)
	subsystems = append(subsystems, utils.LoaderSv1)
	if config.CgrConfig().AttributeSCfg().Enabled {
		subsystems = append(subsystems, utils.AttributeSv1)
	}
	if config.CgrConfig().RalsCfg().RALsEnabled {
		subsystems = append(subsystems, utils.ApierV1)
		subsystems = append(subsystems, utils.ApierV2)
		subsystems = append(subsystems, utils.Responder)
	}
	if config.CgrConfig().CdrsCfg().CDRSEnabled {
		subsystems = append(subsystems, utils.CDRsV1)
		subsystems = append(subsystems, utils.CDRsV2)
	}
	if config.CgrConfig().AnalyzerSCfg().Enabled {
		subsystems = append(subsystems, utils.AnalyzerSv1)
	}
	if config.CgrConfig().SessionSCfg().Enabled {
		subsystems = append(subsystems, utils.SessionSv1)
	}
	if config.CgrConfig().ChargerSCfg().Enabled {
		subsystems = append(subsystems, utils.ChargerSv1)
	}
	if config.CgrConfig().ResourceSCfg().Enabled {
		subsystems = append(subsystems, utils.ResourceSv1)
	}
	if config.CgrConfig().StatSCfg().Enabled {
		subsystems = append(subsystems, utils.StatSv1)
	}
	if config.CgrConfig().ThresholdSCfg().Enabled {
		subsystems = append(subsystems, utils.ThresholdSv1)
	}
	if config.CgrConfig().SupplierSCfg().Enabled {
		subsystems = append(subsystems, utils.SupplierSv1)
	}
	IntRPC = rpcclient.NewRPCClientSet(subsystems, config.CgrConfig().GeneralCfg().InternalTtl)
}

func AddInternalRPCClient(name string, rpc rpcclient.RpcClientConnection) {
	connChan := make(chan rpcclient.RpcClientConnection, 1)
	connChan <- rpc
	err := IntRPC.AddRPCConnection(name, "", "", false, "", "", "",
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.INTERNAL_RPC, connChan, true)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<InternalRCP> Error adding %s to the set: %v", name, err.Error()))
	}
}

func GetInternalRPCClientChanel() chan rpcclient.RpcClientConnection {
	connChan := make(chan rpcclient.RpcClientConnection, 1)
	connChan <- IntRPC
	return connChan
}
