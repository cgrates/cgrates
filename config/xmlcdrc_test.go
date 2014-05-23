/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package config

import (
	"encoding/xml"
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"strings"
	"testing"
)

var cfgDocCdrc *CgrXmlCfgDocument // Will be populated by first test

func TestPopulateRSRFIeld(t *testing.T) {
	cdrcField := CdrcField{Id: "TEST1", Filter: `~effective_caller_id_number:s/(\d+)/+$1/`}
	if err := cdrcField.PopulateRSRField(); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrcField.rsrField == nil {
		t.Error("Failed loading the RSRField")
	}
	cdrcField = CdrcField{Id: "TEST2", Filter: `99`}
	if err := cdrcField.PopulateRSRField(); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrcField.rsrField == nil {
		t.Error("Failed loading the RSRField")
	}
}

func TestSetDefaults(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdrc" type="csv" id="CDRC-CSVDF">
    <enabled>true</enabled>
  </configuration>
</document>`
	var xmlCdrc *CgrXmlCdrcCfg
	reader := strings.NewReader(cfgXmlStr)
	if cfgDocCdrcDf, err := ParseCgrXmlConfig(reader); err != nil {
		t.Error(err.Error())
	} else if cfgDocCdrcDf == nil {
		t.Fatal("Could not parse xml configuration document")
	} else if len(cfgDocCdrcDf.cdrcs) != 1 {
		t.Error("Did not load cdrc")
	} else {
		xmlCdrc = cfgDocCdrcDf.cdrcs["CDRC-CSVDF"]
	}
	dfCfg, _ := NewDefaultCGRConfig()
	xmlCdrc.setDefaults()
	if xmlCdrc.CdrsAddress != dfCfg.CdrcCdrs ||
		xmlCdrc.CdrType != dfCfg.CdrcCdrType ||
		xmlCdrc.CsvSeparator != dfCfg.CdrcCsvSep ||
		xmlCdrc.CdrInDir != dfCfg.CdrcCdrInDir ||
		xmlCdrc.CdrOutDir != dfCfg.CdrcCdrOutDir ||
		xmlCdrc.CdrSourceId != dfCfg.CdrcSourceId ||
		len(xmlCdrc.CdrFields) != len(dfCfg.CdrcCdrFields) {
		t.Error("Failed loading default configuration")
	}
}

func TestParseXmlCdrcConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdrc" type="csv" id="CDRC-CSV1">
    <enabled>true</enabled>
    <cdrs_address>internal</cdrs_address>
    <cdr_type>csv</cdr_type>
    <csv_separator>,</csv_separator>
    <run_delay>0</run_delay>
    <cdr_in_dir>/var/log/cgrates/cdrc/in</cdr_in_dir>
    <cdr_out_dir>/var/log/cgrates/cdrc/out</cdr_out_dir>
    <cdr_source_id>freeswitch_csv</cdr_source_id>
    <fields>
      <field id="accid" filter="0" />
      <field id="reqtype" filter="1" />
      <field id="direction" filter="2" />
      <field id="tenant" filter="3" />
      <field id="category" filter="4" />
      <field id="account" filter="5" />
      <field id="subject" filter="6" />
      <field id="destination" filter="7" />
      <field id="setup_time" filter="8" />
      <field id="answer_time" filter="9" />
      <field id="usage" filter="10" />
      <field id="extr1" filter="11" />
      <field id="extr2" filter="12" />
    </fields>
  </configuration>
</document>`
	var err error
	reader := strings.NewReader(cfgXmlStr)
	if cfgDocCdrc, err = ParseCgrXmlConfig(reader); err != nil {
		t.Error(err.Error())
	} else if cfgDocCdrc == nil {
		t.Fatal("Could not parse xml configuration document")
	}
	if len(cfgDocCdrc.cdrcs) != 1 {
		t.Error("Did not cache")
	}
}

func TestGetCdrcCfgs(t *testing.T) {
	cdrcfgs := cfgDocCdrc.GetCdrcCfgs("CDRC-CSV1")
	if cdrcfgs == nil {
		t.Error("No config instance returned")
	}
	expectCdrc := &CgrXmlCdrcCfg{Enabled: true, CdrsAddress: "internal", CdrType: "csv", CsvSeparator: ",",
		RunDelay: 0, CdrInDir: "/var/log/cgrates/cdrc/in", CdrOutDir: "/var/log/cgrates/cdrc/out", CdrSourceId: "freeswitch_csv"}
	cdrFlds := []*CdrcField{
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.ACCID, Filter: "0"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.REQTYPE, Filter: "1"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.DIRECTION, Filter: "2"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.TENANT, Filter: "3"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.CATEGORY, Filter: "4"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.ACCOUNT, Filter: "5"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.SUBJECT, Filter: "6"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.DESTINATION, Filter: "7"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.SETUP_TIME, Filter: "8"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.ANSWER_TIME, Filter: "9"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: utils.USAGE, Filter: "10"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: "extr1", Filter: "11"},
		&CdrcField{XMLName: xml.Name{Local: "field"}, Id: "extr2", Filter: "12"}}
	for _, fld := range cdrFlds {
		fld.PopulateRSRField()
	}
	expectCdrc.CdrFields = cdrFlds
	if !reflect.DeepEqual(expectCdrc, cdrcfgs["CDRC-CSV1"]) {
		t.Errorf("Expecting: %v, received: %v", expectCdrc, cdrcfgs["CDRC-CSV1"])
	}
}

func TestCdrRSRFields(t *testing.T) {
	cdrcfgs := cfgDocCdrc.GetCdrcCfgs("CDRC-CSV1")
	if cdrcfgs == nil {
		t.Error("No config instance returned")
	}
	eRSRFields := map[string]*utils.RSRField{
		utils.ACCID:       &utils.RSRField{Id: "0"},
		utils.REQTYPE:     &utils.RSRField{Id: "1"},
		utils.DIRECTION:   &utils.RSRField{Id: "2"},
		utils.TENANT:      &utils.RSRField{Id: "3"},
		utils.CATEGORY:    &utils.RSRField{Id: "4"},
		utils.ACCOUNT:     &utils.RSRField{Id: "5"},
		utils.SUBJECT:     &utils.RSRField{Id: "6"},
		utils.DESTINATION: &utils.RSRField{Id: "7"},
		utils.SETUP_TIME:  &utils.RSRField{Id: "8"},
		utils.ANSWER_TIME: &utils.RSRField{Id: "9"},
		utils.USAGE:       &utils.RSRField{Id: "10"},
		"extr1":           &utils.RSRField{Id: "11"},
		"extr2":           &utils.RSRField{Id: "12"},
	}
	if rsrFields := cdrcfgs["CDRC-CSV1"].CdrRSRFields(); !reflect.DeepEqual(rsrFields, eRSRFields) {
		t.Errorf("Expecting: %v, received: %v", eRSRFields, rsrFields)
	}
}
