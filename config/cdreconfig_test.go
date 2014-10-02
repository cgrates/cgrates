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
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

func TestNewCfgCdrFieldsFromIds(t *testing.T) {
	expectedFlds := []*CfgCdrField{
		&CfgCdrField{
			Tag:        utils.CGRID,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.CGRID,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.CGRID}},
			Width:     40,
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        "extra1",
			Type:       utils.CDRFIELD,
			CdrFieldId: "extra1",
			Value: []*utils.RSRField{
				&utils.RSRField{Id: "extra1"}},
			Width:     30,
			Strip:     "xright",
			Padding:   "left",
			Mandatory: false,
		},
	}
	if cdreFlds, err := NewCfgCdrFieldsFromIds(true, utils.CGRID, "extra1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFlds, cdreFlds) {
		t.Errorf("Expected: %v, received: %v", expectedFlds, cdreFlds)
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
	eCdreCfg.ContentFields = []*CfgCdrField{
		&CfgCdrField{
			Tag:        utils.CGRID,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.CGRID,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.CGRID}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.MEDI_RUNID,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.MEDI_RUNID,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.MEDI_RUNID}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.TOR,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.TOR,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.TOR}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.ACCID,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.ACCID,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.ACCID}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.REQTYPE,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.REQTYPE,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.REQTYPE}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.DIRECTION,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.DIRECTION,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.DIRECTION}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.TENANT,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.TENANT,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.TENANT}},
			Mandatory: true},
		&CfgCdrField{
			Tag:        utils.CATEGORY,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.CATEGORY,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.CATEGORY}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.ACCOUNT,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.ACCOUNT,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.ACCOUNT}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.SUBJECT,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.SUBJECT,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.SUBJECT}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.DESTINATION,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.DESTINATION,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.DESTINATION}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.SETUP_TIME,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.SETUP_TIME,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.SETUP_TIME}},
			Layout:    "2006-01-02T15:04:05Z07:00",
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.ANSWER_TIME,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.ANSWER_TIME,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.ANSWER_TIME}},
			Layout:    "2006-01-02T15:04:05Z07:00",
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.USAGE,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.USAGE,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.USAGE}},
			Mandatory: true,
		},
		&CfgCdrField{
			Tag:        utils.COST,
			Type:       utils.CDRFIELD,
			CdrFieldId: utils.COST,
			Value: []*utils.RSRField{
				&utils.RSRField{Id: utils.COST}},
			Mandatory: true,
		},
	}
	if cdreCfg := NewDefaultCdreConfig(); !reflect.DeepEqual(eCdreCfg, cdreCfg) {
		for _, fld := range cdreCfg.ContentFields {
			fmt.Printf("Have field: %+v\n", fld)
		}
		t.Errorf("Expecting: %v, received: %v", eCdreCfg, cdreCfg)
	}
}
