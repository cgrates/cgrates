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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	CDRsAccountsDftOpt   = false
	CDRsAttributesDftOpt = false
	CDRsChargersDftOpt   = false
	CDRsExportDftOpt     = false
	CDRsRatesDftOpt      = false
	CDRsStatsDftOpt      = false
	CDRsThresholdsDftOpt = false
	CDRsRefundDftOpt     = false
	CDRsRerateDftOpt     = false
	CDRsStoreDftOpt      = true
)

type CdrsOpts struct {
	Accounts   []*DynamicBoolOpt
	Attributes []*DynamicBoolOpt
	Chargers   []*DynamicBoolOpt
	Export     []*DynamicBoolOpt
	Rates      []*DynamicBoolOpt
	Stats      []*DynamicBoolOpt
	Thresholds []*DynamicBoolOpt
	Refund     []*DynamicBoolOpt
	Rerate     []*DynamicBoolOpt
	Store      []*DynamicBoolOpt
}

// CdrsCfg is the CDR server
type CdrsCfg struct {
	Enabled          bool       // Enable CDR Server service
	ExtraFields      RSRParsers // Extra fields to store in CDRs
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
		cdrsOpts.Accounts = append(cdrsOpts.Accounts, jsnCfg.Accounts...)
	}
	if jsnCfg.Attributes != nil {
		cdrsOpts.Attributes = append(cdrsOpts.Attributes, jsnCfg.Attributes...)
	}
	if jsnCfg.Chargers != nil {
		cdrsOpts.Chargers = append(cdrsOpts.Chargers, jsnCfg.Chargers...)
	}
	if jsnCfg.Export != nil {
		cdrsOpts.Export = append(cdrsOpts.Export, jsnCfg.Export...)
	}
	if jsnCfg.Rates != nil {
		cdrsOpts.Rates = append(cdrsOpts.Rates, jsnCfg.Rates...)
	}
	if jsnCfg.Stats != nil {
		cdrsOpts.Stats = append(cdrsOpts.Stats, jsnCfg.Stats...)
	}
	if jsnCfg.Thresholds != nil {
		cdrsOpts.Thresholds = append(cdrsOpts.Thresholds, jsnCfg.Thresholds...)
	}
	if jsnCfg.Refund != nil {
		cdrsOpts.Refund = append(cdrsOpts.Refund, jsnCfg.Refund...)
	}
	if jsnCfg.Rerate != nil {
		cdrsOpts.Rerate = append(cdrsOpts.Rerate, jsnCfg.Rerate...)
	}
	if jsnCfg.Store != nil {
		cdrsOpts.Store = append(cdrsOpts.Store, jsnCfg.Store...)
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
		cdrscfg.RateSConns = updateInternalConns(*jsnCdrsCfg.Rates_conns, utils.MetaRates)
	}
	if jsnCdrsCfg.Accounts_conns != nil {
		cdrscfg.AccountSConns = updateInternalConns(*jsnCdrsCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCdrsCfg.Opts != nil {
		cdrscfg.Opts.loadFromJSONCfg(jsnCdrsCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cdrscfg CdrsCfg) AsMapInterface(string) any {
	opts := map[string]any{
		utils.MetaAccounts:   cdrscfg.Opts.Accounts,
		utils.MetaAttributes: cdrscfg.Opts.Attributes,
		utils.MetaChargers:   cdrscfg.Opts.Chargers,
		utils.MetaEEs:        cdrscfg.Opts.Export,
		utils.MetaRates:      cdrscfg.Opts.Rates,
		utils.MetaStats:      cdrscfg.Opts.Stats,
		utils.MetaThresholds: cdrscfg.Opts.Thresholds,
		utils.MetaRefund:     cdrscfg.Opts.Refund,
		utils.MetaRerate:     cdrscfg.Opts.Rerate,
		utils.MetaStore:      cdrscfg.Opts.Store,
	}
	mp := map[string]any{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ExtraFieldsCfg:      cdrscfg.ExtraFields.AsStringSlice(),
		utils.OnlineCDRExportsCfg: slices.Clone(cdrscfg.OnlineCDRExports),
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
	var accS []*DynamicBoolOpt
	if cdrsOpts.Accounts != nil {
		accS = CloneDynamicBoolOpt(cdrsOpts.Accounts)
	}
	var attrS []*DynamicBoolOpt
	if cdrsOpts.Attributes != nil {
		attrS = CloneDynamicBoolOpt(cdrsOpts.Attributes)
	}
	var chrgS []*DynamicBoolOpt
	if cdrsOpts.Chargers != nil {
		chrgS = CloneDynamicBoolOpt(cdrsOpts.Chargers)
	}
	var export []*DynamicBoolOpt
	if cdrsOpts.Export != nil {
		export = CloneDynamicBoolOpt(cdrsOpts.Export)
	}
	var rtS []*DynamicBoolOpt
	if cdrsOpts.Rates != nil {
		rtS = CloneDynamicBoolOpt(cdrsOpts.Rates)
	}
	var stS []*DynamicBoolOpt
	if cdrsOpts.Stats != nil {
		stS = CloneDynamicBoolOpt(cdrsOpts.Stats)
	}
	var thdS []*DynamicBoolOpt
	if cdrsOpts.Thresholds != nil {
		thdS = CloneDynamicBoolOpt(cdrsOpts.Thresholds)
	}
	var refund []*DynamicBoolOpt
	if cdrsOpts.Refund != nil {
		refund = CloneDynamicBoolOpt(cdrsOpts.Refund)
	}
	var rerate []*DynamicBoolOpt
	if cdrsOpts.Rerate != nil {
		rerate = CloneDynamicBoolOpt(cdrsOpts.Rerate)
	}
	var store []*DynamicBoolOpt
	if cdrsOpts.Store != nil {
		store = CloneDynamicBoolOpt(cdrsOpts.Store)
	}
	return &CdrsOpts{
		Accounts:   accS,
		Attributes: attrS,
		Chargers:   chrgS,
		Export:     export,
		Rates:      rtS,
		Stats:      stS,
		Thresholds: thdS,
		Refund:     refund,
		Rerate:     rerate,
		Store:      store,
	}
}

// Clone returns a deep copy of CdrsCfg
func (cdrscfg CdrsCfg) Clone() (cln *CdrsCfg) {
	cln = &CdrsCfg{
		Enabled:       cdrscfg.Enabled,
		ExtraFields:   cdrscfg.ExtraFields.Clone(),
		SMCostRetries: cdrscfg.SMCostRetries,
		Opts:          cdrscfg.Opts.Clone(),
	}
	if cdrscfg.ChargerSConns != nil {
		cln.ChargerSConns = slices.Clone(cdrscfg.ChargerSConns)
	}
	if cdrscfg.AttributeSConns != nil {
		cln.AttributeSConns = slices.Clone(cdrscfg.AttributeSConns)
	}
	if cdrscfg.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(cdrscfg.ThresholdSConns)
	}
	if cdrscfg.StatSConns != nil {
		cln.StatSConns = slices.Clone(cdrscfg.StatSConns)
	}
	if cdrscfg.OnlineCDRExports != nil {
		cln.OnlineCDRExports = slices.Clone(cdrscfg.OnlineCDRExports)
	}
	if cdrscfg.ActionSConns != nil {
		cln.ActionSConns = slices.Clone(cdrscfg.ActionSConns)
	}
	if cdrscfg.EEsConns != nil {
		cln.EEsConns = slices.Clone(cdrscfg.EEsConns)
	}
	if cdrscfg.RateSConns != nil {
		cln.RateSConns = slices.Clone(cdrscfg.RateSConns)
	}
	if cdrscfg.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(cdrscfg.AccountSConns)
	}

	return
}

type CdrsOptsJson struct {
	Accounts   []*DynamicBoolOpt `json:"*accounts"`
	Attributes []*DynamicBoolOpt `json:"*attributes"`
	Chargers   []*DynamicBoolOpt `json:"*chargers"`
	Export     []*DynamicBoolOpt `json:"*ees"`
	Rates      []*DynamicBoolOpt `json:"*rates"`
	Stats      []*DynamicBoolOpt `json:"*stats"`
	Thresholds []*DynamicBoolOpt `json:"*thresholds"`
	Refund     []*DynamicBoolOpt `json:"*refund"`
	Rerate     []*DynamicBoolOpt `json:"*rerate"`
	Store      []*DynamicBoolOpt `json:"*store"`
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled              *bool
	Extra_fields         *[]string
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
	if !DynamicBoolOptEqual(v1.Accounts, v2.Accounts) {
		d.Accounts = v2.Accounts
	}
	if !DynamicBoolOptEqual(v1.Attributes, v2.Attributes) {
		d.Attributes = v2.Attributes
	}
	if !DynamicBoolOptEqual(v1.Chargers, v2.Chargers) {
		d.Chargers = v2.Chargers
	}
	if !DynamicBoolOptEqual(v1.Export, v2.Export) {
		d.Export = v2.Export
	}
	if !DynamicBoolOptEqual(v1.Rates, v2.Rates) {
		d.Rates = v2.Rates
	}
	if !DynamicBoolOptEqual(v1.Stats, v2.Stats) {
		d.Stats = v2.Stats
	}
	if !DynamicBoolOptEqual(v1.Thresholds, v2.Thresholds) {
		d.Thresholds = v2.Thresholds
	}
	if !DynamicBoolOptEqual(v1.Refund, v2.Refund) {
		d.Refund = v2.Refund
	}
	if !DynamicBoolOptEqual(v1.Rerate, v2.Rerate) {
		d.Rerate = v2.Rerate
	}
	if !DynamicBoolOptEqual(v1.Store, v2.Store) {
		d.Store = v2.Store
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
	if !slices.Equal(extra1, extra2) {
		d.Extra_fields = &extra2
	}
	if v1.SMCostRetries != v2.SMCostRetries {
		d.Session_cost_retries = utils.IntPointer(v2.SMCostRetries)
	}
	if !slices.Equal(v1.ChargerSConns, v2.ChargerSConns) {
		d.Chargers_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ChargerSConns))
	}
	if !slices.Equal(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.OnlineCDRExports, v2.OnlineCDRExports) {
		d.Online_cdr_exports = &v2.OnlineCDRExports
	}
	if !slices.Equal(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	if !slices.Equal(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	d.Opts = diffCdrsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
