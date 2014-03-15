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
	eCfg.DefaultTOR = "call"
	eCfg.DefaultTenant = "cgrates.org"
	eCfg.DefaultSubject = "cgrates"
	eCfg.RoundingMethod = utils.ROUNDING_MIDDLE
	eCfg.RoundingDecimals = 4
	eCfg.RaterEnabled = false
	eCfg.RaterBalancer = ""
	eCfg.BalancerEnabled = false
	eCfg.SchedulerEnabled = false
	eCfg.CDRSEnabled = false
	eCfg.CDRSExtraFields = []*utils.RSRField{}
	eCfg.CDRSMediator = ""
	eCfg.CdreCdrFormat = "csv"
	eCfg.CdreExtraFields = []string{}
	eCfg.CdreDir = "/var/log/cgrates/cdr/cdrexport/csv"
	eCfg.CdrcEnabled = false
	eCfg.CdrcCdrs = utils.INTERNAL
	eCfg.CdrcCdrsMethod = "http_cgr"
	eCfg.CdrcRunDelay = time.Duration(0)
	eCfg.CdrcCdrType = "csv"
	eCfg.CdrcCdrInDir = "/var/log/cgrates/cdr/cdrc/in"
	eCfg.CdrcCdrOutDir = "/var/log/cgrates/cdr/cdrc/out"
	eCfg.CdrcSourceId = "freeswitch_csv"
	eCfg.CdrcAccIdField = "0"
	eCfg.CdrcReqTypeField = "1"
	eCfg.CdrcDirectionField = "2"
	eCfg.CdrcTenantField = "3"
	eCfg.CdrcTorField = "4"
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
	eCfg.MediatorRunIds = []string{}
	eCfg.MediatorSubjectFields = []string{}
	eCfg.MediatorReqTypeFields = []string{}
	eCfg.MediatorDirectionFields = []string{}
	eCfg.MediatorTenantFields = []string{}
	eCfg.MediatorTORFields = []string{}
	eCfg.MediatorAccountFields = []string{}
	eCfg.MediatorDestFields = []string{}
	eCfg.MediatorSetupTimeFields = []string{}
	eCfg.MediatorAnswerTimeFields = []string{}
	eCfg.MediatorDurationFields = []string{}
	eCfg.SMEnabled = false
	eCfg.SMSwitchType = FS
	eCfg.SMRater = "internal"
	eCfg.SMRaterReconnects = 3
	eCfg.SMDebitInterval = 10
	eCfg.SMMaxCallDuration = time.Duration(3) * time.Hour
	eCfg.SMRunIds = []string{}
	eCfg.SMReqTypeFields = []string{}
	eCfg.SMDirectionFields = []string{}
	eCfg.SMTenantFields = []string{}
	eCfg.SMTORFields = []string{}
	eCfg.SMAccountFields = []string{}
	eCfg.SMSubjectFields = []string{}
	eCfg.SMDestFields = []string{}
	eCfg.SMSetupTimeFields = []string{}
	eCfg.SMAnswerTimeFields = []string{}
	eCfg.SMDurationFields = []string{}
	eCfg.FreeswitchServer = "127.0.0.1:8021"
	eCfg.FreeswitchPass = "ClueCon"
	eCfg.FreeswitchReconnects = 5
	eCfg.HistoryAgentEnabled = false
	eCfg.HistoryServer = "internal"
	eCfg.HistoryServerEnabled = false
	eCfg.HistoryDir = "/var/log/cgrates/history"
	eCfg.HistorySaveInterval = time.Duration(1) * time.Second
	eCfg.MailerServer = "localhost:25"
	eCfg.MailerAuthUser = "cgrates"
	eCfg.MailerAuthPass = "CGRateS.org"
	eCfg.MailerFromAddr = "cgr-mailer@localhost.localdomain"
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
	cfg.SMSubjectFields = []string{"sample1", "sample2", "sample3"}
	if err := cfg.checkConfigSanity(); err == nil {
		t.Error("Failed to detect config insanity")
	}
}

// Load config from file and make sure we have all set
func TestConfigFromFile(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err := NewCGRConfig(&cfgPth)
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
	eCfg.DefaultTOR = "test"
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
	eCfg.CdreExtraFields = []string{"test"}
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
	eCfg.CdrcTorField = "test"
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
	eCfg.MediatorRunIds = []string{"test"}
	eCfg.MediatorSubjectFields = []string{"test"}
	eCfg.MediatorReqTypeFields = []string{"test"}
	eCfg.MediatorDirectionFields = []string{"test"}
	eCfg.MediatorTenantFields = []string{"test"}
	eCfg.MediatorTORFields = []string{"test"}
	eCfg.MediatorAccountFields = []string{"test"}
	eCfg.MediatorDestFields = []string{"test"}
	eCfg.MediatorSetupTimeFields = []string{"test"}
	eCfg.MediatorAnswerTimeFields = []string{"test"}
	eCfg.MediatorDurationFields = []string{"test"}
	eCfg.SMEnabled = true
	eCfg.SMSwitchType = "test"
	eCfg.SMRater = "test"
	eCfg.SMRaterReconnects = 99
	eCfg.SMDebitInterval = 99
	eCfg.SMMaxCallDuration = time.Duration(99) * time.Second
	eCfg.SMRunIds = []string{"test"}
	eCfg.SMReqTypeFields = []string{"test"}
	eCfg.SMDirectionFields = []string{"test"}
	eCfg.SMTenantFields = []string{"test"}
	eCfg.SMTORFields = []string{"test"}
	eCfg.SMAccountFields = []string{"test"}
	eCfg.SMSubjectFields = []string{"test"}
	eCfg.SMDestFields = []string{"test"}
	eCfg.SMSetupTimeFields = []string{"test"}
	eCfg.SMAnswerTimeFields = []string{"test"}
	eCfg.SMDurationFields = []string{"test"}
	eCfg.FreeswitchServer = "test"
	eCfg.FreeswitchPass = "test"
	eCfg.FreeswitchReconnects = 99
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
