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

type ActionsOpts struct {
	ActionProfileIDs map[string][]string
}

// ActionSCfg is the configuration of ActionS
type ActionSCfg struct {
	Enabled                  bool
	CDRsConns                []string
	EEsConns                 []string
	ThresholdSConns          []string
	StatSConns               []string
	AccountSConns            []string
	Tenants                  *[]string
	IndexedSelects           bool
	StringIndexedFields      *[]string
	PrefixIndexedFields      *[]string
	SuffixIndexedFields      *[]string
	NestedFields             bool
	DynaprepaidActionProfile []string
	Opts                     *ActionsOpts
}

// loadActionSCfg loads the ActionS section of the configuration
func (acS *ActionSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnActionCfg := new(ActionSJsonCfg)
	if err = jsnCfg.GetSection(ctx, ActionSJSON, jsnActionCfg); err != nil {
		return
	}
	return acS.loadFromJSONCfg(jsnActionCfg)
}

func (actOpts *ActionsOpts) loadFromJSONCfg(jsnCfg *ActionsOptsJson) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ActionProfileIDs != nil {
		actOpts.ActionProfileIDs = jsnCfg.ActionProfileIDs
	}

	return nil
}

func (acS *ActionSCfg) loadFromJSONCfg(jsnCfg *ActionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		acS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrs_conns != nil {
		acS.CDRsConns = updateInternalConns(*jsnCfg.Cdrs_conns, utils.MetaCDRs)
	}
	if jsnCfg.Ees_conns != nil {
		acS.EEsConns = updateInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCfg.Thresholds_conns != nil {
		acS.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Stats_conns != nil {
		acS.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Accounts_conns != nil {
		acS.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Tenants != nil {
		acS.Tenants = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Tenants))
	}
	if jsnCfg.Indexed_selects != nil {
		acS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		acS.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		acS.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		acS.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		acS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Dynaprepaid_actionprofile != nil {
		acS.DynaprepaidActionProfile = make([]string, len(*jsnCfg.Dynaprepaid_actionprofile))
		for i, val := range *jsnCfg.Dynaprepaid_actionprofile {
			acS.DynaprepaidActionProfile[i] = val
		}
	}
	if jsnCfg.Opts != nil {
		acS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (acS ActionSCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.MetaActionProfileIDsCfg: acS.Opts.ActionProfileIDs,
	}
	mp := map[string]interface{}{
		utils.EnabledCfg:                acS.Enabled,
		utils.IndexedSelectsCfg:         acS.IndexedSelects,
		utils.NestedFieldsCfg:           acS.NestedFields,
		utils.DynaprepaidActionplansCfg: acS.DynaprepaidActionProfile,
		utils.OptsCfg:                   opts,
	}
	if acS.CDRsConns != nil {
		mp[utils.CDRsConnsCfg] = getInternalJSONConns(acS.CDRsConns)
	}
	if acS.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(acS.ThresholdSConns)
	}
	if acS.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(acS.StatSConns)
	}
	if acS.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(acS.AccountSConns)
	}
	if acS.EEsConns != nil {
		mp[utils.EEsConnsCfg] = getInternalJSONConns(acS.EEsConns)
	}
	if acS.Tenants != nil {
		mp[utils.Tenants] = utils.CloneStringSlice(*acS.Tenants)
	}
	if acS.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*acS.StringIndexedFields)
	}
	if acS.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*acS.PrefixIndexedFields)
	}
	if acS.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*acS.SuffixIndexedFields)
	}
	return mp
}

func (ActionSCfg) SName() string             { return ActionSJSON }
func (acS ActionSCfg) CloneSection() Section { return acS.Clone() }

func (actOpts *ActionsOpts) Clone() *ActionsOpts {
	return &ActionsOpts{
		ActionProfileIDs: actOpts.ActionProfileIDs,
	}
}

// Clone returns a deep copy of ActionSCfg
func (acS ActionSCfg) Clone() (cln *ActionSCfg) {
	cln = &ActionSCfg{
		Enabled:        acS.Enabled,
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
		Opts:           acS.Opts.Clone(),
	}
	if acS.CDRsConns != nil {
		cln.CDRsConns = utils.CloneStringSlice(acS.CDRsConns)
	}
	if acS.ThresholdSConns != nil {
		cln.ThresholdSConns = utils.CloneStringSlice(acS.ThresholdSConns)
	}
	if acS.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(acS.StatSConns)
	}
	if acS.AccountSConns != nil {
		cln.AccountSConns = utils.CloneStringSlice(acS.AccountSConns)
	}
	if acS.EEsConns != nil {
		cln.EEsConns = utils.CloneStringSlice(acS.EEsConns)
	}
	if acS.Tenants != nil {
		cln.Tenants = utils.SliceStringPointer(utils.CloneStringSlice(*acS.Tenants))
	}
	if acS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*acS.StringIndexedFields))
	}
	if acS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*acS.PrefixIndexedFields))
	}
	if acS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*acS.SuffixIndexedFields))
	}
	if acS.DynaprepaidActionProfile != nil {
		cln.DynaprepaidActionProfile = make([]string, len(acS.DynaprepaidActionProfile))
		for i, con := range acS.DynaprepaidActionProfile {
			cln.DynaprepaidActionProfile[i] = con
		}
	}
	return
}

type ActionsOptsJson struct {
	ActionProfileIDs map[string][]string `json:"*actionProfileIDs"`
}

// Action service config section
type ActionSJsonCfg struct {
	Enabled                   *bool
	Cdrs_conns                *[]string
	Ees_conns                 *[]string
	Thresholds_conns          *[]string
	Stats_conns               *[]string
	Accounts_conns            *[]string
	Tenants                   *[]string
	Indexed_selects           *bool
	String_indexed_fields     *[]string
	Prefix_indexed_fields     *[]string
	Suffix_indexed_fields     *[]string
	Nested_fields             *bool // applies when indexed fields is not defined
	Dynaprepaid_actionprofile *[]string
	Opts                      *ActionsOptsJson
}

func diffActionsOptsJsonCfg(d *ActionsOptsJson, v1, v2 *ActionsOpts) *ActionsOptsJson {
	if d == nil {
		d = new(ActionsOptsJson)
	}
	d.ActionProfileIDs = diffMapStringStringSlice(d.ActionProfileIDs, v1.ActionProfileIDs, v2.ActionProfileIDs)
	return d
}

func diffActionSJsonCfg(d *ActionSJsonCfg, v1, v2 *ActionSCfg) *ActionSJsonCfg {
	if d == nil {
		d = new(ActionSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.CDRsConns, v2.CDRsConns) {
		d.Cdrs_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CDRsConns))
	}
	if !utils.SliceStringEqual(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}

	if v1.Tenants != v2.Tenants {
		d.Tenants = utils.SliceStringPointer(utils.CloneStringSlice(*v2.Tenants))
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if v1.StringIndexedFields != v2.StringIndexedFields {
		d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	}
	if v1.PrefixIndexedFields != v2.PrefixIndexedFields {
		d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	}
	if v1.SuffixIndexedFields != v2.SuffixIndexedFields {
		d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	}
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !utils.SliceStringEqual(v1.DynaprepaidActionProfile, v2.DynaprepaidActionProfile) {
		d.Dynaprepaid_actionprofile = utils.SliceStringPointer(getInternalJSONConns(v2.DynaprepaidActionProfile))
	}
	d.Opts = diffActionsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
