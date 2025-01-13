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
)

var ThresholdsProfileIDsDftOpt = []string{}

const ThresholdsProfileIgnoreFiltersDftOpt = false

type ThresholdsOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
}

// ThresholdSCfg the threshold config section
type ThresholdSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StoreInterval          time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	ActionSConns           []string // connections towards ActionS
	Opts                   *ThresholdsOpts
}

// loadThresholdSCfg loads the ThresholdS section of the configuration
func (t *ThresholdSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnThresholdSCfg := new(ThresholdSJsonCfg)
	if err = jsnCfg.GetSection(ctx, ThresholdSJSON, jsnThresholdSCfg); err != nil {
		return
	}
	return t.loadFromJSONCfg(jsnThresholdSCfg)
}

func (thdOpts *ThresholdsOpts) loadFromJSONCfg(jsnCfg *ThresholdsOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		thdOpts.ProfileIDs = append(thdOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var profileIgnFltr []*DynamicBoolOpt
		profileIgnFltr, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		if err != nil {
			return
		}
		thdOpts.ProfileIgnoreFilters = append(profileIgnFltr, thdOpts.ProfileIgnoreFilters...)
	}
	return
}

func (t *ThresholdSCfg) loadFromJSONCfg(jsnCfg *ThresholdSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		t.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		t.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Store_interval != nil {
		if t.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		t.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		t.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		t.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		t.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		t.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		t.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Actions_conns != nil {
		t.ActionSConns = updateInternalConns(*jsnCfg.Actions_conns, utils.MetaActions)
	}
	if jsnCfg.Opts != nil {
		err = t.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (t ThresholdSCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.MetaProfileIDs:           t.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFilters: t.Opts.ProfileIgnoreFilters,
	}
	mp := map[string]any{
		utils.EnabledCfg:        t.Enabled,
		utils.IndexedSelectsCfg: t.IndexedSelects,
		utils.NestedFieldsCfg:   t.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.OptsCfg:           opts,
	}
	if t.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = t.StoreInterval.String()
	}

	if t.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*t.StringIndexedFields)
	}
	if t.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*t.PrefixIndexedFields)
	}
	if t.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*t.SuffixIndexedFields)
	}
	if t.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*t.ExistsIndexedFields)
	}
	if t.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*t.NotExistsIndexedFields)
	}
	if t.ActionSConns != nil {
		mp[utils.ActionSConnsCfg] = getInternalJSONConns(t.ActionSConns)
	}
	return mp
}

func (ThresholdSCfg) SName() string           { return ThresholdSJSON }
func (t ThresholdSCfg) CloneSection() Section { return t.Clone() }

func (thdOpts *ThresholdsOpts) Clone() *ThresholdsOpts {
	var thIDs []*DynamicStringSliceOpt
	if thdOpts.ProfileIDs != nil {
		thIDs = CloneDynamicStringSliceOpt(thdOpts.ProfileIDs)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if thdOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(thdOpts.ProfileIgnoreFilters)
	}
	return &ThresholdsOpts{
		ProfileIDs:           thIDs,
		ProfileIgnoreFilters: profileIgnoreFilters,
	}
}

// Clone returns a deep copy of ThresholdSCfg
func (t ThresholdSCfg) Clone() (cln *ThresholdSCfg) {
	cln = &ThresholdSCfg{
		Enabled:        t.Enabled,
		IndexedSelects: t.IndexedSelects,
		StoreInterval:  t.StoreInterval,
		NestedFields:   t.NestedFields,
		Opts:           t.Opts.Clone(),
	}

	if t.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*t.StringIndexedFields))
	}
	if t.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*t.PrefixIndexedFields))
	}
	if t.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*t.SuffixIndexedFields))
	}
	if t.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*t.ExistsIndexedFields))
	}
	if t.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*t.NotExistsIndexedFields))
	}
	if t.ActionSConns != nil {
		cln.ActionSConns = slices.Clone(t.ActionSConns)
	}
	return
}

type ThresholdsOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
}

// Threshold service config section
type ThresholdSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	Store_interval           *string
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Actions_conns            *[]string
	Opts                     *ThresholdsOptsJson
}

func diffThresholdsOptsJsonCfg(d *ThresholdsOptsJson, v1, v2 *ThresholdsOpts) *ThresholdsOptsJson {
	if d == nil {
		d = new(ThresholdsOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	return d
}

func diffThresholdSJsonCfg(d *ThresholdSJsonCfg, v1, v2 *ThresholdSCfg) *ThresholdSJsonCfg {
	if d == nil {
		d = new(ThresholdSJsonCfg)
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
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	d.Exists_indexed_fields = diffIndexSlice(d.Exists_indexed_fields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.Notexists_indexed_fields = diffIndexSlice(d.Notexists_indexed_fields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !slices.Equal(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	d.Opts = diffThresholdsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
