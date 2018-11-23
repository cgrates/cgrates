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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLoaderProcessContentSingleFile(t *testing.T) {
	attrsCSV := `#Tenant,ID,Contexts,FilterIDs,ActivationInterval,FieldName,Initial,Substitute,Append,Weight
cgrates.org,TestLoader1,*sessions;*cdrs,*string:Account:1007,2014-01-14T00:00:00Z,Account,*any,1001,false,10
cgrates.org,TestLoader1,lcr,*string:Account:1008;*string:Account:1009,,Subject,*any,1001,true,
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: []*config.FCTemplate{
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
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(attrsCSV))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: map[string]*openedCSVFile{
			"Attributes.csv": &openedCSVFile{fileName: "Attributes.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaAttributes); err != nil {
		t.Error(err)
	}
	eAP := &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "TestLoader1",
		Contexts: []string{utils.MetaSessionS, utils.MetaCDRs, "lcr"},
		FilterIDs: []string{"*string:Account:1007",
			"*string:Account:1008", "*string:Account:1009"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "Account",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  "Subject",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     true,
			}},
		Weight: 10.0,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAP.Attributes, ap.Attributes) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}
}

func TestLoaderProcessContentMultiFiles(t *testing.T) {
	file1CSV := `ignored,ignored,ignored,ignored,ignored,Subject,*any,1001,ignored,ignored`
	file2CSV := `ignored,TestLoader2`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContentMultiFiles",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: []*config.FCTemplate{
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
	if err := ldr.processContent(utils.MetaAttributes); err != nil {
		t.Error(err)
	}
	eAP := &engine.AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "TestLoader2",
		Contexts: []string{utils.ANY},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "Subject",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     true,
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
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}
}

func TestLoaderProcessResource(t *testing.T) {
	resProfiles := `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Weight[9],Thresholds[10]
cgrates.org,ResGroup1,*string:Account:1001,2014-07-29T15:00:00Z,1s,2,call,true,true,10,
cgrates.org,ResGroup2,*string:Account:1002,2014-07-29T15:00:00Z,3600s,2,premium_call,true,true,10,
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: []*config.FCTemplate{
			&config.FCTemplate{Tag: "Tenant",
				FieldId:   "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "TTL",
				FieldId: "UsageTTL",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Limit",
				FieldId: "Limit",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "AllocationMessage",
				FieldId: "AllocationMessage",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				FieldId: "Blocker",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Stored",
				FieldId: "Stored",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Thresholds",
				FieldId: "Thresholds",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(resProfiles))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: map[string]*openedCSVFile{
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaResources); err != nil {
		t.Error(err)
	}
	eResPrf1 := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"*string:Account:1001"},
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
		ID:        "ResGroup2",
		FilterIDs: []string{"*string:Account:1002"},
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
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eResPrf1, resPrf) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eResPrf1), utils.ToJSON(resPrf))
	}
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eResPrf2, resPrf) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eResPrf2), utils.ToJSON(resPrf))
	}
}

