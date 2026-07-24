// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// HTTP config section
type HTTPCfg struct {
	HTTPJsonRPCURL        string            // JSON RPC relative URL ("" to disable)
	HTTPWSURL             string            // WebSocket relative URL ("" to disable)
	HTTPFreeswitchCDRsURL string            // Freeswitch CDRS relative URL ("" to disable)
	HTTPCDRsURL           string            // CDRS relative URL ("" to disable)
	HTTPUseBasicAuth      bool              // Use basic auth for HTTP API
	HTTPAuthUsers         map[string]string // Basic auth user:password map (base64 passwords)
}

// loadFromJsonCfg loads Database config from JsonCfg
func (httpcfg *HTTPCfg) loadFromJsonCfg(jsnHttpCfg *HTTPJsonCfg) (err error) {
	if jsnHttpCfg == nil {
		return nil
	}
	if jsnHttpCfg.Json_rpc_url != nil {
		httpcfg.HTTPJsonRPCURL = *jsnHttpCfg.Json_rpc_url
	}
	if jsnHttpCfg.Ws_url != nil {
		httpcfg.HTTPWSURL = *jsnHttpCfg.Ws_url
	}
	if jsnHttpCfg.Freeswitch_cdrs_url != nil {
		httpcfg.HTTPFreeswitchCDRsURL = *jsnHttpCfg.Freeswitch_cdrs_url
	}
	if jsnHttpCfg.Http_Cdrs != nil {
		httpcfg.HTTPCDRsURL = *jsnHttpCfg.Http_Cdrs
	}
	if jsnHttpCfg.Use_basic_auth != nil {
		httpcfg.HTTPUseBasicAuth = *jsnHttpCfg.Use_basic_auth
	}
	if jsnHttpCfg.Auth_users != nil {
		httpcfg.HTTPAuthUsers = *jsnHttpCfg.Auth_users
	}

	return nil
}

func (httpcfg *HTTPCfg) AsMapInterface() map[string]any {
	httpUsers := make(map[string]any, len(httpcfg.HTTPAuthUsers))
	for key, item := range httpcfg.HTTPAuthUsers {
		httpUsers[key] = item
	}

	return map[string]any{
		utils.HTTPJsonRPCURLCfg:        httpcfg.HTTPJsonRPCURL,
		utils.HTTPWSURLCfg:             httpcfg.HTTPWSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.HTTPFreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.HTTPCDRsURL,
		utils.HTTPUseBasicAuthCfg:      httpcfg.HTTPUseBasicAuth,
		utils.HTTPAuthUsersCfg:         httpUsers,
	}
}
