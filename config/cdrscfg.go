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
	"github.com/cgrates/cgrates/utils"
)

type CdrsCfg struct {
	CDRSEnabled          bool              // Enable CDR Server service
	CDRSExtraFields      []*utils.RSRField // Extra fields to store in CDRs
	CDRSStoreCdrs        bool              // store cdrs in storDb
	CDRSSMCostRetries    int
	CDRSChargerSConns    []*HaPoolConfig
	CDRSRaterConns       []*HaPoolConfig // address where to reach the Rater for cost calculation: <""|internal|x.y.z.y:1234>
	CDRSPubSubSConns     []*HaPoolConfig // address where to reach the pubsub service: <""|internal|x.y.z.y:1234>
	CDRSAttributeSConns  []*HaPoolConfig // address where to reach the users service: <""|internal|x.y.z.y:1234>
	CDRSUserSConns       []*HaPoolConfig // address where to reach the users service: <""|internal|x.y.z.y:1234>
	CDRSAliaseSConns     []*HaPoolConfig // address where to reach the aliases service: <""|internal|x.y.z.y:1234>
	CDRSThresholdSConns  []*HaPoolConfig // address where to reach the thresholds service
	CDRSStatSConns       []*HaPoolConfig
	CDRSOnlineCDRExports []string // list of CDRE templates to use for real-time CDR exports
}

//loadFromJsonCfg loads Cdrs config from JsonCfg
func (cdrscfg *CdrsCfg) loadFromJsonCfg(jsnCdrsCfg *CdrsJsonCfg) (err error) {
	if jsnCdrsCfg == nil {
		return nil
	}
	if jsnCdrsCfg.Enabled != nil {
		cdrscfg.CDRSEnabled = *jsnCdrsCfg.Enabled
	}
	if jsnCdrsCfg.Extra_fields != nil {
		if cdrscfg.CDRSExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
			return err
		}
	}
	if jsnCdrsCfg.Store_cdrs != nil {
		cdrscfg.CDRSStoreCdrs = *jsnCdrsCfg.Store_cdrs
	}
	if jsnCdrsCfg.Sessions_cost_retries != nil {
		cdrscfg.CDRSSMCostRetries = *jsnCdrsCfg.Sessions_cost_retries
	}
	if jsnCdrsCfg.Chargers_conns != nil {
		cdrscfg.CDRSChargerSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Chargers_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Chargers_conns {
			cdrscfg.CDRSChargerSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSChargerSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Rals_conns != nil {
		cdrscfg.CDRSRaterConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Rals_conns {
			cdrscfg.CDRSRaterConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSRaterConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Pubsubs_conns != nil {
		cdrscfg.CDRSPubSubSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Pubsubs_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Pubsubs_conns {
			cdrscfg.CDRSPubSubSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSPubSubSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.CDRSAttributeSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Attributes_conns {
			cdrscfg.CDRSAttributeSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSAttributeSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Users_conns != nil {
		cdrscfg.CDRSUserSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Users_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Users_conns {
			cdrscfg.CDRSUserSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSUserSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Aliases_conns != nil {
		cdrscfg.CDRSAliaseSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Aliases_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Aliases_conns {
			cdrscfg.CDRSAliaseSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSAliaseSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.CDRSThresholdSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Thresholds_conns {
			cdrscfg.CDRSThresholdSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.CDRSStatSConns = make([]*HaPoolConfig, len(*jsnCdrsCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Stats_conns {
			cdrscfg.CDRSStatSConns[idx] = NewDfltHaPoolConfig()
			cdrscfg.CDRSStatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		for _, expProfile := range *jsnCdrsCfg.Online_cdr_exports {
			cdrscfg.CDRSOnlineCDRExports = append(cdrscfg.CDRSOnlineCDRExports, expProfile)
		}
	}

	return nil
}
