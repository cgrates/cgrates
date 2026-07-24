// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

type LoaderCgrCfg struct {
	TpID           string
	DataPath       string
	DisableReverse bool
	FieldSeparator rune // The separator to use when reading csvs
	CachesConns    []string
	SchedulerConns []string
}

func (ld *LoaderCgrCfg) loadFromJsonCfg(jsnCfg *LoaderCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Tpid != nil {
		ld.TpID = *jsnCfg.Tpid
	}
	if jsnCfg.Data_path != nil {
		ld.DataPath = *jsnCfg.Data_path
	}
	if jsnCfg.Disable_reverse != nil {
		ld.DisableReverse = *jsnCfg.Disable_reverse
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 {
		sepStr := *jsnCfg.Field_separator
		ld.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Caches_conns != nil {
		ld.CachesConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, conn := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				ld.CachesConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			} else {
				ld.CachesConns[idx] = conn
			}
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		ld.SchedulerConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, conn := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				ld.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			} else {
				ld.SchedulerConns[idx] = conn
			}
		}
	}
	return nil
}

func (ld *LoaderCgrCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.TpIDCfg:           ld.TpID,
		utils.DataPathCfg:       ld.DataPath,
		utils.DisableReverseCfg: ld.DisableReverse,
		utils.FieldSeparatorCfg: string(ld.FieldSeparator),
		utils.CachesConnsCfg:    ld.CachesConns,
		utils.SchedulerConnsCfg: ld.SchedulerConns,
	}
}
