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

// ThresholdSCfg the threshold config section
type ThresholdSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StoreInterval       time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
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
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		t.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		t.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		t.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		t.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (t *ThresholdSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        t.Enabled,
		utils.IndexedSelectsCfg: t.IndexedSelects,
		utils.NestedFieldsCfg:   t.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
	}
	if t.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = t.StoreInterval.String()
	}

	if t.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*t.StringIndexedFields))
		for i, item := range *t.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if t.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*t.PrefixIndexedFields))
		for i, item := range *t.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if t.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*t.SuffixIndexedFields))
		for i, item := range *t.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	return
}

// Clone returns a deep copy of ThresholdSCfg
func (t ThresholdSCfg) Clone() (cln *ThresholdSCfg) {
	cln = &ThresholdSCfg{
		Enabled:        t.Enabled,
		IndexedSelects: t.IndexedSelects,
		StoreInterval:  t.StoreInterval,
		NestedFields:   t.NestedFields,
	}

	if t.StringIndexedFields != nil {
		idx := make([]string, len(*t.StringIndexedFields))
		for i, dx := range *t.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if t.PrefixIndexedFields != nil {
		idx := make([]string, len(*t.PrefixIndexedFields))
		for i, dx := range *t.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if t.SuffixIndexedFields != nil {
		idx := make([]string, len(*t.SuffixIndexedFields))
		for i, dx := range *t.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
