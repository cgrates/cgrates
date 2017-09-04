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
	Enabled       bool
	StatSConns    []*HaPoolConfig // Connections towards StatS
	StoreInterval time.Duration   // Dump regularly from cache into dataDB
	ShortCache    *CacheParamConfig
}

func (rlcfg *ResourceSConfig) loadFromJsonCfg(jsnCfg *ResourceSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		rlcfg.StatSConns = make([]*HaPoolConfig, len(*jsnCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Stats_conns {
			rlcfg.StatSConns[idx] = NewDfltHaPoolConfig()
			rlcfg.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Store_interval != nil {
		if rlcfg.StoreInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.Short_cache != nil {
		rlcfg.ShortCache = new(CacheParamConfig)
		if err = rlcfg.ShortCache.loadFromJsonCfg(jsnCfg.Short_cache); err != nil {
			return
		}
	}
	return nil
}
