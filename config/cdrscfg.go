/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	Enabled          bool             // Enable CDR Server service
	ExtraFields      utils.RSRParsers // Extra fields to store in CDRs
	SMCostRetries    int
	Conns            map[string][]*DynamicStringSliceOpt
	OnlineCDRExports []string // list of CDRE templates to use for real-time CDR exports
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

func (cdrsOpts *CdrsOpts) loadFromJSONCfg(jsnCfg *CdrsOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Accounts != nil {
		var accounts []*DynamicBoolOpt
		accounts, err = IfaceToBoolDynamicOpts(jsnCfg.Accounts)
		if err != nil {
			return
		}
		cdrsOpts.Accounts = append(accounts, cdrsOpts.Accounts...)
	}
	if jsnCfg.Attributes != nil {
		var attributes []*DynamicBoolOpt
		attributes, err = IfaceToBoolDynamicOpts(jsnCfg.Attributes)
		if err != nil {
			return
		}
		cdrsOpts.Attributes = append(attributes, cdrsOpts.Attributes...)
	}
	if jsnCfg.Chargers != nil {
		var chargers []*DynamicBoolOpt
		chargers, err = IfaceToBoolDynamicOpts(jsnCfg.Attributes)
		if err != nil {
			return
		}
		cdrsOpts.Chargers = append(chargers, cdrsOpts.Chargers...)
	}
	if jsnCfg.Export != nil {
		var export []*DynamicBoolOpt
		export, err = IfaceToBoolDynamicOpts(jsnCfg.Export)
		if err != nil {
			return
		}
		cdrsOpts.Export = append(export, cdrsOpts.Export...)
	}
	if jsnCfg.Rates != nil {
		var rates []*DynamicBoolOpt
		rates, err = IfaceToBoolDynamicOpts(jsnCfg.Rates)
		if err != nil {
			return
		}
		cdrsOpts.Rates = append(rates, cdrsOpts.Rates...)
	}
	if jsnCfg.Stats != nil {
		var stats []*DynamicBoolOpt
		stats, err = IfaceToBoolDynamicOpts(jsnCfg.Stats)
		if err != nil {
			return
		}
		cdrsOpts.Stats = append(stats, cdrsOpts.Stats...)
	}
	if jsnCfg.Thresholds != nil {
		var thresholds []*DynamicBoolOpt
		thresholds, err = IfaceToBoolDynamicOpts(jsnCfg.Thresholds)
		if err != nil {
			return
		}
		cdrsOpts.Thresholds = append(thresholds, cdrsOpts.Thresholds...)
	}
	if jsnCfg.Refund != nil {
		var refund []*DynamicBoolOpt
		refund, err = IfaceToBoolDynamicOpts(jsnCfg.Refund)
		if err != nil {
			return
		}
		cdrsOpts.Refund = append(refund, cdrsOpts.Refund...)
	}
	if jsnCfg.Rerate != nil {
		var rerate []*DynamicBoolOpt
		rerate, err = IfaceToBoolDynamicOpts(jsnCfg.Rerate)
		if err != nil {
			return
		}
		cdrsOpts.Rerate = append(rerate, cdrsOpts.Rerate...)
	}
	if jsnCfg.Store != nil {
		var store []*DynamicBoolOpt
		store, err = IfaceToBoolDynamicOpts(jsnCfg.Store)
		if err != nil {
			return
		}
		cdrsOpts.Store = append(store, cdrsOpts.Store...)
	}
	return
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
		if cdrscfg.ExtraFields, err = utils.NewRSRParsersFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
			return
		}
	}
	if jsnCdrsCfg.Session_cost_retries != nil {
		cdrscfg.SMCostRetries = *jsnCdrsCfg.Session_cost_retries
	}
	if jsnCdrsCfg.Conns != nil {
		tagged := tagConns(jsnCdrsCfg.Conns)
		for connType, opts := range tagged {
			cdrscfg.Conns[connType] = opts
		}
	}
	if jsnCdrsCfg.Online_cdr_exports != nil {
		cdrscfg.OnlineCDRExports = append(cdrscfg.OnlineCDRExports, *jsnCdrsCfg.Online_cdr_exports...)
	}
	if jsnCdrsCfg.Opts != nil {
		cdrscfg.Opts.loadFromJSONCfg(jsnCdrsCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cdrscfg CdrsCfg) AsMapInterface() any {
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
		utils.ConnsCfg:            stripConns(cdrscfg.Conns),
		utils.OnlineCDRExportsCfg: slices.Clone(cdrscfg.OnlineCDRExports),
		utils.OptsCfg:             opts,
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
		Conns:         CloneConnsOpt(cdrscfg.Conns),
		Opts:          cdrscfg.Opts.Clone(),
	}
	if cdrscfg.OnlineCDRExports != nil {
		cln.OnlineCDRExports = slices.Clone(cdrscfg.OnlineCDRExports)
	}

	return
}

