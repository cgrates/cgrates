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
	"maps"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	ResourcesUsageIDDftOpt  = utils.EmptyString
	ResourcesUsageTTLDftOpt = 72 * time.Hour
	ResourcesUnitsDftOpt    = 1
)

type ResourcesOpts struct {
	UsageID  []*DynamicStringOpt
	UsageTTL []*DynamicDurationOpt
	Units    []*DynamicFloat64Opt
}

// ResourceSCfg is resources section config
type ResourceSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	Conns                  map[string][]*DynamicConns
	StoreInterval          time.Duration // Dump regularly from cache into DB
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	Opts                   *ResourcesOpts
}

// Load loads the ResourceS section of the configuration
func (c *ResourceSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) error {
	jsnRLSCfg := new(ResourceSJsonCfg)
	if err := jsnCfg.GetSection(ctx, ResourceSJSON, jsnRLSCfg); err != nil {
		return err
	}
	return c.loadFromJSONCfg(jsnRLSCfg)
}

func (rsOpts *ResourcesOpts) loadFromJSONCfg(jsnCfg *ResourcesOptsJson) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.UsageID != nil {
		usageID, err := InterfaceToDynamicStringOpts(jsnCfg.UsageID)
		if err != nil {
			return err
		}
		rsOpts.UsageID = append(usageID, rsOpts.UsageID...)
	}
	if jsnCfg.UsageTTL != nil {
		usageTTL, err := IfaceToDurationDynamicOpts(jsnCfg.UsageTTL)
		if err != nil {
			return err
		}
		rsOpts.UsageTTL = append(usageTTL, rsOpts.UsageTTL...)
	}

	if jsnCfg.Units != nil {
		units, err := InterfaceToFloat64DynamicOpts(jsnCfg.Units)
		if err != nil {
			return err
		}
		rsOpts.Units = append(units, rsOpts.Units...)
	}
	return nil
}

func (c *ResourceSCfg) loadFromJSONCfg(jsnCfg *ResourceSJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		c.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		c.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Conns != nil {
		maps.Copy(c.Conns, tagConns(jsnCfg.Conns))
	}
	if jsnCfg.Store_interval != nil {
		ivl, err := utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval)
		if err != nil {
			return err
		}
		c.StoreInterval = ivl
	}
	if jsnCfg.String_indexed_fields != nil {
		c.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.String_indexed_fields))
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		c.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Prefix_indexed_fields))
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		c.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Suffix_indexed_fields))
	}
	if jsnCfg.Exists_indexed_fields != nil {
		c.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Exists_indexed_fields))
	}
	if jsnCfg.Notexists_indexed_fields != nil {
		c.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jsnCfg.Notexists_indexed_fields))
	}
	if jsnCfg.Nested_fields != nil {
		c.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		return c.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (c ResourceSCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaUsageIDCfg:  c.Opts.UsageID,
		utils.MetaUsageTTLCfg: c.Opts.UsageTTL,
		utils.MetaUnitsCfg:    c.Opts.Units,
	}
	mp := map[string]any{
		utils.EnabledCfg:        c.Enabled,
		utils.IndexedSelectsCfg: c.IndexedSelects,
		utils.ConnsCfg:          stripConns(c.Conns),
		utils.NestedFieldsCfg:   c.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.OptsCfg:           opts,
	}
	if c.StringIndexedFields != nil {
		mp[utils.StringIndexedFieldsCfg] = slices.Clone(*c.StringIndexedFields)
	}
	if c.PrefixIndexedFields != nil {
		mp[utils.PrefixIndexedFieldsCfg] = slices.Clone(*c.PrefixIndexedFields)
	}
	if c.SuffixIndexedFields != nil {
		mp[utils.SuffixIndexedFieldsCfg] = slices.Clone(*c.SuffixIndexedFields)
	}
	if c.ExistsIndexedFields != nil {
		mp[utils.ExistsIndexedFieldsCfg] = slices.Clone(*c.ExistsIndexedFields)
	}
	if c.NotExistsIndexedFields != nil {
		mp[utils.NotExistsIndexedFieldsCfg] = slices.Clone(*c.NotExistsIndexedFields)
	}
	if c.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = c.StoreInterval.String()
	}
	return mp
}

func (ResourceSCfg) SName() string           { return ResourceSJSON }
func (c ResourceSCfg) CloneSection() Section { return c.Clone() }

