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
	"time"

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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(eFsAgentConfig), utils.ToJSON(fsAgentCfg))
	}
}

func TestSessionSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2018"),
		Chargers_conns:        &[]string{utils.MetaInternal},
		Rals_conns:            &[]string{utils.MetaInternal},
		Resources_conns:       &[]string{utils.MetaInternal},
		Thresholds_conns:      &[]string{utils.MetaInternal},
		Stats_conns:           &[]string{utils.MetaInternal},
		Routes_conns:          &[]string{utils.MetaInternal},
		Attributes_conns:      &[]string{utils.MetaInternal},
		Cdrs_conns:            &[]string{utils.MetaInternal},
		Replication_conns:     &[]string{"*conn1"},
		Debit_interval:        utils.StringPointer("2"),
		Store_session_costs:   utils.BoolPointer(true),
		Min_call_duration:     utils.StringPointer("1"),
		Max_call_duration:     utils.StringPointer("100"),
		Session_ttl:           utils.StringPointer("0"),
		Session_indexes:       &[]string{},
		Client_protocol:       utils.Float64Pointer(2.5),
		Channel_sync_interval: utils.StringPointer("10"),
		Terminate_attempts:    utils.IntPointer(6),
		Alterable_fields:      &[]string{},
		Min_dur_low_balance:   utils.StringPointer("1"),
		Scheduler_conns:       &[]string{utils.MetaInternal},
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
		ChargerSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)},
		RALsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)},
		ThreshSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)},
		RouteSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)},
		AttrSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)},
		CDRsConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		ReplicationConns:    []string{"*conn1"},
		DebitInterval:       time.Duration(2),
		StoreSCosts:         true,
		MinCallDuration:     time.Duration(1),
		MaxCallDuration:     time.Duration(100),
		SessionTTL:          time.Duration(0),
		SessionIndexes:      utils.StringMap{},
		ClientProtocol:      2.5,
		ChannelSyncInterval: time.Duration(10),
		TerminateAttempts:   6,
		AlterableFields:     utils.StringSet{},
		MinDurLowBalance:    time.Duration(1),
		SchedulerConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler)},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.META_ANY: {}},
			PayloadMaxduration: time.Duration(-1),
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

func TestSessionSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {},

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
		utils.MinCallDurationCfg:     "0",
		utils.MaxCallDurationCfg:     "3h0m0s",
		utils.SessionTTLCfg:          "0",
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      1.0,
		utils.ChannelSyncIntervalCfg: "0",
		utils.TerminateAttemptsCfg:   5,
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sessionSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
func TestSessionSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"sessions": {
			"enabled": true,
			"listen_bijson": "127.0.0.1:2018",
			"chargers_conns": ["*internal"],
			"rals_conns": ["*internal"],
			"cdrs_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"thresholds_conns": ["*internal"],
			"stats_conns": ["*internal"],
			"routes_conns": ["*internal"],
			"attributes_conns": ["*internal"],
			"replication_conns": ["*localhost"],
			"debit_interval": "8s",
			"store_session_costs": true,
			"min_call_duration": "1s",
			"max_call_duration": "1h",
			"session_ttl": "1s",
			"client_protocol": 2.0,
			"terminate_attempts": 10,
			"stir": {
				"allowed_attest": ["any1","any2"],
				"payload_maxduration": "-1",
				"default_attest": "B",
				"publickey_path": "",
				"privatekey_path": "",
			},
			"scheduler_conns": ["*internal"],
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.ListenBijsonCfg:        "127.0.0.1:2018",
		utils.ChargerSConnsCfg:       []string{utils.MetaInternal},
		utils.RALsConnsCfg:           []string{utils.MetaInternal},
		utils.CDRsConnsCfg:           []string{utils.MetaInternal},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal},
		utils.ThresholdSConnsCfg:     []string{utils.MetaInternal},
		utils.StatSConnsCfg:          []string{utils.MetaInternal},
		utils.RouteSConnsCfg:         []string{utils.MetaInternal},
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal},
		utils.ReplicationConnsCfg:    []string{utils.MetaLocalHost},
		utils.DebitIntervalCfg:       "8s",
		utils.StoreSCostsCfg:         true,
		utils.MinCallDurationCfg:     "1s",
		utils.MaxCallDurationCfg:     "1h0m0s",
		utils.SessionTTLCfg:          "1s",
		utils.SessionIndexesCfg:      []string{},
		utils.ClientProtocolCfg:      2.0,
		utils.ChannelSyncIntervalCfg: "0",
		utils.TerminateAttemptsCfg:   10,
		utils.AlterableFieldsCfg:     []string{},
		utils.STIRCfg: map[string]interface{}{
			utils.AllowedAtestCfg:       []string{"any1", "any2"},
			utils.PayloadMaxdurationCfg: "-1",
			utils.DefaultAttestCfg:      "B",
			utils.PublicKeyPathCfg:      "",
			utils.PrivateKeyPathCfg:     "",
		},
		utils.SchedulerConnsCfg: []string{"*internal"},
	}
	cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	rcv := cgrCfg.sessionSCfg.AsMapInterface()
	sort.Strings(rcv[utils.STIRCfg].(map[string]interface{})[utils.AllowedAtestCfg].([]string))
	if !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgloadFromJsonCfg2(t *testing.T) {
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
		MaxWaitConnection:   time.Duration(2),
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

func TestFsAgentCfgAsMapInterface(t *testing.T) {
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.fsAgentCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgAsMapInterface1(t *testing.T) {
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
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
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscocfg)
	}
	if err := fscocfg.loadFromJsonCfg(new(FsConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fscocfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fscocfg)
	}
	json := &FsConnJsonCfg{
		Address:    utils.StringPointer("127.0.0.1:8448"),
		Password:   utils.StringPointer("pass123"),
		Reconnects: utils.IntPointer(5),
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fscocfg))
	}
}

