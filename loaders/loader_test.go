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
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var data = engine.NewInternalDB(nil, nil, false, config.CgrConfig().DataDbCfg().Items)

const (
	ResourcesCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Weight[9],Thresholds[10]
cgrates.org,ResGroup21,*string:~*req.Account:1001,2014-07-29T15:00:00Z,1s,2,call,true,true,10,
cgrates.org,ResGroup22,*string:~*req.Account:dan,2014-07-29T15:00:00Z,3600s,2,premium_call,true,true,10,
`
	StatsCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,TestStats,*string:~*req.Account:1001,2014-07-29T15:00:00Z,100,1s,2,*sum#~*req.Value;*average#~*req.Value,,true,true,20,Th1;Th2
cgrates.org,TestStats,,,,,2,*sum#~*req.Usage,,true,true,20,
cgrates.org,TestStats2,FLTR_1,2014-07-29T15:00:00Z,100,1s,2,*sum#~*req.Value;*sum#~*req.Usage;*average#~*req.Value;*average#~*req.Usage,,true,true,20,Th
cgrates.org,TestStats2,,,,,2,*sum#~*req.Cost;*average#~*req.Cost,,true,true,20,
`
	ThresholdsCSVContent = `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,Threshold1,*string:~*req.Account:1001;*string:~*req.RunID:*default,2014-07-29T15:00:00Z,12,10,1s,true,10,THRESH1,true
`
	FiltersCSVContent = `
#Tenant[0],ID[1],Type[2],Element[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_1,*string,~*req.Account,1001;1002,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*prefix,~*req.Destination,10;20,2014-07-29T15:00:00Z
cgrates.org,FLTR_1,*rsr,~*req.Subject,~^1.*1$,
cgrates.org,FLTR_1,*rsr,~*req.Destination,1002,
cgrates.org,FLTR_ACNT_dan,*string,~*req.Account,dan,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_DE,*destinations,~*req.Destination,DST_DE,2014-07-29T15:00:00Z
cgrates.org,FLTR_DST_NL,*destinations,~*req.Destination,DST_NL,2014-07-29T15:00:00Z
`
	RoutesCSVContent = `
#Tenant[0],ID[1],FilterIDs[2],ActivationInterval[3],Sorting[4],SortingParameters[5],RouteID[6],RouteFilterIDs[7],RouteAccountIDs[8],RouteRatingPlanIDs[9],RouteResourceIDs[10],RouteStatIDs[11],RouteWeight[12],RouteBlocker[13],RouteParameters[14],Weight[15]
cgrates.org,RoutePrf1,*string:~*req.Account:dan,2014-07-29T15:00:00Z,*lc,,route1,FLTR_ACNT_dan,Account1;Account1_1,RPL_1,ResGroup1,Stat1,10,true,param1,20
cgrates.org,RoutePrf1,,,,,route1,,,RPL_2,ResGroup2,,10,,,
cgrates.org,RoutePrf1,,,,,route1,FLTR_DST_DE,Account2,RPL_3,ResGroup3,Stat2,10,,,
cgrates.org,RoutePrf1,,,,,route1,,,,ResGroup4,Stat3,10,,,
`
	AttributesCSVContent = `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ALS1,con1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*string:~*req.Field1:Initial,*req.Field1,*variable,Sub1,true,20
cgrates.org,ALS1,con2;con3,,,,*req.Field2,*variable,Sub2,true,20
`
	ChargersCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,Charger1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*rated,ATTR_1001_SIMPLEAUTH,20
`
	DispatcherCSVContent = `
#Tenant,ID,FilterIDs,ActivationInterval,Strategy,Hosts,Weight
cgrates.org,D1,*any,*string:~*req.Account:1001,2014-07-29T15:00:00Z,*first,,C1,*gt:~*req.Usage:10,10,false,192.168.56.203,20
cgrates.org,D1,,,,*first,,C2,*lt:~*req.Usage:10,10,false,192.168.56.204,
`
	DispatcherHostCSVContent = `
#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],MaxReconnectInterval[6],ConnectTimeout[7],ReplyTimeout[8],Tls[9],ClientKey[10],ClientCertificate[11],CaCertificate[12]
cgrates.org,ALL,127.0.0.1:6012,*json,1,3,5m,1m,2m,false,,,
`
)

func TestLoaderProcessContentSingleFile(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		bufLoaderData: make(map[string][]LoaderData),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "Contexts",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "AttributeFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "Path",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Path: "Type",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Path: "Value",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Path: "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(AttributesCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(AttributesCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(AttributesCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{fileName: utils.AttributesCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expectedErr := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaAttributes, utils.EmptyString); err == nil || err != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoaderProcessContentMultiFiles(t *testing.T) {
	file1CSV := `ignored,ignored,ignored,ignored,ignored,,*req.Subject,1001,ignored,ignored`
	file2CSV := `ignored,TestLoader2`
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContentMultiFiles",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Path: "Tenant",
				Type:      utils.MetaString,
				Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.InfieldSep),
				Mandatory: true},
			{Path: "Contexts",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("*any", utils.InfieldSep)},
			{Path: "Path",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.InfieldSep)},
			{Path: "Value",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaString,
				Value: config.NewRSRParsersMustCompile("10", utils.InfieldSep)},
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "UsageTTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "Limit",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "AllocationMessage",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Path: "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Path: "Stored",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Path: "Thresholds",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(ResourcesCSVContent))
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessFilters",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "Type",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "Element",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "Values",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(FiltersCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(FiltersCSVContent))
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "MaxHits",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "MinHits",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "MinSleep",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Path: "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Path: "ActionIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Path: "Async",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(ThresholdsCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(ThresholdsCSVContent))
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "QueueLength",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "TTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "MinItems",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Path: "MetricIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Path: "MetricFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Path: "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Path: "Stored",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},

			{Path: "ThresholdIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(StatsCSVContent))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {utils.StatsCsv: &openedCSVFile{rdr: rdr, csvRdr: csvRdr}},
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
			{MetricID: "*sum#~*req.Value"},
			{MetricID: "*average#~*req.Value"},
			{MetricID: "*sum#~*req.Usage"},
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
	rdr = io.NopCloser(strings.NewReader(StatsCSVContent))
	csvRdr = csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {utils.StatsCsv: &openedCSVFile{rdr: rdr, csvRdr: csvRdr}},
	}
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestLoaderProcessStatsWrongMetrics(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderProcessStatsWrongMetrics",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
		dataTpls: map[string][]*config.FCTemplate{
			utils.MetaStats: {
				{Path: "MetricIDs",
					Type:  utils.MetaComposed,
					Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
				{Path: "Stored",
					Type:  utils.MetaComposed,
					Value: config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep)},
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(`#Metrics[0],Stored[1]
not_a_valid_metric_type,true,`))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaStats: {
			utils.StatsCsv: &openedCSVFile{fileName: utils.StatsCsv,
				rdr: rdr, csvRdr: csvRdr}},
	}
	expected := "unsupported metric type <not_a_valid_metric_type>"
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err != nil {
		t.Error(err)
	}

	//initialize again but with a valid metric and false stored field
	rdr = io.NopCloser(strings.NewReader(`#Metrics[0],Stored[1]
*sum#~*req.Value,false`))
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "Sorting",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "SortingParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "RouteID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
			{Path: "RouteFilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep)},
			{Path: "RouteAccountIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep)},
			{Path: "RouteRatingplanIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep)},
			{Path: "RouteResourceIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep)},
			{Path: "RouteStatIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep)},
			{Path: "RouteWeight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep)},
			{Path: "RouteBlocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep)},
			{Path: "RouteParameters",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(RoutesCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(RoutesCSVContent))
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

func TestLoaderProcessChargers(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
			{Path: "FilterIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep)},
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep)},
			{Path: "RunID",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep)},
			{Path: "AttributeIDs",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep)},
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep)},
		},
	}
	rdr := io.NopCloser(strings.NewReader(ChargersCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(ChargersCSVContent))
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
	rdr := io.NopCloser(strings.NewReader(DispatcherCSVContent))
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

	//cannot set DispatchersProfile when dryrun is true
	ldr.dryRun = true
	rdr = io.NopCloser(strings.NewReader(DispatcherCSVContent))
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
				Tag:       "ConnectAttempts",
				Path:      "ConnectAttempts",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "Reconnects",
				Path:      "Reconnects",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "MaxReconnectInterval",
				Path:      "MaxReconnectInterval",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ConnectTimeout",
				Path:      "ConnectTimeout",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ReplyTimeout",
				Path:      "ReplyTimeout",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "TLS",
				Path:      "TLS",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ClientKey",
				Path:      "ClientKey",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "ClientCertificate",
				Path:      "ClientCertificate",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
				Mandatory: true,
			},
			{
				Tag:       "CaCertificate",
				Path:      "CaCertificate",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
				Mandatory: true,
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(DispatcherHostCSVContent))
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
			ID:                   "ALL",
			Address:              "127.0.0.1:6012",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           3,
			MaxReconnectInterval: 5 * time.Minute,
			ConnectTimeout:       1 * time.Minute,
			ReplyTimeout:         2 * time.Minute,
			TLS:                  false,
		},
	}

	rcv, err := ldr.dm.GetDispatcherHost("cgrates.org", "ALL",
		true, false, utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eDispHost, rcv) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eDispHost), utils.ToJSON(rcv))
	}

	//cannot set DispatcherHostProfile when dryrun is true
	ldr.dryRun = true
	rdr = io.NopCloser(strings.NewReader(DispatcherHostCSVContent))
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
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	rdr := io.NopCloser(strings.NewReader(AttributesCSVContent))
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
	rdr = io.NopCloser(strings.NewReader(AttributesCSVContent))
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
func TestNewLoaderWithMultiFiles(t *testing.T) {

	ldrCfg := config.CgrConfig().LoaderCfg()[0].Clone()
	ldrCfg.Data[0].Fields = []*config.FCTemplate{
		{Path: "Tenant",
			Type:      utils.MetaString,
			Value:     config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			Mandatory: true},
		{Path: "ID",
			Type:      utils.MetaComposed,
			Value:     config.NewRSRParsersMustCompile("~*file(File2.csv).1", utils.InfieldSep),
			Mandatory: true},
		{Path: "Contexts",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("*any", utils.InfieldSep)},
		{Path: "Path",
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).6", utils.InfieldSep)},
		{Path: "Value",
			Type:  utils.MetaComposed,
			Value: config.NewRSRParsersMustCompile("~*file(File1.csv).7", utils.InfieldSep)},
		{Path: "Weight",
			Type:  utils.MetaString,
			Value: config.NewRSRParsersMustCompile("10", utils.InfieldSep)},
	}
	ldr := NewLoader(engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil), ldrCfg, "", 0, nil, nil, nil)

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
		utils.ResourcesCsv:          {},
		utils.RoutesCsv:             {},
		utils.StatsCsv:              {},
		utils.ThresholdsCsv:         {},
	}
	if !reflect.DeepEqual(expected, openRdrs) {
		t.Errorf("Expected %s,received %s", utils.ToJSON(expected), utils.ToJSON(openRdrs))
	}
}

