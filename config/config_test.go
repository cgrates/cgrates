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
	"reflect"
)

// Make sure defaults did not change by mistake
func TestDefaults(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil  {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	eCfg := &CGRConfig{}
	eCfg.DataDBType = REDIS 
	eCfg.DataDBHost = "127.0.0.1" 
	eCfg.DataDBPort = "6379" 
	eCfg.DataDBName = "10" 
	eCfg.DataDBUser = "" 
	eCfg.DataDBPass = "" 
	eCfg.LogDBType = MONGO 
	eCfg.LogDBHost = "localhost" 
	eCfg.LogDBPort = "27017" 
	eCfg.LogDBName = "cgrates" 
	eCfg.LogDBUser = "" 
	eCfg.LogDBPass = "" 
	eCfg.RPCEncoding = GOB 
	eCfg.DefaultReqType = "" 
	eCfg.DefaultTOR = "0" 
	eCfg.DefaultTenant = "0" 
	eCfg.DefaultSubject = "0" 
	eCfg.RaterEnabled = false 
	eCfg.RaterBalancer = DISABLED 
	eCfg.RaterListen = "127.0.0.1:2012" 
	eCfg.BalancerEnabled = false 
	eCfg.BalancerListen = "127.0.0.1:2013" 
	eCfg.SchedulerEnabled = false 
	eCfg.CDRSListen = "127.0.0.1:2022" 
	eCfg.CDRSfsJSONEnabled = false 
	eCfg.CDRSMediator = INTERNAL 
	eCfg.MediatorEnabled = false 
	eCfg.MediatorListen = "127.0.0.1:2032" 
	eCfg.MediatorCDRInDir = "/var/log/freeswitch/cdr-csv" 
	eCfg.MediatorCDROutDir = "/var/log/cgrates/cdr_out" 
	eCfg.MediatorRater = "127.0.0.1:2012" 
	eCfg.MediatorRaterReconnects = 3 
	eCfg.MediatorPseudoprepaid = false 
	eCfg.MediatorCDRType = "freeswitch_csv" 
	eCfg.SMEnabled = false 
	eCfg.SMSwitchType = FS 
	eCfg.SMRater = "127.0.0.1:2012" 
	eCfg.SMRaterReconnects = 3 
	eCfg.SMDebitInterval = 10 
	eCfg.FreeswitchServer = "127.0.0.1:8021" 
	eCfg.FreeswitchPass = "ClueCon" 
	eCfg.FreeswitchReconnects = 5 
	eCfg.FreeswitchUUIDIdx = "10" 
	eCfg.FreeswitchTORIdx = "-1" 
	eCfg.FreeswitchTenantIdx = "-1" 
	eCfg.FreeswitchDirectionIdx = "-1" 
	eCfg.FreeswitchSubjectIdx = "-1" 
	eCfg.FreeswitchAccountIdx = "-1" 
	eCfg.FreeswitchDestIdx = "-1" 
	eCfg.FreeswitchTimeStartIdx = "-1" 
	eCfg.FreeswitchDurationIdx = "-1"
	if !reflect.DeepEqual(cfg ,eCfg ){ 
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Defaults different than expected!")
	}
}

// Make sure defaults did not change
func TestDefaultsSanity(t *testing.T) {
	cfg := &CGRConfig{}
	errSet := cfg.setDefaults()
	if errSet != nil  {
		t.Log(fmt.Sprintf("Coud not set defaults: %s!", errSet.Error()))
		t.FailNow()
	}
	if (cfg.RaterListen != INTERNAL && 
		(cfg.RaterListen == cfg.BalancerListen || 
			cfg.RaterListen == cfg.CDRSListen ||
			cfg.RaterListen == cfg.MediatorListen )) ||
		(cfg.BalancerListen != INTERNAL && (cfg.BalancerListen == cfg.CDRSListen ||
			cfg.BalancerListen == cfg.MediatorListen ))||
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
	eCfg.DataDBType = "test"
	eCfg.DataDBHost = "test"
	eCfg.DataDBPort = "test"
	eCfg.DataDBName = "test"
	eCfg.DataDBUser = "test"
	eCfg.DataDBPass = "test"
	eCfg.LogDBType = "test"
	eCfg.LogDBHost = "test"
	eCfg.LogDBPort = "test"
	eCfg.LogDBName = "test"
	eCfg.LogDBUser = "test"
	eCfg.LogDBPass = "test"
	eCfg.RPCEncoding = "test"
	eCfg.DefaultReqType = "test"
	eCfg.DefaultTOR = "test"
	eCfg.DefaultTenant = "test"
	eCfg.DefaultSubject = "test"
	eCfg.RaterEnabled = true
	eCfg.RaterBalancer = "test"
	eCfg.RaterListen = "test"
	eCfg.BalancerEnabled = true
	eCfg.BalancerListen = "test"
	eCfg.SchedulerEnabled = true
	eCfg.CDRSListen = "test"
	eCfg.CDRSfsJSONEnabled = true
	eCfg.CDRSMediator = "test"
	eCfg.MediatorEnabled = true
	eCfg.MediatorListen = "test"
	eCfg.MediatorCDRInDir = "test"
	eCfg.MediatorCDROutDir = "test"
	eCfg.MediatorRater = "test"
	eCfg.MediatorRaterReconnects = 99
	eCfg.MediatorPseudoprepaid = true
	eCfg.MediatorCDRType = "test"
	eCfg.SMEnabled = true
	eCfg.SMSwitchType = "test"
	eCfg.SMRater = "test"
	eCfg.SMRaterReconnects = 99
	eCfg.SMDebitInterval = 99
	eCfg.FreeswitchServer = "test"
	eCfg.FreeswitchPass = "test"
	eCfg.FreeswitchReconnects = 99
	eCfg.FreeswitchUUIDIdx = "test"
	eCfg.FreeswitchTORIdx = "test"
	eCfg.FreeswitchTenantIdx = "test"
	eCfg.FreeswitchDirectionIdx = "test"
	eCfg.FreeswitchSubjectIdx = "test"
	eCfg.FreeswitchAccountIdx = "test"
	eCfg.FreeswitchDestIdx = "test"
	eCfg.FreeswitchTimeStartIdx = "test"
	eCfg.FreeswitchDurationIdx = "test"
	if !reflect.DeepEqual(cfg ,eCfg ){ 
		t.Log(eCfg)
		t.Log(cfg)
		t.Error("Loading of configuration from file failed!")
	}
}		
