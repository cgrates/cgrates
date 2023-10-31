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
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type DispatcherHostProfile struct {
	ID        string
	FilterIDs []string
	Weight    float64        // applied in case of multiple connections need to be ordered
	Params    map[string]any // additional parameters stored for a session
	Blocker   bool           // no connection after this one
}

func (dC *DispatcherHostProfile) Clone() (cln *DispatcherHostProfile) {
	cln = &DispatcherHostProfile{
		ID:      dC.ID,
		Weight:  dC.Weight,
		Blocker: dC.Blocker,
	}
	if dC.FilterIDs != nil {
		cln.FilterIDs = slices.Clone(dC.FilterIDs)
	}
	if dC.Params != nil {
		cln.Params = make(map[string]any)
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
	StrategyParams map[string]any         // ie for distribution, set here the pool weights
	Weight         float64                // used for profile sorting on match
	Hosts          DispatcherHostProfiles // dispatch to these connections
}

// DispatcherProfileWithAPIOpts is used in replicatorV1 for dispatcher
type DispatcherProfileWithAPIOpts struct {
	*DispatcherProfile
	APIOpts map[string]any
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
	APIOpts map[string]any
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
			cfg.GeneralCfg().MaxReconnectInterval, cfg.GeneralCfg().ConnectTimeout,
			cfg.GeneralCfg().ReplyTimeout,
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

func (dP *DispatcherProfile) Set(path []string, val any, newBranch bool, _ string) (err error) {
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
			if val != utils.EmptyString {
				dP.Weight, err = utils.IfaceAsFloat64(val)
			}
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
				dP.Hosts = append(dP.Hosts, &DispatcherHostProfile{Params: make(map[string]any)})
			}
			switch path[1] {
			case utils.ID:
				dP.Hosts[len(dP.Hosts)-1].ID = utils.IfaceAsString(val)
			case utils.FilterIDs:
				var valA []string
				valA, err = utils.IfaceAsStringSlice(val)
				dP.Hosts[len(dP.Hosts)-1].FilterIDs = append(dP.Hosts[len(dP.Hosts)-1].FilterIDs, valA...)
			case utils.Weight:
				if val != utils.EmptyString {
					dP.Hosts[len(dP.Hosts)-1].Weight, err = utils.IfaceAsFloat64(val)
				}
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
			dP.Hosts = append(dP.Hosts, &DispatcherHostProfile{Params: make(map[string]any)})
		}
		dP.Hosts[len(dP.Hosts)-1].Params[path[2]] = val
	}
	return
}

func (dH *DispatcherHost) Set(path []string, val any, newBranch bool, _ string) (err error) {
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
		if val != utils.EmptyString {
			dH.ConnectAttempts, err = utils.IfaceAsInt(val)
		}
	case utils.Reconnects:
		if val != utils.EmptyString {
			dH.Reconnects, err = utils.IfaceAsInt(val)
		}
	case utils.MaxReconnectInterval:
		if val != utils.EmptyString {
			dH.MaxReconnectInterval, err = utils.IfaceAsDuration(val)
		}
	case utils.ConnectTimeout:
		dH.ConnectTimeout, err = utils.IfaceAsDuration(val)
	case utils.ReplyTimeout:
		dH.ReplyTimeout, err = utils.IfaceAsDuration(val)
	case utils.TLS:
		dH.TLS, err = utils.IfaceAsBool(val)
	}
	return
}

func (dP *DispatcherProfile) Merge(v2 any) {
	vi := v2.(*DispatcherProfile)
	if len(vi.Tenant) != 0 {
		dP.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		dP.ID = vi.ID
	}
	dP.FilterIDs = append(dP.FilterIDs, vi.FilterIDs...)
	if len(dP.Hosts) == 1 && dP.Hosts[0].ID == utils.EmptyString {
		dP.Hosts = dP.Hosts[:0]
	}
	var equal bool
	for _, hostV2 := range vi.Hosts {
		for _, host := range dP.Hosts {
			if host.ID == hostV2.ID {
				host.Merge(hostV2)
				equal = true
				break
			}
		}
		if !equal && hostV2.ID != utils.EmptyString {
			dP.Hosts = append(dP.Hosts, hostV2)
		}
		equal = false
	}
	if vi.Weight != 0 {
		dP.Weight = vi.Weight
	}
	for k, v := range vi.StrategyParams {
		dP.StrategyParams[k] = v
	}
	if len(vi.Strategy) != 0 {
		dP.Strategy = vi.Strategy
	}
}

func (dspHost *DispatcherHostProfile) Merge(v2 *DispatcherHostProfile) {
	if v2.ID != utils.EmptyString {
		dspHost.ID = v2.ID
	}
	if v2.Weight != 0 {
		dspHost.Weight = v2.Weight
	}
	if v2.Blocker {
		dspHost.Blocker = v2.Blocker
	}
	dspHost.FilterIDs = append(dspHost.FilterIDs, v2.FilterIDs...)
	for k, v := range v2.Params {
		dspHost.Params[k] = v
	}
}

