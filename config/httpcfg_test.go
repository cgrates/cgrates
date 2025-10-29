/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
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
