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

type SarSCfg struct {
	Enabled    bool
	StatSConns []string
}

func (sa *SarSCfg) loadFromJSONCfg(jsnCfg *SarsJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		sa.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		sa.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			sa.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				sa.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}

	}
	return
}

func (sa *SarSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg: sa.Enabled,
	}
	if sa.StatSConns != nil {
		statSConns := make([]string, len(sa.StatSConns))
		for i, item := range sa.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	return
}

func (sa *SarSCfg) Clone() (cln *SarSCfg) {
	cln = &SarSCfg{
		Enabled: sa.Enabled,
	}
	if sa.StatSConns != nil {
		cln.StatSConns = make([]string, len(sa.StatSConns))
		copy(cln.StatSConns, sa.StatSConns)
	}
	return
}
