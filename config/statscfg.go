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

type StatsOpts struct {
	ProfileIDs           []string
	ProfileIgnoreFilters bool
}

// StatSCfg the stats config section
type StatSCfg struct {
	Enabled                bool
	IndexedSelects         bool
	StoreInterval          time.Duration // Dump regularly from cache into dataDB
	StoreUncompressedLimit int
	ThresholdSConns        []string
	StringIndexedFields    *[]string
	PrefixIndexedFields    *[]string
	SuffixIndexedFields    *[]string
	NestedFields           bool
	Opts                   *StatsOpts
	EEsConns               []string
	EEsExporterIDs         []string
}

func (sqOpts *StatsOpts) loadFromJSONCfg(jsnCfg *StatsOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ProfileIDs != nil {
		sqOpts.ProfileIDs = *jsnCfg.ProfileIDs
	}
	if jsnCfg.ProfileIgnoreFilters != nil {
		sqOpts.ProfileIgnoreFilters = *jsnCfg.ProfileIgnoreFilters
	}
}

func (st *StatSCfg) loadFromJSONCfg(jsnCfg *StatServJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		st.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		st.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Store_interval != nil {
		if st.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Store_uncompressed_limit != nil {
		st.StoreUncompressedLimit = *jsnCfg.Store_uncompressed_limit
	}
	if jsnCfg.Thresholds_conns != nil {
		st.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			st.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				st.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Ees_conns != nil {
		st.EEsConns = make([]string, len(*jsnCfg.Ees_conns))
		for idx, connID := range *jsnCfg.Ees_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			st.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				st.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	if jsnCfg.Ees_exporter_ids != nil {
		st.EEsExporterIDs = append(st.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		copy(sif, *jsnCfg.String_indexed_fields)
		st.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		copy(pif, *jsnCfg.Prefix_indexed_fields)
		st.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		copy(sif, *jsnCfg.Suffix_indexed_fields)
		st.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		st.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Opts != nil {
		st.Opts.loadFromJSONCfg(jsnCfg.Opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (st *StatSCfg) AsMapInterface() (initialMP map[string]any) {
	opts := map[string]any{
		utils.MetaProfileIDs:              st.Opts.ProfileIDs,
		utils.MetaProfileIgnoreFiltersCfg: st.Opts.ProfileIgnoreFilters,
	}
	initialMP = map[string]any{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.NestedFieldsCfg:           st.NestedFields,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.OptsCfg:                   opts,
	}
	if st.StoreInterval != 0 {
		initialMP[utils.StoreIntervalCfg] = st.StoreInterval.String()
	}

	eesExporterIDs := make([]string, len(st.EEsExporterIDs))
	copy(eesExporterIDs, st.EEsExporterIDs)

	initialMP[utils.EEsExporterIDsCfg] = eesExporterIDs

	if st.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*st.StringIndexedFields))
		copy(stringIndexedFields, *st.StringIndexedFields)

		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if st.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*st.PrefixIndexedFields))
		copy(prefixIndexedFields, *st.PrefixIndexedFields)

		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if st.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*st.SuffixIndexedFields))
		copy(suffixIndexedFields, *st.SuffixIndexedFields)
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields

	}
	if st.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(st.ThresholdSConns))
		for i, item := range st.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if st.EEsConns != nil {
		eesConns := make([]string, len(st.EEsConns))
		for i, item := range st.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.EEsConnsCfg] = eesConns
	}
	return
}

func (stOpts *StatsOpts) Clone() *StatsOpts {
	return &StatsOpts{
		ProfileIDs:           slices.Clone(stOpts.ProfileIDs),
		ProfileIgnoreFilters: stOpts.ProfileIgnoreFilters,
	}
}

// Clone returns a deep copy of StatSCfg
func (st StatSCfg) Clone() (cln *StatSCfg) {
	cln = &StatSCfg{
		Enabled:                st.Enabled,
		IndexedSelects:         st.IndexedSelects,
		StoreInterval:          st.StoreInterval,
		StoreUncompressedLimit: st.StoreUncompressedLimit,
		NestedFields:           st.NestedFields,
		Opts:                   st.Opts.Clone(),
	}
	if st.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(st.ThresholdSConns))
		copy(cln.ThresholdSConns, st.ThresholdSConns)
	}
	if st.EEsConns != nil {
		cln.EEsConns = make([]string, len(st.EEsConns))
		copy(cln.EEsConns, st.EEsConns)
	}
	if st.EEsExporterIDs != nil {
		cln.EEsExporterIDs = make([]string, len(st.EEsExporterIDs))
		copy(cln.EEsExporterIDs, st.EEsExporterIDs)
	}
	if st.StringIndexedFields != nil {
		idx := make([]string, len(*st.StringIndexedFields))
		copy(idx, *st.StringIndexedFields)
		cln.StringIndexedFields = &idx
	}
	if st.PrefixIndexedFields != nil {
		idx := make([]string, len(*st.PrefixIndexedFields))
		copy(idx, *st.PrefixIndexedFields)
		cln.PrefixIndexedFields = &idx
	}
	if st.SuffixIndexedFields != nil {
		idx := make([]string, len(*st.SuffixIndexedFields))
		copy(idx, *st.SuffixIndexedFields)
		cln.SuffixIndexedFields = &idx
	}
	return
}
