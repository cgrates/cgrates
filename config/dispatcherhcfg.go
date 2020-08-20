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
	Enabled           bool
	DispatchersConns  []string
	HostIDs           map[string][]string
	RegisterInterval  time.Duration
	RegisterTransport string
	RegisterTLS       bool
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
	if jsnCfg.Host_ids != nil {
		for tnt, id := range jsnCfg.Host_ids {
			dps.HostIDs[tnt] = id
		}
	}
	if jsnCfg.Register_interval != nil {
		if dps.RegisterInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Register_interval); err != nil {
			return
		}
	}
	if jsnCfg.Register_transport != nil {
		dps.RegisterTransport = *jsnCfg.Register_transport
	}
	if jsnCfg.Register_tls != nil {
		dps.RegisterTLS = *jsnCfg.Register_tls
	}
	return
}

func (dps *DispatcherHCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.EnabledCfg:           dps.Enabled,
		utils.DispatchersConnsCfg:  dps.DispatchersConns,
		utils.HostIdsCfg:           dps.HostIDs,
		utils.RegisterIntervalCfg:  dps.RegisterInterval,
		utils.RegisterTransportCfg: dps.RegisterTransport,
	}

}
