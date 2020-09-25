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
	tLocal, err := time.LoadLocation("Local")
	if err != nil {
		t.Error(err)
	}
	expected = SureTaxCfg{
		Url:                  utils.EmptyString,
		ClientNumber:         utils.EmptyString,
		ValidationKey:        utils.EmptyString,
		BusinessUnit:         utils.EmptyString,
		Timezone:             tLocal,
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
	eMap := map[string]interface{}{
		utils.UrlCfg:                  utils.EmptyString,
		utils.ClientNumberCfg:         utils.EmptyString,
		utils.ValidationKeyCfg:        utils.EmptyString,
		utils.BusinessUnitCfg:         utils.EmptyString,
		utils.TimezoneCfg:             "Local",
		utils.IncludeLocalCostCfg:     false,
		utils.ReturnFileCodeCfg:       "0",
		utils.ResponseGroupCfg:        "03",
		utils.ResponseTypeCfg:         "D4",
		utils.RegulatoryCodeCfg:       "03",
		utils.ClientTrackingCfg:       "~*req.CGRID",
		utils.CustomerNumberCfg:       "~*req.Subject",
		utils.OrigNumberCfg:           "~*req.Subject",
		utils.TermNumberCfg:           "~*req.Destination",
		utils.BillToNumberCfg:         utils.EmptyString,
		utils.ZipcodeCfg:              utils.EmptyString,
		utils.Plus4Cfg:                utils.EmptyString,
		utils.P2PZipcodeCfg:           utils.EmptyString,
		utils.P2PPlus4Cfg:             utils.EmptyString,
		utils.UnitsCfg:                "1",
		utils.UnitTypeCfg:             "00",
		utils.TaxIncludedCfg:          "0",
		utils.TaxSitusRuleCfg:         "04",
		utils.TransTypeCodeCfg:        "010101",
		utils.SalesTypeCodeCfg:        "R",
		utils.TaxExemptionCodeListCfg: utils.EmptyString,
	}
	if cgrCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sureTaxCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSureTaxCfgAsMapInterface1(t *testing.T) {
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
		"term_number":             "~*req.CGRID",
		"units":                   "7",
		"unit_type":               "02",
		"tax_included":            "1",
		"tax_situs_rule":          "03",
		"trans_type_code":         "01010101",
		"sales_type_code":         "B",
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
		utils.BillToNumberCfg:         utils.EmptyString,
		utils.ZipcodeCfg:              utils.EmptyString,
		utils.Plus4Cfg:                utils.EmptyString,
		utils.P2PZipcodeCfg:           utils.EmptyString,
		utils.P2PPlus4Cfg:             utils.EmptyString,
		utils.UnitsCfg:                "7",
		utils.UnitTypeCfg:             "02",
		utils.TaxIncludedCfg:          "1",
		utils.TaxSitusRuleCfg:         "03",
		utils.TransTypeCodeCfg:        "01010101",
		utils.SalesTypeCodeCfg:        "B",
		utils.TaxExemptionCodeListCfg: utils.EmptyString,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.sureTaxCfg.AsMapInterface(cgrCfg.generalCfg.RSRSep); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
