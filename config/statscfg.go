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
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type StatSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StoreInterval          time.Duration // Dump regularly from cache into dataDB
	StoreUncompressedLimit int
	ThresholdSConns        []string
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	NestedFields           bool
}

func (st *StatSCfg) loadFromJsonCfg(jsnCfg *StatServJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		st.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		st.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Store_interval != nil {
		if st.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Store_uncompressed_limit != nil {
		st.StoreUncompressedLimit = *jsnCfg.Store_uncompressed_limit
	}
	if jsnCfg.Thresholds_conns != nil {
		st.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				st.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				st.ThresholdSConns[idx] = conn
			}
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		st.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		st.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		st.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		st.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (st *StatSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.NestedFieldsCfg:           st.NestedFields,
	}
	if st.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = st.StoreInterval.String()
	} else {
		initialMP[utils.StoreIntervalCfg] = utils.EmptyString
	}
	if st.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*st.StringIndexedFields))
		for i, item := range *st.StringIndexedFields {
			stringIndexedFields[i] = item
		}

		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if st.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*st.PrefixIndexedFields))
		for i, item := range *st.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}

		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if st.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*st.SuffixIndexedFields))
		for i, item := range *st.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields

	}
	if st.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(st.ThresholdSConns))
		for i, item := range st.ThresholdSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds, utils.EmptyString)
			} else {
				thresholdSConns[i] = item
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	return
}