func TestRemoteHostloadFromJsonCfg(t *testing.T) {
	var hpoolcfg, expected RemoteHost
	if err := hpoolcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, hpoolcfg)
	}
	if err := hpoolcfg.loadFromJsonCfg(new(RemoteHostJson)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, hpoolcfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(hpoolcfg))
	}
}

func TestAsteriskAgentCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &AsteriskAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Sessions_conns:         &[]string{utils.MetaInternal},
		Create_cdr:             utils.BoolPointer(true),
		Low_balance_ann_file:   utils.StringPointer("randomFile"),
		Empty_balance_context:  utils.StringPointer("randomContext"),
		Empty_balance_ann_file: utils.StringPointer("randomAnnFile"),
		Asterisk_conns: &[]*AstConnJsonCfg{
			{
				Address:          utils.StringPointer("127.0.0.1:8088"),
				User:             utils.StringPointer(utils.CGRATES),
				Password:         utils.StringPointer("CGRateS.org"),
				Connect_attempts: utils.IntPointer(3),
				Reconnects:       utils.IntPointer(5),
			},
		},
	}
	expected := &AsteriskAgentCfg{
		Enabled:             true,
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		CreateCDR:           true,
		LowBalanceAnnFile:   "randomFile",
		EmptyBalanceContext: "randomContext",
		EmptyBalanceAnnFile: "randomAnnFile",
		AsteriskConns: []*AsteriskConnCfg{{
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
		utils.EnabledCfg:             false,
		utils.SessionSConnsCfg:       []string{"*internal"},
		utils.CreateCdrCfg:           false,
		utils.LowBalanceAnnFileCfg:   utils.EmptyString,
		utils.EmptyBalanceContext:    utils.EmptyString,
		utils.EmptyBalanceAnnFileCfg: utils.EmptyString,
		utils.AsteriskConnsCfg: []map[string]interface{}{
			{utils.AliasCfg: "", utils.AddressCfg: "127.0.0.1:8088", utils.UserCf: "cgrates", utils.Password: "CGRateS.org", utils.ConnectAttemptsCfg: 3, utils.ReconnectsCfg: 5},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
		utils.EnabledCfg:             true,
		utils.SessionSConnsCfg:       []string{"*conn1", "*conn2"},
		utils.CreateCdrCfg:           true,
		utils.LowBalanceAnnFileCfg:   utils.EmptyString,
		utils.EmptyBalanceContext:    utils.EmptyString,
		utils.EmptyBalanceAnnFileCfg: utils.EmptyString,
		utils.AsteriskConnsCfg: []map[string]interface{}{
			{utils.AliasCfg: "", utils.AddressCfg: "127.0.0.1:8089", utils.UserCf: "cgrates", utils.Password: "CGRateS.org", utils.ConnectAttemptsCfg: 5, utils.ReconnectsCfg: 8},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.asteriskAgentCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAsteriskConnCfgloadFromJsonCfg(t *testing.T) {
	var asconcfg, expected AsteriskConnCfg
	if err := asconcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, asconcfg)
	}
	if err := asconcfg.loadFromJsonCfg(new(AstConnJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asconcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, asconcfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(asconcfg))
	}
}
