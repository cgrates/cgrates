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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"strings"
	"testing"
)

var cfgDoc *CgrXmlCfgDocument // Will be populated by first test

func TestXmlCdreCfgParseXmlConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdre" id="CDRE-FW1">
    <cdr_format>fwv</cdr_format>
    <data_usage_multiply_factor>1.0</data_usage_multiply_factor>
    <cost_multiply_factor>0.0</cost_multiply_factor>
    <cost_rounding_decimals>-1</cost_rounding_decimals>
    <cost_shift_digits>0</cost_shift_digits>
    <mask_destination_id>MASKED_DESTINATIONS</mask_destination_id>
    <mask_length>0</mask_length>
    <export_dir>/var/log/cgrates/cdre</export_dir>
    <export_template>
      <header>
        <fields>
          <field tag="TypeOfRecord" type="constant" value="10" width="2" />
          <field tag="Filler1" type="filler" width="3" />
          <field tag="DistributorCode" type="constant" value="VOI" width="3" />
          <field tag="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
          <field tag="LastCdr" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
          <field tag="FileCreationfTime" type="metatag" value="time_now" layout="020106150400" width="12" />
          <field tag="Version" type="constant" value="01" width="2" />
          <field tag="Filler2" type="filler" width="105" />
        </fields>
      </header>
      <content>
        <fields>
          <field tag="TypeOfRecord" type="constant" value="20" width="2" />
          <field tag="Account" type="cdrfield" value="cgrid" width="12" mandatory="true" />
          <field tag="Subject" type="cdrfield" value="subject" strip="left" padding="left" width="5" />
          <field tag="CLI" type="cdrfield" value="cli" strip="xright" width="15" />
          <field tag="Destination" type="cdrfield" value="destination" strip="xright" width="24" />
          <field tag="TOR" type="constant" value="02" width="2" />
          <field tag="SubtypeTOR" type="constant" value="11" width="4" />
          <field tag="SetupTime" type="cdrfield" value="start_time" layout="020106150400" width="12" />
          <field tag="Duration" type="cdrfield" value="duration" width="6" />
          <field tag="DataVolume" type="filler" width="6" />
          <field tag="TaxCode" type="constant" value="1" width="1" />
          <field tag="OperatorCode" type="cdrfield" value="operator" width="2" />
          <field tag="ProductId" type="cdrfield" value="productid" width="5" />
          <field tag="NetworkId" type="constant" value="3" width="1" />
          <field tag="CallId" type="cdrfield" value="accid" width="16" />
          <field tag="Filler" type="filler" width="8" />
          <field tag="Filler" type="filler" width="8" />
          <field tag="TerminationCode" type="cdrfield" value="~cost_details:s/&quot;MatchedDestId&quot;:&quot;.+_(\s\s\s\s\s)&quot;/$1/" width="5" />
          <field tag="Cost" type="cdrfield" value="cost" padding="zeroleft" width="9" />
          <field tag="CalledMask" type="cdrfield" value="calledmask" width="1" />
        </fields>
      </content>
      <trailer>
        <fields>
          <field tag="TypeOfRecord" type="constant" value="90" width="2" />
          <field tag="Filler1" type="filler" width="3" />
          <field tag="DistributorCode" type="constant" value="VOI" width="3" />
          <field tag="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
          <field tag="NumberOfRecords" type="metatag" value="cdrs_number" padding="zeroleft" width="6" />
          <field tag="CdrsDuration" type="metatag" value="cdrs_duration" padding="zeroleft" width="8" />
          <field tag="FirstCdrTime" type="metatag" value="first_cdr_time" layout="020106150400" width="12" />
          <field tag="LastCdrTime" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
          <field tag="Filler1" type="filler" width="93" />
        </fields>
      </trailer>
    </export_template>
  </configuration>
  <configuration section="cdre" type="csv" id="CHECK-CSV1">
  <export_template>
    <content>
     <fields>
      <field tag="CGRID" type="cdrfield" value="cgrid" width="40"/>
      <field tag="RatingSubject" type="cdrfield" value="subject" width="24" padding="left" strip="xright" mandatory="true"/>
      <field tag="Usage" type="cdrfield" value="usage" layout="seconds" width="6" padding="right" mandatory="true"/>
      <field tag="AccountReference" type="http_post" value="https://localhost:8000" width="10" strip="xright" padding="left" mandatory="true" />
      <field tag="AccountType" type="http_post" value="https://localhost:8000" width="10" strip="xright" padding="left" mandatory="true" />
      <field tag="MultipleMed1" type="combimed" value="cost" strip="xright" padding="left" mandatory="true" filter="~mediation_runid:s/DEFAULT/SECOND_RUN/"/>
     </fields>
    </content>
   </export_template>
  </configuration>