func TestLoaderAttributesAsStructErrType(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderAttributesAsStructErrType",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Path: "Weight",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	attributeCsv := `
#Weight
true
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoaderAttributesAsStructErrConversion",
		bufLoaderData: map[string][]LoaderData{},
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	attributeCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadResourcesAsStructErr",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Path: "Blocker",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	resourcesCsv := `
#Blocker
NOT_A_BOOLEAN
`
	rdr := io.NopCloser(strings.NewReader(resourcesCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadResourcesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Path: "UsageTTL",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	resourcesCsv := `
#UsageTTL
12ss
`
	rdr := io.NopCloser(strings.NewReader(resourcesCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadFiltersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	filtersCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(filtersCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadFiltersAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	filtersCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(filtersCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadStatsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	statsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(statsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadStatsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	statsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(statsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadThresholdsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadThresholdsAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadRoutesAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadRoutesAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadChargersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadChargersAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadDispatchersAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadDispatcherHostsAsStructErrConversion",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Path: "ActivationInterval",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#ActivationInterval
* * * * * *
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	ldr := &Loader{
		ldrID:         "TestLoadDispatcherHostsAsStructErrType",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{Path: "PK",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep)},
		},
	}
	thresholdsCsv := `
#PK
NOT_UINT
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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

func TestLoadAndRemoveResources(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoadAndRemoveResources",
		bufLoaderData: make(map[string][]LoaderData),
		dryRun:        true,
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	resourcesCSV := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := io.NopCloser(strings.NewReader(resourcesCSV))
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
	rdr = io.NopCloser(strings.NewReader(resourcesCSV))
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
	rdr = io.NopCloser(strings.NewReader(resourcesCSV))
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
	rdr = io.NopCloser(strings.NewReader(resourcesCSV))
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
	rdr = io.NopCloser(strings.NewReader(resourcesCSV))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveFilterContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveFilterContents",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	filtersCsv := `
#Tenant[0],ID[0]
cgrates.org,FILTERS_REM_1
`
	rdr := io.NopCloser(strings.NewReader(filtersCsv))
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
	rdr = io.NopCloser(strings.NewReader(filtersCsv))
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
	rdr = io.NopCloser(strings.NewReader(filtersCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaFilters, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveStatsContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoaderProcessContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaStats: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	statsCsv := `
#Tenant[0],ProfileID[1]
cgrates.org,REM_STATS_1
`
	rdr := io.NopCloser(strings.NewReader(statsCsv))
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
	rdr = io.NopCloser(strings.NewReader(statsCsv))
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
	rdr = io.NopCloser(strings.NewReader(statsCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveThresholdsContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveThresholdsContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaThresholds: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	thresholdsCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_THRESHOLDS_1,
`
	rdr := io.NopCloser(strings.NewReader(thresholdsCsv))
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
	rdr = io.NopCloser(strings.NewReader(thresholdsCsv))
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
	rdr = io.NopCloser(strings.NewReader(thresholdsCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveRoutesContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveRoutesContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRoutes: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	routesCsv := `
#Tenant[0],ID[1]
cgrates.org,ROUTES_REM_1
`
	rdr := io.NopCloser(strings.NewReader(routesCsv))
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
	rdr = io.NopCloser(strings.NewReader(routesCsv))
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
	rdr = io.NopCloser(strings.NewReader(routesCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaRoutes, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveChargersContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveChargersContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaChargers: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	routesCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_ROUTES_1
`
	rdr := io.NopCloser(strings.NewReader(routesCsv))
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
	rdr = io.NopCloser(strings.NewReader(routesCsv))
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
	rdr = io.NopCloser(strings.NewReader(routesCsv))
	rdr = io.NopCloser(strings.NewReader(routesCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaChargers, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveDispatchersContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveDispatchersContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatchers: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	dispatchersCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_DISPATCHERS_1
`
	rdr := io.NopCloser(strings.NewReader(dispatchersCsv))
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
	if err := ldr.removeContent(utils.MetaDispatchers, utils.EmptyString); err != utils.ErrDSPProfileNotFound {
		t.Error(err)
	}

	//cannot remove DispatchersProfile when dryrun is true
	ldr.dryRun = true
	rdr = io.NopCloser(strings.NewReader(dispatchersCsv))
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
	rdr = io.NopCloser(strings.NewReader(dispatchersCsv))
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
	expected := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaDispatchers, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveDispatcherHostsContent(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveDispatcherHostsContent",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaDispatcherHosts: {
			{Path: "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Path: "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	dispatchersHostsCsv := `
#Tenant[0],ID[1]
cgrates.org,REM_DISPATCHERH_1
`
	rdr := io.NopCloser(strings.NewReader(dispatchersHostsCsv))
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
	if err := ldr.removeContent(utils.MetaDispatcherHosts, utils.EmptyString); err != utils.ErrDSPHostNotFound {
		t.Error(err)
	}

	//cannot remove DispatcherHosts when dryrun is true
	ldr.dryRun = true
	rdr = io.NopCloser(strings.NewReader(dispatchersHostsCsv))
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
	rdr := io.NopCloser(strings.NewReader(DispatcherHostCSVContent))
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
	expectedErr := utils.ErrNoDatabaseConn
	if err := ldr.processContent(utils.MetaDispatcherHosts, utils.EmptyString); err == nil || err != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestLoaderListenAndServe(t *testing.T) {
	ldr := &Loader{}
	stopChan := make(chan struct{}, 1)
	go func() {
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

func TestRemoveThresholdsMockError(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveThresholdsMockError",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
		dataTpls: map[string][]*config.FCTemplate{
			utils.MetaThresholds: {
				{Path: "Tenant",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
					Mandatory: true},
				{Path: "ID",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
					Mandatory: true},
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(`#Tenant[0],ID[1]
	cgrates.org,REM_THRESHOLDS_1,`))
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

	expected := utils.ErrNoDatabaseConn
	ldr.dm = engine.NewDataManager(&engine.DataDBMock{
		GetThresholdProfileDrvF: func(tenant, id string) (tp *engine.ThresholdProfile, err error) {
			return &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "REM_THRESHOLDS_1",
			}, nil
		},
		SetThresholdProfileDrvF: func(tp *engine.ThresholdProfile) (err error) { return expected },
		RemThresholdProfileDrvF: func(tenant, id string) (err error) { return expected },
	}, config.CgrConfig().CacheCfg(), nil)
	if err := ldr.processContent(utils.MetaThresholds, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.removeContent(utils.MetaThresholds, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveStatQueueMockError(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestRemoveStatQueueError",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
		dataTpls: map[string][]*config.FCTemplate{
			utils.MetaStats: {
				{Path: "Tenant",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
					Mandatory: true},
				{Path: "ID",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
					Mandatory: true},
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(`#Tenant[0],ProfileID[1]
cgrates.org,REM_STATS_1`))
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

	expected := utils.ErrNoDatabaseConn
	ldr.dm = engine.NewDataManager(&engine.DataDBMock{
		GetStatQueueProfileDrvF: func(tenant, id string) (sq *engine.StatQueueProfile, err error) { return nil, nil },
		SetStatQueueProfileDrvF: func(sq *engine.StatQueueProfile) (err error) { return expected },
		RemStatQueueProfileDrvF: func(tenant, id string) (err error) { return expected },
	}, config.CgrConfig().CacheCfg(), nil)

	if err := ldr.removeContent(utils.MetaStats, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.processContent(utils.MetaStats, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveResourcesMockError(t *testing.T) {
	ldr := &Loader{
		ldrID:         "TestLoadAndRemoveResources",
		bufLoaderData: make(map[string][]LoaderData),
		dm:            engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:      "UTC",
		dataTpls: map[string][]*config.FCTemplate{
			utils.MetaResources: {
				{Path: "Tenant",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
					Mandatory: true},
				{Path: "ID",
					Type:      utils.MetaComposed,
					Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
					Mandatory: true},
			},
		},
	}
	rdr := io.NopCloser(strings.NewReader(`	#Tenant[0],ID[1]
	cgrates.org,NewRes1`))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"Resources.csv": &openedCSVFile{fileName: "Resources.csv",
				rdr: rdr, csvRdr: rdrCsv}},
	}

	expected := utils.ErrNoDatabaseConn
	ldr.dm = engine.NewDataManager(&engine.DataDBMock{
		GetResourceProfileDrvF:    func(tnt, id string) (*engine.ResourceProfile, error) { return nil, nil },
		SetResourceProfileDrvF:    func(rp *engine.ResourceProfile) error { return expected },
		RemoveResourceProfileDrvF: func(tnt, id string) error { return expected },
	}, config.CgrConfig().CacheCfg(), nil)

	if err := ldr.removeContent(utils.MetaResources, utils.EmptyString); err == nil || err != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := ldr.processContent(utils.MetaResources, utils.EmptyString); err == nil || err != expected {
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

type ccMock struct {
	calls map[string]func(args any, reply any) error
}

func (ccM *ccMock) Call(_ *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(args, reply)
	}
}

func TestStoreLoadedDataAttributes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()

	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:             nil,
		Tenant:              "",
		AttributeProfileIDs: []string{"cgrates.org:attributesID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Attributes": {
			{
				"Tenant": "cgrates.org",
				"ID":     "attributesID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaAttributes, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataResources(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:            nil,
		Tenant:             "",
		ResourceIDs:        []string{"cgrates.org:resourcesID"},
		ResourceProfileIDs: []string{"cgrates.org:resourcesID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Resources": {
			{
				"Tenant": "cgrates.org",
				"ID":     "resourcesID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaResources, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:   nil,
		Tenant:    "",
		FilterIDs: []string{"cgrates.org:filtersID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Filters": {
			{
				"Tenant": "cgrates.org",
				"ID":     "filtersID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaFilters, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataStats(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:              nil,
		Tenant:               "",
		StatsQueueIDs:        []string{"cgrates.org:statsID"},
		StatsQueueProfileIDs: []string{"cgrates.org:statsID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"StatsQueue": {
			{
				"Tenant": "cgrates.org",
				"ID":     "statsID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaStats, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataThresholds(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:             nil,
		Tenant:              "",
		ThresholdIDs:        []string{"cgrates.org:thresholdsID"},
		ThresholdProfileIDs: []string{"cgrates.org:thresholdsID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Thresholds": {
			{
				"Tenant": "cgrates.org",
				"ID":     "thresholdsID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaThresholds, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataRoutes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:         nil,
		Tenant:          "",
		RouteProfileIDs: []string{"cgrates.org:routesID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Routes": {
			{
				"Tenant": "cgrates.org",
				"ID":     "routesID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaRoutes, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataChargers(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:           nil,
		Tenant:            "",
		ChargerProfileIDs: []string{"cgrates.org:chargersID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Chargers": {
			{
				"Tenant": "cgrates.org",
				"ID":     "chargersID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaChargers, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataDispatchers(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:              nil,
		Tenant:               "",
		DispatcherProfileIDs: []string{"cgrates.org:dispatchersID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"Dispatchers": {
			{
				"Tenant": "cgrates.org",
				"ID":     "dispatchersID",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaDispatchers, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataDispatcherHosts(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:           nil,
		Tenant:            "",
		DispatcherHostIDs: []string{"cgrates.org:dispatcherHostsID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	// ldr := &Loader{

	// }
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", 0, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"DispatcherHosts": {
			{
				"Tenant":  "cgrates.org",
				"ID":      "dispatcherHostsID",
				"Address": "192.168.100.1",
			},
		},
	}
	if err := ldr.storeLoadedData(utils.MetaDispatcherHosts, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
}

func TestStoreLoadedDataWithDelay(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().CachingDelay = 5 * time.Millisecond
	argExpect := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts:           nil,
		Tenant:            "",
		DispatcherHostIDs: []string{"cgrates.org:dispatcherHostsID"},
	}
	cM := &ccMock{
		calls: map[string]func(args any, reply any) error{
			utils.CacheSv1ReloadCache: func(args any, reply any) error {
				if !reflect.DeepEqual(args, argExpect) {
					t.Errorf("Expected %v \nbut received %v", utils.ToJSON(argExpect), utils.ToJSON(args))
				}
				return nil
			},
			utils.CacheSv1Clear: func(args any, reply any) error {
				return nil
			},
		},
	}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cM
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	cacheConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	loaderCfg := config.CgrConfig().LoaderCfg()
	fltrS := engine.NewFilterS(cfg, connMgr, dm)
	ldr := NewLoader(dm, loaderCfg[0], "", cfg.GeneralCfg().CachingDelay, fltrS, connMgr, cacheConns)
	lds := map[string][]LoaderData{
		"DispatcherHosts": {
			{
				"Tenant":  "cgrates.org",
				"ID":      "dispatcherHostsID",
				"Address": "192.168.100.1",
			},
		},
	}
	tStart := time.Now()
	if err := ldr.storeLoadedData(utils.MetaDispatcherHosts, lds, utils.MetaReload); err != nil {
		t.Error(err)
	}
	got := time.Since(tStart)
	want := cfg.GeneralCfg().CachingDelay
	if diff := got - want; diff < 0 || diff > 15*time.Millisecond {
		t.Errorf("storeLoadedData duration = %v, want at least %v (diff %v, margin 15ms)", got, want, diff)
	}
}

func TestLoaderLoaderallFilesPresent(t *testing.T) {
	tests := []struct {
		name     string
		ldrType  string
		rdrs     map[string]map[string]*openedCSVFile
		expected bool
	}{
		{
			name:    "All files present",
			ldrType: "type1",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": &openedCSVFile{}, "file3": &openedCSVFile{}},
			},
			expected: true,
		},
		{
			name:    "Some files nil",
			ldrType: "type1",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": nil, "file3": &openedCSVFile{}},
			},
			expected: false,
		},
		{
			name:     "No files for type",
			ldrType:  "type1",
			rdrs:     map[string]map[string]*openedCSVFile{},
			expected: true,
		},
		{
			name:    "Type not present",
			ldrType: "type2",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": &openedCSVFile{}, "file3": &openedCSVFile{}},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ldr := &Loader{rdrs: tt.rdrs}
			result := ldr.allFilesPresent(tt.ldrType)
			reflect.DeepEqual(tt.expected, result)
		})
	}
}

func TestLoadergetLdrType(t *testing.T) {
	tests := []struct {
		name     string
		fName    string
		rdrs     map[string]map[string]*openedCSVFile
		expected string
	}{
		{
			name:  "File name exists in one loader type",
			fName: "file1",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": &openedCSVFile{}},
				"type2": {"file3": &openedCSVFile{}, "file4": &openedCSVFile{}},
			},
			expected: "type1",
		},
		{
			name:  "File name does not exist in any loader type",
			fName: "file5",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": &openedCSVFile{}},
				"type2": {"file3": &openedCSVFile{}, "file4": &openedCSVFile{}},
			},
			expected: "",
		},
		{
			name:  "Multiple loader types, file name present in one",
			fName: "file4",
			rdrs: map[string]map[string]*openedCSVFile{
				"type1": {"file1": &openedCSVFile{}, "file2": &openedCSVFile{}},
				"type2": {"file3": &openedCSVFile{}, "file4": &openedCSVFile{}},
				"type3": {"file5": &openedCSVFile{}},
			},
			expected: "type2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ldr := &Loader{rdrs: tt.rdrs}
			result := ldr.getLdrType(tt.fName)
			reflect.DeepEqual(tt.expected, result)
		})
	}
}

func TestLoaderIsFolderLocked(t *testing.T) {
	tests := []struct {
		name           string
		lockFilepath   string
		expectedLocked bool
		expectedErr    error
	}{
		{
			name:           "No lock file path",
			lockFilepath:   "",
			expectedLocked: false,
			expectedErr:    nil,
		},
		{
			name:           "Lock file does not exist",
			lockFilepath:   "nonexistent/path/to/lockfile",
			expectedLocked: false,
			expectedErr:    nil,
		},
		{
			name:           "Error checking lock file",
			lockFilepath:   "invalid/file/path",
			expectedLocked: false,
			expectedErr:    os.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ldr := &Loader{lockFilepath: tt.lockFilepath}
			locked, _ := ldr.isFolderLocked()
			if locked != tt.expectedLocked {
				t.Errorf("expected locked: %v, got: %v", tt.expectedLocked, locked)
			}

		})
	}
}

func TestProcessFiles(t *testing.T) {
	tempDir := t.TempDir()
	tempFileName := path.Join(tempDir, "testfile.csv")
	fileContent := "header1,header2\nvalue1,value2\n"
	if err := os.WriteFile(tempFileName, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to create temp CSV file: %v", err)
	}
	ldr := &Loader{
		tpInDir:  tempDir,
		fieldSep: ",",
		rdrs: map[string]map[string]*openedCSVFile{
			"testLoader": {
				"testfile.csv": &openedCSVFile{},
			},
		},
	}
	if err := ldr.processFiles("testLoader", "cacheOption", "storeOption"); err != nil {
		t.Errorf("processFiles returned an error: %v", err)
	}
	if _, ok := ldr.rdrs["testLoader"]["testfile.csv"]; !ok {
		t.Error("File was not added to ldr.rdrs")
	}
}

func TestProcessFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFileName := path.Join(tempDir, "testfile.csv")
	fileContent := "header1,header2\nvalue1,value2\n"
	err := os.WriteFile(tempFileName, []byte(fileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp CSV file: %v", err)
	}
	ldr := &Loader{}
	err = ldr.processFile("testfile.csv")
	if err != nil {
		t.Errorf("processFile returned an error: %v", err)
	}
	if _, ok := ldr.rdrs["testLoader"]["testfile.csv"]; ok {
		t.Errorf("File was not added to ldr.rdrs")
	}
}
