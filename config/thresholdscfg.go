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
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ThresholdsOpts struct {
	ThresholdIDs []string
}

// ThresholdSCfg the threshold config section
type ThresholdSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StoreInterval       time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	ActionSConns        []string // connections towards ActionS
	Opts                *ThresholdsOpts
}

func (thdOpts *ThresholdsOpts) loadFromJSONCfg(jsnCfg *ThresholdsOptsJson) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ThresholdIDs != nil {
		thdOpts.ThresholdIDs = *jsnCfg.ThresholdIDs
	}

	return nil
}

func (t *ThresholdSCfg) loadFromJSONCfg(jsnCfg *ThresholdSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		t.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		t.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Store_interval != nil {
		if t.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		t.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		t.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		t.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		t.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Actions_conns != nil {
		t.ActionSConns = updateInternalConns(*jsnCfg.Actions_conns, utils.MetaActions)
	}
	if jsnCfg.Opts != nil {
		t.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (t *ThresholdSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.MetaThresholdIDsCfg: t.Opts.ThresholdIDs,
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        t.Enabled,
		utils.IndexedSelectsCfg: t.IndexedSelects,
		utils.NestedFieldsCfg:   t.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.OptsCfg:           opts,
	}
	if t.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = t.StoreInterval.String()
	}

	if t.StringIndexedFields != nil {
		initialMP[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*t.StringIndexedFields)
	}
	if t.PrefixIndexedFields != nil {
		initialMP[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*t.PrefixIndexedFields)
	}
	if t.SuffixIndexedFields != nil {
		initialMP[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*t.SuffixIndexedFields)
	}
	if t.ActionSConns != nil {
		initialMP[utils.ActionSConnsCfg] = getInternalJSONConns(t.ActionSConns)
	}
	return
}

func (thdOpts *ThresholdsOpts) Clone() *ThresholdsOpts {
	var thdIDs []string
	if thdOpts.ThresholdIDs != nil {
		thdIDs = utils.CloneStringSlice(thdOpts.ThresholdIDs)
	}
	return &ThresholdsOpts{
		ThresholdIDs: thdIDs,
	}
}

// Clone returns a deep copy of ThresholdSCfg
func (t ThresholdSCfg) Clone() (cln *ThresholdSCfg) {
	cln = &ThresholdSCfg{
		Enabled:        t.Enabled,
		IndexedSelects: t.IndexedSelects,
		StoreInterval:  t.StoreInterval,
		NestedFields:   t.NestedFields,
	}

	if t.Opts != nil {
		cln.Opts = t.Opts.Clone()
	}
	if t.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*t.StringIndexedFields))
	}
	if t.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*t.PrefixIndexedFields))
	}
	if t.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*t.SuffixIndexedFields))
	}
	if t.ActionSConns != nil {
		cln.ActionSConns = utils.CloneStringSlice(t.ActionSConns)
	}
	return
}

type ThresholdsOptsJson struct {
	ThresholdIDs *[]string `json:"*thresholdIDs"`
}

// Threshold service config section
type ThresholdSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Actions_conns         *[]string
	Opts                  *ThresholdsOptsJson
}

func diffThresholdsOptsJsonCfg(d *ThresholdsOptsJson, v1, v2 *ThresholdsOpts) *ThresholdsOptsJson {
	if d == nil {
		d = new(ThresholdsOptsJson)
	}
	if !utils.SliceStringEqual(v1.ThresholdIDs, v2.ThresholdIDs) {
		d.ThresholdIDs = utils.SliceStringPointer(v2.ThresholdIDs)
	}
	return d
}

func diffThresholdSJsonCfg(d *ThresholdSJsonCfg, v1, v2 *ThresholdSCfg) *ThresholdSJsonCfg {
	if d == nil {
		d = new(ThresholdSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	d.Opts = diffThresholdsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
