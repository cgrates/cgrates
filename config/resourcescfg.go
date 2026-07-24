// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	if jsnCfg.Nested_fields != nil {
		rlcfg.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (rlcfg *ResourceSConfig) AsMapInterface() map[string]any {
	thresholdSConns := make([]string, len(rlcfg.ThresholdSConns))
	for i, item := range rlcfg.ThresholdSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
		if item == buf {
			thresholdSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds, utils.EmptyString)
		} else {
			thresholdSConns[i] = item
		}
	}
	stringIndexedFields := []string{}
	if rlcfg.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*rlcfg.StringIndexedFields))
		for i, item := range *rlcfg.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if rlcfg.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*rlcfg.PrefixIndexedFields))
		for i, item := range *rlcfg.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	var storeInterval string = ""
	if rlcfg.StoreInterval != 0 {
		storeInterval = rlcfg.StoreInterval.String()
	}
	return map[string]any{
		utils.EnabledCfg:             rlcfg.Enabled,
		utils.IndexedSelectsCfg:      rlcfg.IndexedSelects,
		utils.ThresholdSConnsCfg:     thresholdSConns,
		utils.StoreIntervalCfg:       storeInterval,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.NestedFieldsCfg:        rlcfg.NestedFields,
	}

}
