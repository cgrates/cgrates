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

func TestDiameterAgentCfgloadFromJsonCfg(t *testing.T) {
	var dacfg, expected DiameterAgentCfg
	if err := dacfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dacfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dacfg)
	}
	if err := dacfg.loadFromJsonCfg(new(DiameterAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dacfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, dacfg)
	}
	cfgJSONStr := `{
"diameter_agent": {
	"enabled": false,											// enables the diameter agent: <true|false>
	"listen": "127.0.0.1:3868",									// address where to listen for diameter requests <x.y.z.y:1234>
	"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	// path towards directory holding additional dictionaries to load
	"sessions_conns": ["*internal"],
	"origin_host": "CGR-DA",									// diameter Origin-Host AVP used in replies
	"origin_realm": "cgrates.org",								// diameter Origin-Realm AVP used in replies
	"vendor_id": 0,												// diameter Vendor-Id AVP used in replies
	"product_name": "CGRateS",									// diameter Product-Name AVP used in replies
	"synced_conn_requests": true,
	"templates":{},
	"request_processors": [],
},
}`
	expected = DiameterAgentCfg{
		Listen:           "127.0.0.1:3868",
		DictionariesPath: "/usr/share/cgrates/diameter/dict/",
		SessionSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		OriginHost:       "CGR-DA",
		OriginRealm:      "cgrates.org",
		VendorId:         0,
		ProductName:      "CGRateS",
		SyncedConnReqs:   true,
		Templates:        make(map[string][]*FCTemplate),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dacfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dacfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(dacfg))
	}
}

func TestDiameterAgentCfgAsMapInterface(t *testing.T) {
	var dacfg DiameterAgentCfg
	cfgJSONStr := `{
	"diameter_agent": {
		"enabled": false,											
		"listen": "127.0.0.1:3868",									
		"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	
		"sessions_conns": ["*internal"],
		"origin_host": "CGR-DA",									
		"origin_realm": "cgrates.org",								
		"vendor_id": 0,												
		"product_name": "CGRateS",									
		"synced_conn_requests": true,
		"templates":{},
		"request_processors": [],
	},
}`
	eMap := map[string]any{
		"asr_template":         "",
		"concurrent_requests":  0,
		"dictionaries_path":    "/usr/share/cgrates/diameter/dict/",
		"enabled":              false,
		"listen":               "127.0.0.1:3868",
		"listen_net":           "",
		"origin_host":          "CGR-DA",
		"origin_realm":         "cgrates.org",
		"product_name":         "CGRateS",
		"sessions_conns":       []string{"*internal"},
		"synced_conn_requests": true,
		"vendor_id":            0,
		"templates":            map[string][]map[string]any{},
		"request_processors":   []map[string]any{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dacfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if rcv := dacfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestDiameterCfgloadFromJsonCfg(t *testing.T) {
	strErr := "test`"
	id := "t"
	str := "test"
	slc := []string{"val1", "val2"}
	fcs := []*FcTemplateJsonCfg{{Value: &str}}

	d := DiameterAgentCfg{
		RequestProcessors: []*RequestProcessor{
			{
				ID: str,
			},
		},
	}

	tests := []struct {
		name string
		js   *DiameterAgentJsonCfg
		sep  string
		exp  string
	}{
		{
			name: "session conns",
			js: &DiameterAgentJsonCfg{
				Sessions_conns: &[]string{"val1", "val2"},
			},
			sep: "",
			exp: "",
		},
		{
			name: "Templates error",
			js: &DiameterAgentJsonCfg{
				Templates: map[string][]*FcTemplateJsonCfg{"test": {{Value: &strErr}}},
			},
			sep: "",
			exp: "Unclosed unspilit syntax",
		},
		{
			name: "Request processors",
			js: &DiameterAgentJsonCfg{
				Request_processors: &[]*ReqProcessorJsnCfg{{
					ID:             &id,
					Filters:        &slc,
					Tenant:         &str,
					Timezone:       &str,
					Flags:          &slc,
					Request_fields: &fcs,
					Reply_fields:   &fcs,
				}},
			},
			sep: "",
			exp: "",
		},
		{
			name: "Request processors load data into one set",
			js: &DiameterAgentJsonCfg{
				Request_processors: &[]*ReqProcessorJsnCfg{{
					ID:             &str,
					Filters:        &slc,
					Tenant:         &str,
					Timezone:       &str,
					Flags:          &slc,
					Request_fields: &fcs,
					Reply_fields:   &fcs,
				}},
			},
			sep: "",
			exp: "",
		},
		{
			name: "Request processors error",
			js: &DiameterAgentJsonCfg{
				Request_processors: &[]*ReqProcessorJsnCfg{{
					ID:             &str,
					Filters:        &slc,
					Tenant:         &strErr,
					Timezone:       &str,
					Flags:          &slc,
					Request_fields: &fcs,
					Reply_fields:   &fcs,
				}},
			},
			sep: "",
			exp: "Unclosed unspilit syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := d.loadFromJsonCfg(tt.js, tt.sep)

			if err != nil {
				if err.Error() != tt.exp {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestDiameterAgentCfgAsMapInterface2(t *testing.T) {
	str := "test"
	slc := []string{"val1", "val2"}
	bl := false

	ds := DiameterAgentCfg{
		Enabled:          bl,
		ListenNet:        str,
		Listen:           str,
		DictionariesPath: str,
		SessionSConns:    slc,
		OriginHost:       str,
		OriginRealm:      str,
		VendorId:         1,
		ProductName:      str,
		ConcurrentReqs:   1,
		SyncedConnReqs:   bl,
		ASRTemplate:      str,
		Templates: map[string][]*FCTemplate{
			"test": {{Value: RSRParsers{{Rules: "test"}}}},
		},
		RequestProcessors: []*RequestProcessor{{}},
	}

	exp := map[string]any{
		utils.EnabledCfg:          ds.Enabled,
		utils.ListenNetCfg:        ds.ListenNet,
		utils.ListenCfg:           ds.Listen,
		utils.DictionariesPathCfg: ds.DictionariesPath,
		utils.SessionSConnsCfg:    slc,
		utils.OriginHostCfg:       ds.OriginHost,
		utils.OriginRealmCfg:      ds.OriginRealm,
		utils.VendorIdCfg:         ds.VendorId,
		utils.ProductNameCfg:      ds.ProductName,
		utils.ConcurrentReqsCfg:   ds.ConcurrentReqs,
		utils.SyncedConnReqsCfg:   ds.SyncedConnReqs,
		utils.ASRTemplateCfg:      ds.ASRTemplate,
		utils.TemplatesCfg: map[string][]map[string]any{
			"test": {{"value": "test"}},
		},
		utils.RequestProcessorsCfg: []map[string]any{
			{
				utils.IDCfg:            "",
				utils.TenantCfg:        "",
				utils.FiltersCfg:       []string{},
				utils.FlagsCfg:         map[string][]string{},
				utils.TimezoneCfgC:     "",
				utils.RequestFieldsCfg: []map[string]any{},
				utils.ReplyFieldsCfg:   []map[string]any{},
			},
		},
	}

	rcv := ds.AsMapInterface("")

	if rcv == nil {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}
