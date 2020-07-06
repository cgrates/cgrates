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

func TestSureTaxCfgloadFromJsonCfg(t *testing.T) {
	var sureTaxCfg, expected SureTaxCfg
	if err := sureTaxCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sureTaxCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, sureTaxCfg)
	}
	if err := sureTaxCfg.loadFromJsonCfg(new(SureTaxJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sureTaxCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, sureTaxCfg)
	}
	cfgJSONStr := `{
	"suretax": {
		"url": "",
		"client_number": "",
		"validation_key": "",
		"business_unit": "",
		"timezone": "Local",
		"include_local_cost": false,
		"return_file_code": "0",
		"response_group": "03",
		"response_type": "D4",
		"regulatory_code": "03",
		"client_tracking": "~*req.CGRID",
		"customer_number": "~*req.Subject",
		"orig_number":  "~*req.Subject",
		"term_number": "~*req.Destination",
		"bill_to_number": "",
		"zipcode": "",
		"plus4": "",
		"p2pzipcode": "",
		"p2pplus4": "",
		"units": "1",
		"unit_type": "00",
		"tax_included": "0",
		"tax_situs_rule": "04",
		"trans_type_code": "010101",
		"sales_type_code": "R",
		"tax_exemption_code_list": "",
	},
}`
	expected = SureTaxCfg{
		Url:                  utils.EmptyString,
		ClientNumber:         utils.EmptyString,
		ValidationKey:        utils.EmptyString,
		BusinessUnit:         utils.EmptyString,
		Timezone:             &time.Location{},
		IncludeLocalCost:     false,
		ReturnFileCode:       "0",
		ResponseGroup:        "03",
		ResponseType:         "D4",
		RegulatoryCode:       "03",
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", utils.INFIELD_SEP),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", utils.INFIELD_SEP),
		BillToNumber:         NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
		Zipcode:              NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
		Plus4:                NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
		P2PZipcode:           NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
		P2PPlus4:             NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
		Units:                NewRSRParsersMustCompile("1", utils.INFIELD_SEP),
		UnitType:             NewRSRParsersMustCompile("00", utils.INFIELD_SEP),
		TaxIncluded:          NewRSRParsersMustCompile("0", utils.INFIELD_SEP),
		TaxSitusRule:         NewRSRParsersMustCompile("04", utils.INFIELD_SEP),
		TransTypeCode:        NewRSRParsersMustCompile("010101", utils.INFIELD_SEP),
		SalesTypeCode:        NewRSRParsersMustCompile("R", utils.INFIELD_SEP),
		TaxExemptionCodeList: NewRSRParsersMustCompile(utils.EmptyString, utils.INFIELD_SEP),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if sureTax, err := jsnCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sureTaxCfg.loadFromJsonCfg(sureTax); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sureTaxCfg) {
		t.Errorf("Expected: %+v,\nRecived: %+v", utils.ToJSON(expected), utils.ToJSON(sureTaxCfg))
	}
}

func TestSureTaxCfgAsMapInterface(t *testing.T) {
	var sureTaxCfg SureTaxCfg
	cfgJSONStr := `{
	"suretax": {
		"url": "",
		"client_number": "",
		"validation_key": "",
		"business_unit": "",
		"timezone": "UTC",
		"include_local_cost": false,
		"return_file_code": "0",
		"response_group": "03",
		"response_type": "D4",
		"regulatory_code": "03",
		"client_tracking": "~*req.CGRID",
		"customer_number": "~*req.Subject",
		"orig_number":  "~*req.Subject",
		"term_number": "~*req.Destination",
		"bill_to_number": "",
		"zipcode": "",
		"plus4": "",
		"p2pzipcode": "",
		"p2pplus4": "",
		"units": "1",
		"unit_type": "00",
		"tax_included": "0",
		"tax_situs_rule": "04",
		"trans_type_code": "010101",
		"sales_type_code": "R",
		"tax_exemption_code_list": "",
	},
}`
	eMap := map[string]interface{}{
		"url":                     "",
		"client_number":           "",
		"validation_key":          "",
		"business_unit":           "",
		"timezone":                "UTC",
		"include_local_cost":      false,
		"return_file_code":        "0",
		"response_group":          "03",
		"response_type":           "D4",
		"regulatory_code":         "03",
		"client_tracking":         "~*req.CGRID",
		"customer_number":         "~*req.Subject",
		"orig_number":             "~*req.Subject",
		"term_number":             "~*req.Destination",
		"bill_to_number":          "",
		"zipcode":                 "",
		"plus4":                   "",
		"p2pzipcode":              "",
		"p2pplus4":                "",
		"units":                   "1",
		"unit_type":               "00",
		"tax_included":            "0",
		"tax_situs_rule":          "04",
		"trans_type_code":         "010101",
		"sales_type_code":         "R",
		"tax_exemption_code_list": "",
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if sureTax, err := jsnCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sureTaxCfg.loadFromJsonCfg(sureTax); err != nil {
		t.Error(err)
	} else if rcv := sureTaxCfg.AsMapInterface(utils.EmptyString); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
