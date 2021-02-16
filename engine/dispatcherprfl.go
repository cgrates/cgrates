/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"math/rand"
	"sort"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type DispatcherHostProfile struct {
	ID        string
	FilterIDs []string
	Weight    float64                // applied in case of multiple connections need to be ordered
	Params    map[string]interface{} // additional parameters stored for a session
	Blocker   bool                   // no connection after this one
}

func (dC *DispatcherHostProfile) Clone() (cln *DispatcherHostProfile) {
	cln = &DispatcherHostProfile{
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

type DispatcherHostProfiles []*DispatcherHostProfile

// Sort is part of sort interface, sort based on Weight
func (dHPrfls DispatcherHostProfiles) Sort() {
	sort.Slice(dHPrfls, func(i, j int) bool { return dHPrfls[i].Weight > dHPrfls[j].Weight })
}

// ReorderFromIndex will consider idx as starting point for the reordered slice
func (dHPrfls DispatcherHostProfiles) ReorderFromIndex(idx int) {
	initConns := dHPrfls.Clone()
	for i := 0; i < len(dHPrfls); i++ {
		if idx > len(dHPrfls)-1 {
			idx = 0
		}
		dHPrfls[i] = initConns[idx]
		idx++
	}
	return
}

// Shuffle will mix the connections in place
func (dHPrfls DispatcherHostProfiles) Shuffle() {
	rand.Shuffle(len(dHPrfls), func(i, j int) {
		dHPrfls[i], dHPrfls[j] = dHPrfls[j], dHPrfls[i]
	})
	return
}

func (dHPrfls DispatcherHostProfiles) Clone() (cln DispatcherHostProfiles) {
	cln = make(DispatcherHostProfiles, len(dHPrfls))
	for i, dHPrfl := range dHPrfls {
		cln[i] = dHPrfl.Clone()
	}
	return
}

func (dHPrfls DispatcherHostProfiles) HostIDs() (hostIDs []string) {
	hostIDs = make([]string, len(dHPrfls))
	for i, hostPrfl := range dHPrfls {
		hostIDs[i] = hostPrfl.ID
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
	Hosts              DispatcherHostProfiles // dispatch to these connections
}

// DispatcherProfileWithOpts is used in replicatorV1 for dispatcher
type DispatcherProfileWithOpts struct {
	*DispatcherProfile
	Opts map[string]interface{}
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

// DispatcherHost represents one virtual host used by dispatcher
type DispatcherHost struct {
	Tenant  string
	ID      string
	Conn    *config.RemoteHost
	rpcConn rpcclient.ClientConnector
}

// DispatcherHostWithOpts is used in replicatorV1 for dispatcher
type DispatcherHostWithOpts struct {
	*DispatcherHost
	Opts map[string]interface{}
}

// TenantID returns the tenant concatenated with the ID
func (dH *DispatcherHost) TenantID() string {
	return utils.ConcatenatedKey(dH.Tenant, dH.ID)
}

// Call will build and cache the connection if it is not defined yet then will execute the method on conn
func (dH *DispatcherHost) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if dH.rpcConn == nil {
		// connect the rpcConn
		cfg := config.CgrConfig()
		if dH.rpcConn, err = NewRPCConnection(dH.Conn,
			cfg.TLSCfg().ClientKey,
			cfg.TLSCfg().ClientCerificate, cfg.TLSCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			IntRPC.GetInternalChanel(), false, nil); err != nil {
			return
		}

	}
	return dH.rpcConn.Call(serviceMethod, args, reply)
}

type DispatcherHostIDs []string

// ReorderFromIndex will consider idx as starting point for the reordered slice
func (dHPrflIDs DispatcherHostIDs) ReorderFromIndex(idx int) {
	initConns := dHPrflIDs.Clone()
	for i := 0; i < len(dHPrflIDs); i++ {
		if idx > len(dHPrflIDs)-1 {
			idx = 0
		}
		dHPrflIDs[i] = initConns[idx]
		idx++
	}
	return
}

// Shuffle will mix the connections in place
func (dHPrflIDs DispatcherHostIDs) Shuffle() {
	rand.Shuffle(len(dHPrflIDs), func(i, j int) {
		dHPrflIDs[i], dHPrflIDs[j] = dHPrflIDs[j], dHPrflIDs[i]
	})
	return
}

func (dHPrflIDs DispatcherHostIDs) Clone() (cln DispatcherHostIDs) {
	cln = make(DispatcherHostIDs, len(dHPrflIDs))
	for i, dhID := range dHPrflIDs {
		cln[i] = dhID
	}
	return
}
