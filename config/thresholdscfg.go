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

func (t *ThresholdSCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.EnabledCfg:             t.Enabled,
		utils.IndexedSelectsCfg:      t.IndexedSelects,
		utils.StoreIntervalCfg:       t.StoreInterval,
		utils.StringIndexedFieldsCfg: t.StringIndexedFields,
		utils.PrefixIndexedFieldsCfg: t.PrefixIndexedFields,
		utils.NestedFieldsCfg:        t.NestedFields,
	}
}
