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
     <field name="TypeOfRecord" type="constant" value="10" width="2"/>
     <field name="Filler1" type="filler" width="3"/>
     <field name="DistributorCode" type="constant" value="VOI" width="3"/>
     <field name="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5"/>
     <field name="LastCdr" type="metatag" value="last_cdr_time" layout="020106150400" width="12"/>
     <field name="FileCreationfTime" type="metatag" value="time_now" layout="020106150400" width="12"/>
     <field name="Version" type="constant" value="01" width="2"/>
     <field name="Filler2" type="filler" width="105"/>
    </fields>
   </header>
   <content>
    <fields>
     <field name="TypeOfRecord" type="constant" value="20" width="2"/>
     <field name="Account" type="cdrfield" value="cgrid" width="12" mandatory="true"/>
     <field name="Subject" type="cdrfield" value="subject" strip="left" padding="left" width="5"/>
     <field name="CLI" type="cdrfield" value="cli" strip="xright" width="15"/>
     <field name="Destination" type="cdrfield" value="destination" strip="xright" width="24"/>
     <field name="TOR" type="constant" value="02" width="2"/>
     <field name="SubtypeTOR" type="constant" value="11" width="4"/>
     <field name="SetupTime" type="cdrfield" value="start_time" layout="020106150400" width="12"/>
     <field name="Duration" type="cdrfield" value="duration" width="6"/>
     <field name="DataVolume" type="filler" width="6"/>
     <field name="TaxCode" type="constant" value="1" width="1"/>
     <field name="OperatorCode" type="cdrfield" value="operator" width="2"/>
     <field name="ProductId" type="cdrfield" value="productid" width="5"/>
     <field name="NetworkId" type="constant" value="3" width="1"/>
     <field name="CallId" type="cdrfield" value="accid" width="16"/>
     <field name="Filler" type="filler" width="8"/>
     <field name="Filler" type="filler" width="8"/>
     <field name="TerminationCode" type="cdrfield" value='~cost_details:s/"MatchedDestId":".+_(\s\s\s\s\s)"/$1/' width="5"/>
     <field name="Cost" type="cdrfield" value="cost" padding="zeroleft" width="9"/>
     <field name="CalledMask" type="cdrfield" value="calledmask" width="1"/>
    </fields>
   </content>
   <trailer>
    <fields>
     <field name="TypeOfRecord" type="constant" value="90" width="2"/>
     <field name="Filler1" type="filler" width="3"/>
     <field name="DistributorCode" type="constant" value="VOI" width="3"/>
     <field name="FileSeqNr" type="metatag" value="export_id" padding="zeroleft" width="5"/>
     <field name="NumberOfRecords" type="metatag" value="cdrs_number" padding="zeroleft" width="6"/>
     <field name="CdrsDuration" type="metatag" value="cdrs_duration" padding="zeroleft" width="8"/>
     <field name="FirstCdrTime" type="metatag" value="first_cdr_time" layout="020106150400" width="12"/>
     <field name="LastCdrTime" type="metatag" value="last_cdr_time" layout="020106150400" width="12"/>
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
