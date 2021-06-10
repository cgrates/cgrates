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
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

var cfg *CGRConfig
var err error

func TestNewDefaultConfigError(t *testing.T) {
	if _, err = newCGRConfig([]byte(CGRATES_CFG_JSON)); err != nil {
		t.Error(err)
	}
}

func TestNewCgrConfigFromBytesError(t *testing.T) {
	cfg := []byte(`{
"cores": {
	"caps": "0",
}
}`)
	expected := "json: cannot unmarshal string into Go struct field CoreSJsonCfg.Caps of type int"
	if _, err := newCGRConfig(cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v,\n received %+v", expected, err)
	}
}

func TestNewCgrConfigFromBytesDecodeError(t *testing.T) {
	cfg := []byte(`invalidSection`)
	expected := "invalid character 'i' looking for beginning of value around line 1 and position 1\n line: \"invalidSection\""
	if _, err := newCGRConfig(cfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v,\n received %+q", expected, err)
	}
}

func TestCgrCfgConfigSharing(t *testing.T) {
	cfg = NewDefaultCGRConfig()
	SetCgrConfig(cfg)
	cfgReturn := CgrConfig()
	if !reflect.DeepEqual(cfgReturn, cfg) {
		t.Errorf("Retrieved %v, Expected %v", cfgReturn, cfg)
	}
}

func TestCgrCfgLoadWithDefaults(t *testing.T) {
	jsnCfg := `
{
"freeswitch_agent": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3, "alias":"123"},
		{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5, "alias":"124"}
	],
},

}`
	eCgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	eCgrCfg.fsAgentCfg.Enabled = true
	eCgrCfg.fsAgentCfg.EventSocketConns = []*FsConnCfg{
		{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3, Alias: "123"},
		{Address: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5, Alias: "124"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg)
	}
}

func TestCgrCfgDataDBPortWithoutDynamic(t *testing.T) {
	jsnCfg := `
{
"data_db": {
	"db_type": "mongo",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().Type != utils.Mongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Type, utils.Mongo)
	} else if cgrCfg.DataDbCfg().Port != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Port, "6379")
	}
	jsnCfg = `
{
"data_db": {
	"db_type": "internal",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().Type != utils.Internal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Type, utils.Internal)
	} else if cgrCfg.DataDbCfg().Port != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Port, "6379")
	}
}

func TestCgrCfgDataDBPortWithDymanic(t *testing.T) {
	jsnCfg := `
{
"data_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().Type != utils.Mongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Type, utils.Mongo)
	} else if cgrCfg.DataDbCfg().Port != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Port, "27017")
	}
	jsnCfg = `
{
"data_db": {
	"db_type": "internal",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().Type != utils.Internal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Type, utils.Internal)
	} else if cgrCfg.DataDbCfg().Port != "internal" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().Port, "internal")
	}
}

func TestCgrCfgStorDBPortWithoutDynamic(t *testing.T) {
	jsnCfg := `
{
"stor_db": {
	"db_type": "mongo",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.StorDbCfg().Type != utils.Mongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.Mongo)
	} else if cgrCfg.StorDbCfg().Port != "3306" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Port, "3306")
	}
}

func TestCgrCfgStorDBPortWithDymanic(t *testing.T) {
	jsnCfg := `
{
"stor_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.StorDbCfg().Type != utils.Mongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.Mongo)
	} else if cgrCfg.StorDbCfg().Port != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Port, "27017")
	}
}

func TestCgrCfgListener(t *testing.T) {
	jsnCfg := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.ListenCfg().RPCGOBTLSListen != "127.0.0.1:2023" {
		t.Errorf("Expected: 127.0.0.1:2023 , received: %+v", cgrCfg.ListenCfg().RPCGOBTLSListen)
	} else if cgrCfg.ListenCfg().RPCJSONTLSListen != "127.0.0.1:2022" {
		t.Errorf("Expected: 127.0.0.1:2022 , received: %+v", cgrCfg.ListenCfg().RPCJSONTLSListen)
	}
}

func TestHttpAgentCfg(t *testing.T) {
	jsnCfg := `
{
"http_agent": [
	{
		"id": "conecto1",
		"url": "/conecto",					// relative URL for requests coming in
		"sessions_conns": ["*internal"],
		"request_payload":	"*url",			// source of input data <*url>
		"reply_payload":	"*xml",			// type of output data <*xml>
		"request_processors": [],
	}
],
}
	`
	eCgrCfg := NewDefaultCGRConfig()
	eCgrCfg.httpAgentCfg = []*HTTPAgentCfg{
		{
			ID:                "conecto1",
			URL:               "/conecto",
			RequestPayload:    utils.MetaUrl,
			ReplyPayload:      utils.MetaXml,
			SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
			RequestProcessors: nil,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.HTTPAgentCfg(), cgrCfg.HTTPAgentCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.httpAgentCfg), utils.ToJSON(cgrCfg.httpAgentCfg))
	}
}

func TestCgrCfgJSONDefaultsGeneral(t *testing.T) {
	if cgrCfg.GeneralCfg().RoundingDecimals != 5 {
		t.Errorf("Expected: 5, received: %+v", cgrCfg.GeneralCfg().RoundingDecimals)
	}
	if cgrCfg.GeneralCfg().DBDataEncoding != "msgpack" {
		t.Errorf("Expected: msgpack, received: %+v", cgrCfg.GeneralCfg().DBDataEncoding)
	}
	if expected := "/var/spool/cgrates/tpe"; cgrCfg.GeneralCfg().TpExportPath != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, cgrCfg.GeneralCfg().TpExportPath)
	}
	if cgrCfg.GeneralCfg().PosterAttempts != 3 {
		t.Errorf("Expected: 3, received: %+v", cgrCfg.GeneralCfg().PosterAttempts)
	}
	if expected := "/var/spool/cgrates/failed_posts"; cgrCfg.GeneralCfg().FailedPostsDir != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, cgrCfg.GeneralCfg().FailedPostsDir)
	}
	if cgrCfg.GeneralCfg().DefaultReqType != "*rated" {
		t.Errorf("Expected: *rated, received: %+v", cgrCfg.GeneralCfg().DefaultReqType)
	}
	if cgrCfg.GeneralCfg().DefaultCategory != "call" {
		t.Errorf("Expected: call, received: %+v", cgrCfg.GeneralCfg().DefaultCategory)
	}
	if cgrCfg.GeneralCfg().DefaultTenant != "cgrates.org" {
		t.Errorf("Expected: cgrates.org, received: %+v", cgrCfg.GeneralCfg().DefaultTenant)
	}
	if cgrCfg.GeneralCfg().DefaultCaching != utils.MetaReload {
		t.Errorf("Expected: *reload, received: %+v", cgrCfg.GeneralCfg().DefaultCaching)
	}
	if cgrCfg.GeneralCfg().DefaultTimezone != "Local" {
		t.Errorf("Expected: Local, received: %+v", cgrCfg.GeneralCfg().DefaultTimezone)
	}
	if cgrCfg.GeneralCfg().ConnectAttempts != 5 {
		t.Errorf("Expected: 3, received: %+v", cgrCfg.GeneralCfg().ConnectAttempts)
	}
	if cgrCfg.GeneralCfg().Reconnects != -1 {
		t.Errorf("Expected: -1, received: %+v", cgrCfg.GeneralCfg().Reconnects)
	}
	if cgrCfg.GeneralCfg().ConnectTimeout != time.Second {
		t.Errorf("Expected: 1s, received: %+v", cgrCfg.GeneralCfg().ConnectTimeout)
	}
	if cgrCfg.GeneralCfg().ReplyTimeout != 2*time.Second {
		t.Errorf("Expected: 2s, received: %+v", cgrCfg.GeneralCfg().ReplyTimeout)
	}
	if cgrCfg.GeneralCfg().LockingTimeout != 0 {
		t.Errorf("Expected: 0, received: %+v", cgrCfg.GeneralCfg().LockingTimeout)
	}
	if cgrCfg.GeneralCfg().Logger != utils.MetaSysLog {
		t.Errorf("Expected: %+v, received: %+v", utils.MetaSysLog, cgrCfg.GeneralCfg().Logger)
	}
	if cgrCfg.GeneralCfg().LogLevel != 6 {
		t.Errorf("Expected: 6, received: %+v", cgrCfg.GeneralCfg().LogLevel)
	}
	if cgrCfg.GeneralCfg().DigestSeparator != "," {
		t.Errorf("Expected: utils.CSVSep , received: %+v", cgrCfg.GeneralCfg().DigestSeparator)
	}
	if cgrCfg.GeneralCfg().DigestEqual != ":" {
		t.Errorf("Expected: ':' , received: %+v", cgrCfg.GeneralCfg().DigestEqual)
	}
}

func TestCgrCfgJSONDefaultsListen(t *testing.T) {
	if cgrCfg.ListenCfg().RPCJSONListen != "127.0.0.1:2012" {
		t.Errorf("Expected: 127.0.0.1:2012 , received: %+v", cgrCfg.ListenCfg().RPCJSONListen)
	}
	if cgrCfg.ListenCfg().RPCGOBListen != "127.0.0.1:2013" {
		t.Errorf("Expected: 127.0.0.1:2013 , received: %+v", cgrCfg.ListenCfg().RPCGOBListen)
	}
	if cgrCfg.ListenCfg().HTTPListen != "127.0.0.1:2080" {
		t.Errorf("Expected: 127.0.0.1:2080 , received: %+v", cgrCfg.ListenCfg().HTTPListen)
	}
	if cgrCfg.ListenCfg().RPCJSONTLSListen != "127.0.0.1:2022" {
		t.Errorf("Expected: 127.0.0.1:2022 , received: %+v", cgrCfg.ListenCfg().RPCJSONListen)
	}
	if cgrCfg.ListenCfg().RPCGOBTLSListen != "127.0.0.1:2023" {
		t.Errorf("Expected: 127.0.0.1:2023 , received: %+v", cgrCfg.ListenCfg().RPCGOBListen)
	}
	if cgrCfg.ListenCfg().HTTPTLSListen != "127.0.0.1:2280" {
		t.Errorf("Expected: 127.0.0.1:2280 , received: %+v", cgrCfg.ListenCfg().HTTPListen)
	}
}

func TestCgrCfgJSONDefaultsjsnDataDb(t *testing.T) {
	if cgrCfg.DataDbCfg().Type != "redis" {
		t.Errorf("Expecting: redis , received: %+v", cgrCfg.DataDbCfg().Type)
	}
	if cgrCfg.DataDbCfg().Host != "127.0.0.1" {
		t.Errorf("Expecting: 127.0.0.1 , received: %+v", cgrCfg.DataDbCfg().Host)
	}
	if cgrCfg.DataDbCfg().Port != "6379" {
		t.Errorf("Expecting: 6379 , received: %+v", cgrCfg.DataDbCfg().Port)
	}
	if cgrCfg.DataDbCfg().Name != "10" {
		t.Errorf("Expecting: 10 , received: %+v", cgrCfg.DataDbCfg().Name)
	}
	if cgrCfg.DataDbCfg().User != "cgrates" {
		t.Errorf("Expecting: cgrates , received: %+v", cgrCfg.DataDbCfg().User)
	}
	if cgrCfg.DataDbCfg().Password != "" {
		t.Errorf("Expecting:  , received: %+v", cgrCfg.DataDbCfg().Password)
	}
	if len(cgrCfg.DataDbCfg().RmtConns) != 0 {
		t.Errorf("Expecting:  0, received: %+v", len(cgrCfg.DataDbCfg().RmtConns))
	}
	if len(cgrCfg.DataDbCfg().RplConns) != 0 {
		t.Errorf("Expecting:  0, received: %+v", len(cgrCfg.DataDbCfg().RplConns))
	}
}

func TestCgrCfgJSONDefaultsStorDB(t *testing.T) {
	if cgrCfg.StorDbCfg().Type != "mysql" {
		t.Errorf("Expecting: mysql , received: %+v", cgrCfg.StorDbCfg().Type)
	}
	if cgrCfg.StorDbCfg().Host != "127.0.0.1" {
		t.Errorf("Expecting: 127.0.0.1 , received: %+v", cgrCfg.StorDbCfg().Host)
	}
	if cgrCfg.StorDbCfg().Port != "3306" {
		t.Errorf("Expecting: 3306 , received: %+v", cgrCfg.StorDbCfg().Port)
	}
	if cgrCfg.StorDbCfg().Name != "cgrates" {
		t.Errorf("Expecting: cgrates , received: %+v", cgrCfg.StorDbCfg().Name)
	}
	if cgrCfg.StorDbCfg().User != "cgrates" {
		t.Errorf("Expecting: cgrates , received: %+v", cgrCfg.StorDbCfg().User)
	}
	if cgrCfg.StorDbCfg().Password != "" {
		t.Errorf("Expecting: , received: %+v", cgrCfg.StorDbCfg().Password)
	}
	if cgrCfg.StorDbCfg().Opts[utils.SQLMaxOpenConnsCfg] != 100. {
		t.Errorf("Expecting: 100 , received: %+v", cgrCfg.StorDbCfg().Opts[utils.SQLMaxOpenConnsCfg])
	}
	if cgrCfg.StorDbCfg().Opts[utils.SQLMaxIdleConnsCfg] != 10. {
		t.Errorf("Expecting: 10 , received: %+v", cgrCfg.StorDbCfg().Opts[utils.SQLMaxIdleConnsCfg])
	}
	if !reflect.DeepEqual(cgrCfg.StorDbCfg().StringIndexedFields, []string{}) {
		t.Errorf("Expecting: %+v , received: %+v", []string{}, cgrCfg.StorDbCfg().StringIndexedFields)
	}
	if !reflect.DeepEqual(cgrCfg.StorDbCfg().PrefixIndexedFields, []string{}) {
		t.Errorf("Expecting: %+v , received: %+v", []string{}, cgrCfg.StorDbCfg().PrefixIndexedFields)
	}
}

func TestCgrCfgJSONDefaultsCDRS(t *testing.T) {
	eCdrsCfg := &CdrsCfg{
		Enabled:         false,
		StoreCdrs:       true,
		SMCostRetries:   5,
		ChargerSConns:   []string{},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		ActionSConns:    []string{},
		EEsConns:        []string{},
		ExtraFields:     RSRParsers{},
	}
	if !reflect.DeepEqual(eCdrsCfg, cgrCfg.cdrsCfg) {
		t.Errorf("Expecting: %+v , received: %+v", eCdrsCfg, cgrCfg.cdrsCfg)
	}
}

func TestCgrCfgJSONLoadCDRS(t *testing.T) {
	jsnCfg := `
{
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
},
}
	`
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg)
	if err != nil {
		t.Error(err)
	}
	if !cgrCfg.CdrsCfg().Enabled {
		t.Errorf("Expecting: true , received: %+v", cgrCfg.CdrsCfg().Enabled)
	}
	expected := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().ChargerSConns, expected) {
		t.Errorf("Expecting: %+v , received: %+v", expected, cgrCfg.CdrsCfg().ChargerSConns)
	}
}

func TestCgrCfgJSONDefaultsSMGenericCfg(t *testing.T) {
	eSessionSCfg := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		CDRsConns:           []string{},
		ResSConns:           []string{},
		ThreshSConns:        []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttrSConns:          []string{},
		ReplicationConns:    []string{},
		DebitInterval:       0,
		StoreSCosts:         false,
		SessionTTL:          0,
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.MetaAny}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		ActionSConns: []string{},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
	}
	if !reflect.DeepEqual(eSessionSCfg, cgrCfg.sessionSCfg) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSessionSCfg), utils.ToJSON(cgrCfg.sessionSCfg))
	}
}

func TestCgrCfgJSONDefaultsCacheCFG(t *testing.T) {
	eCacheCfg := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheResourceProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheResources: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheEventResources: {Limit: -1,
				TTL: 0, StaticTTL: false},
			utils.CacheStatQueueProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheStatQueues: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheThresholdProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheThresholds: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheFilters: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRouteProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAttributeProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheChargerProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRateProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherHosts: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheActionProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAccounts: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheResourceFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheStatFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheThresholdFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRouteFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAttributeFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheChargerFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRateProfilesFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRateFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheActionProfilesFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAccountsFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheReverseFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherRoutes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherLoads: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatchers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDiameterMessages: {Limit: -1,
				TTL: 3 * time.Hour, StaticTTL: false},
			utils.CacheRPCResponses: {Limit: 0,
				TTL: 2 * time.Second, StaticTTL: false},
			utils.CacheClosedSessions: {Limit: -1,
				TTL: 10 * time.Second, StaticTTL: false},
			utils.CacheEventCharges: {Limit: -1,
				TTL: 10 * time.Second, StaticTTL: false},
			utils.CacheCDRIDs: {Limit: -1,
				TTL: 10 * time.Minute, StaticTTL: false},
			utils.CacheLoadIDs: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRPCConnections: {Limit: -1,
				TTL: 0, StaticTTL: false},
			utils.CacheUCH: {Limit: -1,
				TTL: 3 * time.Hour, StaticTTL: false},
			utils.CacheSTIR: {Limit: -1,
				TTL: 3 * time.Hour, StaticTTL: false},
			utils.CacheCapsEvents: {Limit: -1},

			utils.CacheVersions: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPResources: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPStats: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPThresholds: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPFilters: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheSessionCostsTBL: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheCDRsTBL: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPRoutes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPAttributes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPChargers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPDispatchers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPDispatcherHosts: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPActionProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPRateProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPAccounts: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.MetaAPIBan: {Limit: -1,
				TTL: 2 * time.Minute, StaticTTL: false, Precache: false},
			utils.CacheReplicationHosts: {Limit: 0,
				TTL: 0, StaticTTL: false, Precache: false},
		},
		ReplicationConns: []string{},
	}

	if !reflect.DeepEqual(eCacheCfg, cgrCfg.CacheCfg()) {
		t.Errorf("received: %s, \nexpecting: %s",
			utils.ToJSON(cgrCfg.CacheCfg()), utils.ToJSON(eCacheCfg))
	}
}

func TestCgrCfgJSONDefaultsFsAgentConfig(t *testing.T) {
	eFsAgentCfg := &FsAgentCfg{
		Enabled:             false,
		SessionSConns:       []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:       true,
		CreateCdr:           false,
		ExtraFields:         RSRParsers{},
		EmptyBalanceContext: "",
		EmptyBalanceAnnFile: "",
		MaxWaitConnection:   2 * time.Second,
		EventSocketConns: []*FsConnCfg{{
			Address:    "127.0.0.1:8021",
			Password:   "ClueCon",
			Reconnects: 5,
			Alias:      "127.0.0.1:8021",
		}},
	}

	if !reflect.DeepEqual(cgrCfg.fsAgentCfg, eFsAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.fsAgentCfg), utils.ToJSON(eFsAgentCfg))
	}
}

func TestCgrCfgJSONDefaultsKamAgentConfig(t *testing.T) {
	eKamAgentCfg := &KamAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCdr:     false,
		EvapiConns: []*KamConnCfg{{
			Address:    "127.0.0.1:8448",
			Reconnects: 5,
		}},
	}
	if !reflect.DeepEqual(cgrCfg.kamAgentCfg, eKamAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v",
			utils.ToJSON(cgrCfg.kamAgentCfg), utils.ToJSON(eKamAgentCfg))
	}
}

func TestCgrCfgJSONDefaultssteriskAgentCfg(t *testing.T) {
	eAstAgentCfg := &AsteriskAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{
			{Address: "127.0.0.1:8088",
				User: "cgrates", Password: "CGRateS.org",
				ConnectAttempts: 3, Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.asteriskAgentCfg, eAstAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.asteriskAgentCfg, eAstAgentCfg)
	}
}

func TestCgrCfgJSONDefaultFiltersCfg(t *testing.T) {
	eFiltersCfg := &FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
		AdminSConns:    []string{},
	}
	if !reflect.DeepEqual(cgrCfg.filterSCfg, eFiltersCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.filterSCfg, eFiltersCfg)
	}
}

func TestCgrCfgJSONDefaultSChargerSCfg(t *testing.T) {
	eChargerSCfg := &ChargerSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		AttributeSConns:     []string{},
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
	}
	if !reflect.DeepEqual(eChargerSCfg, cgrCfg.chargerSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eChargerSCfg, cgrCfg.chargerSCfg)
	}
}

