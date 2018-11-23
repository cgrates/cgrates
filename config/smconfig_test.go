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
	cfgJSONStr := `{
"sessions": {
	"enabled": false,						// starts session manager service: <true|false>
	"listen_bijson": "127.0.0.1:2014",		// address where to listen for bidirectional JSON-RPC requests
	"chargers_conns": [],					// address where to reach the charger service, empty to disable charger functionality: <""|*internal|x.y.z.y:1234>
	"rals_conns": [
		{"address": "*internal"}			// address where to reach the RALs <""|*internal|127.0.0.1:2013>
	],
	"cdrs_conns": [
		{"address": "*internal"}			// address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	],
	"resources_conns": [],					// address where to reach the ResourceS <""|*internal|127.0.0.1:2013>
	"thresholds_conns": [],					// address where to reach the ThresholdS <""|*internal|127.0.0.1:2013>
	"stats_conns": [],						// address where to reach the StatS <""|*internal|127.0.0.1:2013>
	"suppliers_conns": [],					// address where to reach the SupplierS <""|*internal|127.0.0.1:2013>
	"attributes_conns": [],					// address where to reach the AttributeS <""|*internal|127.0.0.1:2013>
	"session_replication_conns": [],		// replicate sessions towards these session services
	"debit_interval": "0s",					// interval to perform debits on.
	"min_call_duration": "0s",				// only authorize calls with allowed duration higher than this
	"max_call_duration": "3h",				// maximum call duration a prepaid call can last
	"session_ttl": "0s",					// time after a session with no updates is terminated, not defined by default
	//"session_ttl_max_delay": "",			// activates session_ttl randomization and limits the maximum possible delay
	//"session_ttl_last_used": "",			// tweak LastUsed for sessions timing-out, not defined by default
	//"session_ttl_usage": "",				// tweak Usage for sessions timing-out, not defined by default
	"session_indexes": [],					// index sessions based on these fields for GetActiveSessions API
	"client_protocol": 1.0,					// version of protocol to use when acting as JSON-PRC client <"0","1.0">
	"channel_sync_interval": "0",			// sync channels regularly (0 to disable sync session)
},
}`
	expected = SessionSCfg{
		ListenBijson:            "127.0.0.1:2014",
		ChargerSConns:           []*HaPoolConfig{},
		RALsConns:               []*HaPoolConfig{{Address: "*internal"}},
		ResSConns:               []*HaPoolConfig{},
		ThreshSConns:            []*HaPoolConfig{},
		StatSConns:              []*HaPoolConfig{},
		SupplSConns:             []*HaPoolConfig{},
		AttrSConns:              []*HaPoolConfig{},
		CDRsConns:               []*HaPoolConfig{{Address: "*internal"}},
		SessionReplicationConns: []*HaPoolConfig{},
		MaxCallDuration:         time.Duration(3 * time.Hour),
		SessionIndexes:          map[string]bool{},
		ClientProtocol:          1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSesCfg, err := jsnCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sescfg.loadFromJsonCfg(jsnSesCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sescfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(sescfg))
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
	"sessions_conns": [
		{"address": "*internal"}			// connection towards session service: <*internal>
	],
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
		SessionSConns:     []*HaPoolConfig{{Address: "*internal"}},
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
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnFsAgCfg, err := jsnCfg.FreeswitchAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = fsagcfg.loadFromJsonCfg(jsnFsAgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, fsagcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(fsagcfg))
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

func TestHaPoolConfigloadFromJsonCfg(t *testing.T) {
	var hpoolcfg, expected HaPoolConfig
	if err := hpoolcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, hpoolcfg)
	}
	if err := hpoolcfg.loadFromJsonCfg(new(HaPoolJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(hpoolcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, hpoolcfg)
	}
	json := &HaPoolJsonCfg{
		Address:     utils.StringPointer("127.0.0.1:8448"),
		Synchronous: utils.BoolPointer(true),
	}
	expected = HaPoolConfig{
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
	"sessions_conns": [
		{"address": "*internal"}			// connection towards session service: <*internal>
	],
	"create_cdr": false,					// create CDR out of events and sends it to CDRS component
	"asterisk_conns":[						// instantiate connections to multiple Asterisk servers
		{"address": "127.0.0.1:8088", "user": "cgrates", "password": "CGRateS.org", "connect_attempts": 3,"reconnects": 5}
	],
},
}`
	expected = AsteriskAgentCfg{
		Enabled:       true,
		SessionSConns: []*HaPoolConfig{{Address: "*internal"}},
		AsteriskConns: []*AsteriskConnCfg{{
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnAsAgCfg, err := jsnCfg.AsteriskAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = asagcfg.loadFromJsonCfg(jsnAsAgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, asagcfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(asagcfg))
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
