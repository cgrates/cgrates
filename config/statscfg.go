// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	if jsnCfg.Nested_fields != nil {
		st.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (st *StatSCfg) AsMapInterface() map[string]any {
	var storeInterval string = ""
	if st.StoreInterval != 0 {
		storeInterval = st.StoreInterval.String()
	}
	stringIndexedFields := []string{}
	if st.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*st.StringIndexedFields))
		for i, item := range *st.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if st.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*st.PrefixIndexedFields))
		for i, item := range *st.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	thresholdSConns := make([]string, len(st.ThresholdSConns))
	for i, item := range st.ThresholdSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
		if item == buf {
			thresholdSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds, utils.EmptyString)
		} else {
			thresholdSConns[i] = item
		}
	}

	return map[string]any{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreIntervalCfg:          storeInterval,
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.ThresholdSConnsCfg:        thresholdSConns,
		utils.StringIndexedFieldsCfg:    stringIndexedFields,
		utils.PrefixIndexedFieldsCfg:    prefixIndexedFields,
		utils.NestedFieldsCfg:           st.NestedFields,
	}

}
