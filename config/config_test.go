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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

var cfg *CGRConfig
var err error

func TestCgrCfgConfigSharing(t *testing.T) {
	cfg, _ = NewDefaultCGRConfig()
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
	eCgrCfg, err := NewDefaultCGRConfig()
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
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MONGO {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MONGO)
	} else if cgrCfg.DataDbCfg().DataDbPort != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "6379")
	}
	jsnCfg = `
{
"data_db": {
	"db_type": "internal",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().DataDbType != utils.INTERNAL {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.INTERNAL)
	} else if cgrCfg.DataDbCfg().DataDbPort != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "6379")
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
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MONGO {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MONGO)
	} else if cgrCfg.DataDbCfg().DataDbPort != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "27017")
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
	} else if cgrCfg.DataDbCfg().DataDbType != utils.INTERNAL {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.INTERNAL)
	} else if cgrCfg.DataDbCfg().DataDbPort != "internal" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "internal")
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
	} else if cgrCfg.StorDbCfg().Type != utils.MONGO {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.MONGO)
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
	} else if cgrCfg.StorDbCfg().Type != utils.MONGO {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.MONGO)
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
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.httpAgentCfg = []*HttpAgentCfg{
		{
			ID:                "conecto1",
			Url:               "/conecto",
			RequestPayload:    utils.MetaUrl,
			ReplyPayload:      utils.MetaXml,
			SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
			RequestProcessors: nil,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.HttpAgentCfg(), cgrCfg.HttpAgentCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.httpAgentCfg), utils.ToJSON(cgrCfg.httpAgentCfg))
	}
}

func TestCgrCfgLoadJSONDefaults(t *testing.T) {
	cgrCfg, err = NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
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
	if cgrCfg.GeneralCfg().ConnectTimeout != 1*time.Second {
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
		t.Errorf("Expected: utils.CSV_SEP , received: %+v", cgrCfg.GeneralCfg().DigestSeparator)
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
	if cgrCfg.DataDbCfg().DataDbType != "redis" {
		t.Errorf("Expecting: redis , received: %+v", cgrCfg.DataDbCfg().DataDbType)
	}
	if cgrCfg.DataDbCfg().DataDbHost != "127.0.0.1" {
		t.Errorf("Expecting: 127.0.0.1 , received: %+v", cgrCfg.DataDbCfg().DataDbHost)
	}
	if cgrCfg.DataDbCfg().DataDbPort != "6379" {
		t.Errorf("Expecting: 6379 , received: %+v", cgrCfg.DataDbCfg().DataDbPort)
	}
	if cgrCfg.DataDbCfg().DataDbName != "10" {
		t.Errorf("Expecting: 10 , received: %+v", cgrCfg.DataDbCfg().DataDbName)
	}
	if cgrCfg.DataDbCfg().DataDbUser != "cgrates" {
		t.Errorf("Expecting: cgrates , received: %+v", cgrCfg.DataDbCfg().DataDbUser)
	}
	if cgrCfg.DataDbCfg().DataDbPass != "" {
		t.Errorf("Expecting:  , received: %+v", cgrCfg.DataDbCfg().DataDbPass)
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
	if cgrCfg.StorDbCfg().Opts[utils.MaxOpenConnsCfg] != 100. {
		t.Errorf("Expecting: 100 , received: %+v", cgrCfg.StorDbCfg().Opts[utils.MaxOpenConnsCfg])
	}
	if cgrCfg.StorDbCfg().Opts[utils.MaxIdleConnsCfg] != 10. {
		t.Errorf("Expecting: 10 , received: %+v", cgrCfg.StorDbCfg().Opts[utils.MaxIdleConnsCfg])
	}
	if !reflect.DeepEqual(cgrCfg.StorDbCfg().StringIndexedFields, []string{}) {
		t.Errorf("Expecting: %+v , received: %+v", []string{}, cgrCfg.StorDbCfg().StringIndexedFields)
	}
	if !reflect.DeepEqual(cgrCfg.StorDbCfg().PrefixIndexedFields, []string{}) {
		t.Errorf("Expecting: %+v , received: %+v", []string{}, cgrCfg.StorDbCfg().PrefixIndexedFields)
	}
}

func TestCgrCfgJSONDefaultsRALs(t *testing.T) {
	eHaPoolcfg := []string{}

	if cgrCfg.RalsCfg().Enabled != false {
		t.Errorf("Expecting: false , received: %+v", cgrCfg.RalsCfg().Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.RalsCfg().ThresholdSConns, eHaPoolcfg) {
		t.Errorf("Expecting: %+v , received: %+v", eHaPoolcfg, cgrCfg.RalsCfg().ThresholdSConns)
	}
	if cgrCfg.RalsCfg().RpSubjectPrefixMatching != false {
		t.Errorf("Expecting: false , received: %+v", cgrCfg.RalsCfg().RpSubjectPrefixMatching)
	}
	eMaxCU := map[string]time.Duration{
		utils.ANY:   time.Duration(189 * time.Hour),
		utils.VOICE: time.Duration(72 * time.Hour),
		utils.DATA:  time.Duration(107374182400),
		utils.SMS:   time.Duration(10000),
		utils.MMS:   time.Duration(10000),
	}
	if !reflect.DeepEqual(eMaxCU, cgrCfg.RalsCfg().MaxComputedUsage) {
		t.Errorf("Expecting: %+v , received: %+v", eMaxCU, cgrCfg.RalsCfg().MaxComputedUsage)
	}
	if cgrCfg.RalsCfg().MaxIncrements != int(1000000) {
		t.Errorf("Expecting: 1000000 , received: %+v", cgrCfg.RalsCfg().MaxIncrements)
	}
	eBalRatingSbj := map[string]string{
		utils.ANY:   "*zero1ns",
		utils.VOICE: "*zero1s",
	}
	if !reflect.DeepEqual(eBalRatingSbj, cgrCfg.RalsCfg().BalanceRatingSubject) {
		t.Errorf("Expecting: %+v , received: %+v", eBalRatingSbj, cgrCfg.RalsCfg().BalanceRatingSubject)
	}
}

func TestCgrCfgJSONDefaultsScheduler(t *testing.T) {
	eSchedulerCfg := &SchedulerCfg{
		Enabled:   false,
		CDRsConns: []string{},
		Filters:   []string{},
	}
	if !reflect.DeepEqual(cgrCfg.schedulerCfg, eSchedulerCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.schedulerCfg, eSchedulerCfg)
	}
}

func TestCgrCfgJSONDefaultsCDRS(t *testing.T) {
	eCdrsCfg := &CdrsCfg{
		Enabled:         false,
		StoreCdrs:       true,
		SMCostRetries:   5,
		ChargerSConns:   []string{},
		RaterConns:      []string{},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
		SchedulerConns:  []string{},
		EEsConns:        []string{},
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
	"rals_conns": ["*internal"],
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
	expected = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().RaterConns, expected) {
		t.Errorf("Expecting: %+v , received: %+v", expected, cgrCfg.CdrsCfg().RaterConns)
	}
}

func TestCgrCfgJSONDefaultsSMGenericCfg(t *testing.T) {
	eSessionSCfg := &SessionSCfg{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		ChargerSConns:       []string{},
		RALsConns:           []string{},
		CDRsConns:           []string{},
		ResSConns:           []string{},
		ThreshSConns:        []string{},
		StatSConns:          []string{},
		RouteSConns:         []string{},
		AttrSConns:          []string{},
		ReplicationConns:    []string{},
		DebitInterval:       0 * time.Second,
		StoreSCosts:         false,
		SessionTTL:          0 * time.Second,
		SessionIndexes:      utils.StringMap{},
		ClientProtocol:      1.0,
		ChannelSyncInterval: 0,
		TerminateAttempts:   5,
		AlterableFields:     utils.NewStringSet([]string{}),
		STIRCfg: &STIRcfg{
			AllowedAttest:      utils.NewStringSet([]string{utils.META_ANY}),
			PayloadMaxduration: -1,
			DefaultAttest:      "A",
		},
		SchedulerConns: []string{},
	}
	if !reflect.DeepEqual(eSessionSCfg, cgrCfg.sessionSCfg) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSessionSCfg), utils.ToJSON(cgrCfg.sessionSCfg))
	}
}

func TestCgrCfgJSONDefaultsCacheCFG(t *testing.T) {
	eCacheCfg := &CacheCfg{
		Partitions: map[string]*CacheParamCfg{
			utils.CacheDestinations: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheReverseDestinations: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRatingPlans: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRatingProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheActions: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheActionPlans: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAccountActionPlans: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheActionTriggers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheSharedGroups: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTimings: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
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
			utils.CacheReverseFilterIndexes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherRoutes: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatcherLoads: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDispatchers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheDiameterMessages: {Limit: -1,
				TTL: time.Duration(3 * time.Hour), StaticTTL: false},
			utils.CacheRPCResponses: {Limit: 0,
				TTL: time.Duration(2 * time.Second), StaticTTL: false},
			utils.CacheClosedSessions: {Limit: -1,
				TTL: time.Duration(10 * time.Second), StaticTTL: false},
			utils.CacheEventCharges: {Limit: -1,
				TTL: time.Duration(10 * time.Second), StaticTTL: false},
			utils.CacheCDRIDs: {Limit: -1,
				TTL: time.Duration(10 * time.Minute), StaticTTL: false},
			utils.CacheLoadIDs: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheRPCConnections: {Limit: -1,
				TTL: 0, StaticTTL: false},
			utils.CacheUCH: {Limit: -1,
				TTL: time.Duration(3 * time.Hour), StaticTTL: false},
			utils.CacheSTIR: {Limit: -1,
				TTL: time.Duration(3 * time.Hour), StaticTTL: false},

			utils.CacheVersions: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheAccounts: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPTimings: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPDestinations: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPRates: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPDestinationRates: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPRatingPlans: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPRatingProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPSharedGroups: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPActions: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPActionPlans: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPActionTriggers: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.CacheTBLTPAccountActions: {Limit: -1,
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
			utils.CacheTBLTPRateProfiles: {Limit: -1,
				TTL: 0, StaticTTL: false, Precache: false},
			utils.MetaAPIBan: {Limit: -1,
				TTL: 2 * time.Minute, StaticTTL: false, Precache: false},
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
		SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
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
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
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
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
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
		ApierSConns:    []string{},
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
		RALsConns:           []string{},
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
		SessionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorId:          0,
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
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.VendorId, testDA.VendorId) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.VendorId, testDA.VendorId)
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

func TestCgrCfgJSONDefaultsMailer(t *testing.T) {
	if cgrCfg.MailerCfg().MailerServer != "localhost" {
		t.Error(cgrCfg.MailerCfg().MailerServer)
	}
	if cgrCfg.MailerCfg().MailerAuthUser != "cgrates" {
		t.Error(cgrCfg.MailerCfg().MailerAuthUser)
	}
	if cgrCfg.MailerCfg().MailerAuthPass != "CGRateS.org" {
		t.Error(cgrCfg.MailerCfg().MailerAuthPass)
	}
	if cgrCfg.MailerCfg().MailerFromAddr != "cgr-mailer@localhost.localdomain" {
		t.Error(cgrCfg.MailerCfg().MailerFromAddr)
	}
}

func TestCgrCfgJSONDefaultsSureTax(t *testing.T) {
	localt, err := time.LoadLocation("Local")
	if err != nil {
		t.Error("time parsing error", err)
	}
	eSureTaxCfg := &SureTaxCfg{
		Url:                  "",
		ClientNumber:         "",
		ValidationKey:        "",
		BusinessUnit:         "",
		Timezone:             localt,
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
		BillToNumber:         NewRSRParsersMustCompile("", utils.INFIELD_SEP),
		Zipcode:              NewRSRParsersMustCompile("", utils.INFIELD_SEP),
		P2PZipcode:           NewRSRParsersMustCompile("", utils.INFIELD_SEP),
		P2PPlus4:             NewRSRParsersMustCompile("", utils.INFIELD_SEP),
		Units:                NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
		UnitType:             NewRSRParsersMustCompile("00", utils.INFIELD_SEP),
		TaxIncluded:          NewRSRParsersMustCompile("0", utils.INFIELD_SEP),
		TaxSitusRule:         NewRSRParsersMustCompile("04", utils.INFIELD_SEP),
		TransTypeCode:        NewRSRParsersMustCompile("010101", utils.INFIELD_SEP),
		SalesTypeCode:        NewRSRParsersMustCompile("R", utils.INFIELD_SEP),
		TaxExemptionCodeList: NewRSRParsersMustCompile("", utils.INFIELD_SEP),
	}

	if !reflect.DeepEqual(cgrCfg.sureTaxCfg, eSureTaxCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.sureTaxCfg, eSureTaxCfg)
	}
}

func TestCgrCfgJSONDefaultsHTTP(t *testing.T) {
	if cgrCfg.HTTPCfg().HTTPJsonRPCURL != "/jsonrpc" {
		t.Errorf("expecting: /jsonrpc , received: %+v", cgrCfg.HTTPCfg().HTTPJsonRPCURL)
	}
	if cgrCfg.HTTPCfg().HTTPWSURL != "/ws" {
		t.Errorf("expecting: /ws , received: %+v", cgrCfg.HTTPCfg().HTTPWSURL)
	}
	if cgrCfg.HTTPCfg().HTTPFreeswitchCDRsURL != "/freeswitch_json" {
		t.Errorf("expecting: /freeswitch_json , received: %+v", cgrCfg.HTTPCfg().HTTPFreeswitchCDRsURL)
	}
	if cgrCfg.HTTPCfg().HTTPCDRsURL != "/cdr_http" {
		t.Errorf("expecting: /cdr_http , received: %+v", cgrCfg.HTTPCfg().HTTPCDRsURL)
	}
	if cgrCfg.HTTPCfg().HTTPUseBasicAuth != false {
		t.Errorf("expecting: false , received: %+v", cgrCfg.HTTPCfg().HTTPUseBasicAuth)
	}
	if !reflect.DeepEqual(cgrCfg.HTTPCfg().HTTPAuthUsers, map[string]string{}) {
		t.Errorf("expecting: %+v , received: %+v", map[string]string{}, cgrCfg.HTTPCfg().HTTPAuthUsers)
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
	dbs := []string{utils.MONGO, utils.REDIS, utils.MYSQL, utils.INTERNAL}
	for _, dbtype := range dbs {
		host := dbdf.dbHost(dbtype, flagInput)
		if host != utils.LOCALHOST {
			t.Errorf("received: %+v, expecting: %+v", host, utils.LOCALHOST)
		}
		user := dbdf.dbUser(dbtype, flagInput)
		if user != utils.CGRATES {
			t.Errorf("received: %+v, expecting: %+v", user, utils.CGRATES)
		}
		port := dbdf.dbPort(dbtype, flagInput)
		if port != dbdf[dbtype]["DbPort"] {
			t.Errorf("received: %+v, expecting: %+v", port, dbdf[dbtype]["DbPort"])
		}
		name := dbdf.dbName(dbtype, flagInput)
		if name != dbdf[dbtype]["DbName"] {
			t.Errorf("received: %+v, expecting: %+v", name, dbdf[dbtype]["DbName"])
		}
		pass := dbdf.dbPass(dbtype, flagInput)
		if pass != dbdf[dbtype]["DbPass"] {
			t.Errorf("received: %+v, expecting: %+v", pass, dbdf[dbtype]["DbPass"])
		}
	}
}

func TestDbDefaults(t *testing.T) {
	dbdf := newDbDefaults()
	flagInput := "NonMetaDynamic"
	dbs := []string{utils.MONGO, utils.REDIS, utils.MYSQL, utils.INTERNAL, utils.POSTGRES}
	for _, dbtype := range dbs {
		host := dbdf.dbHost(dbtype, flagInput)
		if host != flagInput {
			t.Errorf("Expected %+v, received %+v", flagInput, host)
		}
		user := dbdf.dbUser(dbtype, flagInput)
		if user != flagInput {
			t.Errorf("Expected %+v, received %+v", flagInput, user)
		}
		port := dbdf.dbPort(dbtype, "1234")
		if port != "1234" {
			t.Errorf("Expected %+v, received %+v", "1234", port)
		}
		name := dbdf.dbName(dbtype, utils.CGRATES)
		if name != utils.CGRATES {
			t.Errorf("Expected %+v, received %+v", utils.CGRATES, name)
		}
		pass := dbdf.dbPass(dbtype, utils.EmptyString)
		if pass != utils.EmptyString {
			t.Errorf("Expected %+v, received %+v", utils.EmptyString, pass)
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
	expected := "json: cannot unmarshal string into Go struct field RPCConnsJson.PoolSize of type int"
	cgrCfg, err := NewDefaultCGRConfig()
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
            "node_id": "ENGINE1",
            "locking_timeout": "0",
            "failed_posts_ttl": "0s",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
            "min_call_duration": "1s",
            "max_call_duration": []
        }
    }
}`
	expected := "json: cannot unmarshal array into Go struct field GeneralJsonCfg.Max_call_duration of type string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadGeneralCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadCacheCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadListenCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadHTTPCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDataDBCfgError(t *testing.T) {
	cfgJSONStr := `{
"data_db": {
	"db_host": 127.0,
	}
}`
	expected := "json: cannot unmarshal number into Go struct field DbJsonCfg.Db_host of type string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDataDBCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadStorDBCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadFilterSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadRalSCfgError(t *testing.T) {
	cfgJSONStr := `{
	"rals": {	
	    "stats_conns": "*internal",
    },
}`
	expected := "json: cannot unmarshal string into Go struct field RalsJsonCfg.Stats_conns of type []string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRalSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadSchedulerCfgError(t *testing.T) {
	cfgJSONStr := `{
	"schedulers": {
       "filters": "randomFilter",
    },
}`
	expected := "json: cannot unmarshal string into Go struct field SchedulerJsonCfg.Filters of type []string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSchedulerCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadCdrsCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSessionSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadFreeswitchAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadKamAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAsteriskAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDiameterAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRadiusAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDNSAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadHttpAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAttributeSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadChargerSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadResourceSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadStatSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadThresholdSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadLoaderSCfgError(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"run_delay": "0",
		},
	],	
}`
	expected := "json: cannot unmarshal string into Go struct field LoaderJsonCfg.Run_delay of type int"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadLoaderSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRouteSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadMailerCfgError(t *testing.T) {
	cfgJSONStr := `{
	"mailer": {
		"server": 1234,
		},
}`
	expected := "json: cannot unmarshal number into Go struct field MailerJsonCfg.Server of type string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadMailerCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSureTaxCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDispatcherSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadDispatcherHCfgError(t *testing.T) {
	cfgJSONStr := `{
		"dispatcherh":{
			"register_interval": 5,
		},		
}`
	expected := "json: cannot unmarshal number into Go struct field DispatcherHJsonCfg.Register_interval of type string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadDispatcherHCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadLoaderCgrCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadMigratorCgrCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadTlsCgrCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAnalyzerCgrCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadAPIBanCgrCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoadApierCfgError(t *testing.T) {
	myJSONStr := `{
    "apiers": {
       "scheduler_conns": "*internal",
    },
}`
	expected := "json: cannot unmarshal string into Go struct field ApierJsonCfg.Scheduler_conns of type []string"
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(myJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadApierCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadErsCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadEesCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadRateSCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadSIPAgentCfg(cgrCfgJson); err == nil || err.Error() != expected {
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
	cgrConfig, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	if cgrCfgJson, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if err := cgrConfig.loadTemplateSCfg(cgrCfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestCgrLoaderCfgITDefaults(t *testing.T) {
	eCfg := LoaderSCfgs{
		{
			Id:             utils.MetaDefault,
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:       "ProfileID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
						{
							Tag:    "Contexts",
							Path:   "Contexts",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{
							Tag:    "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{
							Tag:    "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{
							Tag:    "AttributeFilterIDs",
							Path:   "AttributeFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{
							Tag:    "Path",
							Path:   "Path",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Type",
							Path:   "Type",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Value",
							Path:   "Value",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Type",
							Path:   "Type",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Element",
							Path:   "Element",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Values",
							Path:   "Values",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "UsageTTL",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Limit",
							Path:   "Limit",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "AllocationMessage",
							Path:   "AllocationMessage",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:   "MetricIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MetricFilterIDs",
							Path:   "MetricFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP),
							Layout: time.RFC3339},

						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MaxHits",
							Path:   "MaxHits",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MinHits",
							Path:   "MinHits",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "MinSleep",
							Path:   "MinSleep",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActionIDs",
							Path:   "ActionIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Async",
							Path:   "Async",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteID",
							Path:   "RouteID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteFilterIDs",
							Path:   "RouteFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteAccountIDs",
							Path:   "RouteAccountIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteRatingPlanIDs",
							Path:   "RouteRatingPlanIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteResourceIDs",
							Path:   "RouteResourceIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteStatIDs",
							Path:   "RouteStatIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteWeight",
							Path:   "RouteWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteBlocker",
							Path:   "RouteBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RouteParameters",
							Path:   "RouteParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.15", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "RunID",
							Path:   "RunID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "AttributeIDs",
							Path:   "AttributeIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Contexts",
							Path:   "Contexts",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Strategy",
							Path:   "Strategy",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "StrategyParameters",
							Path:   "StrategyParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnID",
							Path:   "ConnID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnFilterIDs",
							Path:   "ConnFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnWeight",
							Path:   "ConnWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnBlocker",
							Path:   "ConnBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnParameters",
							Path:   "ConnParameters",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Address",
							Path:   "Address",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Transport",
							Path:   "Transport",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "TLS",
							Path:   "TLS",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ActivationInterval",
							Path:   "ActivationInterval",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "Weight",
							Path:   "Weight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "ConnectFee",
							Path:   "ConnectFee",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RoundingMethod",
							Path:   "RoundingMethod",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RoundingDecimals",
							Path:   "RoundingDecimals",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "MinCost",
							Path:   "MinCost",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCost",
							Path:   "MaxCost",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCostStrategy",
							Path:   "MaxCostStrategy",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateID",
							Path:   "RateID",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateFilterIDs",
							Path:   "RateFilterIDs",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateActivationStart",
							Path:   "RateActivationStart",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateWeight",
							Path:   "RateWeight",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.14", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateBlocker",
							Path:   "RateBlocker",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.15", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateIntervalStart",
							Path:   "RateIntervalStart",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.16", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateValue",
							Path:   "RateValue",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.17", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateUnit",
							Path:   "RateUnit",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.18", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
						{Tag: "RateIncrement",
							Path:   "RateIncrement",
							Type:   utils.MetaVariable,
							Value:  NewRSRParsersMustCompile("~*req.19", utils.INFIELD_SEP),
							Layout: time.RFC3339,
						},
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
		FieldSeparator:  rune(utils.CSV_SEP),
		CachesConns:     []string{utils.MetaLocalHost},
		SchedulerConns:  []string{utils.MetaLocalHost},
		GapiCredentials: json.RawMessage(`".gapi/credentials.json"`),
		GapiToken:       json.RawMessage(`".gapi/token.json"`),
	}
	if !reflect.DeepEqual(cgrCfg.LoaderCgrCfg(), eLdrCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.LoaderCgrCfg()), utils.ToJSON(eLdrCfg))
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
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.tlsCfg = &TlsCfg{
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
	} else if !reflect.DeepEqual(eCgrCfg.TlsCfg(), cgrCfg.TlsCfg()) {
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
	aCfg := &ApierCfg{
		Enabled:         false,
		CachesConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		SchedulerConns:  []string{},
		AttributeSConns: []string{},
		EEsConns:        []string{},
	}
	if !reflect.DeepEqual(cgrCfg.apier, aCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.apier, aCfg)
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
	}
	if !reflect.DeepEqual(cgrCfg.rateSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.rateSCfg, eCfg)
	}
}

func TestCgrCfgV1GetConfigSection(t *testing.T) {
	jsnCfg := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`
	expected := map[string]interface{}{
		"http":         ":2080",
		"http_tls":     "127.0.0.1:2280",
		"rpc_gob":      ":2013",
		"rpc_gob_tls":  "127.0.0.1:2023",
		"rpc_json":     ":2012",
		"rpc_json_tls": "127.0.0.1:2022",
	}
	var rcv map[string]interface{}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(jsnCfg); err != nil {
		t.Error(err)
	} else if err := cgrCfg.V1GetConfigSection(&SectionWithOpts{Section: LISTEN_JSN}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCgrCdfEventReader(t *testing.T) {
	eCfg := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Readers: []*EventReaderCfg{
			{
				ID:               utils.MetaDefault,
				Type:             utils.META_NONE,
				FieldSep:         ",",
				HeaderDefineChar: ":",
				RunDelay:         0,
				ConcurrentReqs:   1024,
				SourcePath:       "/var/spool/cgrates/ers/in",
				ProcessedPath:    "/var/spool/cgrates/ers/out",
				XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
				Tenant:           nil,
				Timezone:         utils.EmptyString,
				Filters:          []string{},
				Flags:            utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
				},
				CacheDumpFields: []*FCTemplate{},
				Opts:            make(map[string]interface{}),
			},
		},
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
				TTL:       time.Duration(5 * time.Second),
				StaticTTL: false,
			},
		},
		Exporters: []*EventExporterCfg{
			{
				ID:            utils.MetaDefault,
				Type:          utils.META_NONE,
				FieldSep:      ",",
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
		ID:               utils.MetaDefault,
		Type:             utils.META_NONE,
		FieldSep:         ",",
		HeaderDefineChar: ":",
		RunDelay:         0,
		ConcurrentReqs:   1024,
		SourcePath:       "/var/spool/cgrates/ers/in",
		ProcessedPath:    "/var/spool/cgrates/ers/out",
		XmlRootPath:      utils.HierarchyPath{utils.EmptyString},
		Tenant:           nil,
		Timezone:         utils.EmptyString,
		Filters:          nil,
		Flags:            utils.FlagsWithParams{},
		Fields: []*FCTemplate{
			{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
			{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP), Mandatory: true, Layout: time.RFC3339},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
		Opts:            make(map[string]interface{}),
	}
	for _, v := range eCfg.Fields {
		v.ComputePath()
	}
	if !reflect.DeepEqual(cgrCfg.dfltEvRdr, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.dfltEvRdr), utils.ToJSON(eCfg))
	}

}

func TestCgrCfgEventExporterDefault(t *testing.T) {
	eCfg := &EventExporterCfg{
		ID:            utils.MetaDefault,
		Type:          utils.META_NONE,
		FieldSep:      ",",
		Tenant:        nil,
		ExportPath:    "/var/spool/cgrates/ees",
		Attempts:      1,
		Timezone:      utils.EmptyString,
		Filters:       nil,
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
	eCfg := make(RpcConns)
	// hardoded the *internal and *localhost connections
	eCfg[utils.MetaInternal] = &RPCConn{
		Strategy: rpcclient.PoolFirst,
		PoolSize: 0,
		Conns: []*RemoteHost{
			{
				Address: utils.MetaInternal,
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
		Url:     "/configs/",
		RootDir: "/var/spool/cgrates/configs",
	}
	if !reflect.DeepEqual(cgrCfg.configSCfg, eCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.configSCfg), utils.ToJSON(eCfg))
	}
}
