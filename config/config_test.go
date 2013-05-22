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
	"testing"
)

func TestConfig(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err := NewCGRConfig(&cfgPth)
	if err != nil {
		t.Log(fmt.Sprintf("Could not parse config: %s!", err))
		t.FailNow()
	}
	if cfg.DataDBType != "test" ||
		cfg.DataDBHost != "test" ||
		cfg.DataDBPort != "test" ||
		cfg.DataDBName != "test" ||
		cfg.DataDBUser != "test" ||
		cfg.DataDBPass != "test" ||
		cfg.LogDBType != "test" ||
		cfg.LogDBHost != "test" ||
		cfg.LogDBPort != "test" ||
		cfg.LogDBName != "test" ||
		cfg.LogDBUser != "test" ||
		cfg.LogDBPass != "test" ||
		cfg.RPCEncoding != "test" ||
		cfg.RaterEnabled != true ||
		cfg.RaterBalancer != "test" ||
		cfg.RaterListen != "test" ||
		cfg.BalancerEnabled != true ||
		cfg.BalancerListen != "test" ||
		cfg.SchedulerEnabled != true ||
		cfg.SMEnabled != true ||
		cfg.SMSwitchType != "test" ||
		cfg.SMRater != "test" ||
		cfg.SMDebitInterval != 11 ||
		cfg.MediatorEnabled != true ||
		cfg.MediatorCDRInDir != "test" ||
		cfg.MediatorCDROutDir != "test" ||
		cfg.MediatorRater != "test" ||
		cfg.MediatorSkipDB != true ||
		cfg.MediatorPseudoprepaid != true ||
		cfg.FreeswitchServer != "test" ||
		cfg.FreeswitchPass != "test" ||
		cfg.FreeswitchDirectionIdx != "test" ||
		cfg.FreeswitchTORIdx != "test" ||
		cfg.FreeswitchTenantIdx != "test" ||
		cfg.FreeswitchSubjectIdx != "test" ||
		cfg.FreeswitchAccountIdx != "test" ||
		cfg.FreeswitchDestIdx != "test" ||
		cfg.FreeswitchTimeStartIdx != "test" ||
		cfg.FreeswitchDurationIdx != "test" ||
		cfg.FreeswitchUUIDIdx != "test" {
			t.Log(cfg.DataDBType)
			t.Log(cfg.DataDBHost)
			t.Log(cfg.DataDBPort)
			t.Log(cfg.DataDBName)
			t.Log(cfg.DataDBUser)
			t.Log(cfg.DataDBPass)
			t.Log(cfg.LogDBType)
			t.Log(cfg.LogDBHost)
			t.Log(cfg.LogDBPort)
			t.Log(cfg.LogDBName)
			t.Log(cfg.LogDBUser)
			t.Log(cfg.LogDBPass)
			t.Log(cfg.RPCEncoding)
			t.Log(cfg.RaterEnabled)
			t.Log(cfg.RaterBalancer)
			t.Log(cfg.RaterListen) 
			t.Log(cfg.BalancerEnabled)
			t.Log(cfg.BalancerListen) 
			t.Log(cfg.SchedulerEnabled)
			t.Log(cfg.SMEnabled)
			t.Log(cfg.SMSwitchType) 
			t.Log(cfg.SMRater)
			t.Log(cfg.SMDebitInterval)
			t.Log(cfg.MediatorEnabled)
			t.Log(cfg.MediatorCDRInDir)
			t.Log(cfg.MediatorCDROutDir) 
			t.Log(cfg.MediatorRater)
			t.Log(cfg.MediatorSkipDB) 
			t.Log(cfg.MediatorPseudoprepaid) 
			t.Log(cfg.FreeswitchServer) 
			t.Log(cfg.FreeswitchPass)
			t.Log(cfg.FreeswitchDirectionIdx)
			t.Log(cfg.FreeswitchTORIdx) 
			t.Log(cfg.FreeswitchTenantIdx) 
			t.Log(cfg.FreeswitchSubjectIdx) 
			t.Log(cfg.FreeswitchAccountIdx)
			t.Log(cfg.FreeswitchDestIdx) 
			t.Log(cfg.FreeswitchTimeStartIdx)
			t.Log(cfg.FreeswitchDurationIdx)
			t.Log(cfg.FreeswitchUUIDIdx)
			t.Error("Config file read failed!")
		}
}

