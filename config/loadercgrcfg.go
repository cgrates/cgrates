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
	"encoding/json"

	"github.com/cgrates/cgrates/utils"
)

// LoaderCgrCfg the config for cgr-loader
type LoaderCgrCfg struct {
	TpID            string
	DataPath        string
	DisableReverse  bool
	FieldSeparator  rune // The separator to use when reading csvs
	CachesConns     []string
	SchedulerConns  []string
	GapiCredentials json.RawMessage
	GapiToken       json.RawMessage
}

func (ld *LoaderCgrCfg) loadFromJSONCfg(jsnCfg *LoaderCfgJson) (err error) {
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
			ld.CachesConns[idx] = conn
			if conn == utils.MetaInternal {
				ld.CachesConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			}
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		ld.SchedulerConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, conn := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ld.SchedulerConns[idx] = conn
			if conn == utils.MetaInternal {
				ld.SchedulerConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)
			}
		}
	}
	if jsnCfg.Gapi_credentials != nil {
		ld.GapiCredentials = *jsnCfg.Gapi_credentials
	}
	if jsnCfg.Gapi_token != nil {
		ld.GapiToken = *jsnCfg.Gapi_token
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (ld *LoaderCgrCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.TpIDCfg:           ld.TpID,
		utils.DataPathCfg:       ld.DataPath,
		utils.DisableReverseCfg: ld.DisableReverse,
		utils.FieldSepCfg:       string(ld.FieldSeparator),
	}
	if ld.CachesConns != nil {
		cacheSConns := make([]string, len(ld.CachesConns))
		for i, item := range ld.CachesConns {
			cacheSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cacheSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CachesConnsCfg] = cacheSConns
	}
	if ld.SchedulerConns != nil {
		schedulerSConns := make([]string, len(ld.SchedulerConns))
		for i, item := range ld.SchedulerConns {
			schedulerSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler) {
				schedulerSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SchedulerConnsCfg] = schedulerSConns
	}
	if ld.GapiCredentials != nil {
		initialMP[utils.GapiCredentialsCfg] = ld.GapiCredentials
	}
	if ld.GapiToken != nil {
		initialMP[utils.GapiTokenCfg] = ld.GapiToken
	}
	return
}

// Clone returns a deep copy of LoaderCgrCfg
func (ld LoaderCgrCfg) Clone() (cln *LoaderCgrCfg) {
	cln = &LoaderCgrCfg{
		TpID:            ld.TpID,
		DataPath:        ld.DataPath,
		DisableReverse:  ld.DisableReverse,
		FieldSeparator:  ld.FieldSeparator,
		GapiCredentials: json.RawMessage(string([]byte(ld.GapiCredentials))),
		GapiToken:       json.RawMessage(string([]byte(ld.GapiToken))),
	}

	if ld.CachesConns != nil {
		cln.CachesConns = make([]string, len(ld.CachesConns))
		for i, k := range ld.CachesConns {
			cln.CachesConns[i] = k
		}
	}
	if ld.SchedulerConns != nil {
		cln.SchedulerConns = make([]string, len(ld.SchedulerConns))
		for i, k := range ld.SchedulerConns {
			cln.SchedulerConns[i] = k
		}
	}
	return
}
