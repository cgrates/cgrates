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
	"encoding/csv"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestLoaderProcessContentSingleFile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "AttributeFilterIDs",
				Path:  "AttributeFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "Path",
				Path:  "Path",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": &openedCSVFile{fileName: "Attributes.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}

	//cannot set AttributeProfile when dryrun is true
	ldr.dryRun = true
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//processContent successfully when dryrun is false
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": &openedCSVFile{fileName: "Attributes.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}

	eAP := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1", "con2", "con3"},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.InfieldSep),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.InfieldSep),
			}},
		Blocker: true,
		Weight:  20,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "ALS1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAP.Attributes, ap.Attributes) {
		t.Errorf("expecting: %s, \n received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}

	//cannot set AttributeProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{fileName: utils.AttributesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoaderProcessContentMultiFiles(t *testing.T) {
	file1CSV := `ignored,ignored,ignored,ignored,ignored,,*req.Subject,1001,ignored,ignored`
	file2CSV := `ignored,TestLoader2`
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContentMultiFiles",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaString,
				Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("*any", utils.InfieldSep)},
			{Tag: "Path",
				Path:  "Path",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.InfieldSep)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("10", utils.InfieldSep)},
		},
	}
	rdr1 := ioutil.NopCloser(strings.NewReader(file1CSV))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	rdr2 := ioutil.NopCloser(strings.NewReader(file2CSV))
	csvRdr2 := csv.NewReader(rdr2)
	csvRdr2.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"File1.csv": &openedCSVFile{fileName: "File1.csv",
				rdr: rdr1, csvRdr: csvRdr1},
			"File2.csv": &openedCSVFile{fileName: "File2.csv",
				rdr: rdr2, csvRdr: csvRdr2}},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}
	eAP := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "TestLoader2",
		FilterIDs: []string{},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "Subject",
				FilterIDs: []string{},
				Value:     config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			}},
		Weight: 10.0,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAP, ap) {
		t.Errorf("expecting: %s, \n received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}
}

func TestLoaderProcessResource(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "TTL",
				Path:  "UsageTTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "Limit",
				Path:  "Limit",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "AllocationMessage",
				Path:  "AllocationMessage",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "Thresholds",
				Path:  "Thresholds",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ResourcesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err != nil {
		t.Error(err)
	}
	eResPrf1 := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ResGroup21",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		UsageTTL:          time.Second,
		AllocationMessage: "call",
		Weight:            10,
		Limit:             2,
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{},
	}
	eResPrf2 := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ResGroup22",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		UsageTTL:          3600 * time.Second,
		AllocationMessage: "premium_call",
		Weight:            10,
		Limit:             2,
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{},
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup21",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eResPrf1, resPrf) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eResPrf1), utils.ToJSON(resPrf))
	}
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup22",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eResPrf2, resPrf) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eResPrf2), utils.ToJSON(resPrf))
	}
}

func TestLoaderProcessFilters(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessFilters",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "Element",
				Path:  "Element",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "Values",
				Path:  "Values",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.FiltersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			"Filters.csv": &openedCSVFile{fileName: "Filters.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}

	//Cannot set filterProfile when dryrun is true
	ldr.dryRun = true
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//processContent when dryrun is false
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(engine.FiltersCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			"Filters.csv": &openedCSVFile{fileName: "Filters.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err != nil {
		t.Error(err)
	}

	eFltr1 := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Values:  []string{"1001", "1002"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"10", "20"},
			},
			{
				Type:    utils.MetaRSR,
				Element: "~*req.Subject",
				Values:  []string{"~^1.*1$"},
			},
			{
				Type:    utils.MetaRSR,
				Element: "~*req.Destination",
				Values:  []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	if err := eFltr1.Compile(); err != nil {
		t.Error(err)
	}
	// Compile Value for rsr fields
	if err := eFltr1.Rules[2].CompileValues(); err != nil {
		t.Error(err)
	}
	// Compile Value for rsr fields
	if err := eFltr1.Rules[3].CompileValues(); err != nil {
		t.Error(err)
	}
	eFltr2 := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_DST_DE",
		Rules: []*engine.FilterRule{
			{
				Type:    "*destinations",
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"DST_DE"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	if err := eFltr2.Compile(); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if fltr, err := ldr.dm.GetFilter("cgrates.org", "FLTR_1",
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFltr1, fltr) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eFltr1), utils.ToJSON(fltr))
	}
	if fltr, err := ldr.dm.GetFilter("cgrates.org", "FLTR_DST_DE",
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFltr2, fltr) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eFltr2), utils.ToJSON(fltr))
	}
}

