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
	ApierSConns    []string
}

func (fSCfg *FilterSCfg) loadFromJSONCfg(jsnCfg *FilterSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Stats_conns != nil {
		fSCfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			fSCfg.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				fSCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		fSCfg.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, connID := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			fSCfg.ResourceSConns[idx] = connID
			if connID == utils.MetaInternal {
				fSCfg.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			}
		}
	}
	if jsnCfg.Apiers_conns != nil {
		fSCfg.ApierSConns = make([]string, len(*jsnCfg.Apiers_conns))
		for idx, connID := range *jsnCfg.Apiers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			fSCfg.ApierSConns[idx] = connID
			if connID == utils.MetaInternal {
				fSCfg.ApierSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (fSCfg *FilterSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = make(map[string]interface{})
	if fSCfg.StatSConns != nil {
		statSConns := make([]string, len(fSCfg.StatSConns))
		for i, item := range fSCfg.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if fSCfg.ResourceSConns != nil {
		resourceSConns := make([]string, len(fSCfg.ResourceSConns))
		for i, item := range fSCfg.ResourceSConns {
			resourceSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resourceSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resourceSConns
	}
	if fSCfg.ApierSConns != nil {
		apierConns := make([]string, len(fSCfg.ApierSConns))
		for i, item := range fSCfg.ApierSConns {
			apierConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier) {
				apierConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ApierSConnsCfg] = apierConns
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
	if fSCfg.ApierSConns != nil {
		cln.ApierSConns = utils.CloneStringSlice(fSCfg.ApierSConns)
	}
	return
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Apiers_conns    *[]string
}

func diffFilterSJsonCfg(d *FilterSJsonCfg, v1, v2 *FilterSCfg) *FilterSJsonCfg {
	if d == nil {
		d = new(FilterSJsonCfg)
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = &v2.StatSConns
	}
	if !utils.SliceStringEqual(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = &v2.ResourceSConns
	}
	if !utils.SliceStringEqual(v1.ApierSConns, v2.ApierSConns) {
		d.Apiers_conns = &v2.ApierSConns
	}
	return d
}
