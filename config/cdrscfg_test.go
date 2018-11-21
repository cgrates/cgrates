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

	"github.com/cgrates/cgrates/utils"
)

func TestCdrsCfgloadFromJsonCfg(t *testing.T) {
	var cdrscfg, expected CdrsCfg
	if err := cdrscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrscfg)
	}
	if err := cdrscfg.loadFromJsonCfg(new(CdrsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, cdrscfg)
	}
	cfgJSONStr := `{
"cdrs": {
	"enabled": false,						// start the CDR Server service:  <true|false>
	"extra_fields": [],						// extra fields to store in CDRs for non-generic CDRs
	"store_cdrs": true,						// store cdrs in storDb
	"sessions_cost_retries": 5,				// number of queries to sessions_costs before recalculating CDR
	"chargers_conns": [],					// address where to reach the charger service, empty to disable charger functionality: <""|*internal|x.y.z.y:1234>
	"rals_conns": [
		{"address": "*internal"}			// address where to reach the Rater for cost calculation, empty to disable functionality: <""|*internal|x.y.z.y:1234>
	],
	"pubsubs_conns": [],					// address where to reach the pubusb service, empty to disable pubsub functionality: <""|*internal|x.y.z.y:1234>
	"attributes_conns": [],					// address where to reach the attribute service, empty to disable attributes functionality: <""|*internal|x.y.z.y:1234>
	"users_conns": [],						// address where to reach the user service, empty to disable user profile functionality: <""|*internal|x.y.z.y:1234>
	"aliases_conns": [],					// address where to reach the aliases service, empty to disable aliases functionality: <""|*internal|x.y.z.y:1234>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"online_cdr_exports":[],				// list of CDRE profiles to use for real-time CDR exports
	},
}`
	expected = CdrsCfg{
		CDRSStoreCdrs:       true,
		CDRSSMCostRetries:   5,
		CDRSChargerSConns:   []*HaPoolConfig{},
		CDRSRaterConns:      []*HaPoolConfig{{Address: utils.MetaInternal}},
		CDRSPubSubSConns:    []*HaPoolConfig{},
		CDRSAttributeSConns: []*HaPoolConfig{},
		CDRSUserSConns:      []*HaPoolConfig{},
		CDRSAliaseSConns:    []*HaPoolConfig{},
		CDRSThresholdSConns: []*HaPoolConfig{},
		CDRSStatSConns:      []*HaPoolConfig{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cdrscfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(cdrscfg))
	}
}
