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

// AttributeSCfg is the configuration of attribute service
type AttributeSCfg struct {
	Enabled             bool
	ResourceSConns      []string
	StatSConns          []string
	ApierSConns         []string
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	ProcessRuns         int
	NestedFields        bool
}

func (alS *AttributeSCfg) loadFromJSONCfg(jsnCfg *AttributeSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		alS.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			alS.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				alS.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		alS.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, connID := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			alS.ResourceSConns[idx] = connID
			if connID == utils.MetaInternal {
				alS.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			}
		}
	}
	if jsnCfg.Apiers_conns != nil {
		alS.ApierSConns = make([]string, len(*jsnCfg.Apiers_conns))
		for idx, connID := range *jsnCfg.Apiers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			alS.ApierSConns[idx] = connID
			if connID == utils.MetaInternal {
				alS.ApierSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
			}
		}
	}
	if jsnCfg.Indexed_selects != nil {
		alS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		alS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		alS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		alS.SuffixIndexedFields = &sif
	}
	if jsnCfg.Process_runs != nil {
		alS.ProcessRuns = *jsnCfg.Process_runs
	}
	if jsnCfg.Nested_fields != nil {
		alS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (alS *AttributeSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        alS.Enabled,
		utils.IndexedSelectsCfg: alS.IndexedSelects,
		utils.ProcessRunsCfg:    alS.ProcessRuns,
		utils.NestedFieldsCfg:   alS.NestedFields,
	}
	if alS.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*alS.StringIndexedFields))
		for i, item := range *alS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if alS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*alS.PrefixIndexedFields))
		for i, item := range *alS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if alS.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*alS.SuffixIndexedFields))
		for i, item := range *alS.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if alS.StatSConns != nil {
		statSConns := make([]string, len(alS.StatSConns))
		for i, item := range alS.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}

	if alS.ResourceSConns != nil {
		resourceSConns := make([]string, len(alS.ResourceSConns))
		for i, item := range alS.ResourceSConns {
			resourceSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resourceSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resourceSConns
	}
	if alS.ApierSConns != nil {
		apierSConns := make([]string, len(alS.ApierSConns))
		for i, item := range alS.ApierSConns {
			apierSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier) {
				apierSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ApierSConnsCfg] = apierSConns
	}
	return
}

// Clone returns a deep copy of AttributeSCfg
func (alS AttributeSCfg) Clone() (cln *AttributeSCfg) {
	cln = &AttributeSCfg{
		Enabled:        alS.Enabled,
		IndexedSelects: alS.IndexedSelects,
		NestedFields:   alS.NestedFields,
		ProcessRuns:    alS.ProcessRuns,
	}
	if alS.ResourceSConns != nil {
		cln.ResourceSConns = make([]string, len(alS.ResourceSConns))
		for i, con := range alS.ResourceSConns {
			cln.ResourceSConns[i] = con
		}
	}
	if alS.StatSConns != nil {
		cln.StatSConns = make([]string, len(alS.StatSConns))
		for i, con := range alS.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if alS.ApierSConns != nil {
		cln.ApierSConns = make([]string, len(alS.ApierSConns))
		for i, con := range alS.ApierSConns {
			cln.ApierSConns[i] = con
		}
	}

	if alS.StringIndexedFields != nil {
		idx := make([]string, len(*alS.StringIndexedFields))
		for i, dx := range *alS.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if alS.PrefixIndexedFields != nil {
		idx := make([]string, len(*alS.PrefixIndexedFields))
		for i, dx := range *alS.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if alS.SuffixIndexedFields != nil {
		idx := make([]string, len(*alS.SuffixIndexedFields))
		for i, dx := range *alS.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
