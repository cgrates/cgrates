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

// RouteSCfg is the configuration of route service
type RouteSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	AttributeSConns     []string
	ResourceSConns      []string
	StatSConns          []string
	RALsConns           []string
	DefaultRatio        int
	NestedFields        bool
}

func (rts *RouteSCfg) loadFromJsonCfg(jsnCfg *RouteSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rts.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		rts.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		rts.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		rts.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		rts.SuffixIndexedFields = &sif
	}
	if jsnCfg.Attributes_conns != nil {
		rts.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rts.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				rts.AttributeSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		rts.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, conn := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rts.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			} else {
				rts.ResourceSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		rts.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rts.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			} else {
				rts.StatSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Rals_conns != nil {
		rts.RALsConns = make([]string, len(*jsnCfg.Rals_conns))
		for idx, conn := range *jsnCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rts.RALsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			} else {
				rts.RALsConns[idx] = conn
			}
		}
	}
	if jsnCfg.Default_ratio != nil {
		rts.DefaultRatio = *jsnCfg.Default_ratio
	}
	if jsnCfg.Nested_fields != nil {
		rts.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (rts *RouteSCfg) AsMapInterface() map[string]interface{} {

	stringIndexedFields := []string{}
	if rts.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*rts.StringIndexedFields))
		for i, item := range *rts.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if rts.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*rts.PrefixIndexedFields))
		for i, item := range *rts.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	attributeSConns := make([]string, len(rts.AttributeSConns))
	for i, item := range rts.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}
	ralSConns := make([]string, len(rts.RALsConns))
	for i, item := range rts.RALsConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)

		if item == buf {
			ralSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResponder, utils.EmptyString)
		} else {
			ralSConns[i] = item
		}
	}
	resourceSConns := make([]string, len(rts.ResourceSConns))
	for i, item := range rts.ResourceSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
		if item == buf {
			resourceSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResources, utils.EmptyString)
		} else {
			resourceSConns[i] = item
		}
	}
	statSConns := make([]string, len(rts.StatSConns))
	for i, item := range rts.StatSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
		if item == buf {
			statSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaStatS, utils.EmptyString)
		} else {
			statSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:             rts.Enabled,
		utils.IndexedSelectsCfg:      rts.IndexedSelects,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.AttributeSConnsCfg:     attributeSConns,
		utils.ResourceSConnsCfg:      resourceSConns,
		utils.StatSConnsCfg:          statSConns,
		utils.RALsConnsCfg:           ralSConns,
		utils.DefaultRatioCfg:        rts.DefaultRatio,
		utils.NestedFieldsCfg:        rts.NestedFields,
	}

}
