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

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	AttributeSConns     []string
	NestedFields        bool
	AnySubsystem        bool
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
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		dps.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		dps.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		dps.SuffixIndexedFields = &sif
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
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (dps *DispatcherSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        dps.Enabled,
		utils.IndexedSelectsCfg: dps.IndexedSelects,
		utils.NestedFieldsCfg:   dps.NestedFields,
		utils.AnySubsystemCfg:   dps.AnySubsystem,
	}
	if dps.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*dps.StringIndexedFields))
		for i, item := range *dps.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if dps.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*dps.PrefixIndexedFields))
		for i, item := range *dps.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if dps.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*dps.SuffixIndexedFields))
		for i, item := range *dps.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if dps.AttributeSConns != nil {
		attributeSConns := make([]string, len(dps.AttributeSConns))
		for i, item := range dps.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	return
}

// Clone returns a deep copy of DispatcherSCfg
func (dps DispatcherSCfg) Clone() (cln *DispatcherSCfg) {
	cln = &DispatcherSCfg{
		Enabled:        dps.Enabled,
		IndexedSelects: dps.IndexedSelects,
		NestedFields:   dps.NestedFields,
		AnySubsystem:   dps.AnySubsystem,
	}

	if dps.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(dps.AttributeSConns))
		for i, conn := range dps.AttributeSConns {
			cln.AttributeSConns[i] = conn
		}
	}
	if dps.StringIndexedFields != nil {
		idx := make([]string, len(*dps.StringIndexedFields))
		for i, dx := range *dps.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if dps.PrefixIndexedFields != nil {
		idx := make([]string, len(*dps.PrefixIndexedFields))
		for i, dx := range *dps.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if dps.SuffixIndexedFields != nil {
		idx := make([]string, len(*dps.SuffixIndexedFields))
		for i, dx := range *dps.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
