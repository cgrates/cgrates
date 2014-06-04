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

func TestXmlCdreCfgPopulateCdreRSRFIeld(t *testing.T) {
	cdreField := CgrXmlCfgCdrField{Name: "TEST1", Type: "cdrfield", Value: `~effective_caller_id_number:s/(\d+)/+$1/`}
	if err := cdreField.populateRSRField(); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdreField.valueAsRsrField == nil {
		t.Error("Failed loading the RSRField")
	}
	valRSRField, _ := utils.NewRSRField(`~effective_caller_id_number:s/(\d+)/+$1/`)
	if recv := cdreField.ValueAsRSRField(); !reflect.DeepEqual(valRSRField, recv) {
		t.Errorf("Expecting %v, received %v", valRSRField, recv)
	}
	cdreField = CgrXmlCfgCdrField{Name: "TEST1", Type: "constant", Value: `someval`}
	if err := cdreField.populateRSRField(); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdreField.valueAsRsrField != nil {
		t.Error("Should not load the RSRField")
	}
}

func TestXmlCdreCfgParseXmlConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdre" type="fixed_width" id="CDRE-FW1">
    <cdr_format>fwv</cdr_format>
    <data_usage_multiply_factor>0.0</data_usage_multiply_factor>
    <cost_multiply_factor>0.0</cost_multiply_factor>
    <cost_rounding_decimals>-1</cost_rounding_decimals>
    <cost_shift_digits>0</cost_shift_digits>
    <mask_destination_id>MASKED_DESTINATIONS</mask_destination_id>
    <mask_length>0</mask_length>
    <export_dir>/var/log/cgrates/cdre</export_dir>
    <export_template>
      <header>
        <fields>
          <field name="TypeOfRecord" type="constant" value="10" width="2" />
          <field name="Filler1" type="filler" width="3" />
          <field name="DistributorCode" type="constant" value="VOI" width="3" />
          <field name="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
          <field name="LastCdr" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
          <field name="FileCreationfTime" type="metatag" value="time_now" layout="020106150400" width="12" />
          <field name="Version" type="constant" value="01" width="2" />
          <field name="Filler2" type="filler" width="105" />
        </fields>
      </header>
      <content>
        <fields>
          <field name="TypeOfRecord" type="constant" value="20" width="2" />
          <field name="Account" type="cdrfield" value="cgrid" width="12" mandatory="true" />
          <field name="Subject" type="cdrfield" value="subject" strip="left" padding="left" width="5" />
          <field name="CLI" type="cdrfield" value="cli" strip="xright" width="15" />
          <field name="Destination" type="cdrfield" value="destination" strip="xright" width="24" />
          <field name="TOR" type="constant" value="02" width="2" />
          <field name="SubtypeTOR" type="constant" value="11" width="4" />
          <field name="SetupTime" type="cdrfield" value="start_time" layout="020106150400" width="12" />
          <field name="Duration" type="cdrfield" value="duration" width="6" multiply_factor_voice="1000" />
          <field name="DataVolume" type="filler" width="6" />
          <field name="TaxCode" type="constant" value="1" width="1" />
          <field name="OperatorCode" type="cdrfield" value="operator" width="2" />
          <field name="ProductId" type="cdrfield" value="productid" width="5" />
          <field name="NetworkId" type="constant" value="3" width="1" />
          <field name="CallId" type="cdrfield" value="accid" width="16" />
          <field name="Filler" type="filler" width="8" />
          <field name="Filler" type="filler" width="8" />
          <field name="TerminationCode" type="cdrfield" value="~cost_details:s/&quot;MatchedDestId&quot;:&quot;.+_(\s\s\s\s\s)&quot;/$1/" width="5" />
          <field name="Cost" type="cdrfield" value="cost" padding="zeroleft" width="9" />
          <field name="CalledMask" type="cdrfield" value="calledmask" width="1" />
        </fields>
      </content>
      <trailer>
        <fields>
          <field name="TypeOfRecord" type="constant" value="90" width="2" />
          <field name="Filler1" type="filler" width="3" />
          <field name="DistributorCode" type="constant" value="VOI" width="3" />
          <field name="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
          <field name="NumberOfRecords" type="metatag" value="cdrs_number" padding="zeroleft" width="6" />
          <field name="CdrsDuration" type="metatag" value="cdrs_duration" padding="zeroleft" width="8" />
          <field name="FirstCdrTime" type="metatag" value="first_cdr_time" layout="020106150400" width="12" />
          <field name="LastCdrTime" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
          <field name="Filler1" type="filler" width="93" />
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
	if len(cfgDoc.cdres) != 1 {
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
}

func TestXmlCdreCfgAsCdreConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<document type="cgrates/xml">
  <configuration section="cdre" type="fixed_width" id="CDRE-FW2">
    <cdr_format>fwv</cdr_format>
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
          <field name="TypeOfRecord" type="constant" value="10" width="2" />
          <field name="LastCdr" type="metatag" value="last_cdr_time" layout="020106150400" width="12" />
        </fields>
      </header>
      <content>
        <fields>
          <field name="OperatorCode" type="cdrfield" value="operator" width="2" />
          <field name="ProductId" type="cdrfield" value="productid" width="5" />
          <field name="NetworkId" type="constant" value="3" width="1" />
          <field name="FromHttpPost1" type="http_post" value="https://localhost:8000" width="10" strip="xright" padding="left" />
        </fields>
      </content>
      <trailer>
        <fields>
          <field name="DistributorCode" type="constant" value="VOI" width="3" />
          <field name="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5" />
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
		DataUsageMultiplyFactor: 1024.0,
		CostMultiplyFactor:      1.19,
		CostRoundingDecimals:    -1,
		CostShiftDigits:         -3,
		MaskDestId:              "MASKED_DESTINATIONS",
		MaskLength:              1,
		ExportDir:               "/var/log/cgrates/cdre",
	}
	eCdreCfg.HeaderFields = []*CdreCdrField{
		&CdreCdrField{
			Name:  "TypeOfRecord",
			Type:  "constant",
			Value: "10",
			Width: 2},
		&CdreCdrField{
			Name:   "LastCdr",
			Type:   "metatag",
			Value:  "last_cdr_time",
			Layout: "020106150400",
			Width:  12},
	}
	eCdreCfg.ContentFields = []*CdreCdrField{
		&CdreCdrField{
			Name:            "OperatorCode",
			Type:            "cdrfield",
			Value:           "operator",
			Width:           2,
			valueAsRsrField: &utils.RSRField{Id: "operator"},
		},
		&CdreCdrField{
			Name:            "ProductId",
			Type:            "cdrfield",
			Value:           "productid",
			Width:           5,
			valueAsRsrField: &utils.RSRField{Id: "productid"},
		},
		&CdreCdrField{
			Name:  "NetworkId",
			Type:  "constant",
			Value: "3",
			Width: 1,
		},
		&CdreCdrField{
			Name:    "FromHttpPost1",
			Type:    "http_post",
			Value:   "https://localhost:8000",
			Width:   10,
			Strip:   "xright",
			Padding: "left",
		},
	}
	eCdreCfg.TrailerFields = []*CdreCdrField{
		&CdreCdrField{
			Name:  "DistributorCode",
			Type:  "constant",
			Value: "VOI",
			Width: 3,
		},
		&CdreCdrField{
			Name:    "FileSeqNr",
			Type:    "metatag",
			Value:   "export_id",
			Width:   5,
			Padding: "zeroleft",
		},
	}
	if rcvCdreCfg := xmlCdreCfgs["CDRE-FW2"].AsCdreConfig(); !reflect.DeepEqual(rcvCdreCfg, eCdreCfg) {
		t.Errorf("Expecting: %v, received: %v", eCdreCfg, rcvCdreCfg)
	}
}
