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

	"github.com/cgrates/cgrates/utils"
)

type TrendSCfg struct {
	Enabled                bool
	StatSConns             []string
	ThresholdSConns        []string
	ScheduledIDs           map[string][]string
	StoreInterval          time.Duration
	StoreUncompressedLimit int
	EEsConns               []string
	EEsExporterIDs         []string
}

func (sa *TrendSCfg) loadFromJSONCfg(jsnCfg *TrendsJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		sa.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		sa.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			sa.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				sa.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		sa.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			sa.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				sa.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Scheduled_ids != nil {
		sa.ScheduledIDs = jsnCfg.Scheduled_ids
	}
	if jsnCfg.Store_interval != nil {
		if sa.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Store_uncompressed_limit != nil {
		sa.StoreUncompressedLimit = *jsnCfg.Store_uncompressed_limit
	}
	if jsnCfg.Ees_conns != nil {
		sa.EEsConns = make([]string, len(*jsnCfg.Ees_conns))
		for idx, connID := range *jsnCfg.Ees_conns {
			sa.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				sa.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	if jsnCfg.Ees_exporter_ids != nil {
		sa.EEsExporterIDs = append(sa.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	return
}

func (sa *TrendSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:                sa.Enabled,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.StoreUncompressedLimitCfg: sa.StoreUncompressedLimit,
	}
	if sa.StatSConns != nil {
		statSConns := make([]string, len(sa.StatSConns))
		for i, item := range sa.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if sa.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(sa.ThresholdSConns))
		for i, item := range sa.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if sa.ScheduledIDs != nil {
		initialMP[utils.ScheduledIDsCfg] = sa.ScheduledIDs
	}
	if sa.EEsConns != nil {
		eesConns := make([]string, len(sa.EEsConns))
		for i, item := range sa.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.EEsConnsCfg] = eesConns
	}
	eesExporterIDs := make([]string, len(sa.EEsExporterIDs))
	copy(eesExporterIDs, sa.EEsExporterIDs)
	initialMP[utils.EEsExporterIDsCfg] = eesExporterIDs
	return
}

func (sa *TrendSCfg) Clone() (cln *TrendSCfg) {
	cln = &TrendSCfg{
		Enabled:                sa.Enabled,
		StoreInterval:          sa.StoreInterval,
		StoreUncompressedLimit: sa.StoreUncompressedLimit,
	}
	if sa.StatSConns != nil {
		cln.StatSConns = make([]string, len(sa.StatSConns))
		copy(cln.StatSConns, sa.StatSConns)
	}
	if sa.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(sa.ThresholdSConns))
		copy(cln.ThresholdSConns, sa.ThresholdSConns)
	}
	if sa.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range sa.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
	}
	if sa.EEsConns != nil {
		cln.EEsConns = make([]string, len(sa.EEsConns))
		copy(cln.EEsConns, sa.EEsConns)
	}
	if sa.EEsExporterIDs != nil {
		cln.EEsExporterIDs = make([]string, len(sa.EEsExporterIDs))
		copy(cln.EEsExporterIDs, sa.EEsExporterIDs)
	}
	return

}
