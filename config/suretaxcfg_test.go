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

func TestSureTaxCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Url:                     utils.StringPointer("randomURL"),
		Client_number:           utils.StringPointer("randomClient"),
		Validation_key:          utils.StringPointer("randomKey"),
		Business_unit:           utils.StringPointer("randomUnit"),
		Timezone:                utils.StringPointer("Local"),
		Include_local_cost:      utils.BoolPointer(true),
		Return_file_code:        utils.StringPointer("1"),
		Response_group:          utils.StringPointer("06"),
		Response_type:           utils.StringPointer("A3"),
		Regulatory_code:         utils.StringPointer("06"),
		Client_tracking:         utils.StringPointer("~*req.Destination1"),
		Customer_number:         utils.StringPointer("~*req.Destination1"),
		Orig_number:             utils.StringPointer("~*req.Destination1"),
		Term_number:             utils.StringPointer("~*req.CGRID"),
		Bill_to_number:          utils.StringPointer(utils.EmptyString),
		Zipcode:                 utils.StringPointer(utils.EmptyString),
		Plus4:                   utils.StringPointer(utils.EmptyString),
		P2PZipcode:              utils.StringPointer(utils.EmptyString),
		P2PPlus4:                utils.StringPointer(utils.EmptyString),
		Units:                   utils.StringPointer("1"),
		Unit_type:               utils.StringPointer("00"),
		Tax_included:            utils.StringPointer("0"),
		Tax_situs_rule:          utils.StringPointer("04"),
		Trans_type_code:         utils.StringPointer("010101"),
		Sales_type_code:         utils.StringPointer("R"),
		Tax_exemption_code_list: utils.StringPointer(utils.EmptyString),
	}
	tLocal, err := time.LoadLocation("Local")
	if err != nil {
		t.Error(err)
	}
	expected := &SureTaxCfg{
		Url:                  "randomURL",
		ClientNumber:         "randomClient",
		ValidationKey:        "randomKey",
		BusinessUnit:         "randomUnit",
		Timezone:             tLocal,
		IncludeLocalCost:     true,
		ReturnFileCode:       "1",
		ResponseGroup:        "06",
		ResponseType:         "A3",
		RegulatoryCode:       "06",
		ClientTracking:       NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Destination1", utils.INFIELD_SEP),
		TermNumber:           NewRSRParsersMustCompile("~*req.CGRID", utils.INFIELD_SEP),
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
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.sureTaxCfg) {
		t.Errorf("Expecetd %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.sureTaxCfg))
	}
}

func TestSureTaxCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Timezone: utils.StringPointer("/"),
	}
	expected := "time: invalid location name"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase3(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Client_tracking: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase4(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Customer_number: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase5(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Orig_number: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase6(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Term_number: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase7(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Bill_to_number: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase8(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Zipcode: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase9(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Plus4: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase10(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		P2PZipcode: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase11(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		P2PPlus4: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase12(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Units: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase13(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Unit_type: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase14(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Tax_included: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase15(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Tax_situs_rule: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase16(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Trans_type_code: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase17(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Sales_type_code: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgloadFromJsonCfgCase18(t *testing.T) {
	cfgJSON := &SureTaxJsonCfg{
		Tax_exemption_code_list: utils.StringPointer("a{*"),
	}
	expected := "invalid converter terminator in rule: <a{*>"
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.sureTaxCfg.loadFromJsonCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestSureTaxCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"suretax": {
		"timezone":                "UTC",
		"include_local_cost":      true,
		"return_file_code":        "0",
		"response_group":          "04",
		"response_type":           "A4",
		"regulatory_code":         "04",
		"client_tracking":         "~*req.Destination",
		"customer_number":         "~*req.Destination",
		"orig_number":             "~*req.Destination",
        "bill_to_number": "randomNumber",					
	    "zipcode": "randomCode",							
	    "plus4": "randomValue",							
	    "p2pzipcode": "randomCode",						
	    "p2pplus4": "randomValue",
		"term_number":             "~*req.CGRID",
		"units":                   "7",
		"unit_type":               "02",
		"tax_included":            "1",
		"tax_situs_rule":          "03",
		"trans_type_code":         "01010101",
		"sales_type_code":         "B",
        "tax_exemption_code_list": "randomCode"
    },
}`
	eMap := map[string]interface{}{
		utils.UrlCfg:                  utils.EmptyString,
		utils.ClientNumberCfg:         utils.EmptyString,
		utils.ValidationKeyCfg:        utils.EmptyString,
		utils.BusinessUnitCfg:         utils.EmptyString,
		utils.TimezoneCfg:             "UTC",
		utils.IncludeLocalCostCfg:     true,
		utils.ReturnFileCodeCfg:       "0",
		utils.ResponseGroupCfg:        "04",
		utils.ResponseTypeCfg:         "A4",
		utils.RegulatoryCodeCfg:       "04",
		utils.ClientTrackingCfg:       "~*req.Destination",
		utils.CustomerNumberCfg:       "~*req.Destination",
		utils.OrigNumberCfg:           "~*req.Destination",
		utils.TermNumberCfg:           "~*req.CGRID",
		utils.BillToNumberCfg:         "randomNumber",
		utils.ZipcodeCfg:              "randomCode",
		utils.Plus4Cfg:                "randomValue",
		utils.P2PZipcodeCfg:           "randomCode",
		utils.P2PPlus4Cfg:             "randomValue",
		utils.UnitsCfg:                "7",
		utils.UnitTypeCfg:             "02",
		utils.TaxIncludedCfg:          "1",
		utils.TaxSitusRuleCfg:         "03",
		utils.TransTypeCodeCfg:        "01010101",
		utils.SalesTypeCodeCfg:        "B",
		utils.TaxExemptionCodeListCfg: "randomCode",
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sureTaxCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
