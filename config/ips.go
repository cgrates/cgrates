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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	IPsUsageIDDftOpt = utils.EmptyString
	IPsTTLDftOpt     = 72 * time.Hour
	IPsUnitsDftOpt   = 1
)

// IPsJsonCfg holds the unparsed ips section configuration as found in the
// config file.
type IPsJsonCfg struct {
	Enabled                *bool        `json:"enabled"`
	IndexedSelects         *bool        `json:"indexed_selects"`
	StoreInterval          *string      `json:"store_interval"`
	StringIndexedFields    *[]string    `json:"string_indexed_fields"`
	PrefixIndexedFields    *[]string    `json:"prefix_indexed_fields"`
	SuffixIndexedFields    *[]string    `json:"suffix_indexed_fields"`
	ExistsIndexedFields    *[]string    `json:"exists_indexed_fields"`
	NotExistsIndexedFields *[]string    `json:"not_exists_indexed_fields"`
	NestedFields           *bool        `json:"nested_fields"`
	Opts                   *IPsOptsJson `json:"opts"`
}

// IPsCfg represents the configuration of the IPs module.
type IPsCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StoreInterval          time.Duration
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	ExistsIndexedFields    *[]string
	NotExistsIndexedFields *[]string
	NestedFields           bool
	Opts                   *IPsOpts
}

// Load loads the IPs section of the configuration.
func (rlcfg *IPsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRLSCfg := new(IPsJsonCfg)
	if err = jsnCfg.GetSection(ctx, IPsJSON, jsnRLSCfg); err != nil {
		return
	}
	return rlcfg.loadFromJSONCfg(jsnRLSCfg)
}

func (IPsCfg) SName() string           { return IPsJSON }
func (c IPsCfg) CloneSection() Section { return c.Clone() }

func (c *IPsCfg) loadFromJSONCfg(jc *IPsJsonCfg) error {
	if jc == nil {
		return nil
	}
	if jc.Enabled != nil {
		c.Enabled = *jc.Enabled
	}
	if jc.IndexedSelects != nil {
		c.IndexedSelects = *jc.IndexedSelects
	}
	if jc.StoreInterval != nil {
		v, err := utils.ParseDurationWithNanosecs(*jc.StoreInterval)
		if err != nil {
			return err
		}
		c.StoreInterval = v
	}
	if jc.StringIndexedFields != nil {
		sif := slices.Clone(*jc.StringIndexedFields)
		c.StringIndexedFields = &sif
	}
	if jc.PrefixIndexedFields != nil {
		pif := slices.Clone(*jc.PrefixIndexedFields)
		c.PrefixIndexedFields = &pif
	}
	if jc.SuffixIndexedFields != nil {
		sif := slices.Clone(*jc.SuffixIndexedFields)
		c.SuffixIndexedFields = &sif
	}
	if jc.ExistsIndexedFields != nil {
		eif := slices.Clone(*jc.ExistsIndexedFields)
		c.ExistsIndexedFields = &eif
	}
	if jc.NotExistsIndexedFields != nil {
		c.NotExistsIndexedFields = utils.SliceStringPointer(slices.Clone(*jc.NotExistsIndexedFields))
	}
	if jc.NestedFields != nil {
		c.NestedFields = *jc.NestedFields
	}
	if jc.Opts != nil {
		if err := c.Opts.loadFromJSONCfg(jc.Opts); err != nil {
			return err
		}
	}
	return nil
}

// Clone returns a deep copy of IPsCfg.
func (c IPsCfg) Clone() *IPsCfg {
	clone := &IPsCfg{
		Enabled:        c.Enabled,
		IndexedSelects: c.IndexedSelects,
		StoreInterval:  c.StoreInterval,
		NestedFields:   c.NestedFields,
		Opts:           c.Opts.Clone(),
	}
	if c.StringIndexedFields != nil {
		idx := slices.Clone(*c.StringIndexedFields)
		clone.StringIndexedFields = &idx
	}
	if c.PrefixIndexedFields != nil {
		idx := slices.Clone(*c.PrefixIndexedFields)
		clone.PrefixIndexedFields = &idx
	}
	if c.SuffixIndexedFields != nil {
		idx := slices.Clone(*c.SuffixIndexedFields)
		clone.SuffixIndexedFields = &idx
	}
	if c.ExistsIndexedFields != nil {
		idx := slices.Clone(*c.ExistsIndexedFields)
		clone.ExistsIndexedFields = &idx
	}
	if c.NotExistsIndexedFields != nil {
		idx := slices.Clone(*c.NotExistsIndexedFields)
		clone.NotExistsIndexedFields = &idx
	}
	return clone
}

