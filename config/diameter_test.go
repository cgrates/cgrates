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
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDiameterAgentCfgloadFromJsonCfg(t *testing.T) {
	jsonCFG := &DiameterAgentJsonCfg{
		Enabled:          utils.BoolPointer(true),
		ListenNet:        utils.StringPointer("tcp"),
		Listen:           utils.StringPointer("127.0.0.1:3868"),
		DictionariesPath: utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		CEApplications:   &[]string{"Base"},
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{utils.MetaInternal, "*conn1"}},
			},
			utils.MetaStats: {
				{ConnIDs: []string{utils.MetaInternal, "*conn1"}},
			},
			utils.MetaThresholds: {
				{ConnIDs: []string{utils.MetaInternal, "*conn1"}},
			},
		},
		OriginHost:         utils.StringPointer("CGR-DA"),
		OriginRealm:        utils.StringPointer("cgrates.org"),
		VendorID:           utils.IntPointer(0),
		ProductName:        utils.StringPointer("randomName"),
		SyncedConnRequests: utils.BoolPointer(true),
		ASRTemplate:        utils.StringPointer("randomTemplate"),
		RARTemplate:        utils.StringPointer("randomTemplate"),
		ForcedDisconnect:   utils.StringPointer("forced"),
		RequestProcessors: &[]*ReqProcessorJsnCfg{
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
		CEApplications:   []string{"Base"},
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"}},
			},
			utils.MetaStats: {
				{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}},
			},
			utils.MetaThresholds: {
				{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}},
			},
		},

		ConnStatusStatQueueIDs: []string{},
		ConnStatusThresholdIDs: []string{},
		OriginHost:             "CGR-DA",
		OriginRealm:            "cgrates.org",
		VendorID:               0,
		ProductName:            "randomName",
		SyncedConnReqs:         true,
		ASRTemplate:            "randomTemplate",
		RARTemplate:            "randomTemplate",
		ForcedDisconnect:       "forced",
		RequestProcessors: []*RequestProcessor{
			{
				ID:       "cgrates",
				Timezone: "Local",
			},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.diameterAgentCfg.loadFromJSONCfg(jsonCFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.diameterAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.diameterAgentCfg))
	}

	jsonCFG = nil
	if err := jsnCfg.diameterAgentCfg.loadFromJSONCfg(jsonCFG); err != nil {
		t.Error(err)
	}
}

