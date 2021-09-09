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

type AttributesOpts struct {
	AttributeIDs []string
	ProcessRuns  int
	ProfileRuns  int
}

// AttributeSCfg is the configuration of attribute service
type AttributeSCfg struct {
	Enabled             bool
	ResourceSConns      []string
	StatSConns          []string
	AccountSConns       []string
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	Opts                *AttributesOpts
}

func (attrOpts *AttributesOpts) loadFromJSONCfg(jsnCfg *AttributesOptsJson) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.AttributeIDs != nil {
		attrOpts.AttributeIDs = *jsnCfg.AttributeIDs
	}
	if jsnCfg.ProcessRuns != nil {
		attrOpts.ProcessRuns = *jsnCfg.ProcessRuns
	}
	if jsnCfg.ProfileRuns != nil {
		attrOpts.ProfileRuns = *jsnCfg.ProfileRuns
	}

	return nil
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
		alS.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		alS.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		alS.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		alS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		alS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (alS *AttributeSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.MetaAttributeIDsCfg: alS.Opts.AttributeIDs,
		utils.MetaProcessRunsCfg:  alS.Opts.ProcessRuns,
		utils.MetaProfileRunsCfg:  alS.Opts.ProfileRuns,
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        alS.Enabled,
		utils.IndexedSelectsCfg: alS.IndexedSelects,
		utils.NestedFieldsCfg:   alS.NestedFields,
		utils.OptsCfg:           opts,
	}
	if alS.StringIndexedFields != nil {
		initialMP[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*alS.StringIndexedFields)
	}
	if alS.PrefixIndexedFields != nil {
		initialMP[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*alS.PrefixIndexedFields)
	}
	if alS.SuffixIndexedFields != nil {
		initialMP[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*alS.SuffixIndexedFields)
	}
	if alS.StatSConns != nil {
		initialMP[utils.StatSConnsCfg] = getInternalJSONConns(alS.StatSConns)
	}

	if alS.ResourceSConns != nil {
		initialMP[utils.ResourceSConnsCfg] = getInternalJSONConns(alS.ResourceSConns)
	}
	if alS.AccountSConns != nil {
		initialMP[utils.AccountSConnsCfg] = getInternalJSONConns(alS.AccountSConns)
	}
	return
}

func (attrOpts *AttributesOpts) Clone() *AttributesOpts {
	var attrIDs []string
	if attrOpts.AttributeIDs != nil {
		attrIDs = utils.CloneStringSlice(attrOpts.AttributeIDs)
	}
	return &AttributesOpts{
		AttributeIDs: attrIDs,
		ProcessRuns:  attrOpts.ProcessRuns,
		ProfileRuns:  attrOpts.ProfileRuns,
	}
}

// Clone returns a deep copy of AttributeSCfg
func (alS AttributeSCfg) Clone() (cln *AttributeSCfg) {
	cln = &AttributeSCfg{
		Enabled:        alS.Enabled,
		IndexedSelects: alS.IndexedSelects,
		NestedFields:   alS.NestedFields,
	}
	if alS.Opts != nil {
		cln.Opts = alS.Opts.Clone()
	}
	if alS.ResourceSConns != nil {
		cln.ResourceSConns = utils.CloneStringSlice(alS.ResourceSConns)
	}
	if alS.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(alS.StatSConns)
	}
	if alS.AccountSConns != nil {
		cln.AccountSConns = utils.CloneStringSlice(alS.AccountSConns)
	}

	if alS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*alS.StringIndexedFields))
	}
	if alS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*alS.PrefixIndexedFields))
	}
	if alS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*alS.SuffixIndexedFields))
	}
	return
}

type AttributesOptsJson struct {
	AttributeIDs *[]string `json:"*attributeIDs"`
	ProcessRuns  *int      `json:"*processRuns"`
	ProfileRuns  *int      `json:"*profileRuns"`
}

// Attribute service config section
type AttributeSJsonCfg struct {
	Enabled               *bool
	Stats_conns           *[]string
	Resources_conns       *[]string
	Accounts_conns        *[]string
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Opts                  *AttributesOptsJson
}

func diffAttributesOptsJsonCfg(d *AttributesOptsJson, v1, v2 *AttributesOpts) *AttributesOptsJson {
	if d == nil {
		d = new(AttributesOptsJson)
	}
	if !utils.SliceStringEqual(v1.AttributeIDs, v2.AttributeIDs) {
		d.AttributeIDs = utils.SliceStringPointer(v2.AttributeIDs)
	}
	if v1.ProcessRuns != v2.ProcessRuns {
		d.ProcessRuns = utils.IntPointer(v2.ProcessRuns)
	}
	if v1.ProfileRuns != v2.ProfileRuns {
		d.ProfileRuns = utils.IntPointer(v2.ProfileRuns)
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
	if !utils.SliceStringEqual(v1.ResourceSConns, v2.ResourceSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
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
	d.Opts = diffAttributesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}

func diffIndexSlice(d, v1, v2 *[]string) *[]string {
	if v2 == nil {
		return nil
	}
	if v1 == nil || !utils.SliceStringEqual(*v1, *v2) {
		d = utils.SliceStringPointer(utils.CloneStringSlice(*v2))
	}
	return d
}
