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

// FilterSCfg the filters config section
type FilterSCfg struct {
	StatSConns     []string
	ResourceSConns []string
	AccountSConns  []string
	TrendSConns    []string
	RankingSConns  []string
}

// loadFilterSCfg loads the FilterS section of the configuration
func (fSCfg *FilterSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnFilterSCfg := new(FilterSJsonCfg)
	if err = jsnCfg.GetSection(ctx, FilterSJSON, jsnFilterSCfg); err != nil {
		return
	}
	return fSCfg.loadFromJSONCfg(jsnFilterSCfg)
}

func (fSCfg *FilterSCfg) loadFromJSONCfg(jsnCfg *FilterSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Stats_conns != nil {
		fSCfg.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Resources_conns != nil {
		fSCfg.ResourceSConns = updateInternalConns(*jsnCfg.Resources_conns, utils.MetaResources)
	}
	if jsnCfg.Accounts_conns != nil {
		fSCfg.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Trends_conns != nil {
		fSCfg.TrendSConns = updateInternalConns(*jsnCfg.Trends_conns, utils.MetaTrends)
	}
	if jsnCfg.Rankings_conns != nil {
		fSCfg.RankingSConns = updateInternalConns(*jsnCfg.Rankings_conns, utils.MetaRankings)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (fSCfg FilterSCfg) AsMapInterface() any {
	mp := make(map[string]any)
	if fSCfg.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		mp[utils.ResourceSConnsCfg] = getInternalJSONConns(fSCfg.ResourceSConns)
	}
	if fSCfg.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(fSCfg.AccountSConns)
	}
	if fSCfg.TrendSConns != nil {
		mp[utils.TrendSConnsCfg] = getInternalJSONConns(fSCfg.TrendSConns)
	}
	if fSCfg.RankingSConns != nil {
		mp[utils.RankingSConnsCfg] = getInternalJSONConns(fSCfg.RankingSConns)
	}
	return mp
}

func (FilterSCfg) SName() string               { return FilterSJSON }
func (fSCfg FilterSCfg) CloneSection() Section { return fSCfg.Clone() }

// Clone returns a deep copy of FilterSCfg
func (fSCfg FilterSCfg) Clone() (cln *FilterSCfg) {
	cln = new(FilterSCfg)
	if fSCfg.StatSConns != nil {
		cln.StatSConns = slices.Clone(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		cln.ResourceSConns = slices.Clone(fSCfg.ResourceSConns)
	}
	if fSCfg.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(fSCfg.AccountSConns)
	}
	if fSCfg.TrendSConns != nil {
		cln.TrendSConns = slices.Clone(fSCfg.TrendSConns)
	}
	if fSCfg.RankingSConns != nil {
		cln.RankingSConns = slices.Clone(fSCfg.RankingSConns)
	}
	return
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Accounts_conns  *[]string
	Trends_conns    *[]string
	Rankings_conns  *[]string
}

func diffFilterSJsonCfg(d *FilterSJsonCfg, v1, v2 *FilterSCfg) *FilterSJsonCfg {
	if d == nil {
		d = new(FilterSJsonCfg)
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	if !slices.Equal(v1.TrendSConns, v2.TrendSConns) {
		d.Trends_conns = utils.SliceStringPointer(getInternalJSONConns(v2.TrendSConns))
	}
	if !slices.Equal(v1.RankingSConns, v2.RankingSConns) {
		d.Rankings_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RankingSConns))
	}
	return d
}
