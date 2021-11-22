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

type RoutesOpts struct {
	Context      string
	IgnoreErrors bool
	MaxCost      interface{}
	Limit        *int
	Offset       *int
	ProfileCount *int
}

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
	Opts                *RoutesOpts
}

func (rtsOpts *RoutesOpts) loadFromJSONCfg(jsnCfg *RoutesOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Context != nil {
		rtsOpts.Context = *jsnCfg.Context
	}
	if jsnCfg.IgnoreErrors != nil {
		rtsOpts.IgnoreErrors = *jsnCfg.IgnoreErrors
	}
	if jsnCfg.MaxCost != nil {
		rtsOpts.MaxCost = jsnCfg.MaxCost
	}
	if jsnCfg.Limit != nil {
		rtsOpts.Limit = jsnCfg.Limit
	}
	if jsnCfg.Offset != nil {
		rtsOpts.Offset = jsnCfg.Offset
	}
	if jsnCfg.ProfileCount != nil {
		rtsOpts.ProfileCount = jsnCfg.ProfileCount
	}
}

func (rts *RouteSCfg) loadFromJSONCfg(jsnCfg *RouteSJsonCfg) (err error) {
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
			rts.AttributeSConns[idx] = conn
			if conn == utils.MetaInternal {
				rts.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		rts.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, conn := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			rts.ResourceSConns[idx] = conn
			if conn == utils.MetaInternal {
				rts.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		rts.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			rts.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				rts.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Rals_conns != nil {
		rts.RALsConns = make([]string, len(*jsnCfg.Rals_conns))
		for idx, conn := range *jsnCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			rts.RALsConns[idx] = conn
			if conn == utils.MetaInternal {
				rts.RALsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			}
		}
	}
	if jsnCfg.Default_ratio != nil {
		rts.DefaultRatio = *jsnCfg.Default_ratio
	}
	if jsnCfg.Nested_fields != nil {
		rts.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		rts.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (rts *RouteSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.OptsContext:         rts.Opts.Context,
		utils.MetaIgnoreErrorsCfg: rts.Opts.IgnoreErrors,
	}
	if rts.Opts.MaxCost != nil {
		opts[utils.MetaMaxCostCfg] = rts.Opts.MaxCost
	}
	if rts.Opts.Limit != nil {
		opts[utils.MetaLimitCfg] = *rts.Opts.Limit
	}
	if rts.Opts.Offset != nil {
		opts[utils.MetaOffsetCfg] = *rts.Opts.Offset
	}
	if rts.Opts.ProfileCount != nil {
		opts[utils.MetaProfileCountCfg] = rts.Opts.ProfileCount
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        rts.Enabled,
		utils.IndexedSelectsCfg: rts.IndexedSelects,
		utils.DefaultRatioCfg:   rts.DefaultRatio,
		utils.NestedFieldsCfg:   rts.NestedFields,
		utils.OptsCfg:           opts,
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
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if rts.RALsConns != nil {
		ralSConns := make([]string, len(rts.RALsConns))
		for i, item := range rts.RALsConns {
			ralSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder) {
				ralSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.RALsConnsCfg] = ralSConns
	}
	if rts.ResourceSConns != nil {
		resourceSConns := make([]string, len(rts.ResourceSConns))
		for i, item := range rts.ResourceSConns {
			resourceSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resourceSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resourceSConns
	}
	if rts.StatSConns != nil {
		statSConns := make([]string, len(rts.StatSConns))
		for i, item := range rts.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	return
}

func (rts *RoutesOpts) Clone() (cln *RoutesOpts) {
	cln = &RoutesOpts{
		Context:      rts.Context,
		IgnoreErrors: rts.IgnoreErrors,
		MaxCost:      rts.MaxCost,
	}
	if rts.ProfileCount != nil {
		cln.ProfileCount = utils.IntPointer(*rts.ProfileCount)
	}
	if rts.Limit != nil {
		cln.Limit = utils.IntPointer(*rts.Limit)
	}
	if rts.Offset != nil {
		cln.Offset = utils.IntPointer(*rts.Offset)
	}
	return
}

// Clone returns a deep copy of RouteSCfg
func (rts RouteSCfg) Clone() (cln *RouteSCfg) {
	cln = &RouteSCfg{
		Enabled:        rts.Enabled,
		IndexedSelects: rts.IndexedSelects,
		DefaultRatio:   rts.DefaultRatio,
		NestedFields:   rts.NestedFields,
		Opts:           rts.Opts.Clone(),
	}
	if rts.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(rts.AttributeSConns))
		for i, con := range rts.AttributeSConns {
			cln.AttributeSConns[i] = con
		}
	}
	if rts.ResourceSConns != nil {
		cln.ResourceSConns = make([]string, len(rts.ResourceSConns))
		for i, con := range rts.ResourceSConns {
			cln.ResourceSConns[i] = con
		}
	}
	if rts.StatSConns != nil {
		cln.StatSConns = make([]string, len(rts.StatSConns))
		for i, con := range rts.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if rts.RALsConns != nil {
		cln.RALsConns = make([]string, len(rts.RALsConns))
		for i, con := range rts.RALsConns {
			cln.RALsConns[i] = con
		}
	}
	if rts.StringIndexedFields != nil {
		idx := make([]string, len(*rts.StringIndexedFields))
		for i, dx := range *rts.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if rts.PrefixIndexedFields != nil {
		idx := make([]string, len(*rts.PrefixIndexedFields))
		for i, dx := range *rts.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if rts.SuffixIndexedFields != nil {
		idx := make([]string, len(*rts.SuffixIndexedFields))
		for i, dx := range *rts.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
