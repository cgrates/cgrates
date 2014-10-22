/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestConfigSharing(t *testing.T) {
	cfg, _ := NewDefaultCGRConfig()
	SetCgrConfig(cfg)
	cfgReturn := CgrConfig()
	if !reflect.DeepEqual(cfgReturn, cfg) {
		t.Errorf("Retrieved %v, Expected %v", cfgReturn, cfg)
	}
}

// Make sure defaults did not change by mistake
func TestDefaults(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	eCfg := &CGRConfig{}
	eCfg.RatingDBType = REDIS
	eCfg.RatingDBHost = "127.0.0.1"
	eCfg.RatingDBPort = "6379"
	eCfg.RatingDBName = "10"
	eCfg.RatingDBUser = ""
	eCfg.RatingDBPass = ""
	eCfg.AccountDBType = REDIS
	eCfg.AccountDBHost = "127.0.0.1"
	eCfg.AccountDBPort = "6379"
	eCfg.AccountDBName = "11"
	eCfg.AccountDBUser = ""
	eCfg.AccountDBPass = ""
	eCfg.StorDBType = utils.MYSQL
	eCfg.StorDBHost = "localhost"
	eCfg.StorDBPort = "3306"
	eCfg.StorDBName = "cgrates"
	eCfg.StorDBUser = "cgrates"
	eCfg.StorDBPass = "CGRateS.org"
	eCfg.StorDBMaxOpenConns = 100
	eCfg.StorDBMaxIdleConns = 10
	eCfg.DBDataEncoding = utils.MSGPACK
	eCfg.RPCJSONListen = "127.0.0.1:2012"
	eCfg.RPCGOBListen = "127.0.0.1:2013"
	eCfg.HTTPListen = "127.0.0.1:2080"
	eCfg.DefaultReqType = utils.RATED
	eCfg.DefaultCategory = "call"
	eCfg.DefaultTenant = "cgrates.org"
	eCfg.DefaultSubject = "cgrates"
	eCfg.RoundingDecimals = 10
	eCfg.HttpSkipTlsVerify = false
	eCfg.TpExportPath = "/var/log/cgrates/tpe"
	eCfg.XmlCfgDocument = nil
	eCfg.RaterEnabled = false
	eCfg.RaterBalancer = ""
	eCfg.BalancerEnabled = false
	eCfg.SchedulerEnabled = false
	eCfg.CdreDefaultInstance = NewDefaultCdreConfig()
	eCfg.CdrcInstances = []*CdrcConfig{NewDefaultCdrcConfig()}
	eCfg.CDRSEnabled = false
	eCfg.CDRSExtraFields = []*utils.RSRField{}
	eCfg.CDRSMediator = ""
	eCfg.CDRSStats = ""
	eCfg.CDRSStoreDisable = false
	eCfg.CDRStatsEnabled = false
	eCfg.CDRStatConfig = &CdrStatsConfig{Id: utils.META_DEFAULT, QueueLength: 50, TimeWindow: time.Duration(1) * time.Hour, Metrics: []string{"ASR", "ACD", "ACC"}}
	eCfg.MediatorEnabled = false
	eCfg.MediatorRater = utils.INTERNAL
	eCfg.MediatorReconnects = 3
	eCfg.MediatorStats = ""
	eCfg.MediatorStoreDisable = false
	eCfg.SMEnabled = false
	eCfg.SMSwitchType = FS
	eCfg.SMRater = utils.INTERNAL
	eCfg.SMCdrS = ""
	eCfg.SMReconnects = 3
	eCfg.SMDebitInterval = 10
	eCfg.SMMinCallDuration = time.Duration(0)
	eCfg.SMMaxCallDuration = time.Duration(3) * time.Hour
	eCfg.FreeswitchServer = "127.0.0.1:8021"
	eCfg.FreeswitchPass = "ClueCon"
	eCfg.FreeswitchReconnects = 5
	eCfg.FSMinDurLowBalance = time.Duration(5) * time.Second
	eCfg.FSLowBalanceAnnFile = ""
	eCfg.FSEmptyBalanceContext = ""
	eCfg.FSEmptyBalanceAnnFile = ""
	eCfg.FSCdrExtraFields = []*utils.RSRField{}
	eCfg.OsipsListenUdp = "127.0.0.1:2020"
	eCfg.OsipsMiAddr = "127.0.0.1:8020"
	eCfg.OsipsEvSubscInterval = time.Duration(60) * time.Second
	eCfg.OsipsReconnects = 3
	eCfg.DerivedChargers = make(utils.DerivedChargers, 0)
	eCfg.CombinedDerivedChargers = true
	eCfg.HistoryAgentEnabled = false
	eCfg.HistoryServer = utils.INTERNAL
	eCfg.HistoryServerEnabled = false
	eCfg.HistoryDir = "/var/log/cgrates/history"
	eCfg.HistorySaveInterval = time.Duration(1) * time.Second
	eCfg.MailerServer = "localhost:25"
	eCfg.MailerAuthUser = "cgrates"
	eCfg.MailerAuthPass = "CGRateS.org"
	eCfg.MailerFromAddr = "cgr-mailer@localhost.localdomain"
	eCfg.DataFolderPath = "/usr/share/cgrates/"
	if !reflect.DeepEqual(cfg, eCfg) {
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Defaults different than expected!")
	}

}

