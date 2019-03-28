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
	CDRSChargerSConns    []*RemoteHost
	CDRSRaterConns       []*RemoteHost // address where to reach the Rater for cost calculation: <""|internal|x.y.z.y:1234>
	CDRSAttributeSConns  []*RemoteHost // address where to reach the users service: <""|internal|x.y.z.y:1234>
	CDRSThresholdSConns  []*RemoteHost // address where to reach the thresholds service
	CDRSStatSConns       []*RemoteHost
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
	if jsnCdrsCfg.Session_cost_retries != nil {
		cdrscfg.CDRSSMCostRetries = *jsnCdrsCfg.Session_cost_retries
	}
	if jsnCdrsCfg.Chargers_conns != nil {
		cdrscfg.CDRSChargerSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Chargers_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Chargers_conns {
			cdrscfg.CDRSChargerSConns[idx] = NewDfltRemoteHost()
			cdrscfg.CDRSChargerSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Rals_conns != nil {
		cdrscfg.CDRSRaterConns = make([]*RemoteHost, len(*jsnCdrsCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Rals_conns {
			cdrscfg.CDRSRaterConns[idx] = NewDfltRemoteHost()
			cdrscfg.CDRSRaterConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.CDRSAttributeSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Attributes_conns {
			cdrscfg.CDRSAttributeSConns[idx] = NewDfltRemoteHost()
			cdrscfg.CDRSAttributeSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.CDRSThresholdSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Thresholds_conns {
			cdrscfg.CDRSThresholdSConns[idx] = NewDfltRemoteHost()
			cdrscfg.CDRSThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.CDRSStatSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Stats_conns {
			cdrscfg.CDRSStatSConns[idx] = NewDfltRemoteHost()
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
