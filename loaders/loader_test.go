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
	ldr.dataTpls = map[string][]*config.CfgCdrField{
		utils.MetaAttributes: []*config.CfgCdrField{
			&config.CfgCdrField{Tag: "TenantID",
				FieldId:   "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("0", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "ProfileID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "Contexts",
				FieldId: "Contexts",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "FieldName",
				FieldId: "FieldName",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Initial",
				FieldId: "Initial",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Substitute",
				FieldId: "Substitute",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Append",
				FieldId: "Append",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP)},
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
				Substitute: "1001",
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  "Subject",
				Initial:    utils.ANY,
				Substitute: "1001",
				Append:     true,
			}},
		Weight: 10.0,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader1",
		false, utils.NonTransactional); err != nil {
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
	ldr.dataTpls = map[string][]*config.CfgCdrField{
		utils.MetaAttributes: []*config.CfgCdrField{
			&config.CfgCdrField{Tag: "TenantID",
				FieldId:   "Tenant",
				Type:      utils.MetaString,
				Value:     utils.ParseRSRFieldsMustCompile("^cgrates.org", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "ProfileID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("File2.csv:1", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "Contexts",
				FieldId: "Contexts",
				Type:    utils.MetaString,
				Value:   utils.ParseRSRFieldsMustCompile("^*any", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "FieldName",
				FieldId: "FieldName",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("File1.csv:5", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Initial",
				FieldId: "Initial",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("File1.csv:6", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Substitute",
				FieldId: "Substitute",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("File1.csv:7", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Append",
				FieldId: "Append",
				Type:    utils.MetaString,
				Value:   utils.ParseRSRFieldsMustCompile("^true", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.MetaString,
				Value:   utils.ParseRSRFieldsMustCompile("^10", utils.INFIELD_SEP)},
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
				Substitute: "1001",
				Append:     true,
			}},
		Weight: 10.0,
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader2",
		false, utils.NonTransactional); err != nil {
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
	ldr.dataTpls = map[string][]*config.CfgCdrField{
		utils.MetaResources: []*config.CfgCdrField{
			&config.CfgCdrField{Tag: "Tenant",
				FieldId:   "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("0", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "ID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "FilterIDs",
				FieldId: "FilterIDs",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "TTL",
				FieldId: "UsageTTL",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Limit",
				FieldId: "Limit",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "AllocationMessage",
				FieldId: "AllocationMessage",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Blocker",
				FieldId: "Blocker",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Stored",
				FieldId: "Stored",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Weight",
				FieldId: "Weight",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "Thresholds",
				FieldId: "Thresholds",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP)},
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
	}
	if len(ldr.bufLoaderData) != 0 {
		t.Errorf("wrong buffer content: %+v", ldr.bufLoaderData)
	}
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eResPrf1, resPrf) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eResPrf1), utils.ToJSON(resPrf))
	}
	if resPrf, err := ldr.dm.GetResourceProfile("cgrates.org", "ResGroup2",
		false, utils.NonTransactional); err != nil {
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
`
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID:         "TestLoaderProcessFilters",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.CfgCdrField{
		utils.MetaFilters: []*config.CfgCdrField{
			&config.CfgCdrField{Tag: "Tenant",
				FieldId:   "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("0", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "ID",
				FieldId:   "ID",
				Type:      utils.META_COMPOSED,
				Value:     utils.ParseRSRFieldsMustCompile("1", utils.INFIELD_SEP),
				Mandatory: true},
			&config.CfgCdrField{Tag: "FilterType",
				FieldId: "FilterType",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "FilterFieldName",
				FieldId: "FilterFieldName",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "FilterFieldValues",
				FieldId: "FilterFieldValues",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP)},
			&config.CfgCdrField{Tag: "ActivationInterval",
				FieldId: "ActivationInterval",
				Type:    utils.META_COMPOSED,
				Value:   utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP)},
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
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFltr1, fltr) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eFltr1), utils.ToJSON(fltr))
	}
	if fltr, err := ldr.dm.GetFilter("cgrates.org", "FLTR_DST_DE",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFltr2, fltr) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eFltr2), utils.ToJSON(fltr))
	}
}
