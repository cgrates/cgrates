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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type TrendSCfg struct {
	Enabled                bool
	StoreInterval          time.Duration
	StoreUncompressedLimit int
	StatSConns             []string
	ScheduledIDs           map[string][]string
	ThresholdSConns        []string
	EEsConns               []string
	EEsExporterIDs         []string
}

func (t *TrendSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnTrendSCfg := new(TrendSJsonCfg)
	if err = jsnCfg.GetSection(ctx, TrendSJSON, jsnTrendSCfg); err != nil {
		return
	}
	return t.loadFromJSONCfg(jsnTrendSCfg)
}

func (t *TrendSCfg) loadFromJSONCfg(jsnCfg *TrendSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		t.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Store_interval != nil {
		if t.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.Store_uncompressed_limit != nil {
		t.StoreUncompressedLimit = *jsnCfg.Store_uncompressed_limit
	}
	if jsnCfg.Stats_conns != nil {
		t.StatSConns = tagInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Scheduled_ids != nil {
		t.ScheduledIDs = jsnCfg.Scheduled_ids
	}
	if jsnCfg.Thresholds_conns != nil {
		t.ThresholdSConns = tagInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Ees_conns != nil {
		t.EEsConns = tagInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCfg.Ees_exporter_ids != nil {
		t.EEsExporterIDs = append(t.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	return
}

func (t *TrendSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:                t.Enabled,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.StoreUncompressedLimitCfg: t.StoreUncompressedLimit,
		utils.EEsExporterIDsCfg:         slices.Clone(t.EEsExporterIDs),
	}
	if t.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = t.StoreInterval.String()
	}
	if t.StatSConns != nil {
		mp[utils.StatSConnsCfg] = stripInternalConns(t.StatSConns)
	}
	if t.ScheduledIDs != nil {
		mp[utils.ScheduledIDsCfg] = t.ScheduledIDs
	}
	if t.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = stripInternalConns(t.ThresholdSConns)
	}
	if t.EEsConns != nil {
		mp[utils.EEsConnsCfg] = stripInternalConns(t.EEsConns)
	}
	return mp
}

func (TrendSCfg) SName() string           { return TrendSJSON }
func (t TrendSCfg) CloneSection() Section { return t.Clone() }

func (t *TrendSCfg) Clone() (cln *TrendSCfg) {
	cln = &TrendSCfg{
		Enabled:                t.Enabled,
		StoreInterval:          t.StoreInterval,
		StoreUncompressedLimit: t.StoreUncompressedLimit,
	}
	if t.StatSConns != nil {
		cln.StatSConns = slices.Clone(t.StatSConns)
	}
	if t.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range t.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
	}
	if t.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(t.ThresholdSConns)
	}
	if t.EEsConns != nil {
		cln.EEsConns = slices.Clone(t.EEsConns)
	}
	if t.EEsExporterIDs != nil {
		cln.EEsExporterIDs = slices.Clone(t.EEsExporterIDs)
	}
	return
}

type TrendSJsonCfg struct {
	Enabled                  *bool
	Store_interval           *string
	Store_uncompressed_limit *int
	Stats_conns              *[]string
	Scheduled_ids            map[string][]string
	Thresholds_conns         *[]string
	Ees_conns                *[]string
	Ees_exporter_ids         *[]string
}

func diffTrendsJsonCfg(d *TrendSJsonCfg, v1, v2 *TrendSCfg) *TrendSJsonCfg {
	if d == nil {
		d = new(TrendSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	if v1.StoreUncompressedLimit != v2.StoreUncompressedLimit {
		d.Store_uncompressed_limit = utils.IntPointer(v2.StoreUncompressedLimit)
	}
	d.Scheduled_ids = diffMapStringSlice(d.Scheduled_ids, v1.ScheduledIDs, v2.ScheduledIDs)
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(stripInternalConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(stripInternalConns(v2.EEsConns))
	}
	if !slices.Equal(v1.EEsExporterIDs, v2.EEsExporterIDs) {
		d.Ees_exporter_ids = &v2.EEsExporterIDs
	}
	return d
}
