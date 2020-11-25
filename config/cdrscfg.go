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

// CdrsCfg is the CDR server
type CdrsCfg struct {
	Enabled          bool       // Enable CDR Server service
	ExtraFields      RSRParsers // Extra fields to store in CDRs
	StoreCdrs        bool       // store cdrs in storDb
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

// loadFromJSONCfg loads Cdrs config from JsonCfg
func (cdrscfg *CdrsCfg) loadFromJSONCfg(jsnCdrsCfg *CdrsJsonCfg) (err error) {
	if jsnCdrsCfg == nil {
		return nil
	}
	if jsnCdrsCfg.Enabled != nil {
		cdrscfg.Enabled = *jsnCdrsCfg.Enabled
	}
	if jsnCdrsCfg.Extra_fields != nil {
		if cdrscfg.ExtraFields, err = NewRSRParsersFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
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
			cdrscfg.ChargerSConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.ChargerSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
			}
		}
	}
	if jsnCdrsCfg.Rals_conns != nil {
		cdrscfg.RaterConns = make([]string, len(*jsnCdrsCfg.Rals_conns))
		for idx, connID := range *jsnCdrsCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.RaterConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.RaterConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			}
		}
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.AttributeSConns = make([]string, len(*jsnCdrsCfg.Attributes_conns))
		for idx, connID := range *jsnCdrsCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.AttributeSConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.ThresholdSConns = make([]string, len(*jsnCdrsCfg.Thresholds_conns))
		for idx, connID := range *jsnCdrsCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.ThresholdSConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.StatSConns = make([]string, len(*jsnCdrsCfg.Stats_conns))
		for idx, connID := range *jsnCdrsCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
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
			cdrscfg.SchedulerConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			}
		}
	}

	if jsnCdrsCfg.Ees_conns != nil {
		cdrscfg.EEsConns = make([]string, len(*jsnCdrsCfg.Ees_conns))
		for idx, connID := range *jsnCdrsCfg.Ees_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (cdrscfg *CdrsCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:       cdrscfg.Enabled,
		utils.StoreCdrsCfg:     cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg: cdrscfg.SMCostRetries,
	}

	extraFields := make([]string, len(cdrscfg.ExtraFields))
	for i, item := range cdrscfg.ExtraFields {
		extraFields[i] = item.Rules
	}
	initialMP[utils.ExtraFieldsCfg] = extraFields

	onlineCDRExports := make([]string, len(cdrscfg.OnlineCDRExports))
	for i, item := range cdrscfg.OnlineCDRExports {
		onlineCDRExports[i] = item
	}
	initialMP[utils.OnlineCDRExportsCfg] = onlineCDRExports

	if cdrscfg.ChargerSConns != nil {
		chargerSConns := make([]string, len(cdrscfg.ChargerSConns))
		for i, item := range cdrscfg.ChargerSConns {
			chargerSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers) {
				chargerSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ChargerSConnsCfg] = chargerSConns
	}
	if cdrscfg.RaterConns != nil {
		raterConns := make([]string, len(cdrscfg.RaterConns))
		for i, item := range cdrscfg.RaterConns {
			raterConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder) {
				raterConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.RALsConnsCfg] = raterConns
	}
	if cdrscfg.AttributeSConns != nil {
		attributeSConns := make([]string, len(cdrscfg.AttributeSConns))
		for i, item := range cdrscfg.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if cdrscfg.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(cdrscfg.ThresholdSConns))
		for i, item := range cdrscfg.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if cdrscfg.StatSConns != nil {
		statSConns := make([]string, len(cdrscfg.StatSConns))
		for i, item := range cdrscfg.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if cdrscfg.SchedulerConns != nil {
		schedulerConns := make([]string, len(cdrscfg.SchedulerConns))
		for i, item := range cdrscfg.SchedulerConns {
			schedulerConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler) {
				schedulerConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SchedulerConnsCfg] = schedulerConns
	}
	if cdrscfg.EEsConns != nil {
		eesConns := make([]string, len(cdrscfg.EEsConns))
		for i, item := range cdrscfg.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.EEsConnsCfg] = eesConns
	}
	return
}

// Clone returns a deep copy of CdrsCfg
func (cdrscfg CdrsCfg) Clone() (cln *CdrsCfg) {
	cln = &CdrsCfg{
		Enabled:       cdrscfg.Enabled,
		ExtraFields:   cdrscfg.ExtraFields.Clone(),
		StoreCdrs:     cdrscfg.StoreCdrs,
		SMCostRetries: cdrscfg.SMCostRetries,
	}
	if cdrscfg.ChargerSConns != nil {
		cln.ChargerSConns = make([]string, len(cdrscfg.ChargerSConns))
		for i, con := range cdrscfg.ChargerSConns {
			cln.ChargerSConns[i] = con
		}
	}
	if cdrscfg.RaterConns != nil {
		cln.RaterConns = make([]string, len(cdrscfg.RaterConns))
		for i, con := range cdrscfg.RaterConns {
			cln.RaterConns[i] = con
		}
	}
	if cdrscfg.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(cdrscfg.AttributeSConns))
		for i, con := range cdrscfg.AttributeSConns {
			cln.AttributeSConns[i] = con
		}
	}
	if cdrscfg.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(cdrscfg.ThresholdSConns))
		for i, con := range cdrscfg.ThresholdSConns {
			cln.ThresholdSConns[i] = con
		}
	}
	if cdrscfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(cdrscfg.StatSConns))
		for i, con := range cdrscfg.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if cdrscfg.OnlineCDRExports != nil {
		cln.OnlineCDRExports = make([]string, len(cdrscfg.OnlineCDRExports))
		for i, con := range cdrscfg.OnlineCDRExports {
			cln.OnlineCDRExports[i] = con
		}
	}
	if cdrscfg.SchedulerConns != nil {
		cln.SchedulerConns = make([]string, len(cdrscfg.SchedulerConns))
		for i, con := range cdrscfg.SchedulerConns {
			cln.SchedulerConns[i] = con
		}
	}
	if cdrscfg.EEsConns != nil {
		cln.EEsConns = make([]string, len(cdrscfg.EEsConns))
		for i, con := range cdrscfg.EEsConns {
			cln.EEsConns[i] = con
		}
	}

	return
}