func TestParamOverwrite(t *testing.T) {
	cfgPth := "test_data.txt"
	cfg, err := NewCGRConfig(&cfgPth)
	if err != nil {
		t.Log(fmt.Sprintf("Could not parse config: %s!", err))
		t.FailNow()
	}
	if cfg.FreeswitchReconnects != 5 { // one default which is not overwritten in test data
		t.Errorf("FreeswitchReconnects set == %d, expect 5", cfg.FreeswitchReconnects)
	} else if cfg.SchedulerEnabled != true { // one parameter which should be overwritten in test data
		t.Errorf("scheduler_enabled set == %d, expect true", cfg.SchedulerEnabled)
	}
}

// Make sure defaults did not change by mistake
func TestDefaults(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil  {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	if  cfg.DataDBType != REDIS ||
		cfg.DataDBHost != "127.0.0.1" ||
		cfg.DataDBPort != "6379" ||
		cfg.DataDBName != "10" ||
		cfg.DataDBUser != "" ||
		cfg.DataDBPass != "" ||
		cfg.LogDBType != MONGO ||
		cfg.LogDBHost != "localhost" ||
		cfg.LogDBPort != "27017" ||
		cfg.LogDBName != "cgrates" ||
		cfg.LogDBUser != "" ||
		cfg.LogDBPass != "" ||
		cfg.RPCEncoding != GOB ||
		cfg.DefaultTOR != "0" ||
		cfg.DefaultTenant != "0" ||
		cfg.DefaultSubject != "0" ||
		cfg.RaterEnabled != false ||
		cfg.RaterBalancer != DISABLED ||
		cfg.RaterListen != "127.0.0.1:2012" ||
		cfg.BalancerEnabled != false ||
		cfg.BalancerListen != "127.0.0.1:2013" ||
		cfg.SchedulerEnabled != false ||
		cfg.CDRSListen != "127.0.0.1:2022" ||
		cfg.CDRSfsJSONEnabled != false ||
		cfg.MediatorEnabled != false ||
		cfg.MediatorListen != "127.0.0.1:2032" ||
		cfg.MediatorCDRInDir != "/var/log/freeswitch/cdr-csv" ||
		cfg.MediatorCDROutDir != "/var/log/cgrates/cdr_out" ||
		cfg.MediatorRater != "127.0.0.1:2012" ||
		cfg.MediatorRaterReconnects != 3 ||
		cfg.MediatorSkipDB != false ||
		cfg.MediatorPseudoprepaid != false ||
		cfg.MediatorCDRType != "freeswitch_csv" ||
		cfg.SMEnabled != false ||
		cfg.SMSwitchType != FS ||
		cfg.SMRater != "127.0.0.1:2012" ||
		cfg.SMRaterReconnects != 3 ||
		cfg.SMDebitInterval != 10 ||
		cfg.SMDefaultReqType != "" ||
		cfg.FreeswitchServer != "127.0.0.1:8021" ||
		cfg.FreeswitchPass != "ClueCon" ||
		cfg.FreeswitchReconnects != 5 ||
		cfg.FreeswitchUUIDIdx != "10" ||
		cfg.FreeswitchTORIdx != "-1" ||
		cfg.FreeswitchTenantIdx != "-1" ||
		cfg.FreeswitchDirectionIdx != "-1" ||
		cfg.FreeswitchSubjectIdx != "-1" ||
		cfg.FreeswitchAccountIdx != "-1" ||
		cfg.FreeswitchDestIdx != "-1" ||
		cfg.FreeswitchTimeStartIdx != "-1" ||
		cfg.FreeswitchDurationIdx != "-1" {
			t.Error("Defaults different than expected!")
	}

}

// Make sure defaults did not change by mistake
func TestDefaultsSanity(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil  {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	if cfg.RaterListen == cfg.BalancerListen || 
		cfg.RaterListen == cfg.CDRSListen ||
		cfg.RaterListen == cfg.MediatorListen ||
		cfg.BalancerListen == cfg.CDRSListen ||
		cfg.BalancerListen == cfg.MediatorListen ||
		cfg.CDRSListen == cfg.MediatorListen {
			t.Error("Listen defaults on the same port!")
	}
}

		
	
	
	
