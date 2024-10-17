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
	TrendSConns    []string
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
	if jsnCfg.Trends_conns != nil {
		fSCfg.TrendSConns = make([]string, len(*jsnCfg.Trends_conns))
		for idx, connID := range *jsnCfg.Trends_conns {
			fSCfg.TrendSConns[idx] = connID
			if connID == utils.MetaInternal {
				fSCfg.TrendSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (fSCfg *FilterSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = make(map[string]any)
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
	if fSCfg.TrendSConns != nil {
		trendsConns := make([]string, len(fSCfg.TrendSConns))
		for i, item := range fSCfg.TrendSConns {
			trendsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends) {
				trendsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.TrendSConnsCfg] = trendsConns
	}
	return
}

// Clone returns a deep copy of FilterSCfg
func (fSCfg FilterSCfg) Clone() (cln *FilterSCfg) {
	cln = new(FilterSCfg)
	if fSCfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(fSCfg.StatSConns))
		copy(cln.StatSConns, fSCfg.StatSConns)
	}
	if fSCfg.ResourceSConns != nil {
		cln.ResourceSConns = make([]string, len(fSCfg.ResourceSConns))
		copy(cln.ResourceSConns, fSCfg.ResourceSConns)
	}
	if fSCfg.ApierSConns != nil {
		cln.ApierSConns = make([]string, len(fSCfg.ApierSConns))
		copy(cln.ApierSConns, fSCfg.ApierSConns)
	}
	if fSCfg.TrendSConns != nil {
		cln.TrendSConns = make([]string, len(fSCfg.TrendSConns))
		copy(cln.TrendSConns, fSCfg.TrendSConns)
	}
	return
}