func TestLoaderProcessThresholds(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "MaxHits",
				Path:  "MaxHits",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "MinHits",
				Path:  "MinHits",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "MinSleep",
				Path:  "MinSleep",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "ActionIDs",
				Path:  "ActionIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "Async",
				Path:  "Async",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ThresholdsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{fileName: utils.ThresholdsCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eTh1 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "Threshold1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.RunID:*default"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  time.Second,
		Blocker:   true,
		Weight:    10,
		ActionIDs: []string{"THRESH1"},
		Async:     true,
	}
	aps, err := ldr.dm.GetThresholdProfile("cgrates.org", "Threshold1",
		true, false, utils.NonTransactional)
	sort.Strings(eTh1.FilterIDs)
	sort.Strings(aps.FilterIDs)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTh1, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eTh1), utils.ToJSON(aps))
	}

	//cannot set thresholdProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.ThresholdsCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			"Thresholds.csv": &openedCSVFile{fileName: "Thresholds.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessStats(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "QueueLength",
				Path:  "QueueLength",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "TTL",
				Path:  "TTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "MinItems",
				Path:  "MinItems",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "MetricIDs",
				Path:  "MetricIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "MetricFilterIDs",
				Path:  "MetricFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},

			{Tag: "ThresholdIDs",
				Path:  "ThresholdIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.StatsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{fileName: utils.StatsCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eSt1 := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TestStats",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         time.Second,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*sum#~*req.Value",
			},
			{
				MetricID: "*average#~*req.Value",
			},
			{
				MetricID: "*sum#~*req.Usage",
			},
		},
		ThresholdIDs: []string{"Th1", "Th2"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     2,
	}

	aps, err := ldr.dm.GetStatQueueProfile("cgrates.org", "TestStats",
		true, false, utils.NonTransactional)
	//sort the slices of Metrics
	sort.Slice(eSt1.Metrics, func(i, j int) bool { return eSt1.Metrics[i].MetricID < eSt1.Metrics[j].MetricID })
	sort.Slice(aps.Metrics, func(i, j int) bool { return aps.Metrics[i].MetricID < aps.Metrics[j].MetricID })
	sort.Strings(eSt1.ThresholdIDs)
	sort.Strings(aps.ThresholdIDs)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSt1, aps) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eSt1), utils.ToJSON(aps))
	}

	//cannot set statsProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.StatsCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			"Stats.csv": &openedCSVFile{fileName: "Stats.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessStatsWrongMetrics(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessStatsWrongMetrics",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "MetricIDs",
				Path:  "MetricIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep)},
		},
	}
	statsCsv := `
#Metrics[0],Stored[1]
not_a_valid_metric_type,true,
`
	rdr := ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{fileName: utils.StatsCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expected := "SERVER_ERROR: unsupported metric type <not_a_valid_metric_type>"
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//initialize again but with a valid metric and false stored field
	statsCsv = `
#Metrics[0],Stored[1]
*sum#~*req.Value,false
`
	rdr = ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{fileName: utils.StatsCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessRoutes(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "Sorting",
				Path:  "Sorting",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "SortingParameters",
				Path:  "SortingParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "RouteID",
				Path:  "RouteID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "RouteFilterIDs",
				Path:  "RouteFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "RouteAccountIDs",
				Path:  "RouteAccountIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "RouteRatingPlanIDs",
				Path:  "RouteRatingplanIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "RouteResourceIDs",
				Path:  "RouteResourceIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Tag: "RouteStatIDs",
				Path:  "RouteStatIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},
			{Tag: "RouteWeight",
				Path:  "RouteWeight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
			{Tag: "RouteBlocker",
				Path:  "RouteBlocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep)},
			{Tag: "RouteParameters",
				Path:  "RouteParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.RoutesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{fileName: utils.RoutesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}

	eSp := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "RoutePrf1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	sort.Slice(eSp.Routes, func(i, j int) bool {
		return strings.Compare(eSp.Routes[i].ID+strings.Join(eSp.Routes[i].FilterIDs, utils.ConcatenatedKeySep),
			eSp.Routes[j].ID+strings.Join(eSp.Routes[j].FilterIDs, utils.ConcatenatedKeySep)) < 0
	})

	aps, err := ldr.dm.GetRouteProfile("cgrates.org", "RoutePrf1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(aps.Routes, func(i, j int) bool {
		return strings.Compare(aps.Routes[i].ID+strings.Join(aps.Routes[i].FilterIDs, utils.ConcatenatedKeySep),
			aps.Routes[j].ID+strings.Join(aps.Routes[j].FilterIDs, utils.ConcatenatedKeySep)) < 0
	})
	if !reflect.DeepEqual(eSp, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSp), utils.ToJSON(aps))
	}

	//cannot set RoutesProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.RoutesCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{fileName: utils.RoutesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessAccountProfiles(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessAccountProfiles",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAccountProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	actPrflCsv := `
#Tenant[0],ID[1]
cgrates.org,ACTPRF_ID1
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrflCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv, rdr: rdr,
				csvRdr: csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set an AccountProfile while dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(actPrflCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv, rdr: rdr,
				csvRdr: csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessChargers(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "RunID",
				Path:  "RunID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "AttributeIDs",
				Path:  "AttributeIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ChargersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{fileName: utils.ChargersCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eCharger1 := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
		Weight:       20,
	}

	if rcv, err := ldr.dm.GetChargerProfile("cgrates.org", "Charger1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCharger1, rcv) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eCharger1), utils.ToJSON(rcv))
	}

	//cannot set chargerProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.ChargersCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{fileName: utils.ChargersCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessDispatches(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:   "Subsystems",
				Path:  "Subsystems",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
			},
			{
				Tag:   "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
			},
			{
				Tag:   "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
			},
			{
				Tag:   "Strategy",
				Path:  "Strategy",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
			},
			{
				Tag:   "StrategyParameters",
				Path:  "StrategyParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
			},
			{
				Tag:   "ConnID",
				Path:  "ConnID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
			},
			{
				Tag:   "ConnFilterIDs",
				Path:  "ConnFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
			},
			{
				Tag:   "ConnWeight",
				Path:  "ConnWeight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
			},
			{
				Tag:   "ConnBlocker",
				Path:  "ConnBlocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
			},
			{
				Tag:   "ConnParameters",
				Path:  "ConnParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
			},
			{
				Tag:   "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
			},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.DispatcherCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eDisp := &engine.DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "D1",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		StrategyParams: map[string]interface{}{},
		Strategy:       "*first",
		Weight:         20,
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{"*gt:~*req.Usage:10"},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.56.203"},
				Blocker:   false,
			},
			&engine.DispatcherHostProfile{
				ID:        "C2",
				FilterIDs: []string{"*lt:~*req.Usage:10"},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.56.204"},
				Blocker:   false,
			},
		},
	}

	rcv, err := ldr.dm.GetDispatcherProfile("cgrates.org", "D1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(eDisp.Hosts, func(i, j int) bool { return eDisp.Hosts[i].ID < eDisp.Hosts[j].ID })
	sort.Slice(rcv.Hosts, func(i, j int) bool { return rcv.Hosts[i].ID < rcv.Hosts[j].ID })
	if !reflect.DeepEqual(eDisp, rcv) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eDisp), utils.ToJSON(rcv))
	}

	//cannot set DispatchersProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.DispatcherCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessDispatcheHosts(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{
				Tag:       "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:   "Address",
				Path:  "Address",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
			},
			{
				Tag:   "Transport",
				Path:  "Transport",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
			},
			{
				Tag:   "TLS",
				Path:  "TLS",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
			},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.DispatcherHostCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaDispatcherHosts, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eDispHost := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:        "ALL1",
			Address:   "127.0.0.1:2012",
			Transport: utils.MetaJSON,
			TLS:       true,
		},
	}

	rcv, err := ldr.dm.GetDispatcherHost("cgrates.org", "ALL1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eDispHost, rcv) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eDispHost), utils.ToJSON(rcv))
	}

	//cannot set DispatcherHostProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.DispatcherHostCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.processContent(utils.MetaDispatcherHosts, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderRemoveContentSingleFile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": &openedCSVFile{fileName: "Attributes.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	// Add two attributeProfiles
	ap := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1", "con2", "con3"},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.InfieldSep),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.InfieldSep),
			}},
		Blocker: true,
		Weight:  20,
	}
	if err := ldr.dm.SetAttributeProfile(ap, true); err != nil {
		t.Error(err)
	}
	ap.ID = "Attr2"
	if err := ldr.dm.SetAttributeProfile(ap, true); err != nil {
		t.Error(err)
	}

	if err := ldr.removeContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	// make sure the first attribute is deleted
	if _, err := ldr.dm.GetAttributeProfile("cgrates.org", "ALS1",
		true, false, utils.NonTransactional); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// the second should be there
	if rcv, err := ldr.dm.GetAttributeProfile("cgrates.org", "Attr2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ap, rcv) {
		t.Errorf("expecting: %s, \n received: %s",
			utils.ToJSON(ap), utils.ToJSON(rcv))
	}

	//now should be empty, nothing to remove
	if err := ldr.removeContent(utils.MetaAttributes, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": &openedCSVFile{fileName: "Attributes.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessRateProfile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessRateProfile",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "Weights",
				Path:  "Weights",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "MinCost",
				Path:  "MinCost",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "MaxCost",
				Path:  "MaxCost",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "MaxCostStrategy",
				Path:  "MaxCostStrategy",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "RateID",
				Path:  "RateID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "RateFilterIDs",
				Path:  "RateFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "RateActivationTimes",
				Path:  "RateActivationTimes",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Tag: "RateWeights",
				Path:  "RateWeights",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},
			{Tag: "RateBlocker",
				Path:  "RateBlocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
			{Tag: "RateIntervalStart",
				Path:  "RateIntervalStart",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep)},
			{Tag: "RateFixedFee",
				Path:  "RateFixedFee",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep)},
			{Tag: "RateRecurrentFee",
				Path:  "RateRecurrentFee",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep)},
			{Tag: "RateUnit",
				Path:  "RateUnit",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep)},
			{Tag: "RateIncrement",
				Path:  "RateIncrement",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.17", utils.InfieldSep)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.RateProfileCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	eRatePrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: time.Minute,
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(89, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(564, 4),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf), utils.ToJSON(rcv))
	}

	//cannot set RateProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.RateProfileCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

}

