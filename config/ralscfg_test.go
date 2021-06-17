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
	var ralscfg, expected RalsCfg
	if err := ralscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ralscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, ralscfg)
	}
	if err := ralscfg.loadFromJsonCfg(new(RalsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ralscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, ralscfg)
	}
	cfgJSONStr := `{
"rals": {
	"enabled": false,						// enable Rater service: <true|false>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"users_conns": [],						// address where to reach the user service, empty to disable user profile functionality: <""|*internal|x.y.z.y:1234>
	"rp_subject_prefix_matching": false,	// enables prefix matching for the rating profile subject
	"max_computed_usage": {					// do not compute usage higher than this, prevents memory overload
		"*any": "189h",
		"*voice": "72h",
		"*data": "107374182400",
		"*sms": "10000"
	},
},
}`
	ralscfg.MaxComputedUsage = make(map[string]time.Duration)
	expected = RalsCfg{
		Enabled:                 false,
		ThresholdSConns:         []string{},
		StatSConns:              []string{},
		RpSubjectPrefixMatching: false,
		MaxComputedUsage: map[string]time.Duration{
			utils.ANY:   time.Duration(189 * time.Hour),
			utils.VOICE: time.Duration(72 * time.Hour),
			utils.DATA:  time.Duration(107374182400),
			utils.SMS:   time.Duration(10000),
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRalsCfg, err := jsnCfg.RalsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = ralscfg.loadFromJsonCfg(jsnRalsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, ralscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(ralscfg))
	}
}

func TestRalsCfgAsMapInterface(t *testing.T) {
	var ralscfg RalsCfg
	ralscfg.BalanceRatingSubject = make(map[string]string)
	ralscfg.MaxComputedUsage = make(map[string]time.Duration)
	cfgJSONStr := `{
	"rals": {
		"enabled": false,						
		"thresholds_conns": [],					
		"stats_conns": [],									
		"rp_subject_prefix_matching": false,	
		"remove_expired":true,					
		"max_computed_usage": {					
			"*any": "189h",
			"*voice": "72h",
			"*data": "107374182400",
			"*sms": "10000",
			"*mms": "10000"
		},
		"max_increments": 1000000,
		"balance_rating_subject":{				
			"*any": "*zero1ns",
			"*voice": "*zero1s"
		},
		"dynaprepaid_actionplans": [],			
	},
}`
	eMap := map[string]interface{}{
		"enabled":                    false,
		"thresholds_conns":           []string{},
		"stats_conns":                []string{},
		"rp_subject_prefix_matching": false,
		"remove_expired":             true,
		"max_computed_usage": map[string]interface{}{
			"*any":   "189h0m0s",
			"*voice": "72h0m0s",
			"*data":  "107374182400",
			"*sms":   "10000",
			"*mms":   "10000",
		},
		"max_increments": 1000000,
		"balance_rating_subject": map[string]interface{}{
			"*any":   "*zero1ns",
			"*voice": "*zero1s",
		},
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRalsCfg, err := jsnCfg.RalsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = ralscfg.loadFromJsonCfg(jsnRalsCfg); err != nil {
		t.Error(err)
	} else if rcv := ralscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v ,\n recived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
