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

	"github.com/cgrates/cgrates/utils"
)

type AttributesOpts struct {
	ProfileIDs           []string
	ProfileRuns          int
	ProfileIgnoreFilters bool
	ProcessRuns          int
	Context              *string
}

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
	ExistsIndexedFields *[]string
	NestedFields        bool
	AnyContext          bool
	Opts                *AttributesOpts
}

func (attrOpts *AttributesOpts) loadFromJSONCfg(jsnCfg *AttributesOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		attrOpts.ProfileIDs = *jsnCfg.ProfileIDs
	}
	if jsnCfg.ProfileRuns != nil {
		attrOpts.ProfileRuns = *jsnCfg.ProfileRuns
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		attrOpts.ProfileIgnoreFilters = *jsnCfg.ProfileIgnoreFilters
	}
	if jsnCfg.ProcessRuns != nil {
		attrOpts.ProcessRuns = *jsnCfg.ProcessRuns
	}
	if jsnCfg.Context != nil {
		attrOpts.Context = jsnCfg.Context
	}
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
		copy(sif, *jsnCfg.String_indexed_fields)
		alS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		copy(pif, *jsnCfg.Prefix_indexed_fields)
		alS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		copy(sif, *jsnCfg.Suffix_indexed_fields)
		alS.SuffixIndexedFields = &sif
	}
	if jsnCfg.ExistsIndexedFields != nil {
		eif := slices.Clone(*jsnCfg.ExistsIndexedFields)
		alS.ExistsIndexedFields = &eif
	}
	if jsnCfg.Nested_fields != nil {
		alS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Any_context != nil {
		alS.AnyContext = *jsnCfg.Any_context
	}
	if jsnCfg.Opts != nil {
		alS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (alS *AttributeSCfg) AsMapInterface() (initialMP map[string]any) {
	opts := map[string]any{
		utils.MetaProfileIDs:              alS.Opts.ProfileIDs,
		utils.MetaProfileRuns:             alS.Opts.ProfileRuns,
		utils.MetaProfileIgnoreFiltersCfg: alS.Opts.ProfileIgnoreFilters,
		utils.MetaProcessRuns:             alS.Opts.ProcessRuns,
	}
	if alS.Opts.Context != nil {
		opts[utils.OptsContext] = *alS.Opts.Context
	}
	initialMP = map[string]any{
		utils.EnabledCfg:        alS.Enabled,
		utils.IndexedSelectsCfg: alS.IndexedSelects,
		utils.NestedFieldsCfg:   alS.NestedFields,
		utils.AnyContextCfg:     alS.AnyContext,
		utils.OptsCfg:           opts,
	}
	if alS.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*alS.StringIndexedFields))
		copy(stringIndexedFields, *alS.StringIndexedFields)
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if alS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*alS.PrefixIndexedFields))
		copy(prefixIndexedFields, *alS.PrefixIndexedFields)
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if alS.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*alS.SuffixIndexedFields))
		copy(suffixIndexedFields, *alS.SuffixIndexedFields)
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if alS.ExistsIndexedFields != nil {
		eif := slices.Clone(*alS.ExistsIndexedFields)
		initialMP[utils.ExistsIndexedFieldsCfg] = eif
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

func (attrOpts *AttributesOpts) Clone() *AttributesOpts {
	cln := &AttributesOpts{
		ProfileIDs:           slices.Clone(attrOpts.ProfileIDs),
		ProfileRuns:          attrOpts.ProfileRuns,
		ProfileIgnoreFilters: attrOpts.ProfileIgnoreFilters,
		ProcessRuns:          attrOpts.ProcessRuns,
	}
	if attrOpts.Context != nil {
		cln.Context = new(string)
		*cln.Context = *attrOpts.Context
	}
	return cln
}

// Clone returns a deep copy of AttributeSCfg
func (alS AttributeSCfg) Clone() (cln *AttributeSCfg) {
	cln = &AttributeSCfg{
		Enabled:        alS.Enabled,
		IndexedSelects: alS.IndexedSelects,
		NestedFields:   alS.NestedFields,
		AnyContext:     alS.AnyContext,
		Opts:           alS.Opts.Clone(),
	}
	if alS.ResourceSConns != nil {
		cln.ResourceSConns = make([]string, len(alS.ResourceSConns))
		copy(cln.ResourceSConns, alS.ResourceSConns)
	}
	if alS.StatSConns != nil {
		cln.StatSConns = make([]string, len(alS.StatSConns))
		copy(cln.StatSConns, alS.StatSConns)
	}
	if alS.ApierSConns != nil {
		cln.ApierSConns = make([]string, len(alS.ApierSConns))
		copy(cln.ApierSConns, alS.ApierSConns)
	}

	if alS.StringIndexedFields != nil {
		idx := make([]string, len(*alS.StringIndexedFields))
		copy(idx, *alS.StringIndexedFields)
		cln.StringIndexedFields = &idx
	}
	if alS.PrefixIndexedFields != nil {
		idx := make([]string, len(*alS.PrefixIndexedFields))
		copy(idx, *alS.PrefixIndexedFields)
		cln.PrefixIndexedFields = &idx
	}
	if alS.SuffixIndexedFields != nil {
		idx := make([]string, len(*alS.SuffixIndexedFields))
		copy(idx, *alS.SuffixIndexedFields)
		cln.SuffixIndexedFields = &idx
	}
	if alS.ExistsIndexedFields != nil {
		idx := slices.Clone(*alS.ExistsIndexedFields)
		cln.ExistsIndexedFields = &idx
	}
	return
}
