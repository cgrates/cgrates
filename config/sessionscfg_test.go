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
	"sort"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestFsAgentCfgloadFromJsonCfg1(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Create_cdr:     utils.BoolPointer(true),
		Subscribe_park: utils.BoolPointer(true),
		Event_socket_conns: &[]*FsConnJsonCfg{
			{
				Address:    utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
			{
				Address:    utils.StringPointer("2.3.4.5:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
			},
		},
	}
	eFsAgentConfig := &FsAgentCfg{
		Enabled:       true,
		CreateCdr:     true,
		SubscribePark: true,
		EventSocketConns: []*FsConnCfg{
			{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5, Alias: "1.2.3.4:8021"},
			{Address: "2.3.4.5:8021", Password: "ClueCon", Reconnects: 5, Alias: "2.3.4.5:8021"},
		},
	}
	fsAgentCfg := new(FsAgentCfg)
	if err := fsAgentCfg.loadFromJsonCfg(fsAgentJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFsAgentConfig, fsAgentCfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(eFsAgentConfig), utils.ToJSON(fsAgentCfg))
	}
}

func TestSessionSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2018"),
		Chargers_conns:        &[]string{utils.MetaInternal, "*conn1"},
		Rals_conns:            &[]string{utils.MetaInternal, "*conn1"},
		Resources_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Routes_conns:          &[]string{utils.MetaInternal, "*conn1"},
		Attributes_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Cdrs_conns:            &[]string{utils.MetaInternal, "*conn1"},
		Replication_conns:     &[]string{"*conn1"},
		Debit_interval:        utils.StringPointer("2"),
		Store_session_costs:   utils.BoolPointer(true),
		Session_ttl:           utils.StringPointer("0"),
		Session_indexes:       &[]string{},
		Client_protocol:       utils.Float64Pointer(2.5),
		Channel_sync_interval: utils.StringPointer("10"),
		Terminate_attempts:    utils.IntPointer(6),
		Alterable_fields:      &[]string{},
		Min_dur_low_balance:   utils.StringPointer("1"),
		Scheduler_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Stir: &STIRJsonCfg{
			Allowed_attest:      &[]string{utils.META_ANY},
			Payload_maxduration: utils.StringPointer("-1"),
			Default_attest:      utils.StringPointer("A"),
			Publickey_path:      utils.StringPointer("randomPath"),
			Privatekey_path:     utils.StringPointer("randomPath"),
		},
	}
	expected := &SessionSCfg{
		Enabled:             true,
		ListenBijson:        "127.0.0.1:2018",
		ChargerSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		RALsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder), "*conn1"},
		ResSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		ThreshSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS), "*conn1"},
		RouteSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes), "*conn1"},
		AttrSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		CDRsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), "*conn1"},
		ReplicationConns:    []string{"*conn1"},
		DebitInterval:       2,
		StoreSCosts:         true,
		SessionTTL:          0,
		SessionIndexes:      utils.StringMap{},
		ClientProtocol:      2.5,
		ChannelSyncInterval: 10,
		TerminateAttempts:   6,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    1,
		SchedulerConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler), "*conn1"},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.META_ANY: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "randomPath",
			PublicKeyPath:      "randomPath",
		},
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.sessionSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sessionSCfg))
	}
}

func TestSessionSCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Replication_conns: &[]string{utils.MetaInternal},
	}
	expected := "Replication connection ID needs to be different than *internal"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase3(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Debit_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase5(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Session_ttl: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase7(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Channel_sync_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase8(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Min_dur_low_balance: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase9(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Session_ttl_last_usage: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cfgJSON1 := &SessionSJsonCfg{
		Session_ttl_last_used: utils.StringPointer("1ss"),
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON1); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cfgJSON2 := &SessionSJsonCfg{
		Session_ttl_max_delay: utils.StringPointer("1ss"),
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON2); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cfgJSON3 := &SessionSJsonCfg{
		Session_ttl_usage: utils.StringPointer("1ss"),
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON3); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgloadFromJsonCfgCase10(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Session_ttl_last_usage: utils.StringPointer("1"),
		Session_ttl_last_used:  utils.StringPointer("10"),
		Session_ttl_max_delay:  utils.StringPointer("100"),
		Session_ttl_usage:      utils.StringPointer("1"),
	}
	expected := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		RALsConns:           []string{},
		ResSConns:           []string{},
		ThreshSConns:        []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttrSConns:          []string{},
		CDRsConns:           []string{},
		ReplicationConns:    []string{},
		DebitInterval:       0,
		StoreSCosts:         false,
		SessionTTL:          0,
		SessionIndexes:      utils.StringMap{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    0,
		SchedulerConns:      []string{},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.META_ANY: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "",
			PublicKeyPath:      "",
		},
		SessionTTLMaxDelay:  utils.DurationPointer(100),
		SessionTTLLastUsage: utils.DurationPointer(1),
		SessionTTLLastUsed:  utils.DurationPointer(10),
		SessionTTLUsage:     utils.DurationPointer(1),
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.sessionSCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sessionSCfg))
	}
}

func TestSessionSCfgloadFromJsonCfgCase11(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Stir: &STIRJsonCfg{
			Payload_maxduration: utils.StringPointer("1ss"),
		},
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sessionSCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSessionSCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {
          "channel_sync_interval": "1s",
          "session_ttl_max_delay": "3h0m0s",
          "session_ttl_last_used": "0s",
          "session_ttl_usage": "1s",
          "session_ttl_last_usage": "10s",
           "sessions": {
			"stir": {
				"payload_maxduration": "-1",
			},
		},
    },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.ListenBijsonCfg:        "127.0.0.1:2014",
		utils.ChargerSConnsCfg:       []string{},
		utils.RALsConnsCfg:           []string{},
		utils.CDRsConnsCfg:           []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.ThresholdSConnsCfg:     []string{},
		utils.StatSConnsCfg:          []string{},
		utils.RouteSConnsCfg:         []string{},
		utils.AttributeSConnsCfg:     []string{},
		utils.ReplicationConnsCfg:    []string{},
		utils.DebitIntervalCfg:       "0",
		utils.StoreSCostsCfg:         false,
		utils.SessionTTLCfg:          "0",
		utils.SessionTTLMaxDelayCfg:  "3h0m0s",
		utils.SessionTTLLastUsedCfg:  "0s",
		utils.SessionTTLUsageCfg:     "1s",
		utils.SessionTTLLastUsageCfg: "10s",
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      1.0,
		utils.ChannelSyncIntervalCfg: "1s",
		utils.TerminateAttemptsCfg:   5,
		utils.MinDurLowBalanceCfg:    "0",
		utils.AlterableFieldsCfg:     []string{},
		utils.STIRCfg: map[string]interface{}{
			utils.AllowedAtestCfg:       []string{"*any"},
			utils.PayloadMaxdurationCfg: "-1",
			utils.DefaultAttestCfg:      "A",
			utils.PublicKeyPathCfg:      "",
			utils.PrivateKeyPathCfg:     "",
		},
		utils.SchedulerConnsCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sessionSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSessionSCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
		"sessions": {
			"enabled": true,
			"listen_bijson": "127.0.0.1:2018",
			"chargers_conns": ["*internal:*chargers", "*conn1"],
			"rals_conns": ["*internal:*responder", "*conn1"],
			"cdrs_conns": ["*internal:*cdrs", "*conn1"],
			"resources_conns": ["*internal:*resources", "*conn1"],
			"thresholds_conns": ["*internal:*thresholds", "*conn1"],
			"stats_conns": ["*internal:*stats", "*conn1"],
			"routes_conns": ["*internal:*routes", "*conn1"],
			"attributes_conns": ["*internal:*attributes", "*conn1"],
			"replication_conns": ["*localhost"],
			"debit_interval": "8s",
			"store_session_costs": true,
			"session_ttl": "1s",
            "min_dur_low_balance": "1s",
			"client_protocol": 2.0,
			"terminate_attempts": 10,
			"stir": {
				"allowed_attest": ["any1","any2"],
				"payload_maxduration": "1s",
				"default_attest": "B",
				"publickey_path": "",
				"privatekey_path": "",
			},
			"scheduler_conns": ["*internal:*scheduler", "*conn1"],
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.ListenBijsonCfg:        "127.0.0.1:2018",
		utils.ChargerSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.RALsConnsCfg:           []string{utils.MetaInternal, "*conn1"},
		utils.CDRsConnsCfg:           []string{utils.MetaInternal, "*conn1"},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal, "*conn1"},
		utils.ThresholdSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:          []string{utils.MetaInternal, "*conn1"},
		utils.RouteSConnsCfg:         []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.ReplicationConnsCfg:    []string{utils.MetaLocalHost},
		utils.DebitIntervalCfg:       "8s",
		utils.StoreSCostsCfg:         true,
		utils.MinDurLowBalanceCfg:    "1s",
		utils.SessionTTLCfg:          "1s",
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      2.0,
		utils.ChannelSyncIntervalCfg: "0",
		utils.TerminateAttemptsCfg:   10,
		utils.AlterableFieldsCfg:     []string{},
		utils.STIRCfg: map[string]interface{}{
			utils.AllowedAtestCfg:       []string{"any1", "any2"},
			utils.PayloadMaxdurationCfg: "1s",
			utils.DefaultAttestCfg:      "B",
			utils.PublicKeyPathCfg:      "",
			utils.PrivateKeyPathCfg:     "",
		},
		utils.SchedulerConnsCfg: []string{utils.MetaInternal, "*conn1"},
	}
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	rcv := cgrCfg.sessionSCfg.AsMapInterface()
	sort.Strings(rcv[utils.STIRCfg].(map[string]interface{})[utils.AllowedAtestCfg].([]string))
	if !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSessionSCfgAsMapInterfaceCase3(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {
			"stir": {
				"payload_maxduration": "0",
			},
		},
    },
}`
	eMap := map[string]interface{}{
		utils.STIRCfg: map[string]interface{}{
			utils.AllowedAtestCfg:       []string{"*any"},
			utils.PayloadMaxdurationCfg: "0",
			utils.DefaultAttestCfg:      "A",
			utils.PublicKeyPathCfg:      "",
			utils.PrivateKeyPathCfg:     "",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sessionSCfg.AsMapInterface(); !reflect.DeepEqual(eMap[utils.STIRCfg], rcv[utils.STIRCfg]) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap[utils.STIRCfg]), utils.ToJSON(rcv[utils.STIRCfg]))
	}
}

func TestFsAgentCfgloadFromJsonCfgCase1(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Sessions_conns:         &[]string{utils.MetaInternal},
		Create_cdr:             utils.BoolPointer(true),
		Subscribe_park:         utils.BoolPointer(true),
		Low_balance_ann_file:   utils.StringPointer("randomFile"),
		Empty_balance_ann_file: utils.StringPointer("randomEmptyFile"),
		Empty_balance_context:  utils.StringPointer("randomEmptyContext"),
		Max_wait_connection:    utils.StringPointer("2"),
		Extra_fields:           &[]string{},
		Event_socket_conns: &[]*FsConnJsonCfg{
			{
				Address:    utils.StringPointer("1.2.3.4:8021"),
				Password:   utils.StringPointer("ClueCon"),
				Reconnects: utils.IntPointer(5),
				Alias:      utils.StringPointer("127.0.0.1:8021"),
			},
		},
	}
	expected := &FsAgentCfg{
		Enabled:             true,
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		SubscribePark:       true,
		CreateCdr:           true,
		LowBalanceAnnFile:   "randomFile",
		EmptyBalanceAnnFile: "randomEmptyFile",
		EmptyBalanceContext: "randomEmptyContext",
		MaxWaitConnection:   2,
		ExtraFields:         RSRParsers{},
		EventSocketConns: []*FsConnCfg{
			{
				Address:    "1.2.3.4:8021",
				Password:   "ClueCon",
				Reconnects: 5,
				Alias:      "127.0.0.1:8021",
			},
		},
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.fsAgentCfg.loadFromJsonCfg(fsAgentJsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.fsAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.fsAgentCfg))
	}
}

func TestFsAgentCfgloadFromJsonCfgCase2(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Max_wait_connection: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.fsAgentCfg.loadFromJsonCfg(fsAgentJsnCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFsAgentCfgloadFromJsonCfgCase3(t *testing.T) {
	fsAgentJsnCfg := &FreeswitchAgentJsonCfg{
		Extra_fields: &[]string{"a{*"},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.fsAgentCfg.loadFromJsonCfg(fsAgentJsnCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFsAgentCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.SessionSConnsCfg:       []string{"*internal"},
		utils.SubscribeParkCfg:       true,
		utils.CreateCdrCfg:           false,
		utils.ExtraFieldsCfg:         "",
		utils.LowBalanceAnnFileCfg:   "",
		utils.EmptyBalanceContextCfg: "",
		utils.EmptyBalanceAnnFileCfg: "",
		utils.MaxWaitConnectionCfg:   "2s",
		utils.EventSocketConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8021", utils.Password: "ClueCon", utils.ReconnectsCfg: 5, utils.AliasCfg: "127.0.0.1:8021"},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {
          "enabled": true,						
          "sessions_conns": ["*conn1","*conn2"],
	      "subscribe_park": false,					
	      "create_cdr": true,
	      "max_wait_connection": "7s",			
	      "event_socket_conns":[					
		      {"address": "127.0.0.1:8000", "password": "ClueCon123", "reconnects": 8, "alias": "127.0.0.1:8000"}
	],},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.SessionSConnsCfg:       []string{"*conn1", "*conn2"},
		utils.SubscribeParkCfg:       false,
		utils.CreateCdrCfg:           true,
		utils.ExtraFieldsCfg:         "",
		utils.LowBalanceAnnFileCfg:   "",
		utils.EmptyBalanceContextCfg: "",
		utils.EmptyBalanceAnnFileCfg: "",
		utils.MaxWaitConnectionCfg:   "7s",
		utils.EventSocketConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8000", utils.Password: "ClueCon123", utils.ReconnectsCfg: 8, utils.AliasCfg: "127.0.0.1:8000"},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgAsMapInterfaceCase3(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {
          "extra_fields": ["randomFields"],		
          "max_wait_connection": "0",
    }
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.SessionSConnsCfg:       []string{"*internal"},
		utils.SubscribeParkCfg:       true,
		utils.CreateCdrCfg:           false,
		utils.ExtraFieldsCfg:         "randomFields",
		utils.LowBalanceAnnFileCfg:   "",
		utils.EmptyBalanceContextCfg: "",
		utils.EmptyBalanceAnnFileCfg: "",
		utils.MaxWaitConnectionCfg:   "",
		utils.EventSocketConnsCfg: []map[string]interface{}{
			{utils.AddressCfg: "127.0.0.1:8021", utils.Password: "ClueCon", utils.ReconnectsCfg: 5, utils.AliasCfg: "127.0.0.1:8021"},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

}

func TestFsConnCfgloadFromJsonCfg(t *testing.T) {
	var fscocfg, expected FsConnCfg
	if err := fscocfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	if err := fscocfg.loadFromJsonCfg(new(FsConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fscocfg)
	}
	json := &FsConnJsonCfg{
		Address:    utils.StringPointer("127.0.0.1:8448"),
		Password:   utils.StringPointer("pass123"),
		Reconnects: utils.IntPointer(5),
		Alias:      utils.StringPointer("127.0.0.1:8448"),
	}
	expected = FsConnCfg{
		Address:    "127.0.0.1:8448",
		Password:   "pass123",
		Reconnects: 5,
		Alias:      "127.0.0.1:8448",
	}
	if err = fscocfg.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fscocfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(fscocfg))
	}
}

func TestRemoteHostloadFromJsonCfg(t *testing.T) {
	var hpoolcfg, expected RemoteHost
	if err := hpoolcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, hpoolcfg)
	}
	if err := hpoolcfg.loadFromJsonCfg(new(RemoteHostJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, hpoolcfg)
	}
	json := &RemoteHostJson{
		Address:     utils.StringPointer("127.0.0.1:8448"),
		Synchronous: utils.BoolPointer(true),
	}
	expected = RemoteHost{
		Address:     "127.0.0.1:8448",
		Synchronous: true,
	}
	if err = hpoolcfg.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, hpoolcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(hpoolcfg))
	}
}

func TestAsteriskAgentCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{utils.MetaInternal},
		Create_cdr:     utils.BoolPointer(true),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
				Alias:            utils.StringPointer("127.0.0.1:8448"),
				Address:          utils.StringPointer("127.0.0.1:8088"),
				User:             utils.StringPointer(utils.CGRATES),
				Password:         utils.StringPointer("CGRateS.org"),
				Connect_attempts: utils.IntPointer(3),
				Reconnects:       utils.IntPointer(5),
			},
		},
	}
	expected := &AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		CreateCDR:     true,
		AsteriskConns: []*AsteriskConnCfg{{
			Alias:           "127.0.0.1:8448",
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.asteriskAgentCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.asteriskAgentCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.asteriskAgentCfg))
	}
}

func TestAsteriskAgentCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       false,
		utils.SessionSConnsCfg: []string{"*internal"},
		utils.CreateCdrCfg:     false,
		utils.AsteriskConnsCfg: []map[string]interface{}{
			{utils.AliasCfg: "", utils.AddressCfg: "127.0.0.1:8088", utils.UserCf: "cgrates", utils.Password: "CGRateS.org", utils.ConnectAttemptsCfg: 3, utils.ReconnectsCfg: 5},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAsteriskAgentCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {
		"enabled": true,
		"sessions_conns": ["*conn1","*conn2"],
		"create_cdr": true,
		"asterisk_conns":[
			{"address": "127.0.0.1:8089","connect_attempts": 5,"reconnects": 8}
		],
	},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:       true,
		utils.SessionSConnsCfg: []string{"*conn1", "*conn2"},
		utils.CreateCdrCfg:     true,
		utils.AsteriskConnsCfg: []map[string]interface{}{
			{utils.AliasCfg: "", utils.AddressCfg: "127.0.0.1:8089", utils.UserCf: "cgrates", utils.Password: "CGRateS.org", utils.ConnectAttemptsCfg: 5, utils.ReconnectsCfg: 8},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAsteriskConnCfgloadFromJsonCfg(t *testing.T) {
	var asconcfg, expected AsteriskConnCfg
	if err := asconcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asconcfg)
	}
	if err := asconcfg.loadFromJsonCfg(new(AstConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asconcfg)
	}
	json := &AstConnJsonCfg{
		Address:          utils.StringPointer("127.0.0.1:8088"),
		User:             utils.StringPointer("cgrates"),
		Password:         utils.StringPointer("CGRateS.org"),
		Connect_attempts: utils.IntPointer(3),
		Reconnects:       utils.IntPointer(5),
	}
	expected = AsteriskConnCfg{
		Address:         "127.0.0.1:8088",
		User:            "cgrates",
		Password:        "CGRateS.org",
		ConnectAttempts: 3,
		Reconnects:      5,
	}
	if err = asconcfg.loadFromJsonCfg(json); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, asconcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(asconcfg))
	}
}
