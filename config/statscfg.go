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

type StatSCfg struct {
	Enabled         bool
	StoreInterval   time.Duration // Dump regularly from cache into dataDB
	ThresholdSConns []*HaPoolConfig
	IndexedFields   []string
}

func (st *StatSCfg) loadFromJsonCfg(jsnCfg *StatServJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		st.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Store_interval != nil {
		if st.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		st.ThresholdSConns = make([]*HaPoolConfig, len(*jsnCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCfg.Thresholds_conns {
			st.ThresholdSConns[idx] = NewDfltHaPoolConfig()
			st.ThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Indexed_fields != nil {
		st.IndexedFields = make([]string, len(*jsnCfg.Indexed_fields))
		for i, fID := range *jsnCfg.Indexed_fields {
			st.IndexedFields[i] = fID
		}
	}
	return nil
}
