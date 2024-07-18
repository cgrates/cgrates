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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type TrendSCfg struct {
	Enabled         bool
	StatSConns      []string
	ThresholdSConns []string
}

func (t *TrendSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnTrendSCfg := new(TrendSJsonCfg)
	if err = jsnCfg.GetSection(ctx, TrendSJSON, jsnTrendSCfg); err != nil {
		return
	}
	return t.loadFromJSONCfg(jsnTrendSCfg)
}

func (t *TrendSCfg) loadFromJSONCfg(jsnCfg *TrendSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		t.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		t.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Thresholds_conns != nil {
		t.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	return
}

func (t *TrendSCfg) AsMapInterface(string) any {
	mp := map[string]any{
		utils.EnabledCfg: t.Enabled,
	}
	if t.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(t.StatSConns)
	}

	if t.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(t.ThresholdSConns)
	}
	return mp
}

func (TrendSCfg) SName() string           { return TrendSJSON }
func (t TrendSCfg) CloneSection() Section { return t.Clone() }

func (t *TrendSCfg) Clone() (cln *TrendSCfg) {
	cln = &TrendSCfg{
		Enabled: t.Enabled,
	}
	if t.StatSConns != nil {
		cln.StatSConns = slices.Clone(t.StatSConns)
	}
	if t.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(t.ThresholdSConns)
	}
	return
}

type TrendSJsonCfg struct {
	Enabled          *bool
	Stats_conns      *[]string
	Thresholds_conns *[]string
}

func diffTrendsJsonCfg(d *TrendSJsonCfg, v1, v2 *TrendSCfg) *TrendSJsonCfg {
	if d == nil {
		d = new(TrendSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	return d
}