func TestLoaderProcessRateProfileRates(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessRateProfile",
		bufLoaderData: make(map[string][]LoaderData),
		flagsTpls:     make(map[string]utils.FlagsWithParams),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Tag: "Weights",
				Path:  "Weights",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Tag: "MinCost",
				Path:  "MinCost",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Tag: "MaxCost",
				Path:  "MaxCost",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Tag: "MaxCostStrategy",
				Path:  "MaxCostStrategy",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Tag: "RateID",
				Path:  "RateID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Tag: "RateFilterIDs",
				Path:  "RateFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Tag: "RateActivationTimes",
				Path:  "RateActivationTimes",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Tag: "RateWeights",
				Path:  "RateWeights",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},
			{Tag: "RateBlocker",
				Path:  "RateBlocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
			{Tag: "RateIntervalStart",
				Path:  "RateIntervalStart",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep)},
			{Tag: "RateFixedFee",
				Path:  "RateFixedFee",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep)},
			{Tag: "RateRecurrentFee",
				Path:  "RateRecurrentFee",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep)},
			{Tag: "RateUnit",
				Path:  "RateUnit",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep)},
			{Tag: "RateIncrement",
				Path:  "RateIncrement",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.17", utils.InfieldSep)},
		},
	}
	ratePrfCnt1 := `
#Tenant,ID,FilterIDs,ActivationInterval,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationTimes,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0.4,0.12,1m,1m
cgrates.org,RP1,,,,,,,RT_WEEK,,,,,1m,,0.06,1m,1s
`
	ratePrfCnt2 := `
#Tenant,ID,FilterIDs,ActivationInterval,Weights,RoundingMethod,RoundingDecimals,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationTimes,RateWeights,RateBlocker,RateIntervalStart,RateValue,RateUnit,RateIncrement
cgrates.org,RP1,,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,,0.06,1m,1s
cgrates.org,RP1,,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,,0.06,1m,1s
`
	rdr1 := ioutil.NopCloser(strings.NewReader(ratePrfCnt1))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr1, csvRdr: csvRdr1}},
	}
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	eRatePrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(4, 1),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: time.Minute,
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf), utils.ToJSON(rcv))
	}

	rdr2 := ioutil.NopCloser(strings.NewReader(ratePrfCnt2))
	csvRdr2 := csv.NewReader(rdr2)
	csvRdr2.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr2, csvRdr: csvRdr2}},
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eRatePrf = &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(4, 1),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: time.Minute,
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	rcv, err = ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf), utils.ToJSON(rcv))
	}

}