func TestRequestProcessorloadFromJsonCfg1(t *testing.T) {
	cfgJSON := &DiameterAgentJsonCfg{
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{
				Tenant: utils.StringPointer("a{*"),
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.diameterAgentCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
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
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{
				ID: utils.StringPointer("random"),
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if err := jsonCfg.diameterAgentCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}
}

func TestDiameterAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"diameter_agent": {
		"enabled": false,
		"listen": "127.0.0.1:3868",
		"dictionaries_path": "/usr/share/cgrates/diameter/dict/",
		"ce_applications": ["Base"],
		"conns": {
			"*sessions": [{"ConnIDs":["*birpc_internal","*internal","*conn1"]}],
			"*stats": [{"ConnIDs":["*birpc_internal","*internal","*conn1"]}],
			"*thresholds": [{"ConnIDs":["*birpc_internal","*internal","*conn1"]}]
		},
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
		utils.CEApplicationsCfg:   []string{"Base"},
		utils.EnabledCfg:          false,
		utils.ForcedDisconnectCfg: "*none",
		utils.ListenCfg:           "127.0.0.1:3868",
		utils.ListenNetCfg:        "tcp",
		utils.OriginHostCfg:       "CGR-DA",
		utils.OriginRealmCfg:      "cgrates.org",
		utils.ProductNameCfg:      "CGRateS",
		utils.RARTemplateCfg:      "",
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaSessionS:   {{ConnIDs: []string{rpcclient.BiRPCInternal, utils.MetaInternal, "*conn1"}}},
			utils.MetaStats:      {{ConnIDs: []string{rpcclient.BiRPCInternal, utils.MetaInternal, "*conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{rpcclient.BiRPCInternal, utils.MetaInternal, "*conn1"}}},
		},
		utils.ConnStatusStatQueueIDsCfg:  []string{},
		utils.ConnStatusThresholdIDsCfg:  []string{},
		utils.SyncedConnReqsCfg:          true,
		utils.VendorIDCfg:                0,
		utils.ConnHealthCheckIntervalCfg: "0s",
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
		rcv := cgrCfg.diameterAgentCfg.AsMapInterface()
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
		utils.ASRTemplateCfg:      "",
		utils.DictionariesPathCfg: "/usr/share/cgrates/diameter",
		utils.EnabledCfg:          true,
		utils.ForcedDisconnectCfg: "*none",
		utils.ListenCfg:           "127.0.0.1:3868",
		utils.ListenNetCfg:        "tcp",
		utils.OriginHostCfg:       "CGR-DA",
		utils.OriginRealmCfg:      "cgrates.org",
		utils.ProductNameCfg:      "CGRateS",
		utils.RARTemplateCfg:      "",
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaSessionS: {{ConnIDs: []string{rpcclient.BiRPCInternal}}},
		},
		utils.ConnStatusStatQueueIDsCfg:  []string{},
		utils.ConnStatusThresholdIDsCfg:  []string{},
		utils.SyncedConnReqsCfg:          false,
		utils.VendorIDCfg:                0,
		utils.ConnHealthCheckIntervalCfg: "0s",
		utils.RequestProcessorsCfg:       []map[string]any{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.diameterAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestDiameterAgentCfgClone(t *testing.T) {
	ban := &DiameterAgentCfg{
		Enabled:          true,
		ListenNet:        "tcp",
		Listen:           "127.0.0.1:3868",
		DictionariesPath: "/usr/share/cgrates/diameter/dict/",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), "*conn1"}},
			},
		},
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
	if rcv.Conns[utils.MetaSessionS][0].ConnIDs[1] = ""; ban.Conns[utils.MetaSessionS][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RequestProcessors[0].ID = ""; ban.RequestProcessors[0].ID != "cgrates" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffDiameterAgentJsonCfg(t *testing.T) {
	var d *DiameterAgentJsonCfg

	v1 := &DiameterAgentCfg{
		Enabled:          false,
		ListenNet:        "tcp",
		Listen:           "localhost:8080",
		DictionariesPath: "/path/",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{"*localhost"}},
			},
			utils.MetaStats: {
				{ConnIDs: []string{"*localhost"}},
			},
			utils.MetaThresholds: {
				{ConnIDs: []string{"*localhost"}},
			},
		},
		OriginHost:              "originHost",
		OriginRealm:             "originRealm",
		ConnStatusStatQueueIDs:  []string{"conn1", "conn2"},
		ConnStatusThresholdIDs:  []string{"conn1", "conn2"},
		ConnHealthCheckInterval: time.Second,
		VendorID:                2,
		ProductName:             "productName",
		SyncedConnReqs:          false,
		ASRTemplate:             "ASRTemplate",
		RARTemplate:             "RARTemplate",
		ForcedDisconnect:        "ForcedDisconnect",
		RequestProcessors:       []*RequestProcessor{},
	}

	v2 := &DiameterAgentCfg{
		Enabled:          true,
		ListenNet:        "udp",
		Listen:           "localhost:8037",
		DictionariesPath: "/path/different",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{"*birpc_internal"}},
			},
			utils.MetaStats: {
				{ConnIDs: []string{"*internal"}},
			},
			utils.MetaThresholds: {
				{ConnIDs: []string{"*internal"}},
			},
		},
		CEApplications:          []string{"Base"},
		OriginHost:              "diffOriginHost",
		OriginRealm:             "diffOriginRealm",
		ConnStatusStatQueueIDs:  []string{"conn2", "conn3"},
		ConnStatusThresholdIDs:  []string{"conn2", "conn3"},
		ConnHealthCheckInterval: 2 * time.Second,
		VendorID:                5,
		ProductName:             "diffProductName",
		SyncedConnReqs:          true,
		ASRTemplate:             "diffASRTemplate",
		RARTemplate:             "diffRARTemplate",
		ForcedDisconnect:        "diffForcedDisconnect",
		RequestProcessors: []*RequestProcessor{
			{
				ID: "id",
			},
		},
	}

	expected := &DiameterAgentJsonCfg{
		Enabled:          utils.BoolPointer(true),
		ListenNet:        utils.StringPointer("udp"),
		Listen:           utils.StringPointer("localhost:8037"),
		DictionariesPath: utils.StringPointer("/path/different"),
		CEApplications:   utils.SliceStringPointer([]string{"Base"}),
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{"*birpc_internal"}},
			},
			utils.MetaStats: {
				{ConnIDs: []string{"*internal"}},
			},
			utils.MetaThresholds: {
				{ConnIDs: []string{"*internal"}},
			},
		},
		OriginHost:              utils.StringPointer("diffOriginHost"),
		OriginRealm:             utils.StringPointer("diffOriginRealm"),
		VendorID:                utils.IntPointer(5),
		ProductName:             utils.StringPointer("diffProductName"),
		SyncedConnRequests:      utils.BoolPointer(true),
		ASRTemplate:             utils.StringPointer("diffASRTemplate"),
		RARTemplate:             utils.StringPointer("diffRARTemplate"),
		ForcedDisconnect:        utils.StringPointer("diffForcedDisconnect"),
		ConnStatusStatQueueIDs:  &[]string{"conn2", "conn3"},
		ConnStatusThresholdIDs:  &[]string{"conn2", "conn3"},
		ConnHealthCheckInterval: utils.StringPointer("2s"),
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{
				ID: utils.StringPointer("id"),
			},
		},
	}

	rcv := diffDiameterAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &DiameterAgentJsonCfg{
		RequestProcessors: &[]*ReqProcessorJsnCfg{
			{},
		},
	}
	rcv = diffDiameterAgentJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiameterAgentCloneSection(t *testing.T) {
	dmtCfg := &DiameterAgentCfg{
		Enabled:          false,
		ListenNet:        "tcp",
		Listen:           "localhost:8080",
		DictionariesPath: "/path/",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{"*localhost"}},
			},
		},
		OriginHost:        "originHost",
		OriginRealm:       "originRealm",
		VendorID:          2,
		ProductName:       "productName",
		SyncedConnReqs:    false,
		ASRTemplate:       "ASRTemplate",
		RARTemplate:       "RARTemplate",
		ForcedDisconnect:  "ForcedDisconnect",
		RequestProcessors: []*RequestProcessor{},
	}

	exp := &DiameterAgentCfg{
		Enabled:          false,
		ListenNet:        "tcp",
		Listen:           "localhost:8080",
		DictionariesPath: "/path/",
		Conns: map[string][]*DynamicConns{
			utils.MetaSessionS: {
				{ConnIDs: []string{"*localhost"}},
			},
		},
		OriginHost:        "originHost",
		OriginRealm:       "originRealm",
		VendorID:          2,
		ProductName:       "productName",
		SyncedConnReqs:    false,
		ASRTemplate:       "ASRTemplate",
		RARTemplate:       "RARTemplate",
		ForcedDisconnect:  "ForcedDisconnect",
		RequestProcessors: []*RequestProcessor{},
	}

	rcv := dmtCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
