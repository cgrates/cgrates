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

type CdrsOpts struct {
	Accounts   []*utils.DynamicBoolOpt
	Attributes []*utils.DynamicBoolOpt
	Chargers   []*utils.DynamicBoolOpt
	Export     []*utils.DynamicBoolOpt
	Rates      []*utils.DynamicBoolOpt
	Stats      []*utils.DynamicBoolOpt
	Thresholds []*utils.DynamicBoolOpt
}

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
	Opts             *CdrsOpts
}

// loadCdrsCfg loads the Cdrs section of the configuration
func (cdrscfg *CdrsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnCdrsCfg := new(CdrsJsonCfg)
	if err = jsnCfg.GetSection(ctx, CDRsJSON, jsnCdrsCfg); err != nil {
		return
	}
	return cdrscfg.loadFromJSONCfg(jsnCdrsCfg)
}

func (cdrsOpts *CdrsOpts) loadFromJSONCfg(jsnCfg *CdrsOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Accounts != nil {
		cdrsOpts.Accounts = utils.MapToDynamicBoolOpts(jsnCfg.Accounts)
	}
	if jsnCfg.Attributes != nil {
		cdrsOpts.Attributes = utils.MapToDynamicBoolOpts(jsnCfg.Attributes)
	}
	if jsnCfg.Chargers != nil {
		cdrsOpts.Chargers = utils.MapToDynamicBoolOpts(jsnCfg.Chargers)
	}
	if jsnCfg.Export != nil {
		cdrsOpts.Export = utils.MapToDynamicBoolOpts(jsnCfg.Export)
	}
	if jsnCfg.Rates != nil {
		cdrsOpts.Rates = utils.MapToDynamicBoolOpts(jsnCfg.Rates)
	}
	if jsnCfg.Stats != nil {
		cdrsOpts.Stats = utils.MapToDynamicBoolOpts(jsnCfg.Stats)
	}
	if jsnCfg.Thresholds != nil {
		cdrsOpts.Thresholds = utils.MapToDynamicBoolOpts(jsnCfg.Thresholds)
	}
}

