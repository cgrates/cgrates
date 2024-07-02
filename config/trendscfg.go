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

type TrendSCfg struct {
	Enabled         bool
	StatSConns      []string
	ThresholdSConns []string
}

func (sa *TrendSCfg) loadFromJSONCfg(jsnCfg *TrendsJsonCfg) (err error) {
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
	if jsnCfg.Thresholds_conns != nil {
		sa.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			sa.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				sa.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	return
}

func (sa *TrendSCfg) AsMapInterface() (initialMP map[string]any) {
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
	if sa.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(sa.ThresholdSConns))
		for i, item := range sa.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	return
}

func (sa *TrendSCfg) Clone() (cln *TrendSCfg) {
	cln = &TrendSCfg{
		Enabled: sa.Enabled,
	}
	if sa.StatSConns != nil {
		cln.StatSConns = make([]string, len(sa.StatSConns))
		copy(cln.StatSConns, sa.StatSConns)
	}
	if sa.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(sa.ThresholdSConns))
		copy(cln.ThresholdSConns, sa.ThresholdSConns)
	}
	return
}
