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
	"encoding/json"
	"path"
	"reflect"
	"sync"
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
	JSN_CFG := `
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg)
	}
}

func TestCgrCfgDataDBPortWithoutDynamic(t *testing.T) {
	JSN_CFG := `
{
"data_db": {
	"db_type": "mongo",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MetaMongo)
	} else if cgrCfg.DataDbCfg().DataDbPort != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "6379")
	}
	JSN_CFG = `
{
"data_db": {
	"db_type": "internal",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MetaInternal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MetaInternal)
	} else if cgrCfg.DataDbCfg().DataDbPort != "6379" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "6379")
	}
}

func TestCgrCfgDataDBPortWithDymanic(t *testing.T) {
	JSN_CFG := `
{
"data_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MetaMongo)
	} else if cgrCfg.DataDbCfg().DataDbPort != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "27017")
	}
	JSN_CFG = `
{
"data_db": {
	"db_type": "internal",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.DataDbCfg().DataDbType != utils.MetaInternal {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbType, utils.MetaInternal)
	} else if cgrCfg.DataDbCfg().DataDbPort != "internal" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.DataDbCfg().DataDbPort, "internal")
	}
}

func TestCgrCfgStorDBPortWithoutDynamic(t *testing.T) {
	JSN_CFG := `
{
"stor_db": {
	"db_type": "mongo",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.StorDbCfg().Type != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.MetaMongo)
	} else if cgrCfg.StorDbCfg().Port != "3306" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Port, "3306")
	}
}

func TestCgrCfgStorDBPortWithDymanic(t *testing.T) {
	JSN_CFG := `
{
"stor_db": {
	"db_type": "mongo",
	"db_port": -1,
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.StorDbCfg().Type != utils.MetaMongo {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Type, utils.MetaMongo)
	} else if cgrCfg.StorDbCfg().Port != "27017" {
		t.Errorf("Expected: %+v, received: %+v", cgrCfg.StorDbCfg().Port, "27017")
	}
}

func TestCgrCfgListener(t *testing.T) {
	JSN_CFG := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.ListenCfg().RPCGOBTLSListen != "127.0.0.1:2023" {
		t.Errorf("Expected: 127.0.0.1:2023 , received: %+v", cgrCfg.ListenCfg().RPCGOBTLSListen)
	} else if cgrCfg.ListenCfg().RPCJSONTLSListen != "127.0.0.1:2022" {
		t.Errorf("Expected: 127.0.0.1:2022 , received: %+v", cgrCfg.ListenCfg().RPCJSONTLSListen)
	}
}

func TestHttpAgentCfg(t *testing.T) {
	JSN_RAW_CFG := `
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_RAW_CFG); err != nil {
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
	if cgrCfg.GeneralCfg().HttpSkipTlsVerify != false {
		t.Errorf("Expected: false, received: %+v", cgrCfg.GeneralCfg().HttpSkipTlsVerify)
	}
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
	if cgrCfg.DataDbCfg().DataDbType != "*redis" {
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
	if cgrCfg.StorDbCfg().Type != "*mysql" {
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
	if cgrCfg.StorDbCfg().MaxOpenConns != 100 {
		t.Errorf("Expecting: 100 , received: %+v", cgrCfg.StorDbCfg().MaxOpenConns)
	}
	if cgrCfg.StorDbCfg().MaxIdleConns != 10 {
		t.Errorf("Expecting: 10 , received: %+v", cgrCfg.StorDbCfg().MaxIdleConns)
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
	emptySlice := []string{}
	var eCdrExtr []*utils.RSRField
	if cgrCfg.CdrsCfg().Enabled != false {
		t.Errorf("Expecting: false , received: %+v", cgrCfg.CdrsCfg().Enabled)
	}
	if !reflect.DeepEqual(eCdrExtr, cgrCfg.CdrsCfg().ExtraFields) {
		t.Errorf("Expecting: %+v , received: %+v", eCdrExtr, cgrCfg.CdrsCfg().ExtraFields)
	}
	if cgrCfg.CdrsCfg().StoreCdrs != true {
		t.Errorf("Expecting: true , received: %+v", cgrCfg.CdrsCfg().StoreCdrs)
	}
	if cgrCfg.CdrsCfg().SMCostRetries != 5 {
		t.Errorf("Expecting: 5 , received: %+v", cgrCfg.CdrsCfg().SMCostRetries)
	}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().RaterConns, emptySlice) {
		t.Errorf("Expecting: %+v , received: %+v", emptySlice, cgrCfg.CdrsCfg().RaterConns)
	}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().ChargerSConns, emptySlice) {
		t.Errorf("Expecting: %+v , received: %+v", emptySlice, cgrCfg.CdrsCfg().ChargerSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().AttributeSConns, emptySlice) {
		t.Errorf("Expecting: %+v , received: %+v", emptySlice, cgrCfg.CdrsCfg().AttributeSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().ThresholdSConns, emptySlice) {
		t.Errorf("Expecting: %+v , received: %+v", emptySlice, cgrCfg.CdrsCfg().ThresholdSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CdrsCfg().StatSConns, emptySlice) {
		t.Errorf("Expecting: %+v , received: %+v", emptySlice, cgrCfg.CdrsCfg().StatSConns)
	}
	if cgrCfg.CdrsCfg().OnlineCDRExports != nil {
		t.Errorf("Expecting: nil , received: %+v", cgrCfg.CdrsCfg().OnlineCDRExports)
	}
}

func TestCgrCfgJSONLoadCDRS(t *testing.T) {
	JSN_RAW_CFG := `
{
"cdrs": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"],
},
}
	`
	cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_RAW_CFG)
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

func TestCgrCfgJSONDefaultsCdreProfiles(t *testing.T) {
	eContentFlds := []*FCTemplate{
		{
			Tag:   "*exp.CGRID",
			Path:  "*exp.CGRID",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.CGRID", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.RunID",
			Path:  "*exp.RunID",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.RunID", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.ToR",
			Path:  "*exp.ToR",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.ToR", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.OriginID",
			Path:  "*exp.OriginID",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.OriginID", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.RequestType",
			Path:  "*exp.RequestType",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.RequestType", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.Tenant",
			Path:  "*exp.Tenant",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Tenant", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.Category",
			Path:  "*exp.Category",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Category", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.Account",
			Path:  "*exp.Account",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.Subject",
			Path:  "*exp.Subject",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
		},
		{
			Tag:   "*exp.Destination",
			Path:  "*exp.Destination",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP),
		},
		{
			Tag:    "*exp.SetupTime",
			Path:   "*exp.SetupTime",
			Type:   "*composed",
			Value:  NewRSRParsersMustCompile("~*req.SetupTime", true, utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00",
		},
		{
			Tag:    "*exp.AnswerTime",
			Path:   "*exp.AnswerTime",
			Type:   "*composed",
			Value:  NewRSRParsersMustCompile("~*req.AnswerTime", true, utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00",
		},
		{
			Tag:   "*exp.Usage",
			Path:  "*exp.Usage",
			Type:  "*composed",
			Value: NewRSRParsersMustCompile("~*req.Usage", true, utils.INFIELD_SEP),
		},
		{
			Tag:              "*exp.Cost",
			Path:             "*exp.Cost",
			Type:             "*composed",
			Value:            NewRSRParsersMustCompile("~*req.Cost", true, utils.INFIELD_SEP),
			RoundingDecimals: utils.IntPointer(4),
		},
	}
	for _, v := range eContentFlds {
		v.ComputePath()
	}
	eCdreCfg := map[string]*CdreCfg{
		utils.MetaDefault: {
			ExportFormat:      utils.MetaFileCSV,
			ExportPath:        "/var/spool/cgrates/cdre",
			Filters:           []string{},
			Synchronous:       false,
			Attempts:          1,
			AttributeSContext: "",
			FieldSeparator:    utils.CSV_SEP,
			Fields:            eContentFlds,
		},
	}
	if !reflect.DeepEqual(cgrCfg.CdreProfiles, eCdreCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.CdreProfiles, eCdreCfg)
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
		SupplSConns:         []string{},
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
		DefaultUsage: map[string]time.Duration{
			utils.META_ANY: 3 * time.Hour,
			utils.VOICE:    3 * time.Hour,
			utils.DATA:     1048576,
			utils.SMS:      1,
		},
	}
	if !reflect.DeepEqual(eSessionSCfg, cgrCfg.sessionSCfg) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSessionSCfg), utils.ToJSON(cgrCfg.sessionSCfg))
	}
}

func TestCgrCfgJSONDefaultsCacheCFG(t *testing.T) {
	eCacheCfg := CacheCfg{
		utils.CacheReverseFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDestinations: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheReverseDestinations: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheRatingPlans: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheRatingProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActions: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActionPlans: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAccountActionPlans: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActionTriggers: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSharedGroups: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheTimings: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResourceProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResources: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheEventResources: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false},
		utils.CacheStatQueueProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheStatQueues: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholdProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholds: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheFilters: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSupplierProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAttributeProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheChargerProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDispatcherProfiles: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDispatcherHosts: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResourceFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheStatFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholdFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSupplierFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAttributeFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheChargerFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDispatcherFilterIndexes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDispatcherRoutes: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDiameterMessages: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(3 * time.Hour), StaticTTL: false},
		utils.CacheRPCResponses: &CacheParamCfg{Limit: 0,
			TTL: time.Duration(2 * time.Second), StaticTTL: false},
		utils.CacheClosedSessions: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(10 * time.Second), StaticTTL: false},
		utils.CacheCDRIDs: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(10 * time.Minute), StaticTTL: false},
		utils.CacheLoadIDs: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheRPCConnections: &CacheParamCfg{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false},
	}

	if !reflect.DeepEqual(eCacheCfg, cgrCfg.CacheCfg()) {
		t.Errorf("received: %s, \nexpecting: %s",
			utils.ToJSON(eCacheCfg), utils.ToJSON(cgrCfg.CacheCfg()))
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
	}
	if !reflect.DeepEqual(eThresholdSCfg, cgrCfg.thresholdSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eThresholdSCfg, cgrCfg.thresholdSCfg)
	}
}

