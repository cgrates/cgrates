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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

var StatsProfileIDsDftOpt = []string{}

const StatsProfileIgnoreFilters = false

type StatsOpts struct {
	ProfileIDs           []*utils.DynamicStringSliceOpt
	ProfileIgnoreFilters []*utils.DynamicBoolOpt
	RoundingDecimals     []*utils.DynamicIntOpt
	PrometheusStatIDs    []*utils.DynamicStringSliceOpt
}

// StatSCfg the stats config section
type StatSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StoreInterval          time.Duration // Dump regularly from cache into dataDB
	StoreUncompressedLimit int
	ThresholdSConns        []string
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	Opts                   *StatsOpts
}

// loadStatSCfg loads the StatS section of the configuration
func (st *StatSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnStatSCfg := new(StatServJsonCfg)
	if err = jsnCfg.GetSection(ctx, StatSJSON, jsnStatSCfg); err != nil {
		return
	}
	return st.loadFromJSONCfg(jsnStatSCfg)
}

func (sqOpts *StatsOpts) loadFromJSONCfg(jsnCfg *StatsOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		sqOpts.ProfileIDs = append(sqOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		sqOpts.ProfileIgnoreFilters = append(sqOpts.ProfileIgnoreFilters, jsnCfg.ProfileIgnoreFilters...)
	}
	if jsnCfg.RoundingDecimals != nil {
		sqOpts.RoundingDecimals = append(sqOpts.RoundingDecimals, jsnCfg.RoundingDecimals...)
	}
	if jsnCfg.PrometheusStatIDs != nil {
		sqOpts.PrometheusStatIDs = append(sqOpts.PrometheusStatIDs, jsnCfg.PrometheusStatIDs...)
	}
}

func (st *StatSCfg) loadFromJSONCfg(jsnCfg *StatServJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		st.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		st.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Store_interval != nil {
		if st.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.Store_uncompressed_limit != nil {
		st.StoreUncompressedLimit = *jsnCfg.Store_uncompressed_limit
	}
	if jsnCfg.Thresholds_conns != nil {
		st.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.String_indexed_fields != nil {
		st.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice((*jsnCfg.String_indexed_fields)))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		st.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice((*jsnCfg.Prefix_indexed_fields)))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		st.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice((*jsnCfg.Suffix_indexed_fields)))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		st.ExistsIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		st.NotExistsIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		st.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		st.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (st StatSCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.MetaProfileIDs:           st.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFilters: st.Opts.ProfileIgnoreFilters,
		utils.OptsRoundingDecimals:     st.Opts.RoundingDecimals,
		utils.OptsPrometheusStatIDs:    st.Opts.PrometheusStatIDs,
	}
	mp := map[string]interface{}{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.NestedFieldsCfg:           st.NestedFields,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.OptsCfg:                   opts,
	}
	if st.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = st.StoreInterval.String()
	}
	if st.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*st.StringIndexedFields)
	}
	if st.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*st.PrefixIndexedFields)
	}
	if st.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*st.SuffixIndexedFields)
	}
	if st.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = utils.CloneStringSlice(*st.ExistsIndexedFields)
	}
	if st.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = utils.CloneStringSlice(*st.NotExistsIndexedFields)
	}
	if st.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(st.ThresholdSConns)
	}
	return mp
}

func (StatSCfg) SName() string            { return StatSJSON }
func (st StatSCfg) CloneSection() Section { return st.Clone() }

func (sqOpts *StatsOpts) Clone() *StatsOpts {
	var sqIDs []*utils.DynamicStringSliceOpt
	if sqOpts.ProfileIDs != nil {
		sqIDs = utils.CloneDynamicStringSliceOpt(sqOpts.ProfileIDs)
	}
	var profileIgnoreFilters []*utils.DynamicBoolOpt
	if sqOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = utils.CloneDynamicBoolOpt(sqOpts.ProfileIgnoreFilters)
	}
	var rounding []*utils.DynamicIntOpt
	if sqOpts.RoundingDecimals != nil {
		rounding = utils.CloneDynamicIntOpt(sqOpts.RoundingDecimals)
	}
	var promMtrcs []*utils.DynamicStringSliceOpt
	if sqOpts.PrometheusStatIDs != nil {
		promMtrcs = utils.CloneDynamicStringSliceOpt(sqOpts.PrometheusStatIDs)
	}
	return &StatsOpts{
		ProfileIDs:           sqIDs,
		ProfileIgnoreFilters: profileIgnoreFilters,
		RoundingDecimals:     rounding,
		PrometheusStatIDs:    promMtrcs,
	}
}

// Clone returns a deep copy of StatSCfg
func (st StatSCfg) Clone() (cln *StatSCfg) {
	cln = &StatSCfg{
		Enabled:                st.Enabled,
		IndexedSelects:         st.IndexedSelects,
		StoreInterval:          st.StoreInterval,
		StoreUncompressedLimit: st.StoreUncompressedLimit,
		NestedFields:           st.NestedFields,
		Opts:                   st.Opts.Clone(),
	}
	if st.ThresholdSConns != nil {
		cln.ThresholdSConns = utils.CloneStringSlice(st.ThresholdSConns)
	}
	if st.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*st.StringIndexedFields))
	}
	if st.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*st.PrefixIndexedFields))
	}
	if st.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*st.SuffixIndexedFields))
	}
	if st.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*st.ExistsIndexedFields))
	}
	if st.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*st.NotExistsIndexedFields))
	}
	return
}

type StatsOptsJson struct {
	ProfileIDs           []*utils.DynamicStringSliceOpt `json:"*profileIDs"`
	ProfileIgnoreFilters []*utils.DynamicBoolOpt        `json:"*profileIgnoreFilters"`
	RoundingDecimals     []*utils.DynamicIntOpt         `json:"*roundingDecimals"`
	PrometheusStatIDs    []*utils.DynamicStringSliceOpt `json:"*prometheusStatIDs"`
}

// Stat service config section
type StatServJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	Store_interval           *string
	Store_uncompressed_limit *int
	Thresholds_conns         *[]string
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Opts                     *StatsOptsJson
}

func diffStatsOptsJsonCfg(d *StatsOptsJson, v1, v2 *StatsOpts) *StatsOptsJson {
	if d == nil {
		d = new(StatsOptsJson)
	}
	if !utils.DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !utils.DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = v2.ProfileIgnoreFilters
	}
	if !utils.DynamicIntOptEqual(v1.RoundingDecimals, v2.RoundingDecimals) {
		d.RoundingDecimals = v2.RoundingDecimals
	}
	if !utils.DynamicStringSliceOptEqual(v1.PrometheusStatIDs, v2.PrometheusStatIDs) {
		d.PrometheusStatIDs = v2.PrometheusStatIDs
	}
	return d
}

func diffStatServJsonCfg(d *StatServJsonCfg, v1, v2 *StatSCfg) *StatServJsonCfg {
	if d == nil {
		d = new(StatServJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	if v1.StoreUncompressedLimit != v2.StoreUncompressedLimit {
		d.Store_uncompressed_limit = utils.IntPointer(v2.StoreUncompressedLimit)
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	d.Exists_indexed_fields = diffIndexSlice(d.Exists_indexed_fields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.Notexists_indexed_fields = diffIndexSlice(d.Notexists_indexed_fields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	d.Opts = diffStatsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
