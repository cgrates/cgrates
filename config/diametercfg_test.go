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

func TestDiameterAgentCfgloadFromJsonCfg(t *testing.T) {
	var dacfg, expected DiameterAgentCfg
	if err := dacfg.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dacfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dacfg)
	}
	if err := dacfg.loadFromJsonCfg(new(DiameterAgentJsonCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dacfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dacfg)
	}
	cfgJSONStr := `{
"diameter_agent": {
	"enabled": false,											// enables the diameter agent: <true|false>
	"listen": "127.0.0.1:3868",									// address where to listen for diameter requests <x.y.z.y:1234>
	"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	// path towards directory holding additional dictionaries to load
	"sessions_conns": [
		{"address": "*internal"}								// connection towards SessionService
	],
	"origin_host": "CGR-DA",									// diameter Origin-Host AVP used in replies
	"origin_realm": "cgrates.org",								// diameter Origin-Realm AVP used in replies
	"vendor_id": 0,												// diameter Vendor-Id AVP used in replies
	"product_name": "CGRateS",									// diameter Product-Name AVP used in replies
	"templates":{},
	"request_processors": [],
},
}`
	expected = DiameterAgentCfg{
		Listen:           "127.0.0.1:3868",
		DictionariesPath: "/usr/share/cgrates/diameter/dict/",
		SessionSConns:    []*HaPoolConfig{{Address: "*internal"}},
		OriginHost:       "CGR-DA",
		OriginRealm:      "cgrates.org",
		VendorId:         0,
		ProductName:      "CGRateS",
		Templates:        make(map[string][]*FCTemplate),
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DiameterAgentJsonCfg(); err != nil {
		t.Error(err)
	} else if err = dacfg.loadFromJsonCfg(jsnDaCfg, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dacfg) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dacfg))
	}
}

func TestDARequestProcessorloadFromJsonCfg(t *testing.T) {
	var dareq, expected DARequestProcessor
	if err := dareq.loadFromJsonCfg(nil, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dareq)
	}
	if err := dareq.loadFromJsonCfg(new(DARequestProcessorJsnCfg), utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dareq, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, dareq)
	}
	json := &DARequestProcessorJsnCfg{
		Id:                  utils.StringPointer("cgrates"),
		Tenant:              utils.StringPointer("tenant"),
		Filters:             &[]string{"filter1", "filter2"},
		Flags:               &[]string{"flag1", "flag2"},
		Timezone:            utils.StringPointer("Local"),
		Continue_on_success: utils.BoolPointer(true),
	}
	expected = DARequestProcessor{
		ID:                "cgrates",
		Tenant:            NewRSRParsersMustCompile("tenant", true, utils.INFIELD_SEP),
		Filters:           []string{"filter1", "filter2"},
		Flags:             utils.StringMap{"flag1": true, "flag2": true},
		Timezone:          "Local",
		ContinueOnSuccess: true,
	}
	if err = dareq.loadFromJsonCfg(json, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, dareq) {
		t.Errorf("Expected: %+v , recived: %+v", utils.ToJSON(expected), utils.ToJSON(dareq))
	}
}
