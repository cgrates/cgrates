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

/*
func TestSetDefaults(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdrc" id="CDRC-CSVDF">
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
		xmlCdrc.CdrFormat != dfCfg.CdrcCdrType ||
		xmlCdrc.CsvSeparator != dfCfg.CdrcCsvSep ||
		xmlCdrc.CdrInDir != dfCfg.CdrcCdrInDir ||
		xmlCdrc.CdrOutDir != dfCfg.CdrcCdrOutDir ||
		xmlCdrc.CdrSourceId != dfCfg.CdrcSourceId ||
		len(xmlCdrc.CdrFields) != len(dfCfg.CdrcCdrFields) {
		t.Error("Failed loading default configuration")
	}
}
*/

func TestParseXmlCdrcConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdrc" id="CDRC-CSV1">
    <enabled>true</enabled>
    <cdrs_address>internal</cdrs_address>
    <cdr_format>csv</cdr_format>
    <field_separator>,</field_separator>
    <run_delay>0</run_delay>
    <cdr_in_dir>/var/log/cgrates/cdrc/in</cdr_in_dir>
    <cdr_out_dir>/var/log/cgrates/cdrc/out</cdr_out_dir>
    <cdr_source_id>freeswitch_csv</cdr_source_id>
    <fields>
      <field tag="accid" value="0;13" />
      <field tag="reqtype" value="1" />
      <field tag="direction" value="2" />
      <field tag="tenant" value="3" />
      <field tag="category" value="4" />
      <field tag="account" value="5" />
      <field tag="subject" value="6" />
      <field tag="destination" value="7" />
      <field tag="setup_time" value="8" />
      <field tag="answer_time" value="9" />
      <field tag="usage" value="10" />
      <field tag="extr1" value="11" />
      <field tag="extr2" value="12" />
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
	enabled := true
	cdrsAddr := "internal"
	cdrFormat := "csv"
	fldSep := ","
	runDelay := int64(0)
	cdrInDir := "/var/log/cgrates/cdrc/in"
	cdrOutDir := "/var/log/cgrates/cdrc/out"
	cdrSrcId := "freeswitch_csv"
	expectCdrc := &CgrXmlCdrcCfg{Enabled: &enabled, CdrsAddress: &cdrsAddr, CdrFormat: &cdrFormat, FieldSeparator: &fldSep,
		RunDelay: &runDelay, CdrInDir: &cdrInDir, CdrOutDir: &cdrOutDir, CdrSourceId: &cdrSrcId}
	accIdTag, reqTypeTag, dirTag, tntTag, categTag, acntTag, subjTag, dstTag, sTimeTag, aTimeTag, usageTag, extr1, extr2 := utils.ACCID,
		utils.REQTYPE, utils.DIRECTION, utils.TENANT, utils.CATEGORY, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.SETUP_TIME, utils.ANSWER_TIME, utils.USAGE, "extr1", "extr2"
	accIdVal, reqVal, dirVal, tntVal, categVal, acntVal, subjVal, dstVal, sTimeVal, aTimeVal, usageVal, extr1Val, extr2Val := "0;13", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"
	cdrFlds := []*XmlCfgCdrField{
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &accIdTag, Value: &accIdVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &reqTypeTag, Value: &reqVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &dirTag, Value: &dirVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &tntTag, Value: &tntVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &categTag, Value: &categVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &acntTag, Value: &acntVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &subjTag, Value: &subjVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &dstTag, Value: &dstVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &sTimeTag, Value: &sTimeVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &aTimeTag, Value: &aTimeVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &usageTag, Value: &usageVal},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &extr1, Value: &extr1Val},
		&XmlCfgCdrField{XMLName: xml.Name{Local: "field"}, Tag: &extr2, Value: &extr2Val}}
	expectCdrc.CdrFields = cdrFlds
	if !reflect.DeepEqual(expectCdrc, cdrcfgs["CDRC-CSV1"]) {
		t.Errorf("Expecting: %v, received: %v", expectCdrc, cdrcfgs["CDRC-CSV1"])
	}
}
