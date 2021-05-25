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
			utils.MetaAny:   "189h0m0s",
			utils.MetaVoice: "72h0m0s",
			utils.MetaData:  "107374182400",
			utils.MetaSMS:   "5000",
			utils.MetaMMS:   "10000",
		},
		Max_increments: utils.IntPointer(1000000),
		Balance_rating_subject: &map[string]string{
			utils.MetaAny:   "*zero1ns",
			utils.MetaVoice: "*zero1s",
		},
	}
	expected := &RalsCfg{
		Enabled:                 true,
		ThresholdSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		CacheSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		RpSubjectPrefixMatching: true,
		RemoveExpired:           true,
		MaxComputedUsage: map[string]time.Duration{
			utils.MetaAny:   189 * time.Hour,
			utils.MetaVoice: 72 * time.Hour,
			utils.MetaData:  107374182400,
			utils.MetaSMS:   5000,
			utils.MetaMMS:   10000,
		},
		MaxIncrements: 1000000,
		BalanceRatingSubject: map[string]string{
			utils.MetaAny:   "*zero1ns",
			utils.MetaVoice: "*zero1s",
		},
	}
	cfg := NewDefaultCGRConfig()
	if err = cfg.ralsCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.ralsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.ralsCfg))
	}
}

func TestRalsCfgFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &RalsJsonCfg{
		Max_computed_usage: &map[string]string{
			utils.MetaAny: "189hh",
		},
	}
	expected := "time: unknown unit \"hh\" in duration \"189hh\""
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.ralsCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
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
		utils.BalanceRatingSubjectCfg: map[string]string{
			"*any":   "*zero1ns",
			"*voice": "*zero1s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
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
		utils.BalanceRatingSubjectCfg: map[string]string{
			"*any":   "*zero1ns",
			"*voice": "*zero1s",
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.ralsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRalsCfgClone(t *testing.T) {
	ban := &RalsCfg{
		Enabled:                 true,
		ThresholdSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:              []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		CacheSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		RpSubjectPrefixMatching: true,
		RemoveExpired:           true,
		MaxComputedUsage: map[string]time.Duration{
			utils.MetaAny:   189 * time.Hour,
			utils.MetaVoice: 72 * time.Hour,
			utils.MetaData:  107374182400,
			utils.MetaSMS:   5000,
			utils.MetaMMS:   10000,
		},
		MaxIncrements: 1000000,
		BalanceRatingSubject: map[string]string{
			utils.MetaAny:   "*zero1ns",
			utils.MetaVoice: "*zero1s",
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ThresholdSConns[1] = ""; ban.ThresholdSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.CacheSConns[1] = ""; ban.CacheSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.MaxComputedUsage[utils.MetaAny] = 0; ban.MaxComputedUsage[utils.MetaAny] != 189*time.Hour {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.BalanceRatingSubject[utils.MetaAny] = ""; ban.BalanceRatingSubject[utils.MetaAny] != "*zero1ns" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
