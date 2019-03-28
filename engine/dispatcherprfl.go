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
	"math/rand"
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type DispatcherConn struct {
	ID        string
	FilterIDs []string
	Weight    float64                // applied in case of multiple connections need to be ordered
	Params    map[string]interface{} // additional parameters stored for a session
	Blocker   bool                   // no connection after this one
}

func (dC *DispatcherConn) Clone() (cln *DispatcherConn) {
	cln = &DispatcherConn{
		ID:      dC.ID,
		Weight:  dC.Weight,
		Blocker: dC.Blocker,
	}
	if dC.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(dC.FilterIDs))
		for i, fltr := range dC.FilterIDs {
			cln.FilterIDs[i] = fltr
		}
	}
	if dC.Params != nil {
		cln.Params = make(map[string]interface{})
		for k, v := range dC.Params {
			cln.Params[k] = v
		}
	}
	return
}

type DispatcherConns []*DispatcherConn

// Sort is part of sort interface, sort based on Weight
func (dConns DispatcherConns) Sort() {
	sort.Slice(dConns, func(i, j int) bool { return dConns[i].Weight > dConns[j].Weight })
}

// ReorderFromIndex will consider idx as starting point for the reordered slice
func (dConns DispatcherConns) ReorderFromIndex(idx int) {
	initConns := dConns.Clone()
	for i := 0; i < len(dConns); i++ {
		if idx > len(dConns)-1 {
			idx = 0
		}
		dConns[i] = initConns[idx]
		idx++
	}
	return
}

// Shuffle will mix the connections in place
func (dConns DispatcherConns) Shuffle() {
	rand.Shuffle(len(dConns), func(i, j int) {
		dConns[i], dConns[j] = dConns[j], dConns[i]
	})
	return
}

func (dConns DispatcherConns) Clone() (cln DispatcherConns) {
	cln = make(DispatcherConns, len(dConns))
	for i, dConn := range dConns {
		cln[i] = dConn.Clone()
	}
	return
}

func (dConns DispatcherConns) ConnIDs() (connIDs []string) {
	connIDs = make([]string, len(dConns))
	for i, conn := range dConns {
		connIDs[i] = conn.ID
	}
	return
}

// DispatcherProfile is the config for one Dispatcher
type DispatcherProfile struct {
	Tenant             string
	ID                 string
	Subsystems         []string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // activation interval
	Strategy           string
	StrategyParams     map[string]interface{} // ie for distribution, set here the pool weights
	Weight             float64                // used for profile sorting on match
	Conns              DispatcherConns        // dispatch to these connections
}

func (dP *DispatcherProfile) TenantID() string {
	return utils.ConcatenatedKey(dP.Tenant, dP.ID)
}

// DispatcherProfiles is a sortable list of Dispatcher profiles
type DispatcherProfiles []*DispatcherProfile

// Sort is part of sort interface, sort based on Weight
func (dps DispatcherProfiles) Sort() {
	sort.Slice(dps, func(i, j int) bool { return dps[i].Weight > dps[j].Weight })
}

<<<<<<< HEAD
type DispatcherHostConn struct {
	Address   string
	Transport string
	TLS       bool
}
=======
// DispatcherHost represents one virtual host with po
>>>>>>> DispatcherHost.GetRPCConnection
type DispatcherHost struct {
	Tenant  string
	ID      string
	Conns   []*config.HaPoolConfig
	rpcConn rpcclient.RpcClientConnection
}

func (dH *DispatcherHost) TenantID() string {
	return utils.ConcatenatedKey(dH.Tenant, dH.ID)
}

// GetRPCConnection builds or returns the cached connection
func (dH *DispatcherHost) GetRPCConnection() (rpcConn rpcclient.RpcClientConnection, err error) {
	if dH.rpcConn == nil {
		cfg := config.CgrConfig()
		if dH.rpcConn, err = NewRPCPool(
			rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			dH.Conns, nil, time.Duration(0), false); err != nil {
			return
		}
	}
	return dH.rpcConn, nil
}
