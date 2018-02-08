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

type ResourceSConfig struct {
	Enabled             bool
	ThresholdSConns     []*HaPoolConfig // Connections towards StatS
	StoreInterval       time.Duration   // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
}

func (rlcfg *ResourceSConfig) loadFromJsonCfg(jsnCfg *ResourceSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Thresholds_conns != nil {
		rlcfg.ThresholdSConns = make([]*HaPoolConfig, len(*jsnCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCfg.Thresholds_conns {
			rlcfg.ThresholdSConns[idx] = NewDfltHaPoolConfig()
			rlcfg.ThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
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
	return nil
}
