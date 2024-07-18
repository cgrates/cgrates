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

type RankingSCfg struct {
	Enabled    bool
	StatSConns []string
}

func (rnk *RankingSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRankingSCfg := new(RankingSJsonCfg)
	if err = jsnCfg.GetSection(ctx, RankingSJSON, jsnRankingSCfg); err != nil {
		return
	}
	return rnk.loadFromJSONCfg(jsnRankingSCfg)
}

func (rnk *RankingSCfg) loadFromJSONCfg(jsnCfg *RankingSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		rnk.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		rnk.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	return
}

func (rnk *RankingSCfg) AsMapInterface(string) any {
	mp := map[string]any{
		utils.EnabledCfg: rnk.Enabled,
	}
	if rnk.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(rnk.StatSConns)
	}
	return mp
}

func (RankingSCfg) SName() string             { return RankingSJSON }
func (rnk RankingSCfg) CloneSection() Section { return rnk.Clone() }

func (rnk *RankingSCfg) Clone() (cln *RankingSCfg) {
	cln = &RankingSCfg{
		Enabled: rnk.Enabled,
	}
	if rnk.StatSConns != nil {
		cln.StatSConns = slices.Clone(rnk.StatSConns)
	}
	return
}

type RankingSJsonCfg struct {
	Enabled     *bool
	Stats_conns *[]string
}

func diffRankingsJsonCfg(d *RankingSJsonCfg, v1, v2 *RankingSCfg) *RankingSJsonCfg {
	if d == nil {
		d = new(RankingSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	return d
}
