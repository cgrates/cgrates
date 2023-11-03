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
	"slices"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ThresholdsOpts struct {
	ProfileIDs           []string
	ProfileIgnoreFilters bool
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
	Opts                *ThresholdsOpts
}

func (thdOpts *ThresholdsOpts) loadFromJSONCfg(jsnCfg *ThresholdsOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		thdOpts.ProfileIDs = *jsnCfg.ProfileIDs
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		thdOpts.ProfileIgnoreFilters = *jsnCfg.ProfileIgnoreFilters
	}
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
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		copy(sif, *jsnCfg.String_indexed_fields)
		t.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		copy(pif, *jsnCfg.Prefix_indexed_fields)
		t.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		copy(sif, *jsnCfg.Suffix_indexed_fields)
		t.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		t.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		t.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (t *ThresholdSCfg) AsMapInterface() (initialMP map[string]any) {
	opts := map[string]any{
		utils.MetaProfileIDs:              t.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFiltersCfg: t.Opts.ProfileIgnoreFilters,
	}
	initialMP = map[string]any{
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
		stringIndexedFields := make([]string, len(*t.StringIndexedFields))
		copy(stringIndexedFields, *t.StringIndexedFields)
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if t.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*t.PrefixIndexedFields))
		copy(prefixIndexedFields, *t.PrefixIndexedFields)
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if t.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*t.SuffixIndexedFields))
		copy(suffixIndexedFields, *t.SuffixIndexedFields)
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	return
}

func (thdOpts *ThresholdsOpts) Clone() *ThresholdsOpts {
	return &ThresholdsOpts{
		ProfileIDs:           slices.Clone(thdOpts.ProfileIDs),
		ProfileIgnoreFilters: thdOpts.ProfileIgnoreFilters,
	}
}

// Clone returns a deep copy of ThresholdSCfg
func (t ThresholdSCfg) Clone() (cln *ThresholdSCfg) {
	cln = &ThresholdSCfg{
		Enabled:        t.Enabled,
		IndexedSelects: t.IndexedSelects,
		StoreInterval:  t.StoreInterval,
		NestedFields:   t.NestedFields,
		Opts:           t.Opts.Clone(),
	}

	if t.StringIndexedFields != nil {
		idx := make([]string, len(*t.StringIndexedFields))
		copy(idx, *t.StringIndexedFields)
		cln.StringIndexedFields = &idx
	}
	if t.PrefixIndexedFields != nil {
		idx := make([]string, len(*t.PrefixIndexedFields))
		copy(idx, *t.PrefixIndexedFields)
		cln.PrefixIndexedFields = &idx
	}
	if t.SuffixIndexedFields != nil {
		idx := make([]string, len(*t.SuffixIndexedFields))
		copy(idx, *t.SuffixIndexedFields)
		cln.SuffixIndexedFields = &idx
	}
	return
}
