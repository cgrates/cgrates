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
	"slices"
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
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(eFsAgentConfig), utils.ToJSON(fsAgentCfg))
	}
}

func TestSessionSCfgloadFromJsonCfg(t *testing.T) {
	var sescfg, expected SessionSCfg
	if err := sescfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sescfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, sescfg)
	}
	if err := sescfg.loadFromJsonCfg(new(SessionSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sescfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, sescfg)
	}
	sescfg.DefaultUsage = make(map[string]time.Duration)
	cfgJSONStr := `{
"sessions": {
	"enabled": false,						// starts session manager service: <true|false>
	"listen_bijson": "127.0.0.1:2014",		// address where to listen for bidirectional JSON-RPC requests
	"chargers_conns": [],					// address where to reach the charger service, empty to disable charger functionality: <""|*internal|x.y.z.y:1234>
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"resources_conns": [],					// address where to reach the ResourceS <""|*internal|127.0.0.1:2013>
	"thresholds_conns": [],					// address where to reach the ThresholdS <""|*internal|127.0.0.1:2013>
	"stats_conns": [],						// address where to reach the StatS <""|*internal|127.0.0.1:2013>
	"suppliers_conns": [],					// address where to reach the SupplierS <""|*internal|127.0.0.1:2013>
	"attributes_conns": [],					// address where to reach the AttributeS <""|*internal|127.0.0.1:2013>
	"replication_conns": [],				// replicate sessions towards these session services
	"debit_interval": "0s",					// interval to perform debits on.
	"session_ttl": "0s",					// time after a session with no updates is terminated, not defined by default
	//"session_ttl_max_delay": "",			// activates session_ttl randomization and limits the maximum possible delay
	//"session_ttl_last_used": "",			// tweak LastUsed for sessions timing-out, not defined by default
	//"session_ttl_usage": "",				// tweak Usage for sessions timing-out, not defined by default
	"session_indexes": [],					// index sessions based on these fields for GetActiveSessions API
	"client_protocol": 1.0,					// version of protocol to use when acting as JSON-PRC client <"0","1.0">
	"channel_sync_interval": "0",			// sync channels regularly (0 to disable sync session)
	"default_usage":{						// the usage if the event is missing the usage field
		"*any": "3h",
		"*voice": "3h",
		"*data": "1048576",
		"*sms": "1",
	},
},
}`
	expected = SessionSCfg{
		ListenBijson:     "127.0.0.1:2014",
		ChargerSConns:    []string{},
		RALsConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		ResSConns:        []string{},
		ThreshSConns:     []string{},
		StatSConns:       []string{},
		SupplSConns:      []string{},
		AttrSConns:       []string{},
		CDRsConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)},
		ReplicationConns: []string{},
		SessionIndexes:   map[string]bool{},
		ClientProtocol:   1,
		DefaultUsage: map[string]time.Duration{
			utils.META_ANY: 3 * time.Hour,
			utils.VOICE:    3 * time.Hour,
			utils.DATA:     1048576,
			utils.SMS:      1,
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSesCfg, err := jsnCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sescfg.loadFromJsonCfg(jsnSesCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sescfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(sescfg))
	}
}

