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

// Returns the first cached default value for a RemoteHost connection
func NewDfltRemoteHost() *RemoteHost {
	if dfltRemoteHost == nil {
		return new(RemoteHost) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltRemoteHost // Copy the value instead of it's pointer
	return &dfltVal
}

func NewDfltRPCConn() *RPCConn {
	return &RPCConn{Strategy: rpcclient.PoolFirst}
}

type RPCConn struct {
	Strategy string
	PoolSize int
	Conns    []*RemoteHost
}

func (rC *RPCConn) loadFromJsonCfg(jsnCfg *RPCConnsJson) (err error) {
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
			rC.Conns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	return
}

func (rC *RPCConn) AsMapInterface() map[string]interface{} {
	conns := make([]map[string]interface{}, len(rC.Conns))
	for i, item := range rC.Conns {
		conns[i] = item.AsMapInterface()
	}

	return map[string]interface{}{
		utils.StrategyCfg: rC.Strategy,
		utils.PoolSize:    rC.PoolSize,
		utils.Conns:       conns,
	}
}

// One connection to Rater
type RemoteHost struct {
	ID          string
	Address     string
	Transport   string
	Synchronous bool
	TLS         bool
}

func (self *RemoteHost) loadFromJsonCfg(jsnCfg *RemoteHostJson) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		self.ID = *jsnCfg.Id
		// ignore defaults if we have ID
		self.Address = utils.EmptyString
		self.Transport = utils.EmptyString
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Transport != nil {
		self.Transport = *jsnCfg.Transport
	}
	if jsnCfg.Synchronous != nil {
		self.Synchronous = *jsnCfg.Synchronous
	}
	if jsnCfg.Tls != nil {
		self.TLS = *jsnCfg.Tls
	}
	return nil
}

func (rh *RemoteHost) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AddressCfg:     rh.Address,
		utils.TransportCfg:   rh.Transport,
		utils.SynchronousCfg: rh.Synchronous,
		utils.TlsCfg:         rh.TLS,
	}
}
