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

type LoaderCgrCfg struct {
	TpID           string
	DataPath       string
	DisableReverse bool
	FieldSeparator rune // The separator to use when reading csvs
	CachesConns    []*HaPoolConfig
	SchedulerConns []*HaPoolConfig
}

func (ld *LoaderCgrCfg) loadFromJsonCfg(jsnCfg *LoaderCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Tpid != nil {
		ld.TpID = *jsnCfg.Tpid
	}
	if jsnCfg.Data_path != nil {
		ld.DataPath = *jsnCfg.Data_path
	}
	if jsnCfg.Disable_reverse != nil {
		ld.DisableReverse = *jsnCfg.Disable_reverse
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 {
		sepStr := *jsnCfg.Field_separator
		ld.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Caches_conns != nil {
		ld.CachesConns = make([]*HaPoolConfig, len(*jsnCfg.Caches_conns))
		for idx, jsnHaCfg := range *jsnCfg.Caches_conns {
			ld.CachesConns[idx] = NewDfltHaPoolConfig()
			ld.CachesConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		ld.SchedulerConns = make([]*HaPoolConfig, len(*jsnCfg.Scheduler_conns))
		for idx, jsnScheHaCfg := range *jsnCfg.Scheduler_conns {
			ld.SchedulerConns[idx] = NewDfltHaPoolConfig()
			ld.SchedulerConns[idx].loadFromJsonCfg(jsnScheHaCfg)
		}
	}
	return nil
}
