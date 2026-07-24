// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

type FilterSCfg struct {
	StatSConns     []string
	ResourceSConns []string
}

func (fSCfg *FilterSCfg) loadFromJsonCfg(jsnCfg *FilterSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Stats_conns != nil {
		fSCfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				fSCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
			} else {
				fSCfg.StatSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		fSCfg.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, connID := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				fSCfg.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			} else {
				fSCfg.ResourceSConns[idx] = connID
			}
		}
	}
	return
}

func (fSCfg *FilterSCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.StatSConnsCfg:     fSCfg.StatSConns,
		utils.ResourceSConnsCfg: fSCfg.ResourceSConns,
	}
}
