// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ThresholdSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StoreInterval       time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	NestedFields        bool
}

func (t *ThresholdSCfg) loadFromJsonCfg(jsnCfg *ThresholdSJsonCfg) (err error) {
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
	if jsnCfg.Nested_fields != nil {
		t.NestedFields = *jsnCfg.Nested_fields
	}
	return nil
}

func (t *ThresholdSCfg) AsMapInterface() map[string]any {
	var storeInterval string = ""
	if t.StoreInterval != 0 {
		storeInterval = t.StoreInterval.String()
	}
	stringIndexedFields := []string{}
	if t.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*t.StringIndexedFields))
		for i, item := range *t.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if t.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*t.PrefixIndexedFields))
		for i, item := range *t.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	return map[string]any{
		utils.EnabledCfg:             t.Enabled,
		utils.IndexedSelectsCfg:      t.IndexedSelects,
		utils.StoreIntervalCfg:       storeInterval,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.NestedFieldsCfg:        t.NestedFields,
	}
}
