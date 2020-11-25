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
	"github.com/cgrates/cgrates/utils"
)

// SchedulerCfg the condig section for scheduler
type SchedulerCfg struct {
	Enabled      bool
	CDRsConns    []string
	ThreshSConns []string
	StatSConns   []string
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
			schdcfg.CDRsConns[idx] = conn
			if conn == utils.MetaInternal {
				schdcfg.CDRsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)
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
			schdcfg.ThreshSConns[idx] = connID
			if connID == utils.MetaInternal {
				schdcfg.ThreshSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		schdcfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			schdcfg.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				schdcfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (schdcfg *SchedulerCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: schdcfg.Enabled,
		utils.FiltersCfg: schdcfg.Filters,
	}
	if schdcfg.CDRsConns != nil {
		cdrsConns := make([]string, len(schdcfg.CDRsConns))
		for i, item := range schdcfg.CDRsConns {
			cdrsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs) {
				cdrsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CDRsConnsCfg] = cdrsConns
	}
	if schdcfg.ThreshSConns != nil {
		thrsConns := make([]string, len(schdcfg.ThreshSConns))
		for i, item := range schdcfg.ThreshSConns {
			thrsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thrsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThreshSConnsCfg] = thrsConns
	}
	if schdcfg.StatSConns != nil {
		stsConns := make([]string, len(schdcfg.StatSConns))
		for i, item := range schdcfg.StatSConns {
			stsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				stsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = stsConns
	}
	return
}

// Clone returns a deep copy of SchedulerCfg
func (schdcfg SchedulerCfg) Clone() (cln *SchedulerCfg) {
	cln = &SchedulerCfg{
		Enabled: schdcfg.Enabled,
	}
	if schdcfg.CDRsConns != nil {
		cln.CDRsConns = make([]string, len(schdcfg.CDRsConns))
		for i, con := range schdcfg.CDRsConns {
			cln.CDRsConns[i] = con
		}
	}
	if schdcfg.ThreshSConns != nil {
		cln.ThreshSConns = make([]string, len(schdcfg.ThreshSConns))
		for i, con := range schdcfg.ThreshSConns {
			cln.ThreshSConns[i] = con
		}
	}
	if schdcfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(schdcfg.StatSConns))
		for i, con := range schdcfg.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if schdcfg.Filters != nil {
		cln.Filters = make([]string, len(schdcfg.Filters))
		for i, con := range schdcfg.Filters {
			cln.Filters[i] = con
		}
	}

	return
}
