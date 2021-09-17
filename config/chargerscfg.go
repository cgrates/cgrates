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
)

// ChargerSCfg is the configuration of charger service
type ChargerSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	AttributeSConns     []string
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
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
		cS.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.String_indexed_fields != nil {
		cS.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		cS.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		cS.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		cS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (cS ChargerSCfg) AsMapInterface(string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:        cS.Enabled,
		utils.IndexedSelectsCfg: cS.IndexedSelects,
		utils.NestedFieldsCfg:   cS.NestedFields,
	}
	if cS.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(cS.AttributeSConns)
	}
	if cS.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*cS.StringIndexedFields)
	}
	if cS.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*cS.PrefixIndexedFields)
	}
	if cS.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*cS.SuffixIndexedFields)
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
		cln.AttributeSConns = utils.CloneStringSlice(cS.AttributeSConns)
	}

	if cS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*cS.StringIndexedFields))
	}
	if cS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*cS.PrefixIndexedFields))
	}
	if cS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*cS.SuffixIndexedFields))
	}
	return
}

// ChargerSJsonCfg service config section
type ChargerSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Attributes_conns      *[]string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
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
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	return d
}