func (rsOpts *ResourcesOpts) Clone() *ResourcesOpts {
	var usageID []*DynamicStringOpt
	if rsOpts.UsageID != nil {
		usageID = CloneDynamicStringOpt(rsOpts.UsageID)
	}
	var usageTTL []*DynamicDurationOpt
	if rsOpts.UsageTTL != nil {
		usageTTL = CloneDynamicDurationOpt(rsOpts.UsageTTL)
	}
	var units []*DynamicFloat64Opt
	if rsOpts.Units != nil {
		units = CloneDynamicFloat64Opt(rsOpts.Units)
	}
	return &ResourcesOpts{
		UsageID:  usageID,
		UsageTTL: usageTTL,
		Units:    units,
	}
}

// Clone returns a deep copy of ResourceSCfg
func (c ResourceSCfg) Clone() *ResourceSCfg {
	cln := &ResourceSCfg{
		Enabled:        c.Enabled,
		IndexedSelects: c.IndexedSelects,
		Conns:          CloneConnsMap(c.Conns),
		StoreInterval:  c.StoreInterval,
		NestedFields:   c.NestedFields,
		Opts:           c.Opts.Clone(),
	}

	if c.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.SliceStringPointer(slices.Clone(*c.StringIndexedFields))
	}
	if c.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.SliceStringPointer(slices.Clone(*c.PrefixIndexedFields))
	}
	if c.SuffixIndexedFields != nil {
		cln.SuffixIndexedFields = utils.SliceStringPointer(slices.Clone(*c.SuffixIndexedFields))
	}
	if c.ExistsIndexedFields != nil {
		cln.ExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*c.ExistsIndexedFields))
	}
	if c.NotExistsIndexedFields != nil {
		cln.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*c.NotExistsIndexedFields))
	}
	return cln
}

type ResourcesOptsJson struct {
	UsageID  []*DynamicInterfaceOpt `json:"*usageID"`
	UsageTTL []*DynamicInterfaceOpt `json:"*usageTTL"`
	Units    []*DynamicInterfaceOpt `json:"*units"`
}

// ResourceSJsonCfg is the resources section JSON config
type ResourceSJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool                      `json:"indexedSelects"`
	Conns                    map[string][]*DynamicConns `json:"conns,omitempty"`
	Store_interval           *string                    `json:"storeInterval"`
	String_indexed_fields    *[]string                  `json:"stringIndexedFields"`
	Prefix_indexed_fields    *[]string                  `json:"prefixIndexedFields"`
	Suffix_indexed_fields    *[]string                  `json:"suffixIndexedFields"`
	Exists_indexed_fields    *[]string                  `json:"existsIndexedFields"`
	Notexists_indexed_fields *[]string                  `json:"notExistsIndexedFields"`
	Nested_fields            *bool                      `json:"nestedFields"` // applies when indexed fields is not defined
	Opts                     *ResourcesOptsJson
}

func diffResourcesOptsJsonCfg(d *ResourcesOptsJson, v1, v2 *ResourcesOpts) *ResourcesOptsJson {
	if d == nil {
		d = new(ResourcesOptsJson)
	}
	if !DynamicStringOptEqual(v1.UsageID, v2.UsageID) {
		d.UsageID = DynamicStringToInterfaceOpts(v2.UsageID)
	}
	if !DynamicDurationOptEqual(v1.UsageTTL, v2.UsageTTL) {
		d.UsageTTL = DurationToIfaceDynamicOpts(v2.UsageTTL)
	}
	if !DynamicFloat64OptEqual(v1.Units, v2.Units) {
		d.Units = Float64ToInterfaceDynamicOpts(v2.Units)
	}
	return d
}

func diffResourceSJsonCfg(d *ResourceSJsonCfg, v1, v2 *ResourceSCfg) *ResourceSJsonCfg {
	if d == nil {
		d = new(ResourceSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.Indexed_selects = utils.BoolPointer(v2.IndexedSelects)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	d.String_indexed_fields = diffIndexSlice(d.String_indexed_fields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.Prefix_indexed_fields = diffIndexSlice(d.Prefix_indexed_fields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.Suffix_indexed_fields = diffIndexSlice(d.Suffix_indexed_fields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	d.Exists_indexed_fields = diffIndexSlice(d.Exists_indexed_fields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.Notexists_indexed_fields = diffIndexSlice(d.Notexists_indexed_fields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.Nested_fields = utils.BoolPointer(v2.NestedFields)
	}
	d.Opts = diffResourcesOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
