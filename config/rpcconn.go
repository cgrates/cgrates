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

package config

import (
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDfltRemoteHost returns the first cached default value for a RemoteHost connection
func NewDfltRemoteHost() *RemoteHost {
	if dfltRemoteHost == nil {
		return new(RemoteHost) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltRemoteHost // Copy the value instead of it's pointer
	return &dfltVal
}

// NewDfltRPCConn returns the default value for a RPCConn
func NewDfltRPCConn() *RPCConn {
	return &RPCConn{Strategy: rpcclient.PoolFirst}
}

// RPCConns the config for all rpc pools
type RPCConns map[string]*RPCConn

// AsMapInterface returns the config as a map[string]interface{}
func (rC RPCConns) AsMapInterface() (rpcConns map[string]interface{}) {
	rpcConns = make(map[string]interface{})
	for key, value := range rC {
		rpcConns[key] = value.AsMapInterface()
	}
	return
}

// Clone returns a deep copy of RPCConns
func (rC RPCConns) Clone() (cln RPCConns) {
	cln = make(RPCConns)
	for id, conn := range rC {
		cln[id] = conn.Clone()
	}
	return
}

// RPCConn the connection pool config
type RPCConn struct {
	Strategy string
	PoolSize int
	Conns    []*RemoteHost
}

func (rC *RPCConn) loadFromJSONCfg(jsnCfg *RPCConnsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Strategy != nil {
		rC.Strategy = *jsnCfg.Strategy
	}
	if jsnCfg.PoolSize != nil {
		rC.PoolSize = *jsnCfg.PoolSize
	}
	if jsnCfg.Conns != nil {
		rC.Conns = make([]*RemoteHost, len(*jsnCfg.Conns))
		for idx, jsnHaCfg := range *jsnCfg.Conns {
			rC.Conns[idx] = NewDfltRemoteHost()
			rC.Conns[idx].loadFromJSONCfg(jsnHaCfg) //To review if the function signature changes
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rC *RPCConn) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.StrategyCfg: rC.Strategy,
		utils.PoolSize:    rC.PoolSize,
	}
	if rC.Conns != nil {
		conns := make([]map[string]interface{}, len(rC.Conns))
		for i, item := range rC.Conns {
			conns[i] = item.AsMapInterface()
		}
		initialMP[utils.Conns] = conns
	}
	return
}

// Clone returns a deep copy of RPCConn
func (rC RPCConn) Clone() (cln *RPCConn) {
	cln = &RPCConn{
		Strategy: rC.Strategy,
		PoolSize: rC.PoolSize,
	}
	if rC.Conns != nil {
		cln.Conns = make([]*RemoteHost, len(rC.Conns))
		for i, req := range rC.Conns {
			cln.Conns[i] = req.Clone()
		}
	}
	return
}

// RemoteHost connection config
type RemoteHost struct {
	Address     string
	Transport   string
	Synchronous bool
	TLS         bool
}

func (rh *RemoteHost) loadFromJSONCfg(jsnCfg *RemoteHostJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Address != nil {
		rh.Address = *jsnCfg.Address
	}
	if jsnCfg.Transport != nil {
		rh.Transport = *jsnCfg.Transport
	}
	if jsnCfg.Synchronous != nil {
		rh.Synchronous = *jsnCfg.Synchronous
	}
	if jsnCfg.Tls != nil {
		rh.TLS = *jsnCfg.Tls
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rh *RemoteHost) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AddressCfg:     rh.Address,
		utils.TransportCfg:   rh.Transport,
		utils.SynchronousCfg: rh.Synchronous,
		utils.TLS:            rh.TLS,
	}
}

// Clone returns a deep copy of RemoteHost
func (rh RemoteHost) Clone() (cln *RemoteHost) {
	return &RemoteHost{
		Address:     rh.Address,
		Transport:   rh.Transport,
		Synchronous: rh.Synchronous,
		TLS:         rh.TLS,
	}
}
