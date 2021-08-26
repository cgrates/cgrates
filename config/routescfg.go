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

// RouteSCfg is the configuration of route service
type RouteSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	AttributeSConns     []string
	ResourceSConns      []string
	StatSConns          []string
	RateSConns          []string
	AccountSConns       []string
	DefaultRatio        int
	DefaultOpts         map[string]interface{}
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
		rts.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		rts.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		rts.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Attributes_conns != nil {
		rts.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.Resources_conns != nil {
		rts.ResourceSConns = updateInternalConns(*jsnCfg.Resources_conns, utils.MetaResources)
	}
	if jsnCfg.Stats_conns != nil {
		rts.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Rates_conns != nil {
		rts.RateSConns = updateInternalConns(*jsnCfg.Rates_conns, utils.MetaRateS)
	}
	if jsnCfg.Accounts_conns != nil {
		rts.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Default_ratio != nil {
		rts.DefaultRatio = *jsnCfg.Default_ratio
	}
	if jsnCfg.Default_opts != nil {
		rts.DefaultOpts = jsnCfg.Default_opts
	}
	if jsnCfg.Nested_fields != nil {
		rts.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (rts *RouteSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        rts.Enabled,
		utils.IndexedSelectsCfg: rts.IndexedSelects,
		utils.DefaultRatioCfg:   rts.DefaultRatio,
		utils.NestedFieldsCfg:   rts.NestedFields,
		utils.DefaultOptsCfg:    rts.DefaultOpts,
	}
	if rts.StringIndexedFields != nil {
		initialMP[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*rts.StringIndexedFields)
	}
	if rts.PrefixIndexedFields != nil {
		initialMP[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rts.PrefixIndexedFields)
	}
	if rts.SuffixIndexedFields != nil {
		initialMP[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rts.SuffixIndexedFields)
	}
	if rts.AttributeSConns != nil {
		initialMP[utils.AttributeSConnsCfg] = getInternalJSONConns(rts.AttributeSConns)
	}
	if rts.ResourceSConns != nil {
		initialMP[utils.ResourceSConnsCfg] = getInternalJSONConns(rts.ResourceSConns)
	}
	if rts.StatSConns != nil {
		initialMP[utils.StatSConnsCfg] = getInternalJSONConns(rts.StatSConns)
	}
	if rts.RateSConns != nil {
		initialMP[utils.RateSConnsCfg] = getInternalJSONConns(rts.RateSConns)
	}
	if rts.AccountSConns != nil {
		initialMP[utils.AccountSConnsCfg] = getInternalJSONConns(rts.AccountSConns)
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
		DefaultOpts:    rts.DefaultOpts,
	}
	if rts.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(rts.AttributeSConns)
	}
	if rts.ResourceSConns != nil {
		cln.ResourceSConns = utils.CloneStringSlice(rts.ResourceSConns)
	}
	if rts.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(rts.StatSConns)
	}
	if rts.RateSConns != nil {
		cln.RateSConns = utils.CloneStringSlice(rts.RateSConns)
	}
	if rts.AccountSConns != nil {
		cln.AccountSConns = utils.CloneStringSlice(rts.AccountSConns)
	}
	if rts.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rts.StringIndexedFields))
	}
	if rts.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rts.PrefixIndexedFields))
	}
	if rts.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rts.SuffixIndexedFields))
	}
	return
}

// Route service config section
type RouteSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
	Resources_conns       *[]string
	Stats_conns           *[]string
	Rates_conns           *[]string
	Accounts_conns        *[]string
	Default_ratio         *int
	Default_opts          map[string]interface{}
}

func diffRouteSJsonCfg(d *RouteSJsonCfg, v1, v2 *RouteSCfg) *RouteSJsonCfg {
	if d == nil {
		d = new(RouteSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !utils.SliceStringEqual(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !utils.SliceStringEqual(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	if v1.DefaultRatio != v2.DefaultRatio {
		d.Default_ratio = utils.IntPointer(v2.DefaultRatio)
	}
	d.Default_opts = diffMap(d.Default_opts, v1.DefaultOpts, v2.DefaultOpts)
	return d
}
