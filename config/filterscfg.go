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

type FilterSCfg struct {
	StatSConns     []*RemoteHost
	ResourceSConns []*RemoteHost
	RALsConns      []*RemoteHost
}

func (fSCfg *FilterSCfg) loadFromJsonCfg(jsnCfg *FilterSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Stats_conns != nil {
		fSCfg.StatSConns = make([]*RemoteHost, len(*jsnCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Stats_conns {
			fSCfg.StatSConns[idx] = NewDfltRemoteHost()
			fSCfg.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Resources_conns != nil {
		fSCfg.ResourceSConns = make([]*RemoteHost, len(*jsnCfg.Resources_conns))
		for idx, jsnHaCfg := range *jsnCfg.Resources_conns {
			fSCfg.ResourceSConns[idx] = NewDfltRemoteHost()
			fSCfg.ResourceSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Rals_conns != nil {
		fSCfg.RALsConns = make([]*RemoteHost, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			fSCfg.RALsConns[idx] = NewDfltRemoteHost()
			fSCfg.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	return
}
