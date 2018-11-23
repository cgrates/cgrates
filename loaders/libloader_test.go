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

package loaders

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDataUpdateFromCSVOneFile(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "TenantID",
			FieldId:   "Tenant",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "ProfileID",
			FieldId:   "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "Contexts",
			FieldId: "Contexts",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "FilterIDs",
			FieldId: "FilterIDs",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "ActivationInterval",
			FieldId: "ActivationInterval",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "FieldName",
			FieldId: "FieldName",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Initial",
			FieldId: "Initial",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Substitute",
			FieldId: "Substitute",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Append",
			FieldId: "Append",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Weight",
			FieldId: "Weight",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
	}

	rows := [][]string{
		[]string{"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:Account:1007", "2014-01-14T00:00:00Z", "Account", "*any", "1001", "false", "10"},
		[]string{"cgrates.org", "ATTR_1", "", "", "", "Subject", "*any", "1001", "true", ""},
	}
	lData := make(LoaderData)
	if err := lData.UpdateFromCSV("Attributes.csv", rows[0], attrSFlds,
		config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData := LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "*sessions;*cdrs",
		"FilterIDs":          "*string:Account:1007",
		"ActivationInterval": "2014-01-14T00:00:00Z",
		"FieldName":          "Account",
		"Initial":            "*any",
		"Substitute":         "1001",
		"Append":             "false",
		"Weight":             "10",
	}
	if !reflect.DeepEqual(eLData, lData) {
		t.Errorf("expecting: %+v, received: %+v", eLData, lData)
	}
	lData = make(LoaderData)
	if err := lData.UpdateFromCSV("Attributes.csv", rows[1], attrSFlds,
		config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData = LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "",
		"FilterIDs":          "",
		"ActivationInterval": "",
		"FieldName":          "Subject",
		"Initial":            "*any",
		"Substitute":         "1001",
		"Append":             "true",
		"Weight":             "",
	}
	if !reflect.DeepEqual(eLData, lData) {
		t.Errorf("expecting: %+v, received: %+v", eLData, lData)
	}
}

func TestDataUpdateFromCSVMultiFiles(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "TenantID",
			FieldId:   "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "ProfileID",
			FieldId:   "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~File2.csv:1", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "Contexts",
			FieldId: "Contexts",
			Type:    utils.MetaString,
			Value:   config.NewRSRParsersMustCompile("*any", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "FieldName",
			FieldId: "FieldName",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~File1.csv:5", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Initial",
			FieldId: "Initial",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~File1.csv:6", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Substitute",
			FieldId: "Substitute",
			Type:    utils.META_COMPOSED,
			Value:   config.NewRSRParsersMustCompile("~File1.csv:7", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Append",
			FieldId: "Append",
			Type:    utils.MetaString,
			Value:   config.NewRSRParsersMustCompile("true", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Weight",
			FieldId: "Weight",
			Type:    utils.MetaString,
			Value:   config.NewRSRParsersMustCompile("10", true, utils.INFIELD_SEP)},
	}

	loadRun1 := map[string][]string{
		"File1.csv": []string{"ignored", "ignored", "ignored", "ignored", "ignored", "Subject", "*any", "1001", "ignored", "ignored"},
		"File2.csv": []string{"ignored", "ATTR_1"},
	}
	lData := make(LoaderData)
	for fName, record := range loadRun1 {
		if err := lData.UpdateFromCSV(fName, record, attrSFlds,
			config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP), nil); err != nil {
			t.Error(err)
		}
	}

	eLData := LoaderData{"Tenant": "cgrates.org",
		"ID":         "ATTR_1",
		"Contexts":   "*any",
		"FieldName":  "Subject",
		"Initial":    "*any",
		"Substitute": "1001",
		"Append":     "true",
		"Weight":     "10",
	}
	if !reflect.DeepEqual(eLData, lData) {
		t.Errorf("expecting: %+v, received: %+v", eLData, lData)
	}
}
