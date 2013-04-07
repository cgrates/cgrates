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
		cfg.RaterEnabled != true ||
		cfg.RaterBalancer != "test" ||
		cfg.RaterListen != "test" ||
		cfg.RaterRPCEncoding != "test" ||
		cfg.BalancerEnabled != true ||
		cfg.BalancerListen != "test" ||
		cfg.BalancerRPCEncoding != "test" ||
		cfg.SchedulerEnabled != true ||
		cfg.SMEnabled != true ||
		cfg.SMSwitchType != "test" ||
		cfg.SMRater != "test" ||
		cfg.SMDebitPeriod != 11 ||
		cfg.SMRPCEncoding != "test" ||
		cfg.MediatorEnabled != true ||
		cfg.MediatorCDRPath != "test" ||
		cfg.MediatorCDROutPath != "test" ||
		cfg.MediatorRater != "test" ||
		cfg.MediatorRPCEncoding != "test" ||
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
