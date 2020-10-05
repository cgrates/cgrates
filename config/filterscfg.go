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

type FilterSCfg struct {
	StatSConns     []string
	ResourceSConns []string
	ApierSConns    []string
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
	if jsnCfg.Apiers_conns != nil {
		fSCfg.ApierSConns = make([]string, len(*jsnCfg.Apiers_conns))
		for idx, connID := range *jsnCfg.Apiers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				fSCfg.ApierSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)
			} else {
				fSCfg.ApierSConns[idx] = connID
			}
		}
	}
	return
}

func (fSCfg *FilterSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = make(map[string]interface{})
	if fSCfg.StatSConns != nil {
		statSConns := make([]string, len(fSCfg.StatSConns))
		for i, item := range fSCfg.StatSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS) {
				statSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaStatS, utils.EmptyString)
			} else {
				statSConns[i] = item
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if fSCfg.ResourceSConns != nil {
		resourceSConns := make([]string, len(fSCfg.ResourceSConns))
		for i, item := range fSCfg.ResourceSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resourceSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaResources, utils.EmptyString)
			} else {
				resourceSConns[i] = item
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resourceSConns
	}
	if fSCfg.ApierSConns != nil {
		apierConns := make([]string, len(fSCfg.ApierSConns))
		for i, item := range fSCfg.ApierSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier) {
				apierConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaApier, utils.EmptyString)
			} else {
				apierConns[i] = item
			}
		}
		initialMP[utils.ApierSConnsCfg] = apierConns
	}
	return
}
