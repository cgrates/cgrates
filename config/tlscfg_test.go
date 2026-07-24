// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestTlsCfgloadFromJsonCfg(t *testing.T) {
	var tlscfg, expected TlsCfg
	if err := tlscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tlscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, tlscfg)
	}
	if err := tlscfg.loadFromJsonCfg(new(TlsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tlscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, tlscfg)
	}
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
	expected = TlsCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, tlscfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, tlscfg)
	}
}
func TestTlsCfgAsMapInterface(t *testing.T) {
	var tlscfg TlsCfg
	cfgJSONStr := `	{
	"tls": {
		"server_certificate" : "",			
		"server_key":"",					
		"client_certificate" : "",			
		"client_key":"",					
		"ca_certificate":"",				
		"server_policy":4,					
		"server_name":"",
	},
}`
	eMap := map[string]any{
		"server_certificate": "",
		"server_key":         "",
		"client_certificate": "",
		"client_key":         "",
		"ca_certificate":     "",
		"server_policy":      4,
		"server_name":        "",
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if rcv := tlscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `	{
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
	eMap = map[string]any{
		"server_certificate": "path/To/Server/Cert",
		"server_key":         "path/To/Server/Key",
		"client_certificate": "path/To/Client/Cert",
		"client_key":         "path/To/Client/Key",
		"ca_certificate":     "path/To/CA/Cert",
		"server_name":        "TestServerName",
		"server_policy":      3,
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if rcv := tlscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
