// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// AttributeSCfg is the configuration of attribute service
type AttributeSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
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
	if jsnCfg.Process_runs != nil {
		alS.ProcessRuns = *jsnCfg.Process_runs
	}
	if jsnCfg.Nested_fields != nil {
		alS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

func (alS *AttributeSCfg) AsMapInterface() map[string]any {
	stringIndexedFields := []string{}
	if alS.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*alS.StringIndexedFields))
		for i, item := range *alS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if alS.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*alS.PrefixIndexedFields))
		for i, item := range *alS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	return map[string]any{
		utils.EnabledCfg:             alS.Enabled,
		utils.IndexedSelectsCfg:      alS.IndexedSelects,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.ProcessRunsCfg:         alS.ProcessRuns,
		utils.NestedFieldsCfg:        alS.NestedFields,
	}

}
