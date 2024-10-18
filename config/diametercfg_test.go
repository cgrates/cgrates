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
	"github.com/cgrates/rpcclient"
)

func TestDiameterAgentCfgloadFromJsonCfg(t *testing.T) {
	jsonCFG := &DiameterAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen_net:           utils.StringPointer("tcp"),
		Listen:               utils.StringPointer("127.0.0.1:3868"),
		Dictionaries_path:    utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		Sessions_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Origin_host:          utils.StringPointer("CGR-DA"),
		Origin_realm:         utils.StringPointer("cgrates.org"),
		Vendor_id:            utils.IntPointer(0),
		Product_name:         utils.StringPointer("randomName"),
		Synced_conn_requests: utils.BoolPointer(true),
		Asr_template:         utils.StringPointer("randomTemplate"),
		Rar_template:         utils.StringPointer("randomTemplate"),
		Forced_disconnect:    utils.StringPointer("forced"),
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID:       utils.StringPointer(utils.CGRateSLwr),
				Timezone: utils.StringPointer("Local"),
			},
		},
	}
	expected := &DiameterAgentCfg{
		Enabled:          true,
		ListenNet:        "tcp",
		Listen:           "127.0.0.1:3868",
		DictionariesPath: "/usr/share/cgrates/diameter/dict/",
		SessionSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		OriginHost:       "CGR-DA",
		OriginRealm:      "cgrates.org",
		VendorID:         0,
		ProductName:      "randomName",
		SyncedConnReqs:   true,
		ASRTemplate:      "randomTemplate",
		RARTemplate:      "randomTemplate",
		ForcedDisconnect: "forced",
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.diameterAgentCfg.loadFromJSONCfg(jsonCFG, jsnCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.diameterAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.diameterAgentCfg))
	}
}

func TestRequestProcessorloadFromJsonCfg1(t *testing.T) {
	cfgJSON := &DiameterAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.diameterAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRequestProcessorloadFromJsonCfg2(t *testing.T) {
	cfgJSONStr := `{ 
      "diameter_agent": {
        "request_processors": [
	        {
		       "id": "random",
            },
         ]
       }
}`
	cfgJSON := &DiameterAgentJsonCfg{
		Request_processors: &[]*ReqProcessorJsnCfg{
			{
				ID: utils.StringPointer("random"),
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err = jsonCfg.diameterAgentCfg.loadFromJSONCfg(cfgJSON, jsonCfg.generalCfg.RSRSep); err != nil {
		t.Error(err)
	}
}

func TestDiameterAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"diameter_agent": {
		"enabled": false,											
		"listen": "127.0.0.1:3868",									
		"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	
		"sessions_conns": ["*birpc_internal","*internal", "*conn1"],
		"origin_host": "CGR-DA",									
		"origin_realm": "cgrates.org",								
		"vendor_id": 0,												
		"product_name": "CGRateS",									
		"synced_conn_requests": true,
		"request_processors": [
                        {
                         "id": "cgrates", 
                         "tenant": "1",
                         "filters": [],
                          "flags": ["1"],
                         "request_fields": [
                            {"path": "randomPath"},
                           ],
                         "reply_fields": [
                              {"path": "randomPath"},
                          ],
                        }
        ]
	},
}`
	eMap := map[string]any{
		utils.ASRTemplateCfg:      "",
		utils.DictionariesPathCfg: "/usr/share/cgrates/diameter/dict/",
		utils.EnabledCfg:          false,
		utils.ForcedDisconnectCfg: "*none",
		utils.ListenCfg:           "127.0.0.1:3868",
		utils.ListenNetCfg:        "tcp",
		utils.OriginHostCfg:       "CGR-DA",
		utils.OriginRealmCfg:      "cgrates.org",
		utils.ProductNameCfg:      "CGRateS",
		utils.RARTemplateCfg:      "",
		utils.SessionSConnsCfg:    []string{rpcclient.BiRPCInternal, utils.MetaInternal, "*conn1"},
		utils.SyncedConnReqsCfg:   true,
		utils.VendorIDCfg:         0,
		utils.RequestProcessorsCfg: []map[string]any{
			{
				utils.IDCfg:       utils.CGRateSLwr,
				utils.TenantCfg:   "1",
				utils.FiltersCfg:  []string{},
				utils.FlagsCfg:    []string{"1"},
				utils.TimezoneCfg: utils.EmptyString,
				utils.RequestFieldsCfg: []map[string]any{
					{
						utils.PathCfg: "randomPath",
						utils.TagCfg:  "randomPath",
					},
				},
				utils.ReplyFieldsCfg: []map[string]any{
					{
						utils.PathCfg: "randomPath",
						utils.TagCfg:  "randomPath",
					},
				},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		for _, v := range cgrCfg.diameterAgentCfg.RequestProcessors[0].ReplyFields {
			v.ComputePath()
		}
		for _, v := range cgrCfg.diameterAgentCfg.RequestProcessors[0].RequestFields {
			v.ComputePath()
		}
		rcv := cgrCfg.diameterAgentCfg.AsMapInterface(utils.InfieldSep)
		if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}

func TestDiameterAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"diameter_agent": {
		"enabled": true,
		"dictionaries_path": "/usr/share/cgrates/diameter",			
		"synced_conn_requests": false,
	},
}`
	eMap := map[string]any{
		utils.ASRTemplateCfg:       "",
		utils.DictionariesPathCfg:  "/usr/share/cgrates/diameter",
		utils.EnabledCfg:           true,
		utils.ForcedDisconnectCfg:  "*none",
		utils.ListenCfg:            "127.0.0.1:3868",
		utils.ListenNetCfg:         "tcp",
		utils.OriginHostCfg:        "CGR-DA",
		utils.OriginRealmCfg:       "cgrates.org",
		utils.ProductNameCfg:       "CGRateS",
		utils.RARTemplateCfg:       "",
		utils.SessionSConnsCfg:     []string{rpcclient.BiRPCInternal},
		utils.SyncedConnReqsCfg:    false,
		utils.VendorIDCfg:          0,
		utils.RequestProcessorsCfg: []map[string]any{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.diameterAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestDiameterAgentCfgClone(t *testing.T) {
	ban := &DiameterAgentCfg{
		Enabled:          true,
		ListenNet:        "tcp",
		Listen:           "127.0.0.1:3868",
		DictionariesPath: "/usr/share/cgrates/diameter/dict/",
		SessionSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"},
		OriginHost:       "CGR-DA",
		OriginRealm:      "cgrates.org",
		VendorID:         0,
		ProductName:      "randomName",
		SyncedConnReqs:   true,
		ASRTemplate:      "randomTemplate",
		RARTemplate:      "randomTemplate",
		ForcedDisconnect: "forced",
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.SessionSConns[1] = ""; ban.SessionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RequestProcessors[0].ID = ""; ban.RequestProcessors[0].ID != "cgrates" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
