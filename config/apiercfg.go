// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func (aCfg *ApierCfg) AsMapInterface() map[string]any {
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

	return map[string]any{
		utils.EnabledCfg:         aCfg.Enabled,
		utils.CachesConnsCfg:     cachesConns,
		utils.SchedulerConnsCfg:  schedulerConns,
		utils.AttributeSConnsCfg: attributeSConns,
	}

}
