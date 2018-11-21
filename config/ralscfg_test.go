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
	"strings"
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
	"pubsubs_conns": [],					// address where to reach the pubusb service, empty to disable pubsub functionality: <""|*internal|x.y.z.y:1234>
	"users_conns": [],						// address where to reach the user service, empty to disable user profile functionality: <""|*internal|x.y.z.y:1234>
	"aliases_conns": [],					// address where to reach the aliases service, empty to disable aliases functionality: <""|*internal|x.y.z.y:1234>
	"rp_subject_prefix_matching": false,	// enables prefix matching for the rating profile subject
	"lcr_subject_prefix_matching": false,	// enables prefix matching for the lcr subject
	"max_computed_usage": {					// do not compute usage higher than this, prevents memory overload
		"*any": "189h",
		"*voice": "72h",
		"*data": "107374182400",
		"*sms": "10000"
	},
},
}`
	ralscfg.RALsMaxComputedUsage = make(map[string]time.Duration)
	expected = RalsCfg{
		RALsEnabled:              false,
		RALsThresholdSConns:      []*HaPoolConfig{},
		RALsStatSConns:           []*HaPoolConfig{},
		RALsPubSubSConns:         []*HaPoolConfig{},
		RALsUserSConns:           []*HaPoolConfig{},
		RALsAliasSConns:          []*HaPoolConfig{},
		RpSubjectPrefixMatching:  false,
		LcrSubjectPrefixMatching: false,
		RALsMaxComputedUsage: map[string]time.Duration{
			utils.ANY:   time.Duration(189 * time.Hour),
			utils.VOICE: time.Duration(72 * time.Hour),
			utils.DATA:  time.Duration(107374182400),
			utils.SMS:   time.Duration(10000),
		},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRalsCfg, err := jsnCfg.RalsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = ralscfg.loadFromJsonCfg(jsnRalsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, ralscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(ralscfg))
	}
}
