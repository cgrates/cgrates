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
	RatesProfileIDsDftOpt    []string = []string{}
	RatesUsageDftOpt                  = decimal.New(int64(time.Minute), 0)
	RatesIntervalStartDftOpt          = decimal.New(0, 0)
)

const (
	RatesStartTimeDftOpt            = "*now"
	RatesProfileIgnoreFiltersDftOpt = false
)

type RatesOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	StartTime            []*DynamicStringOpt
	Usage                []*DynamicDecimalOpt
	IntervalStart        []*DynamicDecimalOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
}

// RateSCfg the rates config section
type RateSCfg struct {
	Enabled                    bool
	IndexedSelects             bool
	StringIndexedFields        *[]string
	PrefixIndexedFields        *[]string
	SuffixIndexedFields        *[]string
	ExistsIndexedFields        *[]string
	NotExistsIndexedFields     *[]string
	NestedFields               bool
	RateIndexedSelects         bool
	RateStringIndexedFields    *[]string
	RatePrefixIndexedFields    *[]string
	RateSuffixIndexedFields    *[]string
	RateExistsIndexedFields    *[]string
	RateNotExistsIndexedFields *[]string
	RateNestedFields           bool
	Verbosity                  int
	Opts                       *RatesOpts
}

// loadRateSCfg loads the rates section of the configuration
func (rCfg *RateSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRateCfg := new(RateSJsonCfg)
	if err = jsnCfg.GetSection(ctx, RateSJSON, jsnRateCfg); err != nil {
		return
	}
	return rCfg.loadFromJSONCfg(jsnRateCfg)
}

func (rateOpts *RatesOpts) loadFromJSONCfg(jsnCfg *RatesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		rateOpts.ProfileIDs = append(rateOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.StartTime != nil {
		var startime []*DynamicStringOpt
		startime, err = InterfaceToDynamicStringOpts(jsnCfg.StartTime)
		if err != nil {
			return
		}
		rateOpts.StartTime = append(startime, rateOpts.StartTime...)
	}
	if jsnCfg.Usage != nil {
		var usage []*DynamicDecimalOpt
		if usage, err = IfaceToDecimalBigDynamicOpts(jsnCfg.Usage); err != nil {
			return
		}
		rateOpts.Usage = append(usage, rateOpts.Usage...)
	}
	if jsnCfg.IntervalStart != nil {
		var intervalStart []*DynamicDecimalOpt
		intervalStart, err = IfaceToDecimalBigDynamicOpts(jsnCfg.IntervalStart)
		if err != nil {
			return
		}
		rateOpts.IntervalStart = append(intervalStart, rateOpts.IntervalStart...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var profileIgnFltr []*DynamicBoolOpt
		profileIgnFltr, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		rateOpts.ProfileIgnoreFilters = append(profileIgnFltr, rateOpts.ProfileIgnoreFilters...)
	}
	return
}

func (rCfg *RateSCfg) loadFromJSONCfg(jsnCfg *RateSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		rCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		rCfg.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		rCfg.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		rCfg.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		rCfg.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		rCfg.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		rCfg.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		rCfg.NestedFields = *jsnCfg.Nested_fields
	}

	if jsnCfg.Rate_indexed_selects != nil {
		rCfg.RateIndexedSelects = *jsnCfg.Rate_indexed_selects
	}
	if jsnCfg.Rate_string_indexed_fields != nil {
		rCfg.RateStringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Rate_string_indexed_fields))
	}
	if jsnCfg.Rate_prefix_indexed_fields != nil {
		rCfg.RatePrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Rate_prefix_indexed_fields))
	}
	if jsnCfg.Rate_suffix_indexed_fields != nil {
		rCfg.RateSuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Rate_suffix_indexed_fields))
	}
	if jsnCfg.Rate_exists_indexed_fields != nil {
		rCfg.RateExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Rate_exists_indexed_fields))
	}
	if jsnCfg.Rate_notexists_indexed_fields != nil {
		rCfg.RateNotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Rate_notexists_indexed_fields))
	}
	if jsnCfg.Rate_nested_fields != nil {
		rCfg.RateNestedFields = *jsnCfg.Rate_nested_fields
	}
	if jsnCfg.Verbosity != nil {
		rCfg.Verbosity = *jsnCfg.Verbosity
	}
	if jsnCfg.Opts != nil {
		err = rCfg.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (rCfg RateSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaProfileIDs:           rCfg.Opts.ProfileIDs,
		utils.MetaStartTime:            rCfg.Opts.StartTime,
		utils.MetaUsage:                rCfg.Opts.Usage,
		utils.MetaIntervalStartCfg:     rCfg.Opts.IntervalStart,
		utils.MetaProfileIgnoreFilters: rCfg.Opts.ProfileIgnoreFilters,
	}
	mp := map[string]any{
		utils.EnabledCfg:            rCfg.Enabled,
		utils.IndexedSelectsCfg:     rCfg.IndexedSelects,
		utils.NestedFieldsCfg:       rCfg.NestedFields,
		utils.RateIndexedSelectsCfg: rCfg.RateIndexedSelects,
		utils.RateNestedFieldsCfg:   rCfg.RateNestedFields,
		utils.Verbosity:             rCfg.Verbosity,
		utils.OptsCfg:               opts,
	}
	if rCfg.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*rCfg.StringIndexedFields)
	}
	if rCfg.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*rCfg.PrefixIndexedFields)
	}
	if rCfg.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*rCfg.SuffixIndexedFields)
	}
	if rCfg.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*rCfg.ExistsIndexedFields)
	}
	if rCfg.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*rCfg.NotExistsIndexedFields)
	}
	if rCfg.RateStringIndexedFields != nil {
		mp[utils.RateStringIndexedFieldsCfg] = slices.Clone(*rCfg.RateStringIndexedFields)
	}
	if rCfg.RatePrefixIndexedFields != nil {
		mp[utils.RatePrefixIndexedFieldsCfg] = slices.Clone(*rCfg.RatePrefixIndexedFields)
	}
	if rCfg.RateSuffixIndexedFields != nil {
		mp[utils.RateSuffixIndexedFieldsCfg] = slices.Clone(*rCfg.RateSuffixIndexedFields)
	}
	if rCfg.RateExistsIndexedFields != nil {
		mp[utils.RateExistsIndexedFieldsCfg] = slices.Clone(*rCfg.RateExistsIndexedFields)
	}
	if rCfg.RateNotExistsIndexedFields != nil {
		mp[utils.RateNotExistsIndexedFieldsCfg] = slices.Clone(*rCfg.RateNotExistsIndexedFields)
	}
	return mp
}

