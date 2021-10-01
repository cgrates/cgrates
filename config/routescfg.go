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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type RoutesOpts struct {
	Context      []*utils.DynamicStringOpt
	IgnoreErrors []*utils.DynamicBoolOpt
	MaxCost      []*utils.DynamicInterfaceOpt
	Limit        []*utils.DynamicIntOpt
	Offset       []*utils.DynamicIntOpt
	ProfileCount []*utils.DynamicIntOpt
}

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
	Opts                *RoutesOpts
}

func (rtsOpts *RoutesOpts) loadFromJSONCfg(jsnCfg *RoutesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}

	rtsOpts.Context = utils.MapToDynamicStringOpts(jsnCfg.Context)
	rtsOpts.IgnoreErrors = utils.MapToDynamicBoolOpts(jsnCfg.IgnoreErrors)
	rtsOpts.MaxCost = utils.MapToDynamicInterfaceOpts(jsnCfg.MaxCost)
	rtsOpts.Limit = utils.MapToDynamicIntOpts(jsnCfg.Limit)
	rtsOpts.Offset = utils.MapToDynamicIntOpts(jsnCfg.Offset)
	rtsOpts.ProfileCount = utils.MapToDynamicIntOpts(jsnCfg.ProfileCount)

	return
}

// loadRouteSCfg loads the RouteS section of the configuration
func (rts *RouteSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRouteSCfg := new(RouteSJsonCfg)
	if err = jsnCfg.GetSection(ctx, RouteSJSON, jsnRouteSCfg); err != nil {
		return
	}
	return rts.loadFromJSONCfg(jsnRouteSCfg)
}

func (rts *RouteSCfg) loadFromJSONCfg(jsnCfg *RouteSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
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
	if jsnCfg.Opts != nil {
		rts.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	if jsnCfg.Nested_fields != nil {
		rts.NestedFields = *jsnCfg.Nested_fields
	}
	return
}
func (rts RoutesOpts) Clone() (cln *RoutesOpts) {
	var context []*utils.DynamicStringOpt
	if rts.Context != nil {
		context = utils.CloneDynamicStringOpt(rts.Context)
	}
	var ignoreErrors []*utils.DynamicBoolOpt
	if rts.IgnoreErrors != nil {
		ignoreErrors = utils.CloneDynamicBoolOpt(rts.IgnoreErrors)
	}
	var maxCost []*utils.DynamicInterfaceOpt
	if rts.MaxCost != nil {
		maxCost = utils.CloneDynamicInterfaceOpt(rts.MaxCost)
	}
	var profileCount []*utils.DynamicIntOpt
	if rts.ProfileCount != nil {
		profileCount = utils.CloneDynamicIntOpt(rts.ProfileCount)
	}
	var limit []*utils.DynamicIntOpt
	if rts.Limit != nil {
		limit = utils.CloneDynamicIntOpt(rts.Limit)
	}
	var offset []*utils.DynamicIntOpt
	if rts.Offset != nil {
		offset = utils.CloneDynamicIntOpt(rts.Offset)
	}
	cln = &RoutesOpts{
		Context:      context,
		IgnoreErrors: ignoreErrors,
		MaxCost:      maxCost,
		Limit:        limit,
		Offset:       offset,
		ProfileCount: profileCount,
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rts RouteSCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.OptsContext:         utils.DynamicStringOptsToMap(rts.Opts.Context),
		utils.MetaProfileCountCfg: utils.DynamicIntOptsToMap(rts.Opts.ProfileCount),
		utils.MetaIgnoreErrorsCfg: utils.DynamicBoolOptsToMap(rts.Opts.IgnoreErrors),
		utils.MetaMaxCostCfg:      utils.DynamicInterfaceOptsToMap(rts.Opts.MaxCost),
		utils.MetaLimitCfg:        utils.DynamicIntOptsToMap(rts.Opts.Limit),
		utils.MetaOffsetCfg:       utils.DynamicIntOptsToMap(rts.Opts.Offset),
	}

	mp := map[string]interface{}{
		utils.EnabledCfg:        rts.Enabled,
		utils.IndexedSelectsCfg: rts.IndexedSelects,
		utils.DefaultRatioCfg:   rts.DefaultRatio,
		utils.NestedFieldsCfg:   rts.NestedFields,
		utils.OptsCfg:           opts,
	}
	if rts.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*rts.StringIndexedFields)
	}
	if rts.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rts.PrefixIndexedFields)
	}
	if rts.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rts.SuffixIndexedFields)
	}
	if rts.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(rts.AttributeSConns)
	}
	if rts.ResourceSConns != nil {
		mp[utils.ResourceSConnsCfg] = getInternalJSONConns(rts.ResourceSConns)
	}
	if rts.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(rts.StatSConns)
	}
	if rts.RateSConns != nil {
		mp[utils.RateSConnsCfg] = getInternalJSONConns(rts.RateSConns)
	}
	if rts.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(rts.AccountSConns)
	}
	return mp
}

func (RouteSCfg) SName() string             { return RouteSJSON }
func (rts RouteSCfg) CloneSection() Section { return rts.Clone() }

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

type RoutesOptsJson struct {
	Context      map[string]string      `json:"*context"`
	IgnoreErrors map[string]bool        `json:"*ignoreErrors"`
	MaxCost      map[string]interface{} `json:"*maxCost"`
	Limit        map[string]int         `json:"*limit"`
	Offset       map[string]int         `json:"*offset"`
	ProfileCount map[string]int         `json:"*profileCount"`
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
	Opts                  *RoutesOptsJson
}

func diffRoutesOptsJsonCfg(d *RoutesOptsJson, v1, v2 *RoutesOpts) *RoutesOptsJson {
	if d == nil {
		d = new(RoutesOptsJson)
	}
	if !utils.DynamicStringOptEqual(v1.Context, v2.Context) {
		d.Context = utils.DynamicStringOptsToMap(v2.Context)
	}
	if !utils.DynamicIntOptEqual(v1.Limit, v2.Limit) {
		d.Limit = utils.DynamicIntOptsToMap(v2.Limit)
	}
	if !utils.DynamicIntOptEqual(v1.Offset, v2.Offset) {
		d.Offset = utils.DynamicIntOptsToMap(v2.Offset)
	}
	if !utils.DynamicInterfaceOptEqual(v1.MaxCost, v2.MaxCost) {
		d.MaxCost = utils.DynamicInterfaceOptsToMap(v2.MaxCost)
	}
	if !utils.DynamicBoolOptEqual(v1.IgnoreErrors, v2.IgnoreErrors) {
		d.IgnoreErrors = utils.DynamicBoolOptsToMap(v2.IgnoreErrors)
	}
	if !utils.DynamicIntOptEqual(v1.ProfileCount, v2.ProfileCount) {
		d.ProfileCount = utils.DynamicIntOptsToMap(v2.ProfileCount)
	}
	return d
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
	d.Opts = diffRoutesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
