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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestHTTPCfgloadFromJsonCfg(t *testing.T) {
	var httpcfg, expected HTTPCfg
	if err := httpcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
	}
	if err := httpcfg.loadFromJsonCfg(new(HTTPJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(httpcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, httpcfg)
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
		t.Errorf("Expected: %+v , recived: %+v", expected, httpcfg)
	}
}

func TestHTTPCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"http": {},
}`
	eMap := map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:          "/jsonrpc",
		utils.DispatchersRegistrarURLCfg: "/dispatchers_registrar",
		utils.HTTPWSURLCfg:               "/ws",
		utils.HTTPFreeswitchCDRsURLCfg:   "/freeswitch_json",
		utils.HTTPCDRsURLCfg:             "/cdr_http",
		utils.HTTPUseBasicAuthCfg:        false,
		utils.HTTPAuthUsersCfg:           map[string]interface{}{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestHTTPCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"http": {
       "json_rpc_url": "/rpc",					
	   "ws_url": "",	
	   "use_basic_auth": true,					
	   "auth_users": {"user1": "authenticated", "user2": "authenticated"},
     },
}`
	eMap := map[string]interface{}{
		utils.HTTPJsonRPCURLCfg:          "/rpc",
		utils.DispatchersRegistrarURLCfg: "/dispatchers_registrar",
		utils.HTTPWSURLCfg:               "",
		utils.HTTPFreeswitchCDRsURLCfg:   "/freeswitch_json",
		utils.HTTPCDRsURLCfg:             "/cdr_http",
		utils.HTTPUseBasicAuthCfg:        true,
		utils.HTTPAuthUsersCfg: map[string]interface{}{
			"user1": "authenticated",
			"user2": "authenticated",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.httpCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