func TestLoaderProcessFilters(t *testing.T) {
	filters := `
#Tenant[0],ID[1],FilterType[2],FilterFieldName[3],FilterFieldValues[4],ActivationInterval[5]
cgrates.org,FLTR_1,*string,Account,1001;1002,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*prefix,Destination,10;20,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*rsr,,Subject(~^1.*1$);Destination(1002),
cgrates.org,FLTR_ACNT_dan,*string,Account,dan,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_DE,*destinations,Destination,DST_DE,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_NL,*destinations,Destination,DST_NL,2014-07-29T15:00:00Z
cgrates.org,FLTR_ACNT_1001,*string,Account,1001,2014-07-29T15:00:00Z
cgrates.org,FLTR_ACNT_1002,*string,Account,1002,2014-07-29T15:00:00Z
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessFilters",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: []*config.FCTemplate{
			&config.FCTemplate{Tag: "Tenant",
				FieldId:   "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~0", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "ID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~1", true, utils.INFIELD_SEP),
				Mandatory: true},
			&config.FCTemplate{Tag: "FilterType",
				FieldId: "FilterType",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "FilterFieldName",
				FieldId: "FilterFieldName",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "FilterFieldValues",
				FieldId: "FilterFieldValues",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(filters))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaFilters: map[string]*openedCSVFile{
			"Filters.csv": &openedCSVFile{fileName: "Filters.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaFilters); err != nil {
		t.Error(err)
	}
	eFltr1 := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*engine.FilterRule{
			&engine.FilterRule{
				Type:      utils.MetaString,
				FieldName: utils.Account,
				Values:    []string{"1001", "1002"},
			},
			&engine.FilterRule{
				Type:      "*prefix",
				FieldName: utils.Destination,
				Values:    []string{"10", "20"},
			},
			&engine.FilterRule{
				Type:      "*rsr",
				FieldName: "",
				Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
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
				Type:      "*destinations",
				FieldName: utils.Destination,
				Values:    []string{"DST_DE"},
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
	thresholdCSV := `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,THD_ACNT_1001,*string:Account:1001,2014-07-29T15:00:00Z,1,1,1s,false,10,ACT_LOG_WARNING,false
cgrates.org,THD_ACNT_1002,*string:Account:1002,2014-07-29T15:00:00Z,-1,1,1s,true,10,ACT_LOG_WARNING,true
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: []*config.FCTemplate{
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
			&config.FCTemplate{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MaxHits",
				FieldId: "MaxHits",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinHits",
				FieldId: "MinHits",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinSleep",
				FieldId: "MinSleep",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				FieldId: "Blocker",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActionIDs",
				FieldId: "ActionIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Async",
				FieldId: "Async",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(thresholdCSV))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaThresholds: map[string]*openedCSVFile{
			"Thresholds.csv": &openedCSVFile{fileName: "Thresholds.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaThresholds); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eTh1 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		MaxHits:   1,
		MinHits:   1,
		MinSleep:  time.Duration(1 * time.Second),
		Blocker:   false,
		Weight:    10,
		ActionIDs: []string{"ACT_LOG_WARNING"},
		Async:     false,
	}
	eTh2 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1002",
		FilterIDs: []string{"*string:Account:1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)},
		MaxHits:   -1,
		MinHits:   1,
		MinSleep:  time.Duration(1 * time.Second),
		Blocker:   true,
		Weight:    10,
		ActionIDs: []string{"ACT_LOG_WARNING"},
		Async:     true,
	}
	if aps, err := ldr.dm.GetThresholdProfile("cgrates.org", "THD_ACNT_1001",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTh1, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eTh1), utils.ToJSON(aps))
	}
	if aps, err := ldr.dm.GetThresholdProfile("cgrates.org", "THD_ACNT_1002",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTh2, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eTh2), utils.ToJSON(aps))
	}
}

