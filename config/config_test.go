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

	"github.com/cgrates/cgrates/utils"
)

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
	eCfg.AccountDBName = "10"
	eCfg.AccountDBUser = ""
	eCfg.AccountDBPass = ""
	eCfg.StorDBType = utils.MYSQL
	eCfg.StorDBHost = "localhost"
	eCfg.StorDBPort = "3306"
	eCfg.StorDBName = "cgrates"
	eCfg.StorDBUser = "cgrates"
	eCfg.StorDBPass = "CGRateS.org"
	eCfg.DBDataEncoding = utils.MSGPACK
	eCfg.RPCEncoding = JSON
	eCfg.DefaultReqType = utils.RATED
	eCfg.DefaultTOR = "0"
	eCfg.DefaultTenant = "0"
	eCfg.DefaultSubject = "0"
	eCfg.RoundingMethod = utils.ROUNDING_MIDDLE
	eCfg.RoundingDecimals = 4
	eCfg.RaterEnabled = false
	eCfg.RaterBalancer = DISABLED
	eCfg.RaterListen = "127.0.0.1:2012"
	eCfg.BalancerEnabled = false
	eCfg.BalancerListen = "127.0.0.1:2013"
	eCfg.SchedulerEnabled = false
	eCfg.CDRSEnabled = false
	eCfg.CDRSListen = "127.0.0.1:2022"
	eCfg.CDRSExtraFields = []string{}
	eCfg.CDRSMediator = ""
	eCfg.CDRSExportPath = "/var/log/cgrates/cdr/out"
	eCfg.CDRSExportExtraFields = []string{}
	eCfg.MediatorEnabled = false
	eCfg.MediatorListen = "127.0.0.1:2032"
	eCfg.MediatorRater = "127.0.0.1:2012"
	eCfg.MediatorRaterReconnects = 3
	eCfg.MediatorCDRType = "freeswitch_http_json"
	eCfg.MediatorAccIdField = "accid"
	eCfg.MediatorSubjectFields = []string{"subject"}
	eCfg.MediatorReqTypeFields = []string{"reqtype"}
	eCfg.MediatorDirectionFields = []string{"direction"}
	eCfg.MediatorTenantFields = []string{"tenant"}
	eCfg.MediatorTORFields = []string{"tor"}
	eCfg.MediatorAccountFields = []string{"account"}
	eCfg.MediatorDestFields = []string{"destination"}
	eCfg.MediatorTimeAnswerFields = []string{"time_answer"}
	eCfg.MediatorDurationFields = []string{"duration"}
	eCfg.MediatorCDRInDir = "/var/log/freeswitch/cdr-csv"
	eCfg.MediatorCDROutDir = "/var/log/cgrates/cdr/out/freeswitch/csv"
	eCfg.SMEnabled = false
	eCfg.SMSwitchType = FS
	eCfg.SMRater = "127.0.0.1:2012"
	eCfg.SMRaterReconnects = 3
	eCfg.SMDebitInterval = 10
	eCfg.FreeswitchServer = "127.0.0.1:8021"
	eCfg.FreeswitchPass = "ClueCon"
	eCfg.FreeswitchReconnects = 5
	eCfg.HistoryAgentEnabled = false
	eCfg.HistoryServer = "127.0.0.1:2013"
	eCfg.HistoryServerEnabled = false
	eCfg.HistoryListen = "127.0.0.1:2013"
	eCfg.HistoryPath = "/var/log/cgrates/history"
	eCfg.HistorySavePeriod = "1s"
	if !reflect.DeepEqual(cfg, eCfg) {
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Defaults different than expected!")
	}
}

// Make sure defaults did not change
func TestDefaultsSanity(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	if (cfg.RaterListen != INTERNAL &&
		(cfg.RaterListen == cfg.BalancerListen ||
			cfg.RaterListen == cfg.CDRSListen ||
			cfg.RaterListen == cfg.MediatorListen)) ||
		(cfg.BalancerListen != INTERNAL && (cfg.BalancerListen == cfg.CDRSListen ||
			cfg.BalancerListen == cfg.MediatorListen)) ||
		(cfg.CDRSListen != INTERNAL && cfg.CDRSListen == cfg.MediatorListen) {
		t.Error("Listen defaults on the same port!")
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
	eCfg.RPCEncoding = "test"
	eCfg.DefaultReqType = "test"
	eCfg.DefaultTOR = "test"
	eCfg.DefaultTenant = "test"
	eCfg.DefaultSubject = "test"
	eCfg.RoundingMethod = "test"
	eCfg.RoundingDecimals = 99
	eCfg.RaterEnabled = true
	eCfg.RaterBalancer = "test"
	eCfg.RaterListen = "test"
	eCfg.BalancerEnabled = true
	eCfg.BalancerListen = "test"
	eCfg.SchedulerEnabled = true
	eCfg.CDRSEnabled = true
	eCfg.CDRSListen = "test"
	eCfg.CDRSExtraFields = []string{"test"}
	eCfg.CDRSMediator = "test"
	eCfg.CDRSExportPath = "test"
	eCfg.CDRSExportExtraFields = []string{"test"}
	eCfg.MediatorEnabled = true
	eCfg.MediatorListen = "test"
	eCfg.MediatorRater = "test"
	eCfg.MediatorRaterReconnects = 99
	eCfg.MediatorCDRType = "test"
	eCfg.MediatorAccIdField = "test"
	eCfg.MediatorSubjectFields = []string{"test"}
	eCfg.MediatorReqTypeFields = []string{"test"}
	eCfg.MediatorDirectionFields = []string{"test"}
	eCfg.MediatorTenantFields = []string{"test"}
	eCfg.MediatorTORFields = []string{"test"}
	eCfg.MediatorAccountFields = []string{"test"}
	eCfg.MediatorDestFields = []string{"test"}
	eCfg.MediatorTimeAnswerFields = []string{"test"}
	eCfg.MediatorDurationFields = []string{"test"}
	eCfg.MediatorCDRInDir = "test"
	eCfg.MediatorCDROutDir = "test"
	eCfg.SMEnabled = true
	eCfg.SMSwitchType = "test"
	eCfg.SMRater = "test"
	eCfg.SMRaterReconnects = 99
	eCfg.SMDebitInterval = 99
	eCfg.FreeswitchServer = "test"
	eCfg.FreeswitchPass = "test"
	eCfg.FreeswitchReconnects = 99
	eCfg.HistoryAgentEnabled = true
	eCfg.HistoryServer = "test"
	eCfg.HistoryServerEnabled = true
	eCfg.HistoryListen = "test"
	eCfg.HistoryPath = "test"
	eCfg.HistorySavePeriod = "test"
	if !reflect.DeepEqual(cfg, eCfg) {
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Loading of configuration from file failed!")
	}
}
