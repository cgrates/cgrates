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
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rCfg *RateSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:            rCfg.Enabled,
		utils.IndexedSelectsCfg:     rCfg.IndexedSelects,
		utils.NestedFieldsCfg:       rCfg.NestedFields,
		utils.RateIndexedSelectsCfg: rCfg.RateIndexedSelects,
		utils.RateNestedFieldsCfg:   rCfg.RateNestedFields,
		utils.Verbosity:             rCfg.Verbosity,
	}
	if rCfg.StringIndexedFields != nil {
		initialMP[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.StringIndexedFields)
	}
	if rCfg.PrefixIndexedFields != nil {
		initialMP[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.PrefixIndexedFields)
	}
	if rCfg.SuffixIndexedFields != nil {
		initialMP[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.SuffixIndexedFields)
	}
	if rCfg.RateStringIndexedFields != nil {
		initialMP[utils.RateStringIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RateStringIndexedFields)
	}
	if rCfg.RatePrefixIndexedFields != nil {
		initialMP[utils.RatePrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RatePrefixIndexedFields)
	}
	if rCfg.RateSuffixIndexedFields != nil {
		initialMP[utils.RateSuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rCfg.RateSuffixIndexedFields)
	}
	return
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
	return d
}
