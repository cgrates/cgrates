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

func TestListenCfgloadFromJsonCfg(t *testing.T) {
	var lstcfg, expected ListenCfg
	if err := lstcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, lstcfg)
	}
	if err := lstcfg.loadFromJsonCfg(new(ListenJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(lstcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, lstcfg)
	}
	cfgJSONStr := `{
"listen": {
	"rpc_json": "127.0.0.1:2012",			// RPC JSON listening address
	"rpc_gob": "127.0.0.1:2013",			// RPC GOB listening address
	"http": "127.0.0.1:2080",				// HTTP listening address
	"rpc_json_tls" : "127.0.0.1:2022",		// RPC JSON TLS listening address
	"rpc_gob_tls": "127.0.0.1:2023",		// RPC GOB TLS listening address
	"http_tls": "127.0.0.1:2280",			// HTTP TLS listening address
	}
}`
	expected = ListenCfg{
		RPCJSONListen:    "127.0.0.1:2012",
		RPCGOBListen:     "127.0.0.1:2013",
		HTTPListen:       "127.0.0.1:2080",
		RPCJSONTLSListen: "127.0.0.1:2022",
		RPCGOBTLSListen:  "127.0.0.1:2023",
		HTTPTLSListen:    "127.0.0.1:2280",
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnLstCfg, err := jsnCfg.ListenJsonCfg(); err != nil {
		t.Error(err)
	} else if err = lstcfg.loadFromJsonCfg(jsnLstCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, lstcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, lstcfg)
	}
}