func TestLoaderProcessStats(t *testing.T) {
	statsCSV := `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],Metrics[6],MetricParams[7],Blocker[8],Stored[9],Weight[10],MinItems[11],ThresholdIDs[12]
cgrates.org,Stats1,*string:Account:1001;*string:Account:1002,2014-07-29T15:00:00Z,100,1s,*asr;*acc;*tcc;*acd;*tcd;*pdd,,true,true,20,2,THRESH1;THRESH2
cgrates.org,Stats1,*string:Account:1003,2014-07-29T15:00:00Z,100,1s,*sum;*average,Value,true,true,20,2,THRESH1;THRESH2
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: []*config.FCTemplate{
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
			&config.FCTemplate{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "QueueLength",
				FieldId: "QueueLength",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "TTL",
				FieldId: "TTL",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Metrics",
				FieldId: "Metrics",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MetricParams",
				FieldId: "Parameters",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Blocker",
				FieldId: "Blocker",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Stored",
				FieldId: "Stored",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "MinItems",
				FieldId: "MinItems",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ThresholdIDs",
				FieldId: "ThresholdIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(statsCSV))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: map[string]*openedCSVFile{
			"Stats.csv": &openedCSVFile{fileName: "Stats.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaStats); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eSt1 := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stats1",
		FilterIDs: []string{"*string:Account:1001", "*string:Account:1002", "*string:Account:1003"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         time.Duration(1 * time.Second),
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID: utils.MetaASR,
			},
			&utils.MetricWithParams{
				MetricID: utils.MetaACC,
			},
			&utils.MetricWithParams{
				MetricID: utils.MetaTCC,
			},
			&utils.MetricWithParams{
				MetricID: utils.MetaACD,
			},
			&utils.MetricWithParams{
				MetricID: utils.MetaTCD,
			},
			&utils.MetricWithParams{
				MetricID: utils.MetaPDD,
			},
			&utils.MetricWithParams{
				MetricID:   utils.MetaSum,
				Parameters: "Value",
			},
			&utils.MetricWithParams{
				MetricID:   utils.MetaAverage,
				Parameters: "Value",
			},
		},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     2,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
	}
	if aps, err := ldr.dm.GetStatQueueProfile("cgrates.org", "Stats1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSt1.Tenant, aps.Tenant) {
		t.Errorf("expecting: %s, received: %s", eSt1.Tenant, aps.Tenant)
	} else if !reflect.DeepEqual(eSt1.ID, aps.ID) {
		t.Errorf("expecting: %s, received: %s", eSt1.ID, aps.ID)
	} else if !reflect.DeepEqual(len(eSt1.FilterIDs), len(aps.FilterIDs)) {
		t.Errorf("expecting: %d, received: %d", len(eSt1.FilterIDs), len(aps.FilterIDs))
	} else if !reflect.DeepEqual(eSt1.ActivationInterval, aps.ActivationInterval) {
		t.Errorf("expecting: %s, received: %s", eSt1.ActivationInterval, aps.ActivationInterval)
	} else if !reflect.DeepEqual(eSt1.QueueLength, aps.QueueLength) {
		t.Errorf("expecting: %+v, received: %+v", eSt1.QueueLength, aps.QueueLength)
	} else if !reflect.DeepEqual(eSt1.TTL, aps.TTL) {
		t.Errorf("expecting: %+v, received: %+v", eSt1.TTL, aps.TTL)
	} else if !reflect.DeepEqual(len(eSt1.Metrics), len(aps.Metrics)) {
		t.Errorf("expecting: %d, received: %d", len(eSt1.Metrics), len(aps.Metrics))
	} else if !reflect.DeepEqual(eSt1.Blocker, aps.Blocker) {
		t.Errorf("expecting: %t, received: %t", eSt1.Blocker, aps.Blocker)
	} else if !reflect.DeepEqual(eSt1.Stored, aps.Stored) {
		t.Errorf("expecting: %t, received: %t", eSt1.Stored, aps.Stored)
	} else if !reflect.DeepEqual(eSt1.Weight, aps.Weight) {
		t.Errorf("expecting: %+v, received: %+v", eSt1.Weight, aps.Weight)
	} else if !reflect.DeepEqual(eSt1.MinItems, aps.MinItems) {
		t.Errorf("expecting: %+v, received: %+v", eSt1.MinItems, aps.MinItems)
	} else if !reflect.DeepEqual(len(eSt1.ThresholdIDs), len(aps.ThresholdIDs)) {
		t.Errorf("expecting: %d, received: %d", len(eSt1.ThresholdIDs), len(aps.ThresholdIDs))
	}
}

func TestLoaderProcessSuppliers(t *testing.T) {
	supplierCSV := `
#Tenant[0],ID[1],FilterIDs[2],ActivationInterval[3],Sorting[4],SortingParamameters[5],SupplierID[6],SupplierFilterIDs[7],SupplierAccountIDs[8],SupplierRatingPlanIDs[9],SupplierResourceIDs[10],SupplierStatIDs[11],SupplierWeight[12],SupplierBlocker[13],SupplierParameters[14],Weight[15]
cgrates.org,SPL_WEIGHT_2,,2017-11-27T00:00:00Z,*weight,,supplier1,,,,,,10,,,5
cgrates.org,SPL_WEIGHT_1,FLTR_DST_DE,2017-11-27T00:00:00Z,*weight,,supplier1,,,,,,10,,,10
cgrates.org,SPL_WEIGHT_1,FLTR_DST_DE,,,,supplier2,,,,,,20,,,
cgrates.org,SPL_WEIGHT_1,*string:Account:1007,,,,supplier3,FLTR_ACNT_dan,,,,,15,,,
cgrates.org,SPL_LEASTCOST_1,FLTR_1,2017-11-27T00:00:00Z,*least_cost,,supplier1,,,RP_SPECIAL_1002,resource_spl1,,10,false,,10
cgrates.org,SPL_LEASTCOST_1,,,,,supplier2,,,RP_RETAIL1,resource_spl2,,20,,,
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaSuppliers: []*config.FCTemplate{
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
			&config.FCTemplate{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Sorting",
				FieldId: "Sorting",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SortingParamameters",
				FieldId: "SortingParamameters",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierID",
				FieldId: "SupplierID",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierFilterIDs",
				FieldId: "SupplierFilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~7", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierAccountIDs",
				FieldId: "SupplierAccountIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~8", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierRatingPlanIDs",
				FieldId: "SupplierRatingplanIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~9", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierResourceIDs",
				FieldId: "SupplierResourceIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~10", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierStatIDs",
				FieldId: "SupplierStatIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~11", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierWeight",
				FieldId: "SupplierWeight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~12", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierBlocker",
				FieldId: "SupplierBlocker",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~13", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "SupplierParameters",
				FieldId: "SupplierParameters",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~14", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~15", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(supplierCSV))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaSuppliers: map[string]*openedCSVFile{
			"Suppliers.csv": &openedCSVFile{fileName: "Suppliers.csv",
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaSuppliers); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eSp1 := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_WEIGHT_2",
		FilterIDs: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:     "supplier1",
				Weight: 10,
			},
		},
		Weight: 5,
	}
	eSp3 := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_LEASTCOST_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:            "supplier1",
				RatingPlanIDs: []string{"RP_SPECIAL_1002"},
				ResourceIDs:   []string{"resource_spl1"},
				Blocker:       false,
				Weight:        10,
			},
			&engine.Supplier{
				ID:            "supplier2",
				RatingPlanIDs: []string{"RP_RETAIL1"},
				ResourceIDs:   []string{"resource_spl2"},
				Weight:        20,
			},
		},
		Weight: 10,
	}
	eSp3reverse := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_LEASTCOST_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           "*least_cost",
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:            "supplier2",
				RatingPlanIDs: []string{"RP_RETAIL1"},
				ResourceIDs:   []string{"resource_spl2"},
				Weight:        20,
			},
			&engine.Supplier{
				ID:            "supplier1",
				RatingPlanIDs: []string{"RP_SPECIAL_1002"},
				ResourceIDs:   []string{"resource_spl1"},
				Blocker:       false,
				Weight:        10,
			},
		},
		Weight: 10,
	}
	if aps, err := ldr.dm.GetSupplierProfile("cgrates.org", "SPL_WEIGHT_2",
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSp1, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSp1), utils.ToJSON(aps))
	}

	if aps, err := ldr.dm.GetSupplierProfile("cgrates.org", "SPL_LEASTCOST_1",
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSp3, aps) && !reflect.DeepEqual(eSp3reverse, aps) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eSp3), utils.ToJSON(aps))
	}
}

