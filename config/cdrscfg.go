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
	Enabled          bool              // Enable CDR Server service
	ExtraFields      []*utils.RSRField // Extra fields to store in CDRs
	StoreCdrs        bool              // store cdrs in storDb
	SMCostRetries    int
	ChargerSConns    []*RemoteHost
	RaterConns       []*RemoteHost // address where to reach the Rater for cost calculation: <""|internal|x.y.z.y:1234>
	AttributeSConns  []*RemoteHost // address where to reach the users service: <""|internal|x.y.z.y:1234>
	ThresholdSConns  []*RemoteHost // address where to reach the thresholds service
	StatSConns       []*RemoteHost
	OnlineCDRExports []string // list of CDRE templates to use for real-time CDR exports
}

//loadFromJsonCfg loads Cdrs config from JsonCfg
func (cdrscfg *CdrsCfg) loadFromJsonCfg(jsnCdrsCfg *CdrsJsonCfg) (err error) {
	if jsnCdrsCfg == nil {
		return nil
	}
	if jsnCdrsCfg.Enabled != nil {
		cdrscfg.Enabled = *jsnCdrsCfg.Enabled
	}
	if jsnCdrsCfg.Extra_fields != nil {
		if cdrscfg.ExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
			return err
		}
	}
	if jsnCdrsCfg.Store_cdrs != nil {
		cdrscfg.StoreCdrs = *jsnCdrsCfg.Store_cdrs
	}
	if jsnCdrsCfg.Session_cost_retries != nil {
		cdrscfg.SMCostRetries = *jsnCdrsCfg.Session_cost_retries
	}
	if jsnCdrsCfg.Chargers_conns != nil {
		cdrscfg.ChargerSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Chargers_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Chargers_conns {
			cdrscfg.ChargerSConns[idx] = NewDfltRemoteHost()
			cdrscfg.ChargerSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Rals_conns != nil {
		cdrscfg.RaterConns = make([]*RemoteHost, len(*jsnCdrsCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Rals_conns {
			cdrscfg.RaterConns[idx] = NewDfltRemoteHost()
			cdrscfg.RaterConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.AttributeSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Attributes_conns {
			cdrscfg.AttributeSConns[idx] = NewDfltRemoteHost()
			cdrscfg.AttributeSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.ThresholdSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Thresholds_conns {
			cdrscfg.ThresholdSConns[idx] = NewDfltRemoteHost()
			cdrscfg.ThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.StatSConns = make([]*RemoteHost, len(*jsnCdrsCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCdrsCfg.Stats_conns {
			cdrscfg.StatSConns[idx] = NewDfltRemoteHost()
			cdrscfg.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		for _, expProfile := range *jsnCdrsCfg.Online_cdr_exports {
			cdrscfg.OnlineCDRExports = append(cdrscfg.OnlineCDRExports, expProfile)
		}
	}

	return nil
}
