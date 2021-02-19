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
	var sescfg, expected SessionSCfg
	if err := sescfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sescfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, sescfg)
	}
	if err := sescfg.loadFromJsonCfg(new(SessionSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sescfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, sescfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(sescfg))
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
	eMap := map[string]interface{}{
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
		utils.DefaultUsageCfg: map[string]interface{}{
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
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
	eMap = map[string]interface{}{
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
		utils.DefaultUsageCfg: map[string]interface{}{
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
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFsAgentCfgloadFromJsonCfg2(t *testing.T) {
	var fsagcfg, expected FsAgentCfg
	if err := fsagcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fsagcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fsagcfg)
	}
	if err := fsagcfg.loadFromJsonCfg(new(FreeswitchAgentJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fsagcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, fsagcfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fsagcfg))
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
	eMap := map[string]interface{}{
		"enabled":                false,
		"sessions_conns":         []string{"*internal"},
		"subscribe_park":         true,
		"create_cdr":             false,
		"extra_fields":           "",
		"empty_balance_context":  "",
		"empty_balance_ann_file": "",
		"max_wait_connection":    "2s",
		"event_socket_conns": []map[string]interface{}{
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
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
	var asagcfg, expected AsteriskAgentCfg
	if err := asagcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asagcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, asagcfg)
	}
	if err := asagcfg.loadFromJsonCfg(new(AsteriskAgentJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(asagcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, asagcfg)
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
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(asagcfg))
	}
}

func TestAsteriskAgentCfgAsMapInterface(t *testing.T) {
	var asagcfg AsteriskAgentCfg
	cfgJSONStr := `{
	"asterisk_agent": {
		"enabled": true,
		"sessions_conns": ["*internal"],
		"create_cdr": false,
		"asterisk_conns":[
			{"address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3,"reconnects": 5}
		],
	},
}`
	eMap := map[string]interface{}{
		"enabled":        true,
		"sessions_conns": []string{"*internal"},
		"create_cdr":     false,
		"asterisk_conns": []map[string]interface{}{
			{"alias": "", "address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3, "reconnects": 5},
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAsAgCfg, err := jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = asagcfg.loadFromJsonCfg(jsnAsAgCfg); err != nil {
		t.Error(err)
	} else if rcv := asagcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
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
