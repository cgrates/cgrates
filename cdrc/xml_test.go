/*
Real-time Charging System for Telecom & ISP environments
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

package cdrc

import (
	"bytes"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/ChrisTrenkamp/goxpath"
	"github.com/ChrisTrenkamp/goxpath/tree/xmltree"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cdrXmlBroadsoft = `<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE broadWorksCDR>
<broadWorksCDR version="19.0">
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183384</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210000.104</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <type>Start</type>
    </headerModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183385</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210005.247</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <serviceProvider>MBC</serviceProvider>
      <type>Normal</type>
    </headerModule>
    <basicModule>
      <userNumber>1001</userNumber>
      <groupNumber>2001</groupNumber>
      <direction>Terminating</direction>
      <asCallType>Network</asCallType>
      <callingNumber>1001</callingNumber>
      <callingPresentationIndicator>Public</callingPresentationIndicator>
      <calledNumber>+4986517174963</calledNumber>
      <startTime>20160419210005.247</startTime>
      <userTimeZone>1+020000</userTimeZone>
      <localCallId>25160047719:0</localCallId>
      <answerIndicator>Yes</answerIndicator>
      <answerTime>20160419210006.813</answerTime>
      <releaseTime>20160419210020.296</releaseTime>
      <terminationCause>016</terminationCause>
      <chargeIndicator>y</chargeIndicator>
      <releasingParty>local</releasingParty>
      <userId>1001@cgrates.org</userId>
      <clidPermitted>Yes</clidPermitted>
      <namePermitted>Yes</namePermitted>
    </basicModule>
    <centrexModule>
      <group>CGR_GROUP</group>
      <trunkGroupName>CGR_GROUP/CGR_GROUP_TRUNK30</trunkGroupName>
      <trunkGroupInfo>Normal</trunkGroupInfo>
      <locationList>
        <locationInformation>
          <location>1001@cgrates.org</location>
          <locationType>Primary Device</locationType>
        </locationInformation>
      </locationList>
      <locationUsage>31.882</locationUsage>
    </centrexModule>
    <ipModule>
      <route>gw04.cgrates.org</route>
      <networkCallID>74122796919420162305@172.16.1.2</networkCallID>
      <codec>PCMA/8000</codec>
      <accessDeviceAddress>172.16.1.4</accessDeviceAddress>
      <accessCallID>BW2300052501904161738474465@172.16.1.10</accessCallID>
      <codecUsage>31.882</codecUsage>
      <userAgent>OmniPCX Enterprise R11.0.1 k1.520.22.b</userAgent>
    </ipModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183386</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210006.909</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <serviceProvider>MBC</serviceProvider>
      <type>Normal</type>
    </headerModule>
    <basicModule>
      <userNumber>1002</userNumber>
      <groupNumber>2001</groupNumber>
      <direction>Terminating</direction>
      <asCallType>Network</asCallType>
      <callingNumber>+4986517174964</callingNumber>
      <callingPresentationIndicator>Public</callingPresentationIndicator>
      <calledNumber>1002</calledNumber>
      <startTime>20160419210006.909</startTime>
      <userTimeZone>1+020000</userTimeZone>
      <localCallId>27280048121:0</localCallId>
      <answerIndicator>Yes</answerIndicator>
      <answerTime>20160419210007.037</answerTime>
      <releaseTime>20160419210030.322</releaseTime>
      <terminationCause>016</terminationCause>
      <chargeIndicator>y</chargeIndicator>
      <releasingParty>local</releasingParty>
      <userId>314028947650@cgrates.org</userId>
      <clidPermitted>Yes</clidPermitted>
      <namePermitted>Yes</namePermitted>
    </basicModule>
    <centrexModule>
      <group>CGR_GROUP</group>
      <trunkGroupName>CGR_GROUP/CGR_GROUP_TRUNK65</trunkGroupName>
      <trunkGroupInfo>Normal</trunkGroupInfo>
      <locationList>
        <locationInformation>
          <location>31403456100@cgrates.org</location>
          <locationType>Primary Device</locationType>
        </locationInformation>
      </locationList>
      <locationUsage>26.244</locationUsage>
    </centrexModule>
    <ipModule>
      <route>gw01.cgrates.org</route>
      <networkCallID>108352493719420162306@172.31.250.150</networkCallID>
      <codec>PCMA/8000</codec>
      <accessDeviceAddress>172.16.1.4</accessDeviceAddress>
      <accessCallID>2345300069121904161716512907@172.16.1.10</accessCallID>
      <codecUsage>26.244</codecUsage>
      <userAgent>Altitude vBox</userAgent>
    </ipModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183486</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419211500.104</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <type>End</type>
    </headerModule>
  </cdrData>
</broadWorksCDR>`

func optsNotStrict(s *xmltree.ParseOptions) {
	s.Strict = false
}

func TestXMLElementText(t *testing.T) {
	xp := goxpath.MustParse(path.Join("/broadWorksCDR/cdrData/"))
	xmlTree := xmltree.MustParseXML(bytes.NewBufferString(cdrXmlBroadsoft), optsNotStrict)
	cdrs := goxpath.MustExec(xp, xmlTree, nil)
	cdrWithoutUserNr := cdrs[0]
	if _, err := elementText(cdrWithoutUserNr, "cdrData/basicModule/userNumber"); err != utils.ErrNotFound {
		t.Error(err)
	}
	cdrWithUser := cdrs[1]
	if val, err := elementText(cdrWithUser, "cdrData/basicModule/userNumber"); err != nil {
		t.Error(err)
	} else if val != "1001" {
		t.Errorf("Expecting: 1001, received: %s", val)
	}
	if val, err := elementText(cdrWithUser, "/cdrData/centrexModule/locationList/locationInformation/locationType"); err != nil {
		t.Error(err)
	} else if val != "Primary Device" {
		t.Errorf("Expecting: <Primary Device>, received: <%s>", val)
	}
}

func TestXMLHandlerSubstractUsage(t *testing.T) {
	xp := goxpath.MustParse(path.Join("/broadWorksCDR/cdrData/"))
	xmlTree := xmltree.MustParseXML(bytes.NewBufferString(cdrXmlBroadsoft), optsNotStrict)
	cdrs := goxpath.MustExec(xp, xmlTree, nil)
	cdrWithUsage := cdrs[1]
	if usage, err := handlerSubstractUsage(cdrWithUsage, utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>releaseTime;^|;broadWorksCDR>cdrData>basicModule>answerTime", utils.INFIELD_SEP),
		utils.HierarchyPath([]string{"broadWorksCDR", "cdrData"}), "UTC"); err != nil {
		t.Error(err)
	} else if usage != time.Duration(13483000000) {
		t.Errorf("Expected: 13.483s, received: %v", usage)
	}
}

func TestXMLRPProcess(t *testing.T) {
	cdrcCfgs := []*config.CdrcConfig{
		&config.CdrcConfig{
			ID:                      "TestXML",
			Enabled:                 true,
			CdrFormat:               "xml",
			DataUsageMultiplyFactor: 1024,
			CDRPath:                 utils.HierarchyPath([]string{"broadWorksCDR", "cdrData"}),
			CdrSourceId:             "TestXML",
			CdrFilter:               utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>headerModule>type(Normal)", utils.INFIELD_SEP),
			ContentFields: []*config.CfgCdrField{
				&config.CfgCdrField{Tag: "TOR", Type: utils.META_COMPOSED, FieldId: utils.TOR,
					Value: utils.ParseRSRFieldsMustCompile("^*voice", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "OriginID", Type: utils.META_COMPOSED, FieldId: utils.ACCID,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>localCallId", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "RequestType", Type: utils.META_COMPOSED, FieldId: utils.REQTYPE,
					Value: utils.ParseRSRFieldsMustCompile("^*rated", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Direction", Type: utils.META_COMPOSED, FieldId: utils.DIRECTION,
					Value: utils.ParseRSRFieldsMustCompile("^*out", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Tenant", Type: utils.META_COMPOSED, FieldId: utils.TENANT,
					Value: utils.ParseRSRFieldsMustCompile("~broadWorksCDR>cdrData>basicModule>userId:s/.*@(.*)/${1}/", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Category", Type: utils.META_COMPOSED, FieldId: utils.CATEGORY,
					Value: utils.ParseRSRFieldsMustCompile("^call", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Account", Type: utils.META_COMPOSED, FieldId: utils.ACCOUNT,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>userNumber", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Destination", Type: utils.META_COMPOSED, FieldId: utils.DESTINATION,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>calledNumber", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "SetupTime", Type: utils.META_COMPOSED, FieldId: utils.SETUP_TIME,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>startTime", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "AnswerTime", Type: utils.META_COMPOSED, FieldId: utils.ANSWER_TIME,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>answerTime", utils.INFIELD_SEP), Mandatory: true},
				&config.CfgCdrField{Tag: "Usage", Type: utils.META_HANDLER, FieldId: utils.USAGE, HandlerId: utils.HandlerSubstractUsage,
					Value: utils.ParseRSRFieldsMustCompile("broadWorksCDR>cdrData>basicModule>releaseTime;^|;broadWorksCDR>cdrData>basicModule>answerTime", utils.INFIELD_SEP), Mandatory: true},
			},
		},
	}
	xmlRP, err := NewXMLRecordsProcessor(bytes.NewBufferString(cdrXmlBroadsoft), utils.HierarchyPath([]string{"broadWorksCDR", "cdrData"}), "UTC", true, cdrcCfgs)
	if err != nil {
		t.Error(err)
	}
	var cdrs []*engine.CDR
	for i := 0; i < 4; i++ {
		cdrs, err = xmlRP.ProcessNextRecord()
		if i == 1 { // Take second CDR since the first one cannot be processed
			break
		}
	}
	if err != nil {
		t.Error(err)
	}
	expectedCDRs := []*engine.CDR{
		&engine.CDR{CGRID: "1f045359a0784d15e051d7e41ae30132b139d714", OriginHost: "0.0.0.0", Source: "TestXML", OriginID: "25160047719:0",
			ToR: "*voice", RequestType: "*rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Destination: "+4986517174963",
			SetupTime: time.Date(2016, 4, 19, 21, 0, 5, 247000000, time.UTC), AnswerTime: time.Date(2016, 4, 19, 21, 0, 6, 813000000, time.UTC), Usage: time.Duration(13483000000),
			ExtraFields: map[string]string{}, Cost: -1},
	}
	if !reflect.DeepEqual(expectedCDRs, cdrs) {
		t.Errorf("Expecting: %+v\n, received: %+v\n", expectedCDRs, cdrs)
	}
}
