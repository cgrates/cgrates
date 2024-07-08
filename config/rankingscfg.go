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

import "github.com/cgrates/cgrates/utils"

type RankingSCfg struct {
	Enabled    bool
	StatSConns []string
}

func (sgsCfg *RankingSCfg) loadFromJSONCfg(jsnCfg *RankingsJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		sgsCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		sgsCfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			sgsCfg.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				sgsCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	return
}

func (sgsCfg *RankingSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg: sgsCfg.Enabled,
	}
	if sgsCfg.StatSConns != nil {
		statSConns := make([]string, len(sgsCfg.StatSConns))
		for i, item := range sgsCfg.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	return
}

func (sgscfg *RankingSCfg) Clone() (cln *RankingSCfg) {
	cln = &RankingSCfg{
		Enabled: sgscfg.Enabled,
	}
	if sgscfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(sgscfg.StatSConns))
		copy(cln.StatSConns, sgscfg.StatSConns)
	}
	return
}