// loadFromJSONCfg loads Cdrs config from JsonCfg
func (cdrscfg *CdrsCfg) loadFromJSONCfg(jsnCdrsCfg *CdrsJsonCfg) (err error) {
	if jsnCdrsCfg == nil {
		return
	}
	if jsnCdrsCfg.Enabled != nil {
		cdrscfg.Enabled = *jsnCdrsCfg.Enabled
	}
	if jsnCdrsCfg.Extra_fields != nil {
		if cdrscfg.ExtraFields, err = NewRSRParsersFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
			return
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
	if jsnCdrsCfg.Opts != nil {
		cdrscfg.Opts.loadFromJSONCfg(jsnCdrsCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (cdrscfg CdrsCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.MetaAccounts:   utils.DynamicBoolOptsToMap(cdrscfg.Opts.Accounts),
		utils.MetaAttributes: utils.DynamicBoolOptsToMap(cdrscfg.Opts.Attributes),
		utils.MetaChargers:   utils.DynamicBoolOptsToMap(cdrscfg.Opts.Chargers),
		utils.MetaExport:     utils.DynamicBoolOptsToMap(cdrscfg.Opts.Export),
		utils.MetaRateS:      utils.DynamicBoolOptsToMap(cdrscfg.Opts.Rates),
		utils.MetaStats:      utils.DynamicBoolOptsToMap(cdrscfg.Opts.Stats),
		utils.MetaThresholds: utils.DynamicBoolOptsToMap(cdrscfg.Opts.Thresholds),
	}
	mp := map[string]interface{}{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.StoreCdrsCfg:        cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ExtraFieldsCfg:      cdrscfg.ExtraFields.AsStringSlice(),
		utils.OnlineCDRExportsCfg: utils.CloneStringSlice(cdrscfg.OnlineCDRExports),
		utils.OptsCfg:             opts,
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

func (cdrsOpts *CdrsOpts) Clone() *CdrsOpts {
	var accS []*utils.DynamicBoolOpt
	if cdrsOpts.Accounts != nil {
		accS = utils.CloneDynamicBoolOpt(cdrsOpts.Accounts)
	}
	var attrS []*utils.DynamicBoolOpt
	if cdrsOpts.Attributes != nil {
		attrS = utils.CloneDynamicBoolOpt(cdrsOpts.Attributes)
	}
	var chrgS []*utils.DynamicBoolOpt
	if cdrsOpts.Chargers != nil {
		chrgS = utils.CloneDynamicBoolOpt(cdrsOpts.Chargers)
	}
	var export []*utils.DynamicBoolOpt
	if cdrsOpts.Export != nil {
		export = utils.CloneDynamicBoolOpt(cdrsOpts.Export)
	}
	var rtS []*utils.DynamicBoolOpt
	if cdrsOpts.Rates != nil {
		rtS = utils.CloneDynamicBoolOpt(cdrsOpts.Rates)
	}
	var stS []*utils.DynamicBoolOpt
	if cdrsOpts.Stats != nil {
		stS = utils.CloneDynamicBoolOpt(cdrsOpts.Stats)
	}
	var thdS []*utils.DynamicBoolOpt
	if cdrsOpts.Thresholds != nil {
		thdS = utils.CloneDynamicBoolOpt(cdrsOpts.Thresholds)
	}
	return &CdrsOpts{
		Accounts:   accS,
		Attributes: attrS,
		Chargers:   chrgS,
		Export:     export,
		Rates:      rtS,
		Stats:      stS,
		Thresholds: thdS,
	}
}

// Clone returns a deep copy of CdrsCfg
func (cdrscfg CdrsCfg) Clone() (cln *CdrsCfg) {
	cln = &CdrsCfg{
		Enabled:       cdrscfg.Enabled,
		ExtraFields:   cdrscfg.ExtraFields.Clone(),
		StoreCdrs:     cdrscfg.StoreCdrs,
		SMCostRetries: cdrscfg.SMCostRetries,
		Opts:          cdrscfg.Opts.Clone(),
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

type CdrsOptsJson struct {
	Accounts   map[string]bool `json:"*accountS"`
	Attributes map[string]bool `json:"*attributeS"`
	Chargers   map[string]bool `json:"*chargerS"`
	Export     map[string]bool `json:"*export"`
	Rates      map[string]bool `json:"*rateS"`
	Stats      map[string]bool `json:"*statS"`
	Thresholds map[string]bool `json:"*thresholdS"`
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
	Opts                 *CdrsOptsJson
}

func diffCdrsOptsJsonCfg(d *CdrsOptsJson, v1, v2 *CdrsOpts) *CdrsOptsJson {
	if d == nil {
		d = new(CdrsOptsJson)
	}
	if !utils.DynamicBoolOptEqual(v1.Accounts, v2.Accounts) {
		d.Accounts = utils.DynamicBoolOptsToMap(v2.Accounts)
	}
	if !utils.DynamicBoolOptEqual(v1.Attributes, v2.Attributes) {
		d.Attributes = utils.DynamicBoolOptsToMap(v2.Attributes)
	}
	if !utils.DynamicBoolOptEqual(v1.Chargers, v2.Chargers) {
		d.Chargers = utils.DynamicBoolOptsToMap(v2.Chargers)
	}
	if !utils.DynamicBoolOptEqual(v1.Export, v2.Export) {
		d.Export = utils.DynamicBoolOptsToMap(v2.Export)
	}
	if !utils.DynamicBoolOptEqual(v1.Rates, v2.Rates) {
		d.Rates = utils.DynamicBoolOptsToMap(v2.Rates)
	}
	if !utils.DynamicBoolOptEqual(v1.Stats, v2.Stats) {
		d.Stats = utils.DynamicBoolOptsToMap(v2.Stats)
	}
	if !utils.DynamicBoolOptEqual(v1.Thresholds, v2.Thresholds) {
		d.Thresholds = utils.DynamicBoolOptsToMap(v2.Thresholds)
	}
	return d
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
	d.Opts = diffCdrsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