func TestSanityCheck(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil {
		t.Error("Coud not set defaults: ", errSet.Error())
	}
	if err := cfg.checkConfigSanity(); err != nil {
		t.Error("Invalid defaults: ", err)
	}
	cfg = &CGRConfig{}
	cfg.CDRSStats = utils.INTERNAL
	if err := cfg.checkConfigSanity(); err == nil {
		t.Error("Failed detecting improper CDRStats configuration within CDRS")
	}
	cfg = &CGRConfig{}
	cfg.MediatorStats = utils.INTERNAL
	if err := cfg.checkConfigSanity(); err == nil {
		t.Error("Failed detecting improper CDRStats configuration within Mediator")
	}
	cfg = &CGRConfig{}
	cfg.SMCdrS = utils.INTERNAL
	if err := cfg.checkConfigSanity(); err == nil {
		t.Error("Failed detecting improper CDRS configuration within SM")
	}
}

// Load config from file and make sure we have all set
func TestConfigFromFile(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err := NewCGRConfigFromFile(&cfgPth)
	if err != nil {
		t.Log(fmt.Sprintf("Could not parse config: %s!", err))
		t.FailNow()
	}
	eCfg := &CGRConfig{} // Instance we expect to get out after reading config file
	eCfg.setDefaults()
	eCfg.RatingDBType = "test"
	eCfg.RatingDBHost = "test"
	eCfg.RatingDBPort = "test"
	eCfg.RatingDBName = "test"
	eCfg.RatingDBUser = "test"
	eCfg.RatingDBPass = "test"
	eCfg.AccountDBType = "test"
	eCfg.AccountDBHost = "test"
	eCfg.AccountDBPort = "test"
	eCfg.AccountDBName = "test"
	eCfg.AccountDBUser = "test"
	eCfg.AccountDBPass = "test"
	eCfg.StorDBType = "test"
	eCfg.StorDBHost = "test"
	eCfg.StorDBPort = "test"
	eCfg.StorDBName = "test"
	eCfg.StorDBUser = "test"
	eCfg.StorDBPass = "test"
	eCfg.StorDBMaxOpenConns = 99
	eCfg.StorDBMaxIdleConns = 99
	eCfg.DBDataEncoding = "test"
	eCfg.RPCJSONListen = "test"
	eCfg.RPCGOBListen = "test"
	eCfg.HTTPListen = "test"
	eCfg.DefaultReqType = "test"
	eCfg.DefaultCategory = "test"
	eCfg.DefaultTenant = "test"
	eCfg.DefaultSubject = "test"
	eCfg.RoundingDecimals = 99
	eCfg.HttpSkipTlsVerify = true
	eCfg.TpExportPath = "test"
	eCfg.RaterEnabled = true
	eCfg.RaterBalancer = "test"
	eCfg.BalancerEnabled = true
	eCfg.SchedulerEnabled = true
	eCfg.CDRSEnabled = true
	eCfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "test"}}
	eCfg.CDRSMediator = "test"
	eCfg.CDRSStats = "test"
	eCfg.CDRSStoreDisable = true
	eCfg.CDRStatsEnabled = true
	eCfg.CDRStatConfig = &CdrStatsConfig{Id: utils.META_DEFAULT, QueueLength: 99, TimeWindow: time.Duration(99) * time.Second,
		Metrics: []string{"test"}, TORs: []string{"test"}, CdrHosts: []string{"test"}, CdrSources: []string{"test"}, ReqTypes: []string{"test"}, Directions: []string{"test"},
		Tenants: []string{"test"}, Categories: []string{"test"}, Accounts: []string{"test"}, Subjects: []string{"test"}, DestinationPrefixes: []string{"test"},
		UsageInterval:   []time.Duration{time.Duration(99) * time.Second},
		MediationRunIds: []string{"test"}, RatedAccounts: []string{"test"}, RatedSubjects: []string{"test"}, CostInterval: []float64{99.0}}
	eCfg.CDRSStats = "test"
	eCfg.CdreDefaultInstance = &CdreConfig{
		CdrFormat:               "test",
		FieldSeparator:          utils.CSV_SEP,
		DataUsageMultiplyFactor: 99.0,
		CostMultiplyFactor:      99.0,
		CostRoundingDecimals:    99,
		CostShiftDigits:         99,
		MaskDestId:              "test",
		MaskLength:              99,
		ExportDir:               "test"}
	eCfg.CdreDefaultInstance.ContentFields, _ = NewCfgCdrFieldsFromIds(false, "test")
	cdrcCfg := NewDefaultCdrcConfig()
	cdrcCfg.Enabled = true
	cdrcCfg.CdrsAddress = "test"
	cdrcCfg.RunDelay = time.Duration(99) * time.Second
	cdrcCfg.CdrFormat = "test"
	cdrcCfg.FieldSeparator = ";"
	cdrcCfg.DataUsageMultiplyFactor = 99.0
	cdrcCfg.CdrInDir = "test"
	cdrcCfg.CdrOutDir = "test"
	cdrcCfg.CdrSourceId = "test"
	cdrcCfg.CdrFields = []*CfgCdrField{
		&CfgCdrField{Tag: utils.TOR, Type: utils.CDRFIELD, CdrFieldId: utils.TOR, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.ACCID, Type: utils.CDRFIELD, CdrFieldId: utils.ACCID, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.REQTYPE, Type: utils.CDRFIELD, CdrFieldId: utils.REQTYPE, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.DIRECTION, Type: utils.CDRFIELD, CdrFieldId: utils.DIRECTION, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.TENANT, Type: utils.CDRFIELD, CdrFieldId: utils.TENANT, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.CATEGORY, Type: utils.CDRFIELD, CdrFieldId: utils.CATEGORY, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.ACCOUNT, Type: utils.CDRFIELD, CdrFieldId: utils.ACCOUNT, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.SUBJECT, Type: utils.CDRFIELD, CdrFieldId: utils.SUBJECT, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.DESTINATION, Type: utils.CDRFIELD, CdrFieldId: utils.DESTINATION, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: utils.SETUP_TIME, Type: utils.CDRFIELD, CdrFieldId: utils.SETUP_TIME, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true, Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: utils.ANSWER_TIME, Type: utils.CDRFIELD, CdrFieldId: utils.ANSWER_TIME, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true, Layout: "2006-01-02T15:04:05Z07:00"},
		&CfgCdrField{Tag: utils.USAGE, Type: utils.CDRFIELD, CdrFieldId: utils.USAGE, Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}, Mandatory: true},
		&CfgCdrField{Tag: "test", Type: utils.CDRFIELD, CdrFieldId: "test", Value: []*utils.RSRField{&utils.RSRField{Id: "test"}}},
	}
	eCfg.CdrcInstances = []*CdrcConfig{cdrcCfg}
	eCfg.MediatorEnabled = true
	eCfg.MediatorRater = "test"
	eCfg.MediatorReconnects = 99
	eCfg.MediatorStats = "test"
	eCfg.MediatorStoreDisable = true
	eCfg.SMEnabled = true
	eCfg.SMSwitchType = "test"
	eCfg.SMRater = "test"
	eCfg.SMCdrS = "test"
	eCfg.SMReconnects = 99
	eCfg.SMDebitInterval = 99
	eCfg.SMMinCallDuration = time.Duration(98) * time.Second
	eCfg.SMMaxCallDuration = time.Duration(99) * time.Second
	eCfg.FreeswitchServer = "test"
	eCfg.FreeswitchPass = "test"
	eCfg.FreeswitchReconnects = 99
	eCfg.FSMinDurLowBalance = time.Duration(99) * time.Second
	eCfg.FSLowBalanceAnnFile = "test"
	eCfg.FSEmptyBalanceContext = "test"
	eCfg.FSEmptyBalanceAnnFile = "test"
	eCfg.FSCdrExtraFields = []*utils.RSRField{&utils.RSRField{Id: "test"}}
	eCfg.OsipsListenUdp = "test"
	eCfg.OsipsMiAddr = "test"
	eCfg.OsipsEvSubscInterval = time.Duration(99) * time.Second
	eCfg.OsipsReconnects = 99
	eCfg.DerivedChargers = utils.DerivedChargers{&utils.DerivedCharger{RunId: "test", RunFilters: "", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}
	eCfg.CombinedDerivedChargers = true
	eCfg.HistoryAgentEnabled = true
	eCfg.HistoryServer = "test"
	eCfg.HistoryServerEnabled = true
	eCfg.HistoryDir = "test"
	eCfg.HistorySaveInterval = time.Duration(99) * time.Second
	eCfg.MailerServer = "test"
	eCfg.MailerAuthUser = "test"
	eCfg.MailerAuthPass = "test"
	eCfg.MailerFromAddr = "test"
	eCfg.DataFolderPath = "/usr/share/cgrates/"
	if !reflect.DeepEqual(cfg, eCfg) {
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Loading of configuration from file failed!")
	}

}

