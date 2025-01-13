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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

var AttributesProfileIDsDftOpt = []string{}

const (
	AttributesProcessRunsDftOpt          = 1
	AttributesProfileRunsDftOpt          = 0
	AttributesProfileIgnoreFiltersDftOpt = false
)

type AttributesOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	ProcessRuns          []*DynamicIntOpt
	ProfileRuns          []*DynamicIntOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
}

// AttributeSCfg is the configuration of attribute service
type AttributeSCfg struct {
	Enabled                bool
	ResourceSConns         []string
	StatSConns             []string
	AccountSConns          []string
	IndexedSelects         bool
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	Opts                   *AttributesOpts
}

// loadAttributeSCfg loads the AttributeS section of the configuration
func (alS *AttributeSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnAttributeSCfg := new(AttributeSJsonCfg)
	if err = jsnCfg.GetSection(ctx, AttributeSJSON, jsnAttributeSCfg); err != nil {
		return
	}
	return alS.loadFromJSONCfg(jsnAttributeSCfg)
}

func (attrOpts *AttributesOpts) loadFromJSONCfg(jsnCfg *AttributesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		attrOpts.ProfileIDs = append(attrOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.ProcessRuns != nil {
		var procRuns []*DynamicIntOpt
		procRuns, err = IfaceToIntDynamicOpts(jsnCfg.ProcessRuns)
		procRuns = append(procRuns, attrOpts.ProcessRuns...)
		attrOpts.ProcessRuns = procRuns
	}
	if jsnCfg.ProfileRuns != nil {
		var profRuns []*DynamicIntOpt
		profRuns, err = IfaceToIntDynamicOpts(jsnCfg.ProfileRuns)
		if err != nil {
			return
		}
		profRuns = append(profRuns, attrOpts.ProfileRuns...)
		attrOpts.ProfileRuns = profRuns
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var profIgnFltrs []*DynamicBoolOpt
		profIgnFltrs, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		if err != nil {
			return
		}
		attrOpts.ProfileIgnoreFilters = append(profIgnFltrs, attrOpts.ProfileIgnoreFilters...)
	}
	return
}

func (alS *AttributeSCfg) loadFromJSONCfg(jsnCfg *AttributeSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		alS.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Resources_conns != nil {
		alS.ResourceSConns = updateInternalConns(*jsnCfg.Resources_conns, utils.MetaResources)
	}
	if jsnCfg.Accounts_conns != nil {
		alS.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Indexed_selects != nil {
		alS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		alS.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		alS.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		alS.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		alS.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		alS.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		alS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		err = alS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (alS AttributeSCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.MetaProfileIDs:           alS.Opts.ProfileIDs,
		utils.MetaProcessRunsCfg:       alS.Opts.ProcessRuns,
		utils.MetaProfileRunsCfg:       alS.Opts.ProfileRuns,
		utils.MetaProfileIgnoreFilters: alS.Opts.ProfileIgnoreFilters,
	}
	mp := map[string]any{
		utils.EnabledCfg:        alS.Enabled,
		utils.IndexedSelectsCfg: alS.IndexedSelects,
		utils.NestedFieldsCfg:   alS.NestedFields,
		utils.OptsCfg:           opts,
	}
	if alS.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*alS.StringIndexedFields)
	}
	if alS.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*alS.PrefixIndexedFields)
	}
	if alS.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*alS.SuffixIndexedFields)
	}
	if alS.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*alS.ExistsIndexedFields)
	}
	if alS.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*alS.NotExistsIndexedFields)
	}
	if alS.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(alS.StatSConns)
	}
	if alS.ResourceSConns != nil {
		mp[utils.ResourceSConnsCfg] = getInternalJSONConns(alS.ResourceSConns)
	}
	if alS.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(alS.AccountSConns)
	}
	return mp
}

func (attrOpts AttributesOpts) Clone() *AttributesOpts {
	var attrIDs []*DynamicStringSliceOpt
	if attrOpts.ProfileIDs != nil {
		attrIDs = CloneDynamicStringSliceOpt(attrOpts.ProfileIDs)
	}
	var processRuns []*DynamicIntOpt
	if attrOpts.ProcessRuns != nil {
		processRuns = CloneDynamicIntOpt(attrOpts.ProcessRuns)
	}
	var profileRuns []*DynamicIntOpt
	if attrOpts.ProfileRuns != nil {
		profileRuns = CloneDynamicIntOpt(attrOpts.ProfileRuns)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if attrOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(attrOpts.ProfileIgnoreFilters)
	}
	return &AttributesOpts{
		ProfileIDs:           attrIDs,
		ProcessRuns:          processRuns,
		ProfileRuns:          profileRuns,
		ProfileIgnoreFilters: profileIgnoreFilters,
	}
}

func (AttributeSCfg) SName() string             { return AttributeSJSON }
func (alS AttributeSCfg) CloneSection() Section { return alS.Clone() }

// Clone returns a deep copy of AttributeSCfg
func (alS AttributeSCfg) Clone() (cln *AttributeSCfg) {
	cln = &AttributeSCfg{
		Enabled:        alS.Enabled,
		IndexedSelects: alS.IndexedSelects,
		NestedFields:   alS.NestedFields,
		Opts:           alS.Opts.Clone(),
	}
	if alS.ResourceSConns != nil {
		cln.ResourceSConns = slices.Clone(alS.ResourceSConns)
	}
	if alS.StatSConns != nil {
		cln.StatSConns = slices.Clone(alS.StatSConns)
	}
	if alS.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(alS.AccountSConns)
	}

	if alS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*alS.StringIndexedFields))
	}
	if alS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*alS.PrefixIndexedFields))
	}
	if alS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*alS.SuffixIndexedFields))
	}
	if alS.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*alS.ExistsIndexedFields))
	}
	if alS.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*alS.NotExistsIndexedFields))
	}
	return
}

type AttributesOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	ProcessRuns          []*DynamicInterfaceOpt   `json:"*processRuns"`
	ProfileRuns          []*DynamicInterfaceOpt   `json:"*profileRuns"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
}

// Attribute service config section
type AttributeSJsonCfg struct {
	Enabled                  *bool
	Stats_conns              *[]string
	Resources_conns          *[]string
	Accounts_conns           *[]string
	Indexed_selects          *bool
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Opts                     *AttributesOptsJson
}

func diffAttributesOptsJsonCfg(d *AttributesOptsJson, v1, v2 *AttributesOpts) *AttributesOptsJson {
	if d == nil {
		d = new(AttributesOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicIntOptEqual(v1.ProcessRuns, v2.ProcessRuns) {
		d.ProcessRuns = IntToIfaceDynamicOpts(v2.ProcessRuns)
	}
	if !DynamicIntOptEqual(v1.ProfileRuns, v2.ProfileRuns) {
		d.ProfileRuns = IntToIfaceDynamicOpts(v2.ProfileRuns)
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	return d
}

func diffAttributeSJsonCfg(d *AttributeSJsonCfg, v1, v2 *AttributeSCfg) *AttributeSJsonCfg {
	if d == nil {
		d = new(AttributeSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.ResourceSConns, v2.ResourceSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
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
	d.Opts = diffAttributesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}

func diffIndexSlice(d, v1, v2 *[]string) *[]string {
	if v2 == nil {
		return nil
	}
	if v1 == nil || !slices.Equal(*v1, *v2) {
		d = utils.SliceStringPointer(slices.Clone(*v2))
	}
	return d
}
