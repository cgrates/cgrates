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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

var (
	AccountsProfileIDsDftOpt = []string{}
	AccountsUsageDftOpt      = decimal.New(int64(time.Minute), 0)
)

const AccountsProfileIgnoreFiltersDftOpt = false

type AccountsOpts struct {
	ProfileIDs           []*DynamicStringSliceOpt
	Usage                []*DynamicDecimalOpt
	ProfileIgnoreFilters []*DynamicBoolOpt
}

// AccountSCfg is the configuration of ActionS
type AccountSCfg struct {
	Enabled                bool
	Conns                  map[string][]*DynamicConns
	IndexedSelects         bool
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	MaxIterations          int
	MaxUsage               *utils.Decimal
	Opts                   *AccountsOpts
}

// loadAccountSCfg loads the AccountS section of the configuration
func (acS *AccountSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnActionCfg := new(AccountSJsonCfg)
	if err = jsnCfg.GetSection(ctx, AccountSJSON, jsnActionCfg); err != nil {
		return
	}
	return acS.loadFromJSONCfg(jsnActionCfg)
}

func (accOpts *AccountsOpts) loadFromJSONCfg(jsnCfg *AccountsOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		accOpts.ProfileIDs = append(accOpts.ProfileIDs, jsnCfg.ProfileIDs...)
	}
	if jsnCfg.Usage != nil {
		var usage []*DynamicDecimalOpt
		if usage, err = IfaceToDecimalBigDynamicOpts(jsnCfg.Usage); err != nil {
			return
		}
		accOpts.Usage = append(usage, accOpts.Usage...)
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		var prfIgnFltrs []*DynamicBoolOpt
		prfIgnFltrs, err = IfaceToBoolDynamicOpts(jsnCfg.ProfileIgnoreFilters)
		accOpts.ProfileIgnoreFilters = append(prfIgnFltrs, accOpts.ProfileIgnoreFilters...)
	}
	return
}

func (acS *AccountSCfg) loadFromJSONCfg(jsnCfg *AccountSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		acS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		acS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			acS.Conns[connType] = opts
		}
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
	if jsnCfg.Max_iterations != nil {
		acS.MaxIterations = *jsnCfg.Max_iterations
	}
	if jsnCfg.Max_usage != nil {
		if acS.MaxUsage, err = utils.NewDecimalFromUsage(*jsnCfg.Max_usage); err != nil {
			return err
		}
	}
	if jsnCfg.Opts != nil {
		err = acS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (acS AccountSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaProfileIDs:           acS.Opts.ProfileIDs,
		utils.MetaUsage:                acS.Opts.Usage,
		utils.MetaProfileIgnoreFilters: acS.Opts.ProfileIgnoreFilters,
	}
	mp := map[string]any{
		utils.EnabledCfg:        acS.Enabled,
		utils.IndexedSelectsCfg: acS.IndexedSelects,
		utils.NestedFieldsCfg:   acS.NestedFields,
		utils.MaxIterations:     acS.MaxIterations,
		utils.ConnsCfg:          stripConns(acS.Conns),
		utils.OptsCfg:           opts,
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
	if acS.MaxUsage != nil {
		mp[utils.MaxUsage] = acS.MaxUsage.String()
	}
	return mp
}

func (accOpts *AccountsOpts) Clone() *AccountsOpts {
	var accIDs []*DynamicStringSliceOpt
	if accOpts.ProfileIDs != nil {
		accIDs = CloneDynamicStringSliceOpt(accOpts.ProfileIDs)
	}
	var usage []*DynamicDecimalOpt
	if accOpts.Usage != nil {
		usage = CloneDynamicDecimalOpt(accOpts.Usage)
	}
	var profileIgnoreFilters []*DynamicBoolOpt
	if accOpts.ProfileIgnoreFilters != nil {
		profileIgnoreFilters = CloneDynamicBoolOpt(accOpts.ProfileIgnoreFilters)
	}
	return &AccountsOpts{
		ProfileIDs:           accIDs,
		Usage:                usage,
		ProfileIgnoreFilters: profileIgnoreFilters,
	}
}
func (AccountSCfg) SName() string             { return AccountSJSON }
func (acS AccountSCfg) CloneSection() Section { return acS.Clone() }

// Clone returns a deep copy of AccountSCfg
func (acS AccountSCfg) Clone() (cln *AccountSCfg) {
	cln = &AccountSCfg{
		Enabled:        acS.Enabled,
		Conns:          CloneConnsMap(acS.Conns),
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
		MaxIterations:  acS.MaxIterations,
		MaxUsage:       acS.MaxUsage,
		Opts:           acS.Opts.Clone(),
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
	return
}

type AccountsOptsJson struct {
	ProfileIDs           []*DynamicStringSliceOpt `json:"*profileIDs"`
	Usage                []*DynamicInterfaceOpt   `json:"*usage"`
	ProfileIgnoreFilters []*DynamicInterfaceOpt   `json:"*profileIgnoreFilters"`
}

// Account service config section
type AccountSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	Conns                    map[string][]*DynamicConns `json:"conns,omitempty"`
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Exists_indexed_fields    *[]string
	Notexists_indexed_fields *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Max_iterations           *int
	Max_usage                *string
	Opts                     *AccountsOptsJson
}

func diffAccountsOptsJsonCfg(d *AccountsOptsJson, v1, v2 *AccountsOpts) *AccountsOptsJson {
	if d == nil {
		d = new(AccountsOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ProfileIDs, v2.ProfileIDs) {
		d.ProfileIDs = v2.ProfileIDs
	}
	if !DynamicDecimalOptEqual(v1.Usage, v2.Usage) {
		d.Usage = DecimalToIfaceDynamicOpts(v2.Usage)
	}
	if !DynamicBoolOptEqual(v1.ProfileIgnoreFilters, v2.ProfileIgnoreFilters) {
		d.ProfileIgnoreFilters = BoolToIfaceDynamicOpts(v2.ProfileIgnoreFilters)
	}
	return d
}

func diffAccountSJsonCfg(d *AccountSJsonCfg, v1, v2 *AccountSCfg) *AccountSJsonCfg {
	if d == nil {
		d = new(AccountSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
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
	if v1.MaxIterations != v2.MaxIterations {
		d.Max_iterations = utils.IntPointer(v2.MaxIterations)
	}
	if v2.MaxUsage != nil {
		if v1.MaxUsage == nil ||
			v1.MaxUsage.Cmp(v2.MaxUsage.Big) != 0 {
			d.Max_usage = utils.StringPointer(v2.MaxUsage.String())
		}
	} else {
		d.Max_usage = nil
	}
	d.Opts = diffAccountsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
