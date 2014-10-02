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
	//"fmt"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

func TestNewCfgCdrFieldWithDefaults(t *testing.T) {
	eCdreCdrFld := &CfgCdrField{
		Tag:        utils.CGRID,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.CGRID,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.CGRID}},
		Width:     40,
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.CGRID}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.ORDERID,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.ORDERID,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.ORDERID}},
		Width:     11,
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.ORDERID}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.TOR,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.TOR,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.TOR}},
		Width:     6,
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.TOR}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.ACCID,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.ACCID,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.ACCID}},
		Width:     36,
		Strip:     "left",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.ACCID}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.CDRHOST,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.CDRHOST,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.CDRHOST}},
		Width:     15,
		Strip:     "left",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.CDRHOST}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.CDRSOURCE,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.CDRSOURCE,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.CDRSOURCE}},
		Width:     15,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.CDRSOURCE}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.REQTYPE,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.REQTYPE,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.REQTYPE}},
		Width:     13,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.REQTYPE}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.DIRECTION,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.DIRECTION,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.DIRECTION}},
		Width:     4,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.DIRECTION}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.TENANT,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.TENANT,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.TENANT}},
		Width:     24,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.TENANT}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.CATEGORY,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.CATEGORY,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.CATEGORY}},
		Width:     10,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.CATEGORY}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.ACCOUNT,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.ACCOUNT,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.ACCOUNT}},
		Width:     24,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.ACCOUNT}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.SUBJECT,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.SUBJECT,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.SUBJECT}},
		Width:     24,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.SUBJECT}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.DESTINATION,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.DESTINATION,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.DESTINATION}},
		Width:     24,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.DESTINATION}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.SETUP_TIME,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.SETUP_TIME,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.SETUP_TIME}},
		Width:     30,
		Strip:     "xright",
		Padding:   "left",
		Layout:    "2006-01-02T15:04:05Z07:00",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.SETUP_TIME}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.ANSWER_TIME,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.ANSWER_TIME,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.ANSWER_TIME}},
		Width:     30,
		Strip:     "xright",
		Padding:   "left",
		Layout:    "2006-01-02T15:04:05Z07:00",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.ANSWER_TIME}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.USAGE,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.USAGE,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.USAGE}},
		Width:     30,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.USAGE}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.MEDI_RUNID,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.MEDI_RUNID,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.MEDI_RUNID}},
		Width:     20,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.MEDI_RUNID}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        utils.COST,
		Type:       utils.CDRFIELD,
		CdrFieldId: utils.COST,
		Value: []*utils.RSRField{
			&utils.RSRField{Id: utils.COST}},
		Width:     24,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: true,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: utils.COST}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
	eCdreCdrFld = &CfgCdrField{
		Tag:        "extra_1",
		Type:       utils.CDRFIELD,
		CdrFieldId: "extra_1",
		Value: []*utils.RSRField{
			&utils.RSRField{Id: "extra_1"}},
		Width:     30,
		Strip:     "xright",
		Padding:   "left",
		Mandatory: false,
	}
	if cfgCdrField, err := NewCfgCdrFieldWithDefaults(true, []*utils.RSRField{&utils.RSRField{Id: "extra_1"}}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCdreCdrFld, cfgCdrField) {
		t.Errorf("Expecting: %+v, received: %+v", eCdreCdrFld, cfgCdrField)
	}
}
