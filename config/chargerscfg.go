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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// ChargerSCfg is the configuration of charger service
type ChargerSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	AttributeSConns        []string
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
}

// loadChargerSCfg loads the ChargerS section of the configuration
func (cS *ChargerSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnChargerSCfg := new(ChargerSJsonCfg)
	if err = jsnCfg.GetSection(ctx, ChargerSJSON, jsnChargerSCfg); err != nil {
		return
	}
	return cS.loadFromJSONCfg(jsnChargerSCfg)
}

func (cS *ChargerSCfg) loadFromJSONCfg(jsnCfg *ChargerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		cS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		cS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Attributes_conns != nil {
		cS.AttributeSConns = tagInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.String_indexed_fields != nil {
		cS.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		cS.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		cS.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		cS.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		cS.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		cS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cS ChargerSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:        cS.Enabled,
		utils.IndexedSelectsCfg: cS.IndexedSelects,
		utils.NestedFieldsCfg:   cS.NestedFields,
	}
	if cS.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = stripInternalConns(cS.AttributeSConns)
	}
	if cS.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*cS.StringIndexedFields)
	}
	if cS.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*cS.PrefixIndexedFields)
	}
	if cS.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*cS.SuffixIndexedFields)
	}
	if cS.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*cS.ExistsIndexedFields)
	}
	if cS.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*cS.NotExistsIndexedFields)
	}
	return mp
}

func (ChargerSCfg) SName() string            { return ChargerSJSON }
func (cS ChargerSCfg) CloneSection() Section { return cS.Clone() }

// Clone returns a deep copy of ChargerSCfg
func (cS ChargerSCfg) Clone() (cln *ChargerSCfg) {
	cln = &ChargerSCfg{
		Enabled:        cS.Enabled,
		IndexedSelects: cS.IndexedSelects,
		NestedFields:   cS.NestedFields,
	}
	if cS.AttributeSConns != nil {
		cln.AttributeSConns = slices.Clone(cS.AttributeSConns)
	}

	if cS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*cS.StringIndexedFields))
	}
	if cS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*cS.PrefixIndexedFields))
	}
	if cS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*cS.SuffixIndexedFields))
	}
	if cS.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*cS.ExistsIndexedFields))
	}
	if cS.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*cS.NotExistsIndexedFields))
	}
	return
}

// ChargerSJsonCfg service config section
type ChargerSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	Attributes_conns         *[]string
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
}

func diffChargerSJsonCfg(d *ChargerSJsonCfg, v1, v2 *ChargerSCfg) *ChargerSJsonCfg {
	if d == nil {
		d = new(ChargerSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if !slices.Equal(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(stripInternalConns(v2.AttributeSConns))
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	d.Exists_indexed_fields = diffIndexSlice(d.Exists_indexed_fields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.Notexists_indexed_fields = diffIndexSlice(d.Notexists_indexed_fields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	return d
}
