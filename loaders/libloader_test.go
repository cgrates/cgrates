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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDataUpdateFromCSVOneFile(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "Contexts",
			Path:  "Contexts",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
		{Tag: "FilterIDs",
			Path:  "FilterIDs",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
		{Tag: "ActivationInterval",
			Path:  "ActivationInterval",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
		{Tag: "Path",
			Path:  "Path",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
		{Tag: "Initial",
			Path:  "Initial",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
		{Tag: "Substitute",
			Path:  "Substitute",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
		{Tag: "Append",
			Path:  "Append",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
		{Tag: "Weight",
			Path:  "Weight",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
	}

	rows := [][]string{
		{"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:Account:1007", "2014-01-14T00:00:00Z", "Account", "*any", "1001", "false", "10"},
		{"cgrates.org", "ATTR_1", "", "", "", "Subject", "*any", "1001", "true", ""},
	}
	lData := make(LoaderData)
	if err := lData.UpdateFromCSV("Attributes.csv", rows[0], attrSFlds,
		config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData := LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "*sessions;*cdrs",
		"FilterIDs":          "*string:Account:1007",
		"ActivationInterval": "2014-01-14T00:00:00Z",
		"Path":               "Account",
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
		config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData = LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "",
		"FilterIDs":          "",
		"ActivationInterval": "",
		"Path":               "Subject",
		"Initial":            "*any",
		"Substitute":         "1001",
		"Append":             "true",
		"Weight":             "",
	}
	if !reflect.DeepEqual(eLData, lData) {
		t.Errorf("expecting: %+v, received: %+v", eLData, lData)
	}
}

func TestDataUpdateFromCSVOneFile2(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.MetaVariable,
			Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "Contexts",
			Path:  "Contexts",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
		{Tag: "FilterIDs",
			Path:  "FilterIDs",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
		{Tag: "ActivationInterval",
			Path:  "ActivationInterval",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
		{Tag: "Path",
			Path:  "Path",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
		{Tag: "Initial",
			Path:  "Initial",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
		{Tag: "Substitute",
			Path:  "Substitute",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
		{Tag: "Append",
			Path:  "Append",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
		{Tag: "Weight",
			Path:  "Weight",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
	}

	rows := [][]string{
		{"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:Account:1007", "2014-01-14T00:00:00Z", "Account", "*any", "1001", "false", "10"},
		{"cgrates.org", "ATTR_1", "", "", "", "Subject", "*any", "1001", "true", ""},
	}
	lData := make(LoaderData)
	if err := lData.UpdateFromCSV("Attributes.csv", rows[0], attrSFlds,
		config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData := LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "*sessions;*cdrs",
		"FilterIDs":          "*string:Account:1007",
		"ActivationInterval": "2014-01-14T00:00:00Z",
		"Path":               "Account",
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
		config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), nil); err != nil {
		t.Error(err)
	}
	eLData = LoaderData{"Tenant": "cgrates.org",
		"ID":                 "ATTR_1",
		"Contexts":           "",
		"FilterIDs":          "",
		"ActivationInterval": "",
		"Path":               "Subject",
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
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.INFIELD_SEP),
			Mandatory: true},
		{Tag: "Contexts",
			Path:  "Contexts",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("*any", utils.INFIELD_SEP)},
		{Tag: "Path",
			Path:  "Path",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).5", utils.INFIELD_SEP)},
		{Tag: "Initial",
			Path:  "Initial",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.INFIELD_SEP)},
		{Tag: "Substitute",
			Path:  "Substitute",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.INFIELD_SEP)},
		{Tag: "Append",
			Path:  "Append",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("true", utils.INFIELD_SEP)},
		{Tag: "Weight",
			Path:  "Weight",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("10", utils.INFIELD_SEP)},
	}

	loadRun1 := map[string][]string{
		"File1.csv": {"ignored", "ignored", "ignored", "ignored", "ignored", "Subject", "*any", "1001", "ignored", "ignored"},
		"File2.csv": {"ignored", "ATTR_1"},
	}
	lData := make(LoaderData)
	for fName, record := range loadRun1 {
		if err := lData.UpdateFromCSV(fName, record, attrSFlds,
			config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), nil); err != nil {
			t.Error(err)
		}
	}

	eLData := LoaderData{"Tenant": "cgrates.org",
		"ID":         "ATTR_1",
		"Contexts":   "*any",
		"Path":       "Subject",
		"Initial":    "*any",
		"Substitute": "1001",
		"Append":     "true",
		"Weight":     "10",
	}
	if !reflect.DeepEqual(eLData, lData) {
		t.Errorf("expecting: %+v, received: %+v", eLData, lData)
	}
}

func TestRemoteHostLoaderData(t *testing.T) {
	record := []string{"ignored", "ignored", "Subject", "*any", "1001"}
	fNmae := "File1.csv"
	csvProv := newCsvProvider(record, fNmae)
	exp := "local"
	rcv := csvProv.RemoteHost()
	rcvStr := rcv.String()
	if !reflect.DeepEqual(exp, rcvStr) {
		t.Errorf("Expected %+v, received %+v", exp, rcv)
	}
}

func TestGetRateIDsLoaderData(t *testing.T) {
	ldrData := LoaderData{
		"File1.csv": []string{"Subject", "*any", "1001"},
	}
	expected := "cannot find RateIDs in <map[File1.csv:[Subject *any 1001]]>"
	if _, err := ldrData.GetRateIDs(); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestUpdateFromCsvParseValueError(t *testing.T) {
	ldrData := LoaderData{
		"File1.csv": []string{"Subject", "*any", "1001"},
	}
	tnt := config.NewRSRParsersMustCompile("asd{*duration_seconds}", utils.INFIELD_SEP)
	expected := "time: invalid duration \"asd\""
	if err := ldrData.UpdateFromCSV("File1.csv", nil, nil, tnt, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestUpdateFromCsvWithFiltersError(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Filters:   []string{"*string:~*req.Account:10"},
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.INFIELD_SEP),
			Filters:   []string{"*string:~*req.Account:10"},
			Mandatory: true},
	}
	loadRunStr := map[string][]string{
		"File1.csv": {"cgrates.org", "TEST_1"},
	}
	lData := make(LoaderData)

	dftCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(dftCfg, nil, dm)

	for fName, record := range loadRunStr {
		expected := "Ignoring record: [\"cgrates.org\" \"TEST_1\"] with error : strconv.Atoi: parsing \"Account\": invalid syntax"
		if err := lData.UpdateFromCSV(fName, record, attrSFlds,
			config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), filterS); err == nil || err.Error() != expected {
			t.Errorf("Expected %+v, received %+v", expected, err)
		}
	}
}

func TestUpdateFromCsvWithFiltersContinue(t *testing.T) {
	attrSFlds := []*config.FCTemplate{
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP),
			Filters:   []string{`*string:~*req.2:10`},
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.INFIELD_SEP),
			Filters:   []string{`*string:~*req.2:10`},
			Mandatory: true},
	}
	loadRunStr := map[string][]string{
		"File1.csv": {"Subject", "*any", "1001"},
	}
	lData := make(LoaderData)

	dftCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(dftCfg, nil, dm)

	for fName, record := range loadRunStr {
		if err := lData.UpdateFromCSV(fName, record, attrSFlds,
			config.NewRSRParsersMustCompile("cgrates.org", utils.INFIELD_SEP), filterS); err != nil {
			t.Error(err)
		}
	}
}

func TestLoadersFieldAsInterfaceError(t *testing.T) {
	loadRun1 := map[string][]string{
		"File1.csv": {"ignored", "ignored", "ignored", "ignored", "ignored", "Subject", "*any", "1001", "ignored", "ignored"},
	}
	csvProv := newCsvProvider(loadRun1["File1.csv"], "File1.csv")
	csvProv.String()

	expected := "invalid prefix for : [File2.csv]"
	if _, err := csvProv.FieldAsInterface([]string{"File2.csv"}); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "filter rule <[*file() ]> needs to end in )"
	if _, err := csvProv.FieldAsInterface([]string{"*file()", ""}); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+q", expected, err)
	}

	expected = "filter rule <[*file() File1.csv]> needs to end in )"
	if _, err := csvProv.FieldAsInterface([]string{"*file()", "File1.csv"}); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+q", expected, err)
	}
}
