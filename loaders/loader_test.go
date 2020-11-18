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
			{Tag: "AttributeFilterIDs",
				Path:  "AttributeFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "Path",
				Path:  "Path",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
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
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.INFIELD_SEP),
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
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.INFIELD_SEP)},
			{Tag: "Value",
				Path:  "Value",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("10", utils.INFIELD_SEP)},
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
		Contexts:  []string{utils.ANY},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "Subject",
				FilterIDs: []string{},
				Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "TTL",
				Path:  "UsageTTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "Limit",
				Path:  "Limit",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "AllocationMessage",
				Path:  "AllocationMessage",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "Thresholds",
				Path:  "Thresholds",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "Type",
				Path:  "Type",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "Element",
				Path:  "Element",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "Values",
				Path:  "Values",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "MaxHits",
				Path:  "MaxHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "MinHits",
				Path:  "MinHits",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "MinSleep",
				Path:  "MinSleep",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "ActionIDs",
				Path:  "ActionIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "Async",
				Path:  "Async",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.ThresholdsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: {
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "QueueLength",
				Path:  "QueueLength",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "TTL",
				Path:  "TTL",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "MinItems",
				Path:  "MinItems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "MetricIDs",
				Path:  "MetricIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "MetricFilterIDs",
				Path:  "MetricFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "Blocker",
				Path:  "Blocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "Stored",
				Path:  "Stored",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP)},

			{Tag: "ThresholdIDs",
				Path:  "ThresholdIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(engine.StatsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "Sorting",
				Path:  "Sorting",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "SortingParameters",
				Path:  "SortingParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "RouteID",
				Path:  "RouteID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "RouteFilterIDs",
				Path:  "RouteFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "RouteAccountIDs",
				Path:  "RouteAccountIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "RouteRatingPlanIDs",
				Path:  "RouteRatingplanIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "RouteRateProfileIDs",
				Path:  "RouteRateProfileIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
			{Tag: "RouteResourceIDs",
				Path:  "RouteResourceIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP)},
			{Tag: "RouteStatIDs",
				Path:  "RouteStatIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP)},
			{Tag: "RouteWeight",
				Path:  "RouteWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP)},
			{Tag: "RouteBlocker",
				Path:  "RouteBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.INFIELD_SEP)},
			{Tag: "RouteParameters",
				Path:  "RouteParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.16", utils.INFIELD_SEP)},
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
				RateProfileIDs:  []string{"RT_ALWAYS"},
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
		return strings.Compare(eSp.Routes[i].ID+strings.Join(eSp.Routes[i].FilterIDs, utils.CONCATENATED_KEY_SEP),
			eSp.Routes[j].ID+strings.Join(eSp.Routes[j].FilterIDs, utils.CONCATENATED_KEY_SEP)) < 0
	})

	aps, err := ldr.dm.GetRouteProfile("cgrates.org", "RoutePrf1",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(aps.Routes, func(i, j int) bool {
		return strings.Compare(aps.Routes[i].ID+strings.Join(aps.Routes[i].FilterIDs, utils.CONCATENATED_KEY_SEP),
			aps.Routes[j].ID+strings.Join(aps.Routes[j].FilterIDs, utils.CONCATENATED_KEY_SEP)) < 0
	})
	if !reflect.DeepEqual(eSp, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSp), utils.ToJSON(aps))
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "RunID",
				Path:  "RunID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "AttributeIDs",
				Path:  "AttributeIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:   "Subsystems",
				Path:  "Subsystems",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
			},
			{
				Tag:   "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
			},
			{
				Tag:   "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
			},
			{
				Tag:   "Strategy",
				Path:  "Strategy",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP),
			},
			{
				Tag:   "StrategyParameters",
				Path:  "StrategyParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnID",
				Path:  "ConnID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnFilterIDs",
				Path:  "ConnFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnWeight",
				Path:  "ConnWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnBlocker",
				Path:  "ConnBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP),
			},
			{
				Tag:   "ConnParameters",
				Path:  "ConnParameters",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP),
			},
			{
				Tag:   "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP),
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:   "Address",
				Path:  "Address",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP),
			},
			{
				Tag:   "Transport",
				Path:  "Transport",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP),
			},
			{
				Tag:   "TLS",
				Path:  "TLS",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP),
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
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
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub2", utils.INFIELD_SEP),
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "ConnectFee",
				Path:  "ConnectFee",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "RoundingMethod",
				Path:  "RoundingMethod",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "RoundingDecimals",
				Path:  "RoundingDecimals",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "MinCost",
				Path:  "MinCost",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "MaxCost",
				Path:  "MaxCost",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "MaxCostStrategy",
				Path:  "MaxCostStrategy",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
			{Tag: "RateID",
				Path:  "RateID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP)},
			{Tag: "RateFilterIDs",
				Path:  "RateFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP)},
			{Tag: "RateActivationStart",
				Path:  "RateActivationStart",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP)},
			{Tag: "RateWeight",
				Path:  "RateWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.INFIELD_SEP)},
			{Tag: "RateBlocker",
				Path:  "RateBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.INFIELD_SEP)},
			{Tag: "RateIntervalStart",
				Path:  "RateIntervalStart",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.16", utils.INFIELD_SEP)},
			{Tag: "RateValue",
				Path:  "RateValue",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.17", utils.INFIELD_SEP)},
			{Tag: "RateUnit",
				Path:  "RateUnit",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.18", utils.INFIELD_SEP)},
			{Tag: "RateIncrement",
				Path:  "RateIncrement",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.19", utils.INFIELD_SEP)},
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
	eRatePrf := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf), utils.ToJSON(rcv))
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "FilterIDs",
				Path:  "FilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
			{Tag: "ActivationInterval",
				Path:  "ActivationInterval",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.INFIELD_SEP)},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.INFIELD_SEP)},
			{Tag: "ConnectFee",
				Path:  "ConnectFee",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.INFIELD_SEP)},
			{Tag: "RoundingMethod",
				Path:  "RoundingMethod",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.INFIELD_SEP)},
			{Tag: "RoundingDecimals",
				Path:  "RoundingDecimals",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.INFIELD_SEP)},
			{Tag: "MinCost",
				Path:  "MinCost",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.INFIELD_SEP)},
			{Tag: "MaxCost",
				Path:  "MaxCost",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.INFIELD_SEP)},
			{Tag: "MaxCostStrategy",
				Path:  "MaxCostStrategy",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.INFIELD_SEP)},
			{Tag: "RateID",
				Path:  "RateID",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.INFIELD_SEP)},
			{Tag: "RateFilterIDs",
				Path:  "RateFilterIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.INFIELD_SEP)},
			{Tag: "RateActivationStart",
				Path:  "RateActivationStart",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.INFIELD_SEP)},
			{Tag: "RateWeight",
				Path:  "RateWeight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.INFIELD_SEP)},
			{Tag: "RateBlocker",
				Path:  "RateBlocker",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.INFIELD_SEP)},
			{Tag: "RateIntervalStart",
				Path:  "RateIntervalStart",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.16", utils.INFIELD_SEP)},
			{Tag: "RateValue",
				Path:  "RateValue",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.17", utils.INFIELD_SEP)},
			{Tag: "RateUnit",
				Path:  "RateUnit",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.18", utils.INFIELD_SEP)},
			{Tag: "RateIncrement",
				Path:  "RateIncrement",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.19", utils.INFIELD_SEP)},
		},
	}
	ratePrfCnt1 := `
#Tenant,ID,FilterIDs,ActivationInterval,Weight,ConnectFee,RoundingMethod,RoundingDecimals,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeight,RateBlocker,RateIntervalStart,RateValue,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,,0,0.1,*up,4,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",0,false,0s,0.12,1m,1m
cgrates.org,RP1,,,,,,,,,,RT_WEEK,,,,,1m,0.06,1m,1s
`
	ratePrfCnt2 := `
#Tenant,ID,FilterIDs,ActivationInterval,Weight,ConnectFee,RoundingMethod,RoundingDecimals,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeight,RateBlocker,RateIntervalStart,RateValue,RateUnit,RateIncrement
cgrates.org,RP1,,,,,,,,,,RT_WEEKEND,,"* * * * 0,6",10,false,0s,0.06,1m,1s
cgrates.org,RP1,,,,,,,,,,RT_CHRISTMAS,,* * 24 12 *,30,false,0s,0.06,1m,1s
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
	eRatePrf := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf) {
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
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf) {
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
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "RateIDs",
				Path:  "RateIDs",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
		},
	}

	rPfr := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if err := ldr.dm.SetRateProfile(rPfr, true); err != nil {
		t.Error(err)
	}
	rPfr2 := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP2",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
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
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf) {
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
		Tenant:           "cgrates.org",
		ID:               "RP2",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf2) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf2), utils.ToJSON(rcv))
	}

	eRatePrf3 := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates:            map[string]*engine.Rate{},
	}
	if rcv, err := ldr.dm.GetRateProfile("cgrates.org", "RP1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRatePrf3) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eRatePrf3), utils.ToJSON(rcv))
	}
}

func TestNewLoaderWithMultiFiles(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)

	ldrCfg := config.CgrConfig().LoaderCfg()[0].Clone()
	ldrCfg.Data[0].Fields = []*config.FCTemplate{
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
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.INFIELD_SEP)},
		{Tag: "Value",
			Path:  "Value",
			Type:  utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.INFIELD_SEP)},
		{Tag: "Weight",
			Path:  "Weight",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("10", utils.INFIELD_SEP)},
	}
	ldr := NewLoader(engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil), ldrCfg, "", nil, nil, nil)

	openRdrs := make(utils.StringSet)
	for _, rdr := range ldr.rdrs {
		for fileName := range rdr {
			openRdrs.Add(fileName)
		}
	}
	expected := utils.StringSet{
		"Attributes.csv":         {},
		"Chargers.csv":           {},
		"DispatcherHosts.csv":    {},
		"DispatcherProfiles.csv": {},
		"File1.csv":              {},
		"File2.csv":              {},
		"Filters.csv":            {},
		"RateProfiles.csv":       {},
		"Resources.csv":          {},
		"Routes.csv":             {},
		"Stats.csv":              {},
		"Thresholds.csv":         {},
	}
	if !reflect.DeepEqual(expected, openRdrs) {
		t.Errorf("Expected %s,received %s", utils.ToJSON(expected), utils.ToJSON(openRdrs))
	}
}
