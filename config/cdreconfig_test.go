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
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
		},
		&CdreCdrField{
			Name:            "extra1",
			Type:            utils.CDRFIELD,
			Value:           "extra1",
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: "extra1"},
		},
	}
	if cdreFlds, err := NewCdreCdrFieldsFromIds(utils.CGRID, "extra1"); err != nil {
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
func TestCdreCfgSetDefaultFixedWidthProperties(t *testing.T) {
	cdreCdrFld := &CdreCdrField{
		valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
	}
	eCdreCdrFld := &CdreCdrField{
		Width:           10,
		Strip:           "xright",
		Mandatory:       true,
		valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
	}
	if err := cdreCdrFld.setDefaultFixedWidthProperties(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cdreCdrFld) {
		t.Errorf("Expecting: %v, received: %v", eCdreCdrFld, cdreCdrFld)
	}
}

func TestCdreCfgNewDefaultCdreConfig(t *testing.T) {
	eCdreCfg := new(CdreConfig)
	eCdreCfg.CdrFormat = utils.CSV
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
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CGRID},
		},
		&CdreCdrField{
			Name:            utils.MEDI_RUNID,
			Type:            utils.CDRFIELD,
			Value:           utils.MEDI_RUNID,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.MEDI_RUNID},
		},
		&CdreCdrField{
			Name:            utils.TOR,
			Type:            utils.CDRFIELD,
			Value:           utils.TOR,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.TOR},
		},
		&CdreCdrField{
			Name:            utils.ACCID,
			Type:            utils.CDRFIELD,
			Value:           utils.ACCID,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ACCID},
		},
		&CdreCdrField{
			Name:            utils.REQTYPE,
			Type:            utils.CDRFIELD,
			Value:           utils.REQTYPE,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.REQTYPE},
		},
		&CdreCdrField{
			Name:            utils.DIRECTION,
			Type:            utils.CDRFIELD,
			Value:           utils.DIRECTION,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.DIRECTION},
		},
		&CdreCdrField{
			Name:            utils.TENANT,
			Type:            utils.CDRFIELD,
			Value:           utils.TENANT,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.TENANT},
		},
		&CdreCdrField{
			Name:            utils.CATEGORY,
			Type:            utils.CDRFIELD,
			Value:           utils.CATEGORY,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.CATEGORY},
		},
		&CdreCdrField{
			Name:            utils.ACCOUNT,
			Type:            utils.CDRFIELD,
			Value:           utils.ACCOUNT,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ACCOUNT},
		},
		&CdreCdrField{
			Name:            utils.SUBJECT,
			Type:            utils.CDRFIELD,
			Value:           utils.SUBJECT,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.SUBJECT},
		},
		&CdreCdrField{
			Name:            utils.DESTINATION,
			Type:            utils.CDRFIELD,
			Value:           utils.DESTINATION,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.DESTINATION},
		},
		&CdreCdrField{
			Name:            utils.SETUP_TIME,
			Type:            utils.CDRFIELD,
			Value:           utils.SETUP_TIME,
			Width:           10,
			Strip:           "xright",
			Layout:          "2006-01-02T15:04:05Z07:00",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.SETUP_TIME},
		},
		&CdreCdrField{
			Name:            utils.ANSWER_TIME,
			Type:            utils.CDRFIELD,
			Value:           utils.ANSWER_TIME,
			Width:           10,
			Strip:           "xright",
			Layout:          "2006-01-02T15:04:05Z07:00",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.ANSWER_TIME},
		},
		&CdreCdrField{
			Name:            utils.USAGE,
			Type:            utils.CDRFIELD,
			Value:           utils.USAGE,
			Width:           10,
			Strip:           "xright",
			Mandatory:       true,
			valueAsRsrField: &utils.RSRField{Id: utils.USAGE},
		},
		&CdreCdrField{
			Name:            utils.COST,
			Type:            utils.CDRFIELD,
			Value:           utils.COST,
			Width:           10,
			Strip:           "xright",
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
