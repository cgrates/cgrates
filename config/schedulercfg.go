// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

type SchedulerCfg struct {
	Enabled   bool
	CDRsConns []string
	Filters   []string
}

func (schdcfg *SchedulerCfg) loadFromJsonCfg(jsnCfg *SchedulerJsonCfg) error {
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
	return nil
}

func (schdcfg *SchedulerCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.EnabledCfg:   schdcfg.Enabled,
		utils.CDRsConnsCfg: schdcfg.CDRsConns,
		utils.FiltersCfg:   schdcfg.Filters,
	}
}
