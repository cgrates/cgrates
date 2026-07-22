// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/cgrates/cgrates/utils"
)

// ApierCfg is the configuration of Apier service
type ApierCfg struct {
	Enabled         bool
	CachesConns     []string // connections towards Cache
	SchedulerConns  []string // connections towards Scheduler
	AttributeSConns []string // connections towards AttributeS
	EEsConns        []string // connections towards EEs
}

func (aCfg *ApierCfg) loadFromJSONCfg(jsnCfg *ApierJsonCfg) (err error) {
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
			aCfg.CachesConns[idx] = conn
			if conn == utils.MetaInternal {
				aCfg.CachesConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			}
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		aCfg.SchedulerConns = make([]string, len(*jsnCfg.Scheduler_conns))
		for idx, conn := range *jsnCfg.Scheduler_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.SchedulerConns[idx] = conn
			if conn == utils.MetaInternal {
				aCfg.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			}
		}
	}
	if jsnCfg.Attributes_conns != nil {
		aCfg.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.AttributeSConns[idx] = conn
			if conn == utils.MetaInternal {
				aCfg.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Ees_conns != nil {
		aCfg.EEsConns = make([]string, len(*jsnCfg.Ees_conns))
		for idx, connID := range *jsnCfg.Ees_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				aCfg.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (aCfg *ApierCfg) AsMapInterface() (initialMap map[string]any) {
	initialMap = map[string]any{
		utils.EnabledCfg: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		cachesConns := make([]string, len(aCfg.CachesConns))
		for i, item := range aCfg.CachesConns {
			cachesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cachesConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.CachesConnsCfg] = cachesConns
	}
	if aCfg.SchedulerConns != nil {
		schedulerConns := make([]string, len(aCfg.SchedulerConns))
		for i, item := range aCfg.SchedulerConns {
			schedulerConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler) {
				schedulerConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.SchedulerConnsCfg] = schedulerConns
	}
	if aCfg.AttributeSConns != nil {
		attributeSConns := make([]string, len(aCfg.AttributeSConns))
		for i, item := range aCfg.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.AttributeSConnsCfg] = attributeSConns
	}
	if aCfg.EEsConns != nil {
		eesConns := make([]string, len(aCfg.EEsConns))
		for i, item := range aCfg.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.EEsConnsCfg] = eesConns
	}
	return
}

// Clone returns a deep copy of ApierCfg
func (aCfg *ApierCfg) Clone() (cln *ApierCfg) {
	if aCfg == nil {
		return nil
	}
	cln = &ApierCfg{
		Enabled: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		cln.CachesConns = make([]string, len(aCfg.CachesConns))
		copy(cln.CachesConns, aCfg.CachesConns)
	}
	if aCfg.SchedulerConns != nil {
		cln.SchedulerConns = make([]string, len(aCfg.SchedulerConns))
		copy(cln.SchedulerConns, aCfg.SchedulerConns)
	}
	if aCfg.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(aCfg.AttributeSConns))
		copy(cln.AttributeSConns, aCfg.AttributeSConns)
	}
	if aCfg.EEsConns != nil {
		cln.EEsConns = make([]string, len(aCfg.EEsConns))
		copy(cln.EEsConns, aCfg.EEsConns)
	}
	return
}
