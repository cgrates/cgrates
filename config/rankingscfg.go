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

type RankingSCfg struct {
	Enabled         bool
	StatSConns      []string
	ThresholdSConns []string
	ScheduledIDs    map[string][]string
	StoreInterval   time.Duration
	EEsConns        []string
	EEsExporterIDs  []string
}

func (rnk *RankingSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRankingSCfg := new(RankingSJsonCfg)
	if err = jsnCfg.GetSection(ctx, RankingSJSON, jsnRankingSCfg); err != nil {
		return
	}
	return rnk.loadFromJSONCfg(jsnRankingSCfg)
}

func (rnk *RankingSCfg) loadFromJSONCfg(jsnCfg *RankingSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		rnk.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		rnk.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Thresholds_conns != nil {
		rnk.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Scheduled_ids != nil {
		rnk.ScheduledIDs = jsnCfg.Scheduled_ids
	}
	if jsnCfg.Store_interval != nil {
		if rnk.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.Ees_conns != nil {
		rnk.EEsConns = updateInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	if jsnCfg.Ees_exporter_ids != nil {
		rnk.EEsExporterIDs = append(rnk.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	return
}

func (rnk *RankingSCfg) AsMapInterface(string) any {
	mp := map[string]any{
		utils.EnabledCfg:        rnk.Enabled,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.EEsExporterIDsCfg: slices.Clone(rnk.EEsExporterIDs),
	}
	if rnk.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(rnk.StatSConns)
	}
	if rnk.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(rnk.ThresholdSConns)
	}
	if rnk.ScheduledIDs != nil {
		mp[utils.ScheduledIDsCfg] = rnk.ScheduledIDs
	}
	if rnk.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = rnk.StoreInterval.String()
	}
	if rnk.EEsConns != nil {
		mp[utils.EEsConnsCfg] = getInternalJSONConns(rnk.EEsConns)
	}
	return mp
}

func (RankingSCfg) SName() string             { return RankingSJSON }
func (rnk RankingSCfg) CloneSection() Section { return rnk.Clone() }

func (rnk *RankingSCfg) Clone() (cln *RankingSCfg) {
	cln = &RankingSCfg{
		Enabled:       rnk.Enabled,
		StoreInterval: rnk.StoreInterval,
	}
	if rnk.StatSConns != nil {
		cln.StatSConns = slices.Clone(rnk.StatSConns)
	}
	if rnk.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(rnk.ThresholdSConns)
	}
	if rnk.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range rnk.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
	}
	if rnk.EEsConns != nil {
		cln.EEsConns = slices.Clone(rnk.EEsConns)
	}
	if rnk.EEsExporterIDs != nil {
		cln.EEsExporterIDs = slices.Clone(rnk.EEsExporterIDs)
	}
	return
}

type RankingSJsonCfg struct {
	Enabled          *bool
	Stats_conns      *[]string
	Thresholds_conns *[]string
	Scheduled_ids    map[string][]string
	Store_interval   *string
	Ees_conns        *[]string
	Ees_exporter_ids *[]string
}

func diffRankingsJsonCfg(d *RankingSJsonCfg, v1, v2 *RankingSCfg) *RankingSJsonCfg {
	if d == nil {
		d = new(RankingSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	if !slices.Equal(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	if !slices.Equal(v1.EEsExporterIDs, v2.EEsExporterIDs) {
		d.Ees_exporter_ids = &v2.EEsExporterIDs
	}
	d.Scheduled_ids = diffMapStringSlice(d.Scheduled_ids, v1.ScheduledIDs, v2.ScheduledIDs)
	return d
}
