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

type StatsOpts struct {
	StatIDs []*utils.DynamicStringSliceOpt
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
	if jsnCfg.StatIDs != nil {
		sqOpts.StatIDs = utils.MapToDynamicStringSliceOpts(jsnCfg.StatIDs)
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
		utils.MetaStatIDsCfg: utils.DynamicStringSliceOptsToMap(st.Opts.StatIDs),
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
	if st.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(st.ThresholdSConns)
	}
	return mp
}

func (StatSCfg) SName() string            { return StatSJSON }
func (st StatSCfg) CloneSection() Section { return st.Clone() }

func (sqOpts *StatsOpts) Clone() *StatsOpts {
	var sqIDs []*utils.DynamicStringSliceOpt
	if sqOpts.StatIDs != nil {
		sqIDs = utils.CloneDynamicStringSliceOpt(sqOpts.StatIDs)
	}
	return &StatsOpts{
		StatIDs: sqIDs,
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
	return
}

type StatsOptsJson struct {
	StatIDs map[string][]string `json:"*statIDs"`
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
	Nested_fields            *bool // applies when indexed fields is not defined
	Opts                     *StatsOptsJson
}

func diffStatsOptsJsonCfg(d *StatsOptsJson, v1, v2 *StatsOpts) *StatsOptsJson {
	if d == nil {
		d = new(StatsOptsJson)
	}
	if !utils.DynamicStringSliceOptEqual(v1.StatIDs, v2.StatIDs) {
		d.StatIDs = utils.DynamicStringSliceOptsToMap(v2.StatIDs)
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
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	d.Opts = diffStatsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
