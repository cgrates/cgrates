// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func (rC *RPCConn) AsMapInterface() map[string]any {
	conns := make([]map[string]any, len(rC.Conns))
	for i, item := range rC.Conns {
		conns[i] = item.AsMapInterface()
	}

	return map[string]any{
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

func (rh *RemoteHost) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AddressCfg:     rh.Address,
		utils.TransportCfg:   rh.Transport,
		utils.SynchronousCfg: rh.Synchronous,
		utils.TlsCfg:         rh.TLS,
	}
}
