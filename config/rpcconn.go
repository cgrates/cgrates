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
	return dfltRemoteHost.Clone()
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
	ID          string
	Address     string
	Transport   string
	Synchronous bool
	TLS         bool
}

func (rh *RemoteHost) loadFromJSONCfg(jsnCfg *RemoteHostJson) {
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
		mp[utils.TLS] = rh.TLS
	}
	return
}

// Clone returns a deep copy of RemoteHost
func (rh RemoteHost) Clone() (cln *RemoteHost) {
	return &RemoteHost{
		ID:          rh.ID,
		Address:     rh.Address,
		Transport:   rh.Transport,
		Synchronous: rh.Synchronous,
		TLS:         rh.TLS,
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
		}
	}
	return
}

// Represents one connection instance towards a rater/cdrs server
type RemoteHostJson struct {
	Id          *string
	Address     *string
	Transport   *string
	Synchronous *bool
	Tls         *bool
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
			v1[i].TLS != v2[i].TLS {
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