func (dH *DispatcherHost) Merge(v2 any) {
	vi := v2.(*DispatcherHost)
	if len(vi.Tenant) != 0 {
		dH.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		dH.ID = vi.ID
	}
	if len(vi.Address) != 0 {
		dH.Address = vi.Address
	}
	if len(vi.Transport) != 0 {
		dH.Transport = vi.Transport
	}
	if len(vi.ClientKey) != 0 {
		dH.ClientKey = vi.ClientKey
	}
	if len(vi.ClientCertificate) != 0 {
		dH.ClientCertificate = vi.ClientCertificate
	}
	if len(vi.CaCertificate) != 0 {
		dH.CaCertificate = vi.CaCertificate
	}
	if vi.TLS {
		dH.TLS = vi.TLS
	}
	if vi.ConnectTimeout != 0 {
		dH.ConnectTimeout = vi.ConnectTimeout
	}
	if vi.ReplyTimeout != 0 {
		dH.ReplyTimeout = vi.ReplyTimeout
	}
	if vi.ConnectAttempts != 0 {
		dH.ConnectAttempts = vi.ConnectAttempts
	}
	if vi.Reconnects != 0 {
		dH.Reconnects = vi.Reconnects
	}
	if vi.MaxReconnectInterval != 0 {
		dH.MaxReconnectInterval = vi.MaxReconnectInterval
	}
}

func (dH *DispatcherHost) String() string { return utils.ToJSON(dH) }
func (dH *DispatcherHost) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = dH.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (dH *DispatcherHost) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, utils.ErrNotFound
	case utils.Tenant:
		return dH.Tenant, nil
	case utils.ID:
		return dH.ID, nil
	case utils.Address:
		return dH.Address, nil
	case utils.Transport:
		return dH.Transport, nil
	case utils.ConnectAttempts:
		return dH.ConnectAttempts, nil
	case utils.Reconnects:
		return dH.Reconnects, nil
	case utils.MaxReconnectInterval:
		return dH.MaxReconnectInterval, nil
	case utils.ConnectTimeout:
		return dH.ConnectTimeout, nil
	case utils.ReplyTimeout:
		return dH.ReplyTimeout, nil
	case utils.TLS:
		return dH.TLS, nil
	case utils.ClientKey:
		return dH.ClientKey, nil
	case utils.ClientCertificate:
		return dH.ClientCertificate, nil
	case utils.CaCertificate:
		return dH.CaCertificate, nil
	}
}

func (dP *DispatcherProfile) String() string { return utils.ToJSON(dP) }
func (dP *DispatcherProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = dP.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (dP *DispatcherProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := utils.GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case utils.Hosts:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(dP.Hosts) {
						return dP.Hosts[idx], nil
					}
				case utils.FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(dP.FilterIDs) {
						return dP.FilterIDs[idx], nil
					}
				case utils.StrategyParams:
					return utils.MapStorage(dP.StrategyParams).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return dP.Tenant, nil
		case utils.ID:
			return dP.ID, nil
		case utils.FilterIDs:
			return dP.FilterIDs, nil
		case utils.Weight:
			return dP.Weight, nil
		case utils.Hosts:
			return dP.Hosts, nil
		case utils.Strategy:
			return dP.Strategy, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idxStr := utils.GetPathIndexString(fldPath[0])
	switch fld {
	default:
		return nil, utils.ErrNotFound
	case utils.StrategyParams:
		path := fldPath[1:]
		if idxStr != nil {
			path = append([]string{*idxStr}, path...)
		}
		return utils.MapStorage(dP.StrategyParams).FieldAsInterface(path)
	case utils.Hosts:
		if idxStr == nil {
			return nil, utils.ErrNotFound
		}
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(dP.Hosts) {
			return nil, utils.ErrNotFound
		}
		return dP.Hosts[idx].FieldAsInterface(fldPath[1:])
	}
}

func (dC *DispatcherHostProfile) String() string { return utils.ToJSON(dC) }
func (dC *DispatcherHostProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = dC.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (dC *DispatcherHostProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := utils.GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case utils.FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(dC.FilterIDs) {
						return dC.FilterIDs[idx], nil
					}
				case utils.Params:
					return utils.MapStorage(dC.Params).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, utils.ErrNotFound
		case utils.ID:
			return dC.ID, nil
		case utils.FilterIDs:
			return dC.FilterIDs, nil
		case utils.Weight:
			return dC.Weight, nil
		case utils.Blocker:
			return dC.Blocker, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idxStr := utils.GetPathIndexString(fldPath[0])
	if fld != utils.Params {
		return nil, utils.ErrNotFound
	}
	path := fldPath[1:]
	if idxStr != nil {
		path = append([]string{*idxStr}, path...)
	}
	return utils.MapStorage(dC.Params).FieldAsInterface(path)
}
