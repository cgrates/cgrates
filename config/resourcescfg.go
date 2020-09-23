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

type ResourceSConfig struct {
	Enabled             bool
	IndexedSelects      bool
	ThresholdSConns     []string
	StoreInterval       time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
}

func (rlcfg *ResourceSConfig) loadFromJsonCfg(jsnCfg *ResourceSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		rlcfg.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Thresholds_conns != nil {
		rlcfg.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				rlcfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				rlcfg.ThresholdSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Store_interval != nil {
		if rlcfg.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		rlcfg.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		rlcfg.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		rlcfg.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		rlcfg.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (rlcfg *ResourceSConfig) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        rlcfg.Enabled,
		utils.IndexedSelectsCfg: rlcfg.IndexedSelects,
		utils.NestedFieldsCfg:   rlcfg.NestedFields,
	}
	if rlcfg.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(rlcfg.ThresholdSConns))
		for i, item := range rlcfg.ThresholdSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds, utils.EmptyString)
			} else {
				thresholdSConns[i] = item
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if rlcfg.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*rlcfg.StringIndexedFields))
		for i, item := range *rlcfg.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if rlcfg.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*rlcfg.PrefixIndexedFields))
		for i, item := range *rlcfg.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if rlcfg.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*rlcfg.SuffixIndexedFields))
		for i, item := range *rlcfg.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if rlcfg.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = rlcfg.StoreInterval.String()
	} else {
		initialMP[utils.StoreIntervalCfg] = utils.EmptyString
	}
	return
}
