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

// RateSCfg the rates config section
type RateSCfg struct {
	Enabled                 bool
	IndexedSelects          bool
	StringIndexedFields     *[]string
	PrefixIndexedFields     *[]string
	SuffixIndexedFields     *[]string
	NestedFields            bool
	RateIndexedSelects      bool
	RateStringIndexedFields *[]string
	RatePrefixIndexedFields *[]string
	RateSuffixIndexedFields *[]string
	RateNestedFields        bool
}

func (rCfg *RateSCfg) loadFromJSONCfg(jsnCfg *RateSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		rCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		rCfg.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		rCfg.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		rCfg.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		rCfg.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		rCfg.NestedFields = *jsnCfg.Nested_fields
	}

	if jsnCfg.Rate_indexed_selects != nil {
		rCfg.RateIndexedSelects = *jsnCfg.Rate_indexed_selects
	}
	if jsnCfg.Rate_string_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Rate_string_indexed_fields))
		for i, fID := range *jsnCfg.Rate_string_indexed_fields {
			sif[i] = fID
		}
		rCfg.RateStringIndexedFields = &sif
	}
	if jsnCfg.Rate_prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Rate_prefix_indexed_fields))
		for i, fID := range *jsnCfg.Rate_prefix_indexed_fields {
			pif[i] = fID
		}
		rCfg.RatePrefixIndexedFields = &pif
	}
	if jsnCfg.Rate_suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Rate_suffix_indexed_fields))
		for i, fID := range *jsnCfg.Rate_suffix_indexed_fields {
			sif[i] = fID
		}
		rCfg.RateSuffixIndexedFields = &sif
	}
	if jsnCfg.Rate_nested_fields != nil {
		rCfg.RateNestedFields = *jsnCfg.Rate_nested_fields
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rCfg *RateSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:            rCfg.Enabled,
		utils.IndexedSelectsCfg:     rCfg.IndexedSelects,
		utils.NestedFieldsCfg:       rCfg.NestedFields,
		utils.RateIndexedSelectsCfg: rCfg.RateIndexedSelects,
		utils.RateNestedFieldsCfg:   rCfg.RateNestedFields,
	}
	if rCfg.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*rCfg.StringIndexedFields))
		for i, item := range *rCfg.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if rCfg.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*rCfg.PrefixIndexedFields))
		for i, item := range *rCfg.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if rCfg.SuffixIndexedFields != nil {
		sufixIndexedFields := make([]string, len(*rCfg.SuffixIndexedFields))
		for i, item := range *rCfg.SuffixIndexedFields {
			sufixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = sufixIndexedFields
	}
	if rCfg.RateStringIndexedFields != nil {
		rateStringIndexedFields := make([]string, len(*rCfg.RateStringIndexedFields))
		for i, item := range *rCfg.RateStringIndexedFields {
			rateStringIndexedFields[i] = item
		}
		initialMP[utils.RateStringIndexedFieldsCfg] = rateStringIndexedFields
	}
	if rCfg.RatePrefixIndexedFields != nil {
		ratePrefixIndexedFields := make([]string, len(*rCfg.RatePrefixIndexedFields))
		for i, item := range *rCfg.RatePrefixIndexedFields {
			ratePrefixIndexedFields[i] = item
		}
		initialMP[utils.RatePrefixIndexedFieldsCfg] = ratePrefixIndexedFields
	}
	if rCfg.RateSuffixIndexedFields != nil {
		rateSufixIndexedFields := make([]string, len(*rCfg.RateSuffixIndexedFields))
		for i, item := range *rCfg.RateSuffixIndexedFields {
			rateSufixIndexedFields[i] = item
		}
		initialMP[utils.RateSuffixIndexedFieldsCfg] = rateSufixIndexedFields
	}
	return
}

// Clone returns a deep copy of RateSCfg
func (rCfg RateSCfg) Clone() (cln *RateSCfg) {
	cln = &RateSCfg{
		Enabled:            rCfg.Enabled,
		IndexedSelects:     rCfg.IndexedSelects,
		NestedFields:       rCfg.NestedFields,
		RateIndexedSelects: rCfg.RateIndexedSelects,
		RateNestedFields:   rCfg.RateNestedFields,
	}
	if rCfg.StringIndexedFields != nil {
		idx := make([]string, len(*rCfg.StringIndexedFields))
		for i, dx := range *rCfg.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if rCfg.PrefixIndexedFields != nil {
		idx := make([]string, len(*rCfg.PrefixIndexedFields))
		for i, dx := range *rCfg.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if rCfg.SuffixIndexedFields != nil {
		idx := make([]string, len(*rCfg.SuffixIndexedFields))
		for i, dx := range *rCfg.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}

	if rCfg.RateStringIndexedFields != nil {
		idx := make([]string, len(*rCfg.RateStringIndexedFields))
		for i, dx := range *rCfg.RateStringIndexedFields {
			idx[i] = dx
		}
		cln.RateStringIndexedFields = &idx
	}
	if rCfg.RatePrefixIndexedFields != nil {
		idx := make([]string, len(*rCfg.RatePrefixIndexedFields))
		for i, dx := range *rCfg.RatePrefixIndexedFields {
			idx[i] = dx
		}
		cln.RatePrefixIndexedFields = &idx
	}
	if rCfg.RateSuffixIndexedFields != nil {
		idx := make([]string, len(*rCfg.RateSuffixIndexedFields))
		for i, dx := range *rCfg.RateSuffixIndexedFields {
			idx[i] = dx
		}
		cln.RateSuffixIndexedFields = &idx
	}
	return
}
