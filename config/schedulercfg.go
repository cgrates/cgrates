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

type SchedulerCfg struct {
	Enabled   bool
	CDRsConns []*HaPoolConfig
}

func (schdcfg *SchedulerCfg) loadFromJsonCfg(jsnCfg *SchedulerJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		schdcfg.Enabled = *jsnCfg.Enabled
	}

	if jsnCfg.Cdrs_conns != nil {
		schdcfg.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			schdcfg.CDRsConns[idx] = NewDfltHaPoolConfig()
			schdcfg.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	return nil
}
