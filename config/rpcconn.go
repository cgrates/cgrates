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
	return dfltRemoteHost.Clone()
}

// NewDfltRPCConn returns the default value for a RPCConn
func NewDfltRPCConn() *RPCConn {
	return &RPCConn{Strategy: rpcclient.PoolFirst}
}

// RPCConns the config for all rpc pools
type RPCConns map[string]*RPCConn

func (rC RPCConns) loadFromJSONCfg(jsn RPCConnsJson) {
	// hardoded the *internal connection
	rC[utils.MetaInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address: utils.MetaInternal,
		}},
	}
	rC[rpcclient.BiRPCInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address: rpcclient.BiRPCInternal,
		}},
	}
	rC[utils.MetaLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
		}},
	}
	rC[utils.MetaBiJSONLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2014",
			Transport: rpcclient.BiRPCJSON,
		}},
	}
	for key, val := range jsn {
		rC[key] = NewDfltRPCConn()
		rC[key].loadFromJSONCfg(val)
	}
}

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

func (rC *RPCConn) loadFromJSONCfg(jsnCfg *RPCConnJson) {
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
	if jsnCfg.Client_key != nil {
		rh.ClientKey = *jsnCfg.Client_key
	}
	if jsnCfg.Client_certificate != nil {
		rh.ClientCertificate = *jsnCfg.Client_certificate
	}
	if jsnCfg.Ca_certificate != nil {
		rh.CaCertificate = *jsnCfg.Ca_certificate
	}
	if jsnCfg.Connect_attempts != nil {
		rh.ConnectAttempts = *jsnCfg.Connect_attempts
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
		TLS:               rh.TLS,
		ClientKey:         rh.ClientKey,
		ClientCertificate: rh.ClientCertificate,
		CaCertificate:     rh.CaCertificate,
		ConnectAttempts:   rh.ConnectAttempts,
		Reconnects:        rh.Reconnects,
		ConnectTimeout:    rh.ConnectTimeout,
		ReplyTimeout:      rh.ReplyTimeout,
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
			rh.TLS = newHost.TLS
			rh.ClientKey = newHost.ClientKey
			rh.ClientCertificate = newHost.ClientCertificate
			rh.CaCertificate = newHost.CaCertificate
			rh.ConnectAttempts = newHost.ConnectAttempts
			rh.Reconnects = newHost.Reconnects
			rh.ConnectTimeout = newHost.ConnectTimeout
			rh.ReplyTimeout = newHost.ReplyTimeout
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
			rh.TLS = false
			rh.ClientKey = ""
			rh.ClientCertificate = ""
			rh.CaCertificate = ""
			rh.ConnectAttempts = 0
			rh.Reconnects = 0
			rh.ConnectTimeout = 0
			rh.ReplyTimeout = 0
		}
	}
	return
}

// Represents one connection instance towards a rater/cdrs server
type RemoteHostJson struct {
	Id                 *string
	Address            *string
	Transport          *string
	Synchronous        *bool
	Connect_attempts   *int
	Reconnects         *int
	Connect_timeout    *string
	Reply_timeout      *string
	Tls                *bool
	Client_certificate *string
	Client_key         *string
	Ca_certificate     *string
}

func diffRemoteHostJson(v1, v2 *RemoteHost) (d *RemoteHostJson) {
	d = new(RemoteHostJson)
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.Transport != v2.Transport {
		d.Transport = utils.StringPointer(v2.Transport)
	}
	if v1.Synchronous != v2.Synchronous {
		d.Synchronous = utils.BoolPointer(v2.Synchronous)
	}
	if v1.TLS != v2.TLS {
		d.Tls = utils.BoolPointer(v2.TLS)
	}
	if v1.ClientKey != v2.ClientKey {
		d.Client_key = utils.StringPointer(v2.ClientKey)
	}
	if v1.ClientCertificate != v2.ClientCertificate {
		d.Client_certificate = utils.StringPointer(v2.ClientCertificate)
	}
	if v1.ClientCertificate != v2.ClientCertificate {
		d.Ca_certificate = utils.StringPointer(v2.CaCertificate)
	}
	if v1.ConnectAttempts != v2.ConnectAttempts {
		d.Connect_attempts = utils.IntPointer(v2.ConnectAttempts)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	if v1.ConnectTimeout != v2.ConnectTimeout {
		d.Connect_timeout = utils.StringPointer(v2.ConnectTimeout.String())
	}
	if v1.ReplyTimeout != v2.ReplyTimeout {
		d.Reply_timeout = utils.StringPointer(v2.ReplyTimeout.String())
	}
	return
}

type RPCConnJson struct {
	Strategy *string
	PoolSize *int
	Conns    *[]*RemoteHostJson
}

func diffRPCConnJson(d *RPCConnJson, v1, v2 *RPCConn) *RPCConnJson {
	if d == nil {
		d = new(RPCConnJson)
	}
	if v1.Strategy != v2.Strategy {
		d.Strategy = utils.StringPointer(v2.Strategy)
	}
	if v1.PoolSize != v2.PoolSize {
		d.PoolSize = utils.IntPointer(v2.PoolSize)
	}
	if v2.Conns != nil {
		conns := make([]*RemoteHostJson, len(v2.Conns))
		dft := NewDfltRemoteHost()
		for i, conn := range v2.Conns {
			conns[i] = diffRemoteHostJson(dft, conn)
		}
		d.Conns = &conns
	}
	return d
}

func equalsRemoteHosts(v1, v2 []*RemoteHost) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].Address != v2[i].Address ||
			v1[i].Transport != v2[i].Transport ||
			v1[i].Synchronous != v2[i].Synchronous ||
			v1[i].TLS != v2[i].TLS ||
			v1[i].ClientKey != v2[i].ClientKey ||
			v1[i].ClientCertificate != v2[i].ClientCertificate ||
			v1[i].CaCertificate != v2[i].CaCertificate ||
			v1[i].ConnectAttempts != v2[i].ConnectAttempts ||
			v1[i].Reconnects != v2[i].Reconnects ||
			v1[i].ConnectTimeout != v2[i].ConnectTimeout ||
			v1[i].ReplyTimeout != v2[i].ReplyTimeout {
			return false
		}
	}
	return true
}

func equalsRPCConn(v1, v2 *RPCConn) bool {
	return (v1 == nil && v2 == nil) ||
		(v1 != nil && v2 != nil &&
			v1.Strategy == v2.Strategy &&
			v1.PoolSize == v2.PoolSize &&
			equalsRemoteHosts(v1.Conns, v2.Conns))
}

type RPCConnsJson map[string]*RPCConnJson

func diffRPCConnsJson(d RPCConnsJson, v1, v2 RPCConns) RPCConnsJson {
	if d == nil {
		d = make(RPCConnsJson)
	}
	dft := NewDfltRPCConn()
	for k, val := range v2 {
		if !equalsRPCConn(v1[k], val) {
			d[k] = diffRPCConnJson(d[k], dft, val)
		}
	}
	return d
}
