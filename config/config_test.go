/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	eCfg.DBDataEncoding = utils.MSGPACK
	eCfg.RPCJSONListen = "127.0.0.1:2012"
	eCfg.RPCGOBListen = "127.0.0.1:2013"
	eCfg.HTTPListen = "127.0.0.1:2080"
	eCfg.DefaultReqType = utils.RATED
	eCfg.DefaultCategory = "call"
	eCfg.DefaultTenant = "cgrates.org"
	eCfg.DefaultSubject = "cgrates"
	eCfg.RoundingMethod = utils.ROUNDING_MIDDLE
	eCfg.RoundingDecimals = 4
	eCfg.XmlCfgDocument = nil
	eCfg.RaterEnabled = false
	eCfg.RaterBalancer = ""
	eCfg.BalancerEnabled = false
	eCfg.SchedulerEnabled = false
	eCfg.CDRSEnabled = false
	eCfg.CDRSExtraFields = []*utils.RSRField{}
	eCfg.CDRSMediator = ""
	eCfg.CdreCdrFormat = "csv"
	eCfg.CdreMaskDestId = ""
	eCfg.CdreMaskLength = 0
	eCfg.CdreCostShiftDigits = 0
	eCfg.CdreDir = "/var/log/cgrates/cdre"
	eCfg.CdrcEnabled = false
	eCfg.CdrcCdrs = utils.INTERNAL
	eCfg.CdrcCdrsMethod = "http_cgr"
	eCfg.CdrcRunDelay = time.Duration(0)
	eCfg.CdrcCdrType = "csv"
	eCfg.CdrcCdrInDir = "/var/log/cgrates/cdrc/in"
	eCfg.CdrcCdrOutDir = "/var/log/cgrates/cdrc/out"
	eCfg.CdrcSourceId = "freeswitch_csv"
	eCfg.CdrcAccIdField = "0"
	eCfg.CdrcReqTypeField = "1"
	eCfg.CdrcDirectionField = "2"
	eCfg.CdrcTenantField = "3"
	eCfg.CdrcCategoryField = "4"
	eCfg.CdrcAccountField = "5"
	eCfg.CdrcSubjectField = "6"
	eCfg.CdrcDestinationField = "7"
	eCfg.CdrcSetupTimeField = "8"
	eCfg.CdrcAnswerTimeField = "9"
	eCfg.CdrcDurationField = "10"
	eCfg.CdrcExtraFields = []string{}
	eCfg.MediatorEnabled = false
	eCfg.MediatorRater = "internal"
	eCfg.MediatorRaterReconnects = 3
	eCfg.SMEnabled = false
	eCfg.SMSwitchType = FS
	eCfg.SMRater = "internal"
	eCfg.SMRaterReconnects = 3
	eCfg.SMDebitInterval = 10
	eCfg.SMMaxCallDuration = time.Duration(3) * time.Hour
	eCfg.FreeswitchServer = "127.0.0.1:8021"
	eCfg.FreeswitchPass = "ClueCon"
	eCfg.FreeswitchReconnects = 5
	eCfg.DerivedChargers = make(utils.DerivedChargers, 0)
	eCfg.CombinedDerivedChargers = true
	eCfg.HistoryAgentEnabled = false
	eCfg.HistoryServer = "internal"
	eCfg.HistoryServerEnabled = false
	eCfg.HistoryDir = "/var/log/cgrates/history"
	eCfg.HistorySaveInterval = time.Duration(1) * time.Second
	eCfg.MailerServer = "localhost:25"
	eCfg.MailerAuthUser = "cgrates"
	eCfg.MailerAuthPass = "CGRateS.org"
	eCfg.MailerFromAddr = "cgr-mailer@localhost.localdomain"
	eCfg.CdreExportedFields = []*utils.RSRField{
		&utils.RSRField{Id: utils.CGRID},
		&utils.RSRField{Id: utils.MEDI_RUNID},
		&utils.RSRField{Id: utils.ACCID},
		&utils.RSRField{Id: utils.CDRHOST},
		&utils.RSRField{Id: utils.REQTYPE},
		&utils.RSRField{Id: utils.DIRECTION},
		&utils.RSRField{Id: utils.TENANT},
		&utils.RSRField{Id: utils.Category},
		&utils.RSRField{Id: utils.ACCOUNT},
		&utils.RSRField{Id: utils.SUBJECT},
		&utils.RSRField{Id: utils.DESTINATION},
		&utils.RSRField{Id: utils.SETUP_TIME},
		&utils.RSRField{Id: utils.ANSWER_TIME},
		&utils.RSRField{Id: utils.DURATION},
		&utils.RSRField{Id: utils.COST},
	}
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
	cfg.CdreCdrFormat = utils.CDRE_FIXED_WIDTH
	if err := cfg.checkConfigSanity(); err == nil {
		t.Error("Failed to detect fixed_width dependency on xml configuration")
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
	eCfg.DBDataEncoding = "test"
	eCfg.RPCJSONListen = "test"
	eCfg.RPCGOBListen = "test"
	eCfg.HTTPListen = "test"
	eCfg.DefaultReqType = "test"
	eCfg.DefaultCategory = "test"
	eCfg.DefaultTenant = "test"
	eCfg.DefaultSubject = "test"
	eCfg.RoundingMethod = "test"
	eCfg.RoundingDecimals = 99
	eCfg.RaterEnabled = true
	eCfg.RaterBalancer = "test"
	eCfg.BalancerEnabled = true
	eCfg.SchedulerEnabled = true
	eCfg.CDRSEnabled = true
	eCfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "test"}}
	eCfg.CDRSMediator = "test"
	eCfg.CdreCdrFormat = "test"
	eCfg.CdreMaskDestId = "test"
	eCfg.CdreMaskLength = 99
	eCfg.CdreCostShiftDigits = 99
	eCfg.CdreExportedFields = []*utils.RSRField{&utils.RSRField{Id: "test"}}
	eCfg.CdreDir = "test"
	eCfg.CdrcEnabled = true
	eCfg.CdrcCdrs = "test"
	eCfg.CdrcCdrsMethod = "test"
	eCfg.CdrcRunDelay = time.Duration(99) * time.Second
	eCfg.CdrcCdrType = "test"
	eCfg.CdrcCdrInDir = "test"
	eCfg.CdrcCdrOutDir = "test"
	eCfg.CdrcSourceId = "test"
	eCfg.CdrcAccIdField = "test"
	eCfg.CdrcReqTypeField = "test"
	eCfg.CdrcDirectionField = "test"
	eCfg.CdrcTenantField = "test"
	eCfg.CdrcCategoryField = "test"
	eCfg.CdrcAccountField = "test"
	eCfg.CdrcSubjectField = "test"
	eCfg.CdrcDestinationField = "test"
	eCfg.CdrcSetupTimeField = "test"
	eCfg.CdrcAnswerTimeField = "test"
	eCfg.CdrcDurationField = "test"
	eCfg.CdrcExtraFields = []string{"test"}
	eCfg.MediatorEnabled = true
	eCfg.MediatorRater = "test"
	eCfg.MediatorRaterReconnects = 99
	eCfg.SMEnabled = true
	eCfg.SMSwitchType = "test"
	eCfg.SMRater = "test"
	eCfg.SMRaterReconnects = 99
	eCfg.SMDebitInterval = 99
	eCfg.SMMaxCallDuration = time.Duration(99) * time.Second
	eCfg.FreeswitchServer = "test"
	eCfg.FreeswitchPass = "test"
	eCfg.FreeswitchReconnects = 99
	eCfg.DerivedChargers = utils.DerivedChargers{&utils.DerivedCharger{RunId: "test", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", DurationField: "test"}}
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
		RSRules: []*utils.ReSearchReplace{&utils.ReSearchReplace{regexp.MustCompile(`(\d+)`), "+$1"}}}}) {
		t.Errorf("Unexpected value for config CdrsExtraFields: %v", cfg.CDRSExtraFields)
	}
	eFieldsCfg = []byte(`[cdrs]
extra_fields = extr1,extr2,
`)
	if _, err := NewCGRConfigFromBytes(eFieldsCfg); err == nil {
		t.Error("Failed to detect empty field in the end of extra fields defition")
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
	if cfg, err := NewCGRConfigFromBytes(eFieldsCfg); err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(cfg.CdreExportedFields, []*utils.RSRField{&utils.RSRField{Id: "cgrid"}, &utils.RSRField{Id: "mediation_runid"}, &utils.RSRField{Id: "accid"}}) {
		t.Errorf("Unexpected value for CdrsExtraFields: %v", cfg.CDRSExtraFields)
	}
	eFieldsCfg = []byte(`[cdre]
cdr_format = csv
export_template = cgrid,mediation_runid,accid,
`)
	if _, err := NewCGRConfigFromBytes(eFieldsCfg); err == nil {
		t.Error("Failed to detect empty field in the end of export_template defition")
	}
	eFieldsCfg = []byte(`[cdre]
cdr_format = csv
export_template = cgrid,~accid:s/(\d)/$1,runid
`)
	if _, err := NewCGRConfigFromBytes(eFieldsCfg); err == nil {
		t.Error("Failed to detect failed RSRParsing")
	}
}
