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

// DispatcherHCfg is the configuration of dispatcher hosts
type DispatcherHCfg struct {
	Enabled          bool
	DispatchersConns []string
	Hosts            map[string][]*DispatcherHRegistarCfg
	RegisterInterval time.Duration
}

func (dps *DispatcherHCfg) loadFromJsonCfg(jsnCfg *DispatcherHJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		dps.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Dispatchers_conns != nil {
		dps.DispatchersConns = make([]string, len(*jsnCfg.Dispatchers_conns))
		copy(dps.DispatchersConns, *jsnCfg.Dispatchers_conns)
	}
	if jsnCfg.Hosts != nil {
		for tnt, hosts := range jsnCfg.Hosts {
			for _, hostJSON := range hosts {
				dps.Hosts[tnt] = append(dps.Hosts[tnt], NewDispatcherHRegistarCfg(hostJSON))
			}
		}
	}
	if jsnCfg.Register_interval != nil {
		if dps.RegisterInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Register_interval); err != nil {
			return
		}
	}
	return
}

func (dps *DispatcherHCfg) AsMapInterface() map[string]interface{} {
	hosts := make(map[string][]map[string]interface{})
	for tnt, hs := range dps.Hosts {
		for _, h := range hs {
			hosts[tnt] = append(hosts[tnt], h.AsMapInterface())
		}
	}
	return map[string]interface{}{
		utils.EnabledCfg:          dps.Enabled,
		utils.DispatchersConnsCfg: dps.DispatchersConns,
		utils.HostsCfg:            hosts,
		utils.RegisterIntervalCfg: dps.RegisterInterval,
	}
}

type DispatcherHRegistarCfg struct {
	ID                string
	RegisterTransport string
	RegisterTLS       bool
}

func NewDispatcherHRegistarCfg(jsnCfg DispatcherHRegistarJsonCfg) (dhr *DispatcherHRegistarCfg) {
	dhr = new(DispatcherHRegistarCfg)
	if jsnCfg.Id != nil {
		dhr.ID = *jsnCfg.Id
	}
	dhr.RegisterTransport = utils.MetaJSON
	if jsnCfg.Register_transport != nil {
		dhr.RegisterTransport = *jsnCfg.Register_transport
	}
	if jsnCfg.Register_tls != nil {
		dhr.RegisterTLS = *jsnCfg.Register_tls
	}
	return
}

func (dhr *DispatcherHRegistarCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.IDCfg:                dhr.ID,
		utils.RegisterTransportCfg: dhr.RegisterTransport,
		utils.RegisterTLSCfg:       dhr.RegisterTLS,
	}
}