func TestLoaderRemoveRateProfileRates(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderRemoveRateProfileRates",
		bufLoaderData: make(map[string][]LoaderData),
		flagsTpls:     make(map[string]utils.FlagsWithParams),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "RateIDs",
				Path:  "RateIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
		},
	}

	rPfr := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
					{
						IntervalStart: time.Minute,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
				},
			},
		},
	}
	if err := ldr.dm.SetRateProfile(rPfr, true); err != nil {
		t.Error(err)
	}
	rPfr2 := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP2",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
					{
						IntervalStart: time.Minute,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
				},
			},
		},
	}
	if err := ldr.dm.SetRateProfile(rPfr2, true); err != nil {
		t.Error(err)
	}

	ratePrfCnt1 := `
#Tenant,ID,RateIDs
cgrates.org,RP1,RT_WEEKEND
`
	ratePrfCnt2 := `
#Tenant,ID,RateIDs
cgrates.org,RP2,RT_WEEKEND;RT_CHRISTMAS
cgrates.org,RP1,
`
	rdr1 := ioutil.NopCloser(strings.NewReader(ratePrfCnt1))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr1, csvRdr: csvRdr1}},
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}

	eRatePrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
					{
						IntervalStart: time.Minute,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
				},
			},
		},
	}
	rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf), utils.ToJSON(rcv))
	}

	rdr2 := ioutil.NopCloser(strings.NewReader(ratePrfCnt2))
	csvRdr2 := csv.NewReader(rdr2)
	csvRdr2.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr2, csvRdr: csvRdr2}},
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}

	eRatePrf2 := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP2",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
					},
					{
						IntervalStart: time.Minute,
					},
				},
			},
		},
	}
	rcv, err = ldr.dm.GetRateProfile("cgrates.org", "RP2",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf2.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf2) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf2), utils.ToJSON(rcv))
	}

	eRatePrf3 := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates:           map[string]*engine.Rate{},
	}
	rcv, err = ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()
	eRatePrf3.Compile()
	if !reflect.DeepEqual(rcv, eRatePrf3) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf3), utils.ToJSON(rcv))
	}
}

func TestRemoveRateProfileFlagsError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderRemoveRateProfileRates",
		bufLoaderData: make(map[string][]LoaderData),
		flagsTpls:     make(map[string]utils.FlagsWithParams),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	ratePrfCnt1 := `
#Tenant,ID
cgrates.org,RP2
`
	rPfr := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
	}
	if err := ldr.dm.SetRateProfile(rPfr, true); err != nil {
		t.Error(err)
	}

	rdr1 := ioutil.NopCloser(strings.NewReader(ratePrfCnt1))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{fileName: utils.RateProfilesCsv,
				rdr: rdr1, csvRdr: csvRdr1}},
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	expectedErr := "cannot find RateIDs in <map[ID:RP2 Tenant:cgrates.org]>"
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestNewLoaderWithMultiFiles(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)

	ldrCfg := config.CgrConfig().LoaderCfg()[0].Clone()
	ldrCfg.Data[0].Fields = []*config.FCTemplate{
		{Tag: "TenantID",
			Path:      "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			Mandatory: true},
		{Tag: "ProfileID",
			Path:      "ID",
			Type:      utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.InfieldSep),
			Mandatory: true},
		{Tag: "Contexts",
			Path:  "Contexts",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("*any", utils.InfieldSep)},
		{Tag: "Path",
			Path:  "Path",
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.InfieldSep)},
		{Tag: "Value",
			Path:  "Value",
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.InfieldSep)},
		{Tag: "Weight",
			Path:  "Weight",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("10", utils.InfieldSep)},
	}
	ldr := NewLoader(engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil), ldrCfg, "", nil, nil, nil)

	openRdrs := make(utils.StringSet)
	for _, rdr := range ldr.rdrs {
		for fileName := range rdr {
			openRdrs.Add(fileName)
		}
	}
	expected := utils.StringSet{
		utils.AttributesCsv:         {},
		utils.ChargersCsv:           {},
		utils.DispatcherHostsCsv:    {},
		utils.DispatcherProfilesCsv: {},
		"File1.csv":                 {},
		"File2.csv":                 {},
		utils.FiltersCsv:            {},
		utils.RateProfilesCsv:       {},
		utils.ResourcesCsv:          {},
		utils.RoutesCsv:             {},
		utils.StatsCsv:              {},
		utils.ThresholdsCsv:         {},
		utils.ActionProfilesCsv:     {},
		utils.AccountProfilesCsv:    {},
	}
	if !reflect.DeepEqual(expected, openRdrs) {
		t.Errorf("Expected %s,received %s", utils.ToJSON(expected), utils.ToJSON(openRdrs))
	}
}

