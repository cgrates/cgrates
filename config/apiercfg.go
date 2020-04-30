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

import "github.com/cgrates/cgrates/utils"

// ApierCfg is the configuration of Apier service
type ApierCfg struct {
	Enabled         bool
	CachesConns     []string // connections towards Cache
	SchedulerConns  []string // connections towards Scheduler
	AttributeSConns []string // connections towards AttributeS
}

func (aCfg *ApierCfg) loadFromJsonCfg(jsnCfg *ApierJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Caches_conns != nil {
		aCfg.CachesConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, conn := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				aCfg.CachesConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			} else {
				aCfg.CachesConns[idx] = conn
			}
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		aCfg.SchedulerConns = make([]string, len(*jsnCfg.Scheduler_conns))
		for idx, conn := range *jsnCfg.Scheduler_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				aCfg.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			} else {
				aCfg.SchedulerConns[idx] = conn
			}
		}
	}
	if jsnCfg.Attributes_conns != nil {
		aCfg.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				aCfg.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				aCfg.AttributeSConns[idx] = conn
			}
		}
	}

	return nil
}

func (aCfg *ApierCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.EnabledCfg:         aCfg.Enabled,
		utils.CachesConnsCfg:     aCfg.CachesConns,
		utils.SchedulerConnsCfg:  aCfg.SchedulerConns,
		utils.AttributeSConnsCfg: aCfg.AttributeSConns,
	}

}
