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

func TestRalsCfgFromJsonCfg(t *testing.T) {
	cfgJSON := &RalsJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Thresholds_conns:           &[]string{utils.MetaInternal},
		Stats_conns:                &[]string{utils.MetaInternal},
		CacheS_conns:               &[]string{utils.MetaInternal},
		Rp_subject_prefix_matching: utils.BoolPointer(true),
		Remove_expired:             utils.BoolPointer(true),
		Max_computed_usage: &map[string]string{
			utils.ANY:   "189h0m0s",
			utils.VOICE: "72h0m0s",
			utils.DATA:  "107374182400",
			utils.SMS:   "5000",
			utils.MMS:   "10000",
		},
		Max_increments: utils.IntPointer(1000000),
		Balance_rating_subject: &map[string]string{
			utils.META_ANY:   "*zero1ns",
			utils.META_VOICE: "*zero1s",
		},
		Dynaprepaid_actionplans: &[]string{},
	}
	expected := &RalsCfg{
		Enabled:                 true,
		ThresholdSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		StatSConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)},
		CacheSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		RpSubjectPrefixMatching: true,
		RemoveExpired:           true,
		MaxComputedUsage: map[string]time.Duration{
			utils.ANY:   time.Duration(189 * time.Hour),
			utils.VOICE: time.Duration(72 * time.Hour),
			utils.DATA:  time.Duration(107374182400),
			utils.SMS:   time.Duration(5000),
			utils.MMS:   time.Duration(10000),
		},
		MaxIncrements: 1000000,
		BalanceRatingSubject: map[string]string{
			utils.META_ANY:   "*zero1ns",
			utils.META_VOICE: "*zero1s",
		},
		DynaprepaidActionPlans: []string{},
	}
	if cfgJson, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = cfgJson.ralsCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfgJson.ralsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfgJson.ralsCfg))
	}
}

func TestRalsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"rals": {
        "enabled": true,						
	    "thresholds_conns": ["*internal"],					
	    "stats_conns": ["*conn1","*conn2"],						
	    "users_conns": ["*internal"],						
	    "rp_subject_prefix_matching": true,	
	    "max_computed_usage": {					// do not compute usage higher than this, prevents memory overload
		   "*voice": "48h",
		   "*sms": "5000"
        }, 
    },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 true,
		utils.ThresholdSConnsCfg:         []string{"*internal"},
		utils.StatSConnsCfg:              []string{"*conn1", "*conn2"},
		utils.CachesConnsCfg:             []string{"*internal"},
		utils.RpSubjectPrefixMatchingCfg: true,
		utils.RemoveExpiredCfg:           true,
		utils.MaxComputedUsageCfg: map[string]interface{}{
			"*any":   "189h0m0s",
			"*voice": "48h0m0s",
			"*data":  "107374182400",
			"*sms":   "5000",
			"*mms":   "10000",
		},
		utils.MaxIncrementsCfg: 1000000,
		utils.BalanceRatingSubjectCfg: map[string]interface{}{
			"*any":   "*zero1ns",
			"*voice": "*zero1s",
		},
		utils.Dynaprepaid_actionplansCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.ralsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRalsCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
     "rals": {}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 false,
		utils.ThresholdSConnsCfg:         []string{},
		utils.StatSConnsCfg:              []string{},
		utils.CachesConnsCfg:             []string{"*internal"},
		utils.RpSubjectPrefixMatchingCfg: false,
		utils.RemoveExpiredCfg:           true,
		utils.MaxComputedUsageCfg: map[string]interface{}{
			"*any":   "189h0m0s",
			"*voice": "72h0m0s",
			"*data":  "107374182400",
			"*sms":   "10000",
			"*mms":   "10000",
		},
		utils.MaxIncrementsCfg: 1000000,
		utils.BalanceRatingSubjectCfg: map[string]interface{}{
			"*any":   "*zero1ns",
			"*voice": "*zero1s",
		},
		utils.Dynaprepaid_actionplansCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.ralsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
