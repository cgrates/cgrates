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

import "github.com/cgrates/cgrates/utils"

type AccountsOpts struct {
	AccountIDs []string
}

// AccountSCfg is the configuration of ActionS
type AccountSCfg struct {
	Enabled             bool
	AttributeSConns     []string
	RateSConns          []string
	ThresholdSConns     []string
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	MaxIterations       int
	MaxUsage            *utils.Decimal
	Opts                *AccountsOpts
}

func (accOpts *AccountsOpts) loadFromJSONCfg(jsnCfg *AccountsOptsJson) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.AccountIDs != nil {
		accOpts.AccountIDs = *jsnCfg.AccountIDs
	}

	return nil
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
	if jsnCfg.Attributes_conns != nil {
		acS.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.Rates_conns != nil {
		acS.RateSConns = updateInternalConns(*jsnCfg.Rates_conns, utils.MetaRateS)
	}
	if jsnCfg.Thresholds_conns != nil {
		acS.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
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
	if jsnCfg.Max_iterations != nil {
		acS.MaxIterations = *jsnCfg.Max_iterations
	}
	if jsnCfg.Max_usage != nil {
		if acS.MaxUsage, err = utils.NewDecimalFromUsage(*jsnCfg.Max_usage); err != nil {
			return err
		}
	}
	if jsnCfg.Opts != nil {
		acS.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (acS *AccountSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.MetaAccountIDsCfg: acS.Opts.AccountIDs,
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        acS.Enabled,
		utils.IndexedSelectsCfg: acS.IndexedSelects,
		utils.NestedFieldsCfg:   acS.NestedFields,
		utils.MaxIterations:     acS.MaxIterations,
		utils.OptsCfg:           opts,
	}
	if acS.AttributeSConns != nil {
		initialMP[utils.AttributeSConnsCfg] = getInternalJSONConns(acS.AttributeSConns)
	}
	if acS.RateSConns != nil {
		initialMP[utils.RateSConnsCfg] = getInternalJSONConns(acS.RateSConns)
	}
	if acS.ThresholdSConns != nil {
		initialMP[utils.ThresholdSConnsCfg] = getInternalJSONConns(acS.ThresholdSConns)
	}
	if acS.StringIndexedFields != nil {
		initialMP[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*acS.StringIndexedFields)
	}
	if acS.PrefixIndexedFields != nil {
		initialMP[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*acS.PrefixIndexedFields)
	}
	if acS.SuffixIndexedFields != nil {
		initialMP[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*acS.SuffixIndexedFields)
	}
	if acS.MaxUsage != nil {
		initialMP[utils.MaxUsage] = acS.MaxUsage.String()
	}
	return
}

func (accOpts *AccountsOpts) Clone() *AccountsOpts {
	var accIDs []string
	if accOpts.AccountIDs != nil {
		accIDs = utils.CloneStringSlice(accOpts.AccountIDs)
	}
	return &AccountsOpts{
		AccountIDs: accIDs,
	}
}

// Clone returns a deep copy of AccountSCfg
func (acS AccountSCfg) Clone() (cln *AccountSCfg) {
	cln = &AccountSCfg{
		Enabled:        acS.Enabled,
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
		MaxIterations:  acS.MaxIterations,
		MaxUsage:       acS.MaxUsage,
		Opts:           acS.Opts.Clone(),
	}
	if acS.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(acS.AttributeSConns)
	}
	if acS.RateSConns != nil {
		cln.RateSConns = utils.CloneStringSlice(acS.RateSConns)
	}
	if acS.ThresholdSConns != nil {
		cln.ThresholdSConns = utils.CloneStringSlice(acS.ThresholdSConns)
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
	return
}

type AccountsOptsJson struct {
	AccountIDs *[]string `json:"*accountIDs"`
}

// Account service config section
type AccountSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Attributes_conns      *[]string
	Rates_conns           *[]string
	Thresholds_conns      *[]string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Max_iterations        *int
	Max_usage             *string
	Opts                  *AccountsOptsJson
}

func diffAccountsOptsJsonCfg(d *AccountsOptsJson, v1, v2 *AccountsOpts) *AccountsOptsJson {
	if d == nil {
		d = new(AccountsOptsJson)
	}
	if !utils.SliceStringEqual(v1.AccountIDs, v2.AccountIDs) {
		d.AccountIDs = utils.SliceStringPointer(v2.AccountIDs)
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
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !utils.SliceStringEqual(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
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
