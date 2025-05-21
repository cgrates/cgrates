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

	"github.com/cgrates/cgrates/utils"
)

// IPsCfg represents the configuration of the IPs module.
type IPsCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StoreInterval       time.Duration
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	ExistsIndexedFields *[]string
	NestedFields        bool
	Opts                *IPsOpts
}

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
	return clone
}

// AsMapInterface returns the ips config as a map[string]any.
func (c IPsCfg) AsMapInterface() any {
	return map[string]any{
		utils.EnabledCfg:             c.Enabled,
		utils.IndexedSelectsCfg:      c.IndexedSelects,
		utils.NestedFieldsCfg:        c.NestedFields,
		utils.StoreIntervalCfg:       c.StoreInterval.String(),
		utils.StringIndexedFieldsCfg: c.StringIndexedFields,
		utils.PrefixIndexedFieldsCfg: c.PrefixIndexedFields,
		utils.SuffixIndexedFieldsCfg: c.SuffixIndexedFields,
		utils.ExistsIndexedFieldsCfg: c.ExistsIndexedFields,
		utils.OptsCfg:                c.Opts.AsMapInterface(),
	}
}

type IPsOpts struct {
	UsageID string
	TTL     *time.Duration
	Units   float64
}

func (o *IPsOpts) loadFromJSONCfg(jc *IPsOptsJson) error {
	if jc == nil {
		return nil
	}
	if jc.UsageID != nil {
		o.UsageID = *jc.UsageID
	}
	if jc.TTL != nil {
		ttl, err := utils.ParseDurationWithNanosecs(*jc.TTL)
		if err != nil {
			return err
		}
		o.TTL = &ttl
	}
	if jc.Units != nil {
		o.Units = *jc.Units
	}
	return nil
}

// Clone returns a deep copy of IPsOpts.
func (o *IPsOpts) Clone() *IPsOpts {
	cln := &IPsOpts{
		UsageID: o.UsageID,
		Units:   o.Units,
	}
	if o.TTL != nil {
		cln.TTL = new(time.Duration)
		*cln.TTL = *o.TTL
	}
	return cln
}

// AsMapInterface returns the config as a map[string]any.
func (o *IPsOpts) AsMapInterface() map[string]any {
	m := map[string]any{
		utils.MetaUsageIDCfg: o.UsageID,
		utils.MetaUnitsCfg:   o.Units,
	}
	if o.TTL != nil {
		m[utils.MetaTTLCfg] = *o.TTL
	}
	return m
}
