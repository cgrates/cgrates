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
	"errors"
	"fmt"
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
		t.Errorf("Expected: %+v ,received: %+v", expected, sureTaxCfg)
	}
	if err := sureTaxCfg.loadFromJsonCfg(new(SureTaxJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sureTaxCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, sureTaxCfg)
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
		ClientTracking:       NewRSRParsersMustCompile("~*req.CGRID", true, utils.INFIELD_SEP),
		CustomerNumber:       NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
		OrigNumber:           NewRSRParsersMustCompile("~*req.Subject", true, utils.INFIELD_SEP),
		TermNumber:           NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP),
		BillToNumber:         NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
		Zipcode:              NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
		Plus4:                NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
		P2PZipcode:           NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
		P2PPlus4:             NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
		Units:                NewRSRParsersMustCompile("1", true, utils.INFIELD_SEP),
		UnitType:             NewRSRParsersMustCompile("00", true, utils.INFIELD_SEP),
		TaxIncluded:          NewRSRParsersMustCompile("0", true, utils.INFIELD_SEP),
		TaxSitusRule:         NewRSRParsersMustCompile("04", true, utils.INFIELD_SEP),
		TransTypeCode:        NewRSRParsersMustCompile("010101", true, utils.INFIELD_SEP),
		SalesTypeCode:        NewRSRParsersMustCompile("R", true, utils.INFIELD_SEP),
		TaxExemptionCodeList: NewRSRParsersMustCompile(utils.EmptyString, true, utils.INFIELD_SEP),
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if sureTax, err := jsnCfg.SureTaxJsonCfg(); err != nil {
		t.Error(err)
	} else if err = sureTaxCfg.loadFromJsonCfg(sureTax); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, sureTaxCfg) {
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(expected), utils.ToJSON(sureTaxCfg))
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
	eMap := map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSureTaxCFGLoadFromJsonCFG(t *testing.T) {

	str := "test`"

	st := SureTaxCfg{}

	stjCT := SureTaxJsonCfg{
		Client_tracking: &str,
	}

	stjCN := SureTaxJsonCfg{
		Customer_number: &str,
	}

	stjON := SureTaxJsonCfg{
		Orig_number: &str,
	}

	stjTN := SureTaxJsonCfg{
		Term_number: &str,
	}

	stjBTN := SureTaxJsonCfg{
		Bill_to_number: &str,
	}

	stjZ := SureTaxJsonCfg{
		Zipcode: &str,
	}

	stjPP4 := SureTaxJsonCfg{
		P2PPlus4: &str,
	}

	stjU := SureTaxJsonCfg{
		Units: &str,
	}

	stjUT := SureTaxJsonCfg{
		Unit_type: &str,
	}

	stjTI := SureTaxJsonCfg{
		Tax_included: &str,
	}

	stjTSR := SureTaxJsonCfg{
		Tax_situs_rule: &str,
	}

	stjTTC := SureTaxJsonCfg{
		Trans_type_code: &str,
	}

	stjSTC := SureTaxJsonCfg{
		Sales_type_code: &str,
	}

	stjTECL := SureTaxJsonCfg{
		Tax_exemption_code_list: &str,
	}

	stjPPZ := SureTaxJsonCfg{
		P2PZipcode: &str,
	}

	stjP4 := SureTaxJsonCfg{
		Plus4: &str,
	}

	tests := []struct {
		name string
		arg  *SureTaxJsonCfg
	}{
		{
			arg: &stjCT,
		},
		{
			arg: &stjCN,
		},
		{
			arg: &stjON,
		},
		{
			arg: &stjTN,
		},
		{
			arg: &stjBTN,
		},
		{
			arg: &stjZ,
		},
		{
			arg: &stjPP4,
		},
		{
			arg: &stjU,
		},
		{
			arg: &stjUT,
		},
		{
			arg: &stjTI,
		},
		{
			arg: &stjTSR,
		},
		{
			arg: &stjTTC,
		},
		{
			arg: &stjSTC,
		},
		{
			arg: &stjTECL,
		},
		{
			arg: &stjPPZ,
		},
		{
			arg: &stjP4,
		},
	}

	for _, tt := range tests {
		t.Run("check errors", func(t *testing.T) {
			err := st.loadFromJsonCfg(tt.arg)
			exp := fmt.Errorf("Unclosed unspilit syntax")

			if err.Error() != exp.Error() {
				t.Fatalf("recived %s, expected %s", err, exp)
			}
		})
	}

	t.Run("check timezone error", func(t *testing.T) {
		str := "\\test"

		st := SureTaxCfg{}

		stjT := SureTaxJsonCfg{
			Timezone: &str,
		}

		err := st.loadFromJsonCfg(&stjT)
		exp := errors.New("time: invalid location name")

		if err.Error() != exp.Error() {
			t.Fatalf("recived %s, expected %s", err, exp)
		}
	})
}

func TestSureTaxCFGAsMapInterface(t *testing.T) {
	str := "test"
	rsr, _ := NewRSRParsers(str, true, "")

	st := SureTaxCfg{
		Url:                  str,
		ClientNumber:         str,
		ValidationKey:        str,
		BusinessUnit:         str,
		Timezone:             &time.Location{},
		IncludeLocalCost:     false,
		ReturnFileCode:       str,
		ResponseGroup:        str,
		ResponseType:         str,
		RegulatoryCode:       str,
		ClientTracking:       rsr,
		CustomerNumber:       rsr,
		OrigNumber:           rsr,
		TermNumber:           rsr,
		BillToNumber:         rsr,
		Zipcode:              rsr,
		Plus4:                rsr,
		P2PZipcode:           rsr,
		P2PPlus4:             rsr,
		Units:                rsr,
		UnitType:             rsr,
		TaxIncluded:          rsr,
		TaxSitusRule:         rsr,
		TransTypeCode:        rsr,
		SalesTypeCode:        rsr,
		TaxExemptionCodeList: rsr,
	}

	mp := map[string]any{
		utils.UrlCfg:                  st.Url,
		utils.ClientNumberCfg:         st.ClientNumber,
		utils.ValidationKeyCfg:        st.ValidationKey,
		utils.BusinessUnitCfg:         st.BusinessUnit,
		utils.TimezoneCfg:             st.Timezone.String(),
		utils.IncludeLocalCostCfg:     st.IncludeLocalCost,
		utils.ReturnFileCodeCfg:       st.ReturnFileCode,
		utils.ResponseGroupCfg:        st.ResponseGroup,
		utils.ResponseTypeCfg:         st.ResponseType,
		utils.RegulatoryCodeCfg:       st.RegulatoryCode,
		utils.ClientTrackingCfg:       str,
		utils.CustomerNumberCfg:       str,
		utils.OrigNumberCfg:           str,
		utils.TermNumberCfg:           str,
		utils.BillToNumberCfg:         str,
		utils.ZipcodeCfg:              str,
		utils.Plus4Cfg:                str,
		utils.P2PZipcodeCfg:           str,
		utils.P2PPlus4Cfg:             str,
		utils.UnitsCfg:                str,
		utils.UnitTypeCfg:             str,
		utils.TaxIncludedCfg:          str,
		utils.TaxSitusRuleCfg:         str,
		utils.TransTypeCodeCfg:        str,
		utils.SalesTypeCodeCfg:        str,
		utils.TaxExemptionCodeListCfg: str,
	}

	rcv := st.AsMapInterface("")

	if !reflect.DeepEqual(rcv, mp) {
		t.Errorf("recived %v, expected %v", rcv, mp)
	}
}
