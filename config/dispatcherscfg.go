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

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	AttributeSConns     []string
}

// loadDispatcherSCfg loads the DispatcherS section of the configuration
func (dps *DispatcherSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnDispatcherSCfg := new(DispatcherSJsonCfg)
	if err = jsnCfg.GetSection(ctx, DispatcherSJSON, jsnDispatcherSCfg); err != nil {
		return
	}
	return dps.loadFromJSONCfg(jsnDispatcherSCfg)
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
		dps.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		dps.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		dps.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Attributes_conns != nil {
		dps.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.Nested_fields != nil {
		dps.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (dps DispatcherSCfg) AsMapInterface(string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:        dps.Enabled,
		utils.IndexedSelectsCfg: dps.IndexedSelects,
		utils.NestedFieldsCfg:   dps.NestedFields,
	}
	if dps.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*dps.StringIndexedFields)
	}
	if dps.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*dps.PrefixIndexedFields)
	}
	if dps.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*dps.SuffixIndexedFields)
	}
	if dps.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(dps.AttributeSConns)
	}
	return mp
}

func (DispatcherSCfg) SName() string             { return DispatcherSJSON }
func (dps DispatcherSCfg) CloneSection() Section { return dps.Clone() }

// Clone returns a deep copy of DispatcherSCfg
func (dps DispatcherSCfg) Clone() (cln *DispatcherSCfg) {
	cln = &DispatcherSCfg{
		Enabled:        dps.Enabled,
		IndexedSelects: dps.IndexedSelects,
		NestedFields:   dps.NestedFields,
	}

	if dps.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(dps.AttributeSConns)
	}
	if dps.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*dps.StringIndexedFields))
	}
	if dps.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*dps.PrefixIndexedFields))
	}
	if dps.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*dps.SuffixIndexedFields))
	}
	return
}

type DispatcherSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
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
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	return d
}