func TestSessionSCfgAsMapInterface(t *testing.T) {
	var sescfg SessionSCfg
	sescfg.DefaultUsage = make(map[string]time.Duration)
	cfgJSONStr := `{
	"sessions": {
		"enabled": false,
		"listen_bijson": "127.0.0.1:2014",
		"chargers_conns": [],
		"rals_conns": [],
		"cdrs_conns": [],
		"resources_conns": [],
		"thresholds_conns": [],
		"stats_conns": [],
		"suppliers_conns": [],
		"attributes_conns": [],
		"replication_conns": [],
		"debit_interval": "0s",
		"store_session_costs": false,
		"session_ttl": "0s",
		"session_indexes": [],
		"client_protocol": 1.0,
		"channel_sync_interval": "0",
		"terminate_attempts": 5,
		"alterable_fields": [],
		"stir": {
			"allowed_attest": ["*any"],
			"payload_maxduration": "-1",
			"default_attest": "A",
			"publickey_path": "",
			"privatekey_path": "",
		},
		"scheduler_conns": [],
		"default_usage":{						// the usage if the event is missing the usage field
			"*any": "3h",
			"*voice": "3h",
			"*data": "1048576",
			"*sms": "1",
		},
	},
}`
	eMap := map[string]any{
		"enabled":                false,
		"listen_bijson":          "127.0.0.1:2014",
		"chargers_conns":         []string{},
		"rals_conns":             []string{},
		"cdrs_conns":             []string{},
		"resources_conns":        []string{},
		"thresholds_conns":       []string{},
		"stats_conns":            []string{},
		"suppliers_conns":        []string{},
		"attributes_conns":       []string{},
		"replication_conns":      []string{},
		"debit_interval":         "0",
		"store_session_costs":    false,
		"session_ttl":            "0",
		"session_indexes":        []string{},
		"client_protocol":        1.0,
		"channel_sync_interval":  "0",
		"terminate_attempts":     5,
		"alterable_fields":       []string{},
		"session_ttl_last_used":  "0",
		"session_ttl_max_delay":  "0",
		"session_ttl_usage":      "0",
		"session_ttl_last_usage": "0",
		utils.DefaultUsageCfg: map[string]any{
			utils.META_ANY: "3h0m0s",
			utils.VOICE:    "3h0m0s",
			utils.DATA:     "1048576",
			utils.SMS:      "1",
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSesCfg, err := jsnCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sescfg.loadFromJsonCfg(jsnSesCfg); err != nil {
		t.Error(err)
	} else if rcv := sescfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
	cfgJSONStr = `{
		"sessions": {
			"enabled": false,
			"listen_bijson": "127.0.0.1:2014",
			"chargers_conns": ["*internal"],
			"rals_conns": ["*internal"],
			"cdrs_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"thresholds_conns": ["*internal"],
			"stats_conns": ["*internal"],
			"suppliers_conns": ["*internal"],
			"attributes_conns": ["*internal"],
			"replication_conns": ["*localhost"],
			"debit_interval": "0s",
			"store_session_costs": false,
			"min_call_duration": "0s",
			"max_call_duration": "3h",
			"session_ttl": "0s",
			"session_indexes": [],
			"client_protocol": 1.0,
			"channel_sync_interval": "0",
			"terminate_attempts": 5,
			"alterable_fields": [],
			"stir": {
				"allowed_attest": ["*any"],
				"payload_maxduration": "-1",
				"default_attest": "A",
				"publickey_path": "",
				"privatekey_path": "",
			},
			"scheduler_conns": ["*internal"],
			"default_usage":{						// the usage if the event is missing the usage field
				"*any": "3h",
				"*voice": "3h",
				"*data": "1048576",
				"*sms": "1",
			},
		},
	}`
	eMap = map[string]any{
		"enabled":                false,
		"listen_bijson":          "127.0.0.1:2014",
		"chargers_conns":         []string{"*internal"},
		"rals_conns":             []string{"*internal"},
		"cdrs_conns":             []string{"*internal"},
		"resources_conns":        []string{"*internal"},
		"thresholds_conns":       []string{"*internal"},
		"stats_conns":            []string{"*internal"},
		"suppliers_conns":        []string{"*internal"},
		"attributes_conns":       []string{"*internal"},
		"replication_conns":      []string{"*localhost"},
		"debit_interval":         "0",
		"store_session_costs":    false,
		"session_ttl":            "0",
		"session_indexes":        []string{},
		"client_protocol":        1.0,
		"channel_sync_interval":  "0",
		"terminate_attempts":     5,
		"alterable_fields":       []string{},
		"session_ttl_last_used":  "0",
		"session_ttl_max_delay":  "0",
		"session_ttl_usage":      "0",
		"session_ttl_last_usage": "0",
		utils.DefaultUsageCfg: map[string]any{
			utils.META_ANY: "3h0m0s",
			utils.VOICE:    "3h0m0s",
			utils.DATA:     "1048576",
			utils.SMS:      "1",
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSesCfg, err := jsnCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sescfg.loadFromJsonCfg(jsnSesCfg); err != nil {
		t.Error(err)
	} else if rcv := sescfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgloadFromJsonCfg2(t *testing.T) {
	var fsagcfg, expected FsAgentCfg
	if err := fsagcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fsagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fsagcfg)
	}
	if err := fsagcfg.loadFromJsonCfg(new(FreeswitchAgentJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fsagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, fsagcfg)
	}
	cfgJSONStr := `{
"freeswitch_agent": {
	"enabled": false,						// starts the FreeSWITCH agent: <true|false>
	"sessions_conns": ["*internal"],
	"subscribe_park": true,					// subscribe via fsock to receive park events
	"create_cdr": false,					// create CDR out of events and sends them to CDRS component
	"extra_fields": [],						// extra fields to store in auth/CDRs when creating them
	//"min_dur_low_balance": "5s",			// threshold which will trigger low balance warnings for prepaid calls (needs to be lower than debit_interval)
	//"low_balance_ann_file": "",			// file to be played when low balance is reached for prepaid calls
	"empty_balance_context": "",			// if defined, prepaid calls will be transferred to this context on empty balance
	"empty_balance_ann_file": "",			// file to be played before disconnecting prepaid calls on empty balance (applies only if no context defined)
	"max_wait_connection": "2s",			// maximum duration to wait for a connection to be retrieved from the pool
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5,"alias":""}
	],
},
}`
	expected = FsAgentCfg{
		SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		SubscribePark:     true,
		MaxWaitConnection: time.Duration(2 * time.Second),
		ExtraFields:       RSRParsers{},
		EventSocketConns: []*FsConnCfg{{
			Address:    "127.0.0.1:8021",
			Password:   "ClueCon",
			Reconnects: 5,
			Alias:      "127.0.0.1:8021",
		}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsAgCfg, err := jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fsagcfg.loadFromJsonCfg(jsnFsAgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fsagcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(fsagcfg))
	}
}

func TestFsAgentCfgAsMapInterface(t *testing.T) {
	var fsagcfg FsAgentCfg
	cfgJSONStr := `{
	"freeswitch_agent": {
		"enabled": false,
		"sessions_conns": ["*internal"],
		"subscribe_park": true,
		"create_cdr": false,
		"extra_fields": [],
		//"min_dur_low_balance": "5s",
		//"low_balance_ann_file": "",
		"empty_balance_context": "",
		"empty_balance_ann_file": "",
		"max_wait_connection": "2s",
		"event_socket_conns":[
			{"address": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5,"alias":""}
		],
	},
}`
	eMap := map[string]any{
		"enabled":                false,
		"sessions_conns":         []string{"*internal"},
		"subscribe_park":         true,
		"create_cdr":             false,
		"extra_fields":           "",
		"empty_balance_context":  "",
		"empty_balance_ann_file": "",
		"max_wait_connection":    "2s",
		"event_socket_conns": []map[string]any{
			{"address": "127.0.0.1:8021", "password": "ClueCon", "reconnects": 5, "alias": "127.0.0.1:8021"},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsAgCfg, err := jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fsagcfg.loadFromJsonCfg(jsnFsAgCfg); err != nil {
		t.Error(err)
	} else if rcv := fsagcfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
	var asagcfg, expected AsteriskAgentCfg
	if err := asagcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asagcfg)
	}
	if err := asagcfg.loadFromJsonCfg(new(AsteriskAgentJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asagcfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, asagcfg)
	}
	cfgJSONStr := `{
"asterisk_agent": {
	"enabled": true,						// starts the Asterisk agent: <true|false>
	"sessions_conns": ["*internal"],
	"create_cdr": false,					// create CDR out of events and sends it to CDRS component
	"asterisk_conns":[						// instantiate connections to multiple Asterisk servers
		{"address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3,"reconnects": 5}
	],
},
}`
	expected = AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		AsteriskConns: []*AsteriskConnCfg{{
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAsAgCfg, err := jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = asagcfg.loadFromJsonCfg(jsnAsAgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, asagcfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(asagcfg))
	}
}

func TestAsteriskAgentCfgAsMapInterface(t *testing.T) {
	var asagcfg AsteriskAgentCfg
	cfgJSONStr := `{
	"asterisk_agent": {
		"enabled": true,
		"sessions_conns": ["*internal"],
		"create_cdr": false,
		"alterable_fields": ["field1", "field2"],
		"asterisk_conns":[
			{"address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3,"reconnects": 5}
		],
	},
}`
	eMap := map[string]any{
		"enabled":                true,
		"sessions_conns":         []string{"*internal"},
		"create_cdr":             false,
		utils.AlterableFieldsCfg: []string{"field1", "field2"},
		"asterisk_conns": []map[string]any{
			{"alias": "", "address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3, "reconnects": 5},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAsAgCfg, err := jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = asagcfg.loadFromJsonCfg(jsnAsAgCfg); err != nil {
		t.Error(err)
	} else {
		rcv := asagcfg.AsMapInterface()
		slices.Sort(rcv[utils.AlterableFieldsCfg].([]string))
		if !reflect.DeepEqual(eMap, rcv) {
			t.Errorf("expected: %s, received: %s", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
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

func TestSessionSCfgAsMapInterface2(t *testing.T) {
	bl := false
	str := "test"
	slc := []string{"val1"}
	dr := 1 * time.Millisecond
	drs := "1ms"
	fl := 1.2
	nm := 1

	scfg := SessionSCfg{
		Enabled:             bl,
		ListenBijson:        str,
		ChargerSConns:       slc,
		RALsConns:           slc,
		ResSConns:           slc,
		ThreshSConns:        slc,
		StatSConns:          slc,
		SupplSConns:         slc,
		AttrSConns:          slc,
		CDRsConns:           slc,
		ReplicationConns:    slc,
		DebitInterval:       dr,
		StoreSCosts:         bl,
		SessionTTL:          dr,
		SessionTTLMaxDelay:  &dr,
		SessionTTLLastUsed:  &dr,
		SessionTTLUsage:     &dr,
		SessionTTLLastUsage: &dr,
		SessionIndexes:      utils.StringMap{},
		ClientProtocol:      fl,
		ChannelSyncInterval: dr,
		TerminateAttempts:   nm,
		AlterableFields:     utils.StringSet{},
		DefaultUsage:        map[string]time.Duration{},
	}

	exp := map[string]any{
		utils.EnabledCfg:             scfg.Enabled,
		utils.ListenBijsonCfg:        scfg.ListenBijson,
		utils.ChargerSConnsCfg:       slc,
		utils.RALsConnsCfg:           slc,
		utils.ResSConnsCfg:           slc,
		utils.ThreshSConnsCfg:        slc,
		utils.StatSConnsCfg:          slc,
		utils.SupplSConnsCfg:         slc,
		utils.AttrSConnsCfg:          slc,
		utils.CDRsConnsCfg:           slc,
		utils.ReplicationConnsCfg:    scfg.ReplicationConns,
		utils.DebitIntervalCfg:       drs,
		utils.StoreSCostsCfg:         scfg.StoreSCosts,
		utils.SessionTTLCfg:          drs,
		utils.SessionTTLMaxDelayCfg:  drs,
		utils.SessionTTLLastUsedCfg:  drs,
		utils.SessionTTLUsageCfg:     drs,
		utils.SessionTTLLastUsageCfg: drs,
		utils.SessionIndexesCfg:      scfg.SessionIndexes.Slice(),
		utils.ClientProtocolCfg:      scfg.ClientProtocol,
		utils.ChannelSyncIntervalCfg: drs,
		utils.TerminateAttemptsCfg:   scfg.TerminateAttempts,
		utils.AlterableFieldsCfg:     scfg.AlterableFields.AsSlice(),
		utils.DefaultUsageCfg:        map[string]any{},
	}

	rcv := scfg.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, recived %v", exp, rcv)
	}
}

func TestSMConfigGetDefaultUsage(t *testing.T) {
	s := SessionSCfg{}

	rcv := s.GetDefaultUsage("")
	exp := s.DefaultUsage["*any"]

	if rcv != exp {
		t.Errorf("received %v, expected %v", rcv, exp)
	}
}

func TestSMConfigloadFromJsonCfg(t *testing.T) {
	str := "test"
	slc := []string{str}
	bl := true
	td := 1 * time.Second
	nm := 1
	tds := "1s"
	fl := 1.2
	scfg := SessionSCfg{}

	jsnCfg := &SessionSJsonCfg{
		Enabled:                &bl,
		Listen_bijson:          &str,
		Chargers_conns:         &slc,
		Rals_conns:             &slc,
		Resources_conns:        &slc,
		Thresholds_conns:       &slc,
		Stats_conns:            &slc,
		Suppliers_conns:        &slc,
		Cdrs_conns:             &slc,
		Replication_conns:      &slc,
		Attributes_conns:       &slc,
		Debit_interval:         &tds,
		Store_session_costs:    &bl,
		Session_ttl:            &tds,
		Session_ttl_max_delay:  &tds,
		Session_ttl_last_used:  &tds,
		Session_ttl_usage:      &tds,
		Session_ttl_last_usage: &tds,
		Session_indexes:        &[]string{"test"},
		Client_protocol:        &fl,
		Channel_sync_interval:  &tds,
		Terminate_attempts:     &nm,
		Alterable_fields:       &[]string{},
	}

	err := scfg.loadFromJsonCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}
	exp := SessionSCfg{
		Enabled:             bl,
		ListenBijson:        str,
		ChargerSConns:       slc,
		RALsConns:           slc,
		ResSConns:           slc,
		ThreshSConns:        slc,
		StatSConns:          slc,
		SupplSConns:         slc,
		AttrSConns:          slc,
		CDRsConns:           slc,
		ReplicationConns:    slc,
		DebitInterval:       td,
		StoreSCosts:         bl,
		SessionTTL:          td,
		SessionTTLMaxDelay:  &td,
		SessionTTLLastUsed:  &td,
		SessionTTLUsage:     &td,
		SessionTTLLastUsage: &td,
		SessionIndexes:      utils.StringMap{str: bl},
		ClientProtocol:      fl,
		ChannelSyncInterval: td,
		TerminateAttempts:   nm,
		AlterableFields:     utils.StringSet{},
	}

	if !reflect.DeepEqual(scfg, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(scfg))
	}
}

func TestSMConfigloadFromJsonCfgErrors(t *testing.T) {
	slc := []string{utils.MetaInternal}
	str := "test"
	scfg := SessionSCfg{
		DefaultUsage: map[string]time.Duration{str: 1 * time.Second},
	}
	jsnCfg := &SessionSJsonCfg{
		Replication_conns: &slc,
	}
	jsnCfg2 := &SessionSJsonCfg{
		Debit_interval: &str,
	}
	jsnCfg3 := &SessionSJsonCfg{
		Session_ttl: &str,
	}
	jsnCfg4 := &SessionSJsonCfg{
		Session_ttl_max_delay: &str,
	}
	jsnCfg5 := &SessionSJsonCfg{
		Session_ttl_last_used: &str,
	}
	jsnCfg6 := &SessionSJsonCfg{
		Session_ttl_usage: &str,
	}
	jsnCfg7 := &SessionSJsonCfg{
		Session_ttl_last_usage: &str,
	}
	jsnCfg8 := &SessionSJsonCfg{
		Channel_sync_interval: &str,
	}
	jsnCfg9 := &SessionSJsonCfg{
		Default_usage: &map[string]string{str: str},
	}

	tests := []struct {
		name string
		arg  *SessionSJsonCfg
		err  string
	}{
		{
			name: "Replication_conns error",
			arg:  jsnCfg,
			err:  "Replication connection ID needs to be different than *internal",
		},
		{
			name: "Debit_interval error",
			arg:  jsnCfg2,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Session_ttl error",
			arg:  jsnCfg3,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Session_ttl_max_delay error",
			arg:  jsnCfg4,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Session_ttl_last_used error",
			arg:  jsnCfg5,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Session_ttl_usage error",
			arg:  jsnCfg6,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Session_ttl_last_usage error",
			arg:  jsnCfg7,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Channel_sync_interval error",
			arg:  jsnCfg8,
			err:  `time: invalid duration "test"`,
		},
		{
			name: "Default_usage error",
			arg:  jsnCfg9,
			err:  `time: invalid duration "test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scfg.loadFromJsonCfg(tt.arg)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			} else {
				t.Error("was expecting an error")
			}
		})
	}
}

func TestSMConfigFsAgentCfgloadFromJsonCfgErrors(t *testing.T) {
	str := "test"
	slc := []string{"test"}
	slc2 := []string{"test)"}
	self := &FsAgentCfg{}
	jsnCfg := &FreeswitchAgentJsonCfg{
		Sessions_conns: &slc,
		Extra_fields:   &slc2,
	}

	err := self.loadFromJsonCfg(jsnCfg)

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	self2 := &FsAgentCfg{}
	jsnCfg2 := &FreeswitchAgentJsonCfg{
		Max_wait_connection: &str,
	}

	err = self2.loadFromJsonCfg(jsnCfg2)

	if err != nil {
		if err.Error() != `time: invalid duration "test"` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestSMConfigFsAgentCfgAsMapInterface(t *testing.T) {
	fscfg := &FsAgentCfg{
		SessionSConns: []string{"test"},
		ExtraFields: RSRParsers{{
			Rules: "test",
		}},
		Enabled:             false,
		SubscribePark:       false,
		CreateCdr:           false,
		EmptyBalanceContext: "test",
		EmptyBalanceAnnFile: "test",
		MaxWaitConnection:   1 * time.Second,
		EventSocketConns:    []*FsConnCfg{},
	}

	rcv := fscfg.AsMapInterface(":")
	exp := map[string]any{
		utils.EnabledCfg:             fscfg.Enabled,
		utils.SessionSConnsCfg:       []string{"test"},
		utils.SubscribeParkCfg:       fscfg.SubscribePark,
		utils.CreateCdrCfg:           fscfg.CreateCdr,
		utils.ExtraFieldsCfg:         "test",
		utils.EmptyBalanceContextCfg: fscfg.EmptyBalanceContext,
		utils.EmptyBalanceAnnFileCfg: fscfg.EmptyBalanceAnnFile,
		utils.MaxWaitConnectionCfg:   "1s",
		utils.EventSocketConnsCfg:    []map[string]any{},
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestSMConfigAsteriskConnCfgloadFromJsonCfg(t *testing.T) {
	str := "test"
	aConnCfg := &AsteriskConnCfg{}
	jsnCfg := &AstConnJsonCfg{
		Alias: &str,
	}

	err := aConnCfg.loadFromJsonCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	if aConnCfg.Alias != str {
		t.Error("didn't load")
	}
}

func TestSMConfigAsteriskAgentCfgloadFromJsonCfg(t *testing.T) {
	aCfg := &AsteriskAgentCfg{}
	jsnCfg := &AsteriskAgentJsonCfg{
		Sessions_conns: &[]string{"test"},
	}

	err := aCfg.loadFromJsonCfg(jsnCfg)
	if err != nil {
		t.Error(err)
	}

	if aCfg.SessionSConns[0] != "test" {
		t.Error("didn't load")
	}
}

func TestSMConfigAsteriskAgentCfgAsMapInterface(t *testing.T) {
	str := "test"
	aCfg := &AsteriskAgentCfg{
		Enabled:         false,
		SessionSConns:   []string{str},
		CreateCDR:       false,
		AlterableFields: utils.NewStringSet([]string{"field1", "field2"}),
		AsteriskConns: []*AsteriskConnCfg{
			{
				Alias:           str,
				Address:         str,
				User:            str,
				Password:        str,
				ConnectAttempts: 1,
				Reconnects:      1,
			},
		},
	}
	exp := map[string]any{
		utils.EnabledCfg:         aCfg.Enabled,
		utils.SessionSConnsCfg:   []string{str},
		utils.CreateCDRCfg:       aCfg.CreateCDR,
		utils.AlterableFieldsCfg: []string{"field1", "field2"},
		utils.AsteriskConnsCfg:   []map[string]any{aCfg.AsteriskConns[0].AsMapInterface()},
	}

	rcv := aCfg.AsMapInterface()
	slices.Sort(rcv[utils.AlterableFieldsCfg].([]string))
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
