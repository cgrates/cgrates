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
	"strings"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

// NewRPCPool returns a new pool of connection with the given configuration
func NewRPCPool(ctx *context.Context, dispatchStrategy string, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, rpcConnCfgs []*config.RemoteHost,
	internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient interface{}, poolID string, connCache *ltcache.Cache) (rpcPool *rpcclient.RPCPool, err error) {
	var rpcClient birpc.ClientConnector
	var atLestOneConnected bool // If one connected we don't longer return errors
	rpcPool = rpcclient.NewRPCPool(dispatchStrategy, replyTimeout)
	for _, rpcConnCfg := range rpcConnCfgs {
		if rpcConnCfg.Address == utils.EmptyString {
			// in case we have only conns with empty addresse
			// mimic an error to signal that the init was not done
			err = rpcclient.ErrDisconnected
			continue
		}
		if rpcClient, err = NewRPCConnection(ctx, rpcConnCfg, keyPath, certPath, caPath, connAttempts, reconnects,
			connectTimeout, replyTimeout, internalConnChan, lazyConnect, biRPCClient,
			poolID, rpcConnCfg.ID, connCache); err == rpcclient.ErrUnsupportedCodec {
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
	return
}

// NewRPCConnection creates a new connection based on the RemoteHost structure
// connCache is used to cache the connection with ID
func NewRPCConnection(ctx *context.Context, cfg *config.RemoteHost, keyPath, certPath, caPath string, connAttempts, reconnects int,
	connectTimeout, replyTimeout time.Duration, internalConnChan chan birpc.ClientConnector, lazyConnect bool,
	biRPCClient interface{}, poolID, connID string, connCache *ltcache.Cache) (client birpc.ClientConnector, err error) {
	var id string
	if connID != utils.EmptyString {
		id = poolID + utils.ConcatenatedKeySep + connID
		if x, ok := connCache.Get(id); ok && x != nil {
			return x.(birpc.ClientConnector), nil
		}
	}
	if cfg.Address == rpcclient.InternalRPC ||
		cfg.Address == rpcclient.BiRPCInternal {
		client, err = rpcclient.NewRPCClient(ctx, "", "", cfg.TLS, utils.FirstNonEmpty(cfg.ClientKey, keyPath), utils.FirstNonEmpty(cfg.ClientCertificate, certPath), utils.FirstNonEmpty(cfg.CaCertificate, caPath), utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects), utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout), utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout), cfg.Address, internalConnChan, lazyConnect, biRPCClient)
	} else {
		client, err = rpcclient.NewRPCClient(ctx, utils.TCP, cfg.Address, cfg.TLS, utils.FirstNonEmpty(cfg.ClientKey, keyPath), utils.FirstNonEmpty(cfg.ClientCertificate, certPath), utils.FirstNonEmpty(cfg.CaCertificate, caPath),
			utils.FirstIntNonEmpty(cfg.ConnectAttempts, connAttempts),
			utils.FirstIntNonEmpty(cfg.Reconnects, reconnects), utils.FirstDurationNonEmpty(cfg.ConnectTimeout, connectTimeout), utils.FirstDurationNonEmpty(cfg.ReplyTimeout, replyTimeout),
			utils.FirstNonEmpty(cfg.Transport, rpcclient.GOBrpc), nil, lazyConnect, biRPCClient)
	}
	if connID != utils.EmptyString &&
		err == nil {
		connCache.Set(id, client, nil)
	}
	return
}

func NewRPCClientSet(m map[string]chan birpc.ClientConnector) (s RPCClientSet) {
	s = make(RPCClientSet)
	for k, v := range map[string]string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzer):       utils.AnalyzerSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS):         utils.AdminSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes):     utils.AttributeSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):         utils.CacheSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs):           utils.CDRsV1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):       utils.ChargerSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian):       utils.GuardianSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders):        utils.LoaderSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):      utils.ResourceSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS):       utils.SessionSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):          utils.StatSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):         utils.RouteSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     utils.ThresholdSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): utils.ServiceManagerV1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         utils.ConfigSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           utils.CoreSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs):            utils.EeSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS):          utils.RateSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers):    utils.DispatcherSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts):       utils.AccountSv1,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions):        utils.ActionSv1,
	} {
		s[v] = m[k]
	}
	return
}

// RPCClientSet is a RPC ClientConnector for the internal subsystems
type RPCClientSet map[string]chan birpc.ClientConnector

// GetInternalChanel is used when RPCClientSet is passed as internal connection for RPCPool
func (s RPCClientSet) GetInternalChanel() chan birpc.ClientConnector {
	connChan := make(chan birpc.ClientConnector, 1)
	connChan <- s
	return connChan
}

// Call the implementation of the birpc.ClientConnector interface
func (s RPCClientSet) Call(ctx *context.Context, method string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(method, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	connCh, has := s[methodSplit[0]]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	var conn birpc.ClientConnector
	ctx2, cancel := context.WithTimeout(ctx, config.CgrConfig().GeneralCfg().ConnectTimeout)
	select {
	case conn = <-connCh:
		connCh <- conn
		cancel()
		if conn == nil {
			return rpcclient.ErrDisconnected
		}
	case <-ctx2.Done():
		return ctx2.Err()
	}
	return conn.Call(ctx, method, args, reply)
}
