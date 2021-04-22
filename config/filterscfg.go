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
	"github.com/cgrates/cgrates/utils"
)

// FilterSCfg the filters config section
type FilterSCfg struct {
	StatSConns     []string
	ResourceSConns []string
	AdminSConns    []string
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
	if jsnCfg.Admins_conns != nil {
		fSCfg.AdminSConns = updateInternalConns(*jsnCfg.Admins_conns, utils.MetaAdminS)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (fSCfg *FilterSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = make(map[string]interface{})
	if fSCfg.StatSConns != nil {
		initialMP[utils.StatSConnsCfg] = getInternalJSONConns(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		initialMP[utils.ResourceSConnsCfg] = getInternalJSONConns(fSCfg.ResourceSConns)
	}
	if fSCfg.AdminSConns != nil {
		initialMP[utils.AdminSConnsCfg] = getInternalJSONConns(fSCfg.AdminSConns)
	}
	return
}

// Clone returns a deep copy of FilterSCfg
func (fSCfg FilterSCfg) Clone() (cln *FilterSCfg) {
	cln = new(FilterSCfg)
	if fSCfg.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		cln.ResourceSConns = utils.CloneStringSlice(fSCfg.ResourceSConns)
	}
	if fSCfg.AdminSConns != nil {
		cln.AdminSConns = utils.CloneStringSlice(fSCfg.AdminSConns)
	}
	return
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Admins_conns    *[]string
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
	if !utils.SliceStringEqual(v1.AdminSConns, v2.AdminSConns) {
		d.Admins_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AdminSConns))
	}
	return d
}
