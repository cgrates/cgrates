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
	ActionSConns    []string
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
		ld.CachesConns = updateInternalConns(*jsnCfg.Caches_conns, utils.MetaCaches)
	}
	if jsnCfg.Actions_conns != nil {
		ld.ActionSConns = updateInternalConns(*jsnCfg.Actions_conns, utils.MetaActions)
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
		initialMP[utils.CachesConnsCfg] = getInternalJSONConns(ld.CachesConns)
	}
	if ld.ActionSConns != nil {
		initialMP[utils.ActionSConnsCfg] = getInternalJSONConns(ld.ActionSConns)
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
		cln.CachesConns = utils.CloneStringSlice(ld.CachesConns)
	}
	if ld.ActionSConns != nil {
		cln.ActionSConns = utils.CloneStringSlice(ld.ActionSConns)
	}
	return
}

type LoaderCfgJson struct {
	Tpid             *string
	Data_path        *string
	Disable_reverse  *bool
	Field_separator  *string
	Caches_conns     *[]string
	Actions_conns    *[]string
	Gapi_credentials *json.RawMessage
	Gapi_token       *json.RawMessage
}

func diffLoaderCfgJson(d *LoaderCfgJson, v1, v2 *LoaderCgrCfg) *LoaderCfgJson {
	if d == nil {
		d = new(LoaderCfgJson)
	}
	if v1.TpID != v2.TpID {
		d.Tpid = utils.StringPointer(v2.TpID)
	}
	if v1.DataPath != v2.DataPath {
		d.Data_path = utils.StringPointer(v2.DataPath)
	}
	if v1.DisableReverse != v2.DisableReverse {
		d.Disable_reverse = utils.BoolPointer(v2.DisableReverse)
	}
	if v1.FieldSeparator != v2.FieldSeparator {
		d.Field_separator = utils.StringPointer(string(v2.FieldSeparator))
	}
	if !utils.SliceStringEqual(v1.CachesConns, v2.CachesConns) {
		d.Caches_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CachesConns))
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	gc1 := string(v1.GapiCredentials)
	gc2 := string(v2.GapiCredentials)
	if gc1 != gc2 {
		if v2.GapiCredentials != nil {
			rw := json.RawMessage(gc2)
			d.Gapi_credentials = &rw
		} else {
			d.Gapi_credentials = nil
		}
	}
	gt1 := string(v1.GapiToken)
	gt2 := string(v2.GapiToken)
	if gt1 != gt2 {
		if v2.GapiToken != nil {
			rw := json.RawMessage(gt2)
			d.Gapi_token = &rw
		} else {
			d.Gapi_token = nil
		}
	}
	return d
}
