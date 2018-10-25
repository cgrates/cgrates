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

func TestFsAgentConfigLoadFromJsonCfg(t *testing.T) {
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
	eFsAgentConfig := &FsAgentConfig{
		Enabled:       true,
		CreateCdr:     true,
		SubscribePark: true,
		EventSocketConns: []*FsConnConfig{
			{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 5, Alias: "1.2.3.4:8021"},
			{Address: "2.3.4.5:8021", Password: "ClueCon", Reconnects: 5, Alias: "2.3.4.5:8021"},
		},
	}
	fsAgentCfg := new(FsAgentConfig)
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
	} else if jsnSchCfg, err := jsnCfg.SessionSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sescfg.loadFromJsonCfg(jsnSchCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sescfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, sescfg)
	}
}
