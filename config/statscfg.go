/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

var StatsProfileIDsDftOpt = []string{}

const StatsProfileIgnoreFilters = false

type StatsOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
	RoundingDecimals     []*DynamicIntOpt
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
	EEsConns               []string
	EEsExporterIDs         []string
}

// loadStatSCfg loads the StatS section of the configuration
func (st *StatSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnStatSCfg := new(StatServJsonCfg)
	if err = jsnCfg.GetSection(ctx, StatSJSON, jsnStatSCfg); err != nil {
		return
	}
	return st.loadFromJSONCfg(jsnStatSCfg)
}

func (sqOpts *StatsOpts) loadFromJSONCfg(jsnCfg *StatsOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		sqOpts.ProfileIDs = append(sqOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var prfIgnFltrs []*DynamicBoolOpt
		prfIgnFltrs, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		if err != nil {
			return
		}
		sqOpts.ProfileIgnoreFilters = append(prfIgnFltrs, sqOpts.ProfileIgnoreFilters...)
	}
	if jsnCfg.RoundingDecimals != nil {
		var roundDec []*DynamicIntOpt
		roundDec, err = IfaceToIntDynamicOpts(jsnCfg.RoundingDecimals)
		sqOpts.RoundingDecimals = append(roundDec, sqOpts.RoundingDecimals...)
	}
	return
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
		st.ThresholdSConns = tagInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Ees_conns != nil {
		st.EEsConns = tagInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCfg.Ees_exporter_ids != nil {
		st.EEsExporterIDs = append(st.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	if jsnCfg.String_indexed_fields != nil {
		st.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		st.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		st.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		st.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		st.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		st.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		st.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (st StatSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaProfileIDs:           st.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFilters: st.Opts.ProfileIgnoreFilters,
		utils.OptsRoundingDecimals:     st.Opts.RoundingDecimals,
	}
	mp := map[string]any{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.NestedFieldsCfg:           st.NestedFields,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.EEsExporterIDsCfg:         slices.Clone(st.EEsExporterIDs),
		utils.OptsCfg:                   opts,
	}
	if st.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = st.StoreInterval.String()
	}
	if st.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*st.StringIndexedFields)
	}
	if st.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*st.PrefixIndexedFields)
	}
	if st.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*st.SuffixIndexedFields)
	}
	if st.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*st.ExistsIndexedFields)
	}
	if st.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*st.NotExistsIndexedFields)
	}
	if st.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = stripInternalConns(st.ThresholdSConns)
	}
	if st.EEsConns != nil {
		mp[utils.EEsConnsCfg] = stripInternalConns(st.EEsConns)
	}
	return mp
}

func (StatSCfg) SName() string            { return StatSJSON }
func (st StatSCfg) CloneSection() Section { return st.Clone() }

func (sqOpts *StatsOpts) Clone() *StatsOpts {
	var sqIDs []*DynamicStringSliceOpt
	if sqOpts.ProfileIDs != nil {
		sqIDs = CloneDynamicStringSliceOpt(sqOpts.ProfileIDs)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if sqOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(sqOpts.ProfileIgnoreFilters)
	}
	var rounding []*DynamicIntOpt
	if sqOpts.RoundingDecimals != nil {
		rounding = CloneDynamicIntOpt(sqOpts.RoundingDecimals)
	}
	return &StatsOpts{
		ProfileIDs:           sqIDs,
		ProfileIgnoreFilters: profileIgnoreFilters,
		RoundingDecimals:     rounding,
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
		cln.ThresholdSConns = slices.Clone(st.ThresholdSConns)
	}
	if st.EEsConns != nil {
		cln.EEsConns = slices.Clone(st.EEsConns)
	}
	if st.EEsExporterIDs != nil {
		cln.EEsExporterIDs = slices.Clone(st.EEsExporterIDs)
	}
	if st.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*st.StringIndexedFields))
	}
	if st.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*st.PrefixIndexedFields))
	}
	if st.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*st.SuffixIndexedFields))
	}
	if st.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*st.ExistsIndexedFields))
	}
	if st.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*st.NotExistsIndexedFields))
	}
	return
}

type StatsOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
	RoundingDecimals     []*DynamicInterfaceOpt   `json:"*roundingDecimals"`
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
	Ees_conns                *[]string
	Ees_exporter_ids         *[]string
}

func diffStatsOptsJsonCfg(d *StatsOptsJson, v1, v2 *StatsOpts) *StatsOptsJson {
	if d == nil {
		d = new(StatsOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	if !DynamicIntOptEqual(v1.RoundingDecimals, v2.RoundingDecimals) {
		d.RoundingDecimals = IntToIfaceDynamicOpts(v2.RoundingDecimals)
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
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(stripInternalConns(v2.EEsConns))
	}
	if !slices.Equal(v1.EEsExporterIDs, v2.EEsExporterIDs) {
		d.Ees_exporter_ids = &v2.EEsExporterIDs
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
