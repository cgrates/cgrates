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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	AttributeSConns     []string
	NestedFields        bool
}

func (dps *DispatcherSCfg) loadFromJsonCfg(jsnCfg *DispatcherSJsonCfg) (err error) {
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
	if jsnCfg.Attributes_conns != nil {
		dps.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, connID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				dps.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				dps.AttributeSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Nested_fields != nil {
		dps.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (dps *DispatcherSCfg) AsMapInterface() map[string]interface{} {
	stringIndexedFields := []string{}
	if dps.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*dps.StringIndexedFields))
		for i, item := range *dps.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if dps.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*dps.PrefixIndexedFields))
		for i, item := range *dps.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	attributeSConns := make([]string, len(dps.AttributeSConns))
	for i, item := range dps.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:             dps.Enabled,
		utils.IndexedSelectsCfg:      dps.IndexedSelects,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.AttributeSConnsCfg:     attributeSConns,
		utils.NestedFieldsCfg:        dps.NestedFields,
	}

}
