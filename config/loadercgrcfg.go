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

func (ld *LoaderCgrCfg) AsMapInterface() map[string]interface{} {
	gapiCredentials := make([]byte, len(ld.GapiCredentials))
	for i, item := range ld.GapiCredentials {
		gapiCredentials[i] = item
	}

	gapiToken := make([]byte, len(ld.GapiToken))
	for i, item := range ld.GapiToken {
		gapiToken[i] = item
	}

	return map[string]interface{}{
		utils.TpIDCfg:            ld.TpID,
		utils.DataPathCfg:        ld.DataPath,
		utils.DisableReverseCfg:  ld.DisableReverse,
		utils.FieldSeparatorCfg:  ld.FieldSeparator,
		utils.CachesConnsCfg:     ld.CachesConns,
		utils.SchedulerConnsCfg:  ld.SchedulerConns,
		utils.GapiCredentialsCfg: gapiCredentials,
		utils.GapiTokenCfg:       gapiToken,
	}
}
