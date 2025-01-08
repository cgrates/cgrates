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

const (
	DispatchersDispatchersDftOpt = true
)

type DispatchersOpts struct {
	Dispatchers []*DynamicBoolOpt
}

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	AttributeSConns        []string
	Opts                   *DispatchersOpts
}

// loadDispatcherSCfg loads the DispatcherS section of the configuration
func (dps *DispatcherSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnDispatcherSCfg := new(DispatcherSJsonCfg)
	if err = jsnCfg.GetSection(ctx, DispatcherSJSON, jsnDispatcherSCfg); err != nil {
		return
	}
	return dps.loadFromJSONCfg(jsnDispatcherSCfg)
}

func (dspOpts *DispatchersOpts) loadFromJSONCfg(jsnCfg *DispatchersOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Dispatchers != nil {
		dspOpts.Dispatchers = append(dspOpts.Dispatchers, jsnCfg.Dispatchers...)
	}
}

func (dps *DispatcherSCfg) loadFromJSONCfg(jsnCfg *DispatcherSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		dps.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		dps.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		dps.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		dps.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		dps.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		dps.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		dps.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Attributes_conns != nil {
		dps.AttributeSConns = updateInternalConnsWithPrfx(*jsnCfg.Attributes_conns, utils.MetaAttributes, utils.MetaDispatchers)
	}
	if jsnCfg.Nested_fields != nil {
		dps.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		dps.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dps DispatcherSCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.MetaDispatcherSCfg: dps.Opts.Dispatchers,
	}
	mp := map[string]any{
		utils.EnabledCfg:        dps.Enabled,
		utils.IndexedSelectsCfg: dps.IndexedSelects,
		utils.NestedFieldsCfg:   dps.NestedFields,
		utils.OptsCfg:           opts,
	}
	if dps.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*dps.StringIndexedFields)
	}
	if dps.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*dps.PrefixIndexedFields)
	}
	if dps.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*dps.SuffixIndexedFields)
	}
	if dps.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConnsWithPrfx(dps.AttributeSConns, utils.MetaDispatchers)
	}
	if dps.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*dps.ExistsIndexedFields)
	}
	if dps.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*dps.NotExistsIndexedFields)
	}
	return mp
}

func (DispatcherSCfg) SName() string             { return DispatcherSJSON }
func (dps DispatcherSCfg) CloneSection() Section { return dps.Clone() }

func (dspOpts *DispatchersOpts) Clone() *DispatchersOpts {
	var dpS []*DynamicBoolOpt
	if dspOpts.Dispatchers != nil {
		dpS = CloneDynamicBoolOpt(dspOpts.Dispatchers)
	}
	return &DispatchersOpts{
		Dispatchers: dpS,
	}
}

// Clone returns a deep copy of DispatcherSCfg
func (dps DispatcherSCfg) Clone() (cln *DispatcherSCfg) {
	cln = &DispatcherSCfg{
		Enabled:        dps.Enabled,
		IndexedSelects: dps.IndexedSelects,
		NestedFields:   dps.NestedFields,
		Opts:           dps.Opts.Clone(),
	}

	if dps.AttributeSConns != nil {
		cln.AttributeSConns = slices.Clone(dps.AttributeSConns)
	}
	if dps.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*dps.StringIndexedFields))
	}
	if dps.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*dps.PrefixIndexedFields))
	}
	if dps.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*dps.SuffixIndexedFields))
	}
	if dps.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*dps.ExistsIndexedFields))
	}
	if dps.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*dps.NotExistsIndexedFields))
	}
	return
}

type DispatchersOptsJson struct {
	Dispatchers []*DynamicBoolOpt `json:"*dispatchers"`
}

type DispatcherSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Attributes_conns         *[]string
	Opts                     *DispatchersOptsJson
}

func diffDispatchersOptsJsonCfg(d *DispatchersOptsJson, v1, v2 *DispatchersOpts) *DispatchersOptsJson {
	if d == nil {
		d = new(DispatchersOptsJson)
	}
	if !DynamicBoolOptEqual(v1.Dispatchers, v2.Dispatchers) {
		d.Dispatchers = v2.Dispatchers
	}
	return d
}

func diffDispatcherSJsonCfg(d *DispatcherSJsonCfg, v1, v2 *DispatcherSCfg) *DispatcherSJsonCfg {
	if d == nil {
		d = new(DispatcherSJsonCfg)
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
	d.Opts = diffDispatchersOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
