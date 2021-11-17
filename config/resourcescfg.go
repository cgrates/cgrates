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

	"github.com/cgrates/cgrates/utils"
)

type ResourcesOpts struct {
	UsageID  string
	UsageTTL *time.Duration
	Units    float64
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

func (resOpts *ResourcesOpts) loadFromJSONCfg(jsnCfg *ResourcesOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.UsageID != nil {
		resOpts.UsageID = *jsnCfg.UsageID
	}
	if jsnCfg.UsageTTL != nil {
		var ttl time.Duration
		if ttl, err = utils.ParseDurationWithNanosecs(*jsnCfg.UsageTTL); err != nil {
			return err
		}
		resOpts.UsageTTL = utils.DurationPointer(ttl)
	}
	if jsnCfg.Units != nil {
		resOpts.Units = *jsnCfg.Units
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
		rlcfg.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			rlcfg.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				rlcfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Store_interval != nil {
		if rlcfg.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		rlcfg.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		rlcfg.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		rlcfg.SuffixIndexedFields = &sif
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
func (rlcfg *ResourceSConfig) AsMapInterface() (initialMP map[string]interface{}) {
	opts := map[string]interface{}{
		utils.MetaUsageIDCfg: rlcfg.Opts.UsageID,
		utils.MetaUnitsCfg:   rlcfg.Opts.Units,
	}
	if rlcfg.Opts.UsageTTL != nil {
		opts[utils.MetaUsageTTLCfg] = *rlcfg.Opts.UsageTTL
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        rlcfg.Enabled,
		utils.IndexedSelectsCfg: rlcfg.IndexedSelects,
		utils.NestedFieldsCfg:   rlcfg.NestedFields,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.OptsCfg:           opts,
	}
	if rlcfg.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(rlcfg.ThresholdSConns))
		for i, item := range rlcfg.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if rlcfg.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*rlcfg.StringIndexedFields))
		for i, item := range *rlcfg.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if rlcfg.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*rlcfg.PrefixIndexedFields))
		for i, item := range *rlcfg.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if rlcfg.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*rlcfg.SuffixIndexedFields))
		for i, item := range *rlcfg.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if rlcfg.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = rlcfg.StoreInterval.String()
	}
	return
}

func (resOpts *ResourcesOpts) Clone() (cln *ResourcesOpts) {
	cln = &ResourcesOpts{
		UsageID: resOpts.UsageID,
		Units:   resOpts.Units,
	}
	if resOpts.UsageTTL != nil {
		cln.UsageTTL = utils.DurationPointer(*resOpts.UsageTTL)
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
		cln.ThresholdSConns = make([]string, len(rlcfg.ThresholdSConns))
		for i, con := range rlcfg.ThresholdSConns {
			cln.ThresholdSConns[i] = con
		}
	}

	if rlcfg.StringIndexedFields != nil {
		idx := make([]string, len(*rlcfg.StringIndexedFields))
		for i, dx := range *rlcfg.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if rlcfg.PrefixIndexedFields != nil {
		idx := make([]string, len(*rlcfg.PrefixIndexedFields))
		for i, dx := range *rlcfg.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if rlcfg.SuffixIndexedFields != nil {
		idx := make([]string, len(*rlcfg.SuffixIndexedFields))
		for i, dx := range *rlcfg.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
