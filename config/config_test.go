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
"freeswitch_agent": {
	"enabled": true,				// starts SessionManager service: <true|false>
	"event_socket_conns":[					// instantiate connections to multiple FreeSWITCH servers
		{"address": "1.2.3.4:8021", "password": "ClueCon", "reconnects": 3},
		{"address": "1.2.3.5:8021", "password": "ClueCon", "reconnects": 5}
	],
},

}`
	eCgrCfg, err := NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	eCgrCfg.fsAgentCfg.Enabled = true
	eCgrCfg.fsAgentCfg.EventSocketConns = []*FsConnConfig{
		&FsConnConfig{Address: "1.2.3.4:8021", Password: "ClueCon", Reconnects: 3},
		&FsConnConfig{Address: "1.2.3.5:8021", Password: "ClueCon", Reconnects: 5},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(JSN_CFG); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg) {
		t.Errorf("Expected: %+v, received: %+v", eCgrCfg.fsAgentCfg, cgrCfg.fsAgentCfg)
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
				&CfgCdrField{FieldId: "ToR", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile("~7:s/^(voice|data|sms|mms|generic)$/*$1/", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "AnswerTime", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP)},
				&CfgCdrField{FieldId: "Usage", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile("~9:s/^(\\d+)$/${1}s/", utils.INFIELD_SEP)},
			},
			TrailerFields: make([]*CfgCdrField, 0),
			CacheDumpFields: []*CfgCdrField{
				&CfgCdrField{Tag: "CGRID", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.CGRID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RunID", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.MEDI_RUNID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.TOR, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.OriginID, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.RequestType, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Tenant, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Category, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Account, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Subject", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Subject, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Destination, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED,
					Value:  utils.ParseRSRFieldsMustCompile(utils.SetupTime, utils.INFIELD_SEP),
					Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED,
					Value:  utils.ParseRSRFieldsMustCompile(utils.AnswerTime, utils.INFIELD_SEP),
					Layout: "2006-01-02T15:04:05Z07:00"},
				&CfgCdrField{Tag: "Usage", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.Usage, utils.INFIELD_SEP)},
				&CfgCdrField{Tag: "Cost", Type: utils.META_COMPOSED,
					Value: utils.ParseRSRFieldsMustCompile(utils.COST, utils.INFIELD_SEP)},
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
	if cgrCfg.Logger != utils.MetaSysLog {
		t.Error(cgrCfg.Logger)
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
	if cgrCfg.DataDbName != "10" {
		t.Error(cgrCfg.DataDbName)
	}
	if cgrCfg.DataDbUser != "cgrates" {
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
	if cgrCfg.StorDBPass != "" {
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

func TestCgrCfgJSONDefaultsRALs(t *testing.T) {
	//asd
	eHaPoolcfg := []*HaPoolConfig{}

	if cgrCfg.RALsEnabled != false {
		t.Error(cgrCfg.RALsEnabled)
	}

	if !reflect.DeepEqual(cgrCfg.RALsThresholdSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsThresholdSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsCDRStatSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsCDRStatSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsPubSubSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsPubSubSConns)
	}
	if !reflect.DeepEqual(cgrCfg.RALsAttributeSConns, eHaPoolcfg) {
		t.Error(cgrCfg.RALsAttributeSConns)
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
	eMaxCU := map[string]time.Duration{
		utils.ANY:   time.Duration(189 * time.Hour),
		utils.VOICE: time.Duration(72 * time.Hour),
		utils.DATA:  time.Duration(107374182400),
		utils.SMS:   time.Duration(10000),
	}
	if !reflect.DeepEqual(eMaxCU, cgrCfg.RALsMaxComputedUsage) {
		t.Errorf("Expecting: %+v, received: %+v", eMaxCU, cgrCfg.RALsMaxComputedUsage)
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
	if !reflect.DeepEqual(cgrCfg.CDRSAttributeSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSAttributeSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSUserSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSUserSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSAliaseSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSAliaseSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSCDRStatSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSCDRStatSConns)
	}
	if !reflect.DeepEqual(cgrCfg.CDRSThresholdSConns, eHaPoolCfg) {
		t.Error(cgrCfg.CDRSThresholdSConns)
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
		&CfgCdrField{Tag: "CGRID", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("CGRID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "RunID", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("RunID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "TOR", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "OriginID", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "RequestType", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Tenant", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Category", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Account", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Subject", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Destination", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "SetupTime", Type: "*composed",
			Value:  utils.ParseRSRFieldsMustCompile("SetupTime", utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: "AnswerTime", Type: "*composed",
			Value:  utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP),
			Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: "Usage", Type: "*composed",
			Value: utils.ParseRSRFieldsMustCompile("Usage", utils.INFIELD_SEP)},
		&CfgCdrField{Tag: "Cost", Type: "*composed",
			Value:            utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP),
			RoundingDecimals: 4},
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
	eSessionSCfg := &SessionSCfg{
		Enabled:      false,
		ListenBijson: "127.0.0.1:2014",
		RALsConns: []*HaPoolConfig{
			&HaPoolConfig{Address: "*internal"}},
		CDRsConns: []*HaPoolConfig{
			&HaPoolConfig{Address: "*internal"}},
		ResSConns:               []*HaPoolConfig{},
		SupplSConns:             []*HaPoolConfig{},
		AttrSConns:              []*HaPoolConfig{},
		SessionReplicationConns: []*HaPoolConfig{},
		DebitInterval:           0 * time.Second,
		MinCallDuration:         0 * time.Second,
		MaxCallDuration:         3 * time.Hour,
		SessionTTL:              0 * time.Second,
		SessionIndexes:          utils.StringMap{},
		ClientProtocol:          1.0,
	}
	if !reflect.DeepEqual(eSessionSCfg, cgrCfg.sessionSCfg) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSessionSCfg), utils.ToJSON(cgrCfg.sessionSCfg))
	}

}
func TestCgrCfgJSONDefaultsCacheCFG(t *testing.T) {
	eCacheCfg := CacheConfig{
		utils.CacheDestinations: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheReverseDestinations: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheRatingPlans: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheRatingProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheLCRRules: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheCDRStatS: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActions: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActionPlans: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAccountActionPlans: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheActionTriggers: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSharedGroups: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAliases: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheReverseAliases: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheDerivedChargers: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheTimings: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResourceProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResources: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheEventResources: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(1 * time.Minute), StaticTTL: false},
		utils.CacheStatQueueProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(1 * time.Minute), StaticTTL: false, Precache: false},
		utils.CacheStatQueues: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(1 * time.Minute), StaticTTL: false, Precache: false},
		utils.CacheThresholdProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholds: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheFilters: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSupplierProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAttributeProfiles: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResourceFilterIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheResourceFilterRevIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheStatFilterIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheStatFilterRevIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholdFilterIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheThresholdFilterRevIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSupplierFilterIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheSupplierFilterRevIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAttributeFilterIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
		utils.CacheAttributeFilterRevIndexes: &CacheParamConfig{Limit: -1,
			TTL: time.Duration(0), StaticTTL: false, Precache: false},
	}

	if !reflect.DeepEqual(eCacheCfg, cgrCfg.CacheCfg()) {
		t.Errorf("received: %s, \nexpecting: %s",
			utils.ToJSON(eCacheCfg), utils.ToJSON(cgrCfg.CacheCfg()))
	}
}

func TestCgrCfgJSONDefaultsSMFsConfig(t *testing.T) {
	eFsAgentCfg := &FsAgentConfig{
		Enabled:             false,
		SessionSConns:       []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		SubscribePark:       true,
		CreateCdr:           false,
		ExtraFields:         nil,
		EmptyBalanceContext: "",
		EmptyBalanceAnnFile: "",
		ChannelSyncInterval: 5 * time.Minute,
		MaxWaitConnection:   2 * time.Second,
		EventSocketConns: []*FsConnConfig{
			&FsConnConfig{Address: "127.0.0.1:8021",
				Password: "ClueCon", Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.fsAgentCfg, eFsAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.fsAgentCfg, eFsAgentCfg)
	}
}

func TestCgrCfgJSONDefaultsSMKamConfig(t *testing.T) {
	eSmKaCfg := &SmKamConfig{
		Enabled:         false,
		RALsConns:       []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		CDRsConns:       []*HaPoolConfig{&HaPoolConfig{Address: "*internal"}},
		RLsConns:        []*HaPoolConfig{},
		CreateCdr:       false,
		DebitInterval:   10 * time.Second,
		MinCallDuration: 0 * time.Second,
		MaxCallDuration: 3 * time.Hour,
		EvapiConns:      []*KamConnConfig{&KamConnConfig{Address: "127.0.0.1:8448", Reconnects: 5}},
	}
	if !reflect.DeepEqual(cgrCfg.SmKamConfig, eSmKaCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.SmKamConfig, eSmKaCfg)
	}
}

func TestCgrCfgJSONDefaultssteriskAgentCfg(t *testing.T) {
	eAstAgentCfg := &AsteriskAgentCfg{
		Enabled: false,
		SessionSConns: []*HaPoolConfig{
			&HaPoolConfig{Address: "*internal"}},
		CreateCDR: false,
		AsteriskConns: []*AsteriskConnCfg{
			&AsteriskConnCfg{Address: "127.0.0.1:8088",
				User: "cgrates", Password: "CGRateS.org",
				ConnectAttempts: 3, Reconnects: 5}},
	}

	if !reflect.DeepEqual(cgrCfg.asteriskAgentCfg, eAstAgentCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.asteriskAgentCfg, eAstAgentCfg)
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

func TestCgrCfgJSONDefaultFiltersCfg(t *testing.T) {
	eFiltersCfg := &FilterSCfg{
		StatSConns: []*HaPoolConfig{},
	}
	if !reflect.DeepEqual(cgrCfg.filterSCfg, eFiltersCfg) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.filterSCfg, eFiltersCfg)
	}
}

func TestCgrCfgJSONDefaultSAttributeSCfg(t *testing.T) {
	eAliasSCfg := &AttributeSCfg{
		Enabled:             false,
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},
	}
	if !reflect.DeepEqual(eAliasSCfg, cgrCfg.attributeSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eAliasSCfg, cgrCfg.attributeSCfg)
	}
}

func TestCgrCfgJSONDefaultsResLimCfg(t *testing.T) {
	eResLiCfg := &ResourceSConfig{
		Enabled:             false,
		ThresholdSConns:     []*HaPoolConfig{},
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
		StoreInterval:       0,
		ThresholdSConns:     []*HaPoolConfig{},
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
		StringIndexedFields: nil,
		PrefixIndexedFields: &[]string{},

		RALsConns: []*HaPoolConfig{
			&HaPoolConfig{Address: "*internal"},
		},
		ResourceSConns: []*HaPoolConfig{},
		StatSConns:     []*HaPoolConfig{},
	}
	if !reflect.DeepEqual(eSupplSCfg, cgrCfg.supplierSCfg) {
		t.Errorf("received: %+v, expecting: %+v", eSupplSCfg, cgrCfg.supplierSCfg)
	}
}

func TestCgrCfgJSONDefaultsDiameterAgentCfg(t *testing.T) {
	testDA := &DiameterAgentCfg{
		Enabled:         false,
		Listen:          "127.0.0.1:3868",
		DictionariesDir: "/usr/share/cgrates/diameter/dict/",
		SessionSConns: []*HaPoolConfig{
			&HaPoolConfig{Address: "*internal"}},
		PubSubConns:       []*HaPoolConfig{},
		CreateCDR:         true,
		DebitInterval:     5 * time.Minute,
		Timezone:          "",
		OriginHost:        "CGR-DA",
		OriginRealm:       "cgrates.org",
		VendorId:          0,
		ProductName:       "CGRateS",
		RequestProcessors: nil,
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
	if !reflect.DeepEqual(cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.diameterAgentCfg.SessionSConns, testDA.SessionSConns)
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

func TestCgrCfgJSONDefaultsHTTP(t *testing.T) {
	if cgrCfg.HTTPJsonRPCURL != "/jsonrpc" {
		t.Error(cgrCfg.HTTPJsonRPCURL)
	}
	if cgrCfg.HTTPWSURL != "/ws" {
		t.Error(cgrCfg.HTTPWSURL)
	}
	if cgrCfg.HTTPUseBasicAuth != false {
		t.Error(cgrCfg.HTTPUseBasicAuth)
	}
	if !reflect.DeepEqual(cgrCfg.HTTPAuthUsers, map[string]string{}) {
		t.Error(cgrCfg.HTTPAuthUsers)
	}
}

func TestRadiusAgentCfg(t *testing.T) {
	testRA := &RadiusAgentCfg{
		Enabled:            false,
		ListenNet:          "udp",
		ListenAuth:         "127.0.0.1:1812",
		ListenAcct:         "127.0.0.1:1813",
		ClientSecrets:      map[string]string{utils.META_DEFAULT: "CGRateS.org"},
		ClientDictionaries: map[string]string{utils.META_DEFAULT: "/usr/share/cgrates/radius/dict/"},
		SessionSConns:      []*HaPoolConfig{&HaPoolConfig{Address: utils.MetaInternal}},
		CreateCDR:          true,
		CDRRequiresSession: false,
		Timezone:           "",
		RequestProcessors:  nil,
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.Enabled, testRA.Enabled) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.Enabled, testRA.Enabled)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.ListenNet, testRA.ListenNet) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.ListenNet, testRA.ListenNet)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.ListenAuth, testRA.ListenAuth) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.ListenAuth, testRA.ListenAuth)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.ListenAcct, testRA.ListenAcct) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.ListenAcct, testRA.ListenAcct)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.ClientSecrets, testRA.ClientSecrets) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.ClientSecrets, testRA.ClientSecrets)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.ClientDictionaries, testRA.ClientDictionaries) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.ClientDictionaries, testRA.ClientDictionaries)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.SessionSConns, testRA.SessionSConns) {
		t.Errorf("expecting: %+v, received: %+v", cgrCfg.radiusAgentCfg.SessionSConns, testRA.SessionSConns)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.CreateCDR, testRA.CreateCDR) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.radiusAgentCfg.CreateCDR, testRA.CreateCDR)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.CDRRequiresSession, testRA.CDRRequiresSession) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.radiusAgentCfg.CDRRequiresSession, testRA.CDRRequiresSession)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.Timezone, testRA.Timezone) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.radiusAgentCfg.Timezone, testRA.Timezone)
	}
	if !reflect.DeepEqual(cgrCfg.radiusAgentCfg.RequestProcessors, testRA.RequestProcessors) {
		t.Errorf("received: %+v, expecting: %+v", cgrCfg.radiusAgentCfg.RequestProcessors, testRA.RequestProcessors)
	}
}

func TestDbDefaults(t *testing.T) {
	dbdf := NewDbDefaults()
	flagInput := utils.MetaDynamic
	dbs := []string{utils.MONGO, utils.REDIS, utils.MYSQL}
	for _, dbtype := range dbs {
		host := dbdf.DBHost(dbtype, flagInput)
		if host != utils.LOCALHOST {
			t.Errorf("received: %+v, expecting: %+v", host, utils.LOCALHOST)
		}
		user := dbdf.DBUser(dbtype, flagInput)
		if user != utils.CGRATES {
			t.Errorf("received: %+v, expecting: %+v", user, utils.CGRATES)
		}
		port := dbdf.DBPort(dbtype, flagInput)
		if port != dbdf[dbtype]["DbPort"] {
			t.Errorf("received: %+v, expecting: %+v", port, dbdf[dbtype]["DbPort"])
		}
		name := dbdf.DBName(dbtype, flagInput)
		if name != dbdf[dbtype]["DbName"] {
			t.Errorf("received: %+v, expecting: %+v", name, dbdf[dbtype]["DbName"])
		}
		pass := dbdf.DBPass(dbtype, flagInput)
		if pass != dbdf[dbtype]["DbPass"] {
			t.Errorf("received: %+v, expecting: %+v", pass, dbdf[dbtype]["DbPass"])
		}
	}
}
