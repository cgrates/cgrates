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
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	ProcessRuns         int
	NestedFields        bool
}

func (alS *AttributeSCfg) loadFromJsonCfg(jsnCfg *AttributeSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
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

func (alS *AttributeSCfg) AsMapInterface() map[string]interface{} {
	initialMP := map[string]interface{}{
		utils.EnabledCfg:             alS.Enabled,
		utils.IndexedSelectsCfg:      alS.IndexedSelects,
		utils.ProcessRunsCfg:         alS.ProcessRuns,
		utils.NestedFieldsCfg:        alS.NestedFields,
	}
	if alS.StringIndexedFields  != nil {
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
	return initialMP
}
