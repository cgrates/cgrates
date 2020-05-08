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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLoaderProcessContentSingleFile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "AttributeFilterIDs",
				Path:  "AttributeFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Path",
				Path:  "Path",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: map[string]*openedCSVFile{
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
			&engine.Attribute{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", true, utils.INFIELD_SEP),
			},
			&engine.Attribute{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", true, utils.INFIELD_SEP),
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
}

func TestLoaderProcessContentMultiFiles(t *testing.T) {
	file1CSV := `ignored,ignored,ignored,ignored,ignored,,*req.Subject,1001,ignored,ignored`
	file2CSV := `ignored,TestLoader2`
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContentMultiFiles",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaString,
				Value:     config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~File2.csv:1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("*any", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Path",
				Path:  "Path",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~File1.csv:6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~File1.csv:7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("10", true, utils.INFIELD_SEP)},
		},
	}
	rdr1 := ioutil.NopCloser(strings.NewReader(file1CSV))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	rdr2 := ioutil.NopCloser(strings.NewReader(file2CSV))
	csvRdr2 := csv.NewReader(rdr2)
	csvRdr2.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: map[string]*openedCSVFile{
			"File1.csv": &openedCSVFile{fileName: "File1.csv",
				rdr: rdr1, csvRdr: csvRdr1},
			"File2.csv": &openedCSVFile{fileName: "File2.csv",
				rdr: rdr2, csvRdr: csvRdr2}},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err != nil {
		t.Error(err)
	}
	eAP := &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "TestLoader2",
		Contexts: []string{utils.ANY},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				Path:      utils.MetaReq + utils.NestingSep + "Subject",
				FilterIDs: []string{},
				Value:     config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
			}},
		Weight: 10.0,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAP.Attributes, ap.Attributes) {
		t.Errorf("expecting: %s, \n received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}
}

func TestLoaderProcessResource(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: []*config.FCTemplate{
			&config.FCTemplate{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "TTL",
				Path:  "UsageTTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Limit",
				Path:  "Limit",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "AllocationMessage",
				Path:  "AllocationMessage",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Thresholds",
				Path:  "Thresholds",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ResourcesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: map[string]*openedCSVFile{
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
		UsageTTL:          time.Duration(1 * time.Second),
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
		UsageTTL:          time.Duration(3600 * time.Second),
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessFilters",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: []*config.FCTemplate{
			&config.FCTemplate{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Element",
				Path:  "Element",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Values",
				Path:  "Values",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.FiltersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: map[string]*openedCSVFile{
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
			&engine.FilterRule{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001", "1002"},
			},
			&engine.FilterRule{
				Type:    "*prefix",
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"10", "20"},
			},
			&engine.FilterRule{
				Type:    "*rsr",
				Element: "",
				Values:  []string{"~*req.Subject(~^1.*1$)", "~*req.Destination(1002)"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	// Compile Value for rsr fields
	if err := eFltr1.Rules[2].CompileValues(); err != nil {
		t.Error(err)
	}
	eFltr2 := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_DST_DE",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				Type:    "*destinations",
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"DST_DE"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MaxHits",
				Path:  "MaxHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinHits",
				Path:  "MinHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinSleep",
				Path:  "MinSleep",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActionIDs",
				Path:  "ActionIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Async",
				Path:  "Async",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ThresholdsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: map[string]*openedCSVFile{
			"Thresholds.csv": &openedCSVFile{fileName: "Thresholds.csv",
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
		MinSleep:  time.Duration(1 * time.Second),
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
}

func TestLoaderProcessStats(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "QueueLength",
				Path:  "QueueLength",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "TTL",
				Path:  "TTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinItems",
				Path:  "MinItems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MetricIDs",
				Path:  "MetricIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MetricFilterIDs",
				Path:  "MetricFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},

			&config.FCTemplate{Tag: "ThresholdIDs",
				Path:  "ThresholdIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.StatsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: map[string]*openedCSVFile{
			"Stats.csv": &openedCSVFile{fileName: "Stats.csv",
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
		TTL:         time.Duration(1 * time.Second),
		Metrics: []*engine.MetricWithFilters{
			&engine.MetricWithFilters{
				MetricID: "*sum:~*req.Value",
			},
			&engine.MetricWithFilters{
				MetricID: "*average:~*req.Value",
			},
			&engine.MetricWithFilters{
				MetricID: "*sum:~*req.Usage",
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
}

func TestLoaderProcessRoutes(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Sorting",
				Path:  "Sorting",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SortingParameters",
				Path:  "SortingParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteID",
				Path:  "RouteID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteFilterIDs",
				Path:  "RouteFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteAccountIDs",
				Path:  "RouteAccountIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteRatingPlanIDs",
				Path:  "RouteRatingplanIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteResourceIDs",
				Path:  "RouteResourceIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteStatIDs",
				Path:  "RouteStatIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteWeight",
				Path:  "RouteWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteBlocker",
				Path:  "RouteBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~13", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RouteParameters",
				Path:  "RouteParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~14", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~15", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.RoutesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRoutes: map[string]*openedCSVFile{
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
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
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
	eSp2 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			&engine.Route{
				ID:              "supplier1",
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
	eSp3 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
		},
		Weight: 20,
	}
	eSp4 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	eSp5 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
		},
		Weight: 20,
	}
	eSp6 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_ACNT_dan"},
				AccountIDs:      []string{"Account1", "Account1_1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param1",
			},
			&engine.Route{
				ID:              "supplier1",
				RatingPlanIDs:   []string{"RPL_2"},
				ResourceIDs:     []string{"ResGroup2", "ResGroup4"},
				StatIDs:         []string{"Stat3"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			&engine.Route{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account2"},
				RatingPlanIDs:   []string{"RPL_3"},
				ResourceIDs:     []string{"ResGroup3"},
				StatIDs:         []string{"Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}

	if aps, err := ldr.dm.GetRouteProfile("cgrates.org", "SPP_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSp, aps) &&
		!reflect.DeepEqual(eSp2, aps) &&
		!reflect.DeepEqual(eSp3, aps) &&
		!reflect.DeepEqual(eSp4, aps) &&
		!reflect.DeepEqual(eSp5, aps) &&
		!reflect.DeepEqual(eSp6, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSp), utils.ToJSON(aps))
	}
}

func TestLoaderProcessChargers(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RunID",
				Path:  "RunID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "AttributeIDs",
				Path:  "AttributeIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ChargersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: map[string]*openedCSVFile{
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

}

func TestLoaderProcessDispatches(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: []*config.FCTemplate{
			&config.FCTemplate{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			&config.FCTemplate{
				Tag:       "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			&config.FCTemplate{
				Tag:   "Subsystems",
				Path:  "Subsystems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "Strategy",
				Path:  "Strategy",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "StrategyParameters",
				Path:  "StrategyParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ConnID",
				Path:  "ConnID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ConnFilterIDs",
				Path:  "ConnFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ConnWeight",
				Path:  "ConnWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ConnBlocker",
				Path:  "ConnBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "ConnParameters",
				Path:  "ConnParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP),
			},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.DispatcherCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: map[string]*openedCSVFile{
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

}

func TestLoaderProcessDispatcheHosts(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: []*config.FCTemplate{
			&config.FCTemplate{
				Tag:       "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			&config.FCTemplate{
				Tag:       "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			&config.FCTemplate{
				Tag:   "Address",
				Path:  "Address",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "Transport",
				Path:  "Transport",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
			},
			&config.FCTemplate{
				Tag:   "TLS",
				Path:  "TLS",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
			},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.DispatcherHostCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: map[string]*openedCSVFile{
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
		ID:     "ALL1",
		Conns: []*config.RemoteHost{
			&config.RemoteHost{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
			&config.RemoteHost{
				Address:   "127.0.0.1:3012",
				Transport: utils.MetaJSON,
			},
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
}

func TestLoaderRemoveContentSingleFile(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: []*config.FCTemplate{
			&config.FCTemplate{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: map[string]*openedCSVFile{
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
			&engine.Attribute{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", true, utils.INFIELD_SEP),
			},
			&engine.Attribute{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", true, utils.INFIELD_SEP),
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
}
