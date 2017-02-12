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
"sm_freeswitch": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3},
		{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5}
	],
},

}`
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.SmFsConfig.Enabled = true
	eCgrCfg.SmFsConfig.EventSocketConns = []*FsConnConfig{
		&FsConnConfig{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3},
		&FsConnConfig{Address: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.SmFsConfig, cgrCfg.SmFsConfig)
	}
}

func TestCgrCfgCDRC(t *testing.T) {
	JSN_RAW_CFG := `
{
"cdrc": [
	{
		"id": "*default",
		"enabled": true,							// enable CDR client functionality
		"content_fields":[							// import template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
			{"field_id": "ToR", "type": "*composed", "value": "~7:s/^(voice|data|sms|mms|generic)$/*$1/"},
			{"field_id": "AnswerTime", "type": "*composed", "value": "1"},
			{"field_id": "Usage", "type": "*composed", "value": "~9:s/^(\\d+)$/${1}s/"},
		],
	},
],
}`
	eCgrCfg, _ := NewDefaultCGRConfig()
	eCgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"] = []*CdrcConfig{
		&CdrcConfig{
			ID:                       utils.META_DEFAULT,
			Enabled:                  true,
			DryRun:                   false,
			CdrsConns:                []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
			CdrFormat:                "csv",
			FieldSeparator:           rune(','),
			DataUsageMultiplyFactor:  1024,
			Timezone:                 "",
			RunDelay:                 0,
			MaxOpenFiles:             1024,
			CdrInDir:                 "/var/spool/cgrates/cdrc/in",
			CdrOutDir:                "/var/spool/cgrates/cdrc/out",
			FailedCallsPrefix:        "missed_calls",
			CDRPath:                  utils.HierarchyPath([]string{""}),
			CdrSourceId:              "freeswitch_csv",
			ContinueOnSuccess:        false,
			PartialRecordCache:       time.Duration(10 * time.Second),
			PartialCacheExpiryAction: "*dump_to_file",
			HeaderFields:             make([]*CfgCdrField, 0),
			ContentFields: []*CfgCdrField{
				&CfgCdrField{FieldId: "ToR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|mms|generic)$/*$1/", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("~9:s/^(\\d+)$/${1}s/", utils.INFIELD_SEP)},
			},
			TrailerFields: make([]*CfgCdrField, 0),
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.REQTYPE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DIRECTION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.CATEGORY, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SUBJECT, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.DESTINATION, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.SETUP_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.ANSWER_TIME, utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.USAGE, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_RAW_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.CdrcProfiles, cgrCfg.CdrcProfiles) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0], cgrCfg.CdrcProfiles["/var/spool/cgrates/cdrc/in"][0])
	}
}

func TestCgrCfgLoadJSONDefaults(t *testing.T) {
	cgrCfg, err = NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
}

func TestCgrCfgJSONDefaultsGeneral(t *testing.T) {
	if cgrCfg.HttpSkipTlsVerify != false {
		t.Error(cgrCfg.HttpSkipTlsVerify)
	}
	if cgrCfg.RoundingDecimals != 5 {
		t.Error(cgrCfg.RoundingDecimals)
	}
	if cgrCfg.DBDataEncoding != "msgpack" {
		t.Error(cgrCfg.DBDataEncoding)
	}
	if cgrCfg.TpExportPath != "/var/spool/cgrates/tpe" {
		t.Error(cgrCfg.TpExportPath)
	}
	if cgrCfg.PosterAttempts != 3 {
		t.Error(cgrCfg.PosterAttempts)
	}
	if cgrCfg.FailedPostsDir != "/var/spool/cgrates/failed_posts" {
		t.Error(cgrCfg.FailedPostsDir)
	}
	if cgrCfg.DefaultReqType != "*rated" {
		t.Error(cgrCfg.DefaultReqType)
	}
	if cgrCfg.DefaultCategory != "call" {
		t.Error(cgrCfg.DefaultCategory)
	}
	if cgrCfg.DefaultTenant != "cgrates.org" {
		t.Error(cgrCfg.DefaultTenant)
	}
	if cgrCfg.DefaultTimezone != "Local" {
		t.Error(cgrCfg.DefaultTimezone)
	}
	if cgrCfg.ConnectAttempts != 3 {
		t.Error(cgrCfg.ConnectAttempts)
	}
	if cgrCfg.Reconnects != -1 {
		t.Error(cgrCfg.Reconnects)
	}
	if cgrCfg.ConnectTimeout != 1*time.Second {
		t.Error(cgrCfg.ConnectTimeout)
	}
	if cgrCfg.ReplyTimeout != 2*time.Second {
		t.Error(cgrCfg.ReplyTimeout)
	}
	if cgrCfg.ResponseCacheTTL != 0*time.Second {
		t.Error(cgrCfg.ResponseCacheTTL)
	}
	if cgrCfg.InternalTtl != 2*time.Minute {
		t.Error(cgrCfg.InternalTtl)
	}
	if cgrCfg.LockingTimeout != 5*time.Second {
		t.Error(cgrCfg.LockingTimeout)
	}
	if cgrCfg.LogLevel != 6 {
		t.Error(cgrCfg.LogLevel)
	}
}

func TestCgrCfgJSONDefaultsListen(t *testing.T) {
	if cgrCfg.RPCJSONListen != "127.0.0.1:2012" {
		t.Error(cgrCfg.RPCJSONListen)
	}
	if cgrCfg.RPCGOBListen != "127.0.0.1:2013" {
		t.Error(cgrCfg.RPCGOBListen)
	}
	if cgrCfg.HTTPListen != "127.0.0.1:2080" {
		t.Error(cgrCfg.HTTPListen)
	}
}

func TestCgrCfgJSONDefaultsTPdb(t *testing.T) {
	if cgrCfg.TpDbType != "redis" {
		t.Error(cgrCfg.TpDbType)
	}
	if cgrCfg.TpDbHost != "127.0.0.1" {
		t.Error(cgrCfg.TpDbHost)
	}
	if cgrCfg.TpDbPort != "6379" {
		t.Error(cgrCfg.TpDbPort)
	}
	if cgrCfg.TpDbName != "10" {
		t.Error(cgrCfg.TpDbName)
	}
	if cgrCfg.TpDbUser != "" {
		t.Error(cgrCfg.TpDbUser)
	}
	if cgrCfg.TpDbPass != "" {
		t.Error(cgrCfg.TpDbPass)
	}
}

func TestCgrCfgJSONDefaultsjsnDataDb(t *testing.T) {
	if cgrCfg.DataDbType != "redis" {
		t.Error(cgrCfg.DataDbType)
	}
	if cgrCfg.DataDbHost != "127.0.0.1" {
		t.Error(cgrCfg.DataDbHost)
	}
	if cgrCfg.DataDbPort != "6379" {
		t.Error(cgrCfg.DataDbPort)
	}
	if cgrCfg.DataDbName != "11" {
		t.Error(cgrCfg.DataDbName)
	}
	if cgrCfg.DataDbUser != "" {
		t.Error(cgrCfg.DataDbUser)
	}
	if cgrCfg.DataDbPass != "" {
		t.Error(cgrCfg.DataDbPass)
	}
	if cgrCfg.LoadHistorySize != 10 {
		t.Error(cgrCfg.LoadHistorySize)
	}
}

func TestCgrCfgJSONDefaultsStorDB(t *testing.T) {
	if cgrCfg.StorDBType != "mysql" {
		t.Error(cgrCfg.StorDBType)
	}
	if cgrCfg.StorDBHost != "127.0.0.1" {
		t.Error(cgrCfg.StorDBHost)
	}
	if cgrCfg.StorDBPort != "3306" {
		t.Error(cgrCfg.StorDBPort)
	}
	if cgrCfg.StorDBName != "cgrates" {
		t.Error(cgrCfg.StorDBName)
	}
	if cgrCfg.StorDBUser != "cgrates" {
		t.Error(cgrCfg.StorDBUser)
	}
	if cgrCfg.StorDBPass != "CGRateS.org" {
		t.Error(cgrCfg.StorDBPass)
	}
	if cgrCfg.StorDBMaxOpenConns != 100 {
		t.Error(cgrCfg.StorDBMaxOpenConns)
	}
	if cgrCfg.StorDBMaxIdleConns != 10 {
		t.Error(cgrCfg.StorDBMaxIdleConns)
	}
	Eslice := []string{}
	if !reflect.DeepEqual(cgrCfg.StorDBCDRSIndexes, Eslice) {
		t.Error(cgrCfg.StorDBCDRSIndexes)
	}
}

func TestCgrCfgJSONDefaultsBalancer(t *testing.T) {
	if cgrCfg.BalancerEnabled != false {
		t.Error(cgrCfg.BalancerEnabled)
	}
}

func TestCgrCfgJSONDefaultsRALs(t *testing.T) {

	eHaPoolcfg := []*HaPoolConfig{}

	if cgrCfg.RALsEnabled != false {
		t.Error(cgrCfg.RALsEnabled)
	}
	if cgrCfg.RALsBalancer != "" {
		t.Error(cgrCfg.RALsBalancer)
	}
	if !reflect.DeepEqual(cgrCfg.RALsCDRStatSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsCDRStatSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsHistorySConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsHistorySConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsPubSubSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsPubSubSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsUserSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsUserSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsAliasSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsAliasSConns)
	}
	if cgrCfg.RpSubjectPrefixMatching != false {
		t.Error(cgrCfg.RpSubjectPrefixMatching)
	}
	if cgrCfg.LcrSubjectPrefixMatching != false {
		t.Error(cgrCfg.LcrSubjectPrefixMatching)
	}
}

func TestCgrCfgJSONDefaultsScheduler(t *testing.T) {
	if cgrCfg.SchedulerEnabled != false {
		t.Error(cgrCfg.SchedulerEnabled)
	}
}

func TestCgrCfgJSONDefaultsCDRS(t *testing.T) {
	eHaPoolCfg := []*HaPoolConfig{}
	var eCdrExtr []*utils.RSRField
	if cgrCfg.CDRSEnabled != false {
		t.Error(cgrCfg.CDRSEnabled)
	}
	if !reflect.DeepEqual(eCdrExtr, cgrCfg.CDRSExtraFields) {
		t.Errorf(" expecting: %+v, received: %+v", eCdrExtr, cgrCfg.CDRSExtraFields)
	}
	if cgrCfg.CDRSStoreCdrs != true {
		t.Error(cgrCfg.CDRSStoreCdrs)
	}
	if cgrCfg.CDRScdrAccountSummary != false {
		t.Error(cgrCfg.CDRScdrAccountSummary)
	}
	if cgrCfg.CDRSSMCostRetries != 5 {
		t.Error(cgrCfg.CDRSSMCostRetries)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSRaterConns, []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}}) {
		t.Error(cgrCfg.CDRSRaterConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSPubSubSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSPubSubSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSUserSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSUserSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSAliaseSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSAliaseSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSStatSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSStatSConns)
	}
	if cgrCfg.CDRSOnlineCDRExports != nil {
		t.Error(cgrCfg.CDRSOnlineCDRExports)
	}
}

func TestCgrCfgJSONDefaultsCDRStats(t *testing.T) {
	if cgrCfg.CDRStatsEnabled != false {
		t.Error(cgrCfg.CDRStatsEnabled)
	}
	if cgrCfg.CDRStatsSaveInterval != 1*time.Minute {
		t.Error(cgrCfg.CDRStatsSaveInterval)
	}
}

func TestCgrCfgJSONDefaultsCdreProfiles(t *testing.T) {
	eFields := []*CfgCdrField{}
	eContentFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "CGRID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("CGRID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "RunID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RunID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "TOR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Direction", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("SetupTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: "Usage", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Usage", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Cost", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP), RoundingDecimals: 4},
	}
	eCdreCfg := map[string]*CdreConfig{
		"*default": {
			ExportFormat:        utils.MetaFileCSV,
			ExportPath:          "/var/spool/cgrates/cdre",
			Synchronous:         false,
			Attempts:            1,
			FieldSeparator:      ',',
			UsageMultiplyFactor: map[string]float64{utils.ANY: 1.0},
			CostMultiplyFactor:  1.0,
			HeaderFields:        eFields,
			ContentFields:       eContentFlds,
			TrailerFields:       eFields,
		},
	}
	if !reflect.DeepEqual(cgrCfg.CdreProfiles, eCdreCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.CdreProfiles, eCdreCfg)
	}
}

func TestCgrCfgJSONDefaultsSMGenericCfg(t *testing.T) {
	eSmGeCfg := &SmGenericConfig{
		Enabled:             false,
		ListenBijson:        "127.0.0.1:2014",
		RALsConns:           []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CDRsConns:           []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		SMGReplicationConns: []*HaPoolConfig{},
		DebitInterval:       0 * time.Second,
		MinCallDuration:     0 * time.Second,
		MaxCallDuration:     3 * time.Hour,
		SessionTTL:          0 * time.Second,
		SessionIndexes:      utils.StringMap{},
	}

	if !reflect.DeepEqual(cgrCfg.SmGenericConfig, eSmGeCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.SmGenericConfig, eSmGeCfg)
	}

}

func TestCgrCfgJSONDefaultsSMFsConfig(t *testing.T) {
	eSmFsCfg := &SmFsConfig{
		Enabled:             false,
		RALsConns:           []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CDRsConns:           []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		RLsConns:            []*HaPoolConfig{},
		CreateCdr:           false,
		ExtraFields:         nil,
		DebitInterval:       10 * time.Second,
		MinCallDuration:     0 * time.Second,
		MaxCallDuration:     3 * time.Hour,
		MinDurLowBalance:    5 * time.Second,
		LowBalanceAnnFile:   "",
		EmptyBalanceContext: "",
		EmptyBalanceAnnFile: "",
		SubscribePark:       true,
		ChannelSyncInterval: 5 * time.Minute,
		MaxWaitConnection:   2 * time.Second,
		EventSocketConns:    []*FsConnConfig{&FsConnConfig{Address: "127.0.0.1:8021", Password: "ClueCon", Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.SmFsConfig, eSmFsCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.SmFsConfig, eSmFsCfg)
	}
}

func TestCgrCfgJSONDefaultsSMKamConfig(t *testing.T) {
	eSmKaCfg := &SmKamConfig{
		Enabled:         false,
		RALsConns:       []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CDRsConns:       []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CreateCdr:       false,
		DebitInterval:   10 * time.Second,
		MinCallDuration: 0 * time.Second,
		MaxCallDuration: 3 * time.Hour,
		EvapiConns:      []*KamConnConfig{&KamConnConfig{Address: "127.0.0.1:8448", Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.SmKamConfig, eSmKaCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.CdreProfiles, eSmKaCfg)
	}
}

func TestCgrCfgJSONDefaultsSMOsipsConfig(t *testing.T) {
	eSmOpCfg := &SmOsipsConfig{
		Enabled:                 false,
		ListenUdp:               "127.0.0.1:2020",
		RALsConns:               []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CDRsConns:               []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CreateCdr:               false,
		DebitInterval:           10 * time.Second,
		MinCallDuration:         0 * time.Second,
		MaxCallDuration:         3 * time.Hour,
		EventsSubscribeInterval: 60 * time.Second,
		MiAddr:                  "127.0.0.1:8020",
	}

	if !reflect.DeepEqual(cgrCfg.SmOsipsConfig, eSmOpCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.SmOsipsConfig, eSmOpCfg)
	}
}

func TestCgrCfgJSONDefaultsSMAsteriskCfg(t *testing.T) {
	eSmAsCfg := &SMAsteriskCfg{
		Enabled:       false,
		CreateCDR:     false,
		AsteriskConns: []*AsteriskConnCfg{&AsteriskConnCfg{Address: "127.0.0.1:8088", User: "cgrates", Password: "CGRateS.org", ConnectAttempts: 3, Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.smAsteriskCfg, eSmAsCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.smAsteriskCfg, eSmAsCfg)
	}
}

func TestCgrCfgJSONDefaultsHistoryS(t *testing.T) {
	if cgrCfg.HistoryServerEnabled != false {
		t.Error(cgrCfg.HistoryServerEnabled)
	}
	if cgrCfg.HistoryDir != "/var/lib/cgrates/history" {
		t.Error(cgrCfg.HistoryDir)
	}
	if cgrCfg.HistorySaveInterval != 1*time.Second {
		t.Error(cgrCfg.HistorySaveInterval)
	}
}

func TestCgrCfgJSONDefaultsPubSubS(t *testing.T) {
	if cgrCfg.PubSubServerEnabled != false {
		t.Error(cgrCfg.PubSubServerEnabled)
	}
}

func TestCgrCfgJSONDefaultsAliasesS(t *testing.T) {
	if cgrCfg.AliasesServerEnabled != false {
		t.Error(cgrCfg.AliasesServerEnabled)
	}
}

func TestCgrCfgJSONDefaultsUserS(t *testing.T) {
	eStrSlc := []string{}
	if cgrCfg.UserServerEnabled != false {
		t.Error(cgrCfg.UserServerEnabled)
	}

	if !reflect.DeepEqual(cgrCfg.UserServerIndexes, eStrSlc) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.UserServerIndexes, eStrSlc)
	}
}

func TestCgrCfgJSONDefaultsResLimCfg(t *testing.T) {
	eResLiCfg := &ResourceLimiterConfig{
		Enabled:           false,
		CDRStatConns:      []*HaPoolConfig{},
		CacheDumpInterval: 0 * time.Second,
		UsageTTL:          3 * time.Hour,
	}

	if !reflect.DeepEqual(cgrCfg.resourceLimiterCfg, eResLiCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.resourceLimiterCfg, eResLiCfg)
	}

}

func TestCgrCfgJSONDefaultsDiameterAgentCfg(t *testing.T) {
	reqProc := []*DARequestProcessor{&DARequestProcessor{
		Id:                "*default",
		DryRun:            false,
		PublishEvent:      false,
		RequestFilter:     utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Type(0)", utils.INFIELD_SEP),
		Flags:             utils.StringMap{},
		ContinueOnSuccess: false,
		AppendCCA:         true,
		CCRFields: []*CfgCdrField{
			&CfgCdrField{Tag: "TOR", FieldId: "ToR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*voice", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "OriginID", FieldId: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Session-Id", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "RequestType", FieldId: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Direction", FieldId: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*out", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Tenant", FieldId: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Category", FieldId: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^call", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Account", FieldId: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Subject", FieldId: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Destination", FieldId: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Service-Information>IN-Information>Real-Called-Number", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "SetupTime", FieldId: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Event-Timestamp", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "AnswerTime", FieldId: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Event-Timestamp", utils.INFIELD_SEP), Mandatory: true},
			&CfgCdrField{Tag: "Usage", FieldId: "Usage", Type: "*handler", HandlerId: "*ccr_usage", Mandatory: true},
			&CfgCdrField{Tag: "SubscriberID", FieldId: "SubscriberId", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Data", utils.INFIELD_SEP), Mandatory: true},
		},
		CCAFields: []*CfgCdrField{
			&CfgCdrField{Tag: "GrantedUnits", FieldId: "Granted-Service-Unit>CC-Time", Type: "*handler", HandlerId: "*cca_usage", Mandatory: true}},
	},
	}

	testDA := &DiameterAgentCfg{
		Enabled:           false,
		Listen:            "127.0.0.1:3868",
		DictionariesDir:   "/usr/share/cgrates/diameter/dict/",
		SMGenericConns:    []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		PubSubConns:       []*HaPoolConfig{},
		CreateCDR:         true,
		DebitInterval:     5 * time.Minute,
		Timezone:          "",
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorId:          0,
		ProductName:       "CGRateS",
		RequestProcessors: reqProc,
	}

	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Enabled, testDA.Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Listen, testDA.Listen) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.Listen, testDA.Listen)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.DictionariesDir, testDA.DictionariesDir) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.DictionariesDir, testDA.DictionariesDir)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.SMGenericConns, testDA.SMGenericConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.SMGenericConns, testDA.SMGenericConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.PubSubConns, testDA.PubSubConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.PubSubConns, testDA.PubSubConns)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.CreateCDR, testDA.CreateCDR) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.CreateCDR, testDA.CreateCDR)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.DebitInterval, testDA.DebitInterval) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.DebitInterval, testDA.DebitInterval)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Timezone, testDA.Timezone) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.diameterAgentCfg.Timezone, testDA.Timezone)
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
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.RequestProcessors, testDA.RequestProcessors) {
		t.Errorf("expecting: %+v, received: %+v", testDA.RequestProcessors, cgrCfg.diameterAgentCfg.RequestProcessors)
	}
}

func TestCgrCfgJSONDefaultsMailer(t *testing.T) {
	if cgrCfg.MailerServer != "localhost" {
		t.Error(cgrCfg.MailerServer)
	}
	if cgrCfg.MailerAuthUser != "cgrates" {
		t.Error(cgrCfg.MailerAuthUser)
	}
	if cgrCfg.MailerAuthPass != "CGRateS.org" {
		t.Error(cgrCfg.MailerAuthPass)
	}
	if cgrCfg.MailerFromAddr != "cgr-mailer@localhost.localdomain" {
		t.Error(cgrCfg.MailerFromAddr)
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
		ClientTracking:       utils.ParseRSRFieldsMustCompile("CGRID", utils.INFIELD_SEP),
		CustomerNumber:       utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP),
		OrigNumber:           utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP),
		TermNumber:           utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP),
		BillToNumber:         utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
		Zipcode:              utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
		P2PZipcode:           utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
		P2PPlus4:             utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
		Units:                utils.ParseRSRFieldsMustCompile("^1", utils.INFIELD_SEP),
		UnitType:             utils.ParseRSRFieldsMustCompile("^00", utils.INFIELD_SEP),
		TaxIncluded:          utils.ParseRSRFieldsMustCompile("^0", utils.INFIELD_SEP),
		TaxSitusRule:         utils.ParseRSRFieldsMustCompile("^04", utils.INFIELD_SEP),
		TransTypeCode:        utils.ParseRSRFieldsMustCompile("^010101", utils.INFIELD_SEP),
		SalesTypeCode:        utils.ParseRSRFieldsMustCompile("^R", utils.INFIELD_SEP),
		TaxExemptionCodeList: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
	}

	if !reflect.DeepEqual(cgrCfg.sureTaxCfg, eSureTaxCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.sureTaxCfg, eSureTaxCfg)
	}
}