func TestCgrCfgJSONDefaultSupplierSCfg(t *testing.T) {
	eSupplSCfg := &SupplierSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		RALsConns:           []string{},
		DefaultRatio:        1,
	}
	if !reflect.DeepEqual(eSupplSCfg, cgrCfg.supplierSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eSupplSCfg, cgrCfg.supplierSCfg)
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
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", true, utils.INFIELD_SEP),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP),
		BillToNumber:         NewRSRParsersMustCompile("", true, utils.INFIELD_SEP),
		Zipcode:              NewRSRParsersMustCompile("", true, utils.INFIELD_SEP),
		P2PZipcode:           NewRSRParsersMustCompile("", true, utils.INFIELD_SEP),
		P2PPlus4:             NewRSRParsersMustCompile("", true, utils.INFIELD_SEP),
		Units:                NewRSRParsersMustCompile("1", true, utils.INFIELD_SEP),
		UnitType:             NewRSRParsersMustCompile("00", true, utils.INFIELD_SEP),
		TaxIncluded:          NewRSRParsersMustCompile("0", true, utils.INFIELD_SEP),
		TaxSitusRule:         NewRSRParsersMustCompile("04", true, utils.INFIELD_SEP),
		TransTypeCode:        NewRSRParsersMustCompile("010101", true, utils.INFIELD_SEP),
		SalesTypeCode:        NewRSRParsersMustCompile("R", true, utils.INFIELD_SEP),
		TaxExemptionCodeList: NewRSRParsersMustCompile("", true, utils.INFIELD_SEP),
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

