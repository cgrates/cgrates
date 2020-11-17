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
	RateSConns          []string
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
	if jsnCfg.Rates_conns != nil {
		rts.RateSConns = make([]string, len(*jsnCfg.Rates_conns))
		for idx, conn := range *jsnCfg.Rates_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rts.RateSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)
			} else {
				rts.RateSConns[idx] = conn
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

func (rts *RouteSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        rts.Enabled,
		utils.IndexedSelectsCfg: rts.IndexedSelects,
		utils.DefaultRatioCfg:   rts.DefaultRatio,
		utils.NestedFieldsCfg:   rts.NestedFields,
	}
	if rts.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*rts.StringIndexedFields))
		for i, item := range *rts.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if rts.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*rts.PrefixIndexedFields))
		for i, item := range *rts.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if rts.SuffixIndexedFields != nil {
		suffixIndexedFieldsCfg := make([]string, len(*rts.SuffixIndexedFields))
		for i, item := range *rts.SuffixIndexedFields {
			suffixIndexedFieldsCfg[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFieldsCfg
	}
	if rts.AttributeSConns != nil {
		attributeSConns := make([]string, len(rts.AttributeSConns))
		for i, item := range rts.AttributeSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes)
			} else {
				attributeSConns[i] = item
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if rts.RALsConns != nil {
		ralSConns := make([]string, len(rts.RALsConns))
		for i, item := range rts.RALsConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder) {
				ralSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaResponder)
			} else {
				ralSConns[i] = item
			}
		}
		initialMP[utils.RALsConnsCfg] = ralSConns
	}
	if rts.ResourceSConns != nil {
		resourceSConns := make([]string, len(rts.ResourceSConns))
		for i, item := range rts.ResourceSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resourceSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaResources)
			} else {
				resourceSConns[i] = item
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resourceSConns
	}
	if rts.StatSConns != nil {
		statSConns := make([]string, len(rts.StatSConns))
		for i, item := range rts.StatSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS) {
				statSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaStatS)
			} else {
				statSConns[i] = item
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if rts.RateSConns != nil {
		rateSConns := make([]string, len(rts.RateSConns))
		for i, item := range rts.RateSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS) {
				rateSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaRateS)
			} else {
				rateSConns[i] = item
			}
		}
		initialMP[utils.RateSConnsCfg] = rateSConns
	}
	return
}
