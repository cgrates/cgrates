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

// SupplierSCfg is the configuration of supplier service
type SupplierSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	AttributeSConns     []string
	ResourceSConns      []string
	StatSConns          []string
	ResponderSConns     []string
	DefaultRatio        int
	NestedFields        bool
}

func (spl *SupplierSCfg) loadFromJsonCfg(jsnCfg *SupplierSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		spl.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		spl.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		spl.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		spl.PrefixIndexedFields = &pif
	}
	if jsnCfg.Attributes_conns != nil {
		spl.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				spl.AttributeSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		spl.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, conn := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			} else {
				spl.ResourceSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		spl.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			} else {
				spl.StatSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Rals_conns != nil {
		spl.ResponderSConns = make([]string, len(*jsnCfg.Rals_conns))
		for idx, conn := range *jsnCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.ResponderSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			} else {
				spl.ResponderSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Default_ratio != nil {
		spl.DefaultRatio = *jsnCfg.Default_ratio
	}
	if jsnCfg.Nested_fields != nil {
		spl.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (spl *SupplierSCfg) AsMapInterface() map[string]interface{} {
	stringIndexedFields := []string{}
	if spl.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*spl.StringIndexedFields))
		for i, item := range *spl.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if spl.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*spl.PrefixIndexedFields))
		for i, item := range *spl.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	attributeSConns := make([]string, len(spl.AttributeSConns))
	for i, item := range spl.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}
	responderSConns := make([]string, len(spl.ResponderSConns))
	for i, item := range spl.ResponderSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)

		if item == buf {
			responderSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResponder, utils.EmptyString)
		} else {
			responderSConns[i] = item
		}
	}
	resourceSConns := make([]string, len(spl.ResourceSConns))
	for i, item := range spl.ResourceSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
		if item == buf {
			resourceSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResources, utils.EmptyString)
		} else {
			resourceSConns[i] = item
		}
	}
	statSConns := make([]string, len(spl.StatSConns))
	for i, item := range spl.StatSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
		if item == buf {
			statSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaStatS, utils.EmptyString)
		} else {
			statSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:             spl.Enabled,
		utils.IndexedSelectsCfg:      spl.IndexedSelects,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.AttributeSConnsCfg:     attributeSConns,
		utils.ResourceSConnsCfg:      resourceSConns,
		utils.StatSConnsCfg:          statSConns,
		utils.RALsConnsCfg:           responderSConns,
		utils.DefaultRatioCfg:        spl.DefaultRatio,
		utils.NestedFieldsCfg:        spl.NestedFields,
	}

}