func (RateSCfg) SName() string              { return RateSJSON }
func (rCfg RateSCfg) CloneSection() Section { return rCfg.Clone() }

func (rateOpts *RatesOpts) Clone() *RatesOpts {
	var ratePrfIDs []*DynamicStringSliceOpt
	if rateOpts.ProfileIDs != nil {
		ratePrfIDs = CloneDynamicStringSliceOpt(rateOpts.ProfileIDs)
	}
	var startTime []*DynamicStringOpt
	if rateOpts.StartTime != nil {
		startTime = CloneDynamicStringOpt(rateOpts.StartTime)

	}
	var usage []*DynamicDecimalOpt
	if rateOpts.Usage != nil {
		usage = CloneDynamicDecimalOpt(rateOpts.Usage)
	}
	var intervalStart []*DynamicDecimalOpt
	if rateOpts.IntervalStart != nil {
		intervalStart = CloneDynamicDecimalOpt(rateOpts.IntervalStart)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if rateOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(rateOpts.ProfileIgnoreFilters)
	}
	return &RatesOpts{
		ProfileIDs:           ratePrfIDs,
		StartTime:            startTime,
		Usage:                usage,
		IntervalStart:        intervalStart,
		ProfileIgnoreFilters: profileIgnoreFilters,
	}
}

