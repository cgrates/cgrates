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

package loaders

import (
	"encoding/csv"
	"io"
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
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "AttributeFilterIDs",
				Path:  "AttributeFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "Path",
				Path:  "Path",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": {fileName: "Attributes.csv",
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
				Value:     config.NewRSRParsersMustCompile("Sub1", true, utils.INFIELD_SEP),
			},
			{
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
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaString,
				Value:     config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~File2.csv:1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "Contexts",
				Path:  "Contexts",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("*any", true, utils.INFIELD_SEP)},
			{Tag: "Path",
				Path:  "Path",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~File1.csv:6", true, utils.INFIELD_SEP)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~File1.csv:7", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("10", true, utils.INFIELD_SEP)},
		},
	}
	rdr1 := io.NopCloser(strings.NewReader(file1CSV))
	csvRdr1 := csv.NewReader(rdr1)
	csvRdr1.Comment = '#'
	rdr2 := io.NopCloser(strings.NewReader(file2CSV))
	csvRdr2 := csv.NewReader(rdr2)
	csvRdr2.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"File1.csv": {fileName: "File1.csv",
				rdr: rdr1, csvRdr: csvRdr1},
			"File2.csv": {fileName: "File2.csv",
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
			{
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
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "TTL",
				Path:  "UsageTTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "Limit",
				Path:  "Limit",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "AllocationMessage",
				Path:  "AllocationMessage",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			{Tag: "Thresholds",
				Path:  "Thresholds",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.ResourcesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": {fileName: "Resources.csv",
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
		utils.MetaFilters: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "Element",
				Path:  "Element",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "Values",
				Path:  "Values",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.FiltersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: {
			"Filters.csv": {fileName: "Filters.csv",
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Values:  []string{"1001", "1002"},
			},
			{
				Type:    "*prefix",
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Values:  []string{"10", "20"},
			},
			{
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

	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if fltr, err := engine.GetFilter(ldr.dm, "cgrates.org", "FLTR_1",
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFltr1, fltr) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eFltr1), utils.ToJSON(fltr))
	}
	if fltr, err := engine.GetFilter(ldr.dm, "cgrates.org", "FLTR_DST_DE",
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
		utils.MetaThresholds: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "MaxHits",
				Path:  "MaxHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "MinHits",
				Path:  "MinHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "MinSleep",
				Path:  "MinSleep",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			{Tag: "ActionIDs",
				Path:  "ActionIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			{Tag: "Async",
				Path:  "Async",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.ThresholdsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
			"Thresholds.csv": {fileName: "Thresholds.csv",
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
		utils.MetaStats: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "QueueLength",
				Path:  "QueueLength",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "TTL",
				Path:  "TTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "MinItems",
				Path:  "MinItems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			{Tag: "MetricIDs",
				Path:  "MetricIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			{Tag: "MetricFilterIDs",
				Path:  "MetricFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},

			{Tag: "ThresholdIDs",
				Path:  "ThresholdIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.StatsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			"Stats.csv": {fileName: "Stats.csv",
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
			{
				MetricID: "*sum:~Value",
			},
			{
				MetricID: "*average:~Value",
			},
			{
				MetricID: "*sum:~Usage",
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

func TestLoaderProcessSuppliers(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaSuppliers: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "Sorting",
				Path:  "Sorting",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "SortingParamameters",
				Path:  "SortingParamameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "SupplierID",
				Path:  "SupplierID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			{Tag: "SupplierFilterIDs",
				Path:  "SupplierFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			{Tag: "SupplierAccountIDs",
				Path:  "SupplierAccountIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			{Tag: "SupplierRatingPlanIDs",
				Path:  "SupplierRatingplanIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			{Tag: "SupplierResourceIDs",
				Path:  "SupplierResourceIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			{Tag: "SupplierStatIDs",
				Path:  "SupplierStatIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},
			{Tag: "SupplierWeight",
				Path:  "SupplierWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
			{Tag: "SupplierBlocker",
				Path:  "SupplierBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~13", true, utils.INFIELD_SEP)},
			{Tag: "SupplierParameters",
				Path:  "SupplierParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~14", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~15", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.SuppliersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaSuppliers: {
			"Suppliers.csv": {fileName: "Suppliers.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaSuppliers, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}

	eSp3 := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPP_1",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			{
				ID:                 "supplier1",
				FilterIDs:          []string{"FLTR_DST_DE"},
				AccountIDs:         []string{"Account2"},
				RatingPlanIDs:      []string{"RPL_3"},
				ResourceIDs:        []string{"ResGroup3"},
				StatIDs:            []string{"Stat2"},
				Weight:             10,
				Blocker:            false,
				SupplierParameters: utils.EmptyString,
			},
			{
				ID:                 "supplier1",
				FilterIDs:          []string{"FLTR_ACNT_dan"},
				AccountIDs:         []string{"Account1", "Account1_1"},
				RatingPlanIDs:      []string{"RPL_1"},
				ResourceIDs:        []string{"ResGroup1"},
				StatIDs:            []string{"Stat1"},
				Weight:             10,
				Blocker:            true,
				SupplierParameters: "param1",
			},
			{
				ID:                 "supplier1",
				RatingPlanIDs:      []string{"RPL_2"},
				ResourceIDs:        []string{"ResGroup2", "ResGroup4"},
				StatIDs:            []string{"Stat3"},
				Weight:             10,
				Blocker:            false,
				SupplierParameters: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	sort.Slice(eSp3.Suppliers, func(i, j int) bool {
		return strings.Compare(eSp3.Suppliers[i].ID+
			strings.Join(eSp3.Suppliers[i].FilterIDs, utils.CONCATENATED_KEY_SEP),
			eSp3.Suppliers[j].ID+strings.Join(eSp3.Suppliers[j].FilterIDs, utils.CONCATENATED_KEY_SEP)) < 0
	})
	if aps, err := ldr.dm.GetSupplierProfile("cgrates.org", "SPP_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else {
		sort.Slice(aps.Suppliers, func(i, j int) bool {
			return strings.Compare(aps.Suppliers[i].ID+
				strings.Join(aps.Suppliers[i].FilterIDs, utils.CONCATENATED_KEY_SEP),
				aps.Suppliers[j].ID+strings.Join(aps.Suppliers[j].FilterIDs, utils.CONCATENATED_KEY_SEP)) < 0
		})
		if !reflect.DeepEqual(eSp3, aps) {
			t.Errorf("expecting: %s, received: %s",
				utils.ToJSON(eSp3), utils.ToJSON(aps))
		}
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
		utils.MetaChargers: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			{Tag: "RunID",
				Path:  "RunID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			{Tag: "AttributeIDs",
				Path:  "AttributeIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.ChargersCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: {
			utils.ChargersCsv: {fileName: utils.ChargersCsv,
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
		utils.MetaDispatchers: {
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:   "Subsystems",
				Path:  "Subsystems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "Strategy",
				Path:  "Strategy",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "StrategyParameters",
				Path:  "StrategyParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnID",
				Path:  "ConnID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnFilterIDs",
				Path:  "ConnFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnWeight",
				Path:  "ConnWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnBlocker",
				Path:  "ConnBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnParameters",
				Path:  "ConnParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP),
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.DispatcherCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatchers: {
			utils.DispatcherProfilesCsv: {
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
		StrategyParams: map[string]any{},
		Strategy:       "*first",
		Weight:         20,
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{"*gt:~*req.Usage:10"},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.56.203"},
				Blocker:   false,
			},
			&engine.DispatcherHostProfile{
				ID:        "C2",
				FilterIDs: []string{"*lt:~*req.Usage:10"},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.56.204"},
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
		utils.MetaDispatcherHosts: {
			{
				Tag:       "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:   "Address",
				Path:  "Address",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "Transport",
				Path:  "Transport",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP),
			},
			{
				Tag:   "TLS",
				Path:  "TLS",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP),
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.DispatcherHostCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaDispatcherHosts: {
			utils.DispatcherProfilesCsv: {
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
			{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
			{
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
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
		},
	}
	rdr := io.NopCloser(strings.NewReader(engine.AttributesCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			"Attributes.csv": {fileName: "Attributes.csv",
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
				Value:     config.NewRSRParsersMustCompile("Sub1", true, utils.INFIELD_SEP),
			},
			{
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
