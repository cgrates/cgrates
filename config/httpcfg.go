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
	"github.com/cgrates/cgrates/utils"
)

// HTTPCfg is the HTTP config section
type HTTPCfg struct {
	JsonRPCURL        string            // JSON RPC relative URL ("" to disable)
	RegistrarSURL     string            // registrar service relative URL
	WSURL             string            // WebSocket relative URL ("" to disable)
	FreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	CDRsURL           string            // CDRS relative URL ("" to disable)
	UseBasicAuth      bool              // Use basic auth for HTTP API
	AuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
	ClientOpts        map[string]interface{}
}

// loadFromJSONCfg loads Database config from JsonCfg
func (httpcfg *HTTPCfg) loadFromJSONCfg(jsnHTTPCfg *HTTPJsonCfg) (err error) {
	if jsnHTTPCfg == nil {
		return nil
	}
	if jsnHTTPCfg.Json_rpc_url != nil {
		httpcfg.JsonRPCURL = *jsnHTTPCfg.Json_rpc_url
	}
	if jsnHTTPCfg.Registrars_url != nil {
		httpcfg.RegistrarSURL = *jsnHTTPCfg.Registrars_url
	}
	if jsnHTTPCfg.Ws_url != nil {
		httpcfg.WSURL = *jsnHTTPCfg.Ws_url
	}
	if jsnHTTPCfg.Freeswitch_cdrs_url != nil {
		httpcfg.FreeswitchCDRsURL = *jsnHTTPCfg.Freeswitch_cdrs_url
	}
	if jsnHTTPCfg.Http_Cdrs != nil {
		httpcfg.CDRsURL = *jsnHTTPCfg.Http_Cdrs
	}
	if jsnHTTPCfg.Use_basic_auth != nil {
		httpcfg.UseBasicAuth = *jsnHTTPCfg.Use_basic_auth
	}
	if jsnHTTPCfg.Auth_users != nil {
		httpcfg.AuthUsers = *jsnHTTPCfg.Auth_users
	}
	if jsnHTTPCfg.Client_opts != nil {
		for k, v := range jsnHTTPCfg.Client_opts {
			httpcfg.ClientOpts[k] = v
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (httpcfg *HTTPCfg) AsMapInterface() map[string]interface{} {
	clientOpts := make(map[string]interface{})
	for k, v := range httpcfg.ClientOpts {
		clientOpts[k] = v
	}
	return map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:        httpcfg.JsonRPCURL,
		utils.RegistrarSURLCfg:         httpcfg.RegistrarSURL,
		utils.HTTPWSURLCfg:             httpcfg.WSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.FreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.CDRsURL,
		utils.HTTPUseBasicAuthCfg:      httpcfg.UseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpcfg.AuthUsers,
		utils.HTTPClientOptsCfg:        clientOpts,
	}
}

// Clone returns a deep copy of HTTPCfg
func (httpcfg HTTPCfg) Clone() (cln *HTTPCfg) {
	cln = &HTTPCfg{
		JsonRPCURL:        httpcfg.JsonRPCURL,
		RegistrarSURL:     httpcfg.RegistrarSURL,
		WSURL:             httpcfg.WSURL,
		FreeswitchCDRsURL: httpcfg.FreeswitchCDRsURL,
		CDRsURL:           httpcfg.CDRsURL,
		UseBasicAuth:      httpcfg.UseBasicAuth,
		AuthUsers:         make(map[string]string),
		ClientOpts:        make(map[string]interface{}),
	}
	for u, a := range httpcfg.AuthUsers {
		cln.AuthUsers[u] = a
	}
	for o, val := range httpcfg.ClientOpts {
		cln.ClientOpts[o] = val
	}
	return
}

// HTTP config section
type HTTPJsonCfg struct {
	Json_rpc_url        *string
	Registrars_url      *string
	Ws_url              *string
	Freeswitch_cdrs_url *string
	Http_Cdrs           *string
	Use_basic_auth      *bool
	Auth_users          *map[string]string
	Client_opts         map[string]interface{}
}

func diffHTTPJsonCfg(d *HTTPJsonCfg, v1, v2 *HTTPCfg) *HTTPJsonCfg {
	if d == nil {
		d = new(HTTPJsonCfg)
	}

	if v1.JsonRPCURL != v2.JsonRPCURL {
		d.Json_rpc_url = utils.StringPointer(v2.JsonRPCURL)
	}
	if v1.RegistrarSURL != v2.RegistrarSURL {
		d.Registrars_url = utils.StringPointer(v2.RegistrarSURL)
	}
	if v1.WSURL != v2.WSURL {
		d.Ws_url = utils.StringPointer(v2.WSURL)
	}
	if v1.FreeswitchCDRsURL != v2.FreeswitchCDRsURL {
		d.Freeswitch_cdrs_url = utils.StringPointer(v2.FreeswitchCDRsURL)
	}
	if v1.CDRsURL != v2.CDRsURL {
		d.Http_Cdrs = utils.StringPointer(v2.CDRsURL)
	}
	if v1.UseBasicAuth != v2.UseBasicAuth {
		d.Use_basic_auth = utils.BoolPointer(v2.UseBasicAuth)
	}
	if !utils.MapStringStringEqual(v1.AuthUsers, v2.AuthUsers) {
		d.Auth_users = &v2.AuthUsers
	}
	d.Client_opts = diffMap(d.Client_opts, v1.ClientOpts, v2.ClientOpts)
	return d
}

func diffMap(d, v1, v2 map[string]interface{}) map[string]interface{} {
	if d == nil {
		d = make(map[string]interface{})
	}
	for k, v := range v2 {
		if val, has := v1[k]; !has || val != v {
			d[k] = v
		}
	}
	return d
}