// Clone returns a deep copy of RateSCfg
func (rCfg RateSCfg) Clone() (cln *RateSCfg) {
	cln = &RateSCfg{
		Enabled:            rCfg.Enabled,
		IndexedSelects:     rCfg.IndexedSelects,
		NestedFields:       rCfg.NestedFields,
		RateIndexedSelects: rCfg.RateIndexedSelects,
		RateNestedFields:   rCfg.RateNestedFields,
		Verbosity:          rCfg.Verbosity,
		Opts:               rCfg.Opts.Clone(),
	}
	if rCfg.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.StringIndexedFields))
	}
	if rCfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.PrefixIndexedFields))
	}
	if rCfg.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.SuffixIndexedFields))
	}
	if rCfg.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.ExistsIndexedFields))
	}
	if rCfg.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.NotExistsIndexedFields))
	}
	if rCfg.RateStringIndexedFields != nil {
		cln.RateStringIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.RateStringIndexedFields))
	}
	if rCfg.RatePrefixIndexedFields != nil {
		cln.RatePrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.RatePrefixIndexedFields))
	}
	if rCfg.RateSuffixIndexedFields != nil {
		cln.RateSuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.RateSuffixIndexedFields))
	}
	if rCfg.RateExistsIndexedFields != nil {
		cln.RateExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.RateExistsIndexedFields))
	}
	if rCfg.RateNotExistsIndexedFields != nil {
		cln.RateNotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*rCfg.RateNotExistsIndexedFields))
	}
	return
}

type RatesOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	StartTime            []*DynamicInterfaceOpt   `json:"*startTime"`
	Usage                []*DynamicInterfaceOpt   `json:"*usage"`
	IntervalStart        []*DynamicInterfaceOpt   `json:"*intervalStart"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
}

type RateSJsonCfg struct {
	Enabled                       *bool
	Indexed_selects               *bool
	String_indexed_fields         *[]string
	Prefix_indexed_fields         *[]string
	Suffix_indexed_fields         *[]string
	Exists_indexed_fields         *[]string
	Notexists_indexed_fields      *[]string
	Nested_fields                 *bool // applies when indexed fields is not defined
	Rate_indexed_selects          *bool
	Rate_string_indexed_fields    *[]string
	Rate_prefix_indexed_fields    *[]string
	Rate_suffix_indexed_fields    *[]string
	Rate_exists_indexed_fields    *[]string
	Rate_notexists_indexed_fields *[]string
	Rate_nested_fields            *bool // applies when indexed fields is not defined
	Verbosity                     *int
	Opts                          *RatesOptsJson
}

func diffRatesOptsJsonCfg(d *RatesOptsJson, v1, v2 *RatesOpts) *RatesOptsJson {
	if d == nil {
		d = new(RatesOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicStringOptEqual(v1.StartTime, v2.StartTime) {
		d.StartTime = DynamicStringToInterfaceOpts(v2.StartTime)
	}
	if !DynamicDecimalOptEqual(v1.Usage, v2.Usage) {
		d.Usage = DecimalToIfaceDynamicOpts(v2.Usage)
	}
	if !DynamicDecimalOptEqual(v1.IntervalStart, v2.IntervalStart) {
		d.IntervalStart = DecimalToIfaceDynamicOpts(v2.IntervalStart)
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	return d
}

func diffRateSJsonCfg(d *RateSJsonCfg, v1, v2 *RateSCfg) *RateSJsonCfg {
	if d == nil {
		d = new(RateSJsonCfg)
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
	if v1.RateIndexedSelects != v2.RateIndexedSelects {
		d.Rate_indexed_selects = utils.BoolPointer(v2.RateIndexedSelects)
	}
	d.Rate_string_indexed_fields = diffIndexSlice(d.Rate_string_indexed_fields, v1.RateStringIndexedFields, v2.RateStringIndexedFields)
	d.Rate_prefix_indexed_fields = diffIndexSlice(d.Rate_prefix_indexed_fields, v1.RatePrefixIndexedFields, v2.RatePrefixIndexedFields)
	d.Rate_suffix_indexed_fields = diffIndexSlice(d.Rate_suffix_indexed_fields, v1.RateSuffixIndexedFields, v2.RateSuffixIndexedFields)
	d.Rate_exists_indexed_fields = diffIndexSlice(d.Rate_exists_indexed_fields, v1.RateExistsIndexedFields, v2.RateExistsIndexedFields)
	d.Rate_notexists_indexed_fields = diffIndexSlice(d.Rate_notexists_indexed_fields, v1.RateNotExistsIndexedFields, v2.RateNotExistsIndexedFields)
	if v1.RateNestedFields != v2.RateNestedFields {
		d.Rate_nested_fields = utils.BoolPointer(v2.RateNestedFields)
	}
	if v1.Verbosity != v2.Verbosity {
		d.Verbosity = utils.IntPointer(v2.Verbosity)
	}
	d.Opts = diffRatesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
