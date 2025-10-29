/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"slices"

	"github.com/cgrates/cgrates/utils"
)

// ChargerSCfg is the configuration of charger service
type ChargerSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	AttributeSConns     []string
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	ExistsIndexedFields *[]string
	NestedFields        bool
}

func (cS *ChargerSCfg) loadFromJSONCfg(jsnCfg *ChargerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		cS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		cS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Attributes_conns != nil {
		cS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, attrConn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cS.AttributeSConns[idx] = attrConn
			if attrConn == utils.MetaInternal {
				cS.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		copy(sif, *jsnCfg.String_indexed_fields)
		cS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		copy(pif, *jsnCfg.Prefix_indexed_fields)
		cS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		copy(sif, *jsnCfg.Suffix_indexed_fields)
		cS.SuffixIndexedFields = &sif
	}
	if jsnCfg.ExistsIndexedFields != nil {
		eif := slices.Clone(*jsnCfg.ExistsIndexedFields)
		cS.ExistsIndexedFields = &eif
	}
	if jsnCfg.Nested_fields != nil {
		cS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cS *ChargerSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:        cS.Enabled,
		utils.IndexedSelectsCfg: cS.IndexedSelects,
		utils.NestedFieldsCfg:   cS.NestedFields,
	}
	if cS.AttributeSConns != nil {
		attributeSConns := make([]string, len(cS.AttributeSConns))
		for i, item := range cS.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if cS.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*cS.StringIndexedFields))
		copy(stringIndexedFields, *cS.StringIndexedFields)
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if cS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*cS.PrefixIndexedFields))
		copy(prefixIndexedFields, *cS.PrefixIndexedFields)
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if cS.SuffixIndexedFields != nil {
		sufixIndexedFields := make([]string, len(*cS.SuffixIndexedFields))
		copy(sufixIndexedFields, *cS.SuffixIndexedFields)
		initialMP[utils.SuffixIndexedFieldsCfg] = sufixIndexedFields
	}
	if cS.ExistsIndexedFields != nil {
		eif := slices.Clone(*cS.ExistsIndexedFields)
		initialMP[utils.ExistsIndexedFieldsCfg] = eif
	}
	return
}

// Clone returns a deep copy of ChargerSCfg
func (cS ChargerSCfg) Clone() (cln *ChargerSCfg) {
	cln = &ChargerSCfg{
		Enabled:        cS.Enabled,
		IndexedSelects: cS.IndexedSelects,
		NestedFields:   cS.NestedFields,
	}
	if cS.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(cS.AttributeSConns))
		copy(cln.AttributeSConns, cS.AttributeSConns)
	}

	if cS.StringIndexedFields != nil {
		idx := make([]string, len(*cS.StringIndexedFields))
		copy(idx, *cS.StringIndexedFields)
		cln.StringIndexedFields = &idx
	}
	if cS.PrefixIndexedFields != nil {
		idx := make([]string, len(*cS.PrefixIndexedFields))
		copy(idx, *cS.PrefixIndexedFields)
		cln.PrefixIndexedFields = &idx
	}
	if cS.SuffixIndexedFields != nil {
		idx := make([]string, len(*cS.SuffixIndexedFields))
		copy(idx, *cS.SuffixIndexedFields)
		cln.SuffixIndexedFields = &idx
	}
	if cS.ExistsIndexedFields != nil {
		idx := slices.Clone(*cS.ExistsIndexedFields)
		cln.ExistsIndexedFields = &idx
	}
	return
}
