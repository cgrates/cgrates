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
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

var (
	RoutesProfileCountDftOpt = utils.IntPointer(1)
	RoutesUsageDftOpt        = decimal.New(int64(time.Minute), 0)
)

const (
	RoutesContextDftOpt      = "*routes"
	RoutesIgnoreErrorsDftOpt = false
	RoutesMaxCostDftOpt      = utils.EmptyString
)

type RoutesOpts struct {
	Context      []*DynamicStringOpt
	IgnoreErrors []*DynamicBoolOpt
	MaxCost      []*DynamicInterfaceOpt
	Limit        []*DynamicIntPointerOpt
	Offset       []*DynamicIntPointerOpt
	MaxItems     []*DynamicIntPointerOpt
	ProfileCount []*DynamicIntPointerOpt
	Usage        []*DynamicDecimalOpt
}

// RouteSCfg is the configuration of route service
type RouteSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	AttributeSConns        []string
	ResourceSConns         []string
	StatSConns             []string
	RateSConns             []string
	AccountSConns          []string
	DefaultRatio           int
	Opts                   *RoutesOpts
}

// loadRouteSCfg loads the RouteS section of the configuration
func (rts *RouteSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRouteSCfg := new(RouteSJsonCfg)
	if err = jsnCfg.GetSection(ctx, RouteSJSON, jsnRouteSCfg); err != nil {
		return
	}
	return rts.loadFromJSONCfg(jsnRouteSCfg)
}

func (rtsOpts *RoutesOpts) loadFromJSONCfg(jsnCfg *RoutesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Context != nil {
		rtsOpts.Context = append(rtsOpts.Context, jsnCfg.Context...)
	}
	if jsnCfg.IgnoreErrors != nil {
		rtsOpts.IgnoreErrors = append(rtsOpts.IgnoreErrors, jsnCfg.IgnoreErrors...)
	}
	if jsnCfg.MaxCost != nil {
		rtsOpts.MaxCost = append(rtsOpts.MaxCost, jsnCfg.MaxCost...)
	}
	if jsnCfg.Limit != nil {
		rtsOpts.Limit = append(rtsOpts.Limit, IntToIntPointerDynamicOpts(jsnCfg.Limit)...)
	}
	if jsnCfg.Offset != nil {
		rtsOpts.Offset = append(rtsOpts.Offset, IntToIntPointerDynamicOpts(jsnCfg.Offset)...)
	}
	if jsnCfg.MaxItems != nil {
		rtsOpts.MaxItems = append(rtsOpts.MaxItems, IntToIntPointerDynamicOpts(jsnCfg.MaxItems)...)
	}
	if jsnCfg.ProfileCount != nil {
		rtsOpts.ProfileCount = append(rtsOpts.ProfileCount, IntToIntPointerDynamicOpts(jsnCfg.ProfileCount)...)
	}
	if jsnCfg.Usage != nil {
		var usage []*DynamicDecimalOpt
		if usage, err = StringToDecimalBigDynamicOpts(jsnCfg.Usage); err != nil {
			return
		}
		rtsOpts.Usage = append(rtsOpts.Usage, usage...)
	}
	return
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
		rts.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		rts.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		rts.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		rts.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		rts.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
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
		rts.RateSConns = updateInternalConns(*jsnCfg.Rates_conns, utils.MetaRates)
	}
	if jsnCfg.Accounts_conns != nil {
		rts.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Default_ratio != nil {
		rts.DefaultRatio = *jsnCfg.Default_ratio
	}
	if jsnCfg.Nested_fields != nil {
		rts.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		err = rts.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}
func (rts *RoutesOpts) Clone() (cln *RoutesOpts) {
	var context []*DynamicStringOpt
	if rts.Context != nil {
		context = CloneDynamicStringOpt(rts.Context)
	}
	var ignoreErrors []*DynamicBoolOpt
	if rts.IgnoreErrors != nil {
		ignoreErrors = CloneDynamicBoolOpt(rts.IgnoreErrors)
	}
	var maxCost []*DynamicInterfaceOpt
	if rts.MaxCost != nil {
		maxCost = CloneDynamicInterfaceOpt(rts.MaxCost)
	}
	var profileCount []*DynamicIntPointerOpt
	if rts.ProfileCount != nil {
		profileCount = CloneDynamicIntPointerOpt(rts.ProfileCount)
	}
	var limit []*DynamicIntPointerOpt
	if rts.Limit != nil {
		limit = CloneDynamicIntPointerOpt(rts.Limit)
	}
	var offset []*DynamicIntPointerOpt
	if rts.Offset != nil {
		offset = CloneDynamicIntPointerOpt(rts.Offset)
	}
	var maxItems []*DynamicIntPointerOpt
	if rts.MaxItems != nil {
		maxItems = CloneDynamicIntPointerOpt(rts.MaxItems)
	}
	var usage []*DynamicDecimalOpt
	if rts.Usage != nil {
		usage = CloneDynamicDecimalOpt(rts.Usage)
	}
	cln = &RoutesOpts{
		Context:      context,
		IgnoreErrors: ignoreErrors,
		MaxCost:      maxCost,
		Limit:        limit,
		Offset:       offset,
		MaxItems:     maxItems,
		ProfileCount: profileCount,
		Usage:        usage,
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (rts RouteSCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.OptsContext:         rts.Opts.Context,
		utils.MetaProfileCountCfg: rts.Opts.ProfileCount,
		utils.MetaIgnoreErrorsCfg: rts.Opts.IgnoreErrors,
		utils.MetaMaxCostCfg:      rts.Opts.MaxCost,
		utils.MetaLimitCfg:        rts.Opts.Limit,
		utils.MetaOffsetCfg:       rts.Opts.Offset,
		utils.MetaMaxItemsCfg:     rts.Opts.MaxItems,
		utils.MetaUsage:           rts.Opts.Usage,
	}

	mp := map[string]any{
		utils.EnabledCfg:        rts.Enabled,
		utils.IndexedSelectsCfg: rts.IndexedSelects,
		utils.DefaultRatioCfg:   rts.DefaultRatio,
		utils.NestedFieldsCfg:   rts.NestedFields,
		utils.OptsCfg:           opts,
	}
	if rts.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*rts.StringIndexedFields)
	}
	if rts.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*rts.PrefixIndexedFields)
	}
	if rts.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*rts.SuffixIndexedFields)
	}
	if rts.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*rts.ExistsIndexedFields)
	}
	if rts.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*rts.NotExistsIndexedFields)
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
		cln.AttributeSConns = slices.Clone(rts.AttributeSConns)
	}
	if rts.ResourceSConns != nil {
		cln.ResourceSConns = slices.Clone(rts.ResourceSConns)
	}
	if rts.StatSConns != nil {
		cln.StatSConns = slices.Clone(rts.StatSConns)
	}
	if rts.RateSConns != nil {
		cln.RateSConns = slices.Clone(rts.RateSConns)
	}
	if rts.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(rts.AccountSConns)
	}
	if rts.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*rts.StringIndexedFields))
	}
	if rts.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*rts.PrefixIndexedFields))
	}
	if rts.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*rts.SuffixIndexedFields))
	}
	if rts.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rts.ExistsIndexedFields))
	}
	if rts.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rts.NotExistsIndexedFields))
	}
	return
}

