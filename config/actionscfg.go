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

var ActionsProfileIDsDftOpt = []string{}

const (
	ActionsProfileIgnoreFiltersDftOpt = false
	ActionsPosterAttempsDftOpt        = 1
)

type ActionsOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
	PosterAttempts       []*DynamicIntOpt
}

// ActionSCfg is the configuration of ActionS
type ActionSCfg struct {
	Enabled                  bool
	CDRsConns                []string
	EEsConns                 []string
	ThresholdSConns          []string
	StatSConns               []string
	AccountSConns            []string
	AdminSConns              []string
	Tenants                  *[]string
	IndexedSelects           bool
	StringIndexedFields      *[]string
	PrefixIndexedFields      *[]string
	SuffixIndexedFields      *[]string
	ExistsIndexedFields      *[]string
	NotExistsIndexedFields   *[]string
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
		return
	}
	if jsnCfg.ProfileIDs != nil {
		actOpts.ProfileIDs = append(actOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var prfIgnFltr []*DynamicBoolOpt
		prfIgnFltr, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		actOpts.ProfileIgnoreFilters = append(prfIgnFltr, actOpts.ProfileIgnoreFilters...)
	}
	if jsnCfg.PosterAttempts != nil {
		var pstrAtt []*DynamicIntOpt
		pstrAtt, err = IfaceToIntDynamicOpts(jsnCfg.PosterAttempts)
		actOpts.PosterAttempts = append(actOpts.PosterAttempts, pstrAtt...)
	}
	return
}

func (acS *ActionSCfg) loadFromJSONCfg(jsnCfg *ActionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		acS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrs_conns != nil {
		acS.CDRsConns = tagInternalConns(*jsnCfg.Cdrs_conns, utils.MetaCDRs)
	}
	if jsnCfg.Ees_conns != nil {
		acS.EEsConns = tagInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCfg.Thresholds_conns != nil {
		acS.ThresholdSConns = tagInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Stats_conns != nil {
		acS.StatSConns = tagInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Accounts_conns != nil {
		acS.AccountSConns = tagInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Admins_conns != nil {
		acS.AdminSConns = tagInternalConns(*jsnCfg.Admins_conns, utils.MetaAdminS)
	}
	if jsnCfg.Tenants != nil {
		acS.Tenants = utils.SliceStringPointer(slices.Clone(*jsnCfg.Tenants))
	}
	if jsnCfg.Indexed_selects != nil {
		acS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		acS.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		acS.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		acS.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		acS.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		acS.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		acS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Dynaprepaid_actionprofile != nil {
		acS.DynaprepaidActionProfile = make([]string, len(*jsnCfg.Dynaprepaid_actionprofile))
		copy(acS.DynaprepaidActionProfile, *jsnCfg.Dynaprepaid_actionprofile)
	}
	if jsnCfg.Opts != nil {
		err = acS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (acS ActionSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaProfileIDs:           acS.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFilters: acS.Opts.ProfileIgnoreFilters,
		utils.MetaPosterAttempts:       acS.Opts.PosterAttempts,
	}
	mp := map[string]any{
		utils.EnabledCfg:                acS.Enabled,
		utils.IndexedSelectsCfg:         acS.IndexedSelects,
		utils.NestedFieldsCfg:           acS.NestedFields,
		utils.DynaprepaidActionplansCfg: acS.DynaprepaidActionProfile,
		utils.OptsCfg:                   opts,
	}
	if acS.CDRsConns != nil {
		mp[utils.CDRsConnsCfg] = stripInternalConns(acS.CDRsConns)
	}
	if acS.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = stripInternalConns(acS.ThresholdSConns)
	}
	if acS.StatSConns != nil {
		mp[utils.StatSConnsCfg] = stripInternalConns(acS.StatSConns)
	}
	if acS.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = stripInternalConns(acS.AccountSConns)
	}
	if acS.EEsConns != nil {
		mp[utils.EEsConnsCfg] = stripInternalConns(acS.EEsConns)
	}
	if acS.AdminSConns != nil {
		mp[utils.AdminSConnsCfg] = stripInternalConns(acS.AdminSConns)
	}
	if acS.Tenants != nil {
		mp[utils.Tenants] = slices.Clone(*acS.Tenants)
	}
	if acS.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*acS.StringIndexedFields)
	}
	if acS.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*acS.PrefixIndexedFields)
	}
	if acS.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*acS.SuffixIndexedFields)
	}
	if acS.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*acS.ExistsIndexedFields)
	}
	if acS.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*acS.NotExistsIndexedFields)
	}
	return mp
}

