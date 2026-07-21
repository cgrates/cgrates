// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"maps"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type RankingSCfg struct {
	Enabled        bool
	Conns          map[string][]*DynamicConns
	ScheduledIDs   map[string][]string
	StoreInterval  time.Duration
	EEsExporterIDs []string
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		if rnk.Conns == nil {
			rnk.Conns = make(map[string][]*DynamicConns)
		}
		maps.Copy(rnk.Conns, tagged)
	}
	if jsnCfg.Scheduled_ids != nil {
		rnk.ScheduledIDs = jsnCfg.Scheduled_ids
	}
	if jsnCfg.Store_interval != nil {
		if rnk.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return
		}
	}
	if jsnCfg.Ees_exporter_ids != nil {
		rnk.EEsExporterIDs = append(rnk.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	return
}

func (rnk *RankingSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:        rnk.Enabled,
		utils.StoreIntervalCfg:  utils.EmptyString,
		utils.EEsExporterIDsCfg: slices.Clone(rnk.EEsExporterIDs),
	}
	mp[utils.ConnsCfg] = stripConns(rnk.Conns)
	if rnk.ScheduledIDs != nil {
		mp[utils.ScheduledIDsCfg] = rnk.ScheduledIDs
	}
	if rnk.StoreInterval != 0 {
		mp[utils.StoreIntervalCfg] = rnk.StoreInterval.String()
	}
	return mp
}

func (RankingSCfg) SName() string             { return RankingSJSON }
func (rnk RankingSCfg) CloneSection() Section { return rnk.Clone() }

func (rnk *RankingSCfg) Clone() (cln *RankingSCfg) {
	cln = &RankingSCfg{
		Enabled:       rnk.Enabled,
		Conns:         CloneConnsMap(rnk.Conns),
		StoreInterval: rnk.StoreInterval,
	}
	if rnk.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range rnk.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
	}
	if rnk.EEsExporterIDs != nil {
		cln.EEsExporterIDs = slices.Clone(rnk.EEsExporterIDs)
	}
	return
}

type RankingSJsonCfg struct {
	Enabled          *bool
	Conns            map[string][]*DynamicConns `json:"conns,omitempty"`
	Scheduled_ids    map[string][]string        `json:"scheduledIDs"`
	Store_interval   *string                    `json:"storeInterval"`
	Ees_exporter_ids *[]string                  `json:"eesExporterIDs"`
}

func diffRankingsJsonCfg(d *RankingSJsonCfg, v1, v2 *RankingSCfg) *RankingSJsonCfg {
	if d == nil {
		d = new(RankingSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if v1.StoreInterval != v2.StoreInterval {
		d.Store_interval = utils.StringPointer(v2.StoreInterval.String())
	}
	if !slices.Equal(v1.EEsExporterIDs, v2.EEsExporterIDs) {
		d.Ees_exporter_ids = &v2.EEsExporterIDs
	}
	d.Scheduled_ids = diffMapStringSlice(d.Scheduled_ids, v1.ScheduledIDs, v2.ScheduledIDs)
	return d
}
