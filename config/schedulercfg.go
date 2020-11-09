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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type SchedulerCfg struct {
	Enabled      bool
	CDRsConns    []string
	ThreshSConns []string
	Filters      []string
}

func (schdcfg *SchedulerCfg) loadFromJSONCfg(jsnCfg *SchedulerJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		schdcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrs_conns != nil {
		schdcfg.CDRsConns = make([]string, len(*jsnCfg.Cdrs_conns))
		for idx, conn := range *jsnCfg.Cdrs_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				schdcfg.CDRsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)
			} else {
				schdcfg.CDRsConns[idx] = conn
			}
		}
	}
	if jsnCfg.Filters != nil {
		schdcfg.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			schdcfg.Filters[i] = fltr
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		schdcfg.ThreshSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, connID := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				schdcfg.ThreshSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				schdcfg.ThreshSConns[idx] = connID
			}
		}
	}
	return nil
}

func (schdcfg *SchedulerCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: schdcfg.Enabled,
		utils.FiltersCfg: schdcfg.Filters,
	}
	if schdcfg.CDRsConns != nil {
		cdrsConns := make([]string, len(schdcfg.CDRsConns))
		for i, item := range schdcfg.CDRsConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs) {
				cdrsConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaCDRs)
			} else {
				cdrsConns[i] = item
			}
		}
		initialMP[utils.CDRsConnsCfg] = cdrsConns
	}
	if schdcfg.ThreshSConns != nil {
		thrsConns := make([]string, len(schdcfg.ThreshSConns))
		for i, item := range schdcfg.ThreshSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thrsConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaThresholds)
			} else {
				thrsConns[i] = item
			}
		}
		initialMP[utils.ThreshSConnsCfg] = thrsConns
	}
	return
}
