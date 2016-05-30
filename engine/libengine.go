/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"errors"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewRPCPool(dispatchStrategy string, connAttempts, reconnects int, connectTimeout, replyTimeout time.Duration, codec string,
	rpcConnCfgs []*config.HaPoolConfig, internalConnChan chan rpcclient.RpcClientConnection, ttl time.Duration) (*rpcclient.RpcClientPool, error) {
	var rpcClient *rpcclient.RpcClient
	var err error
	rpcPool := rpcclient.NewRpcClientPool(dispatchStrategy)
	atLestOneConnected := false // If one connected we don't longer return errors
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.MetaInternal {
			var internalConn rpcclient.RpcClientConnection
			select {
			case internalConn = <-internalConnChan:
				internalConnChan <- internalConn
			case <-time.After(ttl):
				return nil, errors.New("TTL triggered")
			}
			rpcClient, err = rpcclient.NewRpcClient("", "", connAttempts, reconnects, connectTimeout, replyTimeout, rpcclient.INTERNAL_RPC, internalConn)
		} else {
			rpcClient, err = rpcclient.NewRpcClient("tcp", rpcConnCfg.Address, connAttempts, reconnects, connectTimeout, replyTimeout, codec, nil)
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