func (ActionSCfg) SName() string             { return ActionSJSON }
func (acS ActionSCfg) CloneSection() Section { return acS.Clone() }

func (actOpts *ActionsOpts) Clone() *ActionsOpts {
	var actPrfIDs []*DynamicStringSliceOpt
	if actOpts.ProfileIDs != nil {
		actPrfIDs = CloneDynamicStringSliceOpt(actOpts.ProfileIDs)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if actOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(actOpts.ProfileIgnoreFilters)
	}
	var posterAttempts []*DynamicIntOpt
	if actOpts.PosterAttempts != nil {
		posterAttempts = CloneDynamicIntOpt(actOpts.PosterAttempts)
	}
	return &ActionsOpts{
		ProfileIDs:           actPrfIDs,
		ProfileIgnoreFilters: profileIgnoreFilters,
		PosterAttempts:       posterAttempts,
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
		cln.CDRsConns = slices.Clone(acS.CDRsConns)
	}
	if acS.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(acS.ThresholdSConns)
	}
	if acS.StatSConns != nil {
		cln.StatSConns = slices.Clone(acS.StatSConns)
	}
	if acS.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(acS.AccountSConns)
	}
	if acS.EEsConns != nil {
		cln.EEsConns = slices.Clone(acS.EEsConns)
	}
	if acS.AdminSConns != nil {
		cln.AdminSConns = slices.Clone(acS.AdminSConns)
	}
	if acS.Tenants != nil {
		cln.Tenants = utils.SliceStringPointer(slices.Clone(*acS.Tenants))
	}
	if acS.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*acS.StringIndexedFields))
	}
	if acS.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*acS.PrefixIndexedFields))
	}
	if acS.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*acS.SuffixIndexedFields))
	}
	if acS.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*acS.ExistsIndexedFields))
	}
	if acS.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*acS.NotExistsIndexedFields))
	}
	if acS.DynaprepaidActionProfile != nil {
		cln.DynaprepaidActionProfile = make([]string, len(acS.DynaprepaidActionProfile))
		copy(cln.DynaprepaidActionProfile, acS.DynaprepaidActionProfile)
	}
	return
}

type ActionsOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
	PosterAttempts       []*DynamicInterfaceOpt   `json:"*posterAttempts"`
}

// Action service config section
type ActionSJsonCfg struct {
	Enabled                   *bool
	Cdrs_conns                *[]string
	Ees_conns                 *[]string
	Thresholds_conns          *[]string
	Stats_conns               *[]string
	Accounts_conns            *[]string
	Admins_conns              *[]string
	Tenants                   *[]string
	Indexed_selects           *bool
	String_indexed_fields     *[]string
	Prefix_indexed_fields     *[]string
	Suffix_indexed_fields     *[]string
	Exists_indexed_fields     *[]string
	Notexists_indexed_fields  *[]string
	Nested_fields             *bool // applies when indexed fields is not defined
	Dynaprepaid_actionprofile *[]string
	Opts                      *ActionsOptsJson
}

func diffActionsOptsJsonCfg(d *ActionsOptsJson, v1, v2 *ActionsOpts) *ActionsOptsJson {
	if d == nil {
		d = new(ActionsOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	if !DynamicIntOptEqual(v1.PosterAttempts, v2.PosterAttempts) {
		d.PosterAttempts = IntToIfaceDynamicOpts(v2.PosterAttempts)
	}
	return d
}

func diffActionSJsonCfg(d *ActionSJsonCfg, v1, v2 *ActionSCfg) *ActionSJsonCfg {
	if d == nil {
		d = new(ActionSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.CDRsConns, v2.CDRsConns) {
		d.Cdrs_conns = utils.SliceStringPointer(stripInternalConns(v2.CDRsConns))
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(stripInternalConns(v2.EEsConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(stripInternalConns(v2.StatSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(stripInternalConns(v2.AccountSConns))
	}
	if !slices.Equal(v1.AdminSConns, v2.AdminSConns) {
		d.Admins_conns = utils.SliceStringPointer(stripInternalConns(v2.AdminSConns))
	}

	if v1.Tenants != v2.Tenants {
		d.Tenants = utils.SliceStringPointer(slices.Clone(*v2.Tenants))
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
	if !slices.Equal(v1.DynaprepaidActionProfile, v2.DynaprepaidActionProfile) {
		d.Dynaprepaid_actionprofile = utils.SliceStringPointer(stripInternalConns(v2.DynaprepaidActionProfile))
	}
	d.Opts = diffActionsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
