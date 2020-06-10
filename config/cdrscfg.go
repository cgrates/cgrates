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
	"strings"

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
	EEsConns         []string
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

	if jsnCdrsCfg.Ees_conns != nil {
		cdrscfg.EEsConns = make([]string, len(*jsnCdrsCfg.Ees_conns))
		for idx, connID := range *jsnCdrsCfg.Ees_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				cdrscfg.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			} else {
				cdrscfg.EEsConns[idx] = connID
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
	onlineCDRExports := make([]string, len(cdrscfg.OnlineCDRExports))
	for i, item := range cdrscfg.OnlineCDRExports {
		onlineCDRExports[i] = item
	}

	chargerSConns := make([]string, len(cdrscfg.ChargerSConns))
	for i, item := range cdrscfg.ChargerSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
		if item == buf {
			chargerSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaChargers, utils.EmptyString)
		} else {
			chargerSConns[i] = item
		}
	}
	RALsConns := make([]string, len(cdrscfg.RaterConns))
	for i, item := range cdrscfg.RaterConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)

		if item == buf {
			RALsConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResponder, utils.EmptyString)
		} else {
			RALsConns[i] = item
		}
	}

	attributeSConns := make([]string, len(cdrscfg.AttributeSConns))
	for i, item := range cdrscfg.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}

	thresholdSConns := make([]string, len(cdrscfg.ThresholdSConns))
	for i, item := range cdrscfg.ThresholdSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
		if item == buf {
			thresholdSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds, utils.EmptyString)
		} else {
			thresholdSConns[i] = item
		}
	}
	statSConns := make([]string, len(cdrscfg.StatSConns))
	for i, item := range cdrscfg.StatSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
		if item == buf {
			statSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaStatS, utils.EmptyString)
		} else {
			statSConns[i] = item
		}
	}
	schedulerConns := make([]string, len(cdrscfg.SchedulerConns))
	for i, item := range cdrscfg.SchedulerConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
		if item == buf {
			schedulerConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaScheduler, utils.EmptyString)
		} else {
			schedulerConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.ExtraFieldsCfg:      extraFields,
		utils.StoreCdrsCfg:        cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ChargerSConnsCfg:    chargerSConns,
		utils.RALsConnsCfg:        RALsConns,
		utils.AttributeSConnsCfg:  attributeSConns,
		utils.ThresholdSConnsCfg:  thresholdSConns,
		utils.StatSConnsCfg:       statSConns,
		utils.OnlineCDRExportsCfg: onlineCDRExports,
		utils.SchedulerConnsCfg:   schedulerConns,
	}
}
