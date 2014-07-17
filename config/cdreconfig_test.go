/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

func TestCdreCfgNewCdreCdrFieldsFromIds(t *testing.T) {
	expectedFlds := []*CdreCdrField{
		&CdreCdrField{
			Name:            utils.CGRID,
			Type:            utils.CDRFIELD,
			Value:           utils.CGRID,
			Width:           40,
			Strip:           "",
			Padding:         "",
			Layout:          "",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
		},
		&CdreCdrField{
			Name:            "extra1",
			Type:            utils.CDRFIELD,
			Value:           "extra1",
			Width:           30,
			Strip:           "xright",
			Padding:         "left",
			Layout:          "",
			Mandatory:       false,
			valueAsRsrField: &utils.RSRField{Id: "extra1"},
		},
	}
	if cdreFlds, err := NewCdreCdrFieldsFromIds(true, utils.CGRID, "extra1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFlds, cdreFlds) {
		t.Errorf("Expected: %v, received: %v", expectedFlds, cdreFlds)
	}
}

func TestCdreCfgValueAsRSRField(t *testing.T) {
	cdreCdrFld := &CdreCdrField{
		Name:            utils.CGRID,
		Type:            utils.CDRFIELD,
		Value:           utils.CGRID,
		Width:           10,
		Strip:           "xright",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
	}
	if rsrVal := cdreCdrFld.ValueAsRSRField(); rsrVal != cdreCdrFld.valueAsRsrField {
		t.Error("Unexpected value received: ", rsrVal)
	}
}

func TestCdreCfgNewDefaultCdreConfig(t *testing.T) {
	eCdreCfg := new(CdreConfig)
	eCdreCfg.CdrFormat = utils.CSV
	eCdreCfg.FieldSeparator = utils.CSV_SEP
	eCdreCfg.DataUsageMultiplyFactor = 0.0
	eCdreCfg.CostMultiplyFactor = 0.0
	eCdreCfg.CostRoundingDecimals = -1
	eCdreCfg.CostShiftDigits = 0
	eCdreCfg.MaskDestId = ""
	eCdreCfg.MaskLength = 0
	eCdreCfg.ExportDir = "/var/log/cgrates/cdre"
	eCdreCfg.ContentFields = []*CdreCdrField{
		&CdreCdrField{
			Name:            utils.CGRID,
			Type:            utils.CDRFIELD,
			Value:           utils.CGRID,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
		},
		&CdreCdrField{
			Name:            utils.MEDI_RUNID,
			Type:            utils.CDRFIELD,
			Value:           utils.MEDI_RUNID,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.MEDI_RUNID},
		},
		&CdreCdrField{
			Name:            utils.TOR,
			Type:            utils.CDRFIELD,
			Value:           utils.TOR,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.TOR},
		},
		&CdreCdrField{
			Name:            utils.ACCID,
			Type:            utils.CDRFIELD,
			Value:           utils.ACCID,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ACCID},
		},
		&CdreCdrField{
			Name:            utils.REQTYPE,
			Type:            utils.CDRFIELD,
			Value:           utils.REQTYPE,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.REQTYPE},
		},
		&CdreCdrField{
			Name:            utils.DIRECTION,
			Type:            utils.CDRFIELD,
			Value:           utils.DIRECTION,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.DIRECTION},
		},
		&CdreCdrField{
			Name:            utils.TENANT,
			Type:            utils.CDRFIELD,
			Value:           utils.TENANT,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.TENANT},
		},
		&CdreCdrField{
			Name:            utils.CATEGORY,
			Type:            utils.CDRFIELD,
			Value:           utils.CATEGORY,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CATEGORY},
		},
		&CdreCdrField{
			Name:            utils.ACCOUNT,
			Type:            utils.CDRFIELD,
			Value:           utils.ACCOUNT,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ACCOUNT},
		},
		&CdreCdrField{
			Name:            utils.SUBJECT,
			Type:            utils.CDRFIELD,
			Value:           utils.SUBJECT,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.SUBJECT},
		},
		&CdreCdrField{
			Name:            utils.DESTINATION,
			Type:            utils.CDRFIELD,
			Value:           utils.DESTINATION,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.DESTINATION},
		},
		&CdreCdrField{
			Name:            utils.SETUP_TIME,
			Type:            utils.CDRFIELD,
			Value:           utils.SETUP_TIME,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.SETUP_TIME},
		},
		&CdreCdrField{
			Name:            utils.ANSWER_TIME,
			Type:            utils.CDRFIELD,
			Value:           utils.ANSWER_TIME,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ANSWER_TIME},
		},
		&CdreCdrField{
			Name:            utils.USAGE,
			Type:            utils.CDRFIELD,
			Value:           utils.USAGE,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.USAGE},
		},
		&CdreCdrField{
			Name:            utils.COST,
			Type:            utils.CDRFIELD,
			Value:           utils.COST,
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.COST},
		},
	}
	if cdreCfg, err := NewDefaultCdreConfig(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCfg, cdreCfg) {
		t.Errorf("Expecting: %v, received: %v", eCdreCfg, cdreCfg)
	}
}

func TestCdreCfgSetDefaultFieldProperties(t *testing.T) {
	cdreCdrFld := &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
	}
	eCdreCdrFld := &CdreCdrField{
		Width:           40,
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.ORDERID},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           11,
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.ORDERID},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.TOR},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           6,
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.TOR},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.ACCID},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           36,
		Strip:           "left",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.ACCID},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.CDRHOST},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           15,
		Strip:           "left",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CDRHOST},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.CDRSOURCE},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           15,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CDRSOURCE},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.REQTYPE},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           13,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.REQTYPE},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.DIRECTION},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           4,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.DIRECTION},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.TENANT},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           24,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.TENANT},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.CATEGORY},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           10,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CATEGORY},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.ACCOUNT},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           24,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.ACCOUNT},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.SUBJECT},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           24,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.SUBJECT},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.DESTINATION},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           24,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.DESTINATION},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.SETUP_TIME},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           30,
		Strip:           "xright",
		Padding:         "left",
		Layout:          "2006-01-02T15:04:05Z07:00",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.SETUP_TIME},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.ANSWER_TIME},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           30,
		Strip:           "xright",
		Padding:         "left",
		Layout:          "2006-01-02T15:04:05Z07:00",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.ANSWER_TIME},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.USAGE},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           30,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.USAGE},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.MEDI_RUNID},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           20,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.MEDI_RUNID},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.COST},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           24,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.COST},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
	cdreCdrFld = &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: "extra_1"},
	}
	eCdreCdrFld = &CdreCdrField{
		Width:           30,
		Strip:           "xright",
		Padding:         "left",
		Mandatory:       false,
		valueAsRsrField: &utils.RSRField{Id: "extra_1"},
	}
	if err := cdreCdrFld.setDefaultFieldProperties(true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
}