func TestLoaderProcessChargers(t *testing.T) {
	chargerCSV := `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],RunID[4],AttributeIDs[5],Weight[6]
cgrates.org,Charge1,*string:Account:1001;*string:Account:1001,2014-07-29T15:00:00Z,*rated,Attr1;Attr2,20
cgrates.org,Charge2,*string:Account:1003,2014-07-29T15:00:00Z,*default,Attr3,10
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: []*config.FCTemplate{
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
			&config.FCTemplate{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~3", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "RunID",
				FieldId: "RunID",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~4", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "AttributeIDs",
				FieldId: "AttributeIDs",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~5", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   config.NewRSRParsersMustCompile("~6", true, utils.INFIELD_SEP)},
		},
	}
	rdr := ioutil.NopCloser(strings.NewReader(chargerCSV))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaChargers: map[string]*openedCSVFile{
			utils.ChargersCsv: &openedCSVFile{fileName: utils.ChargersCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaChargers); err != nil {
		t.Error(err)
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	eCharger1 := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charge1",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"Attr1", "Attr2"},
		Weight:       20,
	}
	eCharger1Rev := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charge1",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"Attr2", "Attr1"},
		Weight:       20,
	}
	eCharger2 := &engine.ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charge2",
		FilterIDs: []string{"*string:Account:1003"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 00, 0, 0, time.UTC),
		},
		RunID:        "*default",
		AttributeIDs: []string{"Attr3"},
		Weight:       10,
	}
	if rcv, err := ldr.dm.GetChargerProfile("cgrates.org", "Charge1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCharger1, rcv) && !reflect.DeepEqual(eCharger1Rev, rcv) {
		t.Errorf("expecting: %s, received: %s", utils.ToJSON(eCharger1), utils.ToJSON(rcv))
	}
	if rcv, err := ldr.dm.GetChargerProfile("cgrates.org", "Charge2",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCharger2, rcv) {
		t.Errorf("expecting: %+v, received: %+v", eCharger2, rcv)
	}

}
