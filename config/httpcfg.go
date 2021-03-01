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
	HTTPJsonRPCURL        string            // JSON RPC relative URL ("" to disable)
	RegistrarSURL         string            // registrar service relative URL
	HTTPWSURL             string            // WebSocket relative URL ("" to disable)
	HTTPFreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	HTTPCDRsURL           string            // CDRS relative URL ("" to disable)
	HTTPUseBasicAuth      bool              // Use basic auth for HTTP API
	HTTPAuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
	ClientOpts            map[string]interface{}
}

// loadFromJSONCfg loads Database config from JsonCfg
func (httpcfg *HTTPCfg) loadFromJSONCfg(jsnHTTPCfg *HTTPJsonCfg) (err error) {
	if jsnHTTPCfg == nil {
		return nil
	}
	if jsnHTTPCfg.Json_rpc_url != nil {
		httpcfg.HTTPJsonRPCURL = *jsnHTTPCfg.Json_rpc_url
	}
	if jsnHTTPCfg.Registrars_url != nil {
		httpcfg.RegistrarSURL = *jsnHTTPCfg.Registrars_url
	}
	if jsnHTTPCfg.Ws_url != nil {
		httpcfg.HTTPWSURL = *jsnHTTPCfg.Ws_url
	}
	if jsnHTTPCfg.Freeswitch_cdrs_url != nil {
		httpcfg.HTTPFreeswitchCDRsURL = *jsnHTTPCfg.Freeswitch_cdrs_url
	}
	if jsnHTTPCfg.Http_Cdrs != nil {
		httpcfg.HTTPCDRsURL = *jsnHTTPCfg.Http_Cdrs
	}
	if jsnHTTPCfg.Use_basic_auth != nil {
		httpcfg.HTTPUseBasicAuth = *jsnHTTPCfg.Use_basic_auth
	}
	if jsnHTTPCfg.Auth_users != nil {
		httpcfg.HTTPAuthUsers = *jsnHTTPCfg.Auth_users
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
		utils.HTTPJsonRPCURLCfg:        httpcfg.HTTPJsonRPCURL,
		utils.RegistrarSURLCfg:         httpcfg.RegistrarSURL,
		utils.HTTPWSURLCfg:             httpcfg.HTTPWSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.HTTPFreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.HTTPCDRsURL,
		utils.HTTPUseBasicAuthCfg:      httpcfg.HTTPUseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpcfg.HTTPAuthUsers,
		utils.HTTPClientOptsCfg:        clientOpts,
	}
}

// Clone returns a deep copy of HTTPCfg
func (httpcfg HTTPCfg) Clone() (cln *HTTPCfg) {
	cln = &HTTPCfg{
		HTTPJsonRPCURL:        httpcfg.HTTPJsonRPCURL,
		RegistrarSURL:         httpcfg.RegistrarSURL,
		HTTPWSURL:             httpcfg.HTTPWSURL,
		HTTPFreeswitchCDRsURL: httpcfg.HTTPFreeswitchCDRsURL,
		HTTPCDRsURL:           httpcfg.HTTPCDRsURL,
		HTTPUseBasicAuth:      httpcfg.HTTPUseBasicAuth,
		HTTPAuthUsers:         make(map[string]string),
		ClientOpts:            make(map[string]interface{}),
	}
	for u, a := range httpcfg.HTTPAuthUsers {
		cln.HTTPAuthUsers[u] = a
	}
	for o, val := range httpcfg.ClientOpts {
		cln.ClientOpts[o] = val
	}
	return
}