func TestLoaderActionProfile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "FilterIDs",
				Path:   "FilterIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActivationInterval",
				Path:   "ActivationInterval",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "Weight",
				Path:   "Weight",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "Schedule",
				Path:   "Schedule",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "TargetType",
				Path:   "TargetType",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "TargetIDs",
				Path:   "TargetIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionID",
				Path:   "ActionID",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionFilterIDs",
				Path:   "ActionFilterIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionBlocker",
				Path:   "ActionBlocker",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionTTL",
				Path:   "ActionTTL",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionType",
				Path:   "ActionType",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionOpts",
				Path:   "ActionOpts",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionPath",
				Path:   "ActionPath",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionValue",
				Path:   "ActionValue",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
				Layout: time.RFC3339},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ActionProfileCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{fileName: utils.ActionProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}

	expected := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ONE_TIME_ACT",
		FilterIDs: []string{},
		Weight:    10,
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: utils.NewStringSet([]string{"1001", "1002"}),
		},
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestBalance.Value",
					Value: "10",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_DATA",
				Type: "*set_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestDataBalance.Type",
					Value: "*data",
				}},
			},
			{
				ID:   "TOPUP_TEST_DATA",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestDataBalance.Value",
					Value: "1024",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_VOICE",
				Type: "*set_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestVoiceBalance.Type",
					Value: "*voice",
				}},
			},
			{
				ID:   "TOPUP_TEST_VOICE",
				Type: "*add_balance",
				Diktats: []*engine.APDiktat{{
					Path:  "*balance.TestVoiceBalance.Value",
					Value: "15m15s",
				}, {
					Path:  "*balance.TestVoiceBalance2.Value",
					Value: "15m15s",
				}},
			},
		},
	}

	aps, err := ldr.dm.GetActionProfile("cgrates.org", "ONE_TIME_ACT",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(expected), utils.ToJSON(aps))
	}

	//cannot set ActionProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(engine.ActionProfileCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{fileName: utils.ActionProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderWrongCsv(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderWrongCsv",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "FilterIDs",
				Path:   "FilterIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActivationInterval",
				Path:   "ActivationInterval",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "Weight",
				Path:   "Weight",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "Schedule",
				Path:   "Schedule",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "AccountIDs",
				Path:   "AccountIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionID",
				Path:   "ActionID",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionFilterIDs",
				Path:   "ActionFilterIDs",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionBlocker",
				Path:   "ActionBlocker",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionTTL",
				Path:   "ActionTTL",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionType",
				Path:   "ActionType",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionOpts",
				Path:   "ActionOpts",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionPath",
				Path:   "ActionPath",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
				Layout: time.RFC3339},
			{Tag: "ActionValue",
				Path:   "ActionValue",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
				Layout: time.RFC3339},
		},
	}

	//Not a valid comment beginning of csv
	newCSVContentMiss := `
//Tenant,ID,FilterIDs,ActivationInterval,Weight,Schedule,AccountIDs,ActionID,ActionFilterIDs,ActionBLocker,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,,10,*asap,1001;1002,TOPUP,,false,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_DATA,,false,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_DATA,,false,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_VOICE,,false,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
`

	rdr := ioutil.NopCloser(strings.NewReader(newCSVContentMiss))
	csvRdr := csv.NewReader(rdr)
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expectedErr := "invalid syntax"
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}

	//Missing fields in csv eg:ActionBLocker
	newCSVContent := `
//Tenant,ID,FilterIDs,ActivationInterval,Weight,Schedule,AccountIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,,10,*asap,1001;1002,TOPUP,,false,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_DATA,,false,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_DATA,,false,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,SET_BALANCE_TEST_VOICE,,false,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,TOPUP_TEST_VOICE,,false,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
`
	rdr = ioutil.NopCloser(strings.NewReader(newCSVContent))
	csvRdr = csv.NewReader(rdr)
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expectedErr = "invalid syntax"
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func TestLoaderActionProfileAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderActionProfileAsStructErrType",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
			{Tag: "ActionBlocker",
				Path:   "ActionBlocker",
				Type:   utils.MetaVariable,
				Value:  config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
				Layout: time.RFC3339},
		},
	}
	actPrfCsv := `
#Tenant,ID,ActionBlocker
cgrates.org,12,NOT_A_BOOLEAN
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{fileName: utils.ActionProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expErr := `strconv.ParseBool: parsing "NOT_A_BOOLEAN": invalid syntax`
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func TestLoaderActionProfileAsStructErrTConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderActionProfileAsStructErrType",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "ActivationInterval",
				Path:      "ActivationInterval",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339},
		},
	}
	actPrfCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{fileName: utils.ActionProfilesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expErr := `Unsupported time format`
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func TestLoaderAttributesAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderAttributesAsStructErrType",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	attributeCsv := `
#Weight
true
`
	rdr := ioutil.NopCloser(strings.NewReader(attributeCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{fileName: utils.AttributesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := "strconv.ParseFloat: parsing \"true\": invalid syntax"
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Error(err)
	}
}

func TestLoaderAttributesAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderAttributesAsStructErrConversion",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	attributeCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(attributeCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{fileName: utils.AttributesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadResourcesAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadResourcesAsStructErr",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	resourcesCsv := `
#Blocker
NOT_A_BOOLEAN
`
	rdr := ioutil.NopCloser(strings.NewReader(resourcesCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{fileName: utils.ResourcesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := "strconv.ParseBool: parsing \"NOT_A_BOOLEAN\": invalid syntax"
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadResourcesAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadResourcesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "UsageTTL",
				Path:  "UsageTTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	resourcesCsv := `