func TestCdrsExtraFields(t *testing.T) {
	eFieldsCfg := []byte(`[cdrs]
extra_fields = extr1,extr2
`)
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CDRSExtraFields, []*utils.RSRField{&utils.RSRField{Id: "extr1"}, &utils.RSRField{Id: "extr2"}}) {
		t.Errorf("Unexpected value for CdrsExtraFields: %v", cfg.CDRSExtraFields)
	}
	eFieldsCfg = []byte(`[cdrs]
extra_fields = ~effective_caller_id_number:s/(\d+)/+$1/
`)
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CDRSExtraFields, []*utils.RSRField{&utils.RSRField{Id: "effective_caller_id_number",
		RSRules: []*utils.ReSearchReplace{&utils.ReSearchReplace{SearchRegexp: regexp.MustCompile(`(\d+)`), ReplaceTemplate: "+$1"}}}}) {
		t.Errorf("Unexpected value for config CdrsExtraFields: %v", cfg.CDRSExtraFields)
	}
	eFieldsCfg = []byte(`[cdrs]
extra_fields = extr1,~extr2:s/x.+/
`)
	if _, err := NewCGRConfigFromBytes(eFieldsCfg); err == nil {
		t.Error("Failed to detect failed RSRParsing")
	}

}

func TestCdreExtraFields(t *testing.T) {
	eFieldsCfg := []byte(`[cdre]
cdr_format = csv
export_template = cgrid,mediation_runid,accid
`)
	expectedFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "cgrid", Type: utils.CDRFIELD, CdrFieldId: "cgrid", Value: []*utils.RSRField{&utils.RSRField{Id: "cgrid"}}, Mandatory: true},
		&CfgCdrField{Tag: "mediation_runid", Type: utils.CDRFIELD, CdrFieldId: "mediation_runid", Value: []*utils.RSRField{&utils.RSRField{Id: "mediation_runid"}}, Mandatory: true},
		&CfgCdrField{Tag: "accid", Type: utils.CDRFIELD, CdrFieldId: "accid", Value: []*utils.RSRField{&utils.RSRField{Id: "accid"}}, Mandatory: true},
	}
	expCdreCfg := &CdreConfig{CdrFormat: utils.CSV, FieldSeparator: utils.CSV_SEP, CostRoundingDecimals: -1, ExportDir: "/var/log/cgrates/cdre", ContentFields: expectedFlds}
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CdreDefaultInstance, expCdreCfg) {
		t.Errorf("Expecting: %v, received: %v", expCdreCfg, cfg.CdreDefaultInstance)
	}
	eFieldsCfg = []byte(`[cdre]
cdr_format = csv
export_template = cgrid,~effective_caller_id_number:s/(\d+)/+$1/
`)
	rsrField, _ := utils.NewRSRField(`~effective_caller_id_number:s/(\d+)/+$1/`)
	expectedFlds = []*CfgCdrField{
		&CfgCdrField{Tag: "cgrid", Type: utils.CDRFIELD, CdrFieldId: "cgrid", Value: []*utils.RSRField{&utils.RSRField{Id: "cgrid"}}, Mandatory: true},
		&CfgCdrField{Tag: "effective_caller_id_number", Type: utils.CDRFIELD, CdrFieldId: "effective_caller_id_number", Value: []*utils.RSRField{rsrField}, Mandatory: false}}
	expCdreCfg.ContentFields = expectedFlds
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CdreDefaultInstance, expCdreCfg) {
		t.Errorf("Expecting: %v, received: %v", expCdreCfg, cfg.CdreDefaultInstance)
	}
	eFieldsCfg = []byte(`[cdre]
cdr_format = csv
export_template = cgrid,~accid:s/(\d)/$1,runid
`)
	if _, err := NewCGRConfigFromBytes(eFieldsCfg); err == nil {
		t.Error("Failed to detect failed RSRParsing")
	}
}