func TestDbDefaults(t *testing.T) {
	dbdf := newDbDefaults()
	flagInput := utils.MetaDynamic
	dbs := []string{utils.MetaMongo, utils.MetaRedis, utils.MetaMySQL, utils.MetaInternal}
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
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ProfileID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "Contexts",
							Path:  "Contexts",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "AttributeFilterIDs",
							Path:  "AttributeFilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "Path",
							Path:  "Path",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
						{Tag: "Type",
							Path:  "Type",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
						{Tag: "Value",
							Path:  "Value",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
						{Tag: "Blocker",
							Path:  "Blocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaFilters,
					Filename: utils.FiltersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "Type",
							Path:  "Type",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "Element",
							Path:  "Element",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "Values",
							Path:  "Values",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaResources,
					Filename: utils.ResourcesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "TTL",
							Path:  "UsageTTL",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "Limit",
							Path:  "Limit",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "AllocationMessage",
							Path:  "AllocationMessage",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
						{Tag: "Blocker",
							Path:  "Blocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
						{Tag: "Stored",
							Path:  "Stored",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
						{Tag: "ThresholdIDs",
							Path:  "ThresholdIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaStats,
					Filename: utils.StatsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "QueueLength",
							Path:  "QueueLength",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "TTL",
							Path:  "TTL",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "MinItems",
							Path:  "MinItems",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
						{Tag: "MetricIDs",
							Path:  "MetricIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
						{Tag: "MetricFilterIDs",
							Path:  "MetricFilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
						{Tag: "Blocker",
							Path:  "Blocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
						{Tag: "Stored",
							Path:  "Stored",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},

						{Tag: "ThresholdIDs",
							Path:  "ThresholdIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaThresholds,
					Filename: utils.ThresholdsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "MaxHits",
							Path:  "MaxHits",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "MinHits",
							Path:  "MinHits",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "MinSleep",
							Path:  "MinSleep",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
						{Tag: "Blocker",
							Path:  "Blocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
						{Tag: "ActionIDs",
							Path:  "ActionIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
						{Tag: "Async",
							Path:  "Async",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaSuppliers,
					Filename: utils.SuppliersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "Sorting",
							Path:  "Sorting",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "SortingParamameters",
							Path:  "SortingParamameters",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "SupplierID",
							Path:  "SupplierID",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
						{Tag: "SupplierFilterIDs",
							Path:  "SupplierFilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
						{Tag: "SupplierAccountIDs",
							Path:  "SupplierAccountIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
						{Tag: "SupplierRatingPlanIDs",
							Path:  "SupplierRatingPlanIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
						{Tag: "SupplierResourceIDs",
							Path:  "SupplierResourceIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
						{Tag: "SupplierStatIDs",
							Path:  "SupplierStatIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},
						{Tag: "SupplierWeight",
							Path:  "SupplierWeight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
						{Tag: "SupplierBlocker",
							Path:  "SupplierBlocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~13", true, utils.INFIELD_SEP)},
						{Tag: "SupplierParameters",
							Path:  "SupplierParameters",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~14", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~15", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaChargers,
					Filename: utils.ChargersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
						{Tag: "RunID",
							Path:  "RunID",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
						{Tag: "AttributeIDs",
							Path:  "AttributeIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
					},
				},
				{
					Type:     utils.MetaDispatchers,
					Filename: utils.DispatcherProfilesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "Contexts",
							Path:  "Contexts",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
						},
						{Tag: "FilterIDs",
							Path:  "FilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
						},
						{Tag: "ActivationInterval",
							Path:  "ActivationInterval",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
						},
						{Tag: "Strategy",
							Path:  "Strategy",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP),
						},
						{Tag: "StrategyParameters",
							Path:  "StrategyParameters",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP),
						},
						{Tag: "ConnID",
							Path:  "ConnID",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP),
						},
						{Tag: "ConnFilterIDs",
							Path:  "ConnFilterIDs",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP),
						},
						{Tag: "ConnWeight",
							Path:  "ConnWeight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP),
						},
						{Tag: "ConnBlocker",
							Path:  "ConnBlocker",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP),
						},
						{Tag: "ConnParameters",
							Path:  "ConnParameters",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP),
						},
						{Tag: "Weight",
							Path:  "Weight",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP),
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
							Value:     NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
							Mandatory: true},
						{Tag: "Address",
							Path:  "Address",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
						},
						{Tag: "Transport",
							Path:  "Transport",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
						},
						{Tag: "TLS",
							Path:  "TLS",
							Type:  utils.MetaVariable,
							Value: NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
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
		AttributeSConns:     []string{},
	}
	if !reflect.DeepEqual(cgrCfg.dispatcherSCfg, eDspSCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.dispatcherSCfg, eDspSCfg)
	}
}

func TestCgrLoaderCfgDefault(t *testing.T) {
	eLdrCfg := &LoaderCgrCfg{
		TpID:           "",
		DataPath:       "./",
		DisableReverse: false,
		FieldSeparator: rune(utils.CSV_SEP),
		CachesConns:    []string{utils.MetaLocalHost},
		SchedulerConns: []string{utils.MetaLocalHost},
	}
	if !reflect.DeepEqual(cgrCfg.LoaderCgrCfg(), eLdrCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.LoaderCgrCfg()), utils.ToJSON(eLdrCfg))
	}
}

func TestCgrMigratorCfgDefault(t *testing.T) {
	eMgrCfg := &MigratorCgrCfg{
		OutDataDBType:     "*redis",
		OutDataDBHost:     "127.0.0.1",
		OutDataDBPort:     "6379",
		OutDataDBName:     "10",
		OutDataDBUser:     "cgrates",
		OutDataDBPassword: "",
		OutDataDBEncoding: "msgpack",
		OutStorDBType:     "*mysql",
		OutStorDBHost:     "127.0.0.1",
		OutStorDBPort:     "3306",
		OutStorDBName:     "cgrates",
		OutStorDBUser:     "cgrates",
		OutStorDBPassword: "",
	}
	if !reflect.DeepEqual(cgrCfg.MigratorCgrCfg(), eMgrCfg) {
		t.Errorf("received: %+v, expecting: %+v", utils.ToJSON(cgrCfg.MigratorCgrCfg()), utils.ToJSON(eMgrCfg))
	}
}

func TestCgrMigratorCfg2(t *testing.T) {
	JSN_CFG := `
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

	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if cgrCfg.MigratorCgrCfg().OutDataDBHost != "0.0.0.0" {
		t.Errorf("Expected: 0.0.0.0 , received: %+v", cgrCfg.MigratorCgrCfg().OutDataDBHost)
	} else if cgrCfg.MigratorCgrCfg().OutDataDBPort != "9999" {
		t.Errorf("Expected: 9999, received: %+v", cgrCfg.MigratorCgrCfg().OutDataDBPassword)
	}
}

func TestCfgTlsCfg(t *testing.T) {
	JSN_CFG := `
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.TlsCfg(), cgrCfg.TlsCfg()) {
		t.Errorf("Expected: %s, received: %s",
			utils.ToJSON(eCgrCfg.tlsCfg), utils.ToJSON(cgrCfg.tlsCfg))
	}
}

func TestCgrCfgJSONDefaultAnalyzerSCfg(t *testing.T) {
	aSCfg := &AnalyzerSCfg{
		Enabled: false,
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
	}
	if !reflect.DeepEqual(cgrCfg.apier, aCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.apier, aCfg)
	}
}

func TestCgrCfgV1GetConfigSection(t *testing.T) {
	JSN_CFG := `
{
"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080",
	}
}`
	expected := map[string]any{
		"HTTPListen":       ":2080",
		"HTTPTLSListen":    "127.0.0.1:2280",
		"RPCGOBListen":     ":2013",
		"RPCGOBTLSListen":  "127.0.0.1:2023",
		"RPCJSONListen":    ":2012",
		"RPCJSONTLSListen": "127.0.0.1:2022",
	}
	var rcv map[string]any
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if err := cgrCfg.V1GetConfigSection(&StringWithArgDispatcher{Section: LISTEN_JSN}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", expected, rcv)
	}
}

func TestCgrCdfEventReader(t *testing.T) {
	eCfg := &ERsCfg{
		Enabled:       false,
		SessionSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		Readers: []*EventReaderCfg{
			{
				ID:             utils.MetaDefault,
				Type:           utils.META_NONE,
				FieldSep:       ",",
				RunDelay:       time.Duration(0),
				ConcurrentReqs: 1024,
				SourcePath:     "/var/spool/cgrates/ers/in",
				ProcessedPath:  "/var/spool/cgrates/ers/out",
				XmlRootPath:    nil,
				Tenant:         nil,
				Timezone:       utils.EmptyString,
				Filters:        []string{},
				Flags:          utils.FlagsWithParams{},
				Fields: []*FCTemplate{
					{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
					{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
						Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
				},
				CacheDumpFields: []*FCTemplate{},
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

func TestCgrCfgEventReaderDefault(t *testing.T) {
	eCfg := &EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.META_NONE,
		FieldSep:       ",",
		RunDelay:       time.Duration(0),
		ConcurrentReqs: 1024,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		XmlRootPath:    nil,
		Tenant:         nil,
		Timezone:       utils.EmptyString,
		Filters:        nil,
		Flags:          utils.FlagsWithParams{},
		Fields: []*FCTemplate{
			{Tag: utils.ToR, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.2", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.OriginID, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.3", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.RequestType, Path: utils.MetaCgreq + utils.NestingSep + utils.RequestType, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.4", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Tenant, Path: utils.MetaCgreq + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.6", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Category, Path: utils.MetaCgreq + utils.NestingSep + utils.Category, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.7", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Account, Path: utils.MetaCgreq + utils.NestingSep + utils.Account, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.8", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Subject, Path: utils.MetaCgreq + utils.NestingSep + utils.Subject, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.9", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Destination, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.10", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.SetupTime, Path: utils.MetaCgreq + utils.NestingSep + utils.SetupTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.11", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.AnswerTime, Path: utils.MetaCgreq + utils.NestingSep + utils.AnswerTime, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.12", true, utils.INFIELD_SEP), Mandatory: true},
			{Tag: utils.Usage, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage, Type: utils.MetaVariable,
				Value: NewRSRParsersMustCompile("~*req.13", true, utils.INFIELD_SEP), Mandatory: true},
		},
		CacheDumpFields: make([]*FCTemplate, 0),
	}
	for _, v := range eCfg.Fields {
		v.ComputePath()
	}
	if !reflect.DeepEqual(cgrCfg.dfltEvRdr, eCfg) {
		t.Errorf("received: %+v,\n expecting: %+v", utils.ToJSON(cgrCfg.dfltEvRdr), utils.ToJSON(eCfg))
	}

}

func TestRpcConnsDefaults(t *testing.T) {
	eCfg := make(map[string]*RPCConn)
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

func TestCheckConfigSanity(t *testing.T) {
	// Rater checks
	cfg, _ := NewDefaultCGRConfig()
	cfg.ralsCfg = &RalsCfg{
		Enabled:    true,
		StatSConns: []string{utils.MetaInternal},
	}
	expected := "<StatS> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true
	cfg.ralsCfg.ThresholdSConns = []string{utils.MetaInternal}

	expected = "<ThresholdS> not enabled but requested by <RALs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg = &RalsCfg{
		Enabled:         false,
		StatSConns:      []string{},
		ThresholdSConns: []string{},
	}
	// CDRServer checks
	cfg.thresholdSCfg.Enabled = true
	cfg.cdrsCfg = &CdrsCfg{
		Enabled:       true,
		ChargerSConns: []string{utils.MetaInternal},
	}
	expected = "<ChargerS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.chargerSCfg.Enabled = true
	cfg.cdrsCfg.RaterConns = []string{utils.MetaInternal}

	expected = "<RALs> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.ralsCfg.Enabled = true
	cfg.cdrsCfg.AttributeSConns = []string{utils.MetaInternal}
	expected = "<AttributeS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = false
	cfg.attributeSCfg.Enabled = true
	cfg.cdrsCfg.StatSConns = []string{utils.MetaInternal}
	expected = "<StatS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.statsCfg.Enabled = true
	cfg.cdrsCfg.OnlineCDRExports = []string{"stringy"}
	cfg.CdreProfiles = map[string]*CdreCfg{"stringx": {}}
	expected = "<CDRs> cannot find CDR export template with ID: <stringy>"
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
	cfg.thresholdSCfg.Enabled = false
	cfg.cdrsCfg.OnlineCDRExports = []string{"stringx"}
	cfg.cdrsCfg.ThresholdSConns = []string{utils.MetaInternal}
	expected = "<ThresholdS> not enabled but requested by <CDRs> component."
	if err := cfg.checkConfigSanity(); err == nil || err.Error() != expected {
		t.Errorf("Expecting: %+q  received: %+q", expected, err)
	}
}

func TestGeneralCfg(t *testing.T) {
	var gencfg GeneralCfg
	cfgJSONStr := `{
		"general": {
			"node_id": "",
			"logger":"*syslog",
			"log_level": 6,
			"http_skip_tls_verify": false,
			"rounding_decimals": 5,
			"dbdata_encoding": "*msgpack",
			"tpexport_dir": "/var/spool/cgrates/tpe",
			"poster_attempts": 3,
			"failed_posts_dir": "/var/spool/cgrates/failed_posts",
			"failed_posts_ttl": "5s",
			"default_request_type": "*rated",
			"default_category": "call",
			"default_tenant": "cgrates.org",
			"default_timezone": "Local",
			"default_caching":"*reload",
			"connect_attempts": 5,
			"reconnects": -1,
			"connect_timeout": "1s",
			"reply_timeout": "2s",
			"locking_timeout": "0",
			"digest_separator": ",",
			"digest_equal": ":",
			"rsr_separator": ";",
			"max_parallel_conns": 100,
			"concurrent_requests":  0,
			"concurrent_strategy":  "",
		},
}`
	eMap := map[string]any{
		"node_id":              "",
		"logger":               "*syslog",
		"log_level":            6,
		"http_skip_tls_verify": false,
		"rounding_decimals":    5,
		"dbdata_encoding":      "*msgpack",
		"tpexport_dir":         "/var/spool/cgrates/tpe",
		"poster_attempts":      3,
		"failed_posts_dir":     "/var/spool/cgrates/failed_posts",
		"failed_posts_ttl":     "5s",
		"default_request_type": "*rated",
		"default_category":     "call",
		"default_tenant":       "cgrates.org",
		"default_timezone":     "Local",
		"default_caching":      "*reload",
		"connect_attempts":     5,
		"reconnects":           -1,
		"connect_timeout":      "1s",
		"reply_timeout":        "2s",
		"locking_timeout":      "0",
		"digest_separator":     ",",
		"digest_equal":         ":",
		"rsr_separator":        ";",
		"max_parallel_conns":   100,
		"concurrent_requests":  0,
		"concurrent_strategy":  "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnGenCfg, err := jsnCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if err = gencfg.loadFromJsonCfg(jsnGenCfg); err != nil {
		t.Error(err)
	} else if rcv := gencfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestConfigDBFields(t *testing.T) {
	d := dbDefaults{}
	str := "test"

	tests := []struct {
		name string
		rcv  any
		exp  any
	}{
		{
			name: "dbName diff from dynamic",
			rcv:  d.dbName(str, str),
			exp:  str,
		},
		{
			name: "dbUser diff from dynamic",
			rcv:  d.dbUser(str, str),
			exp:  str,
		},
		{
			name: "dbHost diff from dynamic",
			rcv:  d.dbHost(str, str),
			exp:  str,
		},
		{
			name: "dbPass diff from dynamic",
			rcv:  d.dbPass(str, str),
			exp:  str,
		},
		{
			name: "isHidden true",
			rcv:  isHidden(".test"),
			exp:  true,
		},
		{
			name: "isHidden false",
			rcv:  isHidden("."),
			exp:  false,
		},
	}

	for _, tt := range tests {
		if tt.rcv != tt.exp {
			t.Errorf("Expected: %+v\nReceived: %+v", tt.exp, tt.rcv)
		}
	}
}

func TestConfigLoadErrors(t *testing.T) {
	c := CGRConfig{}
	js := json.RawMessage([]byte(`test`))

	tests := []struct {
		name string
		rcv  error
		exp  string
	}{
		{
			name: "load general cfg error check",
			rcv:  c.loadGeneralCfg(&CgrJsonCfg{GENERAL_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load cache cfg error check",
			rcv:  c.loadCacheCfg(&CgrJsonCfg{CACHE_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load listen cfg error check",
			rcv:  c.loadListenCfg(&CgrJsonCfg{LISTEN_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load HTTP cfg error check",
			rcv:  c.loadHTTPCfg(&CgrJsonCfg{HTTP_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load StorDB cfg error check",
			rcv:  c.loadStorDBCfg(&CgrJsonCfg{STORDB_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load FilterS cfg error check",
			rcv:  c.loadFilterSCfg(&CgrJsonCfg{FilterSjsn: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load RalS cfg error check",
			rcv:  c.loadRalSCfg(&CgrJsonCfg{RALS_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load Scheduler cfg error check",
			rcv:  c.loadSchedulerCfg(&CgrJsonCfg{SCHEDULER_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load drs cfg error check",
			rcv:  c.loadCdrsCfg(&CgrJsonCfg{CDRS_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load SessionS cfg error check",
			rcv:  c.loadSessionSCfg(&CgrJsonCfg{SessionSJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load KamAgent cfg error check",
			rcv:  c.loadKamAgentCfg(&CgrJsonCfg{KamailioAgentJSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load FreeswitchAgent cfg error check",
			rcv:  c.loadFreeswitchAgentCfg(&CgrJsonCfg{FreeSWITCHAgentJSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load AsteriskAgent cfg error check",
			rcv:  c.loadAsteriskAgentCfg(&CgrJsonCfg{AsteriskAgentJSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load DiameterAgent cfg error check",
			rcv:  c.loadDiameterAgentCfg(&CgrJsonCfg{DA_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load RadiusAgent cfg error check",
			rcv:  c.loadRadiusAgentCfg(&CgrJsonCfg{RA_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load DNSAgent cfg error check",
			rcv:  c.loadDNSAgentCfg(&CgrJsonCfg{DNSAgentJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load HttpAgent cfg error check",
			rcv:  c.loadHttpAgentCfg(&CgrJsonCfg{HttpAgentJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load AttributeS cfg error check",
			rcv:  c.loadAttributeSCfg(&CgrJsonCfg{ATTRIBUTE_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load ChargerS cfg error check",
			rcv:  c.loadChargerSCfg(&CgrJsonCfg{ChargerSCfgJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load ResourceS cfg error check",
			rcv:  c.loadResourceSCfg(&CgrJsonCfg{RESOURCES_JSON: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load StatS cfg error check",
			rcv:  c.loadStatSCfg(&CgrJsonCfg{STATS_JSON: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load ThresholdS cfg error check",
			rcv:  c.loadThresholdSCfg(&CgrJsonCfg{THRESHOLDS_JSON: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load SupplierS cfg error check",
			rcv:  c.loadSupplierSCfg(&CgrJsonCfg{SupplierSJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load Mailer cfg error check",
			rcv:  c.loadMailerCfg(&CgrJsonCfg{MAILER_JSN: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load SureTax cfg error check",
			rcv:  c.loadSureTaxCfg(&CgrJsonCfg{SURETAX_JSON: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load DispatcherS cfg error check",
			rcv:  c.loadDispatcherSCfg(&CgrJsonCfg{DispatcherSJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load MigratorCgr cfg error check",
			rcv:  c.loadMigratorCgrCfg(&CgrJsonCfg{CgrMigratorCfgJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load TlsCgr cfg error check",
			rcv:  c.loadTlsCgrCfg(&CgrJsonCfg{TlsCfgJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load AnalyzerCgr cfg error check",
			rcv:  c.loadAnalyzerCgrCfg(&CgrJsonCfg{AnalyzerCfgJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load Apier cfg error check",
			rcv:  c.loadApierCfg(&CgrJsonCfg{ApierS: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load Apier cfg error check",
			rcv:  c.loadApierCfg(&CgrJsonCfg{ApierS: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "load Ers cfg error check",
			rcv:  c.loadErsCfg(&CgrJsonCfg{ERsJson: &js}),
			exp:  "invalid character 'e' in literal true (expecting 'r')",
		},
	}

	for _, tt := range tests {
		if tt.rcv != nil {
			if tt.rcv.Error() != tt.exp {
				t.Errorf("Expected: %+v\nReceived: %+v", tt.exp, tt.rcv)
			}
		} else {
			t.Error("was expecting an error")
		}
	}
}

func TestConfigSureTaxCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{SURETAX_JSON: {}},
	}

	rcv := cfg.SureTaxCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigDiameterAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{DA_JSN: {}},
	}

	rcv := cfg.DiameterAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigRadiusAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{RA_JSN: {}},
	}

	rcv := cfg.RadiusAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigDNSAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{DNSAgentJson: {}},
	}

	rcv := cfg.DNSAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigAttributeSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{ATTRIBUTE_JSN: {}},
	}

	rcv := cfg.AttributeSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigChargerSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{ChargerSCfgJson: {}},
	}

	rcv := cfg.ChargerSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigResourceSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{RESOURCES_JSON: {}},
	}

	rcv := cfg.ResourceSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigStatSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{STATS_JSON: {}},
	}

	rcv := cfg.StatSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigThresholdSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{THRESHOLDS_JSON: {}},
	}

	rcv := cfg.ThresholdSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigSupplierSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{SupplierSJson: {}},
	}

	rcv := cfg.SupplierSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigSessionSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{SessionSJson: {}},
	}

	rcv := cfg.SessionSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigFsAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{FreeSWITCHAgentJSN: {}},
	}

	rcv := cfg.FsAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigKamAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{KamailioAgentJSN: {}},
	}

	rcv := cfg.KamAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigAsteriskAgentCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{AsteriskAgentJSN: {}},
	}

	rcv := cfg.AsteriskAgentCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigFilterSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{FilterSjsn: {}},
	}

	rcv := cfg.FilterSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigLoaderCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{LoaderJson: {}},
	}

	rcv := cfg.LoaderCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigDispatcherSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{DispatcherSJson: {}},
	}

	rcv := cfg.DispatcherSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigSchedulerCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{SCHEDULER_JSN: {}},
	}

	rcv := cfg.SchedulerCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigAnalyzerSCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{AnalyzerCfgJson: {}},
	}

	rcv := cfg.AnalyzerSCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigApierCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{ApierS: {}},
	}

	rcv := cfg.ApierCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigERsCfg(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{ERsJson: {}},
	}

	rcv := cfg.ERsCfg()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigRPCConns(t *testing.T) {
	cfg := &CGRConfig{
		lks: map[string]*sync.RWMutex{RPCConnsJsonName: {}},
	}

	rcv := cfg.RPCConns()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestConfigV1GetConfigSection(t *testing.T) {
	JSN_CFG := `
	{
	"freeswitch_agent": {
		"enabled": true,				// starts SessionManager service: <true|false>
		"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
			{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3, "alias":"123"},
			{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5, "alias":"124"}
		],
	},
	
	}`
	cfg, _ := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG)
	tests := []struct {
		name string
		args *StringWithArgDispatcher
		exp  string
		err  string
	}{
		{
			name: "case GENERAL_JSN",
			args: &StringWithArgDispatcher{Section: GENERAL_JSN},
			exp:  utils.ToJSON(cfg.GeneralCfg()),
			err:  "",
		},
		{
			name: "case DATADB_JSN",
			args: &StringWithArgDispatcher{Section: DATADB_JSN},
			exp:  utils.ToJSON(cfg.DataDbCfg()),
			err:  "",
		},
		{
			name: "case STORDB_JSN",
			args: &StringWithArgDispatcher{Section: STORDB_JSN},
			exp:  utils.ToJSON(cfg.StorDbCfg()),
			err:  "",
		},
		{
			name: "case TlsCfgJson",
			args: &StringWithArgDispatcher{Section: TlsCfgJson},
			exp:  utils.ToJSON(cfg.TlsCfg()),
			err:  "",
		},
		{
			name: "case CACHE_JSN",
			args: &StringWithArgDispatcher{Section: CACHE_JSN},
			exp:  utils.ToJSON(cfg.CacheCfg()),
			err:  "",
		},
		{
			name: "case HTTP_JSN",
			args: &StringWithArgDispatcher{Section: HTTP_JSN},
			exp:  utils.ToJSON(cfg.HTTPCfg()),
			err:  "",
		},
		{
			name: "case FilterSjsn",
			args: &StringWithArgDispatcher{Section: FilterSjsn},
			exp:  utils.ToJSON(cfg.FilterSCfg()),
			err:  "",
		},
		{
			name: "case RALS_JSN",
			args: &StringWithArgDispatcher{Section: RALS_JSN},
			exp:  utils.ToJSON(cfg.RalsCfg()),
			err:  "",
		},
		{
			name: "case SCHEDULER_JSN",
			args: &StringWithArgDispatcher{Section: SCHEDULER_JSN},
			exp:  utils.ToJSON(cfg.SchedulerCfg()),
			err:  "",
		},
		{
			name: "case CDRS_JSN",
			args: &StringWithArgDispatcher{Section: CDRS_JSN},
			exp:  utils.ToJSON(cfg.CdrsCfg()),
			err:  "",
		},
		{
			name: "case SessionSJson",
			args: &StringWithArgDispatcher{Section: SessionSJson},
			exp:  utils.ToJSON(cfg.SessionSCfg()),
			err:  "",
		},
		{
			name: "case FreeSWITCHAgentJSN",
			args: &StringWithArgDispatcher{Section: FreeSWITCHAgentJSN},
			exp:  utils.ToJSON(cfg.FsAgentCfg()),
			err:  "",
		},
		{
			name: "case KamailioAgentJSN",
			args: &StringWithArgDispatcher{Section: KamailioAgentJSN},
			exp:  utils.ToJSON(cfg.KamAgentCfg()),
			err:  "",
		},
		{
			name: "case AsteriskAgentJSN",
			args: &StringWithArgDispatcher{Section: AsteriskAgentJSN},
			exp:  utils.ToJSON(cfg.AsteriskAgentCfg()),
			err:  "",
		},
		{
			name: "case DA_JSN",
			args: &StringWithArgDispatcher{Section: DA_JSN},
			exp:  utils.ToJSON(cfg.DiameterAgentCfg()),
			err:  "",
		},
		{
			name: "case RA_JSN",
			args: &StringWithArgDispatcher{Section: RA_JSN},
			exp:  utils.ToJSON(cfg.RadiusAgentCfg()),
			err:  "",
		},
		{
			name: "case DNSAgentJson",
			args: &StringWithArgDispatcher{Section: DNSAgentJson},
			exp:  utils.ToJSON(cfg.DNSAgentCfg()),
			err:  "",
		},
		{
			name: "case ATTRIBUTE_JSN",
			args: &StringWithArgDispatcher{Section: ATTRIBUTE_JSN},
			exp:  utils.ToJSON(cfg.AttributeSCfg()),
			err:  "",
		},
		{
			name: "case ChargerSCfgJson",
			args: &StringWithArgDispatcher{Section: ChargerSCfgJson},
			exp:  utils.ToJSON(cfg.ChargerSCfg()),
			err:  "",
		},
		{
			name: "case RESOURCES_JSON",
			args: &StringWithArgDispatcher{Section: RESOURCES_JSON},
			exp:  utils.ToJSON(cfg.ResourceSCfg()),
			err:  "",
		},
		{
			name: "case STATS_JSON",
			args: &StringWithArgDispatcher{Section: STATS_JSON},
			exp:  utils.ToJSON(cfg.StatSCfg()),
			err:  "",
		},
		{
			name: "case THRESHOLDS_JSON",
			args: &StringWithArgDispatcher{Section: THRESHOLDS_JSON},
			exp:  utils.ToJSON(cfg.ThresholdSCfg()),
			err:  "",
		},
		{
			name: "case SupplierSJson",
			args: &StringWithArgDispatcher{Section: SupplierSJson},
			exp:  utils.ToJSON(cfg.SupplierSCfg()),
			err:  "",
		},
		{
			name: "case SURETAX_JSON",
			args: &StringWithArgDispatcher{Section: SURETAX_JSON},
			exp:  utils.ToJSON(cfg.SureTaxCfg()),
			err:  "",
		},
		{
			name: "case DispatcherSJson",
			args: &StringWithArgDispatcher{Section: DispatcherSJson},
			exp:  utils.ToJSON(cfg.DispatcherSCfg()),
			err:  "",
		},
		{
			name: "case LoaderJson",
			args: &StringWithArgDispatcher{Section: LoaderJson},
			exp:  utils.ToJSON(cfg.LoaderCfg()),
			err:  "",
		},
		{
			name: "case CgrLoaderCfgJson",
			args: &StringWithArgDispatcher{Section: CgrLoaderCfgJson},
			exp:  utils.ToJSON(cfg.LoaderCgrCfg()),
			err:  "",
		},
		{
			name: "case CgrMigratorCfgJson",
			args: &StringWithArgDispatcher{Section: CgrMigratorCfgJson},
			exp:  utils.ToJSON(cfg.MigratorCgrCfg()),
			err:  "",
		},
		{
			name: "case ApierS",
			args: &StringWithArgDispatcher{Section: ApierS},
			exp:  utils.ToJSON(cfg.ApierCfg()),
			err:  "",
		},
		{
			name: "case CDRE_JSN",
			args: &StringWithArgDispatcher{Section: CDRE_JSN},
			exp:  utils.ToJSON(cfg.CdreProfiles),
			err:  "",
		},
		{
			name: "case ERsJson",
			args: &StringWithArgDispatcher{Section: ERsJson},
			exp:  utils.ToJSON(cfg.ERsCfg()),
			err:  "",
		},
		{
			name: "case RPCConnsJsonName",
			args: &StringWithArgDispatcher{Section: RPCConnsJsonName},
			exp:  utils.ToJSON(cfg.RPCConns()),
			err:  "",
		},
		{
			name: "default",
			args: &StringWithArgDispatcher{Section: "test"},
			exp:  "",
			err:  "Invalid section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := &map[string]any{}
			exp := &map[string]any{}
			json.Unmarshal([]byte(tt.exp), exp)

			err := cfg.V1GetConfigSection(tt.args, rcv)
			if tt.err != "" {
				if err != nil {
					if err.Error() != tt.err {
						t.Error(err)
					}
				}
			}

			if !reflect.DeepEqual(rcv, exp) {
				t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
			}
		})
	}
}

func TestConfigAsMapInterface(t *testing.T) {
	separator := ""
	cfg := CGRConfig{
		lks:             map[string]*sync.RWMutex{},
		MaxCallDuration: 1 * time.Second,
		DataFolderPath:  "test",
		ConfigPath:      "test",
		dfltCdreProfile: &CdreCfg{},
		dfltEvRdr:       &EventReaderCfg{},
		CdreProfiles: map[string]*CdreCfg{"test": {
			ExportFormat:      "test",
			ExportPath:        "test",
			Filters:           []string{"test"},
			Tenant:            "test",
			AttributeSContext: "test",
			Synchronous:       false,
			Attempts:          1,
			FieldSeparator:    'a',
			Fields:            []*FCTemplate{},
		}},
		loaderCfg: LoaderSCfgs{{
			Id:             "test",
			Enabled:        true,
			Tenant:         RSRParsers{},
			DryRun:         false,
			RunDelay:       1 * time.Second,
			LockFileName:   "test",
			CacheSConns:    []string{"test"},
			FieldSeparator: "test",
			TpInDir:        "test",
			TpOutDir:       "test",
			Data:           []*LoaderDataType{},
		}},
		httpAgentCfg: HttpAgentCfgs{{
			ID:                "test",
			Url:               "test",
			SessionSConns:     []string{"test"},
			RequestPayload:    "test",
			ReplyPayload:      "test",
			RequestProcessors: []*RequestProcessor{},
		}},
		ConfigReloads: map[string]chan struct{}{},
		rldChans:      map[string]chan struct{}{},
		rpcConns: map[string]*RPCConn{"test": {
			Strategy: "test",
			PoolSize: 1,
			Conns:    []*RemoteHost{},
		}},
		generalCfg:       &GeneralCfg{},
		dataDbCfg:        &DataDbCfg{},
		storDbCfg:        &StorDbCfg{},
		tlsCfg:           &TlsCfg{},
		cacheCfg:         CacheCfg{},
		listenCfg:        &ListenCfg{},
		httpCfg:          &HTTPCfg{},
		filterSCfg:       &FilterSCfg{},
		ralsCfg:          &RalsCfg{},
		schedulerCfg:     &SchedulerCfg{},
		cdrsCfg:          &CdrsCfg{},
		sessionSCfg:      &SessionSCfg{},
		fsAgentCfg:       &FsAgentCfg{},
		kamAgentCfg:      &KamAgentCfg{},
		asteriskAgentCfg: &AsteriskAgentCfg{},
		diameterAgentCfg: &DiameterAgentCfg{},
		radiusAgentCfg:   &RadiusAgentCfg{},
		dnsAgentCfg:      &DNSAgentCfg{},
		attributeSCfg:    &AttributeSCfg{},
		chargerSCfg:      &ChargerSCfg{},
		resourceSCfg:     &ResourceSConfig{},
		statsCfg:         &StatSCfg{},
		thresholdSCfg:    &ThresholdSCfg{},
		supplierSCfg:     &SupplierSCfg{},
		sureTaxCfg:       &SureTaxCfg{},
		dispatcherSCfg:   &DispatcherSCfg{},
		loaderCgrCfg: &LoaderCgrCfg{
			TpID:           "test",
			DataPath:       "test",
			DisableReverse: false,
			FieldSeparator: 'a',
			CachesConns:    []string{},
			SchedulerConns: []string{},
		},
		migratorCgrCfg: &MigratorCgrCfg{},
		mailerCfg:      &MailerCfg{},
		analyzerSCfg:   &AnalyzerSCfg{},
		apier:          &ApierCfg{},
		ersCfg:         &ERsCfg{},
	}

	rcv := cfg.AsMapInterface(separator)
	rpcConns := make(map[string]map[string]any, len(cfg.rpcConns))
	for key, val := range cfg.rpcConns {
		rpcConns[key] = val.AsMapInterface()
	}

	cdreProfiles := make(map[string]map[string]any)
	for key, val := range cfg.CdreProfiles {
		cdreProfiles[key] = val.AsMapInterface(separator)
	}

	loaderCfg := make([]map[string]any, len(cfg.loaderCfg))
	for i, item := range cfg.loaderCfg {
		loaderCfg[i] = item.AsMapInterface(separator)
	}

	httpAgentCfg := make([]map[string]any, len(cfg.httpAgentCfg))
	for i, item := range cfg.httpAgentCfg {
		httpAgentCfg[i] = item.AsMapInterface(separator)
	}
	exp := map[string]any{

		utils.CdreProfiles:     cdreProfiles,
		utils.LoaderCfg:        loaderCfg,
		utils.HttpAgentCfg:     httpAgentCfg,
		utils.RpcConns:         rpcConns,
		utils.GeneralCfg:       cfg.generalCfg.AsMapInterface(),
		utils.DataDbCfg:        cfg.dataDbCfg.AsMapInterface(),
		utils.StorDbCfg:        cfg.storDbCfg.AsMapInterface(),
		utils.TlsCfg:           cfg.tlsCfg.AsMapInterface(),
		utils.CacheCfg:         cfg.cacheCfg.AsMapInterface(),
		utils.ListenCfg:        cfg.listenCfg.AsMapInterface(),
		utils.HttpCfg:          cfg.httpCfg.AsMapInterface(),
		utils.FilterSCfg:       cfg.filterSCfg.AsMapInterface(),
		utils.RalsCfg:          cfg.ralsCfg.AsMapInterface(),
		utils.SchedulerCfg:     cfg.schedulerCfg.AsMapInterface(),
		utils.CdrsCfg:          cfg.cdrsCfg.AsMapInterface(),
		utils.SessionSCfg:      cfg.sessionSCfg.AsMapInterface(),
		utils.FsAgentCfg:       cfg.fsAgentCfg.AsMapInterface(separator),
		utils.KamAgentCfg:      cfg.kamAgentCfg.AsMapInterface(),
		utils.AsteriskAgentCfg: cfg.asteriskAgentCfg.AsMapInterface(),
		utils.DiameterAgentCfg: cfg.diameterAgentCfg.AsMapInterface(separator),
		utils.RadiusAgentCfg:   cfg.radiusAgentCfg.AsMapInterface(separator),
		utils.DnsAgentCfg:      cfg.dnsAgentCfg.AsMapInterface(separator),
		utils.AttributeSCfg:    cfg.attributeSCfg.AsMapInterface(),
		utils.ChargerSCfg:      cfg.chargerSCfg.AsMapInterface(),
		utils.ResourceSCfg:     cfg.resourceSCfg.AsMapInterface(),
		utils.StatsCfg:         cfg.statsCfg.AsMapInterface(),
		utils.ThresholdSCfg:    cfg.thresholdSCfg.AsMapInterface(),
		utils.SupplierSCfg:     cfg.supplierSCfg.AsMapInterface(),
		utils.SureTaxCfg:       cfg.sureTaxCfg.AsMapInterface(separator),
		utils.DispatcherSCfg:   cfg.dispatcherSCfg.AsMapInterface(),
		utils.LoaderCgrCfg:     cfg.loaderCgrCfg.AsMapInterface(),
		utils.MigratorCgrCfg:   cfg.migratorCgrCfg.AsMapInterface(),
		utils.MailerCfg:        cfg.mailerCfg.AsMapInterface(),
		utils.AnalyzerSCfg:     cfg.analyzerSCfg.AsMapInterface(),
		utils.Apier:            cfg.apier.AsMapInterface(),
		utils.ErsCfg:           cfg.ersCfg.AsMapInterface(separator),
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestConfigGetLoadFunctionsT(t *testing.T) {
	cfg := &CGRConfig{}
	funcMap := cfg.getLoadFunctions()
	expectedKeys := []string{
		GENERAL_JSN, DATADB_JSN, STORDB_JSN, LISTEN_JSN, TlsCfgJson, HTTP_JSN, SCHEDULER_JSN,
		CACHE_JSN, FilterSjsn, RALS_JSN, CDRS_JSN, CDRE_JSN, ERsJson, SessionSJson,
		AsteriskAgentJSN, FreeSWITCHAgentJSN, KamailioAgentJSN, DA_JSN, RA_JSN, HttpAgentJson,
		DNSAgentJson, ATTRIBUTE_JSN, ChargerSCfgJson, RESOURCES_JSON, STATS_JSON, THRESHOLDS_JSON,
		SupplierSJson, LoaderJson, MAILER_JSN, SURETAX_JSON, CgrLoaderCfgJson, CgrMigratorCfgJson,
		DispatcherSJson, AnalyzerCfgJson, ApierS, RPCConnsJsonName,
	}
	if len(funcMap) != len(expectedKeys) {
		t.Errorf("Unexpected number of functions in the map. Expected: %d, Got: %d", len(expectedKeys), len(funcMap))
	}
	for _, key := range expectedKeys {
		if _, ok := funcMap[key]; !ok {
			t.Errorf("Function for key %s not found in the map", key)
		}
	}

}
