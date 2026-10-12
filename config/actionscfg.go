// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	HttpBody             []*DynamicInterfaceOpt
	HttpHdrs             []*DynamicInterfaceOpt
	HttpContentType      []*DynamicStringOpt
	HttpMethod           []*DynamicStringOpt
	HttpURL              []*DynamicStringOpt
	ProfileIDs           []*DynamicStringSliceOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
	PosterAttempts       []*DynamicIntOpt
}

// ActionSCfg is the configuration of ActionS
type ActionSCfg struct {
	Enabled                  bool
	Conns                    map[string][]*DynamicConns
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
	if jsnCfg.HttpBody != nil {
		actOpts.HttpBody = append(jsnCfg.HttpBody, actOpts.HttpBody...)
	}
	if jsnCfg.HttpContentType != nil {
		httpContentType, err := InterfaceToDynamicStringOpts(jsnCfg.HttpContentType)
		if err != nil {
			return err
		}
		actOpts.HttpContentType = append(httpContentType, actOpts.HttpContentType...)
	}
	if jsnCfg.HttpMethod != nil {
		method, err := InterfaceToDynamicStringOpts(jsnCfg.HttpMethod)
		if err != nil {
			return err
		}
		actOpts.HttpMethod = append(method, actOpts.HttpMethod...)
	}
	if jsnCfg.HttpHdrs != nil {
		actOpts.HttpHdrs = append(jsnCfg.HttpHdrs, actOpts.HttpHdrs...)
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
	if jsnCfg.HttpURL != nil {
		url, err := InterfaceToDynamicStringOpts(jsnCfg.HttpURL)
		if err != nil {
			return err
		}
		actOpts.HttpURL = append(url, actOpts.HttpURL...)
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			acS.Conns[connType] = opts
		}
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
		utils.MetaHTTPBody:             acS.Opts.HttpBody,
		utils.MetaHTTPHdrs:             acS.Opts.HttpHdrs,
		utils.MetaHTTPContentType:      acS.Opts.HttpContentType,
		utils.MetaHTTPMethod:           acS.Opts.HttpMethod,
		utils.MetaProfileIDs:           acS.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFilters: acS.Opts.ProfileIgnoreFilters,
		utils.MetaPosterAttempts:       acS.Opts.PosterAttempts,
		utils.MetaHTTPurl:              acS.Opts.HttpURL,
	}
	mp := map[string]any{
		utils.EnabledCfg:                acS.Enabled,
		utils.ConnsCfg:                  stripConns(acS.Conns),
		utils.IndexedSelectsCfg:         acS.IndexedSelects,
		utils.NestedFieldsCfg:           acS.NestedFields,
		utils.DynaprepaidActionplansCfg: acS.DynaprepaidActionProfile,
		utils.OptsCfg:                   opts,
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
	var body []*DynamicInterfaceOpt
	if actOpts.HttpBody != nil {
		body = CloneDynamicInterfaceOpt(actOpts.HttpBody)
	}
	var hdrs []*DynamicInterfaceOpt
	if actOpts.HttpHdrs != nil {
		hdrs = CloneDynamicInterfaceOpt(actOpts.HttpHdrs)
	}
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
		HttpBody:             body,
		HttpHdrs:             hdrs,
		HttpContentType:      CloneDynamicStringOpt(actOpts.HttpContentType),
		HttpMethod:           CloneDynamicStringOpt(actOpts.HttpMethod),
		ProfileIDs:           actPrfIDs,
		ProfileIgnoreFilters: profileIgnoreFilters,
		PosterAttempts:       posterAttempts,
		HttpURL:              CloneDynamicStringOpt(actOpts.HttpURL),
	}
}

// Clone returns a deep copy of ActionSCfg
func (acS ActionSCfg) Clone() (cln *ActionSCfg) {
	cln = &ActionSCfg{
		Enabled:        acS.Enabled,
		Conns:          CloneConnsMap(acS.Conns),
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
		Opts:           acS.Opts.Clone(),
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
	HttpBody             []*DynamicInterfaceOpt   `json:"*httpBody"`
	HttpHdrs             []*DynamicInterfaceOpt   `json:"*httpHdrs"`
	HttpContentType      []*DynamicInterfaceOpt   `json:"*httpContentType"`
	HttpMethod           []*DynamicInterfaceOpt   `json:"*httpMethod"`
	HttpURL              []*DynamicInterfaceOpt   `json:"*httpURL"`
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
	PosterAttempts       []*DynamicInterfaceOpt   `json:"*posterAttempts"`
}

// Action service config section
type ActionSJsonCfg struct {
	Enabled                   *bool
	Conns                     map[string][]*DynamicConns `json:"conns,omitempty"`
	Tenants                   *[]string
	Indexed_selects           *bool     `json:"indexedSelects"`
	String_indexed_fields     *[]string `json:"stringIndexedFields"`
	Prefix_indexed_fields     *[]string `json:"prefixIndexedFields"`
	Suffix_indexed_fields     *[]string `json:"suffixIndexedFields"`
	Exists_indexed_fields     *[]string `json:"existsIndexedFields"`
	Notexists_indexed_fields  *[]string `json:"notExistsIndexedFields"`
	Nested_fields             *bool     `json:"nestedFields"` // applies when indexed fields is not defined
	Dynaprepaid_actionprofile *[]string `json:"dynaprepaidActionProfile"`
	Opts                      *ActionsOptsJson
}

func diffActionsOptsJsonCfg(d *ActionsOptsJson, v1, v2 *ActionsOpts) *ActionsOptsJson {
	if d == nil {
		d = new(ActionsOptsJson)
	}
	if !DynamicInterfaceOptEqual(v1.HttpBody, v2.HttpBody) {
		d.HttpBody = v2.HttpBody
	}
	if !DynamicInterfaceOptEqual(v1.HttpHdrs, v2.HttpHdrs) {
		d.HttpHdrs = v2.HttpHdrs
	}
	if !DynamicStringOptEqual(v1.HttpContentType, v2.HttpContentType) {
		d.HttpContentType = DynamicStringToInterfaceOpts(v2.HttpContentType)
	}
	if !DynamicStringOptEqual(v1.HttpMethod, v2.HttpMethod) {
		d.HttpMethod = DynamicStringToInterfaceOpts(v2.HttpMethod)
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
	if !DynamicStringOptEqual(v1.HttpURL, v2.HttpURL) {
		d.HttpURL = DynamicStringToInterfaceOpts(v2.HttpURL)
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
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
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
