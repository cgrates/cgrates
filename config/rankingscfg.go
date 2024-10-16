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

type RankingSCfg struct {
	Enabled         bool
	StatSConns      []string
	ThresholdSConns []string
	ScheduledIDs    map[string][]string
	StoreInterval   time.Duration
	EEsConns        []string
	EEsExporterIDs  []string
}

func (sgsCfg *RankingSCfg) loadFromJSONCfg(jsnCfg *RankingsJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		sgsCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Stats_conns != nil {
		sgsCfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			sgsCfg.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				sgsCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		sgsCfg.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			sgsCfg.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				sgsCfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Scheduled_ids != nil {
		sgsCfg.ScheduledIDs = jsnCfg.Scheduled_ids
	}
	if jsnCfg.Store_interval != nil {
		if sgsCfg.StoreInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Store_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Ees_conns != nil {
		sgsCfg.EEsConns = make([]string, len(*jsnCfg.Ees_conns))
		for idx, connID := range *jsnCfg.Ees_conns {
			sgsCfg.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				sgsCfg.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	if jsnCfg.Ees_exporter_ids != nil {
		sgsCfg.EEsExporterIDs = append(sgsCfg.EEsExporterIDs, *jsnCfg.Ees_exporter_ids...)
	}
	return
}

func (sgsCfg *RankingSCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:       sgsCfg.Enabled,
		utils.StoreIntervalCfg: utils.EmptyString,
	}
	if sgsCfg.StatSConns != nil {
		statSConns := make([]string, len(sgsCfg.StatSConns))
		for i, item := range sgsCfg.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if sgsCfg.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(sgsCfg.ThresholdSConns))
		for i, item := range sgsCfg.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if sgsCfg.ScheduledIDs != nil {
		initialMP[utils.ScheduledIDsCfg] = sgsCfg.ScheduledIDs
	}
	if sgsCfg.EEsConns != nil {
		eesConns := make([]string, len(sgsCfg.EEsConns))
		for i, item := range sgsCfg.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.EEsConnsCfg] = eesConns
	}
	eesExporterIDs := make([]string, len(sgsCfg.EEsExporterIDs))
	copy(eesExporterIDs, sgsCfg.EEsExporterIDs)
	initialMP[utils.EEsExporterIDsCfg] = eesExporterIDs
	return

}

func (sgscfg *RankingSCfg) Clone() (cln *RankingSCfg) {
	cln = &RankingSCfg{
		Enabled:       sgscfg.Enabled,
		StoreInterval: sgscfg.StoreInterval,
	}
	if sgscfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(sgscfg.StatSConns))
		copy(cln.StatSConns, sgscfg.StatSConns)
	}
	if sgscfg.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(sgscfg.ThresholdSConns))
		copy(cln.ThresholdSConns, sgscfg.ThresholdSConns)
	}
	if sgscfg.ScheduledIDs != nil {
		cln.ScheduledIDs = make(map[string][]string)
		for key, value := range sgscfg.ScheduledIDs {
			cln.ScheduledIDs[key] = slices.Clone(value)
		}
	}
	if sgscfg.EEsConns != nil {
		cln.EEsConns = make([]string, len(sgscfg.EEsConns))
		copy(cln.EEsConns, sgscfg.EEsConns)
	}
	if sgscfg.EEsExporterIDs != nil {
		cln.EEsExporterIDs = make([]string, len(sgscfg.EEsExporterIDs))
		copy(cln.EEsExporterIDs, sgscfg.EEsExporterIDs)
	}
	return
}
