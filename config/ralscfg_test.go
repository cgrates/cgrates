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

func TestRalsCfgFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &RalsJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Thresholds_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:                &[]string{utils.MetaInternal, "*conn1"},
		CacheS_conns:               &[]string{utils.MetaInternal, "*conn1"},
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
		Dynaprepaid_actionplans: &[]string{"randomPlans"},
	}
	expected := &RalsCfg{
		Enabled:                 true,
		ThresholdSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS), "*conn1"},
		CacheSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
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
		DynaprepaidActionPlans: []string{"randomPlans"},
	}
	if cfgJson, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = cfgJson.ralsCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfgJson.ralsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfgJson.ralsCfg))
	}
}

func TestRalsCfgFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &RalsJsonCfg{
		Max_computed_usage: &map[string]string{
			utils.ANY: "189hh",
		},
	}
	expected := "time: unknown unit \"hh\" in duration \"189hh\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.ralsCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRalsCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
	"rals": {
        "enabled": true,						
	    "thresholds_conns": ["*internal:*thresholds", "*conn1"],					
	    "caches_conns": ["*internal:*caches", "*conn1"],						
	    "stats_conns": ["*internal:*stats", "*conn1"],						
	    "rp_subject_prefix_matching": true,	
	    "max_computed_usage": {					// do not compute usage higher than this, prevents memory overload
		   "*voice": "48h",
		   "*sms": "5000"
        }, 
    },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 true,
		utils.ThresholdSConnsCfg:         []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:              []string{utils.MetaInternal, "*conn1"},
		utils.CachesConnsCfg:             []string{utils.MetaInternal, "*conn1"},
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

func TestRalsCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
     "rals": {
           "caches_conns": ["*conn1"],
           "stats_conns": ["*internal:*stats"],
     }
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 false,
		utils.ThresholdSConnsCfg:         []string{},
		utils.StatSConnsCfg:              []string{utils.MetaInternal},
		utils.CachesConnsCfg:             []string{"*conn1"},
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
