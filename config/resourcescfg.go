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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type ResourcesOpts struct {
	UsageID  []*utils.DynamicStringOpt
	UsageTTL []*utils.DynamicDurationOpt
	Units    []*utils.DynamicFloat64Opt
}

// ResourceSConfig is resorces section config
type ResourceSConfig struct {
	Enabled             bool
	IndexedSelects      bool
	ThresholdSConns     []string
	StoreInterval       time.Duration // Dump regularly from cache into dataDB
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	Opts                *ResourcesOpts
}

// loadResourceSCfg loads the ResourceS section of the configuration
func (rlcfg *ResourceSConfig) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRLSCfg := new(ResourceSJsonCfg)
	if err = jsnCfg.GetSection(ctx, ResourceSJSON, jsnRLSCfg); err != nil {
		return
	}
	return rlcfg.loadFromJSONCfg(jsnRLSCfg)
}

func (rsOpts *ResourcesOpts) loadFromJSONCfg(jsnCfg *ResourcesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.UsageID != nil {
		rsOpts.UsageID = utils.MapToDynamicStringOpts(jsnCfg.UsageID)
	}
	if jsnCfg.UsageTTL != nil {
		if rsOpts.UsageTTL, err = utils.MapToDynamicDurationOpts(jsnCfg.UsageTTL); err != nil {
			return
		}
	}
	if jsnCfg.Units != nil {
		rsOpts.Units = utils.MapToDynamicFloat64Opts(jsnCfg.Units)
	}
	return
}

func (rlcfg *ResourceSConfig) loadFromJSONCfg(jsnCfg *ResourceSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		rlcfg.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Thresholds_conns != nil {
		rlcfg.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Store_interval != nil {
		if rlcfg.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		rlcfg.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		rlcfg.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		rlcfg.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		rlcfg.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		err = rlcfg.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rlcfg ResourceSConfig) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.MetaUsageIDCfg:  utils.DynamicStringOptsToMap(rlcfg.Opts.UsageID),
		utils.MetaUsageTTLCfg: utils.DynamicDurationOptsToMap(rlcfg.Opts.UsageTTL),
		utils.MetaUnitsCfg:    utils.DynamicFloat64OptsToMap(rlcfg.Opts.Units),
	}
	mp := map[string]interface{}{
		utils.EnabledCfg:        rlcfg.Enabled,
		utils.IndexedSelectsCfg: rlcfg.IndexedSelects,
		utils.NestedFieldsCfg:   rlcfg.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.OptsCfg:           opts,
	}
	if rlcfg.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(rlcfg.ThresholdSConns)
	}
	if rlcfg.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = utils.CloneStringSlice(*rlcfg.StringIndexedFields)
	}
	if rlcfg.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = utils.CloneStringSlice(*rlcfg.PrefixIndexedFields)
	}
	if rlcfg.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = utils.CloneStringSlice(*rlcfg.SuffixIndexedFields)
	}
	if rlcfg.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = rlcfg.StoreInterval.String()
	}
	return mp
}

func (ResourceSConfig) SName() string               { return ResourceSJSON }
func (rlcfg ResourceSConfig) CloneSection() Section { return rlcfg.Clone() }

func (rsOpts *ResourcesOpts) Clone() (cln *ResourcesOpts) {
	var usageID []*utils.DynamicStringOpt
	if rsOpts.UsageID != nil {
		usageID = utils.CloneDynamicStringOpt(rsOpts.UsageID)
	}
	var usageTTL []*utils.DynamicDurationOpt
	if rsOpts.UsageTTL != nil {
		usageTTL = utils.CloneDynamicDurationOpt(rsOpts.UsageTTL)
	}
	var units []*utils.DynamicFloat64Opt
	if rsOpts.Units != nil {
		units = utils.CloneDynamicFloat64Opt(rsOpts.Units)
	}
	cln = &ResourcesOpts{
		UsageID:  usageID,
		UsageTTL: usageTTL,
		Units:    units,
	}
	return
}

// Clone returns a deep copy of ResourceSConfig
func (rlcfg ResourceSConfig) Clone() (cln *ResourceSConfig) {
	cln = &ResourceSConfig{
		Enabled:        rlcfg.Enabled,
		IndexedSelects: rlcfg.IndexedSelects,
		StoreInterval:  rlcfg.StoreInterval,
		NestedFields:   rlcfg.NestedFields,
		Opts:           rlcfg.Opts.Clone(),
	}
	if rlcfg.ThresholdSConns != nil {
		cln.ThresholdSConns = utils.CloneStringSlice(rlcfg.ThresholdSConns)
	}

	if rlcfg.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rlcfg.StringIndexedFields))
	}
	if rlcfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rlcfg.PrefixIndexedFields))
	}
	if rlcfg.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(utils.CloneStringSlice(*rlcfg.SuffixIndexedFields))
	}
	return
}

type ResourcesOptsJson struct {
	UsageID  map[string]string  `json:"*usageID"`
	UsageTTL map[string]string  `json:"*usageTTL"`
	Units    map[string]float64 `json:"*units"`
}

// ResourceLimiter service config section
type ResourceSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Thresholds_conns      *[]string
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Opts                  *ResourcesOptsJson
}

func diffResourcesOptsJsonCfg(d *ResourcesOptsJson, v1, v2 *ResourcesOpts) *ResourcesOptsJson {
	if d == nil {
		d = new(ResourcesOptsJson)
	}
	if !utils.DynamicStringOptEqual(v1.UsageID, v2.UsageID) {
		d.UsageID = utils.DynamicStringOptsToMap(v2.UsageID)
	}
	if !utils.DynamicDurationOptEqual(v1.UsageTTL, v2.UsageTTL) {
		d.UsageTTL = utils.DynamicDurationOptsToMap(v2.UsageTTL)
	}
	if !utils.DynamicFloat64OptEqual(v1.Units, v2.Units) {
		d.Units = utils.DynamicFloat64OptsToMap(v2.Units)
	}
	return d
}

func diffResourceSJsonCfg(d *ResourceSJsonCfg, v1, v2 *ResourceSConfig) *ResourceSJsonCfg {
	if d == nil {
		d = new(ResourceSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if !utils.SliceStringEqual(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	d.Opts = diffResourcesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
