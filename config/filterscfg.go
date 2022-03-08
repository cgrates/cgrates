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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// FilterSCfg the filters config section
type FilterSCfg struct {
	StatSConns     []string
	ResourceSConns []string
	AccountSConns  []string
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
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (fSCfg FilterSCfg) AsMapInterface(string) interface{} {
	mp := make(map[string]interface{})
	if fSCfg.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		mp[utils.ResourceSConnsCfg] = getInternalJSONConns(fSCfg.ResourceSConns)
	}
	if fSCfg.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(fSCfg.AccountSConns)
	}
	return mp
}

func (FilterSCfg) SName() string               { return FilterSJSON }
func (fSCfg FilterSCfg) CloneSection() Section { return fSCfg.Clone() }

// Clone returns a deep copy of FilterSCfg
func (fSCfg FilterSCfg) Clone() (cln *FilterSCfg) {
	cln = new(FilterSCfg)
	if fSCfg.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		cln.ResourceSConns = utils.CloneStringSlice(fSCfg.ResourceSConns)
	}
	if fSCfg.AccountSConns != nil {
		cln.AccountSConns = utils.CloneStringSlice(fSCfg.AccountSConns)
	}
	return
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Accounts_conns  *[]string
}

func diffFilterSJsonCfg(d *FilterSJsonCfg, v1, v2 *FilterSCfg) *FilterSJsonCfg {
	if d == nil {
		d = new(FilterSJsonCfg)
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !utils.SliceStringEqual(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	return d
}
