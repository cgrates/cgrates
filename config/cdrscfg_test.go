/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestCdrsCfgloadFromJsonCfg(t *testing.T) {
	var cdrscfg, expected CdrsCfg
	if err := cdrscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, cdrscfg)
	}
	if err := cdrscfg.loadFromJsonCfg(new(CdrsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cdrscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, cdrscfg)
	}
	cfgJSONStr := `{
"cdrs": {
	"enabled": false,						// start the CDR Server service:  <true|false>
	"extra_fields": [],						// extra fields to store in CDRs for non-generic CDRs
	"store_cdrs": true,						// store cdrs in storDb
	"session_cost_retries": 5,				// number of queries to sessions_costs before recalculating CDR
	"chargers_conns": [],					// address where to reach the charger service, empty to disable charger functionality: <""|*internal|x.y.z.y:1234>
	"rals_conns": ["*internal"],
	"attributes_conns": [],					// address where to reach the attribute service, empty to disable attributes functionality: <""|*internal|x.y.z.y:1234>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"online_cdr_exports":[],				// list of CDRE profiles to use for real-time CDR exports
	},
}`
	expected = CdrsCfg{
		StoreCdrs:       true,
		SMCostRetries:   5,
		ChargerSConns:   []string{},
		RaterConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		AttributeSConns: []string{},
		ThresholdSConns: []string{},
		StatSConns:      []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cdrscfg) {
		t.Errorf("Expected: %+v , received: %+v", utils.ToJSON(expected), utils.ToJSON(cdrscfg))
	}
}

func TestExtraFieldsinAsMapInterface(t *testing.T) {
	var cdrscfg CdrsCfg
	cfgJSONStr := `{
	"cdrs": {
		"enabled": true,
		"extra_fields": ["PayPalAccount", "LCRProfile", "ResourceID"],
		"chargers_conns":["*localhost"],
		"store_cdrs": true,
		"online_cdr_exports": []
	},
	}`
	expectedExtra := []string{"PayPalAccount", "LCRProfile", "ResourceID"}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		t.Error(err)
	} else if rcv := cdrscfg.AsMapInterface(); !reflect.DeepEqual(rcv[utils.ExtraFieldsCfg], expectedExtra) {
		t.Errorf("Expecting: '%+v', received: '%+v' ", expectedExtra, rcv[utils.ExtraFieldsCfg])
	}
}

func TestCdrsCfgAsMapInterface(t *testing.T) {
	var cdrscfg CdrsCfg
	cfgJSONStr := `{
	"cdrs": {
		"enabled": false,						
		"extra_fields": [],
		"store_cdrs": true,						
		"session_cost_retries": 5,				
		"chargers_conns":["*localhost"],				
		"rals_conns": ["*internal"],
		"attributes_conns": [],					
		"thresholds_conns": [],					
		"stats_conns": [],						
		"online_cdr_exports":[],
		"scheduler_conns": [],				
	},
}`
	eMap := map[string]any{
		"enabled":              false,
		"extra_fields":         []string{},
		"extra_fields_aliases": map[string]string{},
		"store_cdrs":           true,
		"session_cost_retries": 5,
		"chargers_conns":       []string{"*localhost"},
		"rals_conns":           []string{"*internal"},
		"attributes_conns":     []string{},
		"thresholds_conns":     []string{},
		"stats_conns":          []string{},
		"online_cdr_exports":   []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		t.Error(err)
	} else if rcv := cdrscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"cdrs": {
			"enabled": true,						
			"extra_fields": ["PayPalAccount", "LCRProfile", "ResourceID"],
			"store_cdrs": true,						
			"session_cost_retries": 9,				
			"chargers_conns":["*internal"],				
			"rals_conns": ["*internal"],
			"attributes_conns": ["*internal"],					
			"thresholds_conns": ["*internal"],					
			"stats_conns": ["*internal"],						
			"online_cdr_exports":["http_localhost", "amqp_localhost", "http_test_file", "amqp_test_file","aws_test_file","sqs_test_file","kafka_localhost","s3_test_file"],
			"scheduler_conns": ["*internal"],				
		},
	}`
	eMap = map[string]any{
		"enabled":              true,
		"extra_fields":         []string{"PayPalAccount", "LCRProfile", "ResourceID"},
		"extra_fields_aliases": map[string]string{},
		"store_cdrs":           true,
		"session_cost_retries": 9,
		"chargers_conns":       []string{"*internal"},
		"rals_conns":           []string{"*internal"},
		"attributes_conns":     []string{"*internal"},
		"thresholds_conns":     []string{"*internal"},
		"stats_conns":          []string{"*internal"},
		"online_cdr_exports":   []string{"http_localhost", "amqp_localhost", "http_test_file", "amqp_test_file", "aws_test_file", "sqs_test_file", "kafka_localhost", "s3_test_file"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg(); err != nil {
		t.Error(err)
	} else if err = cdrscfg.loadFromJsonCfg(jsnCdrsCfg); err != nil {
		t.Error(err)
	} else if rcv := cdrscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestCdrsCfgloadFromJsonCfg2(t *testing.T) {
	c := CdrsCfg{}
	slc := []string{"val1", "val2"}

	js := CdrsJsonCfg{
		Chargers_conns:     &slc,
		Rals_conns:         &slc,
		Attributes_conns:   &slc,
		Thresholds_conns:   &slc,
		Stats_conns:        &slc,
		Online_cdr_exports: &slc,
	}

	err := c.loadFromJsonCfg(&js)

	if err != nil {
		t.Error(err)
	}

	exp := CdrsCfg{
		ChargerSConns:    slc,
		RaterConns:       slc,
		AttributeSConns:  slc,
		ThresholdSConns:  slc,
		StatSConns:       slc,
		OnlineCDRExports: slc,
	}

	if !reflect.DeepEqual(c, exp) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", c, exp)
	}

	jse := CdrsJsonCfg{
		Extra_fields: &[]string{""},
	}

	err = c.loadFromJsonCfg(&jse)

	if err != nil {
		if err.Error() != "Empty RSRField in rule: " {
			t.Error(err)
		}
	}
}

func TestCdrsCfgAsMapInterface2(t *testing.T) {
	slc := []string{"val1", "val2"}
	bl := false
	nm := 1
	ex := []*utils.RSRField{}

	cdrscfg := CdrsCfg{
		Enabled:          bl,
		ExtraFields:      ex,
		StoreCdrs:        bl,
		SMCostRetries:    nm,
		ChargerSConns:    slc,
		RaterConns:       slc,
		AttributeSConns:  slc,
		ThresholdSConns:  slc,
		StatSConns:       slc,
		OnlineCDRExports: slc,
	}

	exp := map[string]any{
		utils.EnabledCfg:          cdrscfg.Enabled,
		utils.ExtraFieldsCfg:      []string{},
		utils.StoreCdrsCfg:        cdrscfg.StoreCdrs,
		utils.SMCostRetriesCfg:    cdrscfg.SMCostRetries,
		utils.ChargerSConnsCfg:    slc,
		utils.RALsConnsCfg:        slc,
		utils.AttributeSConnsCfg:  slc,
		utils.ThresholdSConnsCfg:  slc,
		utils.StatSConnsCfg:       slc,
		utils.OnlineCDRExportsCfg: slc,
		utils.ExtraFieldAliases:   map[string]string{},
	}

	rcv := cdrscfg.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", rcv, exp)
	}
}