type RoutesOptsJson struct {
	Context      []*DynamicStringOpt    `json:"*context"`
	IgnoreErrors []*DynamicBoolOpt      `json:"*ignoreErrors"`
	MaxCost      []*DynamicInterfaceOpt `json:"*maxCost"`
	Limit        []*DynamicIntOpt       `json:"*limit"`
	Offset       []*DynamicIntOpt       `json:"*offset"`
	MaxItems     []*DynamicIntOpt       `json:"*maxItems"`
	ProfileCount []*DynamicIntOpt       `json:"*profileCount"`
	Usage        []*DynamicStringOpt    `json:"*usage"`
}

// Route service config section
type RouteSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Attributes_conns         *[]string
	Resources_conns          *[]string
	Stats_conns              *[]string
	Rates_conns              *[]string
	Accounts_conns           *[]string
	Default_ratio            *int
	Opts                     *RoutesOptsJson
}

func diffRoutesOptsJsonCfg(d *RoutesOptsJson, v1, v2 *RoutesOpts) *RoutesOptsJson {
	if d == nil {
		d = new(RoutesOptsJson)
	}
	if !DynamicStringOptEqual(v1.Context, v2.Context) {
		d.Context = v2.Context
	}
	if !DynamicIntPointerOptEqual(v1.Limit, v2.Limit) {
		d.Limit = IntPointerToIntDynamicOpts(v2.Limit)
	}
	if !DynamicIntPointerOptEqual(v1.Offset, v2.Offset) {
		d.Offset = IntPointerToIntDynamicOpts(v2.Offset)
	}
	if !DynamicIntPointerOptEqual(v1.MaxItems, v2.MaxItems) {
		d.MaxItems = IntPointerToIntDynamicOpts(v2.MaxItems)
	}
	if !DynamicInterfaceOptEqual(v1.MaxCost, v2.MaxCost) {
		d.MaxCost = v2.MaxCost
	}
	if !DynamicBoolOptEqual(v1.IgnoreErrors, v2.IgnoreErrors) {
		d.IgnoreErrors = v2.IgnoreErrors
	}
	if !DynamicIntPointerOptEqual(v1.ProfileCount, v2.ProfileCount) {
		d.ProfileCount = IntPointerToIntDynamicOpts(v2.ProfileCount)
	}
	if !DynamicDecimalOptEqual(v1.Usage, v2.Usage) {
		d.Usage = DecimalToStringDynamicOpts(v2.Usage)
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
	d.Exists_indexed_fields = diffIndexSlice(d.Exists_indexed_fields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.Notexists_indexed_fields = diffIndexSlice(d.Notexists_indexed_fields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !slices.Equal(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !slices.Equal(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	if v1.DefaultRatio != v2.DefaultRatio {
		d.Default_ratio = utils.IntPointer(v2.DefaultRatio)
	}
	d.Opts = diffRoutesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