#UsageTTL
12ss
`
	rdr := ioutil.NopCloser(strings.NewReader(resourcesCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{fileName: utils.ResourcesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := "time: unknown unit \"ss\" in duration \"12ss\""
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadFiltersAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadFiltersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	filtersCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(filtersCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			utils.FiltersCsv: &openedCSVFile{
				fileName: utils.FiltersCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadFiltersAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadFiltersAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	filtersCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(filtersCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			utils.FiltersCsv: &openedCSVFile{
				fileName: utils.FiltersCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadStatsAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadStatsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	statsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(statsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadStatsAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadStatsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	statsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(statsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadThresholdsAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadThresholdsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadThresholdsAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadThresholdsAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadRoutesAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadRoutesAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{
				fileName: utils.RoutesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadRoutesAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadRoutesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{
				fileName: utils.RoutesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadChargersAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadChargersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{
				fileName: utils.ChargersCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadChargersAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadChargersAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{
				fileName: utils.ChargersCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadDispatchersAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadDispatchersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadDispatcherAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadDispatcherHostsAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadDispatcherHostsAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadDispatcherHostsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherHostsCsv: &openedCSVFile{
				fileName: utils.DispatcherHostsCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaDispatcherHosts, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadRateProfilesAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadRateProfilesAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadRateProfilesAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadRateProfilesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadAccountProfilesAsStructErrConversion(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadAccountProfilesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAccountProfiles: {
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	actPrfCsv := `
#ActivationInterval
* * * * * * *
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "Unsupported time format"
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestProcessContentAccountProfileAsTPError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestProcessContentAccountProfileAsTPError",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAccountProfiles: {
			{Tag: "Tenant",
				Path:  "Tenant",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
			{Tag: "ID",
				Path:  "ID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep)},
			{Tag: "BalanceID",
				Path:  "BalanceID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Tag: "BalanceUnitFactors",
				Path:  "BalanceUnitFactors",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
		},
	}

	accPrfCSv := `
#Tenant,ID,BalanceID,BalanceUnitFactors
cgrates.org,1001,MonetaryBalance,fltr1&fltr2;100;fltr3
`
	rdr := ioutil.NopCloser(strings.NewReader(accPrfCSv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "invlid key: <fltr1&fltr2;100;fltr3> for BalanceUnitFactors"
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadAccountProfilesAsStructErrType(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadAccountProfilesAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAccountProfiles: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	actPrfCsv := `
#PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expectedErr := "cannot update unsupported struct field: 0"
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoadAndRemoveResources(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadAndRemoveResources",
		bufLoaderData: make(map[string][]LoaderData),
		dryRun:        true,
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	resourcesCSV := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}
	//empty database
	if _, err := ldr.dm.GetResourceProfile("cgrates.org", "NewRes1", false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	//because of dryrun, database will be empty again
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err != nil {
		t.Error(err)
	}

	ldr.dryRun = false
	//reinitialized reader because after first process the reader is at the end of the file
	rdr = ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}

	resPrf := &engine.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "NewRes1",
		FilterIDs:    []string{},
		ThresholdIDs: []string{},
	}
	//NOT_FOUND because is resourceProfile is not set
	if _, err := ldr.dm.GetResourceProfile("cgrates.org", "NewRes1", false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err != nil {
		t.Error(err)
	}

	rcv, err := ldr.dm.GetResourceProfile("cgrates.org", "NewRes1", false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resPrf, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(resPrf), utils.ToJSON(rcv))
	}

	//reinitialized reader because seeker it s at the end of the file
	rdr = ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}

	//cannot remove when dryrun is on true
	ldr.dryRun = true
	if err := ldr.removeContent(utils.MetaResources, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//remove successfully when dryrun is false
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}
	if err := ldr.removeContent(utils.MetaResources, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing in database
	if _, err := ldr.dm.GetResourceProfile("cgrates.org", "NewRes1", false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	//nothing to remove
	if err := ldr.removeContent(utils.MetaResources, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot set again ResourceProfile when dataManager is nil
	ldr.dm = nil
	rdr = ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveFilterContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveFilterContents",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	filtersCsv := `
#Tenant[0],ID[0]
cgrates.org,FILTERS_REM_1
`
	rdr := ioutil.NopCloser(strings.NewReader(filtersCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			utils.FiltersCsv: &openedCSVFile{
				fileName: utils.FiltersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	eFltr1 := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FILTERS_REM_1",
	}
	if err := ldr.dm.SetFilter(eFltr1, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaFilters, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaFilters, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove Filter when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(filtersCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			utils.FiltersCsv: &openedCSVFile{
				fileName: utils.FiltersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaFilters, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again FiltersProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(filtersCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			utils.FiltersCsv: &openedCSVFile{
				fileName: utils.FiltersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveStatsContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	statsCsv := `
#Tenant[0],ProfileID[1]
cgrates.org,REM_STATS_1
`
	rdr := ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expStats := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "REM_STATS_1",
	}
	if err := ldr.dm.SetStatQueueProfile(expStats, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove statsQueueProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again StatsProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveThresholdsContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveThresholdsContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	thresholdsCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_THRESHOLDS_1,
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expThresholdPrf := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "REM_THRESHOLDS_1",
	}
	if err := ldr.dm.SetThresholdProfile(expThresholdPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaThresholds, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaThresholds, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove statsQueueProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaThresholds, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again ThresholdsProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveRoutesContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveRoutesContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	routesCsv := `
#Tenant[0],ID[1]
cgrates.org,ROUTES_REM_1
`
	rdr := ioutil.NopCloser(strings.NewReader(routesCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{
				fileName: routesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expRoutes := &engine.RouteProfile{
		Tenant: "cgrates.org",
		ID:     "ROUTES_REM_1",
	}
	if err := ldr.dm.SetRouteProfile(expRoutes, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaRoutes, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaRoutes, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove routeProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(routesCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{
				fileName: routesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaRoutes, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again RoutesProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(routesCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: {
			utils.RoutesCsv: &openedCSVFile{
				fileName: routesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveChargersContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveChargersContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	routesCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_ROUTES_1
`
	rdr := ioutil.NopCloser(strings.NewReader(routesCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{
				fileName: utils.ChargersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expChargers := &engine.ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "REM_ROUTES_1",
	}
	if err := ldr.dm.SetChargerProfile(expChargers, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaChargers, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remvoe from database
	if err := ldr.removeContent(utils.MetaChargers, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove ChargersProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(routesCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{
				fileName: utils.ChargersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaChargers, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again ChargersProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(routesCsv))
	rdr = ioutil.NopCloser(strings.NewReader(routesCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: &openedCSVFile{
				fileName: utils.ChargersCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveDispatchersContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveDispatchersContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	dispatchersCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_DISPATCHERS_1
`
	rdr := ioutil.NopCloser(strings.NewReader(dispatchersCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expDispatchers := &engine.DispatcherProfile{
		Tenant: "cgrates.org",
		ID:     "REM_DISPATCHERS_1",
	}
	if err := ldr.dm.SetDispatcherProfile(expDispatchers, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaDispatchers, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remvoe from database
	if err := ldr.removeContent(utils.MetaDispatchers, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove DispatchersProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(dispatchersCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaDispatchers, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//cannot set again DispatchersProfile when dataManager is nil
	ldr.dm = nil
	ldr.dryRun = false
	rdr = ioutil.NopCloser(strings.NewReader(dispatchersCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveDispatcherHostsContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveDispatcherHostsContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	dispatchersHostsCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_DISPATCHERH_1
`
	rdr := ioutil.NopCloser(strings.NewReader(dispatchersHostsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherHostsCsv: &openedCSVFile{
				fileName: utils.DispatcherHostsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expDispatchers := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID: "REM_DISPATCHERH_1",
		},
	}
	if err := ldr.dm.SetDispatcherHost(expDispatchers); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaDispatcherHosts, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaDispatcherHosts, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove DispatcherHosts when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(dispatchersHostsCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherHostsCsv: &openedCSVFile{
				fileName: utils.DispatcherHostsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaDispatcherHosts, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestProcessContentEmptyDataBase(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            nil,
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{
				Tag:       "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:   "Address",
				Path:  "Address",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
			},
			{
				Tag:   "Transport",
				Path:  "Transport",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
			},
			{
				Tag:   "TLS",
				Path:  "TLS",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
			},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.DispatcherHostCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherProfilesCsv: &openedCSVFile{
				fileName: utils.DispatcherProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expectedErr := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaDispatcherHosts, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestRemoveRateProfileContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveRateProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	rtPrfCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_RATEPROFILE_1
`
	rdr := ioutil.NopCloser(strings.NewReader(rtPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expRtPrf := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "REM_RATEPROFILE_1",
	}
	if err := ldr.dm.SetRateProfile(expRtPrf, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remvoe from database
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove DispatcherHosts when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(rtPrfCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

	ldr.dryRun = true
	ldr.flagsTpls = map[string]utils.FlagsWithParams{
		utils.MetaRateProfiles: {
			utils.MetaPartial: nil,
		},
	}
	rdr = ioutil.NopCloser(strings.NewReader(rtPrfCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestRemoveActionProfileContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	actPrfCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_ACTPROFILE_1
`
	//cannot set ActionProfile when dataManager is nil
	ldr.dm = nil
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//set dataManager
	ldr.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	actRtPrf := &engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "REM_ACTPROFILE_1",
	}
	if err := ldr.dm.SetActionProfile(actRtPrf, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaActionProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaActionProfiles, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove DispatcherHosts when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaActionProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestRemoveAccountProfileContent(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		dm:            nil,
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAccountProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	acntPrfCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_ACTPROFILE_1
`
	rdr := ioutil.NopCloser(strings.NewReader(acntPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaAccountProfiles, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//set dataManager
	ldr.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	acntPrf := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     "REM_ACTPROFILE_1",
	}
	if err := ldr.dm.SetAccountProfile(acntPrf, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaAccountProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//nothing to remove from database
	if err := ldr.removeContent(utils.MetaAccountProfiles, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//cannot remove AccountProfile when dryrun is true
	ldr.dryRun = true
	rdr = ioutil.NopCloser(strings.NewReader(acntPrfCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAccountProfiles: {
			utils.AccountProfilesCsv: &openedCSVFile{
				fileName: utils.AccountProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	if err := ldr.removeContent(utils.MetaAccountProfiles, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestRemoveContentError1(t *testing.T) {
	//use actionProfile to generate an error by giving a wrong csv
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	//wrong start at the beginning of csv
	actPrfCsv := `
//Tenant[0]
cgrates.org,REM_ACTPROFILE_s
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	actRtPrf := &engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "REM_ACTPROFILE_s",
	}
	expectedErr := "NOT_FOUND"
	if err := ldr.dm.SetActionProfile(actRtPrf, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestRemoveContentError2(t *testing.T) {
	//use actionProfile to generate an error by giving a wrong csv
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveActionProfileContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaActionProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	//wrong start at the beginning of csv
	actPrfCsv := `
Tenant[0],ID[1]
cgrates.org,REM_ACTPROFILE_s
`
	rdr := ioutil.NopCloser(strings.NewReader(actPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaActionProfiles: {
			utils.ActionProfilesCsv: &openedCSVFile{
				fileName: utils.ActionProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	actRtPrf := &engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "REM_ACTPROFILE_s",
	}
	expectedErr := "NOT_FOUND"
	if err := ldr.dm.SetActionProfile(actRtPrf, true); err != nil {
		t.Error(err)
	} else if err := ldr.removeContent(utils.MetaActionProfiles, utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoaderListenAndServe(t *testing.T) {
	ldr := &Loader{}
	stopChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(10)
		stopChan <- struct{}{}
	}()

	if err := ldr.ListenAndServe(stopChan); err != nil {
		t.Error(err)
	}

	ldr.runDelay = -1
	if err := ldr.ListenAndServe(stopChan); err != nil {
		t.Error(err)
	}

	ldr.runDelay = 1
	if err := ldr.ListenAndServe(stopChan); err != nil {
		t.Error(err)
	}
}

func TestRemoveRateProfileRatesError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveRateProfileRatesMockError",
		bufLoaderData: make(map[string][]LoaderData),
		flagsTpls:     make(map[string]utils.FlagsWithParams),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Tag: "RateIDs",
				Path:  "RateIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
		},
	}
	rtPrfCsv := `
#Tenant[0],ID[1],RateIDs[2]
cgrates.org,REM_RATEPROFILE_1,RT_WEEKEND
`
	rdr := ioutil.NopCloser(strings.NewReader(rtPrfCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expRtPrf := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "REM_RATEPROFILE_1",
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	ldr.dm = nil
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(err))
	}

	rdr = ioutil.NopCloser(strings.NewReader(rtPrfCsv))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{"INVALID_FLAGS"})
	expected = "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(err))
	}

	ldr.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	if err := ldr.dm.SetRateProfile(expRtPrf, true); err != nil {
		t.Error(err)
	}

	ldr.dm = nil
	ldr.flagsTpls[utils.MetaRateProfiles] = utils.FlagsWithParamsFromSlice([]string{utils.MetaPartial})
	expected = "NO_DATA_BASE_CONNECTION"
	if err := ldr.removeContent(utils.MetaRateProfiles, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(err))
	}
}

func TestRemoveThresholdsMockError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveThresholdsMockError",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	thresholdsCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_THRESHOLDS_1,
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			utils.ThresholdsCsv: &openedCSVFile{
				fileName: utils.ThresholdsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}

	expThresholdPrf := &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "REM_THRESHOLDS_1",
	}

	if err := ldr.dm.SetThresholdProfile(expThresholdPrf, true); err != nil {
		t.Error(err)
	}

	newData := &dataDBMockError{}
	ldr.dm = engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	expected := "NO_DATA_BASE_CONNECTION"
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.removeContent(utils.MetaThresholds, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

}

func TestRemoveStatQueueMockError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestRemoveStatQueueError",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	statsCsv := `
#Tenant[0],ProfileID[1]
cgrates.org,REM_STATS_1
`
	rdr := ioutil.NopCloser(strings.NewReader(statsCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{
				fileName: utils.StatsCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}
	expStats := &engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "REM_STATS_1",
	}

	if err := ldr.dm.SetStatQueueProfile(expStats, true); err != nil {
		t.Error(err)
	}

	newData := &dataDBMockError{}
	ldr.dm = engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	expected := "NO_DATA_BASE_CONNECTION"

	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveResourcesMockError(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	ldr := &Loader{
		ldrID:         "TestLoadAndRemoveResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	resourcesCSV := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := ioutil.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}

	resPrf := &engine.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "NewRes1",
	}

	if err := ldr.dm.SetResourceProfile(resPrf, true); err != nil {
		t.Error(err)
	}

	newData := &dataDBMockError{}
	ldr.dm = engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	expected := "NO_DATA_BASE_CONNECTION"

	if err := ldr.removeContent(utils.MetaResources, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderHandleFolder(t *testing.T) {
	stopChan := make(chan struct{}, 1)
	stopChan <- struct{}{}
	ldr := &Loader{
		ldrID:    "TestLoaderHandleFolder",
		runDelay: 1,
	}
	ldr.handleFolder(stopChan)
}

func TestLoaderServiceEnabled(t *testing.T) {
	//THis is an empty loader, so there is not an active loader
	ldrs := &LoaderService{}
	if rcv := ldrs.Enabled(); rcv {
		t.Errorf("Expected false, received %+v", rcv)
	}
}
