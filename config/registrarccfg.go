/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// RegistrarCCfgs is the configuration of registrarc rpc and dispatcher
type RegistrarCCfgs struct {
	RPC         *RegistrarCCfg
	Dispatchers *RegistrarCCfg
}

func (dps *RegistrarCCfgs) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfgs) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if err = dps.RPC.loadFromJSONCfg(jsnCfg.RPC); err != nil {
		return
	}
	return dps.Dispatchers.loadFromJSONCfg(jsnCfg.Dispatchers)
}

// AsMapInterface returns the config as a map[string]any
func (dps *RegistrarCCfgs) AsMapInterface() (initialMP map[string]any) {
	return map[string]any{
		utils.RPCCfg:        dps.RPC.AsMapInterface(),
		utils.DispatcherCfg: dps.Dispatchers.AsMapInterface(),
	}
}

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfgs) Clone() (cln *RegistrarCCfgs) {
	return &RegistrarCCfgs{
		RPC:         dps.RPC.Clone(),
		Dispatchers: dps.Dispatchers.Clone(),
	}
}

// RegistrarCCfg is the configuration of registrarc
type RegistrarCCfg struct {
	RegistrarSConns []string
	Hosts           map[string][]*RemoteHost
	RefreshInterval time.Duration
}

type RemoteHostJsonWithTenant struct {
	*RemoteHostJson
	Tenant *string
}

func (dps *RegistrarCCfg) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Registrars_conns != nil {
		dps.RegistrarSConns = make([]string, len(*jsnCfg.Registrars_conns))
		copy(dps.RegistrarSConns, *jsnCfg.Registrars_conns)
	}
	if jsnCfg.Hosts != nil {
		for _, hostJSON := range jsnCfg.Hosts {
			conn := new(RemoteHost)
			conn.loadFromJSONCfg(hostJSON.RemoteHostJson)
			if hostJSON.Tenant == nil || *hostJSON.Tenant == "" {
				dps.Hosts[utils.MetaDefault] = append(dps.Hosts[utils.MetaDefault], conn)
			} else {
				dps.Hosts[*hostJSON.Tenant] = append(dps.Hosts[*hostJSON.Tenant], conn)
			}
		}
	}
	if jsnCfg.Refresh_interval != nil {
		if dps.RefreshInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Refresh_interval); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dps *RegistrarCCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.RegistrarsConnsCfg: dps.RegistrarSConns,
		utils.RefreshIntervalCfg: dps.RefreshInterval.String(),
	}
	if dps.RefreshInterval == 0 {
		initialMP[utils.RefreshIntervalCfg] = "0"
	}
	if dps.Hosts != nil {
		hosts := []map[string]any{}
		for tnt, hs := range dps.Hosts {
			for _, h := range hs {
				mp := h.AsMapInterface()
				delete(mp, utils.AddressCfg)
				mp[utils.Tenant] = tnt
				hosts = append(hosts, mp)
			}
		}
		initialMP[utils.HostsCfg] = hosts
	}
	return
}

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfg) Clone() (cln *RegistrarCCfg) {
	cln = &RegistrarCCfg{
		RefreshInterval: dps.RefreshInterval,
		Hosts:           make(map[string][]*RemoteHost),
	}
	if dps.RegistrarSConns != nil {
		cln.RegistrarSConns = make([]string, len(dps.RegistrarSConns))
		copy(cln.RegistrarSConns, dps.RegistrarSConns)
	}
	for tnt, hosts := range dps.Hosts {
		clnH := make([]*RemoteHost, len(hosts))
		for i, host := range hosts {
			clnH[i] = host.Clone()
		}
		cln.Hosts[tnt] = clnH
	}
	return
}