// AsMapInterface returns the ips config as a map[string]any.
func (c IPsCfg) AsMapInterface() any {
	return map[string]any{
		utils.EnabledCfg:                c.Enabled,
		utils.IndexedSelectsCfg:         c.IndexedSelects,
		utils.NestedFieldsCfg:           c.NestedFields,
		utils.StoreIntervalCfg:          c.StoreInterval.String(),
		utils.StringIndexedFieldsCfg:    c.StringIndexedFields,
		utils.PrefixIndexedFieldsCfg:    c.PrefixIndexedFields,
		utils.SuffixIndexedFieldsCfg:    c.SuffixIndexedFields,
		utils.ExistsIndexedFieldsCfg:    c.ExistsIndexedFields,
		utils.NotExistsIndexedFieldsCfg: c.NotExistsIndexedFields,
		utils.OptsCfg:                   c.Opts.AsMapInterface(),
	}
}

type IPsOptsJson struct {
	UsageID []*DynamicInterfaceOpt `json:"*usageID"`
	TTL     []*DynamicInterfaceOpt `json:"*ttl"`
	Units   []*DynamicInterfaceOpt `json:"*units"`
}

type IPsOpts struct {
	UsageID []*DynamicStringOpt
	TTL     []*DynamicDurationOpt
	Units   []*DynamicFloat64Opt
}

func (o *IPsOpts) loadFromJSONCfg(jc *IPsOptsJson) error {
	if jc == nil {
		return nil
	}

	// NOTE: prepend to the existing slice to ensure that the default opts that
	// always match are at the end.
	if jc.UsageID != nil {
		usageID, err := InterfaceToDynamicStringOpts(jc.UsageID)
		if err != nil {
			return err
		}
		o.UsageID = append(usageID, o.UsageID...)
	}
	if jc.TTL != nil {
		ttl, err := IfaceToDurationDynamicOpts(jc.TTL)
		if err != nil {
			return err
		}
		o.TTL = append(ttl, o.TTL...)
	}
	if jc.Units != nil {
		units, err := InterfaceToFloat64DynamicOpts(jc.Units)
		if err != nil {
			return err
		}
		o.Units = append(units, o.Units...)
	}
	return nil
}

// Clone returns a deep copy of IPsOpts.
func (o *IPsOpts) Clone() *IPsOpts {
	return &IPsOpts{
		UsageID: CloneDynamicStringOpt(o.UsageID),
		TTL:     CloneDynamicDurationOpt(o.TTL),
		Units:   CloneDynamicFloat64Opt(o.Units),
	}
}

// AsMapInterface returns the config as a map[string]any.
func (o *IPsOpts) AsMapInterface() map[string]any {
	return map[string]any{
		utils.MetaUsageIDCfg: o.UsageID,
		utils.MetaUnitsCfg:   o.Units,
		utils.MetaTTLCfg:     o.TTL,
	}
}

func diffIPsOptsJsonCfg(d *IPsOptsJson, v1, v2 *IPsOpts) *IPsOptsJson {
	if d == nil {
		d = new(IPsOptsJson)
	}
	if !DynamicStringOptEqual(v1.UsageID, v2.UsageID) {
		d.UsageID = DynamicStringToInterfaceOpts(v2.UsageID)
	}
	if !DynamicDurationOptEqual(v1.TTL, v2.TTL) {
		d.TTL = DurationToIfaceDynamicOpts(v2.TTL)
	}
	if !DynamicFloat64OptEqual(v1.Units, v2.Units) {
		d.Units = Float64ToInterfaceDynamicOpts(v2.Units)
	}
	return d
}

func diffIPsJsonCfg(d *IPsJsonCfg, v1, v2 *IPsCfg) *IPsJsonCfg {
	if d == nil {
		d = new(IPsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.IndexedSelects != v2.IndexedSelects {
		d.IndexedSelects = utils.BoolPointer(v2.IndexedSelects)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.StoreInterval = utils.StringPointer(v2.StoreInterval.String())
	}
	d.StringIndexedFields = diffIndexSlice(d.StringIndexedFields, v1.StringIndexedFields, v2.StringIndexedFields)
	d.PrefixIndexedFields = diffIndexSlice(d.PrefixIndexedFields, v1.PrefixIndexedFields, v2.PrefixIndexedFields)
	d.SuffixIndexedFields = diffIndexSlice(d.SuffixIndexedFields, v1.SuffixIndexedFields, v2.SuffixIndexedFields)
	d.ExistsIndexedFields = diffIndexSlice(d.ExistsIndexedFields, v1.ExistsIndexedFields, v2.ExistsIndexedFields)
	d.NotExistsIndexedFields = diffIndexSlice(d.NotExistsIndexedFields, v1.NotExistsIndexedFields, v2.NotExistsIndexedFields)
	if v1.NestedFields != v2.NestedFields {
		d.NestedFields = utils.BoolPointer(v2.NestedFields)
	}
	d.Opts = diffIPsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
