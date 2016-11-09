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
	cgrCfg, err = NewCGRConfigFromJsonStringWithDefaults(CGRATES_CFG_JSON)
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
	if cgrCfg.HttpPosterAttempts != 3 {
		t.Error(cgrCfg.HttpPosterAttempts)
	}
	if cgrCfg.HttpFailedDir != "/var/spool/cgrates/http_failed" {
		t.Error(cgrCfg.HttpFailedDir)
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

	test1 := []*HaPoolConfig{}

	if cgrCfg.RALsEnabled != false {
		t.Error(cgrCfg.RALsEnabled)
	}
	if cgrCfg.RALsBalancer != "" {
		t.Error(cgrCfg.RALsBalancer)
	}
	if !reflect.DeepEqual(cgrCfg.RALsCDRStatSConns, test1) {
		t.Error(cgrCfg.RALsCDRStatSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsHistorySConns, test1) {
		t.Error(cgrCfg.RALsHistorySConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsPubSubSConns, test1) {
		t.Error(cgrCfg.RALsPubSubSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsUserSConns, test1) {
		t.Error(cgrCfg.RALsUserSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsAliasSConns, test1) {
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

	test1 := []*HaPoolConfig{}

	if cgrCfg.CDRSEnabled != false {
		t.Error(cgrCfg.CDRSEnabled)
	}
	/*
		test4 := *utils.NewRSRField("")

		if !reflect.DeepEqual(cgrCfg.CDRSExtraFields, test4) {
			t.Errorf("expecting: %+v, received: %+v", cgrCfg.CDRSExtraFields, test4)
		}
	*/
	if cgrCfg.CDRSStoreCdrs != true {
		t.Error(cgrCfg.CDRSStoreCdrs)
	}
	if cgrCfg.CDRScdrAccountSummary != false {
		t.Error(cgrCfg.CDRScdrAccountSummary)
	}
	if cgrCfg.CDRSSMCostRetries != 5 {
		t.Error(cgrCfg.CDRSSMCostRetries)
	}
	test3 := []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}}
	if !reflect.DeepEqual(cgrCfg.CDRSRaterConns, test3) {
		t.Error(cgrCfg.CDRSRaterConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSPubSubSConns, test1) {
		t.Error(cgrCfg.CDRSPubSubSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSUserSConns, test1) {
		t.Error(cgrCfg.CDRSUserSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSAliaseSConns, test1) {
		t.Error(cgrCfg.CDRSAliaseSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSStatSConns, test1) {
		t.Error(cgrCfg.CDRSStatSConns)
	}
	test2 := []*CDRReplicationCfg{}
	if !reflect.DeepEqual(cgrCfg.CDRSCdrReplication, test2) {
		t.Error(cgrCfg.CDRSCdrReplication)
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

/*
func TestCgrCfgJSONDefaultsCdreProfiles(t *testing.T) {
	eContentFlds := []*CfgCdrField{
		&CfgCdrField{
			Tag:   "CGRID",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("CGRID", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "RunID",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("RunID", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "ToR",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "OriginID",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "RequestType",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Direction",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Direction", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Tenant",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Category",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Account",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Subject",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:   "Destination",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:    "SetupTime",
			Type:   "*composed",
			Value:  utils.ParseRSRFieldsMustCompile("SetupTime", utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{
			Tag:    "AnswerTime",
			Type:   "*composed",
			Value:  utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{
			Tag:   "Usage",
			Type:  "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Usage", utils.INFIELD_SEP)},
		&CfgCdrField{
			Tag:              "Cost",
			Type:             "*composed",
			Value:            utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP),
			RoundingDecimals: 4},
	}
	test := map[string]*CdreConfig{
		"*default": {
			CdrFormat:                  "csv",
			FieldSeparator:             ',',
			DataUsageMultiplyFactor:    1,
			SMSUsageMultiplyFactor:     1,
			MMSUsageMultiplyFactor:     1,
			GenericUsageMultiplyFactor: 1,
			CostMultiplyFactor:         1,
			ExportDirectory:            "/var/spool/cgrates/cdre",
			HeaderFields:               []*CfgCdrField{},
			ContentFields:              eContentFlds,
			TrailerFields:              []*CfgCdrField{},
		}}
	if !reflect.DeepEqual(cgrCfg.CdreProfiles, test) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.CdreProfiles, test)
	}
}


func TestCgrCfgJSONDefaultsCdrcProfiles(t *testing.T) {

}
*/
func TestCgrCfgJSONDefaultsSMs(t *testing.T) {
	if cgrCfg.SmGenericConfig == nil {
		t.Error(cgrCfg.SmGenericConfig)
	}
	if cgrCfg.SmFsConfig == nil {
		t.Error(cgrCfg.SmFsConfig)
	}
	if cgrCfg.SmKamConfig == nil {
		t.Error(cgrCfg.SmKamConfig)
	}
	if cgrCfg.SmOsipsConfig == nil {
		t.Error(cgrCfg.SmOsipsConfig)
	}
	if cgrCfg.smAsteriskCfg == nil {
		t.Error(cgrCfg.smAsteriskCfg)
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
	if cgrCfg.UserServerEnabled != false {
		t.Error(cgrCfg.UserServerEnabled)
	}
	if cgrCfg.UserServerIndexes == nil {
		t.Error(cgrCfg.UserServerIndexes)
	}
}

func TestCgrCfgJSONDefaultsResLimCfg(t *testing.T) {
	if cgrCfg.resourceLimiterCfg == nil {
		t.Error(cgrCfg.resourceLimiterCfg)
	}
}

func TestCgrCfgJSONDefaultsDiameterAgentCfg(t *testing.T) {
	test := DiameterAgentCfg{
		Enabled:         false,
		Listen:          "127.0.0.1:3868",
		DictionariesDir: "/usr/share/cgrates/diameter/dict/",
		SMGenericConns:  []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
	}

	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Enabled, test.Enabled) {
		t.Error(cgrCfg.diameterAgentCfg.Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Listen, test.Listen) {
		t.Error(cgrCfg.diameterAgentCfg.Listen)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.DictionariesDir, test.DictionariesDir) {
		t.Error(cgrCfg.diameterAgentCfg.DictionariesDir)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.Listen, test.Listen) {
		t.Error(cgrCfg.diameterAgentCfg)
	}
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.SMGenericConns, test.SMGenericConns) {
		t.Error(cgrCfg.diameterAgentCfg.SMGenericConns)
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
	test2 := &SureTaxCfg{
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
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.Url, test2.Url) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.Url, test2.Url)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ClientNumber, test2.ClientNumber) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ClientNumber, test2.ClientNumber)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ValidationKey, test2.ValidationKey) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ValidationKey, test2.ValidationKey)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.BusinessUnit, test2.BusinessUnit) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.BusinessUnit, test2.BusinessUnit)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.Timezone, test2.Timezone) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.Timezone, test2.Timezone)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.IncludeLocalCost, test2.IncludeLocalCost) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.IncludeLocalCost, test2.IncludeLocalCost)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ReturnFileCode, test2.ReturnFileCode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ReturnFileCode, test2.ReturnFileCode)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ResponseGroup, test2.ResponseGroup) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ResponseGroup, test2.ResponseGroup)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ResponseType, test2.ResponseType) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ResponseType, test2.ResponseType)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.RegulatoryCode, test2.RegulatoryCode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.RegulatoryCode, test2.RegulatoryCode)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.ClientTracking, test2.ClientTracking) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.ClientTracking, test2.ClientTracking)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.CustomerNumber, test2.CustomerNumber) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.CustomerNumber, test2.CustomerNumber)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.OrigNumber, test2.OrigNumber) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.OrigNumber, test2.OrigNumber)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.TermNumber, test2.TermNumber) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.TermNumber, test2.TermNumber)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.BillToNumber, test2.BillToNumber) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.BillToNumber, test2.BillToNumber)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.Zipcode, test2.Zipcode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.Zipcode, test2.Zipcode)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.P2PZipcode, test2.P2PZipcode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.P2PZipcode, test2.P2PZipcode)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.P2PPlus4, test2.P2PPlus4) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.P2PPlus4, test2.P2PPlus4)
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.Units, test2.Units) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.Units[0], test2.Units[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.UnitType, test2.UnitType) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.UnitType[0], test2.UnitType[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.TaxIncluded, test2.TaxIncluded) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.TaxIncluded[0], test2.TaxIncluded[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.TaxSitusRule, test2.TaxSitusRule) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.TaxSitusRule[0], test2.TaxSitusRule[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.TransTypeCode, test2.TransTypeCode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.TransTypeCode[0], test2.TransTypeCode[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.SalesTypeCode, test2.SalesTypeCode) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.SalesTypeCode[0], test2.SalesTypeCode[0])
	}
	if !reflect.DeepEqual(cgrCfg.sureTaxCfg.TaxExemptionCodeList, test2.TaxExemptionCodeList) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.sureTaxCfg.TaxExemptionCodeList[0], test2.TaxExemptionCodeList[0])
	}
}
