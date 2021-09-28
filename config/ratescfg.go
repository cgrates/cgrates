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
	"github.com/ericlagergren/decimal"
)

type RatesOpts struct {
	RateProfileIDs map[string][]string
	StartTime      string
	Usage          *decimal.Big
	IntervalStart  *decimal.Big
}

// RateSCfg the rates config section
type RateSCfg struct {
	Enabled                 bool
	IndexedSelects          bool
	StringIndexedFields     *[]string
	PrefixIndexedFields     *[]string
	SuffixIndexedFields     *[]string
	NestedFields            bool
	RateIndexedSelects      bool
	RateStringIndexedFields *[]string
	RatePrefixIndexedFields *[]string
	RateSuffixIndexedFields *[]string
	RateNestedFields        bool
	Verbosity               int
	Opts                    *RatesOpts
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
		return nil
	}
	if jsnCfg.RateProfileIDs != nil {
		rateOpts.RateProfileIDs = jsnCfg.RateProfileIDs
	}
	if jsnCfg.StartTime != nil {
		rateOpts.StartTime = *jsnCfg.StartTime
	}
	if jsnCfg.Usage != nil {
		if rateOpts.Usage, err = utils.StringAsBig(*jsnCfg.Usage); err != nil {
			return
		}
	}
	if jsnCfg.IntervalStart != nil {
		if rateOpts.IntervalStart, err = utils.StringAsBig(*jsnCfg.IntervalStart); err != nil {
			return
		}
	}

	return nil
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
		rCfg.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		rCfg.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		rCfg.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		rCfg.NestedFields = *jsnCfg.Nested_fields
	}

	if jsnCfg.Rate_indexed_selects != nil {
		rCfg.RateIndexedSelects = *jsnCfg.Rate_indexed_selects
	}
	if jsnCfg.Rate_string_indexed_fields != nil {
		rCfg.RateStringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Rate_string_indexed_fields))
	}
	if jsnCfg.Rate_prefix_indexed_fields != nil {
		rCfg.RatePrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Rate_prefix_indexed_fields))
	}
	if jsnCfg.Rate_suffix_indexed_fields != nil {
		rCfg.RateSuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Rate_suffix_indexed_fields))
	}
	if jsnCfg.Rate_nested_fields != nil {
		rCfg.RateNestedFields = *jsnCfg.Rate_nested_fields
	}
	if jsnCfg.Verbosity != nil {
		rCfg.Verbosity = *jsnCfg.Verbosity
	}
	if jsnCfg.Opts != nil {
		rCfg.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rCfg RateSCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.MetaRateProfileIDsCfg: rCfg.Opts.RateProfileIDs,
		utils.MetaStartTime:         rCfg.Opts.StartTime,
		utils.MetaUsage:             rCfg.Opts.Usage,
		utils.MetaIntervalStartCfg:  rCfg.Opts.IntervalStart,
	}
	mp := map[string]interface{}{
		utils.EnabledCfg:            rCfg.Enabled,
		utils.IndexedSelectsCfg:     rCfg.IndexedSelects,
		utils.NestedFieldsCfg:       rCfg.NestedFields,
		utils.RateIndexedSelectsCfg: rCfg.RateIndexedSelects,
		utils.RateNestedFieldsCfg:   rCfg.RateNestedFields,
		utils.Verbosity:             rCfg.Verbosity,
		utils.OptsCfg:               opts,
	}
	if rCfg.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.StringIndexedFields)
	}
	if rCfg.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.PrefixIndexedFields)
	}
	if rCfg.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.SuffixIndexedFields)
	}
	if rCfg.RateStringIndexedFields != nil {
		mp[utils.RateStringIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RateStringIndexedFields)
	}
	if rCfg.RatePrefixIndexedFields != nil {
		mp[utils.RatePrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RatePrefixIndexedFields)
	}
	if rCfg.RateSuffixIndexedFields != nil {
		mp[utils.RateSuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RateSuffixIndexedFields)
	}
	return mp
}

func (RateSCfg) SName() string              { return RateSJSON }
func (rCfg RateSCfg) CloneSection() Section { return rCfg.Clone() }

func (rateOpts *RatesOpts) Clone() *RatesOpts {
	return &RatesOpts{
		RateProfileIDs: rateOpts.RateProfileIDs,
		StartTime:      rateOpts.StartTime,
		Usage:          rateOpts.Usage,
		IntervalStart:  rateOpts.IntervalStart,
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
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.StringIndexedFields))
	}
	if rCfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.PrefixIndexedFields))
	}
	if rCfg.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.SuffixIndexedFields))
	}

	if rCfg.RateStringIndexedFields != nil {
		cln.RateStringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.RateStringIndexedFields))
	}
	if rCfg.RatePrefixIndexedFields != nil {
		cln.RatePrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.RatePrefixIndexedFields))
	}
	if rCfg.RateSuffixIndexedFields != nil {
		cln.RateSuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rCfg.RateSuffixIndexedFields))
	}
	return
}

type RatesOptsJson struct {
	RateProfileIDs map[string][]string `json:"*rateProfileIDs"`
	StartTime      *string             `json:"*startTime"`
	Usage          *string             `json:"*usage"`
	IntervalStart  *string             `json:"*intervalStart"`
}

type RateSJsonCfg struct {
	Enabled                    *bool
	Indexed_selects            *bool
	String_indexed_fields      *[]string
	Prefix_indexed_fields      *[]string
	Suffix_indexed_fields      *[]string
	Nested_fields              *bool // applies when indexed fields is not defined
	Rate_indexed_selects       *bool
	Rate_string_indexed_fields *[]string
	Rate_prefix_indexed_fields *[]string
	Rate_suffix_indexed_fields *[]string
	Rate_nested_fields         *bool // applies when indexed fields is not defined
	Verbosity                  *int
	Opts                       *RatesOptsJson
}

func diffRatesOptsJsonCfg(d *RatesOptsJson, v1, v2 *RatesOpts) *RatesOptsJson {
	if d == nil {
		d = new(RatesOptsJson)
	}
	d.RateProfileIDs = diffMapStringStringSlice(d.RateProfileIDs, v1.RateProfileIDs, v2.RateProfileIDs)
	if v1.StartTime != v2.StartTime {
		d.StartTime = utils.StringPointer(v2.StartTime)
	}
	if v1.Usage.Cmp(v2.Usage) != 0 {
		d.Usage = utils.StringPointer(v2.Usage.String())
	}
	if v1.IntervalStart.Cmp(v2.IntervalStart) != 0 {
		d.IntervalStart = utils.StringPointer(v2.IntervalStart.String())
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
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if v1.RateIndexedSelects != v2.RateIndexedSelects {
		d.Rate_indexed_selects = utils.BoolPointer(v2.RateIndexedSelects)
	}
	d.Rate_string_indexed_fields = diffIndexSlice(d.Rate_string_indexed_fields, v1.RateStringIndexedFields, v2.RateStringIndexedFields)
	d.Rate_prefix_indexed_fields = diffIndexSlice(d.Rate_prefix_indexed_fields, v1.RatePrefixIndexedFields, v2.RatePrefixIndexedFields)
	d.Rate_suffix_indexed_fields = diffIndexSlice(d.Rate_suffix_indexed_fields, v1.RateSuffixIndexedFields, v2.RateSuffixIndexedFields)
	if v1.RateNestedFields != v2.RateNestedFields {
		d.Rate_nested_fields = utils.BoolPointer(v2.RateNestedFields)
	}
	if v1.Verbosity != v2.Verbosity {
		d.Verbosity = utils.IntPointer(v2.Verbosity)
	}
	d.Opts = diffRatesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
