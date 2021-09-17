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
	"github.com/cgrates/birpc/context"
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
	RateSConns       []string
	AccountSConns    []string
}

// loadCdrsCfg loads the Cdrs section of the configuration
func (cdrscfg *CdrsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnCdrsCfg := new(CdrsJsonCfg)
	if err = jsnCfg.GetSection(ctx, CDRsJSON, jsnCdrsCfg); err != nil {
		return
	}
	return cdrscfg.loadFromJSONCfg(jsnCdrsCfg)
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
		cdrscfg.ChargerSConns = updateInternalConns(*jsnCdrsCfg.Chargers_conns, utils.MetaChargers)
	}
	if jsnCdrsCfg.Attributes_conns != nil {
		cdrscfg.AttributeSConns = updateInternalConns(*jsnCdrsCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCdrsCfg.Thresholds_conns != nil {
		cdrscfg.ThresholdSConns = updateInternalConns(*jsnCdrsCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCdrsCfg.Stats_conns != nil {
		cdrscfg.StatSConns = updateInternalConns(*jsnCdrsCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		cdrscfg.OnlineCDRExports = append(cdrscfg.OnlineCDRExports, *jsnCdrsCfg.Online_cdr_exports...)
	}
	if jsnCdrsCfg.Actions_conns != nil {
		cdrscfg.ActionSConns = updateInternalConns(*jsnCdrsCfg.Actions_conns, utils.MetaActions)
	}
	if jsnCdrsCfg.Ees_conns != nil {
		cdrscfg.EEsConns = updateInternalConns(*jsnCdrsCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCdrsCfg.Rates_conns != nil {
		cdrscfg.RateSConns = updateInternalConns(*jsnCdrsCfg.Rates_conns, utils.MetaRateS)
	}
	if jsnCdrsCfg.Accounts_conns != nil {
		cdrscfg.AccountSConns = updateInternalConns(*jsnCdrsCfg.Accounts_conns, utils.MetaAccounts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (cdrscfg CdrsCfg) AsMapInterface(string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.StoreCdrsCfg:        cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ExtraFieldsCfg:      cdrscfg.ExtraFields.AsStringSlice(),
		utils.OnlineCDRExportsCfg: utils.CloneStringSlice(cdrscfg.OnlineCDRExports),
	}

	if cdrscfg.ChargerSConns != nil {
		mp[utils.ChargerSConnsCfg] = getInternalJSONConns(cdrscfg.ChargerSConns)
	}
	if cdrscfg.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(cdrscfg.AttributeSConns)
	}
	if cdrscfg.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(cdrscfg.ThresholdSConns)
	}
	if cdrscfg.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(cdrscfg.StatSConns)
	}
	if cdrscfg.ActionSConns != nil {
		mp[utils.ActionSConnsCfg] = getInternalJSONConns(cdrscfg.ActionSConns)
	}
	if cdrscfg.EEsConns != nil {
		mp[utils.EEsConnsCfg] = getInternalJSONConns(cdrscfg.EEsConns)
	}
	if cdrscfg.RateSConns != nil {
		mp[utils.RateSConnsCfg] = getInternalJSONConns(cdrscfg.RateSConns)
	}
	if cdrscfg.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(cdrscfg.AccountSConns)
	}
	return mp
}

func (CdrsCfg) SName() string                 { return CDRsJSON }
func (cdrscfg CdrsCfg) CloneSection() Section { return cdrscfg.Clone() }

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
	if cdrscfg.RateSConns != nil {
		cln.RateSConns = utils.CloneStringSlice(cdrscfg.RateSConns)
	}
	if cdrscfg.AccountSConns != nil {
		cln.AccountSConns = utils.CloneStringSlice(cdrscfg.AccountSConns)
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
	Rates_conns          *[]string
	Accounts_conns       *[]string
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
		d.Chargers_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ChargerSConns))
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.OnlineCDRExports, v2.OnlineCDRExports) {
		d.Online_cdr_exports = &v2.OnlineCDRExports
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	if !utils.SliceStringEqual(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	if !utils.SliceStringEqual(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !utils.SliceStringEqual(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	return d
}
