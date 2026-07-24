// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestHTTPCfgloadFromJsonCfg(t *testing.T) {
	var httpcfg, expected HTTPCfg
	if err := httpcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, httpcfg)
	}
	if err := httpcfg.loadFromJsonCfg(new(HTTPJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, httpcfg)
	}
	cfgJSONStr := `{
"http": {										// HTTP server configuration
	"json_rpc_url": "/jsonrpc",					// JSON RPC relative URL ("" to disable)
	"ws_url": "/ws",							// WebSockets relative URL ("" to disable)
	"freeswitch_cdrs_url": "/freeswitch_json",	// Freeswitch CDRS relative URL ("" to disable)
	"http_cdrs": "/cdr_http",					// CDRS relative URL ("" to disable)
	"use_basic_auth": false,					// use basic authentication
	"auth_users": {},							// basic authentication usernames and base64-encoded passwords (eg: { "username1": "cGFzc3dvcmQ=", "username2": "cGFzc3dvcmQy "})
	},
}`
	expected = HTTPCfg{
		HTTPJsonRPCURL:        "/jsonrpc",
		HTTPWSURL:             "/ws",
		HTTPFreeswitchCDRsURL: "/freeswitch_json",
		HTTPCDRsURL:           "/cdr_http",
		HTTPUseBasicAuth:      false,
		HTTPAuthUsers:         map[string]string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnhttpCfg, err := jsnCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if err = httpcfg.loadFromJsonCfg(jsnhttpCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, httpcfg)
	}
}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	var httpcfg HTTPCfg
	cfgJSONStr := `{
	"http": {										
		"json_rpc_url": "/jsonrpc",					
		"ws_url": "/ws",							
		"freeswitch_cdrs_url": "/freeswitch_json",	
		"http_cdrs": "/cdr_http",					
		"use_basic_auth": false,					
		"auth_users": {},							
	},
}`

	eMap := map[string]any{
		"json_rpc_url":        "/jsonrpc",
		"ws_url":              "/ws",
		"freeswitch_cdrs_url": "/freeswitch_json",
		"http_cdrs":           "/cdr_http",
		"use_basic_auth":      false,
		"auth_users":          map[string]any{},
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnhttpCfg, err := jsnCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if err = httpcfg.loadFromJsonCfg(jsnhttpCfg); err != nil {
		t.Error(err)
	} else if rcv := httpcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v ,\n received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestHTTPCfgAsMapInterface2(t *testing.T) {
	str := "test"
	bl := true
	mp := map[string]string{"test1": "val1", "test": "val2"}

	httpcfg := HTTPCfg{
		HTTPJsonRPCURL:        str,
		HTTPWSURL:             str,
		HTTPFreeswitchCDRsURL: str,
		HTTPCDRsURL:           str,
		HTTPUseBasicAuth:      bl,
		HTTPAuthUsers:         mp,
	}

	exp := map[string]any{
		utils.HTTPJsonRPCURLCfg:        httpcfg.HTTPJsonRPCURL,
		utils.HTTPWSURLCfg:             httpcfg.HTTPWSURL,
		utils.HTTPFreeswitchCDRsURLCfg: httpcfg.HTTPFreeswitchCDRsURL,
		utils.HTTPCDRsURLCfg:           httpcfg.HTTPCDRsURL,
		utils.HTTPUseBasicAuthCfg:      httpcfg.HTTPUseBasicAuth,
		utils.HTTPAuthUsersCfg:         map[string]any{"test1": "val1", "test": "val2"},
	}

	rcv := httpcfg.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected: %+v ,\n received: %+v", exp, rcv)
	}
}