func TestCgrCfgJSONDefaultsResLimCfg(t *testing.T) {
	eResLiCfg := &ResourceSConfig{
		Enabled:             false,
		IndexedSelects:      true,
		ThresholdSConns:     []string{},
		StoreInterval:       0,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
	}
	if !reflect.DeepEqual(cgrCfg.resourceSCfg, eResLiCfg) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eResLiCfg), utils.ToJSON(cgrCfg.resourceSCfg))
	}

}

func TestCgrCfgJSONDefaultStatsCfg(t *testing.T) {
	eStatsCfg := &StatSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StoreInterval:       0,
		ThresholdSConns:     []string{},
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
	}
	if !reflect.DeepEqual(cgrCfg.statsCfg, eStatsCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.statsCfg, eStatsCfg)
	}
}

func TestCgrCfgJSONDefaultThresholdSCfg(t *testing.T) {
	eThresholdSCfg := &ThresholdSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StoreInterval:       0,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		ActionSConns:        []string{},
	}
	if !reflect.DeepEqual(eThresholdSCfg, cgrCfg.thresholdSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eThresholdSCfg, cgrCfg.thresholdSCfg)
	}
}

func TestCgrCfgJSONDefaultRouteSCfg(t *testing.T) {
	eSupplSCfg := &RouteSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		DefaultRatio:        1,
	}
	if !reflect.DeepEqual(eSupplSCfg, cgrCfg.routeSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eSupplSCfg, cgrCfg.routeSCfg)
	}
}

func TestCgrCfgJSONDefaultsDiameterAgentCfg(t *testing.T) {
	testDA := &DiameterAgentCfg{
		Enabled:           false,
		Listen:            "127.0.0.1:3868",
		ListenNet:         utils.TCP,
		DictionariesPath:  "/usr/share/cgrates/diameter/dict/",
		SessionSConns:     []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorID:          0,
		ProductName:       "CGRateS",
		ConcurrentReqs:    -1,
		RequestProcessors: nil,
	}

	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Listen, testDA.Listen) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Listen, testDA.Listen)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.DictionariesPath, testDA.DictionariesPath) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.DictionariesPath, testDA.DictionariesPath)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.OriginHost, testDA.OriginHost) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.OriginHost, testDA.OriginHost)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.OriginRealm, testDA.OriginRealm) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.OriginRealm, testDA.OriginRealm)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.VendorID, testDA.VendorID) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.VendorID, testDA.VendorID)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.ProductName, testDA.ProductName) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.ProductName, testDA.ProductName)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.ConcurrentReqs, testDA.ConcurrentReqs) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.ConcurrentReqs, testDA.ConcurrentReqs)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.RequestProcessors, testDA.RequestProcessors) {
		t.Errorf("expecting: %+v, received: %+v", testDA.RequestProcessors, cgrCfg.diameterAgentCfg.RequestProcessors)
	}
}

func TestCgrCfgJSONDefaultsSureTax(t *testing.T) {
	localt, err := time.LoadLocation("Local")
	if err != nil {
		t.Error("time parsing error", err)
	}
	eSureTaxCfg := &SureTaxCfg{
		URL:                  "",
		ClientNumber:         "",
		ValidationKey:        "",
		BusinessUnit:         "",
		Timezone:             localt,
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep),
		BillToNumber:         NewRSRParsersMustCompile("", utils.InfieldSep),
		Zipcode:              NewRSRParsersMustCompile("", utils.InfieldSep),
		P2PZipcode:           NewRSRParsersMustCompile("", utils.InfieldSep),
		P2PPlus4:             NewRSRParsersMustCompile("", utils.InfieldSep),
		Units:                NewRSRParsersMustCompile("1", utils.InfieldSep),
		UnitType:             NewRSRParsersMustCompile("00", utils.InfieldSep),
		TaxIncluded:          NewRSRParsersMustCompile("0", utils.InfieldSep),
		TaxSitusRule:         NewRSRParsersMustCompile("04", utils.InfieldSep),
		TransTypeCode:        NewRSRParsersMustCompile("010101", utils.InfieldSep),
		SalesTypeCode:        NewRSRParsersMustCompile("R", utils.InfieldSep),
		TaxExemptionCodeList: NewRSRParsersMustCompile("", utils.InfieldSep),
	}

	if !reflect.DeepEqual(cgrCfg.sureTaxCfg, eSureTaxCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.sureTaxCfg, eSureTaxCfg)
	}
}

func TestCgrCfgJSONDefaultsHTTP(t *testing.T) {
	if cgrCfg.HTTPCfg().JsonRPCURL != "/jsonrpc" {
		t.Errorf("expecting: /jsonrpc , received: %+v", cgrCfg.HTTPCfg().JsonRPCURL)
	}
	if cgrCfg.HTTPCfg().WSURL != "/ws" {
		t.Errorf("expecting: /ws , received: %+v", cgrCfg.HTTPCfg().WSURL)
	}
	if cgrCfg.HTTPCfg().FreeswitchCDRsURL != "/freeswitch_json" {
		t.Errorf("expecting: /freeswitch_json , received: %+v", cgrCfg.HTTPCfg().FreeswitchCDRsURL)
	}
	if cgrCfg.HTTPCfg().CDRsURL != "/cdr_http" {
		t.Errorf("expecting: /cdr_http , received: %+v", cgrCfg.HTTPCfg().CDRsURL)
	}
	if cgrCfg.HTTPCfg().UseBasicAuth != false {
		t.Errorf("expecting: false , received: %+v", cgrCfg.HTTPCfg().UseBasicAuth)
	}
	if !reflect.DeepEqual(cgrCfg.HTTPCfg().AuthUsers, map[string]string{}) {
		t.Errorf("expecting: %+v , received: %+v", map[string]string{}, cgrCfg.HTTPCfg().AuthUsers)
	}
}

func TestRadiusAgentCfg(t *testing.T) {
	testRA := &RadiusAgentCfg{
		Enabled:            false,
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		RequestProcessors:  nil,
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg, testRA) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg, testRA)
	}
}

func TestDbDefaultsMetaDynamic(t *testing.T) {
	dbdf := newDbDefaults()
	flagInput := utils.MetaDynamic
	dbs := []string{utils.Mongo, utils.Redis, utils.MySQL, utils.Internal}
	for _, dbtype := range dbs {
		port := dbdf.dbPort(dbtype, flagInput)
		if port != dbdf[dbtype]["DbPort"] {
			t.Errorf("received: %+v, expecting: %+v", port, dbdf[dbtype]["DbPort"])
		}
		name := dbdf.dbName(dbtype, flagInput)
		if name != dbdf[dbtype]["DbName"] {
			t.Errorf("received: %+v, expecting: %+v", name, dbdf[dbtype]["DbName"])
		}
	}
}

func TestDbDefaults(t *testing.T) {
	dbdf := newDbDefaults()
	dbs := []string{utils.Mongo, utils.Redis, utils.MySQL, utils.Internal, utils.Postgres}
	for _, dbtype := range dbs {
		port := dbdf.dbPort(dbtype, "1234")
		if port != "1234" {
			t.Errorf("Expected %+v, received %+v", "1234", port)
		}
		name := dbdf.dbName(dbtype, utils.CGRateSLwr)
		if name != utils.CGRateSLwr {
			t.Errorf("Expected %+v, received %+v", utils.CGRateSLwr, name)
		}
	}
}

func TestNewCGRConfigFromJSONStringWithDefaultsError(t *testing.T) {
	cfgJSONStr := "invalidJSON"
	expectedErr := "invalid character 'i' looking for beginning of value around line 1 and position 1\n line: \"invalidJSON\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	cfgJSONStr = `{
	"suretax": {
        "tax_exemption_code_list": "a{*"
    },
}`
	expectedErr = "invalid converter terminator in rule: <a{*>"
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestIsHidden(t *testing.T) {
	file := ".newFile"
	if !isHidden(file) {
		t.Errorf("File is not hidden")
	}

	file = "."
	if isHidden(file) {
		t.Errorf("Invalid input")
	}
}

func TestLazySanityCheck(t *testing.T) {
	cfgJSONStr := `{
      "cdrs": {
	     "online_cdr_exports":["http_localhost", "amqp_localhost", "aws_test_file", "sqs_test_file", "kafka_localhost", "s3_test_file"],
	  },
      "ees": {
            "exporters": [
            {
                  "id": "http_localhost",
			      "type": "*s3_json_map",
			      "fields":[
                      {"tag": "CGRID", "path": "*exp.CGRID", "type": "*variable", "value": "~*req.CGRID"}
                  ]
            }]
	  }
},
`
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	cgrCfg.LazySanityCheck()
}

func TestLoadRPCConnsError(t *testing.T) {
	cfgJSONStr := `{
     "rpc_conns": {
	     "*localhost": {
		     "conns": [
                  {"address": "127.0.0.1:2018", "TLS": true, "synchronous": true, "transport": "*json"},
             ],
             "poolSize": "two",
	      },
     },		
}`
	expected := "json: cannot unmarshal string into Go struct field RPCConnJson.PoolSize of type int"
	cgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrJSONCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrCfg.loadRPCConns(cgrJSONCfg); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadGeneralCfgError(t *testing.T) {
	cfgJSONStr := `{
      "general": {
            "node_id": [],
            "locking_timeout": "0",
            "failed_posts_ttl": "0s",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
        }
    }
}`
	expected := "json: cannot unmarshal array into Go struct field GeneralJsonCfg.Node_id of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadGeneralCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCacheCfgError(t *testing.T) {
	cfgJSONStr := `{
"caches":{
    "replication_conns": 2,
	},
}`
	expected := "json: cannot unmarshal number into Go struct field CacheJsonCfg.Replication_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadCacheCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadListenCfgError(t *testing.T) {
	cfgJSONStr := `{
	"listen": {	
        "http_tls": 1206,			
	}
}`
	expected := "json: cannot unmarshal number into Go struct field ListenJsonCfg.Http_tls of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadListenCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadHTTPCfgError(t *testing.T) {
	cfgJSONStr := `{
	"http": {
	   "auth_users": "user1",
     },
}`
	expected := "json: cannot unmarshal string into Go struct field HTTPJsonCfg.Auth_users of type map[string]string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadHTTPCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDataDBCfgErrorCase1(t *testing.T) {
	cfgJSONStr := `{
"data_db": {
	"db_host": 127.0,
	}
}`
	expected := "json: cannot unmarshal number into Go struct field DbJsonCfg.Db_host of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDataDBCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDataDBCfgErrorCase2(t *testing.T) {
	cfgJSONStr := `{
"data_db": {
	"remote_conns":["*internal"],
	}
}`
	expected := "Remote connection ID needs to be different than <*internal> "
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else {
		cgrConfig.dataDbCfg.RmtConns = []string{utils.MetaInternal}
		if err := cgrConfig.loadDataDBCfg(cgrCfgJSON); err == nil || err.Error() != expected {
			t.Errorf("Expected %+v, received %+v", expected, err)
		}
	}
}

func TestLoadStorDbCfgError(t *testing.T) {
	cfgJSONStr := `{
"stor_db": {
	"db_type": "*internal",
	"db_port": "-1",
	}
}`
	expected := "json: cannot unmarshal string into Go struct field DbJsonCfg.Db_port of type int"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadStorDBCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadFilterSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": "*internal",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field FilterSJsonCfg.Stats_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadFilterSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCdrsCfgError(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {	
        "ees_conns": "*internal",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field CdrsJsonCfg.Ees_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadCdrsCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSessionSCfgError(t *testing.T) {
	cfgJSONStr := `{
	"sessions": {
          "session_ttl_usage": 1,
    },
}`
	expected := "json: cannot unmarshal number into Go struct field SessionSJsonCfg.Session_ttl_usage of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSessionSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadFreeswitchAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"freeswitch_agent": {
          "sessions_conns": "*conn1",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field FreeswitchAgentJsonCfg.Sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadFreeswitchAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadKamAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
		"kamailio_agent": {
			"timezone": 1234,
		},
	}`
	expected := "json: cannot unmarshal number into Go struct field KamAgentJsonCfg.Timezone of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadKamAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAsteriskAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"asterisk_agent": {
		"sessions_conns": "*conn1",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field AsteriskAgentJsonCfg.Sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAsteriskAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDiameterAgentCfgError(t *testing.T) {
	cfgJSONStr := `{ 
      "diameter_agent": {
        "request_processors": [
	        {
		       "id": 1,
            },
         ]
      }
}`
	expected := "json: cannot unmarshal number into Go struct field ReqProcessorJsnCfg.Request_processors.ID of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDiameterAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRadiusAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"radius_agent": {	
         "listen_auth": 1,
     },
}`
	expected := "json: cannot unmarshal number into Go struct field RadiusAgentJsonCfg.Listen_auth of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRadiusAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDNSAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
		"dns_agent": {
			"listen": 1278,
		},
	}`
	expected := "json: cannot unmarshal number into Go struct field DNSAgentJsonCfg.Listen of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDNSAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadHttpAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
