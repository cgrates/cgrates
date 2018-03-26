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

var attrsCSV = `#Tenant,ID,Contexts,FilterIDs,ActivationInterval,FieldName,Initial,Substitute,Append,Weight
cgrates.org,TestLoader1,*sessions;*cdrs,*string:Account:1007,2014-01-14T00:00:00Z,Account,*any,1001,false,10
cgrates.org,TestLoader1,lcr,*string:Account:1008;*string:Account:1009,,Subject,*any,1001,true,
`

func TestLoaderProcessContentSingleFile(t *testing.T) {
	data, _ := engine.NewMapStorage()
	ldr := &Loader{
		ldrID: "TestLoaderProcessContent",
		//rdrs          map[string]map[string]*openedCSVFile
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
	if ap, err := ldr.dm.GetAttributeProfile("cgrates.org", "TestLoader1",
		false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAP.Attributes, ap.Attributes) {
		t.Errorf("expecting: %s, received: %s",
			utils.ToJSON(eAP), utils.ToJSON(ap))
	}
}
