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
	ChargerSConns    []string
	RaterConns       []string
	AttributeSConns  []string
	ThresholdSConns  []string
	StatSConns       []string
	OnlineCDRExports []string // list of CDRE templates to use for real-time CDR exports
	SchedulerConns   []string
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
		cdrscfg.ChargerSConns = make([]string, len(*jsnCdrsCfg.Chargers_conns))
		for idx, connID := range *jsnCdrsCfg.Chargers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.ChargerSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
			} else {
				cdrscfg.ChargerSConns[idx] = connID
			}
		}
	}
	if jsnCdrsCfg.Rals_conns != nil {
		cdrscfg.RaterConns = make([]string, len(*jsnCdrsCfg.Rals_conns))
		for idx, connID := range *jsnCdrsCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.RaterConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			} else {
				cdrscfg.RaterConns[idx] = connID
			}
		}
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.AttributeSConns = make([]string, len(*jsnCdrsCfg.Attributes_conns))
		for idx, connID := range *jsnCdrsCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				cdrscfg.AttributeSConns[idx] = connID
			}
		}
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.ThresholdSConns = make([]string, len(*jsnCdrsCfg.Thresholds_conns))
		for idx, connID := range *jsnCdrsCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				cdrscfg.ThresholdSConns[idx] = connID
			}
		}
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.StatSConns = make([]string, len(*jsnCdrsCfg.Stats_conns))
		for idx, connID := range *jsnCdrsCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
			} else {
				cdrscfg.StatSConns[idx] = connID
			}
		}
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		for _, expProfile := range *jsnCdrsCfg.Online_cdr_exports {
			cdrscfg.OnlineCDRExports = append(cdrscfg.OnlineCDRExports, expProfile)
		}
	}
	if jsnCdrsCfg.Scheduler_conns != nil {
		cdrscfg.SchedulerConns = make([]string, len(*jsnCdrsCfg.Scheduler_conns))
		for idx, connID := range *jsnCdrsCfg.Scheduler_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			} else {
				cdrscfg.SchedulerConns[idx] = connID
			}
		}
	}

	return nil
}

func (cdrscfg *CdrsCfg) AsMapInterface() map[string]interface{} {
	extraFields := make([]string, len(cdrscfg.ExtraFields))
	for i, item := range cdrscfg.ExtraFields {
		extraFields[i] = item.Rules
	}

	return map[string]interface{}{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.ExtraFieldsCfg:      extraFields,
		utils.StoreCdrsCfg:        cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ChargerSConnsCfg:    cdrscfg.ChargerSConns,
		utils.RALsConnsCfg:        cdrscfg.RaterConns,
		utils.AttributeSConnsCfg:  cdrscfg.AttributeSConns,
		utils.ThresholdSConnsCfg:  cdrscfg.ThresholdSConns,
		utils.StatSConnsCfg:       cdrscfg.StatSConns,
		utils.OnlineCDRExportsCfg: cdrscfg.OnlineCDRExports,
	}
}
