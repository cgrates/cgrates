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
	"time"

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
	ID                string
	Address           string
	Transport         string
	Synchronous       bool
	ConnectAttempts   int
	Reconnects        int
	ConnectTimeout    time.Duration
	ReplyTimeout      time.Duration
	TLS               bool
	ClientKey         string
	ClientCertificate string
	CaCertificate     string
}

func (rh *RemoteHost) loadFromJSONCfg(jsnCfg *RemoteHostJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Id != nil {
		rh.ID = *jsnCfg.Id
		// ignore defaults if we have ID
		rh.Address = utils.EmptyString
		rh.Transport = utils.EmptyString
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
	if jsnCfg.Key_path != nil {
		rh.ClientKey = *jsnCfg.Key_path
	}
	if jsnCfg.Cert_path != nil {
		rh.ClientCertificate = *jsnCfg.Cert_path
	}
	if jsnCfg.Ca_path != nil {
		rh.CaCertificate = *jsnCfg.Ca_path
	}
	if jsnCfg.Conn_attempts != nil {
		rh.ConnectAttempts = *jsnCfg.Conn_attempts
	}
	if jsnCfg.Reconnects != nil {
		rh.Reconnects = *jsnCfg.Reconnects
	}
	if jsnCfg.Connect_timeout != nil {
		if rh.ConnectTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.Connect_timeout); err != nil {
			return err
		}
	}
	if jsnCfg.Reply_timeout != nil {
		if rh.ReplyTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.Reply_timeout); err != nil {
			return err
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rh *RemoteHost) AsMapInterface() (mp map[string]interface{}) {
	mp = map[string]interface{}{
		utils.AddressCfg:   rh.Address,
		utils.TransportCfg: rh.Transport,
	}
	if rh.ID != utils.EmptyString {
		mp[utils.IDCfg] = rh.ID
	}
	if rh.Synchronous {
		mp[utils.SynchronousCfg] = rh.Synchronous
	}
	if rh.TLS {
		mp[utils.TLSNoCaps] = rh.TLS
	}
	if rh.ClientKey != utils.EmptyString {
		mp[utils.KeyPathCgr] = rh.ClientKey
	}
	if rh.ClientCertificate != utils.EmptyString {
		mp[utils.CertPathCgr] = rh.ClientCertificate
	}
	if rh.CaCertificate != utils.EmptyString {
		mp[utils.CAPathCgr] = rh.CaCertificate
	}
	if rh.ConnectAttempts != 0 {
		mp[utils.ConnectAttemptsCfg] = rh.ConnectAttempts
	}
	if rh.Reconnects != 0 {
		mp[utils.ReconnectsCfg] = rh.Reconnects
	}
	if rh.ConnectTimeout != 0 {
		mp[utils.ConnectTimeoutCfg] = rh.ConnectTimeout
	}
	if rh.ReplyTimeout != 0 {
		mp[utils.ReplyTimeoutCfg] = rh.ReplyTimeout
	}
	return
}

// Clone returns a deep copy of RemoteHost
func (rh RemoteHost) Clone() (cln *RemoteHost) {
	return &RemoteHost{
		ID:                rh.ID,
		Address:           rh.Address,
		Transport:         rh.Transport,
		Synchronous:       rh.Synchronous,
		ConnectAttempts:   rh.ConnectAttempts,
		Reconnects:        rh.Reconnects,
		ConnectTimeout:    rh.ConnectTimeout,
		ReplyTimeout:      rh.ReplyTimeout,
		TLS:               rh.TLS,
		ClientKey:         rh.ClientKey,
		ClientCertificate: rh.ClientCertificate,
		CaCertificate:     rh.CaCertificate,
	}
}

// UpdateRPCCons will parse each conn and update only
// the conns that have the same ID
func UpdateRPCCons(rpcConns RPCConns, newHosts map[string]*RemoteHost) (connIDs utils.StringSet) {
	connIDs = make(utils.StringSet)
	for rpcKey, rpcPool := range rpcConns {
		for _, rh := range rpcPool.Conns {
			newHost, has := newHosts[rh.ID]
			if !has {
				continue
			}
			connIDs.Add(rpcKey)
			rh.Address = newHost.Address
			rh.Transport = newHost.Transport
			rh.Synchronous = newHost.Synchronous
			rh.ConnectAttempts = newHost.ConnectAttempts
			rh.Reconnects = newHost.Reconnects
			rh.ConnectTimeout = newHost.ConnectTimeout
			rh.ReplyTimeout = newHost.ReplyTimeout
			rh.TLS = newHost.TLS
			rh.ClientKey = newHost.ClientKey
			rh.ClientCertificate = newHost.ClientCertificate
			rh.CaCertificate = newHost.CaCertificate
		}
	}
	return
}

// RemoveRPCCons will parse each conn and reset only
// the conns that have the same ID
func RemoveRPCCons(rpcConns RPCConns, hosts utils.StringSet) (connIDs utils.StringSet) {
	connIDs = make(utils.StringSet)
	for rpcKey, rpcPool := range rpcConns {
		for _, rh := range rpcPool.Conns {
			if !hosts.Has(rh.ID) {
				continue
			}
			connIDs.Add(rpcKey)
			rh.Address = ""
			rh.Transport = ""
			rh.Synchronous = false
			rh.ConnectAttempts = 0
			rh.Reconnects = 0
			rh.ConnectTimeout = 0
			rh.ReplyTimeout = 0
			rh.TLS = false
			rh.ClientKey = ""
			rh.ClientCertificate = ""
			rh.CaCertificate = ""
		}
	}
	return
}
