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
	"strings"
	"testing"
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
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnhttpCfg, err := jsnCfg.HttpJsonCfg(); err != nil {
		t.Error(err)
	} else if err = httpcfg.loadFromJsonCfg(jsnhttpCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, httpcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, httpcfg)
	}
}
