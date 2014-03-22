/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"strings"
	"testing"
)

var cfgDoc *CgrXmlCfgDocument // Will be populated by first test

func TestParseXmlConfig(t *testing.T) {
	cfgXmlStr := `<?xml version="1.0" encoding="UTF-8" ?>
<document type="cgrates/xml">
  <configuration section="cdre" type="fixed_width" id="CDRE-FW1">
   <header>
    <fields>
     <field name="RecordType" type="constant" value="10" width="2"/>
     <field name="Filler1" type="filler" width="3"/>
     <field name="NetworkProviderCode" type="constant" value="VOI" width="3"/>
     <field name="FileSeqNr" type="metatag" id="exportid" lpadding="true" padding_char="0" width="5"/>
     <field name="CutOffTime" type="metatag" id="time_lastcdr" layout="020106150400" width="12"/>
     <field name="FileCreationfTime" type="metatag" id="time_now" layout="020106150400" width="12"/>
     <field name="FileSpecificationVersion" type="constant" value="01" width="2"/>
     <field name="Filler2" type="filler" width="105"/>
    </fields>
   </header>
   <content>
    <fields>
     <field name="RecordType" type="constant" value="20" width="2"/>
     <field name="SIPTrunkID" type="cdrfield" id="cgrid" width="12"/>
     <field name="ConnectionNumber" type="cdrfield" id="subject" strip="left" padding="left" padding_char="0" width="5"/>
     <field name="ANumber" type="cdrfield" id="cli" strip="xright" width="15"/>
     <field name="CalledNumber" type="cdrfield" id="destination" strip="xright" width="24"/>
     <field name="ServiceType" type="constant" value="02" width="2"/>
     <field name="ServiceIdentification" type="constant" value="11" width="4"/>
     <field name="StartChargingDateTime" type="cdrfield" id="start_time" layout="020106150400" width="12"/>
     <field name="ChargeableTime" type="cdrfield" id="duration" width="6"/>
     <field name="DataVolume" type="filler" width="6"/>
     <field name="TaxCode" type="constant" value="1" width="1"/>
     <field name="OperatorTAPCode" type="cdrfield" id="opertapcode" width="2"/>
     <field name="ProductNumber" type="cdrfield" id="productnumber" width="5"/>
     <field name="NetworkSubtype" type="constant" value="3" width="1"/>
     <field name="SessionID" type="cdrfield" id="accid" width="16"/>
     <field name="VolumeUP" type="filler" width="8"/>
     <field name="VolumeDown" type="filler" width="8"/>
     <field name="TerminatingOperator" type="concatenated_cdrfield" id="tapcode,operatorcode" width="5"/>
     <field name="EndCharge" type="metatag" id="total_cost" lpadding="true" padding_char="0" width="9"/>
     <field name="CallMaskingIndicator" type="cdrfield" id="calledmask" width="1"/>
    </fields>
   </content>
   <trailer>
    <fields>
     <field name="RecordType" type="constant" value="90" width="2"/>
     <field name="Filler1" type="filler" width="3"/>
     <field name="NetworkProviderCode" type="constant" value="VOI" width="3"/>
     <field name="FileSeqNr" type="metatag" id="exportid" lpadding="true" padding_char="0" width="5"/>
     <field name="TotalNrRecords" type="metatag" id="nr_cdrs" lpadding="true" padding_char="0" width="6"/>
     <field name="TotalDurRecords" type="metatag" id="dur_cdrs" lpadding="true" padding_char="0" width="8"/>
     <field name="EarliestCDRTime" type="metatag" id="first_cdr_time" layout="020106150400" width="12"/>
     <field name="LatestCDRTime" type="metatag" id="last_cdr_time" layout="020106150400" width="12"/>
     <field name="Filler1" type="filler" width="93"/>
    </fields>
   </trailer>
  </configuration>
</document>`
	var err error
	reader := strings.NewReader(cfgXmlStr)
	if cfgDoc, err = ParseCgrXmlConfig(reader); err != nil {
		t.Error(err.Error())
	} else if cfgDoc == nil {
		t.Fatal("Could not parse xml configuration document")
	}
}

func TestCacheCdreFWCfgs(t *testing.T) {
	if len(cfgDoc.cdrefws) != 0 {
		t.Error("Cache should be empty before caching")
	}
	if err := cfgDoc.cacheCdreFWCfgs(); err != nil {
		t.Error(err)
	} else if len(cfgDoc.cdrefws) != 1 {
		t.Error("Did not cache")
	}
}

func TestGetCdreFWCfg(t *testing.T) {
	cdreFWCfg, err := cfgDoc.GetCdreFWCfg("CDRE-FW1")
	if err != nil {
		t.Error(err)
	} else if cdreFWCfg == nil {
		t.Error("Could not parse CdreFw instance")
	}
	if len(cdreFWCfg.Header.Fields) != 8 {
		t.Error("Unexpected number of header fields parsed", len(cdreFWCfg.Header.Fields))
	}
	if len(cdreFWCfg.Content.Fields) != 20 {
		t.Error("Unexpected number of content fields parsed", len(cdreFWCfg.Content.Fields))
	}
	if len(cdreFWCfg.Trailer.Fields) != 9 {
		t.Error("Unexpected number of trailer fields parsed", len(cdreFWCfg.Trailer.Fields))
	}
}
