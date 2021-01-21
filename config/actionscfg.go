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

// ActionSCfg is the configuration of ActionS
type ActionSCfg struct {
	Enabled             bool
	CDRsConns           []string
	Tenants             *[]string
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
}

func (acS *ActionSCfg) loadFromJSONCfg(jsnCfg *ActionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Cdrs_conns != nil {
		acS.CDRsConns = make([]string, len(*jsnCfg.Cdrs_conns))
		for idx, connID := range *jsnCfg.Cdrs_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			acS.CDRsConns[idx] = connID
			if connID == utils.MetaInternal {
				acS.CDRsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)
			}
		}
	}
	if jsnCfg.Enabled != nil {
		acS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Tenants != nil {
		tnt := make([]string, len(*jsnCfg.Tenants))
		for i, fID := range *jsnCfg.Tenants {
			tnt[i] = fID
		}
		acS.Tenants = &tnt
	}
	if jsnCfg.Indexed_selects != nil {
		acS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		acS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		acS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		acS.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		acS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (acS *ActionSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        acS.Enabled,
		utils.IndexedSelectsCfg: acS.IndexedSelects,
		utils.NestedFieldsCfg:   acS.NestedFields,
	}
	if acS.CDRsConns != nil {
		CDRsConns := make([]string, len(acS.CDRsConns))
		for i, item := range acS.CDRsConns {
			CDRsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs) {
				CDRsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CDRsConnsCfg] = CDRsConns
	}
	if acS.Tenants != nil {
		Tenants := make([]string, len(*acS.Tenants))
		for i, item := range *acS.Tenants {
			Tenants[i] = item
		}
		initialMP[utils.Tenants] = Tenants
	}
	if acS.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*acS.StringIndexedFields))
		for i, item := range *acS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if acS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*acS.PrefixIndexedFields))
		for i, item := range *acS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if acS.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*acS.SuffixIndexedFields))
		for i, item := range *acS.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	return
}

// Clone returns a deep copy of ActionSCfg
func (acS ActionSCfg) Clone() (cln *ActionSCfg) {
	cln = &ActionSCfg{
		Enabled:        acS.Enabled,
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
	}
	if acS.CDRsConns != nil {
		cln.CDRsConns = make([]string, len(acS.CDRsConns))
		for i, con := range acS.CDRsConns {
			cln.CDRsConns[i] = con
		}
	}
	if acS.Tenants != nil {
		tnt := make([]string, len(*acS.Tenants))
		for i, dx := range *acS.Tenants {
			tnt[i] = dx
		}
		cln.Tenants = &tnt
	}
	if acS.StringIndexedFields != nil {
		idx := make([]string, len(*acS.StringIndexedFields))
		for i, dx := range *acS.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if acS.PrefixIndexedFields != nil {
		idx := make([]string, len(*acS.PrefixIndexedFields))
		for i, dx := range *acS.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if acS.SuffixIndexedFields != nil {
		idx := make([]string, len(*acS.SuffixIndexedFields))
		for i, dx := range *acS.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
