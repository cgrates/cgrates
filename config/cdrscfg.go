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
	AttributeSConns  []string
	ThresholdSConns  []string
	StatSConns       []string
	OnlineCDRExports []string // list of CDRE templates to use for real-time CDR exports
	ActionSConns     []string
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
				cdrscfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		cdrscfg.OnlineCDRExports = append(cdrscfg.OnlineCDRExports, *jsnCdrsCfg.Online_cdr_exports...)
	}
	if jsnCdrsCfg.Actions_conns != nil {
		cdrscfg.ActionSConns = make([]string, len(*jsnCdrsCfg.Actions_conns))
		for idx, connID := range *jsnCdrsCfg.Actions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			cdrscfg.ActionSConns[idx] = connID
			if connID == utils.MetaInternal {
				cdrscfg.ActionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions)
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
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if cdrscfg.ActionSConns != nil {
		actionsConns := make([]string, len(cdrscfg.ActionSConns))
		for i, item := range cdrscfg.ActionSConns {
			actionsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions) {
				actionsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ActionSConnsCfg] = actionsConns
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
		cln.ChargerSConns = utils.CloneStringSlice(cdrscfg.ChargerSConns)
	}
	if cdrscfg.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(cdrscfg.AttributeSConns)
	}
	if cdrscfg.ThresholdSConns != nil {
		cln.ThresholdSConns = utils.CloneStringSlice(cdrscfg.ThresholdSConns)
	}
	if cdrscfg.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(cdrscfg.StatSConns)
	}
	if cdrscfg.OnlineCDRExports != nil {
		cln.OnlineCDRExports = utils.CloneStringSlice(cdrscfg.OnlineCDRExports)
	}
	if cdrscfg.ActionSConns != nil {
		cln.ActionSConns = utils.CloneStringSlice(cdrscfg.ActionSConns)
	}
	if cdrscfg.EEsConns != nil {
		cln.EEsConns = utils.CloneStringSlice(cdrscfg.EEsConns)
	}

	return
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled              *bool
	Extra_fields         *[]string
	Store_cdrs           *bool
	Session_cost_retries *int
	Chargers_conns       *[]string
	Attributes_conns     *[]string
	Thresholds_conns     *[]string
	Stats_conns          *[]string
	Online_cdr_exports   *[]string
	Actions_conns        *[]string
	Ees_conns            *[]string
}

func diffCdrsJsonCfg(d *CdrsJsonCfg, v1, v2 *CdrsCfg) *CdrsJsonCfg {
	if d == nil {
		d = new(CdrsJsonCfg)
	}

	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	extra1 := v1.ExtraFields.AsStringSlice()
	extra2 := v2.ExtraFields.AsStringSlice()
	if !utils.SliceStringEqual(extra1, extra2) {
		d.Extra_fields = &extra2
	}
	if v1.StoreCdrs != v2.StoreCdrs {
		d.Store_cdrs = utils.BoolPointer(v2.StoreCdrs)
	}
	if v1.SMCostRetries != v2.SMCostRetries {
		d.Session_cost_retries = utils.IntPointer(v2.SMCostRetries)
	}
	if !utils.SliceStringEqual(v1.ChargerSConns, v2.ChargerSConns) {
		d.Chargers_conns = &v2.ChargerSConns
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = &v2.AttributeSConns
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = &v2.ThresholdSConns
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = &v2.StatSConns
	}
	if !utils.SliceStringEqual(v1.OnlineCDRExports, v2.OnlineCDRExports) {
		d.Online_cdr_exports = &v2.OnlineCDRExports
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = &v2.ActionSConns
	}
	if !utils.SliceStringEqual(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = &v2.EEsConns
	}
	return d
}
