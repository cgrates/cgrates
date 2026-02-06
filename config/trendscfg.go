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
	Conns                  map[string][]*DynamicStringSliceOpt
	ScheduledIDs           map[string][]string
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			t.Conns[connType] = opts
		}
	}
	if jsnCfg.Scheduled_ids != nil {
		t.ScheduledIDs = jsnCfg.Scheduled_ids
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
	mp[utils.ConnsCfg] = stripConns(t.Conns)
	if t.ScheduledIDs != nil {
		mp[utils.ScheduledIDsCfg] = t.ScheduledIDs
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
		Conns:                  CloneConnsOpt(t.Conns),
	}
	if t.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range t.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
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
	Conns                    map[string][]*DynamicStringSliceOpt `json:"conns,omitempty"`
	Scheduled_ids            map[string][]string
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
	if !ConnsEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	d.Scheduled_ids = diffMapStringSlice(d.Scheduled_ids, v1.ScheduledIDs, v2.ScheduledIDs)
	if !slices.Equal(v1.EEsExporterIDs, v2.EEsExporterIDs) {
		d.Ees_exporter_ids = &v2.EEsExporterIDs
	}
	return d
}
