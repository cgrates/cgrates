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
	"strings"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
		cln.FilterIDs = utils.CloneStringSlice(dC.FilterIDs)
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
}

// Shuffle will mix the connections in place
func (dHPrfls DispatcherHostProfiles) Shuffle() {
	rand.Shuffle(len(dHPrfls), func(i, j int) {
		dHPrfls[i], dHPrfls[j] = dHPrfls[j], dHPrfls[i]
	})
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
	Tenant         string
	ID             string
	FilterIDs      []string
	Strategy       string
	StrategyParams map[string]interface{} // ie for distribution, set here the pool weights
	Weight         float64                // used for profile sorting on match
	Hosts          DispatcherHostProfiles // dispatch to these connections
}

// DispatcherProfileWithAPIOpts is used in replicatorV1 for dispatcher
type DispatcherProfileWithAPIOpts struct {
	*DispatcherProfile
	APIOpts map[string]interface{}
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
	Tenant string
	*config.RemoteHost
	rpcConn birpc.ClientConnector
}

// DispatcherHostWithOpts is used in replicatorV1 for dispatcher
type DispatcherHostWithAPIOpts struct {
	*DispatcherHost
	APIOpts map[string]interface{}
}

// TenantID returns the tenant concatenated with the ID
func (dH *DispatcherHost) TenantID() string {
	return utils.ConcatenatedKey(dH.Tenant, dH.ID)
}

// GetConn will build and cache the connection if it is not defined yet
func (dH *DispatcherHost) GetConn(ctx *context.Context, cfg *config.CGRConfig, iPRCCh chan birpc.ClientConnector) (_ birpc.ClientConnector, err error) {
	if dH.rpcConn == nil {
		// connect the rpcConn
		if dH.rpcConn, err = NewRPCConnection(ctx, dH.RemoteHost,
			cfg.TLSCfg().ClientKey,
			cfg.TLSCfg().ClientCerificate, cfg.TLSCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			iPRCCh, false, nil,
			utils.EmptyString, utils.EmptyString, nil); err != nil {
			return
		}

	}
	return dH.rpcConn, nil
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
}

// Shuffle will mix the connections in place
func (dHPrflIDs DispatcherHostIDs) Shuffle() {
	rand.Shuffle(len(dHPrflIDs), func(i, j int) {
		dHPrflIDs[i], dHPrflIDs[j] = dHPrflIDs[j], dHPrflIDs[i]
	})
}

func (dHPrflIDs DispatcherHostIDs) Clone() (cln DispatcherHostIDs) {
	cln = make(DispatcherHostIDs, len(dHPrflIDs))
	copy(cln, dHPrflIDs)
	return
}

func (dP *DispatcherProfile) Set(path []string, val interface{}, newBranch bool, _ string) (err error) {
	switch len(path) {
	default:
		return utils.ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], utils.StrategyParams) &&
				path[0][14] == '[' && path[0][len(path[0])-1] == ']' {
				dP.StrategyParams[path[0][15:len(path[0])-1]] = val
				return
			}
			return utils.ErrWrongPath
		case utils.Tenant:
			dP.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			dP.ID = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			dP.FilterIDs = append(dP.FilterIDs, valA...)
		case utils.Strategy:
			dP.Strategy = utils.IfaceAsString(val)
		case utils.Weight:
			dP.Weight, err = utils.IfaceAsFloat64(val)
		case utils.StrategyParams:
			dP.StrategyParams, err = utils.NewMapFromCSV(utils.IfaceAsString(val))
		}
	case 2:
		switch path[0] {
		default:
			return utils.ErrWrongPath
		case utils.StrategyParams:
			dP.StrategyParams[path[1]] = val
		case utils.Hosts:
			if len(dP.Hosts) == 0 || newBranch {
				dP.Hosts = append(dP.Hosts, &DispatcherHostProfile{Params: make(map[string]interface{})})
			}
			switch path[1] {
			case utils.ID:
				dP.Hosts[len(dP.Hosts)-1].ID = utils.IfaceAsString(val)
			case utils.FilterIDs:
				var valA []string
				valA, err = utils.IfaceAsStringSlice(val)
				dP.Hosts[len(dP.Hosts)-1].FilterIDs = append(dP.Hosts[len(dP.Hosts)-1].FilterIDs, valA...)
			case utils.Weight:
				dP.Hosts[len(dP.Hosts)-1].Weight, err = utils.IfaceAsFloat64(val)
			case utils.Blocker:
				dP.Hosts[len(dP.Hosts)-1].Blocker, err = utils.IfaceAsBool(val)
			case utils.Params:
				dP.Hosts[len(dP.Hosts)-1].Params, err = utils.NewMapFromCSV(utils.IfaceAsString(val))
			default:
				if strings.HasPrefix(path[1], utils.Params) &&
					path[1][6] == '[' && path[1][len(path[1])-1] == ']' {
					dP.Hosts[len(dP.Hosts)-1].Params[path[1][7:len(path[1])-1]] = val
					return
				}
				return utils.ErrWrongPath
			}
		}
	case 3:
		if path[0] != utils.Hosts ||
			path[1] != utils.Params {
			return utils.ErrWrongPath
		}
		if len(dP.Hosts) == 0 || newBranch {
			dP.Hosts = append(dP.Hosts, &DispatcherHostProfile{Params: make(map[string]interface{})})
		}
		dP.Hosts[len(dP.Hosts)-1].Params[path[2]] = val
	}
	return
}

func (dH *DispatcherHost) Set(path []string, val interface{}, newBranch bool, _ string) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}
	switch path[0] {
	default:
		return utils.ErrWrongPath
	case utils.Tenant:
		dH.Tenant = utils.IfaceAsString(val)
	case utils.ID:
		dH.ID = utils.IfaceAsString(val)
	case utils.Address:
		dH.Address = utils.IfaceAsString(val)
	case utils.Transport:
		dH.Transport = utils.IfaceAsString(val)
	case utils.ClientKey:
		dH.ClientKey = utils.IfaceAsString(val)
	case utils.ClientCertificate:
		dH.ClientCertificate = utils.IfaceAsString(val)
	case utils.CaCertificate:
		dH.CaCertificate = utils.IfaceAsString(val)
	case utils.ConnectAttempts:
		dH.ConnectAttempts, err = utils.IfaceAsTInt(val)
	case utils.Reconnects:
		dH.Reconnects, err = utils.IfaceAsTInt(val)
	case utils.ConnectTimeout:
		dH.ConnectTimeout, err = utils.IfaceAsDuration(val)
	case utils.ReplyTimeout:
		dH.ReplyTimeout, err = utils.IfaceAsDuration(val)
	case utils.TLS:
		dH.TLS, err = utils.IfaceAsBool(val)
	}
	return
}
