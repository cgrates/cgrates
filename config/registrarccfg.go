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
)

// RegistrarCCfgs is the configuration of registrarc rpc and dispatcher
type RegistrarCCfgs struct {
	RPC        *RegistrarCCfg
	Dispatcher *RegistrarCCfg
}

func (dps *RegistrarCCfgs) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfgs) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if err = dps.RPC.loadFromJSONCfg(jsnCfg.RPC); err != nil {
		return
	}
	return dps.Dispatcher.loadFromJSONCfg(jsnCfg.Dispatcher)
}

// AsMapInterface returns the config as a map[string]interface{}
func (dps *RegistrarCCfgs) AsMapInterface() (initialMP map[string]interface{}) {
	return map[string]interface{}{
		utils.RPCCfg:        dps.RPC.AsMapInterface(),
		utils.DispatcherCfg: dps.Dispatcher.AsMapInterface(),
	}
}

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfgs) Clone() (cln *RegistrarCCfgs) {
	return &RegistrarCCfgs{
		RPC:        dps.RPC.Clone(),
		Dispatcher: dps.Dispatcher.Clone(),
	}
}

// RegistrarCCfg is the configuration of registrarc
type RegistrarCCfg struct {
	Enabled         bool
	RegistrarSConns []string
	Hosts           map[string][]*RemoteHost
	RefreshInterval time.Duration
}

func (dps *RegistrarCCfg) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		dps.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Registrars_conns != nil {
		dps.RegistrarSConns = make([]string, len(*jsnCfg.Registrars_conns))
		copy(dps.RegistrarSConns, *jsnCfg.Registrars_conns)
	}
	if jsnCfg.Hosts != nil {
		for tnt, hosts := range jsnCfg.Hosts {
			for _, hostJSON := range hosts {
				conn := NewDfltRemoteHost()
				conn.loadFromJSONCfg(hostJSON)
				dps.Hosts[tnt] = append(dps.Hosts[tnt], conn)
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

// AsMapInterface returns the config as a map[string]interface{}
func (dps *RegistrarCCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:         dps.Enabled,
		utils.RegistrarsConnsCfg: dps.RegistrarSConns,
		utils.RefreshIntervalCfg: dps.RefreshInterval.String(),
	}
	if dps.RefreshInterval == 0 {
		initialMP[utils.RefreshIntervalCfg] = "0"
	}
	if dps.Hosts != nil {
		hosts := make(map[string][]map[string]interface{})
		for tnt, hs := range dps.Hosts {
			for _, h := range hs {
				hosts[tnt] = append(hosts[tnt], h.AsMapInterface())
			}
		}
		initialMP[utils.HostsCfg] = hosts
	}
	return
}

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfg) Clone() (cln *RegistrarCCfg) {
	cln = &RegistrarCCfg{
		Enabled:         dps.Enabled,
		RefreshInterval: dps.RefreshInterval,
		Hosts:           make(map[string][]*RemoteHost),
	}
	if dps.RegistrarSConns != nil {
		cln.RegistrarSConns = make([]string, len(dps.RegistrarSConns))
		for i, k := range dps.RegistrarSConns {
			cln.RegistrarSConns[i] = k
		}
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
