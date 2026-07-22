// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"slices"

	"github.com/cgrates/cgrates/utils"
)

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	ExistsIndexedFields *[]string
	AttributeSConns     []string
	NestedFields        bool
	AnySubsystem        bool
	PreventLoop         bool
}

func (dps *DispatcherSCfg) loadFromJSONCfg(jsnCfg *DispatcherSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		dps.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		dps.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		copy(sif, *jsnCfg.String_indexed_fields)
		dps.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		copy(pif, *jsnCfg.Prefix_indexed_fields)
		dps.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		copy(sif, *jsnCfg.Suffix_indexed_fields)
		dps.SuffixIndexedFields = &sif
	}
	if jsnCfg.ExistsIndexedFields != nil {
		eif := slices.Clone(*jsnCfg.ExistsIndexedFields)
		dps.ExistsIndexedFields = &eif
	}
	if jsnCfg.Attributes_conns != nil {
		dps.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, connID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			dps.AttributeSConns[idx] = connID
			if connID == utils.MetaInternal {
				dps.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Nested_fields != nil {
		dps.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Any_subsystem != nil {
		dps.AnySubsystem = *jsnCfg.Any_subsystem
	}
	if jsnCfg.Prevent_loop != nil {
		dps.PreventLoop = *jsnCfg.Prevent_loop
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (dps *DispatcherSCfg) AsMapInterface() (mp map[string]any) {
	mp = map[string]any{
		utils.EnabledCfg:        dps.Enabled,
		utils.IndexedSelectsCfg: dps.IndexedSelects,
		utils.NestedFieldsCfg:   dps.NestedFields,
		utils.AnySubsystemCfg:   dps.AnySubsystem,
		utils.PreventLoopCfg:    dps.PreventLoop,
	}
	if dps.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*dps.StringIndexedFields))
		copy(stringIndexedFields, *dps.StringIndexedFields)
		mp[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if dps.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*dps.PrefixIndexedFields))
		copy(prefixIndexedFields, *dps.PrefixIndexedFields)
		mp[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if dps.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*dps.SuffixIndexedFields))
		copy(suffixIndexedFields, *dps.SuffixIndexedFields)
		mp[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if dps.ExistsIndexedFields != nil {
		eif := slices.Clone(*dps.ExistsIndexedFields)
		mp[utils.ExistsIndexedFieldsCfg] = eif
	}
	if dps.AttributeSConns != nil {
		attributeSConns := make([]string, len(dps.AttributeSConns))
		for i, item := range dps.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		mp[utils.AttributeSConnsCfg] = attributeSConns
	}
	return
}

// Clone returns a deep copy of DispatcherSCfg
func (dps *DispatcherSCfg) Clone() (cln *DispatcherSCfg) {
	if dps == nil {
		return nil
	}
	cln = &DispatcherSCfg{
		Enabled:        dps.Enabled,
		IndexedSelects: dps.IndexedSelects,
		NestedFields:   dps.NestedFields,
		AnySubsystem:   dps.AnySubsystem,
		PreventLoop:    dps.PreventLoop,
	}

	if dps.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(dps.AttributeSConns))
		copy(cln.AttributeSConns, dps.AttributeSConns)
	}
	if dps.StringIndexedFields != nil {
		idx := make([]string, len(*dps.StringIndexedFields))
		copy(idx, *dps.StringIndexedFields)
		cln.StringIndexedFields = &idx
	}
	if dps.PrefixIndexedFields != nil {
		idx := make([]string, len(*dps.PrefixIndexedFields))
		copy(idx, *dps.PrefixIndexedFields)
		cln.PrefixIndexedFields = &idx
	}
	if dps.SuffixIndexedFields != nil {
		idx := make([]string, len(*dps.SuffixIndexedFields))
		copy(idx, *dps.SuffixIndexedFields)
		cln.SuffixIndexedFields = &idx
	}
	if dps.ExistsIndexedFields != nil {
		idx := slices.Clone(*dps.ExistsIndexedFields)
		cln.ExistsIndexedFields = &idx
	}
	return
}