type CdrsOptsJson struct {
	Accounts   []*DynamicInterfaceOpt `json:"*accounts"`
	Attributes []*DynamicInterfaceOpt `json:"*attributes"`
	Chargers   []*DynamicInterfaceOpt `json:"*chargers"`
	Export     []*DynamicInterfaceOpt `json:"*ees"`
	Rates      []*DynamicInterfaceOpt `json:"*rates"`
	Stats      []*DynamicInterfaceOpt `json:"*stats"`
	Thresholds []*DynamicInterfaceOpt `json:"*thresholds"`
	Refund     []*DynamicInterfaceOpt `json:"*refund"`
	Rerate     []*DynamicInterfaceOpt `json:"*rerate"`
	Store      []*DynamicInterfaceOpt `json:"*store"`
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled              *bool
	Extra_fields         *[]string
	Session_cost_retries *int
	Conns                map[string][]*DynamicStringSliceOpt `json:"conns,omitempty"`
	Online_cdr_exports   *[]string
	Opts                 *CdrsOptsJson
}

func diffCdrsOptsJsonCfg(d *CdrsOptsJson, v1, v2 *CdrsOpts) *CdrsOptsJson {
	if d == nil {
		d = new(CdrsOptsJson)
	}
	if !DynamicBoolOptEqual(v1.Accounts, v2.Accounts) {
		d.Accounts = BoolToIfaceDynamicOpts(v2.Accounts)
	}
	if !DynamicBoolOptEqual(v1.Attributes, v2.Attributes) {
		d.Attributes = BoolToIfaceDynamicOpts(v2.Attributes)
	}
	if !DynamicBoolOptEqual(v1.Chargers, v2.Chargers) {
		d.Chargers = BoolToIfaceDynamicOpts(v2.Chargers)
	}
	if !DynamicBoolOptEqual(v1.Export, v2.Export) {
		d.Export = BoolToIfaceDynamicOpts(v2.Export)
	}
	if !DynamicBoolOptEqual(v1.Rates, v2.Rates) {
		d.Rates = BoolToIfaceDynamicOpts(v2.Rates)
	}
	if !DynamicBoolOptEqual(v1.Stats, v2.Stats) {
		d.Stats = BoolToIfaceDynamicOpts(v2.Stats)
	}
	if !DynamicBoolOptEqual(v1.Thresholds, v2.Thresholds) {
		d.Thresholds = BoolToIfaceDynamicOpts(v2.Thresholds)
	}
	if !DynamicBoolOptEqual(v1.Refund, v2.Refund) {
		d.Refund = BoolToIfaceDynamicOpts(v2.Refund)
	}
	if !DynamicBoolOptEqual(v1.Rerate, v2.Rerate) {
		d.Rerate = BoolToIfaceDynamicOpts(v2.Rerate)
	}
	if !DynamicBoolOptEqual(v1.Store, v2.Store) {
		d.Store = BoolToIfaceDynamicOpts(v2.Store)
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
	if !ConnsEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if !slices.Equal(v1.OnlineCDRExports, v2.OnlineCDRExports) {
		d.Online_cdr_exports = &v2.OnlineCDRExports
	}
	d.Opts = diffCdrsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