</document>`
	var err error
	reader := strings.NewReader(cfgXmlStr)
	if cfgDoc, err = ParseCgrXmlConfig(reader); err != nil {
		t.Error(err.Error())
	} else if cfgDoc == nil {
		t.Fatal("Could not parse xml configuration document")
	}
	if len(cfgDoc.cdres) != 2 {
		t.Error("Did not cache")
	}
}

func TestXmlCdreCfgGetCdreCfg(t *testing.T) {
	cdreFWCfg := cfgDoc.GetCdreCfgs("CDRE-FW1")
	if cdreFWCfg == nil {
		t.Error("Could not parse CdreFw instance")
	}
	if len(cdreFWCfg["CDRE-FW1"].Header.Fields) != 8 {
		t.Error("Unexpected number of header fields parsed", len(cdreFWCfg["CDRE-FW1"].Header.Fields))
	}
	if len(cdreFWCfg["CDRE-FW1"].Content.Fields) != 20 {
		t.Error("Unexpected number of content fields parsed", len(cdreFWCfg["CDRE-FW1"].Content.Fields))
	}
	if len(cdreFWCfg["CDRE-FW1"].Trailer.Fields) != 9 {
		t.Error("Unexpected number of trailer fields parsed", len(cdreFWCfg["CDRE-FW1"].Trailer.Fields))
	}
	cdreCsvCfg1 := cfgDoc.GetCdreCfgs("CHECK-CSV1")
	if cdreCsvCfg1 == nil {
		t.Error("Could not parse CdreFw instance")
	}
	if len(cdreCsvCfg1["CHECK-CSV1"].Content.Fields) != 6 {
		t.Error("Unexpected number of content fields parsed", len(cdreCsvCfg1["CHECK-CSV1"].Content.Fields))
	}
}

func TestNewCdreConfigFromXmlCdreCfg(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdre" type="fixed_width" id="CDRE-FW2">
    <cdr_format>fwv</cdr_format>
    <field_separator>;</field_separator>
    <data_usage_multiply_factor>1024.0</data_usage_multiply_factor>
    <cost_multiply_factor>1.19</cost_multiply_factor>
    <cost_rounding_decimals>-1</cost_rounding_decimals>
    <cost_shift_digits>-3</cost_shift_digits>
    <mask_destination_id>MASKED_DESTINATIONS</mask_destination_id>
    <mask_length>1</mask_length>
    <export_dir>/var/log/cgrates/cdre</export_dir>
    <export_template>
      <header>
        <fields>
          <field tag="TypeOfRecord" type="constant" value="10" width="2" />
          <field tag="LastCdr" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
        </fields>
      </header>
      <content>
        <fields>
          <field tag="OperatorCode" type="cdrfield" value="operator" width="2" />
          <field tag="ProductId" type="cdrfield" value="productid" width="5" />
          <field tag="NetworkId" type="constant" value="3" width="1" />
          <field tag="FromHttpPost1" type="http_post" value="https://localhost:8000" width="10" strip="xright" padding="left" />
          <field tag="CombiMed1" type="combimed" value="cost" width="10" strip="xright" padding="left" filter="~mediation_runid:s/DEFAULT/SECOND_RUN/"/>
        </fields>
      </content>
      <trailer>
        <fields>
          <field tag="DistributorCode" type="constant" value="VOI" width="3" />
          <field tag="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
        </fields>
      </trailer>
    </export_template>
  </configuration>
</document>`
	var err error
	reader := strings.NewReader(cfgXmlStr)
	if cfgDoc, err = ParseCgrXmlConfig(reader); err != nil {
		t.Error(err.Error())
	} else if cfgDoc == nil {
		t.Fatal("Could not parse xml configuration document")
	}
	xmlCdreCfgs := cfgDoc.GetCdreCfgs("CDRE-FW2")
	if xmlCdreCfgs == nil {
		t.Error("Could not parse XmlCdre instance")
	}
	eCdreCfg := &CdreConfig{
		CdrFormat:               "fwv",
		FieldSeparator:          ';',
		DataUsageMultiplyFactor: 1024.0,
		CostMultiplyFactor:      1.19,
		CostRoundingDecimals:    -1,
		CostShiftDigits:         -3,
		MaskDestId:              "MASKED_DESTINATIONS",
		MaskLength:              1,
		ExportDir:               "/var/log/cgrates/cdre",
	}
	fltrCombiMed, _ := utils.ParseRSRFields("~mediation_runid:s/DEFAULT/SECOND_RUN/", utils.INFIELD_SEP)
	torVal, _ := utils.ParseRSRFields("^10", utils.INFIELD_SEP)
	lastCdrVal, _ := utils.ParseRSRFields("^last_cdr_time", utils.INFIELD_SEP)
	eCdreCfg.HeaderFields = []*CfgCdrField{
		&CfgCdrField{
			Tag:   "TypeOfRecord",
			Type:  "constant",
			Value: torVal,
			Width: 2,
		},
		&CfgCdrField{
			Tag:        "LastCdr",
			Type:       "metatag",
			CdrFieldId: "last_cdr_time",
			Value:      lastCdrVal,
			Layout:     "020106150400",
			Strip:      "xright",
			Padding:    "left",
			Width:      12,
		},
	}
	networkIdVal, _ := utils.ParseRSRFields("^3", utils.INFIELD_SEP)
	fromHttpPost1Val, _ := utils.ParseRSRFields("^https://localhost:8000", utils.INFIELD_SEP)
	eCdreCfg.ContentFields = []*CfgCdrField{
		&CfgCdrField{
			Tag:        "OperatorCode",
			Type:       "cdrfield",
			CdrFieldId: "operator",
			Value: []*utils.RSRField{
				&utils.RSRField{Id: "operator"}},
			Width:   2,
			Strip:   "xright",
			Padding: "left",
		},
		&CfgCdrField{
			Tag:        "ProductId",
			Type:       "cdrfield",
			CdrFieldId: "productid",
			Value: []*utils.RSRField{
				&utils.RSRField{Id: "productid"}},
			Width:   5,
			Strip:   "xright",
			Padding: "left",
		},
		&CfgCdrField{
			Tag:   "NetworkId",
			Type:  "constant",
			Value: networkIdVal,
			Width: 1,
		},
		&CfgCdrField{
			Tag:     "FromHttpPost1",
			Type:    "http_post",
			Value:   fromHttpPost1Val,
			Width:   10,
			Strip:   "xright",
			Padding: "left",
		},
		&CfgCdrField{
			Tag:        "CombiMed1",
			Type:       "combimed",
			CdrFieldId: "cost",
			Value: []*utils.RSRField{
				&utils.RSRField{Id: "cost"}},
			Width:     10,
			Strip:     "xright",
			Padding:   "left",
			Filter:    fltrCombiMed,
			Mandatory: true,
		},
	}
	distribCodeVal, _ := utils.ParseRSRFields("^VOI", utils.INFIELD_SEP)
	fileSeqNrVal, _ := utils.ParseRSRFields("^export_id", utils.INFIELD_SEP)
	eCdreCfg.TrailerFields = []*CfgCdrField{
		&CfgCdrField{
			Tag:   "DistributorCode",
			Type:  "constant",
			Value: distribCodeVal,
			Width: 3,
		},
		&CfgCdrField{
			Tag:        "FileSeqNr",
			Type:       "metatag",
			CdrFieldId: "export_id",
			Value:      fileSeqNrVal,
			Width:      5,
			Strip:      "xright",
			Padding:    "zeroleft",
		},
	}
	if rcvCdreCfg, err := NewCdreConfigFromXmlCdreCfg(xmlCdreCfgs["CDRE-FW2"]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvCdreCfg, eCdreCfg) {
		t.Errorf("Expecting: %v, received: %v", eCdreCfg, rcvCdreCfg)
	}
}
