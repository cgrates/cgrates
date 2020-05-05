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
	cachesConns := make([]string, len(aCfg.CachesConns))
	for i, item := range aCfg.CachesConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
		if item == buf {
			cachesConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaCaches, utils.EmptyString)
		} else {
			cachesConns[i] = item
		}
	}
	schedulerConns := make([]string, len(aCfg.SchedulerConns))
	for i, item := range aCfg.SchedulerConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
		if item == buf {
			schedulerConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaScheduler, utils.EmptyString)
		} else {
			schedulerConns[i] = item
		}
	}
	attributeSConns := make([]string, len(aCfg.AttributeSConns))
	for i, item := range aCfg.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:         aCfg.Enabled,
		utils.CachesConnsCfg:     cachesConns,
		utils.SchedulerConnsCfg:  schedulerConns,
		utils.AttributeSConnsCfg: attributeSConns,
	}

}