"http_agent": [
	{
		"id": ["randomID"],
		},
	],	
}`
	expected := "json: cannot unmarshal array into Go struct field HttpAgentJsonCfg.Id of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadHTTPAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAttributeSCfgError(t *testing.T) {
	cfgJSONStr := `{
"attributes": {
	"process_runs": "3",						
	},		
}`
	expected := "json: cannot unmarshal string into Go struct field AttributeSJsonCfg.Process_runs of type int"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAttributeSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadChargerSCfgError(t *testing.T) {
	cfgJSONStr := `{
	"chargers": {
		"prefix_indexed_fields": "prefix",	
	},	
}`
	expected := "json: cannot unmarshal string into Go struct field ChargerSJsonCfg.Prefix_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadChargerSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadResourceSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"resources": {
            "string_indexed_fields": "*req.index1",
		},	
	}`
	expected := "json: cannot unmarshal string into Go struct field ResourceSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadResourceSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadStatSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"stats": {								
            "string_indexed_fields": "*req.string",
		},	
}`
	expected := "json: cannot unmarshal string into Go struct field StatServJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadStatSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadThresholdSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"thresholds": {
			"store_interval": 96,						
		},		
}`
	expected := "json: cannot unmarshal number into Go struct field ThresholdSJsonCfg.Store_interval of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadThresholdSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadLoaderSCfgError(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"run_delay": 0,
		},
	],	
}`
	expected := "json: cannot unmarshal number into Go struct field LoaderJsonCfg.Run_delay of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadLoaderSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRouteSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"routes": {
            "string_indexed_fields": "*req.string",
		},
	}`
	expected := "json: cannot unmarshal string into Go struct field RouteSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRouteSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSureTaxCfgError(t *testing.T) {
	cfgJSONStr := `{
	"suretax": {
		"sales_type_code": 123,
    },
}`
	expected := "json: cannot unmarshal number into Go struct field SureTaxJsonCfg.Sales_type_code of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSureTaxCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDispatcherSCfgError(t *testing.T) {
	cfgJSONStr := `{
		"dispatchers":{
			"attributes_conns": "*internal",
		},
}`
	expected := "json: cannot unmarshal string into Go struct field DispatcherSJsonCfg.Attributes_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDispatcherSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDispatcherHCfgError(t *testing.T) {
	cfgJSONStr := `{
		"registrarc":{
			"dispatchers":{
				"refresh_interval": 5,
			},		
		},		
}`
	expected := "json: cannot unmarshal number into Go struct field RegistrarCJsonCfg.Dispatchers.Refresh_interval of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRegistrarCCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadLoaderCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
	"loader": {
		"caches_conns":"*localhost",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field LoaderCfgJson.Caches_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadLoaderCgrCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadMigratorCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
	"migrator": {
        "users_filters": "users",
	},
}`
	expected := "json: cannot unmarshal string into Go struct field MigratorCfgJson.Users_filters of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadMigratorCgrCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadTlsCgrCfgError(t *testing.T) {
	cfgJSONStr := `	{
	"tls":{
		"server_policy": "3",					
	},
}`
	expected := "json: cannot unmarshal string into Go struct field TlsJsonCfg.Server_policy of type int"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadTLSCgrCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAnalyzerCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{
            "enabled": 10,  
        },
    }
}`
	expected := "json: cannot unmarshal number into Go struct field AnalyzerSJsonCfg.Enabled of type bool"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAnalyzerCgrCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAPIBanCgrCfgError(t *testing.T) {
	cfgJSONStr := `{
		"apiban":{
			"enabled": "no",
		},
}`
	expected := "json: cannot unmarshal string into Go struct field APIBanJsonCfg.Enabled of type bool"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAPIBanCgrCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadApierCfgError(t *testing.T) {
	myJSONStr := `{
    "admins": {
       "actions_conns": "*internal",
    },
}`
	expected := "json: cannot unmarshal string into Go struct field AdminSJsonCfg.Actions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(myJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadApierCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadErsCfgError(t *testing.T) {
	cfgJSONStr := `{
"ers": {										
	"sessions_conns": "*internal",
},
}`
	expected := "json: cannot unmarshal string into Go struct field ERsJsonCfg.Sessions_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadErsCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadEesCfgError(t *testing.T) {
	cfgJSONStr := `{
      "ees": { 
            "attributes_conns": "*conn1",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field EEsJsonCfg.Attributes_conns of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadEesCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCoreSCfgError(t *testing.T) {
	cfgJSONStr := `{
      "cores": { 
            "caps": "1",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field CoreSJsonCfg.Caps of type int"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadCoreSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRateSCfgError(t *testing.T) {
	cfgJSONStr := `{
     "rates": {
	         "string_indexed_fields": "*req.index",
     },
}`
	expected := "json: cannot unmarshal string into Go struct field RateSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRateSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSIPAgentCfgError(t *testing.T) {
	cfgJSONStr := `{
	"sip_agent": {
		"request_processors": [
             {
               "id": 1234,
             },
		],
	},
}`
	expected := "json: cannot unmarshal number into Go struct field ReqProcessorJsnCfg.Request_processors.ID of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSIPAgentCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadTemplateSCfgError(t *testing.T) {
	cfgJSONStr := `{
     "templates": {
           "custom_template": [
              {
                "tag": 1234,
            },
           ],
     }
}`
	expected := "json: cannot unmarshal number into Go struct field FcTemplateJsonCfg.Tag of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadTemplateSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	cfgJSONStr = `{
     "templates": {
           "custom_template": [
              {
                "value": "a{*",
            },
           ],
     }
}`
	expected = "invalid converter terminator in rule: <a{*>"
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadTemplateSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadConfigsCfgError(t *testing.T) {
	cfgJSONStr := `{
      "configs": {
          "url": 123,
      },
}`
	expected := "json: cannot unmarshal number into Go struct field ConfigSCfgJson.Url of type string"
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadConfigSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSuretaxConfig(t *testing.T) {
	tLocal, err := time.LoadLocation("Local")
	if err != nil {
		t.Error(err)
	}
	expected := &SureTaxCfg{
		URL:                  "",
		ClientNumber:         "",
		ValidationKey:        "",
		BusinessUnit:         "",
		Timezone:             tLocal,
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", utils.InfieldSep),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", utils.InfieldSep),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep),
		BillToNumber:         NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Zipcode:              NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Plus4:                NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		P2PZipcode:           NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		P2PPlus4:             NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
		Units:                NewRSRParsersMustCompile("1", utils.InfieldSep),
		UnitType:             NewRSRParsersMustCompile("00", utils.InfieldSep),
		TaxIncluded:          NewRSRParsersMustCompile("0", utils.InfieldSep),
		TaxSitusRule:         NewRSRParsersMustCompile("04", utils.InfieldSep),
		TransTypeCode:        NewRSRParsersMustCompile("010101", utils.InfieldSep),
		SalesTypeCode:        NewRSRParsersMustCompile("R", utils.InfieldSep),
		TaxExemptionCodeList: NewRSRParsersMustCompile(utils.EmptyString, utils.InfieldSep),
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.SureTaxCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestDiameterAgentConfig(t *testing.T) {
	expected := &DiameterAgentCfg{
		Enabled:           false,
		ListenNet:         "tcp",
		Listen:            "127.0.0.1:3868",
		DictionariesPath:  "/usr/share/cgrates/diameter/dict/",
		SessionSConns:     []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorID:          0,
		ProductName:       "CGRateS",
		ConcurrentReqs:    -1,
		SyncedConnReqs:    false,
		ASRTemplate:       "",
		RARTemplate:       "",
		ForcedDisconnect:  "*none",
		RequestProcessors: nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.DiameterAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRadiusAgentConfig(t *testing.T) {
	expected := &RadiusAgentCfg{
		Enabled:            false,
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.MetaDefault: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.MetaDefault: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		RequestProcessors:  nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.RadiusAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestDNSAgentConfig(t *testing.T) {
	expected := &DNSAgentCfg{
		Enabled:           false,
		Listen:            "127.0.0.1:2053",
		ListenNet:         "udp",
		SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:          "",
		RequestProcessors: nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.DNSAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAttributeSConfig(t *testing.T) {
	expected := &AttributeSCfg{
		Enabled:             false,
		AdminSConns:         []string{},
		StatSConns:          []string{},
		ResourceSConns:      []string{},
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		ProcessRuns:         1,
		NestedFields:        false,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.AttributeSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestChargersConfig(t *testing.T) {
	expected := &ChargerSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		AttributeSConns:     []string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.ChargerSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestResourceSConfig(t *testing.T) {
	expected := &ResourceSConfig{
		Enabled:             false,
		IndexedSelects:      true,
		StoreInterval:       0,
		ThresholdSConns:     []string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.ResourceSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestStatSConfig(t *testing.T) {
	expected := &StatSCfg{
		Enabled:                false,
		IndexedSelects:         true,
		StoreInterval:          0,
		StoreUncompressedLimit: 0,
		ThresholdSConns:        []string{},
		PrefixIndexedFields:    &[]string{},
		SuffixIndexedFields:    &[]string{},
		NestedFields:           false,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.StatSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestThresholdSConfig(t *testing.T) {
	expected := &ThresholdSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StoreInterval:       0,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        false,
		ActionSConns:        []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.ThresholdSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRouteSConfig(t *testing.T) {
	expected := &RouteSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		DefaultRatio:        1,
		NestedFields:        false,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.RouteSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestSessionSConfig(t *testing.T) {
	expected := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
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
		SessionIndexes:      utils.StringSet{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.StringSet{},
		ActionSConns:        []string{},
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.StringSet{utils.MetaAny: {}},
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
			PrivateKeyPath:     "",
			PublicKeyPath:      "",
		},
		DefaultUsage: map[string]time.Duration{
			utils.MetaAny:   3 * time.Hour,
			utils.MetaVoice: 3 * time.Hour,
			utils.MetaData:  1048576,
			utils.MetaSMS:   1,
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.SessionSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestFsAgentConfig(t *testing.T) {
	expected := &FsAgentCfg{
		Enabled:             false,
		SessionSConns:       []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		SubscribePark:       true,
		CreateCdr:           false,
		LowBalanceAnnFile:   "",
		EmptyBalanceAnnFile: "",
		EmptyBalanceContext: "",
		MaxWaitConnection:   2000000000,
		ExtraFields:         RSRParsers{},
		EventSocketConns: []*FsConnCfg{
			{
				Address:    "127.0.0.1:8021",
				Password:   "ClueCon",
				Reconnects: 5,
				Alias:      "127.0.0.1:8021",
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.FsAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestKamAgentConfig(t *testing.T) {
	expected := &KamAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCdr:     false,
		EvapiConns:    []*KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 5, Alias: ""}},
		Timezone:      "",
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.KamAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAsteriskAgentConfig(t *testing.T) {
	expected := &AsteriskAgentCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS)},
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{{
			Alias:           "",
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 3,
			Reconnects:      5,
		}},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.AsteriskAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestFilterSConfig(t *testing.T) {
	expected := &FilterSCfg{
		StatSConns:     []string{},
		ResourceSConns: []string{},
		AdminSConns:    []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.FilterSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestLoaderConfig(t *testing.T) {
	ten, err := NewRSRParsers("", utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	expected := LoaderSCfgs{
		{
			Enabled:        false,
			ID:             utils.MetaDefault,
			Tenant:         ten,
			LockFileName:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data:           nil,
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.LoaderCfg()
	newConfig[0].Data = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestDispatcherSConfig(t *testing.T) {
	expected := &DispatcherSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		NestedFields:        false,
		AnySubsystem:        true,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.DispatcherSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAnalyzerConfig(t *testing.T) {
	expected := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		TTL:             24 * time.Hour,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.AnalyzerSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestApierConfig(t *testing.T) {
	expected := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		ActionSConns:    []string{},
		AttributeSConns: []string{},
		EEsConns:        []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.AdminSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestERSConfig(t *testing.T) {
	expected := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{"*internal:*sessions"},
		Readers: []*EventReaderCfg{
			{
				ID:                  utils.MetaDefault,
				Type:                utils.MetaNone,
				RunDelay:            0,
				ConcurrentReqs:      1024,
				SourcePath:          "/var/spool/cgrates/ers/in",
				ProcessedPath:       "/var/spool/cgrates/ers/out",
				Tenant:              nil,
				Timezone:            utils.EmptyString,
				Filters:             []string{},
				Flags:               utils.FlagsWithParams{},
				Fields:              nil,
				CacheDumpFields:     make([]*FCTemplate, 0),
				PartialCommitFields: make([]*FCTemplate, 0),
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
				},
			},
		},
		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.ERsCfg()
	newConfig.Readers[0].Fields = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestEEsNoLksConfig(t *testing.T) {
	expected := &EEsCfg{
		Enabled:         false,
		AttributeSConns: []string{},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -1,
				TTL:       5 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				Tenant:        nil,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				contentFields: []*FCTemplate{},
				Fields:        []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.EEsNoLksCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRateSConfig(t *testing.T) {
	expected := &RateSCfg{
		Enabled:                 false,
		IndexedSelects:          true,
		PrefixIndexedFields:     &[]string{},
		SuffixIndexedFields:     &[]string{},
		NestedFields:            false,
		RateIndexedSelects:      true,
		RatePrefixIndexedFields: &[]string{},
		RateSuffixIndexedFields: &[]string{},
		RateNestedFields:        false,
		Verbosity:               1000,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.RateSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestSIPAgentConfig(t *testing.T) {
	expected := &SIPAgentCfg{
		Enabled:             false,
		Listen:              "127.0.0.1:5060",
		ListenNet:           "udp",
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Timezone:            "",
		RetransmissionTimer: 1000000000,
		RequestProcessors:   nil,
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.SIPAgentCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRPCConnsConfig(t *testing.T) {
	expected := RPCConns{
		utils.MetaBiJSONLocalHost: {
			Strategy: rpcclient.PoolFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{{
				Address:   "127.0.0.1:2014",
				Transport: rpcclient.BiRPCJSON,
			}},
		},
		utils.MetaInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     utils.MetaInternal,
					Transport:   utils.EmptyString,
					Synchronous: false,
					TLS:         false,
				},
			},
		},
		rpcclient.BiRPCInternal: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     rpcclient.BiRPCInternal,
					Transport:   utils.EmptyString,
					Synchronous: false,
					TLS:         false,
				},
			},
		},
		utils.MetaLocalHost: {
			Strategy: utils.MetaFirst,
			PoolSize: 0,
			Conns: []*RemoteHost{
				{
					Address:     "127.0.0.1:2012",
					Transport:   "*json",
					Synchronous: false,
					TLS:         false,
				},
			},
		},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.RPCConns()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestTemplatesConfig(t *testing.T) {
	expected := FCTemplates{
		"*err": {
			{
				Tag:       "SessionId",
				Type:      "*variable",
				Path:      "*rep.Session-Id",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     NewRSRParsersMustCompile("~*req.Session-Id", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "OriginHost",
				Type:      "*variable",
				Path:      "*rep.Origin-Host",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     NewRSRParsersMustCompile("~*vars.OriginHost", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "OriginRealm",
				Type:      "*variable",
				Path:      "*rep.Origin-Realm",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     NewRSRParsersMustCompile("~*vars.OriginRealm", utils.InfieldSep),
				Mandatory: true,
			},
		},
		"*errSip": {
			{
				Tag:       "Request",
				Type:      "*constant",
				Path:      "*rep.Request",
				Layout:    "2006-01-02T15:04:05Z07:00",
				Value:     NewRSRParsersMustCompile("SIP/2.0 500 Internal Server Error", utils.InfieldSep),
				Mandatory: true,
			},
		},
		"*cca":           nil,
		"*asr":           nil,
		"*rar":           nil,
		utils.MetaCdrLog: nil,
	}
	for _, value := range expected {
		for _, elem := range value {
			elem.ComputePath()
		}
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.TemplatesCfg()
	newConfig["*cca"] = nil
	newConfig["*asr"] = nil
	newConfig["*rar"] = nil
	newConfig[utils.MetaCdrLog] = nil
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestConfigsConfig(t *testing.T) {
	expected := &ConfigSCfg{
		Enabled: false,
		URL:     "/configs/",
		RootDir: "/var/spool/cgrates/configs",
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.ConfigSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestAPIBanConfig(t *testing.T) {
	expected := &APIBanCfg{
		Enabled: false,
		Keys:    []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig := cgrConfig.APIBanCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestRLockSections(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cgrCfg.rLockSections()
	cgrCfg.rUnlockSections()
}

func TestLockSections(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cgrCfg.lockSections()
	cgrCfg.unlockSections()
}

func TestRLockAndRUnlock(t *testing.T) {
	cgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cgrCfg.RLocks("attributes", "ees", "general")
	cgrCfg.RUnlocks("attributes", "ees", "general")
}

func TestCgrLoaderCfgITDefaults(t *testing.T) {
	eCfg := LoaderSCfgs{
		{
			ID:             utils.MetaDefault,
			Enabled:        false,
			DryRun:         false,
			RunDelay:       0,
			LockFileName:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Data: []*LoaderDataType{
				{
					Type:     utils.MetaAttributes,
					Filename: utils.AttributesCsv,
					Fields: []*FCTemplate{
						{Tag: "TenantID",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:       "ProfileID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:    "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{
							Tag:    "AttributeFilterIDs",
							Path:   "AttributeFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{
							Tag:    "Path",
							Path:   "Path",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Type",
							Path:   "Type",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Value",
							Path:   "Value",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaFilters,
					Filename: utils.FiltersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Type",
							Path:   "Type",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Element",
							Path:   "Element",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Values",
							Path:   "Values",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaResources,
					Filename: utils.ResourcesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "UsageTTL",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Limit",
							Path:   "Limit",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AllocationMessage",
							Path:   "AllocationMessage",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaStats,
					Filename: utils.StatsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:   "MetricIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricFilterIDs",
							Path:   "MetricFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},

						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaThresholds,
					Filename: utils.ThresholdsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MaxHits",
							Path:   "MaxHits",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinHits",
							Path:   "MinHits",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinSleep",
							Path:   "MinSleep",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionProfileIDs",
							Path:   "ActionProfileIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Async",
							Path:   "Async",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRoutes,
					Filename: utils.RoutesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteID",
							Path:   "RouteID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteFilterIDs",
							Path:   "RouteFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteAccountIDs",
							Path:   "RouteAccountIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteRatingPlanIDs",
							Path:   "RouteRatingPlanIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteResourceIDs",
							Path:   "RouteResourceIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteStatIDs",
							Path:   "RouteStatIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteWeight",
							Path:   "RouteWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteBlocker",
							Path:   "RouteBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteParameters",
							Path:   "RouteParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaChargers,
					Filename: utils.ChargersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RunID",
							Path:   "RunID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AttributeIDs",
							Path:   "AttributeIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaDispatchers,
					Filename: utils.DispatcherProfilesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Contexts",
							Path:   "Contexts",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Strategy",
							Path:   "Strategy",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "StrategyParameters",
							Path:   "StrategyParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "ConnID",
							Path:   "ConnID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "ConnFilterIDs",
							Path:   "ConnFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "ConnWeight",
							Path:   "ConnWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "ConnBlocker",
							Path:   "ConnBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "ConnParameters",
							Path:   "ConnParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaDispatcherHosts,
					Filename: utils.DispatcherHostsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Address",
							Path:   "Address",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Transport",
							Path:   "Transport",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "TLS",
							Path:   "TLS",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaRateProfiles,
					Filename: utils.RateProfilesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MinCost",
							Path:   "MinCost",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCost",
							Path:   "MaxCost",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCostStrategy",
							Path:   "MaxCostStrategy",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateID",
							Path:   "RateID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateFilterIDs",
							Path:   "RateFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateActivationTimes",
							Path:   "RateActivationTimes",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateWeight",
							Path:   "RateWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateBlocker",
							Path:   "RateBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateIntervalStart",
							Path:   "RateIntervalStart",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateFixedFee",
							Path:   "RateFixedFee",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateRecurrentFee",
							Path:   "RateRecurrentFee",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateUnit",
							Path:   "RateUnit",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateIncrement",
							Path:   "RateIncrement",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout: time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaActionProfiles,
					Filename: utils.ActionProfilesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TargetType",
							Path:   "TargetType",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TargetIDs",
							Path:   "TargetIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionID",
							Path:   "ActionID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionFilterIDs",
							Path:   "ActionFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionBlocker",
							Path:   "ActionBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionTTL",
							Path:   "ActionTTL",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionType",
							Path:   "ActionType",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionOpts",
							Path:   "ActionOpts",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionPath",
							Path:   "ActionPath",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionValue",
							Path:   "ActionValue",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaAccounts,
					Filename: utils.AccountsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceID",
							Path:   "BalanceID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceFilterIDs",
							Path:   "BalanceFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceWeight",
							Path:   "BalanceWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceBlocker",
							Path:   "BalanceBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceType",
							Path:   "BalanceType",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceOpts",
							Path:   "BalanceOpts",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceCostIncrements",
							Path:   "BalanceCostIncrements",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceAttributeIDs",
							Path:   "BalanceAttributeIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceRateProfileIDs",
							Path:   "BalanceRateProfileIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceUnitFactors",
							Path:   "BalanceUnitFactors",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceUnits",
							Path:   "BalanceUnits",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
			},
		},
	}
	for _, profile := range eCfg {
		for _, fields := range profile.Data {
			for _, v := range fields.Fields {
				v.ComputePath()
			}
		}
	}
	if !reflect.DeepEqual(eCfg, cgrCfg.loaderCfg) {
		t.Errorf("received: %+v, \n expecting: %+v",
			utils.ToJSON(eCfg), utils.ToJSON(cgrCfg.loaderCfg))
	}
}

func TestCgrCfgJSONDefaultDispatcherSCfg(t *testing.T) {
	eDspSCfg := &DispatcherSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		AnySubsystem:        true,
	}
	if !reflect.DeepEqual(cgrCfg.dispatcherSCfg, eDspSCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.dispatcherSCfg, eDspSCfg)
	}
}

func TestCgrLoaderCfgDefault(t *testing.T) {
	eLdrCfg := &LoaderCgrCfg{
		TpID:            "",
		DataPath:        "./",
		DisableReverse:  false,
		FieldSeparator:  utils.CSVSep,
		CachesConns:     []string{utils.MetaLocalHost},
		ActionSConns:    []string{utils.MetaLocalHost},
		GapiCredentials: json.RawMessage(`".gapi/credentials.json"`),
		GapiToken:       json.RawMessage(`".gapi/token.json"`),
	}
	if !reflect.DeepEqual(cgrCfg.LoaderCgrCfg(), eLdrCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.LoaderCgrCfg()), utils.ToJSON(eLdrCfg))
	}
}

func TestLoadConfigDBCfgErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()

	dbJsonCfg := func(section string) (*DbJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	jsnCfg := mockDb{
		DbJsonCfgF: dbJsonCfg,
	}
	if err := cfg.loadConfigDBCfg(&jsnCfg); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestGetReloadChan(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	sectID := "test"
	rcv := cfg.GetReloadChan(sectID)
	if rcv != nil {
		t.Errorf("Expected %v \b but received \n %v", nil, rcv)
	}

	cfg.initChanels()
	sectID = GeneralJSON
	rcv = cfg.GetReloadChan(sectID)
	expected := make(chan struct{})
	if len(rcv) != len(expected) {
		t.Error("Channels should have the same length")
	}
}

func TestCgrMigratorCfgDefault(t *testing.T) {
	eMgrCfg := &MigratorCgrCfg{
		OutDataDBType:     "redis",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     "cgrates",
		OutDataDBPassword: "",
		OutDataDBEncoding: "msgpack",
		OutStorDBType:     "mysql",
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     "cgrates",
		OutStorDBUser:     "cgrates",
		OutStorDBPassword: "",
		OutDataDBOpts: map[string]interface{}{
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		OutStorDBOpts: make(map[string]interface{}),
	}
	if !reflect.DeepEqual(cgrCfg.MigratorCgrCfg(), eMgrCfg) {
		t.Errorf("expected: %+v, received: %+v", utils.ToJSON(eMgrCfg), utils.ToJSON(cgrCfg.MigratorCgrCfg()))
	}
}

func TestCgrMigratorCfg2(t *testing.T) {
	jsnCfg := `
{
"migrator": {
	"out_datadb_type": "redis",
	"out_datadb_host": "0.0.0.0",
	"out_datadb_port": "9999",
	"out_datadb_name": "9999",
	"out_datadb_user": "cgrates",
	"out_datadb_password": "",
	"out_datadb_encoding" : "msgpack",
	"out_stordb_type": "mysql",
	"out_stordb_host": "0.0.0.0",
	"out_stordb_port": "9999",
	"out_stordb_name": "cgrates",
	"out_stordb_user": "cgrates",
	"out_stordb_password": "",
},
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.MigratorCgrCfg().OutDataDBHost != "0.0.0.0" {
		t.Errorf("Expected: 0.0.0.0 , received: %+v", cgrCfg.MigratorCgrCfg().OutDataDBHost)
	} else if cgrCfg.MigratorCgrCfg().OutDataDBPort != "9999" {
		t.Errorf("Expected: 9999, received: %+v", cgrCfg.MigratorCgrCfg().OutDataDBPassword)
	}
}

func TestCfgTlsCfg(t *testing.T) {
	jsnCfg := `
	{
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
	eCgrCfg := NewDefaultCGRConfig()
	eCgrCfg.tlsCfg = &TLSCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.TLSCfg(), cgrCfg.TLSCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.tlsCfg), utils.ToJSON(cgrCfg.tlsCfg))
	}
}

func TestCgrCfgJSONDefaultAnalyzerSCfg(t *testing.T) {
	aSCfg := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		TTL:             24 * time.Hour,
	}
	if !reflect.DeepEqual(cgrCfg.analyzerSCfg, aSCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.analyzerSCfg, aSCfg)
	}
}

func TestNewCGRConfigFromPathNotFound(t *testing.T) {
	fpath := path.Join("/usr", "share", "cgrates", "conf", "samples", "notValid")
	_, err := NewCGRConfigFromPath(fpath)
	if err == nil || err.Error() != utils.ErrPathNotReachable(fpath).Error() {
		t.Fatalf("Expected %s ,received %s", utils.ErrPathNotReachable(fpath), err)
	}
	fpath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo", "cgrates.json")
	cfg, err := NewCGRConfigFromPath(fpath)
	if err == nil {
		t.Fatalf("Expected error,received %v", cfg)
	}
	fpath = "https://not_a_reacheble_website"
	_, err = NewCGRConfigFromPath(fpath)
	if err == nil || err.Error() != utils.ErrPathNotReachable(fpath).Error() {
		t.Fatalf("Expected %s ,received %s", utils.ErrPathNotReachable(fpath), err)
	}
	cfg, err = NewCGRConfigFromPath("https://github.com/")
	if err == nil {
		t.Fatalf("Expected error,received %v", cfg)
	}
}

func TestCgrCfgJSONDefaultApierCfg(t *testing.T) {
	aCfg := &AdminSCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		ActionSConns:    []string{},
		AttributeSConns: []string{},
		EEsConns:        []string{},
	}
	if !reflect.DeepEqual(cgrCfg.admS, aCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.admS, aCfg)
	}
}

func TestCgrCfgJSONDefaultRateCfg(t *testing.T) {
	eCfg := &RateSCfg{
		Enabled:                 false,
		IndexedSelects:          true,
		StringIndexedFields:     nil,
		PrefixIndexedFields:     &[]string{},
		SuffixIndexedFields:     &[]string{},
		NestedFields:            false,
		RateIndexedSelects:      true,
		RateStringIndexedFields: nil,
		RatePrefixIndexedFields: &[]string{},
		RateSuffixIndexedFields: &[]string{},
		RateNestedFields:        false,
		Verbosity:               1000,
	}
	if !reflect.DeepEqual(cgrCfg.rateSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.rateSCfg, eCfg)
	}
}

func TestCgrCfgV1GetConfigAllConfig(t *testing.T) {
	var rcv map[string]interface{}
	cgrCfg := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	expected := cgrCfg.AsMapInterface(cgrCfg.GeneralCfg().RSRSep)
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{}, &rcv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{}, &rcv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCgrCfgV1GetConfigSectionLoader(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		LoaderSJSON: []map[string]interface{}{
			{
				utils.IDCfg:           "*default",
				utils.EnabledCfg:      false,
				utils.TenantCfg:       utils.EmptyString,
				utils.DryRunCfg:       false,
				utils.RunDelayCfg:     "0",
				utils.LockFileNameCfg: ".cgr.lck",
				utils.CachesConnsCfg:  []string{utils.MetaInternal},
				utils.FieldSepCfg:     ",",
				utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
				utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
				utils.DataCfg:         []map[string]interface{}{},
			},
		},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[LoaderSJSON].([]map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[LoaderSJSON])
	} else {
		mp[0][utils.DataCfg] = []map[string]interface{}{}
		if !reflect.DeepEqual(expected[LoaderSJSON], mp) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected[LoaderSJSON]), utils.ToJSON(mp))
		}
	}
}

func TestCgrCfgV1GetConfigSectionHTTPAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		HTTPAgentJSON: []map[string]interface{}{},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestCgrCfgV1GetConfigSectionCoreS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		CoreSJSON: map[string]interface{}{
			utils.CapsCfg:              0,
			utils.CapsStrategyCfg:      utils.MetaBusy,
			utils.CapsStatsIntervalCfg: "0",
			utils.ShutdownTimeoutCfg:   "1s",
		},
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CoreSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestCgrCfgV1GetConfigListen(t *testing.T) {
	jsnCfg := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`
	expected := map[string]interface{}{
		"listen": map[string]interface{}{
			"http":         ":2080",
			"http_tls":     "127.0.0.1:2280",
			"rpc_gob":      ":2013",
			"rpc_gob_tls":  "127.0.0.1:2023",
			"rpc_json":     ":2012",
			"rpc_json_tls": "127.0.0.1:2022",
		},
	}
	var rcv map[string]interface{}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if err := cgrCfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ListenJSON}}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestV1GetConfigGeneral(t *testing.T) {
	var reply map[string]interface{}
	cfgJSONStr := `{
      "general": {
            "node_id": "ENGINE1",
            "locking_timeout": "0",
            "failed_posts_ttl": "0s",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
        }
}`
	expected := map[string]interface{}{
		utils.NodeIDCfg:           "ENGINE1",
		utils.LoggerCfg:           "*syslog",
		utils.LogLevelCfg:         6,
		utils.RoundingDecimalsCfg: 5,
		utils.DBDataEncodingCfg:   "*msgpack",
		utils.TpExportPathCfg:     "/var/spool/cgrates/tpe",
		utils.PosterAttemptsCfg:   3,
		utils.FailedPostsDirCfg:   "/var/spool/cgrates/failed_posts",
		utils.FailedPostsTTLCfg:   "0",
		utils.DefaultReqTypeCfg:   "*rated",
		utils.DefaultCategoryCfg:  "call",
		utils.DefaultTenantCfg:    "cgrates.org",
		utils.DefaultTimezoneCfg:  "Local",
		utils.DefaultCachingCfg:   "*reload",
		utils.ConnectAttemptsCfg:  5,
		utils.ReconnectsCfg:       -1,
		utils.ConnectTimeoutCfg:   "0",
		utils.ReplyTimeoutCfg:     "0",
		utils.LockingTimeoutCfg:   "0",
		utils.DigestSeparatorCfg:  ",",
		utils.DigestEqualCfg:      ":",
		utils.RSRSepCfg:           ";",
		utils.MaxParallelConnsCfg: 100,
	}
	expected = map[string]interface{}{
		GeneralJSON: expected,
	}
	cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr)
	if err != nil {
		t.Error(err)
	}
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{GeneralJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDataDB(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		utils.DataDbTypeCfg:          "*redis",
		utils.DataDbHostCfg:          "127.0.0.1",
		utils.DataDbPortCfg:          6379,
		utils.DataDbNameCfg:          "10",
		utils.DataDbUserCfg:          "cgrates",
		utils.DataDbPassCfg:          "",
		utils.ReplicationFilteredCfg: false,
		utils.RemoteConnIDCfg:        "",
		utils.ReplicationCache:       "",
		utils.OptsCfg:                map[string]interface{}{},
		utils.RemoteConnsCfg:         []string{},
		utils.ReplicationConnsCfg:    []string{},
		utils.ItemsCfg:               map[string]interface{}{},
	}
	expected = map[string]interface{}{
		DataDBJSON: expected,
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DataDBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[DataDBJSON].(map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[DataDBJSON])
	} else {
		mp[utils.ItemsCfg] = map[string]interface{}{}
		mp[utils.OptsCfg] = map[string]interface{}{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigStorDB(t *testing.T) {
	var reply map[string]interface{}
	var empty []string
	expected := map[string]interface{}{
		utils.DataDbTypeCfg:          "*mysql",
		utils.DataDbHostCfg:          "127.0.0.1",
		utils.DataDbPortCfg:          3306,
		utils.DataDbNameCfg:          "cgrates",
		utils.DataDbUserCfg:          "cgrates",
		utils.DataDbPassCfg:          "",
		utils.StringIndexedFieldsCfg: []string{},
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.RemoteConnsCfg:         empty,
		utils.ReplicationConnsCfg:    empty,
		utils.OptsCfg: map[string]interface{}{
			utils.SQLMaxOpenConnsCfg:    100.,
			utils.SQLMaxIdleConnsCfg:    10.,
			utils.SQLConnMaxLifetimeCfg: 0.,
			utils.MongoQueryTimeoutCfg:  "10s",
			utils.SSLModeCfg:            "disable",
			utils.MysqlLocation:         "Local",
		},
		utils.ItemsCfg: map[string]interface{}{},
	}
	expected = map[string]interface{}{
		StorDBJSON: expected,
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{StorDBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[StorDBJSON].(map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[StorDBJSON])
	} else {
		mp[utils.ItemsCfg] = map[string]interface{}{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigTLS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		TlsJSON: map[string]interface{}{
			utils.ServerCerificateCfg: "",
			utils.ServerKeyCfg:        "",
			utils.ServerPolicyCfg:     4,
			utils.ServerNameCfg:       "",
			utils.ClientCerificateCfg: "",
			utils.ClientKeyCfg:        "",
			utils.CaCertificateCfg:    "",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{TlsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigCache(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		CacheJSON: map[string]interface{}{
			utils.PartitionsCfg:       map[string]interface{}{},
			utils.ReplicationConnsCfg: []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CacheJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[CacheJSON].(map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[CacheJSON])
	} else {
		mp[utils.PartitionsCfg] = map[string]interface{}{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigHTTP(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		HTTPJSON: map[string]interface{}{
			utils.HTTPJsonRPCURLCfg:        "/jsonrpc",
			utils.RegistrarSURLCfg:         "/registrar",
			utils.HTTPWSURLCfg:             "/ws",
			utils.HTTPFreeswitchCDRsURLCfg: "/freeswitch_json",
			utils.HTTPCDRsURLCfg:           "/cdr_http",
			utils.HTTPUseBasicAuthCfg:      false,
			utils.HTTPAuthUsersCfg:         map[string]string{},
			utils.HTTPClientOptsCfg: map[string]interface{}{
				utils.HTTPClientTLSClientConfigCfg:       false,
				utils.HTTPClientTLSHandshakeTimeoutCfg:   "10s",
				utils.HTTPClientDisableKeepAlivesCfg:     false,
				utils.HTTPClientDisableCompressionCfg:    false,
				utils.HTTPClientMaxIdleConnsCfg:          100.,
				utils.HTTPClientMaxIdleConnsPerHostCfg:   2.,
				utils.HTTPClientMaxConnsPerHostCfg:       0.,
				utils.HTTPClientIdleConnTimeoutCfg:       "90s",
				utils.HTTPClientResponseHeaderTimeoutCfg: "0",
				utils.HTTPClientExpectContinueTimeoutCfg: "0",
				utils.HTTPClientForceAttemptHTTP2Cfg:     true,
				utils.HTTPClientDialTimeoutCfg:           "30s",
				utils.HTTPClientDialFallbackDelayCfg:     "300ms",
				utils.HTTPClientDialKeepAliveCfg:         "30s",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigFilterS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		FilterSJSON: map[string]interface{}{
			utils.StatSConnsCfg:     []string{},
			utils.ResourceSConnsCfg: []string{},
			utils.AdminSConnsCfg:    []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{FilterSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigCdrs(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		CDRsJSON: map[string]interface{}{
			utils.EnabledCfg:          false,
			utils.ExtraFieldsCfg:      []string{},
			utils.StoreCdrsCfg:        true,
			utils.SessionCostRetires:  5,
			utils.ChargerSConnsCfg:    []string{},
			utils.AttributeSConnsCfg:  []string{},
			utils.ThresholdSConnsCfg:  []string{},
			utils.StatSConnsCfg:       []string{},
			utils.OnlineCDRExportsCfg: []string{},
			utils.ActionSConnsCfg:     []string{},
			utils.EEsConnsCfg:         []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{CDRsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSessionS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		SessionSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.ListenBijsonCfg:        "127.0.0.1:2014",
			utils.ListenBigobCfg:         "",
			utils.ChargerSConnsCfg:       []string{},
			utils.CDRsConnsCfg:           []string{},
			utils.ResourceSConnsCfg:      []string{},
			utils.ThresholdSConnsCfg:     []string{},
			utils.StatSConnsCfg:          []string{},
			utils.RouteSConnsCfg:         []string{},
			utils.AttributeSConnsCfg:     []string{},
			utils.ActionSConnsCfg:        []string{},
			utils.ReplicationConnsCfg:    []string{},
			utils.DebitIntervalCfg:       "0",
			utils.StoreSCostsCfg:         false,
			utils.SessionIndexesCfg:      []string{},
			utils.ClientProtocolCfg:      1.0,
			utils.SessionTTLCfg:          "0",
			utils.ChannelSyncIntervalCfg: "0",
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
			utils.DefaultUsageCfg: map[string]string{
				utils.MetaAny:   "3h0m0s",
				utils.MetaVoice: "3h0m0s",
				utils.MetaData:  "1048576",
				utils.MetaSMS:   "1",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SessionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigFsAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		FreeSWITCHAgentJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.SessionSConnsCfg:       []string{rpcclient.BiRPCInternal},
			utils.SubscribeParkCfg:       true,
			utils.CreateCdrCfg:           false,
			utils.ExtraFieldsCfg:         []string{},
			utils.LowBalanceAnnFileCfg:   "",
			utils.EmptyBalanceContextCfg: "",
			utils.EmptyBalanceAnnFileCfg: "",
			utils.MaxWaitConnectionCfg:   "2s",
			utils.EventSocketConnsCfg: []map[string]interface{}{
				{
					utils.AddressCfg:    "127.0.0.1:8021",
					utils.Password:      "ClueCon",
					utils.ReconnectsCfg: 5,
					utils.AliasCfg:      "127.0.0.1:8021"},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{FreeSWITCHAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigKamailioAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		KamailioAgentJSON: map[string]interface{}{
			utils.EnabledCfg:       false,
			utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal},
			utils.CreateCdrCfg:     false,
			utils.TimezoneCfg:      "",
			utils.EvapiConnsCfg: []map[string]interface{}{
				{
					utils.AddressCfg:    "127.0.0.1:8448",
					utils.ReconnectsCfg: 5,
					utils.AliasCfg:      "",
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{KamailioAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAsteriskAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		AsteriskAgentJSON: map[string]interface{}{
			utils.EnabledCfg:       false,
			utils.SessionSConnsCfg: []string{rpcclient.BiRPCInternal},
			utils.CreateCdrCfg:     false,
			utils.AsteriskConnsCfg: []map[string]interface{}{
				{
					utils.AliasCfg:           "",
					utils.AddressCfg:         "127.0.0.1:8088",
					utils.UserCf:             "cgrates",
					utils.Password:           "CGRateS.org",
					utils.ConnectAttemptsCfg: 3,
					utils.ReconnectsCfg:      5,
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AsteriskAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDiameterAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		DiameterAgentJSON: map[string]interface{}{
			utils.ASRTemplateCfg:        "",
			utils.ConcurrentRequestsCfg: -1,
			utils.DictionariesPathCfg:   "/usr/share/cgrates/diameter/dict/",
			utils.EnabledCfg:            false,
			utils.ForcedDisconnectCfg:   "*none",
			utils.ListenCfg:             "127.0.0.1:3868",
			utils.ListenNetCfg:          "tcp",
			utils.OriginHostCfg:         "CGR-DA",
			utils.OriginRealmCfg:        "cgrates.org",
			utils.ProductNameCfg:        "CGRateS",
			utils.RARTemplateCfg:        "",
			utils.SessionSConnsCfg:      []string{rpcclient.BiRPCInternal},
			utils.SyncedConnReqsCfg:     false,
			utils.VendorIDCfg:           0,
			utils.RequestProcessorsCfg:  []map[string]interface{}{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DiameterAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigRadiusAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		RadiusAgentJSON: map[string]interface{}{
			utils.EnabledCfg:    false,
			utils.ListenNetCfg:  "udp",
			utils.ListenAuthCfg: "127.0.0.1:1812",
			utils.ListenAcctCfg: "127.0.0.1:1813",
			utils.ClientSecretsCfg: map[string]string{
				utils.MetaDefault: "CGRateS.org",
			},
			utils.ClientDictionariesCfg: map[string]string{
				utils.MetaDefault: "/usr/share/cgrates/radius/dict/",
			},
			utils.SessionSConnsCfg:     []string{"*internal"},
			utils.RequestProcessorsCfg: []map[string]interface{}{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RadiusAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDNSAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		DNSAgentJSON: map[string]interface{}{
			utils.EnabledCfg:           false,
			utils.ListenCfg:            "127.0.0.1:2053",
			utils.ListenNetCfg:         "udp",
			utils.SessionSConnsCfg:     []string{utils.MetaInternal},
			utils.TimezoneCfg:          "",
			utils.RequestProcessorsCfg: []map[string]interface{}{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DNSAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAttribute(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		AttributeSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.StatSConnsCfg:          []string{},
			utils.ResourceSConnsCfg:      []string{},
			utils.AdminSConnsCfg:         []string{},
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.ProcessRunsCfg:         1,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AttributeSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigChargers(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ChargerSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.AttributeSConnsCfg:     []string{},
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.SuffixIndexedFieldsCfg: []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ChargerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigResourceS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ResourceSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.StoreIntervalCfg:       utils.EmptyString,
			utils.ThresholdSConnsCfg:     []string{},
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ResourceSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigStats(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		StatSJSON: map[string]interface{}{
			utils.EnabledCfg:                false,
			utils.StoreIntervalCfg:          utils.EmptyString,
			utils.StoreUncompressedLimitCfg: 0,
			utils.ThresholdSConnsCfg:        []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.NestedFieldsCfg:           false,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{StatSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigThresholds(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ThresholdSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.StoreIntervalCfg:       utils.EmptyString,
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.ActionSConnsCfg:        []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ThresholdSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigAcounts(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		AccountSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.IndexedSelectsCfg:      true,
			utils.AttributeSConnsCfg:     []string{},
			utils.RateSConnsCfg:          []string{},
			utils.ThresholdSConnsCfg:     []string{},
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.MaxIterations:          1000,
			utils.MaxUsage:               "259200000000000", // 72h in ns
		},
	}
	cfg := NewDefaultCGRConfig()
	if err := cfg.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AccountSJSON}}, &reply); err != nil {
		t.Error(expected)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigRoutes(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		RouteSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.AttributeSConnsCfg:     []string{},
			utils.ResourceSConnsCfg:      []string{},
			utils.StatSConnsCfg:          []string{},
			utils.DefaultRatioCfg:        1,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RouteSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSuretax(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		SureTaxJSON: map[string]interface{}{
			utils.URLCfg:                  utils.EmptyString,
			utils.ClientNumberCfg:         utils.EmptyString,
			utils.ValidationKeyCfg:        utils.EmptyString,
			utils.BusinessUnitCfg:         utils.EmptyString,
			utils.TimezoneCfg:             "UTC",
			utils.IncludeLocalCostCfg:     false,
			utils.ReturnFileCodeCfg:       "0",
			utils.ResponseGroupCfg:        "03",
			utils.ResponseTypeCfg:         "D4",
			utils.RegulatoryCodeCfg:       "03",
			utils.ClientTrackingCfg:       "~*req.CGRID",
			utils.CustomerNumberCfg:       "~*req.Subject",
			utils.OrigNumberCfg:           "~*req.Subject",
			utils.TermNumberCfg:           "~*req.Destination",
			utils.BillToNumberCfg:         utils.EmptyString,
			utils.ZipcodeCfg:              utils.EmptyString,
			utils.Plus4Cfg:                utils.EmptyString,
			utils.P2PZipcodeCfg:           utils.EmptyString,
			utils.P2PPlus4Cfg:             utils.EmptyString,
			utils.UnitsCfg:                "1",
			utils.UnitTypeCfg:             "00",
			utils.TaxIncludedCfg:          "0",
			utils.TaxSitusRuleCfg:         "04",
			utils.TransTypeCodeCfg:        "010101",
			utils.SalesTypeCodeCfg:        "R",
			utils.TaxExemptionCodeListCfg: utils.EmptyString,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	cfgCgr.SureTaxCfg().Timezone = time.UTC
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SureTaxJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDispatcherS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		DispatcherSJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.IndexedSelectsCfg:      true,
			utils.PrefixIndexedFieldsCfg: []string{},
			utils.SuffixIndexedFieldsCfg: []string{},
			utils.NestedFieldsCfg:        false,
			utils.AttributeSConnsCfg:     []string{},
			utils.AnySubsystemCfg:        true,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{DispatcherSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigDispatcherH(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		RegistrarCJSON: map[string]interface{}{
			utils.DispatcherCfg: map[string]interface{}{
				utils.RegistrarsConnsCfg: []string{},
				utils.HostsCfg:           map[string][]map[string]interface{}{},
				utils.RefreshIntervalCfg: "5m0s",
			},
			utils.RPCCfg: map[string]interface{}{
				utils.RegistrarsConnsCfg: []string{},
				utils.HostsCfg:           map[string][]map[string]interface{}{},
				utils.RefreshIntervalCfg: "5m0s",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RegistrarCJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionLoader(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		LoaderJSON: map[string]interface{}{
			utils.TpIDCfg:            "",
			utils.DataPathCfg:        "./",
			utils.DisableReverseCfg:  false,
			utils.FieldSepCfg:        ",",
			utils.CachesConnsCfg:     []string{"*localhost"},
			utils.ActionSConnsCfg:    []string{"*localhost"},
			utils.GapiCredentialsCfg: json.RawMessage(`".gapi/credentials.json"`),
			utils.GapiTokenCfg:       json.RawMessage(`".gapi/token.json"`),
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionMigrator(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		MigratorJSON: map[string]interface{}{
			utils.OutDataDBTypeCfg:     "redis",
			utils.OutDataDBHostCfg:     "127.0.0.1",
			utils.OutDataDBPortCfg:     "6379",
			utils.OutDataDBNameCfg:     "10",
			utils.OutDataDBUserCfg:     "cgrates",
			utils.OutDataDBPasswordCfg: "",
			utils.OutDataDBEncodingCfg: "msgpack",
			utils.OutStorDBTypeCfg:     "mysql",
			utils.OutStorDBHostCfg:     "127.0.0.1",
			utils.OutStorDBPortCfg:     "3306",
			utils.OutStorDBNameCfg:     "cgrates",
			utils.OutStorDBUserCfg:     "cgrates",
			utils.OutStorDBPasswordCfg: "",
			utils.UsersFiltersCfg:      []string{},
			utils.OutStorDBOptsCfg:     map[string]interface{}{},
			utils.OutDataDBOptsCfg: map[string]interface{}{
				utils.RedisSentinelNameCfg:       "",
				utils.RedisClusterCfg:            false,
				utils.RedisClusterSyncCfg:        "5s",
				utils.RedisClusterOnDownDelayCfg: "0",
				utils.RedisTLS:                   false,
				utils.RedisClientCertificate:     "",
				utils.RedisClientKey:             "",
				utils.RedisCACertificate:         "",
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{MigratorJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionApierS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		AdminSJSON: map[string]interface{}{
			utils.EnabledCfg:         false,
			utils.CachesConnsCfg:     []string{utils.MetaInternal},
			utils.ActionSConnsCfg:    []string{},
			utils.AttributeSConnsCfg: []string{},
			utils.EEsConnsCfg:        []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AdminSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionEES(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		EEsJSON: map[string]interface{}{
			utils.EnabledCfg:         false,
			utils.AttributeSConnsCfg: []string{},
			utils.CacheCfg: map[string]interface{}{
				utils.MetaFileCSV: map[string]interface{}{
					utils.LimitCfg:     -1,
					utils.PrecacheCfg:  false,
					utils.ReplicateCfg: false,
					utils.TTLCfg:       "5s",
					utils.StaticTTLCfg: false,
				},
			},
			utils.ExportersCfg: []map[string]interface{}{
				{
					utils.IDCfg:               utils.MetaDefault,
					utils.TypeCfg:             utils.MetaNone,
					utils.ExportPathCfg:       "/var/spool/cgrates/ees",
					utils.OptsCfg:             map[string]interface{}{},
					utils.TenantCfg:           utils.EmptyString,
					utils.TimezoneCfg:         utils.EmptyString,
					utils.FiltersCfg:          []string{},
					utils.FlagsCfg:            []string{},
					utils.AttributeIDsCfg:     []string{},
					utils.AttributeContextCfg: utils.EmptyString,
					utils.SynchronousCfg:      false,
					utils.AttemptsCfg:         1,
					utils.FieldsCfg:           []map[string]interface{}{},
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{EEsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionERS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ERsJSON: map[string]interface{}{
			utils.EnabledCfg:       false,
			utils.SessionSConnsCfg: []string{utils.MetaInternal},
			utils.ReadersCfg: []map[string]interface{}{
				{
					utils.FiltersCfg:             []string{},
					utils.FlagsCfg:               []string{},
					utils.IDCfg:                  "*default",
					utils.ProcessedPathCfg:       "/var/spool/cgrates/ers/out",
					utils.RunDelayCfg:            "0",
					utils.SourcePathCfg:          "/var/spool/cgrates/ers/in",
					utils.TenantCfg:              utils.EmptyString,
					utils.TimezoneCfg:            utils.EmptyString,
					utils.CacheDumpFieldsCfg:     []map[string]interface{}{},
					utils.PartialCommitFieldsCfg: []map[string]interface{}{},
					utils.ConcurrentRequestsCfg:  1024,
					utils.TypeCfg:                utils.MetaNone,
					utils.FieldsCfg:              []string{},
					utils.OptsCfg: map[string]interface{}{
						"csvFieldSeparator":   ",",
						"csvHeaderDefineChar": ":",
						"csvRowLength":        0.,
						"xmlRootPath":         "",
						"partialOrderField":   "~*req.AnswerTime",
					},
				},
			},
			utils.PartialCacheTTLCfg:    "1s",
			utils.PartialCacheActionCfg: utils.MetaNone,
			utils.PartialPathCfg:        utils.EmptyString,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ERsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[ERsJSON].(map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[ERsJSON])
	} else {
		mp[utils.ReadersCfg].([]map[string]interface{})[0][utils.FieldsCfg] = []string{}
		if !reflect.DeepEqual(mp, expected[ERsJSON]) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected[ERsJSON]), utils.ToJSON(mp))
		}
	}
}

func TestV1GetConfigSectionRPConns(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		RPCConnsJSON: map[string]interface{}{
			utils.MetaBiJSONLocalHost: map[string]interface{}{
				utils.PoolSize:    0,
				utils.StrategyCfg: utils.MetaFirst,
				utils.Conns: []map[string]interface{}{
					{
						utils.AddressCfg:   "127.0.0.1:2014",
						utils.TransportCfg: rpcclient.BiRPCJSON,
					},
				},
			},
			utils.MetaLocalHost: map[string]interface{}{
				utils.PoolSize:    0,
				utils.StrategyCfg: utils.MetaFirst,
				utils.Conns: []map[string]interface{}{
					{
						utils.AddressCfg:   "127.0.0.1:2012",
						utils.TransportCfg: "*json",
					},
				},
			},
			utils.MetaInternal: map[string]interface{}{
				utils.StrategyCfg: utils.MetaFirst,
				utils.PoolSize:    0,
				utils.Conns: []map[string]interface{}{
					{
						utils.AddressCfg:   utils.MetaInternal,
						utils.TransportCfg: utils.EmptyString,
					},
				},
			},
			rpcclient.BiRPCInternal: map[string]interface{}{
				utils.StrategyCfg: utils.MetaFirst,
				utils.PoolSize:    0,
				utils.Conns: []map[string]interface{}{
					{
						utils.AddressCfg:   rpcclient.BiRPCInternal,
						utils.TransportCfg: utils.EmptyString,
					},
				},
			},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RPCConnsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionSIPAgent(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		SIPAgentJSON: map[string]interface{}{
			utils.EnabledCfg:             false,
			utils.ListenCfg:              "127.0.0.1:5060",
			utils.ListenNetCfg:           "udp",
			utils.SessionSConnsCfg:       []string{utils.MetaInternal},
			utils.TimezoneCfg:            utils.EmptyString,
			utils.RetransmissionTimerCfg: "1s",
			utils.RequestProcessorsCfg:   []map[string]interface{}{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{SIPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionTemplates(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		TemplatesJSON: map[string][]map[string]interface{}{
			utils.MetaErr: {
				{utils.TagCfg: "SessionId", utils.PathCfg: "*rep.Session-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginHost", utils.PathCfg: "*rep.Origin-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.OriginHost", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginRealm", utils.PathCfg: "*rep.Origin-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.OriginRealm", utils.MandatoryCfg: true},
			},
			utils.MetaASR: {
				{utils.TagCfg: "SessionId", utils.PathCfg: "*diamreq.Session-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Session-Id", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginHost", utils.PathCfg: "*diamreq.Origin-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Destination-Host", utils.MandatoryCfg: true},
				{utils.TagCfg: "OriginRealm", utils.PathCfg: "*diamreq.Origin-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Destination-Realm", utils.MandatoryCfg: true},
				{utils.TagCfg: "DestinationRealm", utils.PathCfg: "*diamreq.Destination-Realm", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Origin-Realm", utils.MandatoryCfg: true},
				{utils.TagCfg: "DestinationHost", utils.PathCfg: "*diamreq.Destination-Host", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*req.Origin-Host", utils.MandatoryCfg: true},
				{utils.TagCfg: "AuthApplicationId", utils.PathCfg: "*diamreq.Auth-Application-Id", utils.TypeCfg: "*variable",
					utils.ValueCfg: "~*vars.*appid", utils.MandatoryCfg: true},
			},
			utils.MetaCCA:    {},
			utils.MetaRAR:    {},
			"*errSip":        {},
			utils.MetaCdrLog: {},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{TemplatesJSON}}, &reply); err != nil {
		t.Error(err)
	} else if mp, can := reply[TemplatesJSON].(map[string][]map[string]interface{}); !can {
		t.Errorf("Unexpected type: %t", reply[TemplatesJSON])
	} else {
		mp[utils.MetaCCA] = []map[string]interface{}{}
		mp[utils.MetaRAR] = []map[string]interface{}{}
		mp["*errSip"] = []map[string]interface{}{}
		mp[utils.MetaCdrLog] = []map[string]interface{}{}
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func TestV1GetConfigSectionConfigs(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ConfigSJSON: map[string]interface{}{
			utils.EnabledCfg: true,
			utils.URLCfg:     "/configs/",
			utils.RootDirCfg: "/var/spool/cgrates/configs",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	cfgCgr.ConfigSCfg().Enabled = true
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ConfigSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}

	var result string
	cfgCgr2 := NewDefaultCGRConfig()
	if err = cfgCgr2.V1SetConfig(context.Background(), &SetConfigArgs{Config: reply, DryRun: true}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if cfgCgr := NewDefaultCGRConfig(); !reflect.DeepEqual(cfgCgr.ConfigSCfg(), cfgCgr2.ConfigSCfg()) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(cfgCgr.ConfigSCfg()), utils.ToJSON(cfgCgr2.ConfigSCfg()))
	}

	cfgCgr2 = NewDefaultCGRConfig()
	if err = cfgCgr2.V1SetConfig(context.Background(), &SetConfigArgs{Config: reply}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if !reflect.DeepEqual(cfgCgr.ConfigSCfg(), cfgCgr2.ConfigSCfg()) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(cfgCgr.ConfigSCfg()), utils.ToJSON(cfgCgr2.ConfigSCfg()))
	}
}

func TestV1GetConfigSectionAPIBans(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		APIBanJSON: map[string]interface{}{
			utils.EnabledCfg: false,
			utils.KeysCfg:    []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{APIBanJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionAnalyzer(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		AnalyzerSJSON: map[string]interface{}{
			utils.EnabledCfg:         false,
			utils.CleanupIntervalCfg: "1h0m0s",
			utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
			utils.IndexTypeCfg:       utils.MetaScorch,
			utils.TTLCfg:             "24h0m0s",
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{AnalyzerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionRateS(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		RateSJSON: map[string]interface{}{
			utils.EnabledCfg:                 false,
			utils.IndexedSelectsCfg:          true,
			utils.PrefixIndexedFieldsCfg:     []string{},
			utils.SuffixIndexedFieldsCfg:     []string{},
			utils.NestedFieldsCfg:            false,
			utils.RateIndexedSelectsCfg:      true,
			utils.RatePrefixIndexedFieldsCfg: []string{},
			utils.RateSuffixIndexedFieldsCfg: []string{},
			utils.RateNestedFieldsCfg:        false,
			utils.Verbosity:                  1000,
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{RateSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestV1GetConfigSectionInvalidSection(t *testing.T) {
	var reply map[string]interface{}
	expected := "Invalid section "
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{"invalidSection"}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ReloadConfigEmptyConfig(t *testing.T) {
	var reply string
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1SetConfig(context.Background(), &SetConfigArgs{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected output: %+v", reply)
	}
}

func TestV1ReloadConfigUnmarshalError(t *testing.T) {
	var reply string
	expected := "json: unsupported type: chan int"
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1SetConfig(context.Background(), &SetConfigArgs{
		Config: map[string]interface{}{
			StatSJSON: make(chan int),
		},
	},
		&reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ReloadConfigJSONWithLocks(t *testing.T) {
	var reply string
	section := map[string]interface{}{
		"inexistentSection": map[string]interface{}{},
	}
	expected := "Invalid section <inexistentSection> "
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1SetConfig(context.Background(), &SetConfigArgs{Config: section}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+q, received %+q", expected, err)
	}
}

func TestV1GetConfigAsJSONGeneral(t *testing.T) {
	var reply string
	strJSON := `{
		"general": {
			"node_id": "ENGINE1",
		}
	}`
	expected := `{"general":{"connect_attempts":5,"connect_timeout":"1s","dbdata_encoding":"*msgpack","default_caching":"*reload","default_category":"call","default_request_type":"*rated","default_tenant":"cgrates.org","default_timezone":"Local","digest_equal":":","digest_separator":",","failed_posts_dir":"/var/spool/cgrates/failed_posts","failed_posts_ttl":"5s","locking_timeout":"0","log_level":6,"logger":"*syslog","max_parallel_conns":100,"node_id":"ENGINE1","poster_attempts":3,"reconnects":-1,"reply_timeout":"2s","rounding_decimals":5,"rsr_separator":";","tpexport_dir":"/var/spool/cgrates/tpe"}}`
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(strJSON); err != nil {
		t.Error(err)
	} else if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{GeneralJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDataDB(t *testing.T) {
	var reply string
	expected := `{"data_db":{"db_host":"127.0.0.1","db_name":"10","db_password":"","db_port":6379,"db_type":"*redis","db_user":"cgrates","items":{"*accounts":{"remote":false,"replicate":false},"*action_profiles":{"remote":false,"replicate":false},"*actions":{"remote":false,"replicate":false},"*attribute_profiles":{"remote":false,"replicate":false},"*charger_profiles":{"remote":false,"replicate":false},"*dispatcher_hosts":{"remote":false,"replicate":false},"*dispatcher_profiles":{"remote":false,"replicate":false},"*filters":{"remote":false,"replicate":false},"*indexes":{"remote":false,"replicate":false},"*load_ids":{"remote":false,"replicate":false},"*rate_profiles":{"remote":false,"replicate":false},"*resource_profiles":{"remote":false,"replicate":false},"*resources":{"remote":false,"replicate":false},"*route_profiles":{"remote":false,"replicate":false},"*statqueue_profiles":{"remote":false,"replicate":false},"*statqueues":{"remote":false,"replicate":false},"*threshold_profiles":{"remote":false,"replicate":false},"*thresholds":{"remote":false,"replicate":false}},"opts":{"mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0","redisClusterSync":"5s","redisSentinel":"","redisTLS":false},"remote_conn_id":"","remote_conns":[],"replication_cache":"","replication_conns":[],"replication_filtered":false}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DataDBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONStorDB(t *testing.T) {
	var reply string
	expected := `{"stor_db":{"db_host":"127.0.0.1","db_name":"cgrates","db_password":"","db_port":3306,"db_type":"*mysql","db_user":"cgrates","items":{"*cdrs":{"remote":false,"replicate":false},"*session_costs":{"remote":false,"replicate":false},"*tp_accounts":{"remote":false,"replicate":false},"*tp_action_profiles":{"remote":false,"replicate":false},"*tp_attributes":{"remote":false,"replicate":false},"*tp_chargers":{"remote":false,"replicate":false},"*tp_dispatcher_hosts":{"remote":false,"replicate":false},"*tp_dispatcher_profiles":{"remote":false,"replicate":false},"*tp_filters":{"remote":false,"replicate":false},"*tp_rate_profiles":{"remote":false,"replicate":false},"*tp_resources":{"remote":false,"replicate":false},"*tp_routes":{"remote":false,"replicate":false},"*tp_stats":{"remote":false,"replicate":false},"*tp_thresholds":{"remote":false,"replicate":false},"*versions":{"remote":false,"replicate":false}},"opts":{"mongoQueryTimeout":"10s","mysqlLocation":"Local","sqlConnMaxLifetime":0,"sqlMaxIdleConns":10,"sqlMaxOpenConns":100,"sslMode":"disable"},"prefix_indexed_fields":[],"remote_conns":null,"replication_conns":null,"string_indexed_fields":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{StorDBJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTls(t *testing.T) {
	var reply string
	expected := `{"tls":{"ca_certificate":"","client_certificate":"","client_key":"","server_certificate":"","server_key":"","server_name":"","server_policy":4}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{TlsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTCache(t *testing.T) {
	var reply string
	expected := `{"caches":{"partitions":{"*account_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*accounts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*action_profile_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*action_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*apiban":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*attribute_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*caps_events":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*cdr_ids":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*cdrs":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*charger_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*charger_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*closed_sessions":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_hosts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_loads":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_routes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatchers":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*event_charges":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*filters":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*load_ids":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_profile_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*replication_hosts":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resource_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resource_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*reverse_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*route_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*route_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rpc_connections":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rpc_responses":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*session_costs":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*stat_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*statqueue_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*statqueues":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*stir":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*threshold_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*thresholds":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_accounts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_action_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_attributes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_chargers":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_dispatcher_hosts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_dispatcher_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_filters":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_rate_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_routes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_stats":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_thresholds":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*uch":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*versions":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""}},"replication_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CacheJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTListen(t *testing.T) {
	var reply string
	expected := `{"listen":{"http":"127.0.0.1:2080","http_tls":"127.0.0.1:2280","rpc_gob":"127.0.0.1:2013","rpc_gob_tls":"127.0.0.1:2023","rpc_json":"127.0.0.1:2012","rpc_json_tls":"127.0.0.1:2022"}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ListenJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAccounts(t *testing.T) {
	var reply string
	expected := `{"accounts":{"attributes_conns":[],"enabled":false,"indexed_selects":true,"max_iterations":1000,"max_usage":"259200000000000","nested_fields":false,"prefix_indexed_fields":[],"rates_conns":[],"suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cfg := NewDefaultCGRConfig()
	if err := cfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AccountSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONHTTP(t *testing.T) {
	var reply string
	expected := `{"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0","forceAttemptHttp2":true,"idleConnTimeout":"90s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0","skipTlsVerify":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFilterS(t *testing.T) {
	var reply string
	expected := `{"filters":{"admins_conns":[],"resources_conns":[],"stats_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{FilterSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCdrs(t *testing.T) {
	var reply string
	expected := `{"cdrs":{"actions_conns":[],"attributes_conns":[],"chargers_conns":[],"ees_conns":[],"enabled":false,"extra_fields":[],"online_cdr_exports":[],"session_cost_retries":5,"stats_conns":[],"store_cdrs":true,"thresholds_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CDRsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSessionS(t *testing.T) {
	var reply string
	expected := `{"sessions":{"actions_conns":[],"alterable_fields":[],"attributes_conns":[],"cdrs_conns":[],"channel_sync_interval":"0","chargers_conns":[],"client_protocol":1,"debit_interval":"0","default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":false,"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","replication_conns":[],"resources_conns":[],"routes_conns":[],"session_indexes":[],"session_ttl":"0","stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SessionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFreeSwitchAgent(t *testing.T) {
	var reply string
	expected := `{"freeswitch_agent":{"create_cdr":false,"empty_balance_ann_file":"","empty_balance_context":"","enabled":false,"event_socket_conns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","password":"ClueCon","reconnects":5}],"extra_fields":[],"low_balance_ann_file":"","max_wait_connection":"2s","sessions_conns":["*birpc_internal"],"subscribe_park":true}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{FreeSWITCHAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONFKamailioAgent(t *testing.T) {
	var reply string
	expected := `{"kamailio_agent":{"create_cdr":false,"enabled":false,"evapi_conns":[{"address":"127.0.0.1:8448","alias":"","reconnects":5}],"sessions_conns":["*birpc_internal"],"timezone":""}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{KamailioAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAsteriskAgent(t *testing.T) {
	var reply string
	expected := `{"asterisk_agent":{"asterisk_conns":[{"address":"127.0.0.1:8088","alias":"","connect_attempts":3,"password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"create_cdr":false,"enabled":false,"sessions_conns":["*birpc_internal"]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AsteriskAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONADiameterAgent(t *testing.T) {
	var reply string
	expected := `{"diameter_agent":{"asr_template":"","concurrent_requests":-1,"dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":false,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"","request_processors":[],"sessions_conns":["*birpc_internal"],"synced_conn_requests":false,"vendor_id":0}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DiameterAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONARadiusAgent(t *testing.T) {
	var reply string
	expected := `{"radius_agent":{"client_dictionaries":{"*default":"/usr/share/cgrates/radius/dict/"},"client_secrets":{"*default":"CGRateS.org"},"enabled":false,"listen_acct":"127.0.0.1:1813","listen_auth":"127.0.0.1:1812","listen_net":"udp","request_processors":[],"sessions_conns":["*internal"]}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RadiusAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDNSAgent(t *testing.T) {
	var reply string
	expected := `{"dns_agent":{"enabled":false,"listen":"127.0.0.1:2053","listen_net":"udp","request_processors":[],"sessions_conns":["*internal"],"timezone":""}}`
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DNSAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAttributes(t *testing.T) {
	var reply string
	expected := `{"attributes":{"admins_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"process_runs":1,"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AttributeSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONChargerS(t *testing.T) {
	var reply string
	expected := `{"chargers":{"attributes_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ChargerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONResourceS(t *testing.T) {
	var reply string
	expected := `{"resources":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ResourceSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONStatS(t *testing.T) {
	var reply string
	expected := `{"stats":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{StatSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONThresholdS(t *testing.T) {
	var reply string
	expected := `{"thresholds":{"actions_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ThresholdSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRouteS(t *testing.T) {
	var reply string
	expected := `{"routes":{"attributes_conns":[],"default_ratio":1,"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RouteSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSureTax(t *testing.T) {
	var reply string
	expected := `{"suretax":{"bill_to_number":"","business_unit":"","client_number":"","client_tracking":"~*req.CGRID","customer_number":"~*req.Subject","include_local_cost":false,"orig_number":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatory_code":"03","response_group":"03","response_type":"D4","return_file_code":"0","sales_type_code":"R","tax_exemption_code_list":"","tax_included":"0","tax_situs_rule":"04","term_number":"~*req.Destination","timezone":"UTC","trans_type_code":"010101","unit_type":"00","units":"1","url":"","validation_key":"","zipcode":""}}`
	cgrCfg := NewDefaultCGRConfig()

	cgrCfg.SureTaxCfg().Timezone = time.UTC
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SureTaxJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDispatcherS(t *testing.T) {
	var reply string
	expected := `{"dispatchers":{"any_subsystem":true,"attributes_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"suffix_indexed_fields":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{DispatcherSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONDispatcherH(t *testing.T) {
	var reply string
	expected := `{"registrarc":{"dispatchers":{"hosts":{},"refresh_interval":"5m0s","registrars_conns":[]},"rpc":{"hosts":{},"refresh_interval":"5m0s","registrars_conns":[]}}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RegistrarCJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONLoaders(t *testing.T) {
	var reply string
	expected := `{"loaders":[{"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"AttributeFilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.4"},{"path":"Path","tag":"Path","type":"*variable","value":"~*req.5"},{"path":"Type","tag":"Type","type":"*variable","value":"~*req.6"},{"path":"Value","tag":"Value","type":"*variable","value":"~*req.7"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.8"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.6"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.7"},{"path":"MetricFilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.4"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.5"},{"path":"RouteID","tag":"RouteID","type":"*variable","value":"~*req.6"},{"path":"RouteFilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.7"},{"path":"RouteAccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.8"},{"path":"RouteRatingPlanIDs","tag":"RouteRatingPlanIDs","type":"*variable","value":"~*req.9"},{"path":"RouteResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.10"},{"path":"RouteStatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.11"},{"path":"RouteWeight","tag":"RouteWeight","type":"*variable","value":"~*req.12"},{"path":"RouteBlocker","tag":"RouteBlocker","type":"*variable","value":"~*req.13"},{"path":"RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.14"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.4"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.5"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Contexts","tag":"Contexts","type":"*variable","value":"~*req.2"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.3"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.4"},{"path":"Strategy","tag":"Strategy","type":"*variable","value":"~*req.5"},{"path":"StrategyParameters","tag":"StrategyParameters","type":"*variable","value":"~*req.6"},{"path":"ConnID","tag":"ConnID","type":"*variable","value":"~*req.7"},{"path":"ConnFilterIDs","tag":"ConnFilterIDs","type":"*variable","value":"~*req.8"},{"path":"ConnWeight","tag":"ConnWeight","type":"*variable","value":"~*req.9"},{"path":"ConnBlocker","tag":"ConnBlocker","type":"*variable","value":"~*req.10"},{"path":"ConnParameters","tag":"ConnParameters","type":"*variable","value":"~*req.11"}],"file_name":"DispatcherProfiles.csv","flags":null,"type":"*dispatchers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Address","tag":"Address","type":"*variable","value":"~*req.2"},{"path":"Transport","tag":"Transport","type":"*variable","value":"~*req.3"},{"path":"TLS","tag":"TLS","type":"*variable","value":"~*req.4"}],"file_name":"DispatcherHosts.csv","flags":null,"type":"*dispatcher_hosts"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"path":"RateID","tag":"RateID","type":"*variable","value":"~*req.7"},{"path":"RateFilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"path":"RateActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"path":"RateWeight","tag":"RateWeight","type":"*variable","value":"~*req.10"},{"path":"RateBlocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"path":"RateIntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"path":"RateFixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"path":"RateRecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"path":"RateUnit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"path":"RateIncrement","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"RateProfiles.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.4"},{"path":"TargetType","tag":"TargetType","type":"*variable","value":"~*req.5"},{"path":"TargetIDs","tag":"TargetIDs","type":"*variable","value":"~*req.6"},{"path":"ActionID","tag":"ActionID","type":"*variable","value":"~*req.7"},{"path":"ActionFilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.8"},{"path":"ActionBlocker","tag":"ActionBlocker","type":"*variable","value":"~*req.9"},{"path":"ActionTTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"path":"ActionType","tag":"ActionType","type":"*variable","value":"~*req.11"},{"path":"ActionOpts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"path":"ActionPath","tag":"ActionPath","type":"*variable","value":"~*req.13"},{"path":"ActionValue","tag":"ActionValue","type":"*variable","value":"~*req.14"}],"file_name":"ActionProfiles.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"BalanceID","tag":"BalanceID","type":"*variable","value":"~*req.4"},{"path":"BalanceFilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.5"},{"path":"BalanceWeight","tag":"BalanceWeight","type":"*variable","value":"~*req.6"},{"path":"BalanceBlocker","tag":"BalanceBlocker","type":"*variable","value":"~*req.7"},{"path":"BalanceType","tag":"BalanceType","type":"*variable","value":"~*req.8"},{"path":"BalanceOpts","tag":"BalanceOpts","type":"*variable","value":"~*req.9"},{"path":"BalanceCostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.10"},{"path":"BalanceAttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.11"},{"path":"BalanceRateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.12"},{"path":"BalanceUnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.13"},{"path":"BalanceUnits","tag":"BalanceUnits","type":"*variable","value":"~*req.14"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.15"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"dry_run":false,"enabled":false,"field_separator":",","id":"*default","lock_filename":".cgr.lck","run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}]}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCgrLoader(t *testing.T) {
	var reply string
	expected := `{"loader":{"actions_conns":["*localhost"],"caches_conns":["*localhost"],"data_path":"./","disable_reverse":false,"field_separator":",","gapi_credentials":".gapi/credentials.json","gapi_token":".gapi/token.json","tpid":""}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{LoaderJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCgrMigrator(t *testing.T) {
	var reply string
	expected := `{"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","out_datadb_opts":{"redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0","redisClusterSync":"5s","redisSentinel":"","redisTLS":false},"out_datadb_password":"","out_datadb_port":"6379","out_datadb_type":"redis","out_datadb_user":"cgrates","out_stordb_host":"127.0.0.1","out_stordb_name":"cgrates","out_stordb_opts":{},"out_stordb_password":"","out_stordb_port":"3306","out_stordb_type":"mysql","out_stordb_user":"cgrates","users_filters":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{MigratorJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONApierS(t *testing.T) {
	var reply string
	expected := `{"admins":{"actions_conns":[],"attributes_conns":[],"caches_conns":["*internal"],"ees_conns":[],"enabled":false}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AdminSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCfgEES(t *testing.T) {
	var reply string
	expected := `{"ees":{"attributes_conns":[],"cache":{"*file_csv":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"enabled":false,"exporters":[{"attempts":1,"attribute_context":"","attribute_ids":[],"export_path":"/var/spool/cgrates/ees","fields":[],"filters":[],"flags":[],"id":"*default","opts":{},"synchronous":false,"tenant":"","timezone":"","type":"*none"}]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{EEsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCfgERS(t *testing.T) {
	var reply string
	expected := `{"ers":{"enabled":false,"partial_cache_action":"*none","partial_cache_ttl":"1s","partial_path":"","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"partialOrderField":"~*req.AnswerTime","xmlRootPath":""},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","run_delay":"0","source_path":"/var/spool/cgrates/ers/in","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ERsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONSIPAgent(t *testing.T) {
	var reply string
	expected := `{"sip_agent":{"enabled":false,"listen":"127.0.0.1:5060","listen_net":"udp","request_processors":[],"retransmission_timer":"1s","sessions_conns":["*internal"],"timezone":""}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{SIPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONConfigS(t *testing.T) {
	var reply string
	expected := `{"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{ConfigSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONApiBan(t *testing.T) {
	var reply string
	expected := `{"apiban":{"enabled":false,"keys":[]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{APIBanJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRPCConns(t *testing.T) {
	var reply string
	expected := `{"rpc_conns":{"*bijson_localhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RPCConnsJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONTemplates(t *testing.T) {
	var reply string
	expected := `{"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{TemplatesJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONHTTPAgent(t *testing.T) {
	var reply string
	expected := `{"http_agent":[]}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{HTTPAgentJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONAnalyzer(t *testing.T) {
	var reply string
	expected := `{"analyzers":{"cleanup_interval":"1h0m0s","db_path":"/var/spool/cgrates/analyzers","enabled":false,"index_type":"*scorch","ttl":"24h0m0s"}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{AnalyzerSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONRateS(t *testing.T) {
	var reply string
	expected := `{"rates":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"rate_indexed_selects":true,"rate_nested_fields":false,"rate_prefix_indexed_fields":[],"rate_suffix_indexed_fields":[],"suffix_indexed_fields":[],"verbosity":1000}}`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{RateSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1GetConfigAsJSONCoreS(t *testing.T) {
	var reply string
	expected := `{"cores":{"caps":10,"caps_stats_interval":"0","caps_strategy":"*busy","shutdown_timeout":"1s"}}`
	cgrCfg := NewDefaultCGRConfig()

	cgrCfg.coreSCfg.Caps = 10
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{CoreSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if expected != reply {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}

	var result string
	cfgCgr2 := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfgCgr2.rldChans[section] = make(chan struct{}, 1)
	}
	if err = cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: reply, DryRun: true}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if cgrCfg := NewDefaultCGRConfig(); !reflect.DeepEqual(cgrCfg.CoreSCfg(), cfgCgr2.CoreSCfg()) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(cgrCfg.CoreSCfg()), utils.ToJSON(cfgCgr2.CoreSCfg()))
	}
	for _, section := range sortedCfgSections {
		cfgCgr2.rldChans[section] = make(chan struct{}, 1)
	}
	if err = cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: reply}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result")
	} else if !reflect.DeepEqual(cgrCfg.CoreSCfg(), cfgCgr2.CoreSCfg()) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(cgrCfg.CoreSCfg()), utils.ToJSON(cfgCgr2.CoreSCfg()))
	}
}

func TestV1GetConfigAsJSONCheckConfigSanity(t *testing.T) {
	var result string
	args := `{
		"chargers": {
	        "enabled": true,
            "attributes_conns": ["*internal"]
    }
}`
	expected := `<AttributeS> not enabled but requested by <ChargerS> component`
	cfgCgr2 := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfgCgr2.rldChans[section] = make(chan struct{}, 1)
	}

	if err = cfgCgr2.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: args}, &result); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1GetConfigAsJSONInvalidSection(t *testing.T) {
	var reply string
	expected := `Invalid section `
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{Sections: []string{"InvalidSection"}}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1GetConfigAsJSONAllConfig(t *testing.T) {
	cfgJSON := `{
      "general": {
	      "node_id": "ENGINE1",											
	  }
}`
	var reply string
	expected := `{"accounts":{"attributes_conns":[],"enabled":false,"indexed_selects":true,"max_iterations":1000,"max_usage":"259200000000000","nested_fields":false,"prefix_indexed_fields":[],"rates_conns":[],"suffix_indexed_fields":[],"thresholds_conns":[]},"actions":{"accounts_conns":[],"cdrs_conns":[],"dynaprepaid_actionprofile":[],"ees_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"stats_conns":[],"suffix_indexed_fields":[],"tenants":[],"thresholds_conns":[]},"admins":{"actions_conns":[],"attributes_conns":[],"caches_conns":["*internal"],"ees_conns":[],"enabled":false},"analyzers":{"cleanup_interval":"1h0m0s","db_path":"/var/spool/cgrates/analyzers","enabled":false,"index_type":"*scorch","ttl":"24h0m0s"},"apiban":{"enabled":false,"keys":[]},"asterisk_agent":{"asterisk_conns":[{"address":"127.0.0.1:8088","alias":"","connect_attempts":3,"password":"CGRateS.org","reconnects":5,"user":"cgrates"}],"create_cdr":false,"enabled":false,"sessions_conns":["*birpc_internal"]},"attributes":{"admins_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"process_runs":1,"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]},"caches":{"partitions":{"*account_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*accounts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*action_profile_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*action_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*apiban":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2m0s"},"*attribute_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*attribute_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*caps_events":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*cdr_ids":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10m0s"},"*cdrs":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*charger_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*charger_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*closed_sessions":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*diameter_messages":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*dispatcher_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_hosts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_loads":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatcher_routes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*dispatchers":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*event_charges":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"10s"},"*event_resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*filters":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*load_ids":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_profile_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rate_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*replication_hosts":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resource_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resource_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*reverse_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*route_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*route_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rpc_connections":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*rpc_responses":{"limit":0,"precache":false,"replicate":false,"static_ttl":false,"ttl":"2s"},"*session_costs":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*stat_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*statqueue_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*statqueues":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*stir":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*threshold_filter_indexes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*threshold_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*thresholds":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_accounts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_action_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_attributes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_chargers":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_dispatcher_hosts":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_dispatcher_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_filters":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_rate_profiles":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_resources":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_routes":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_stats":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*tp_thresholds":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""},"*uch":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"3h0m0s"},"*versions":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":""}},"replication_conns":[]},"cdrs":{"actions_conns":[],"attributes_conns":[],"chargers_conns":[],"ees_conns":[],"enabled":false,"extra_fields":[],"online_cdr_exports":[],"session_cost_retries":5,"stats_conns":[],"store_cdrs":true,"thresholds_conns":[]},"chargers":{"attributes_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"suffix_indexed_fields":[]},"config_db":{"db_host":"","db_name":"","db_password":"","db_port":0,"db_type":"*internal","db_user":"","items":{},"opts":{"mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0","redisClusterSync":"5s","redisSentinel":"","redisTLS":false},"remote_conn_id":"","remote_conns":null,"replication_cache":"","replication_conns":null,"replication_filtered":false},"configs":{"enabled":false,"root_dir":"/var/spool/cgrates/configs","url":"/configs/"},"cores":{"caps":0,"caps_stats_interval":"0","caps_strategy":"*busy","shutdown_timeout":"1s"},"data_db":{"db_host":"127.0.0.1","db_name":"10","db_password":"","db_port":6379,"db_type":"*redis","db_user":"cgrates","items":{"*accounts":{"remote":false,"replicate":false},"*action_profiles":{"remote":false,"replicate":false},"*actions":{"remote":false,"replicate":false},"*attribute_profiles":{"remote":false,"replicate":false},"*charger_profiles":{"remote":false,"replicate":false},"*dispatcher_hosts":{"remote":false,"replicate":false},"*dispatcher_profiles":{"remote":false,"replicate":false},"*filters":{"remote":false,"replicate":false},"*indexes":{"remote":false,"replicate":false},"*load_ids":{"remote":false,"replicate":false},"*rate_profiles":{"remote":false,"replicate":false},"*resource_profiles":{"remote":false,"replicate":false},"*resources":{"remote":false,"replicate":false},"*route_profiles":{"remote":false,"replicate":false},"*statqueue_profiles":{"remote":false,"replicate":false},"*statqueues":{"remote":false,"replicate":false},"*threshold_profiles":{"remote":false,"replicate":false},"*thresholds":{"remote":false,"replicate":false}},"opts":{"mongoQueryTimeout":"10s","redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0","redisClusterSync":"5s","redisSentinel":"","redisTLS":false},"remote_conn_id":"","remote_conns":[],"replication_cache":"","replication_conns":[],"replication_filtered":false},"diameter_agent":{"asr_template":"","concurrent_requests":-1,"dictionaries_path":"/usr/share/cgrates/diameter/dict/","enabled":false,"forced_disconnect":"*none","listen":"127.0.0.1:3868","listen_net":"tcp","origin_host":"CGR-DA","origin_realm":"cgrates.org","product_name":"CGRateS","rar_template":"","request_processors":[],"sessions_conns":["*birpc_internal"],"synced_conn_requests":false,"vendor_id":0},"dispatchers":{"any_subsystem":true,"attributes_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"suffix_indexed_fields":[]},"dns_agent":{"enabled":false,"listen":"127.0.0.1:2053","listen_net":"udp","request_processors":[],"sessions_conns":["*internal"],"timezone":""},"ees":{"attributes_conns":[],"cache":{"*file_csv":{"limit":-1,"precache":false,"replicate":false,"static_ttl":false,"ttl":"5s"}},"enabled":false,"exporters":[{"attempts":1,"attribute_context":"","attribute_ids":[],"export_path":"/var/spool/cgrates/ees","fields":[],"filters":[],"flags":[],"id":"*default","opts":{},"synchronous":false,"tenant":"","timezone":"","type":"*none"}]},"ers":{"enabled":false,"partial_cache_action":"*none","partial_cache_ttl":"1s","partial_path":"","readers":[{"cache_dump_fields":[],"concurrent_requests":1024,"fields":[{"mandatory":true,"path":"*cgreq.ToR","tag":"ToR","type":"*variable","value":"~*req.2"},{"mandatory":true,"path":"*cgreq.OriginID","tag":"OriginID","type":"*variable","value":"~*req.3"},{"mandatory":true,"path":"*cgreq.RequestType","tag":"RequestType","type":"*variable","value":"~*req.4"},{"mandatory":true,"path":"*cgreq.Tenant","tag":"Tenant","type":"*variable","value":"~*req.6"},{"mandatory":true,"path":"*cgreq.Category","tag":"Category","type":"*variable","value":"~*req.7"},{"mandatory":true,"path":"*cgreq.Account","tag":"Account","type":"*variable","value":"~*req.8"},{"mandatory":true,"path":"*cgreq.Subject","tag":"Subject","type":"*variable","value":"~*req.9"},{"mandatory":true,"path":"*cgreq.Destination","tag":"Destination","type":"*variable","value":"~*req.10"},{"mandatory":true,"path":"*cgreq.SetupTime","tag":"SetupTime","type":"*variable","value":"~*req.11"},{"mandatory":true,"path":"*cgreq.AnswerTime","tag":"AnswerTime","type":"*variable","value":"~*req.12"},{"mandatory":true,"path":"*cgreq.Usage","tag":"Usage","type":"*variable","value":"~*req.13"}],"filters":[],"flags":[],"id":"*default","opts":{"csvFieldSeparator":",","csvHeaderDefineChar":":","csvRowLength":0,"partialOrderField":"~*req.AnswerTime","xmlRootPath":""},"partial_commit_fields":[],"processed_path":"/var/spool/cgrates/ers/out","run_delay":"0","source_path":"/var/spool/cgrates/ers/in","tenant":"","timezone":"","type":"*none"}],"sessions_conns":["*internal"]},"filters":{"admins_conns":[],"resources_conns":[],"stats_conns":[]},"freeswitch_agent":{"create_cdr":false,"empty_balance_ann_file":"","empty_balance_context":"","enabled":false,"event_socket_conns":[{"address":"127.0.0.1:8021","alias":"127.0.0.1:8021","password":"ClueCon","reconnects":5}],"extra_fields":[],"low_balance_ann_file":"","max_wait_connection":"2s","sessions_conns":["*birpc_internal"],"subscribe_park":true},"general":{"connect_attempts":5,"connect_timeout":"1s","dbdata_encoding":"*msgpack","default_caching":"*reload","default_category":"call","default_request_type":"*rated","default_tenant":"cgrates.org","default_timezone":"Local","digest_equal":":","digest_separator":",","failed_posts_dir":"/var/spool/cgrates/failed_posts","failed_posts_ttl":"5s","locking_timeout":"0","log_level":6,"logger":"*syslog","max_parallel_conns":100,"node_id":"ENGINE1","poster_attempts":3,"reconnects":-1,"reply_timeout":"2s","rounding_decimals":5,"rsr_separator":";","tpexport_dir":"/var/spool/cgrates/tpe"},"http":{"auth_users":{},"client_opts":{"dialFallbackDelay":"300ms","dialKeepAlive":"30s","dialTimeout":"30s","disableCompression":false,"disableKeepAlives":false,"expectContinueTimeout":"0","forceAttemptHttp2":true,"idleConnTimeout":"90s","maxConnsPerHost":0,"maxIdleConns":100,"maxIdleConnsPerHost":2,"responseHeaderTimeout":"0","skipTlsVerify":false,"tlsHandshakeTimeout":"10s"},"freeswitch_cdrs_url":"/freeswitch_json","http_cdrs":"/cdr_http","json_rpc_url":"/jsonrpc","registrars_url":"/registrar","use_basic_auth":false,"ws_url":"/ws"},"http_agent":[],"kamailio_agent":{"create_cdr":false,"enabled":false,"evapi_conns":[{"address":"127.0.0.1:8448","alias":"","reconnects":5}],"sessions_conns":["*birpc_internal"],"timezone":""},"listen":{"http":"127.0.0.1:2080","http_tls":"127.0.0.1:2280","rpc_gob":"127.0.0.1:2013","rpc_gob_tls":"127.0.0.1:2023","rpc_json":"127.0.0.1:2012","rpc_json_tls":"127.0.0.1:2022"},"loader":{"actions_conns":["*localhost"],"caches_conns":["*localhost"],"data_path":"./","disable_reverse":false,"field_separator":",","gapi_credentials":".gapi/credentials.json","gapi_token":".gapi/token.json","tpid":""},"loaders":[{"caches_conns":["*internal"],"data":[{"fields":[{"mandatory":true,"path":"Tenant","tag":"TenantID","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ProfileID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"AttributeFilterIDs","tag":"AttributeFilterIDs","type":"*variable","value":"~*req.4"},{"path":"Path","tag":"Path","type":"*variable","value":"~*req.5"},{"path":"Type","tag":"Type","type":"*variable","value":"~*req.6"},{"path":"Value","tag":"Value","type":"*variable","value":"~*req.7"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.8"}],"file_name":"Attributes.csv","flags":null,"type":"*attributes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Type","tag":"Type","type":"*variable","value":"~*req.2"},{"path":"Element","tag":"Element","type":"*variable","value":"~*req.3"},{"path":"Values","tag":"Values","type":"*variable","value":"~*req.4"}],"file_name":"Filters.csv","flags":null,"type":"*filters"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"UsageTTL","tag":"TTL","type":"*variable","value":"~*req.4"},{"path":"Limit","tag":"Limit","type":"*variable","value":"~*req.5"},{"path":"AllocationMessage","tag":"AllocationMessage","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.8"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.9"}],"file_name":"Resources.csv","flags":null,"type":"*resources"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"QueueLength","tag":"QueueLength","type":"*variable","value":"~*req.4"},{"path":"TTL","tag":"TTL","type":"*variable","value":"~*req.5"},{"path":"MinItems","tag":"MinItems","type":"*variable","value":"~*req.6"},{"path":"MetricIDs","tag":"MetricIDs","type":"*variable","value":"~*req.7"},{"path":"MetricFilterIDs","tag":"MetricFilterIDs","type":"*variable","value":"~*req.8"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.9"},{"path":"Stored","tag":"Stored","type":"*variable","value":"~*req.10"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.11"}],"file_name":"Stats.csv","flags":null,"type":"*stats"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"MaxHits","tag":"MaxHits","type":"*variable","value":"~*req.4"},{"path":"MinHits","tag":"MinHits","type":"*variable","value":"~*req.5"},{"path":"MinSleep","tag":"MinSleep","type":"*variable","value":"~*req.6"},{"path":"Blocker","tag":"Blocker","type":"*variable","value":"~*req.7"},{"path":"ActionProfileIDs","tag":"ActionProfileIDs","type":"*variable","value":"~*req.8"},{"path":"Async","tag":"Async","type":"*variable","value":"~*req.9"}],"file_name":"Thresholds.csv","flags":null,"type":"*thresholds"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Sorting","tag":"Sorting","type":"*variable","value":"~*req.4"},{"path":"SortingParameters","tag":"SortingParameters","type":"*variable","value":"~*req.5"},{"path":"RouteID","tag":"RouteID","type":"*variable","value":"~*req.6"},{"path":"RouteFilterIDs","tag":"RouteFilterIDs","type":"*variable","value":"~*req.7"},{"path":"RouteAccountIDs","tag":"RouteAccountIDs","type":"*variable","value":"~*req.8"},{"path":"RouteRatingPlanIDs","tag":"RouteRatingPlanIDs","type":"*variable","value":"~*req.9"},{"path":"RouteResourceIDs","tag":"RouteResourceIDs","type":"*variable","value":"~*req.10"},{"path":"RouteStatIDs","tag":"RouteStatIDs","type":"*variable","value":"~*req.11"},{"path":"RouteWeight","tag":"RouteWeight","type":"*variable","value":"~*req.12"},{"path":"RouteBlocker","tag":"RouteBlocker","type":"*variable","value":"~*req.13"},{"path":"RouteParameters","tag":"RouteParameters","type":"*variable","value":"~*req.14"}],"file_name":"Routes.csv","flags":null,"type":"*routes"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"RunID","tag":"RunID","type":"*variable","value":"~*req.4"},{"path":"AttributeIDs","tag":"AttributeIDs","type":"*variable","value":"~*req.5"}],"file_name":"Chargers.csv","flags":null,"type":"*chargers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Contexts","tag":"Contexts","type":"*variable","value":"~*req.2"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.3"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.4"},{"path":"Strategy","tag":"Strategy","type":"*variable","value":"~*req.5"},{"path":"StrategyParameters","tag":"StrategyParameters","type":"*variable","value":"~*req.6"},{"path":"ConnID","tag":"ConnID","type":"*variable","value":"~*req.7"},{"path":"ConnFilterIDs","tag":"ConnFilterIDs","type":"*variable","value":"~*req.8"},{"path":"ConnWeight","tag":"ConnWeight","type":"*variable","value":"~*req.9"},{"path":"ConnBlocker","tag":"ConnBlocker","type":"*variable","value":"~*req.10"},{"path":"ConnParameters","tag":"ConnParameters","type":"*variable","value":"~*req.11"}],"file_name":"DispatcherProfiles.csv","flags":null,"type":"*dispatchers"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"Address","tag":"Address","type":"*variable","value":"~*req.2"},{"path":"Transport","tag":"Transport","type":"*variable","value":"~*req.3"},{"path":"TLS","tag":"TLS","type":"*variable","value":"~*req.4"}],"file_name":"DispatcherHosts.csv","flags":null,"type":"*dispatcher_hosts"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"MinCost","tag":"MinCost","type":"*variable","value":"~*req.4"},{"path":"MaxCost","tag":"MaxCost","type":"*variable","value":"~*req.5"},{"path":"MaxCostStrategy","tag":"MaxCostStrategy","type":"*variable","value":"~*req.6"},{"path":"RateID","tag":"RateID","type":"*variable","value":"~*req.7"},{"path":"RateFilterIDs","tag":"RateFilterIDs","type":"*variable","value":"~*req.8"},{"path":"RateActivationTimes","tag":"RateActivationTimes","type":"*variable","value":"~*req.9"},{"path":"RateWeight","tag":"RateWeight","type":"*variable","value":"~*req.10"},{"path":"RateBlocker","tag":"RateBlocker","type":"*variable","value":"~*req.11"},{"path":"RateIntervalStart","tag":"RateIntervalStart","type":"*variable","value":"~*req.12"},{"path":"RateFixedFee","tag":"RateFixedFee","type":"*variable","value":"~*req.13"},{"path":"RateRecurrentFee","tag":"RateRecurrentFee","type":"*variable","value":"~*req.14"},{"path":"RateUnit","tag":"RateUnit","type":"*variable","value":"~*req.15"},{"path":"RateIncrement","tag":"RateIncrement","type":"*variable","value":"~*req.16"}],"file_name":"RateProfiles.csv","flags":null,"type":"*rate_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"Schedule","tag":"Schedule","type":"*variable","value":"~*req.4"},{"path":"TargetType","tag":"TargetType","type":"*variable","value":"~*req.5"},{"path":"TargetIDs","tag":"TargetIDs","type":"*variable","value":"~*req.6"},{"path":"ActionID","tag":"ActionID","type":"*variable","value":"~*req.7"},{"path":"ActionFilterIDs","tag":"ActionFilterIDs","type":"*variable","value":"~*req.8"},{"path":"ActionBlocker","tag":"ActionBlocker","type":"*variable","value":"~*req.9"},{"path":"ActionTTL","tag":"ActionTTL","type":"*variable","value":"~*req.10"},{"path":"ActionType","tag":"ActionType","type":"*variable","value":"~*req.11"},{"path":"ActionOpts","tag":"ActionOpts","type":"*variable","value":"~*req.12"},{"path":"ActionPath","tag":"ActionPath","type":"*variable","value":"~*req.13"},{"path":"ActionValue","tag":"ActionValue","type":"*variable","value":"~*req.14"}],"file_name":"ActionProfiles.csv","flags":null,"type":"*action_profiles"},{"fields":[{"mandatory":true,"path":"Tenant","tag":"Tenant","type":"*variable","value":"~*req.0"},{"mandatory":true,"path":"ID","tag":"ID","type":"*variable","value":"~*req.1"},{"path":"FilterIDs","tag":"FilterIDs","type":"*variable","value":"~*req.2"},{"path":"Weight","tag":"Weight","type":"*variable","value":"~*req.3"},{"path":"BalanceID","tag":"BalanceID","type":"*variable","value":"~*req.4"},{"path":"BalanceFilterIDs","tag":"BalanceFilterIDs","type":"*variable","value":"~*req.5"},{"path":"BalanceWeight","tag":"BalanceWeight","type":"*variable","value":"~*req.6"},{"path":"BalanceBlocker","tag":"BalanceBlocker","type":"*variable","value":"~*req.7"},{"path":"BalanceType","tag":"BalanceType","type":"*variable","value":"~*req.8"},{"path":"BalanceOpts","tag":"BalanceOpts","type":"*variable","value":"~*req.9"},{"path":"BalanceCostIncrements","tag":"BalanceCostIncrements","type":"*variable","value":"~*req.10"},{"path":"BalanceAttributeIDs","tag":"BalanceAttributeIDs","type":"*variable","value":"~*req.11"},{"path":"BalanceRateProfileIDs","tag":"BalanceRateProfileIDs","type":"*variable","value":"~*req.12"},{"path":"BalanceUnitFactors","tag":"BalanceUnitFactors","type":"*variable","value":"~*req.13"},{"path":"BalanceUnits","tag":"BalanceUnits","type":"*variable","value":"~*req.14"},{"path":"ThresholdIDs","tag":"ThresholdIDs","type":"*variable","value":"~*req.15"}],"file_name":"Accounts.csv","flags":null,"type":"*accounts"}],"dry_run":false,"enabled":false,"field_separator":",","id":"*default","lock_filename":".cgr.lck","run_delay":"0","tenant":"","tp_in_dir":"/var/spool/cgrates/loader/in","tp_out_dir":"/var/spool/cgrates/loader/out"}],"migrator":{"out_datadb_encoding":"msgpack","out_datadb_host":"127.0.0.1","out_datadb_name":"10","out_datadb_opts":{"redisCACertificate":"","redisClientCertificate":"","redisClientKey":"","redisCluster":false,"redisClusterOndownDelay":"0","redisClusterSync":"5s","redisSentinel":"","redisTLS":false},"out_datadb_password":"","out_datadb_port":"6379","out_datadb_type":"redis","out_datadb_user":"cgrates","out_stordb_host":"127.0.0.1","out_stordb_name":"cgrates","out_stordb_opts":{},"out_stordb_password":"","out_stordb_port":"3306","out_stordb_type":"mysql","out_stordb_user":"cgrates","users_filters":[]},"radius_agent":{"client_dictionaries":{"*default":"/usr/share/cgrates/radius/dict/"},"client_secrets":{"*default":"CGRateS.org"},"enabled":false,"listen_acct":"127.0.0.1:1813","listen_auth":"127.0.0.1:1812","listen_net":"udp","request_processors":[],"sessions_conns":["*internal"]},"rates":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"rate_indexed_selects":true,"rate_nested_fields":false,"rate_prefix_indexed_fields":[],"rate_suffix_indexed_fields":[],"suffix_indexed_fields":[],"verbosity":1000},"registrarc":{"dispatchers":{"hosts":{},"refresh_interval":"5m0s","registrars_conns":[]},"rpc":{"hosts":{},"refresh_interval":"5m0s","registrars_conns":[]}},"resources":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[],"thresholds_conns":[]},"routes":{"attributes_conns":[],"default_ratio":1,"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"resources_conns":[],"stats_conns":[],"suffix_indexed_fields":[]},"rpc_conns":{"*bijson_localhost":{"conns":[{"address":"127.0.0.1:2014","transport":"*birpc_json"}],"poolSize":0,"strategy":"*first"},"*birpc_internal":{"conns":[{"address":"*birpc_internal","transport":""}],"poolSize":0,"strategy":"*first"},"*internal":{"conns":[{"address":"*internal","transport":""}],"poolSize":0,"strategy":"*first"},"*localhost":{"conns":[{"address":"127.0.0.1:2012","transport":"*json"}],"poolSize":0,"strategy":"*first"}},"sessions":{"actions_conns":[],"alterable_fields":[],"attributes_conns":[],"cdrs_conns":[],"channel_sync_interval":"0","chargers_conns":[],"client_protocol":1,"debit_interval":"0","default_usage":{"*any":"3h0m0s","*data":"1048576","*sms":"1","*voice":"3h0m0s"},"enabled":false,"listen_bigob":"","listen_bijson":"127.0.0.1:2014","min_dur_low_balance":"0","replication_conns":[],"resources_conns":[],"routes_conns":[],"session_indexes":[],"session_ttl":"0","stats_conns":[],"stir":{"allowed_attest":["*any"],"default_attest":"A","payload_maxduration":"-1","privatekey_path":"","publickey_path":""},"store_session_costs":false,"terminate_attempts":5,"thresholds_conns":[]},"sip_agent":{"enabled":false,"listen":"127.0.0.1:5060","listen_net":"udp","request_processors":[],"retransmission_timer":"1s","sessions_conns":["*internal"],"timezone":""},"stats":{"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","store_uncompressed_limit":0,"suffix_indexed_fields":[],"thresholds_conns":[]},"stor_db":{"db_host":"127.0.0.1","db_name":"cgrates","db_password":"","db_port":3306,"db_type":"*mysql","db_user":"cgrates","items":{"*cdrs":{"remote":false,"replicate":false},"*session_costs":{"remote":false,"replicate":false},"*tp_accounts":{"remote":false,"replicate":false},"*tp_action_profiles":{"remote":false,"replicate":false},"*tp_attributes":{"remote":false,"replicate":false},"*tp_chargers":{"remote":false,"replicate":false},"*tp_dispatcher_hosts":{"remote":false,"replicate":false},"*tp_dispatcher_profiles":{"remote":false,"replicate":false},"*tp_filters":{"remote":false,"replicate":false},"*tp_rate_profiles":{"remote":false,"replicate":false},"*tp_resources":{"remote":false,"replicate":false},"*tp_routes":{"remote":false,"replicate":false},"*tp_stats":{"remote":false,"replicate":false},"*tp_thresholds":{"remote":false,"replicate":false},"*versions":{"remote":false,"replicate":false}},"opts":{"mongoQueryTimeout":"10s","mysqlLocation":"Local","sqlConnMaxLifetime":0,"sqlMaxIdleConns":10,"sqlMaxOpenConns":100,"sslMode":"disable"},"prefix_indexed_fields":[],"remote_conns":null,"replication_conns":null,"string_indexed_fields":[]},"suretax":{"bill_to_number":"","business_unit":"","client_number":"","client_tracking":"~*req.CGRID","customer_number":"~*req.Subject","include_local_cost":false,"orig_number":"~*req.Subject","p2pplus4":"","p2pzipcode":"","plus4":"","regulatory_code":"03","response_group":"03","response_type":"D4","return_file_code":"0","sales_type_code":"R","tax_exemption_code_list":"","tax_included":"0","tax_situs_rule":"04","term_number":"~*req.Destination","timezone":"UTC","trans_type_code":"010101","unit_type":"00","units":"1","url":"","validation_key":"","zipcode":""},"templates":{"*asr":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"}],"*cca":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"path":"*rep.Result-Code","tag":"ResultCode","type":"*constant","value":"2001"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"},{"mandatory":true,"path":"*rep.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"mandatory":true,"path":"*rep.CC-Request-Type","tag":"CCRequestType","type":"*variable","value":"~*req.CC-Request-Type"},{"mandatory":true,"path":"*rep.CC-Request-Number","tag":"CCRequestNumber","type":"*variable","value":"~*req.CC-Request-Number"}],"*cdrLog":[{"mandatory":true,"path":"*cdr.ToR","tag":"ToR","type":"*variable","value":"~*req.BalanceType"},{"mandatory":true,"path":"*cdr.OriginHost","tag":"OriginHost","type":"*constant","value":"127.0.0.1"},{"mandatory":true,"path":"*cdr.RequestType","tag":"RequestType","type":"*constant","value":"*none"},{"mandatory":true,"path":"*cdr.Tenant","tag":"Tenant","type":"*variable","value":"~*req.Tenant"},{"mandatory":true,"path":"*cdr.Account","tag":"Account","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Subject","tag":"Subject","type":"*variable","value":"~*req.Account"},{"mandatory":true,"path":"*cdr.Cost","tag":"Cost","type":"*variable","value":"~*req.Cost"},{"mandatory":true,"path":"*cdr.Source","tag":"Source","type":"*constant","value":"*cdrLog"},{"mandatory":true,"path":"*cdr.Usage","tag":"Usage","type":"*constant","value":"1"},{"mandatory":true,"path":"*cdr.RunID","tag":"RunID","type":"*variable","value":"~*req.ActionType"},{"mandatory":true,"path":"*cdr.SetupTime","tag":"SetupTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.AnswerTime","tag":"AnswerTime","type":"*constant","value":"*now"},{"mandatory":true,"path":"*cdr.PreRated","tag":"PreRated","type":"*constant","value":"true"}],"*err":[{"mandatory":true,"path":"*rep.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*rep.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*vars.OriginHost"},{"mandatory":true,"path":"*rep.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*vars.OriginRealm"}],"*errSip":[{"mandatory":true,"path":"*rep.Request","tag":"Request","type":"*constant","value":"SIP/2.0 500 Internal Server Error"}],"*rar":[{"mandatory":true,"path":"*diamreq.Session-Id","tag":"SessionId","type":"*variable","value":"~*req.Session-Id"},{"mandatory":true,"path":"*diamreq.Origin-Host","tag":"OriginHost","type":"*variable","value":"~*req.Destination-Host"},{"mandatory":true,"path":"*diamreq.Origin-Realm","tag":"OriginRealm","type":"*variable","value":"~*req.Destination-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Realm","tag":"DestinationRealm","type":"*variable","value":"~*req.Origin-Realm"},{"mandatory":true,"path":"*diamreq.Destination-Host","tag":"DestinationHost","type":"*variable","value":"~*req.Origin-Host"},{"mandatory":true,"path":"*diamreq.Auth-Application-Id","tag":"AuthApplicationId","type":"*variable","value":"~*vars.*appid"},{"path":"*diamreq.Re-Auth-Request-Type","tag":"ReAuthRequestType","type":"*constant","value":"0"}]},"thresholds":{"actions_conns":[],"enabled":false,"indexed_selects":true,"nested_fields":false,"prefix_indexed_fields":[],"store_interval":"","suffix_indexed_fields":[]},"tls":{"ca_certificate":"","client_certificate":"","client_key":"","server_certificate":"","server_key":"","server_name":"","server_policy":4}}`
	cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSON)
	if err != nil {
		t.Fatal(err)
	}
	cgrCfg.SureTaxCfg().Timezone = time.UTC
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{}, &reply); err != nil {
		t.Fatal(err)
	} else if expected != reply {
		t.Fatalf("Expected %+v \n, received %+v", expected, reply)
	}
	if err := cgrCfg.V1GetConfigAsJSON(context.Background(), &SectionWithAPIOpts{}, &reply); err != nil {
		t.Fatal(err)
	} else if expected != reply {
		t.Fatalf("Expected %+v \n, received %+v", expected, reply)
	}
}

func TestV1ReloadConfigFromJSONEmptyConfig(t *testing.T) {
	var reply string
	cgrCfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cgrCfg.rldChans[section] = make(chan struct{}, 1)
	}
	if err := cgrCfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: utils.EmptyString}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply")
	}
}

func TestV1ReloadConfigFromJSONInvalidSection(t *testing.T) {
	var reply string
	expected := "invalid character 'I' looking for beginning of value around line 1 and position 1\n line: \"InvalidSection\""
	cgrCfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cgrCfg.rldChans[section] = make(chan struct{}, 1)
	}
	if err := cgrCfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{Config: "InvalidSection"}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCgrCdfEventReader(t *testing.T) {
	eCfg := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.MetaNone,
				RunDelay:       0,
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields:     []*FCTemplate{},
				PartialCommitFields: []*FCTemplate{},
				Opts: map[string]interface{}{
					"csvFieldSeparator":   ",",
					"csvHeaderDefineChar": ":",
					"csvRowLength":        0.,
					"xmlRootPath":         "",
					"partialOrderField":   "~*req.AnswerTime",
				},
			},
		},
		PartialCacheTTL:    time.Second,
		PartialCacheAction: utils.MetaNone,
	}
	for _, profile := range eCfg.Readers {
		for _, v := range profile.Fields {
			v.ComputePath()
		}
	}
	if !reflect.DeepEqual(cgrCfg.ersCfg, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.ersCfg), utils.ToJSON(eCfg))
	}
}

func TestCgrCdfEventExporter(t *testing.T) {
	eCfg := &EEsCfg{
		Enabled:         false,
		AttributeSConns: []string{},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFileCSV: {
				Limit:     -1,
				TTL:       5 * time.Second,
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.MetaNone,
				Tenant:        nil,
				ExportPath:    "/var/spool/cgrates/ees",
				Attempts:      1,
				Timezone:      utils.EmptyString,
				Filters:       []string{},
				AttributeSIDs: []string{},
				Flags:         utils.FlagsWithParams{},
				Fields:        []*FCTemplate{},
				contentFields: []*FCTemplate{},
				headerFields:  []*FCTemplate{},
				trailerFields: []*FCTemplate{},
				Opts:          make(map[string]interface{}),
			},
		},
	}
	if !reflect.DeepEqual(cgrCfg.eesCfg, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.eesCfg), utils.ToJSON(eCfg))
	}
}

func TestCgrCfgEventReaderDefault(t *testing.T) {
	eCfg := &EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 1024,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Tenant:         nil,
		Timezone:       utils.EmptyString,
		Filters:        []string{},
		Flags:          utils.FlagsWithParams{},
		Fields: []*FCTemplate{
			{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.2", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.3", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.4", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.6", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.7", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.AccountField, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.8", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.9", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.10", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.11", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.12", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.13", utils.InfieldSep), Mandatory: true, Layout: time.RFC3339},
		},
		CacheDumpFields:     make([]*FCTemplate, 0),
		PartialCommitFields: make([]*FCTemplate, 0),
		Opts: map[string]interface{}{
			"csvFieldSeparator":   ",",
			"csvHeaderDefineChar": ":",
			"csvRowLength":        0.,
			"xmlRootPath":         "",
			"partialOrderField":   "~*req.AnswerTime",
		},
	}
	for _, v := range eCfg.Fields {
		v.ComputePath()
	}
	if !reflect.DeepEqual(cgrCfg.dfltEvRdr, eCfg) {
		t.Errorf("expecting: %+v \nreceived: %+v,", utils.ToJSON(eCfg), utils.ToJSON(cgrCfg.dfltEvRdr))
	}

}

func TestCgrCfgEventExporterDefault(t *testing.T) {
	eCfg := &EventExporterCfg{
		ID:            utils.MetaDefault,
		Type:          utils.MetaNone,
		Tenant:        nil,
		ExportPath:    "/var/spool/cgrates/ees",
		Attempts:      1,
		Timezone:      utils.EmptyString,
		Filters:       []string{},
		AttributeSIDs: []string{},
		Flags:         utils.FlagsWithParams{},
		contentFields: []*FCTemplate{},
		Fields:        []*FCTemplate{},
		headerFields:  []*FCTemplate{},
		trailerFields: []*FCTemplate{},
		Opts:          make(map[string]interface{}),
	}
	if !reflect.DeepEqual(cgrCfg.dfltEvExp, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.dfltEvExp), utils.ToJSON(eCfg))
	}

}

func TestRpcConnsDefaults(t *testing.T) {
	eCfg := make(RPCConns)
	// hardoded the *internal and *localhost connections
	eCfg[utils.MetaBiJSONLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{{
			Address:   "127.0.0.1:2014",
			Transport: rpcclient.BiRPCJSON,
		}},
	}
	eCfg[utils.MetaInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address: utils.MetaInternal,
			},
		},
	}
	eCfg[rpcclient.BiRPCInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address: rpcclient.BiRPCInternal,
			},
		},
	}
	eCfg[utils.MetaLocalHost] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
			},
		},
	}
	if !reflect.DeepEqual(cgrCfg.rpcConns, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.rpcConns), utils.ToJSON(eCfg))
	}
}

func TestCgrCfgJSONDefaultsConfigS(t *testing.T) {
	eCfg := &ConfigSCfg{
		Enabled: false,
		URL:     "/configs/",
		RootDir: "/var/spool/cgrates/configs",
	}
	if !reflect.DeepEqual(cgrCfg.configSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.configSCfg), utils.ToJSON(eCfg))
	}
}

func TestLoadConfigFromHTTP(t *testing.T) {
	cfgCgr := NewDefaultCGRConfig()

	url := "inexistentURL"
	expected := "parse \"inexistentURL\": invalid URI for request"
	if err := cfgCgr.loadConfigFromHTTP(url, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadConfigFromReaderError(t *testing.T) {
	expectedErrFile := "open randomfile.go: no such file or directory"
	file, err := os.Open("randomfile.go")
	expectedErr := "invalid argument"
	cgrCfg := NewDefaultCGRConfig()
	if err == nil || err.Error() != expectedErrFile {
		t.Errorf("Expected %+v, receivewd %+v", expectedErrFile, err)
	} else if err := cgrCfg.loadConfigFromReader(file, []func(ConfigDB) error{cfg.loadFromJSONCfg}, true); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadConfigFromReaderLoadFunctionsError(t *testing.T) {
	cfgJSONStr := `{
     "data_db": {								
	    "db_type": 123		
     }
}`
	expected := `json: cannot unmarshal number into Go struct field DbJsonCfg.Db_type of type string`
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.loadConfigFromReader(strings.NewReader(cfgJSONStr),
		[]func(jsonCfg ConfigDB) error{cfg.loadDataDBCfg},
		true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadCfgFromJSONWithLocksInvalidSeciton(t *testing.T) {
	expected := "Invalid section: <invalidSection> "
	cfg := NewDefaultCGRConfig()
	if err = cfg.loadCfgWithLocks("/random/path", "invalidSection"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCGRConfigClone(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	rcv := cfg.Clone()
	cfg.rldChans = nil
	rcv.rldChans = nil
	cfg.lks = nil
	rcv.lks = nil
	if !reflect.DeepEqual(cfg, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg), utils.ToJSON(rcv))
	}

	if !reflect.DeepEqual(cfg.dfltEvRdr, rcv.dfltEvRdr) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dfltEvRdr), utils.ToJSON(rcv.dfltEvRdr))
	}
	if !reflect.DeepEqual(cfg.dfltEvExp, rcv.dfltEvExp) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dfltEvExp), utils.ToJSON(rcv.dfltEvExp))
	}
	if !reflect.DeepEqual(cfg.loaderCfg, rcv.loaderCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.loaderCfg), utils.ToJSON(rcv.loaderCfg))
	}
	if !reflect.DeepEqual(cfg.httpAgentCfg, rcv.httpAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.httpAgentCfg), utils.ToJSON(rcv.httpAgentCfg))
	}
	if !reflect.DeepEqual(cfg.rpcConns, rcv.rpcConns) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.rpcConns), utils.ToJSON(rcv.rpcConns))
	}
	if !reflect.DeepEqual(cfg.templates, rcv.templates) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.templates), utils.ToJSON(rcv.templates))
	}
	if !reflect.DeepEqual(cfg.generalCfg, rcv.generalCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.generalCfg), utils.ToJSON(rcv.generalCfg))
	}
	if !reflect.DeepEqual(cfg.dataDbCfg, rcv.dataDbCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dataDbCfg), utils.ToJSON(rcv.dataDbCfg))
	}
	if !reflect.DeepEqual(cfg.storDbCfg, rcv.storDbCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.storDbCfg), utils.ToJSON(rcv.storDbCfg))
	}
	if !reflect.DeepEqual(cfg.tlsCfg, rcv.tlsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.tlsCfg), utils.ToJSON(rcv.tlsCfg))
	}
	if !reflect.DeepEqual(cfg.cacheCfg, rcv.cacheCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.cacheCfg), utils.ToJSON(rcv.cacheCfg))
	}
	if !reflect.DeepEqual(cfg.listenCfg, rcv.listenCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.listenCfg), utils.ToJSON(rcv.listenCfg))
	}
	if !reflect.DeepEqual(cfg.httpCfg, rcv.httpCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.httpCfg), utils.ToJSON(rcv.httpCfg))
	}
	if !reflect.DeepEqual(cfg.filterSCfg, rcv.filterSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.filterSCfg), utils.ToJSON(rcv.filterSCfg))
	}
	if !reflect.DeepEqual(cfg.cdrsCfg, rcv.cdrsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.cdrsCfg), utils.ToJSON(rcv.cdrsCfg))
	}
	if !reflect.DeepEqual(cfg.sessionSCfg, rcv.sessionSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sessionSCfg), utils.ToJSON(rcv.sessionSCfg))
	}
	if !reflect.DeepEqual(cfg.fsAgentCfg, rcv.fsAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.fsAgentCfg), utils.ToJSON(rcv.fsAgentCfg))
	}
	if !reflect.DeepEqual(cfg.kamAgentCfg, rcv.kamAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.kamAgentCfg), utils.ToJSON(rcv.kamAgentCfg))
	}
	if !reflect.DeepEqual(cfg.asteriskAgentCfg, rcv.asteriskAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.asteriskAgentCfg), utils.ToJSON(rcv.asteriskAgentCfg))
	}
	if !reflect.DeepEqual(cfg.diameterAgentCfg, rcv.diameterAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.diameterAgentCfg), utils.ToJSON(rcv.diameterAgentCfg))
	}
	if !reflect.DeepEqual(cfg.radiusAgentCfg, rcv.radiusAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.radiusAgentCfg), utils.ToJSON(rcv.radiusAgentCfg))
	}
	if !reflect.DeepEqual(cfg.dnsAgentCfg, rcv.dnsAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dnsAgentCfg), utils.ToJSON(rcv.dnsAgentCfg))
	}
	if !reflect.DeepEqual(cfg.attributeSCfg, rcv.attributeSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.attributeSCfg), utils.ToJSON(rcv.attributeSCfg))
	}
	if !reflect.DeepEqual(cfg.chargerSCfg, rcv.chargerSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.chargerSCfg), utils.ToJSON(rcv.chargerSCfg))
	}
	if !reflect.DeepEqual(cfg.resourceSCfg, rcv.resourceSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.resourceSCfg), utils.ToJSON(rcv.resourceSCfg))
	}
	if !reflect.DeepEqual(cfg.statsCfg, rcv.statsCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.statsCfg), utils.ToJSON(rcv.statsCfg))
	}
	if !reflect.DeepEqual(cfg.thresholdSCfg, rcv.thresholdSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.thresholdSCfg), utils.ToJSON(rcv.thresholdSCfg))
	}
	if !reflect.DeepEqual(cfg.routeSCfg, rcv.routeSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.routeSCfg), utils.ToJSON(rcv.routeSCfg))
	}
	if !reflect.DeepEqual(cfg.sureTaxCfg, rcv.sureTaxCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sureTaxCfg), utils.ToJSON(rcv.sureTaxCfg))
	}
	if !reflect.DeepEqual(cfg.dispatcherSCfg, rcv.dispatcherSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.dispatcherSCfg), utils.ToJSON(rcv.dispatcherSCfg))
	}
	if !reflect.DeepEqual(cfg.registrarCCfg, rcv.registrarCCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.registrarCCfg), utils.ToJSON(rcv.registrarCCfg))
	}
	if !reflect.DeepEqual(cfg.loaderCgrCfg, rcv.loaderCgrCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.loaderCgrCfg), utils.ToJSON(rcv.loaderCgrCfg))
	}
	if !reflect.DeepEqual(cfg.migratorCgrCfg, rcv.migratorCgrCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.migratorCgrCfg), utils.ToJSON(rcv.migratorCgrCfg))
	}
	if !reflect.DeepEqual(cfg.analyzerSCfg, rcv.analyzerSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.analyzerSCfg), utils.ToJSON(rcv.analyzerSCfg))
	}
	if !reflect.DeepEqual(cfg.admS, rcv.admS) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.admS), utils.ToJSON(rcv.admS))
	}
	if !reflect.DeepEqual(cfg.ersCfg, rcv.ersCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.ersCfg), utils.ToJSON(rcv.ersCfg))
	}
	if !reflect.DeepEqual(cfg.eesCfg, rcv.eesCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.eesCfg), utils.ToJSON(rcv.eesCfg))
	}
	if !reflect.DeepEqual(cfg.rateSCfg, rcv.rateSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.rateSCfg), utils.ToJSON(rcv.rateSCfg))
	}
	if !reflect.DeepEqual(cfg.sipAgentCfg, rcv.sipAgentCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.sipAgentCfg), utils.ToJSON(rcv.sipAgentCfg))
	}
	if !reflect.DeepEqual(cfg.configSCfg, rcv.configSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.configSCfg), utils.ToJSON(rcv.configSCfg))
	}
	if !reflect.DeepEqual(cfg.apiBanCfg, rcv.apiBanCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.apiBanCfg), utils.ToJSON(rcv.apiBanCfg))
	}
	if !reflect.DeepEqual(cfg.coreSCfg, rcv.coreSCfg) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cfg.coreSCfg), utils.ToJSON(rcv.coreSCfg))
	}
}

func TestActionSConfig(t *testing.T) {
	expected := &ActionSCfg{
		Enabled:                  false,
		EEsConns:                 []string{},
		CDRsConns:                []string{},
		ThresholdSConns:          []string{},
		StatSConns:               []string{},
		AccountSConns:            []string{},
		IndexedSelects:           true,
		Tenants:                  &[]string{},
		StringIndexedFields:      nil,
		PrefixIndexedFields:      &[]string{},
		SuffixIndexedFields:      &[]string{},
		NestedFields:             false,
		DynaprepaidActionProfile: []string{},
	}
	cgrConfig := NewDefaultCGRConfig()
	newConfig := cgrConfig.ActionSCfg()
	if !reflect.DeepEqual(expected, newConfig) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(newConfig))
	}
}

func TestV1GetConfigSectionActionSJSON(t *testing.T) {
	var reply map[string]interface{}
	expected := map[string]interface{}{
		ActionSJSON: map[string]interface{}{
			utils.EnabledCfg:                false,
			utils.EEsConnsCfg:               []string{},
			utils.CDRsConnsCfg:              []string{},
			utils.ThresholdSConnsCfg:        []string{},
			utils.StatSConnsCfg:             []string{},
			utils.AccountSConnsCfg:          []string{},
			utils.Tenants:                   []string{},
			utils.IndexedSelectsCfg:         true,
			utils.PrefixIndexedFieldsCfg:    []string{},
			utils.SuffixIndexedFieldsCfg:    []string{},
			utils.NestedFieldsCfg:           false,
			utils.DynaprepaidActionplansCfg: []string{},
		},
	}
	cfgCgr := NewDefaultCGRConfig()
	if err := cfgCgr.V1GetConfig(context.Background(), &SectionWithAPIOpts{Sections: []string{ActionSJSON}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func TestLoadActionSCfgError(t *testing.T) {
	cfgJSONStr := `{
      "actions": { 
            "string_indexed_fields": "*req.index",
	  }
    }`
	expected := "json: cannot unmarshal string into Go struct field ActionSJsonCfg.String_indexed_fields of type []string"
	cgrConfig := NewDefaultCGRConfig()

	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadActionSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadAccountSCfgError(t *testing.T) {
	cfgJSONStr := `{
"accounts": {								
	"enabled": "not_bool",
   }
}`
	expected := "json: cannot unmarshal string into Go struct field AccountSJsonCfg.Enabled of type bool"
	cfg := NewDefaultCGRConfig()

	if cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cfg.loadAccountSCfg(cgrCfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCGRConfigGetDP(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.LockSections(HTTPAgentJSON, LoaderJSON, ChargerSJSON)
	cfg.UnlockSections(HTTPAgentJSON, LoaderJSON, ChargerSJSON)
	exp := utils.MapStorage(cfg.AsMapInterface(cfg.generalCfg.RSRSep))
	dp := cfg.GetDataProvider()
	if !reflect.DeepEqual(dp, exp) {
		t.Errorf("Expected %+v, received %+v", exp, err)
	}
}

func TestStoreCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg.Enabled = true
	cfg.attributeSCfg.AdminSConns = []string{"*internal"}
	args := &SectionWithAPIOpts{
		Sections: []string{AttributeSJSON},
	}
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	rcv, err := db.AttributeServJsonCfg()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv.Enabled, utils.BoolPointer(true)) {
		t.Errorf("Expected %v \n but received \n %v", utils.BoolPointer(false), rcv.Enabled)
	} else if !reflect.DeepEqual(rcv.Admins_conns, &[]string{"*internal"}) {
		t.Errorf("Expected %v \n but received \n %v", &[]string{"*internal"}, rcv.Admins_conns)
	}
}

func TestSetCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:             true,
		ResourceSConns:      []string{"*internal"},
		StatSConns:          []string{"*internal"},
		AdminSConns:         []string{"*internal"},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"field1"},
		SuffixIndexedFields: &[]string{"field1"},
		PrefixIndexedFields: &[]string{"field1"},
		ProcessRuns:         2,
		NestedFields:        true,
	}
	cfg.rldChans[DataDBJSON] = make(chan struct{}, 1)
	cfg.rldChans[AttributeSJSON] = make(chan struct{}, 1)
	args := &SetConfigArgs{
		Config: map[string]interface{}{
			"attributes": &AttributeSJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Resources_conns:       &[]string{"*localhost"},
				Stats_conns:           &[]string{"*localhost"},
				Admins_conns:          &[]string{"*localhost"},
				Indexed_selects:       utils.BoolPointer(false),
				String_indexed_fields: &[]string{"field2"},
				Suffix_indexed_fields: &[]string{"field2"},
				Prefix_indexed_fields: &[]string{"field2"},
				Process_runs:          utils.IntPointer(3),
				Nested_fields:         utils.BoolPointer(false),
			},
		},
	}
	expected := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Resources_conns:       &[]string{"*localhost"},
		Stats_conns:           &[]string{"*localhost"},
		Admins_conns:          &[]string{"*localhost"},
		Indexed_selects:       utils.BoolPointer(false),
		String_indexed_fields: &[]string{"field2"},
		Suffix_indexed_fields: &[]string{"field2"},
		Prefix_indexed_fields: &[]string{"field2"},
		Process_runs:          utils.IntPointer(3),
		Nested_fields:         utils.BoolPointer(false),
	}
	var reply string
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	rcv, err := db.AttributeServJsonCfg()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func TestSetNilCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:             true,
		ResourceSConns:      []string{"*internal"},
		StatSConns:          []string{"*internal"},
		AdminSConns:         []string{"*internal"},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"field1"},
		SuffixIndexedFields: &[]string{"field1"},
		PrefixIndexedFields: &[]string{"field1"},
		ProcessRuns:         2,
		NestedFields:        true,
	}
	cfg.rldChans[DataDBJSON] = make(chan struct{}, 1)
	cfg.rldChans[AttributeSJSON] = make(chan struct{}, 1)
	attributes := &AttributeSJsonCfg{}
	args := &SetConfigArgs{
		Config: map[string]interface{}{
			"attributes": attributes,
		},
	}
	expected := &AttributeSJsonCfg{}
	var reply string
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

	rcv, err := db.AttributeServJsonCfg()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func TestReloadCfgInDb(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	db := &CgrJsonCfg{}
	cfg.db = db
	cfg.attributeSCfg = &AttributeSCfg{
		Enabled:             true,
		ResourceSConns:      []string{"*internal"},
		StatSConns:          []string{"*internal"},
		AdminSConns:         []string{"*internal"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"field1"},
		SuffixIndexedFields: &[]string{"field1"},
		PrefixIndexedFields: &[]string{"field1"},
		ProcessRuns:         2,
		NestedFields:        true,
	}
	var reply string
	cfg.rldChans[AttributeSJSON] = make(chan struct{}, 1)
	cfg.rldChans[DataDBJSON] = make(chan struct{}, 1)
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "attributes_internal")
	jsn := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Resources_conns:       &[]string{"*localhost"},
		Stats_conns:           &[]string{"*localhost"},
		Admins_conns:          &[]string{"*localhost"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"field2"},
		Suffix_indexed_fields: &[]string{"field2"},
		Prefix_indexed_fields: &[]string{"field2"},
		Process_runs:          utils.IntPointer(3),
		Nested_fields:         utils.BoolPointer(false),
	}
	db.SetSection(context.Background(), AttributeSJSON, jsn)
	expected := &AttributeSCfg{
		Enabled:             false,
		ResourceSConns:      []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		AdminSConns:         []string{"*localhost"},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"field2"},
		SuffixIndexedFields: &[]string{"field2"},
		PrefixIndexedFields: &[]string{"field2"},
		ProcessRuns:         3,
		NestedFields:        false,
	}
	args2 := &ReloadArgs{
		Section: AttributeSJSON,
	}
	if err := cfg.V1ReloadConfig(context.Background(), args2, &reply); err != nil {
		t.Error(err)
	}

	rcv := cfg.AttributeSCfg()
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
