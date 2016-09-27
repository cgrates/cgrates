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

type ResourceLimiterConfig struct {
	Enabled           bool
	CDRStatConns      []*HaPoolConfig // Connections towards CDRStatS
	CacheDumpInterval time.Duration   // Dump regularly from cache into dataDB
	UsageTTL          time.Duration   // Auto-Expire usage units older than this duration
}

func (rlcfg *ResourceLimiterConfig) loadFromJsonCfg(jsnCfg *ResourceLimiterServJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrstats_conns != nil {
		rlcfg.CDRStatConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrstats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrstats_conns {
			rlcfg.CDRStatConns[idx] = NewDfltHaPoolConfig()
			rlcfg.CDRStatConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cache_dump_interval != nil {
		if rlcfg.CacheDumpInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Cache_dump_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Usage_ttl != nil {
		if rlcfg.UsageTTL, err = utils.ParseDurationWithSecs(*jsnCfg.Usage_ttl); err != nil {
			return err
		}
	}
	return nil
}
