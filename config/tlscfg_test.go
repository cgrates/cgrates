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

func TestTlsCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &TlsJsonCfg{
		Server_certificate: utils.StringPointer("path/To/Server/Cert"),
		Server_key:         utils.StringPointer("path/To/Server/Key"),
		Ca_certificate:     utils.StringPointer("path/To/CA/Cert"),
		Client_certificate: utils.StringPointer("path/To/Client/Cert"),
		Client_key:         utils.StringPointer("path/To/Client/Key"),
		Server_name:        utils.StringPointer("TestServerName"),
		Server_policy:      utils.IntPointer(3),
	}
	expected := &TLSCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.tlsCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.tlsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.tlsCfg))
	}
}

func TestTlsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `	{
	"tls": {},
}`
	eMap := map[string]interface{}{
		utils.ServerCerificateCfg: utils.EmptyString,
		utils.ServerKeyCfg:        utils.EmptyString,
		utils.ServerPolicyCfg:     4,
		utils.ServerNameCfg:       utils.EmptyString,
		utils.ClientCerificateCfg: utils.EmptyString,
		utils.ClientKeyCfg:        utils.EmptyString,
		utils.CaCertificateCfg:    utils.EmptyString,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.tlsCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestTlsCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `	{
	"tls":{
		"server_certificate" : "path/To/Server/Cert",			
		"server_key":"path/To/Server/Key",					
		"client_certificate" : "path/To/Client/Cert",			
		"client_key":"path/To/Client/Key",					
		"ca_certificate":"path/To/CA/Cert",							
		"server_name":"TestServerName",
		"server_policy":3,					
	},
}`
	eMap := map[string]interface{}{
		utils.ServerCerificateCfg: "path/To/Server/Cert",
		utils.ServerKeyCfg:        "path/To/Server/Key",
		utils.ServerPolicyCfg:     3,
		utils.ServerNameCfg:       "TestServerName",
		utils.ClientCerificateCfg: "path/To/Client/Cert",
		utils.ClientKeyCfg:        "path/To/Client/Key",
		utils.CaCertificateCfg:    "path/To/CA/Cert",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.tlsCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestTLSCfgClone(t *testing.T) {
	ban := &TLSCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ServerPolicy = 0; ban.ServerPolicy != 3 {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffTlsJsonCfg(t *testing.T) {
	var d *TlsJsonCfg

	v1 := &TLSCfg{
		ServerCerificate: "server_certificate",
		ServerKey:        "server_key",
		ServerPolicy:     1,
		ServerName:       "server_name",
		ClientCerificate: "client_certificate",
		ClientKey:        "client_key",
		CaCertificate:    "ca_certificate",
	}

	v2 := &TLSCfg{
		ServerCerificate: "server_certificate2",
		ServerKey:        "server_key2",
		ServerPolicy:     2,
		ServerName:       "server_name2",
		ClientCerificate: "client_certificate2",
		ClientKey:        "client_key2",
		CaCertificate:    "ca_certificate2",
	}

	expected := &TlsJsonCfg{
		Server_certificate: utils.StringPointer("server_certificate2"),
		Server_key:         utils.StringPointer("server_key2"),
		Server_policy:      utils.IntPointer(2),
		Server_name:        utils.StringPointer("server_name2"),
		Client_certificate: utils.StringPointer("client_certificate2"),
		Client_key:         utils.StringPointer("client_key2"),
		Ca_certificate:     utils.StringPointer("ca_certificate2"),
	}

	rcv := diffTlsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &TlsJsonCfg{}
	rcv = diffTlsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
