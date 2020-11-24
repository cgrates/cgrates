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

// ChargerSCfg is the configuration of charger service
type ChargerSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	AttributeSConns     []string
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
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
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		cS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		cS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		cS.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		cS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (cS *ChargerSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
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
		for i, item := range *cS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if cS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*cS.PrefixIndexedFields))
		for i, item := range *cS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if cS.SuffixIndexedFields != nil {
		sufixIndexedFields := make([]string, len(*cS.SuffixIndexedFields))
		for i, item := range *cS.SuffixIndexedFields {
			sufixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = sufixIndexedFields
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
		for i, con := range cS.AttributeSConns {
			cln.AttributeSConns[i] = con
		}
	}

	if cS.StringIndexedFields != nil {
		idx := make([]string, len(*cS.StringIndexedFields))
		for i, dx := range *cS.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if cS.PrefixIndexedFields != nil {
		idx := make([]string, len(*cS.PrefixIndexedFields))
		for i, dx := range *cS.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if cS.SuffixIndexedFields != nil {
		idx := make([]string, len(*cS.SuffixIndexedFields))
		for i, dx := range *cS.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
