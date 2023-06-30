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

package config

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/antchfx/xmlquery"
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

func TestXMLElementText(t *testing.T) {
	doc, err := xmlquery.Parse(strings.NewReader(cdrXmlBroadsoft))
	if err != nil {
		t.Error(err)
	}
	cdrs := xmlquery.Find(doc, path.Join("/broadWorksCDR/cdrData/"))
	cdrWithoutUserNr := cdrs[0]
	if _, err := ElementText(cdrWithoutUserNr, "basicModule/userNumber"); err != utils.ErrNotFound {
		t.Error(err)
	}
	cdrWithUser := cdrs[1]
	if val, err := ElementText(cdrWithUser, "basicModule/userNumber"); err != nil {
		t.Error(err)
	} else if val != "1001" {
		t.Errorf("Expecting: 1001, received: %s", val)
	}
	if val, err := ElementText(cdrWithUser, "centrexModule/locationList/locationInformation/locationType"); err != nil {
		t.Error(err)
	} else if val != "Primary Device" {
		t.Errorf("Expecting: <Primary Device>, received: <%s>", val)
	}
}

var xmlContent = `<?xml version="1.0" encoding="UTF-8"?>
<File xmlns="http://www.metaswitch.com/cfs/billing/V1.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" compatibility="9">
	<FileHeader seqnum="169">
		<EquipmentType>Metaswitch CFS</EquipmentType>
		<EquipmentId></EquipmentId>
		<CreateTime>1510225200002</CreateTime>
	</FileHeader>
	<CDRs>
		<Call seqnum="0000000001" error="no" longcall="false" testcall="false" class="0" operator="false" correlator="397828983391" connected="false">
			<CallType>National</CallType>
			<Features/>
			<ReleasingParty>Orig</ReleasingParty>
			<ReleaseReason type="q850" loc="u">19</ReleaseReason>
			<ReleaseReason type="sip">480</ReleaseReason>
			<ReleaseReason type="internal">No answer</ReleaseReason>
			<InternalIndex>223007622</InternalIndex>
			<OrigParty xsi:type="BusinessLinePartyType" subscribergroup="Subscribers in Guernsey, NJ" billingtype="flat rate" privacy="false" cpc="normal" ani-ii="00">
				<SubscriberAddr type="e164">+27110493421</SubscriberAddr>
				<CallingPartyAddr type="e164">+27110493421</CallingPartyAddr>
				<ChargeAddr type="e164">+27110493421</ChargeAddr>
				<BusinessGroupName>Ro_test</BusinessGroupName>
				<SIPCallId>0gQAAC8WAAACBAAALxYAAD57SAEV7ekv/OSKkO7qmD82OmbfHO+Z7wIZJkXdCv8z@10.170.248.200</SIPCallId>
			</OrigParty>
			<TermParty xsi:type="NetworkTrunkPartyType">
				<TrunkGroup type="sip" trunkname="IMS Core">
					<TrunkGroupId>1</TrunkGroupId>
					<TrunkMemberId>0</TrunkMemberId>
				</TrunkGroup>
				<SIPCallId>8824071D@10.170.248.140</SIPCallId>
				<Reason type="q850">19</Reason>
				<Reason type="sip">480</Reason>
			</TermParty>
			<RoutingInfo>
				<RequestedAddr type="unknown">0763371551</RequestedAddr>
				<DestAddr type="e164">+270763371551</DestAddr>
				<RoutedAddr type="national">0763371551</RoutedAddr>
				<CallingPartyRoutedAddr type="national">110493421</CallingPartyRoutedAddr>
				<CallingPartyOrigAddr type="national">110493421</CallingPartyOrigAddr>
			</RoutingInfo>
			<CarrierSelectInfo>
				<CarrierOperatorInvolved>False</CarrierOperatorInvolved>
				<SelectionMethod>NetworkDefault</SelectionMethod>
				<NetworkId>0</NetworkId>
			</CarrierSelectInfo>
			<SignalingInfo>
				<MediaCapabilityRequested>Speech</MediaCapabilityRequested>
				<PChargingVector>
					<icidvalue>13442698e525ad2c21251f76479ab2b4</icidvalue>
					<origioi>voice.lab.liquidtelecom.net</origioi>
				</PChargingVector>
			</SignalingInfo>
			<IcSeizeTime>1510225513055</IcSeizeTime>
			<OgSeizeTime>1510225513304</OgSeizeTime>
			<RingingTime>1510225514836</RingingTime>
			<ConnectTime/>
			<DisconnectTime>1510225516981</DisconnectTime>
			<ReleaseTime>1510225516981</ReleaseTime>
			<CompleteTime>1510225516981</CompleteTime>
		</Call>
		<Call seqnum="0000000002" error="no" longcall="false" testcall="false" class="0" operator="false" correlator="402123969565" connected="true">
			<CallType>Premium</CallType>
			<Features/>
			<ReleasingParty>Orig</ReleasingParty>
			<ReleaseReason type="q850" loc="u">16</ReleaseReason>
			<ReleaseReason type="internal">No error</ReleaseReason>
			<InternalIndex>223007623</InternalIndex>
			<OrigParty xsi:type="BusinessLinePartyType" subscribergroup="Subscribers in Guernsey, NJ" billingtype="flat rate" privacy="false" cpc="normal" ani-ii="00">
				<SubscriberAddr type="e164">+27110493421</SubscriberAddr>
				<CallingPartyAddr type="e164">+27110493421</CallingPartyAddr>
				<ChargeAddr type="e164">+27110493421</ChargeAddr>
				<BusinessGroupName>Ro_test</BusinessGroupName>
				<SIPCallId>0gQAAC8WAAACBAAALxYAAPkyWDO29Do1SyxNi5UV71mJYEIEkfNa9wCFCCjY2asU@10.170.248.200</SIPCallId>
			</OrigParty>
			<TermParty xsi:type="NetworkTrunkPartyType">
				<TrunkGroup type="sip" trunkname="IMS Core">
					<TrunkGroupId>1</TrunkGroupId>
					<TrunkMemberId>0</TrunkMemberId>
				</TrunkGroup>
				<SIPCallId>8E450FA1@10.170.248.140</SIPCallId>
			</TermParty>
			<RoutingInfo>
				<RequestedAddr type="unknown">0843073451</RequestedAddr>
				<DestAddr type="e164">+270843073451</DestAddr>
				<RoutedAddr type="national">0843073451</RoutedAddr>
				<CallingPartyRoutedAddr type="national">110493421</CallingPartyRoutedAddr>
				<CallingPartyOrigAddr type="national">110493421</CallingPartyOrigAddr>
			</RoutingInfo>
			<CarrierSelectInfo>
				<CarrierOperatorInvolved>False</CarrierOperatorInvolved>
				<SelectionMethod>NetworkDefault</SelectionMethod>
				<NetworkId>0</NetworkId>
			</CarrierSelectInfo>
			<SignalingInfo>
				<MediaCapabilityRequested>Speech</MediaCapabilityRequested>
				<PChargingVector>
					<icidvalue>46d7974398c2671016afccc3f2c428c7</icidvalue>
					<origioi>voice.lab.liquidtelecom.net</origioi>
				</PChargingVector>
			</SignalingInfo>
			<IcSeizeTime>1510225531933</IcSeizeTime>
			<OgSeizeTime>1510225532183</OgSeizeTime>
			<RingingTime>1510225534973</RingingTime>
			<ConnectTime>1510225539364</ConnectTime>
			<DisconnectTime>1510225593101</DisconnectTime>
			<ReleaseTime>1510225593101</ReleaseTime>
			<CompleteTime>1510225593101</CompleteTime>
		</Call>
		<Call seqnum="0000000003" error="no" longcall="false" testcall="false" class="0" operator="false" correlator="406419270822" connected="true">
			<CallType>International</CallType>
			<Features/>
			<ReleasingParty>Orig</ReleasingParty>
			<ReleaseReason type="q850" loc="u">16</ReleaseReason>
			<ReleaseReason type="internal">No error</ReleaseReason>
			<InternalIndex>223007624</InternalIndex>
			<OrigParty xsi:type="BusinessLinePartyType" subscribergroup="Subscribers in Guernsey, NJ" billingtype="flat rate" privacy="false" cpc="normal" ani-ii="00">
				<SubscriberAddr type="e164">+27110493421</SubscriberAddr>
				<CallingPartyAddr type="e164">+27110493421</CallingPartyAddr>
				<ChargeAddr type="e164">+27110493421</ChargeAddr>
				<BusinessGroupName>Ro_test</BusinessGroupName>
				<SIPCallId>0gQAAC8WAAACBAAALxYAAJrUscTicyU5GtjPyQnpAeuNmz9p/bdOoR/Mk9RXciOI@10.170.248.200</SIPCallId>
			</OrigParty>
			<TermParty xsi:type="NetworkTrunkPartyType">
				<TrunkGroup type="sip" trunkname="IMS Core">
					<TrunkGroupId>1</TrunkGroupId>
					<TrunkMemberId>0</TrunkMemberId>
				</TrunkGroup>
				<SIPCallId>BC8B2801@10.170.248.140</SIPCallId>
			</TermParty>
			<RoutingInfo>
				<RequestedAddr type="unknown">263772822306</RequestedAddr>
				<DestAddr type="e164">+263772822306</DestAddr>
				<RoutedAddr type="e164">263772822306</RoutedAddr>
				<CallingPartyRoutedAddr type="national">110493421</CallingPartyRoutedAddr>
				<CallingPartyOrigAddr type="national">110493421</CallingPartyOrigAddr>
			</RoutingInfo>
			<CarrierSelectInfo>
				<CarrierOperatorInvolved>False</CarrierOperatorInvolved>
				<SelectionMethod>NetworkDefault</SelectionMethod>
				<NetworkId>0</NetworkId>
			</CarrierSelectInfo>
			<SignalingInfo>
				<MediaCapabilityRequested>Speech</MediaCapabilityRequested>
				<PChargingVector>
					<icidvalue>750b8b73e41ba7b59b21240758522268</icidvalue>
					<origioi>voice.lab.liquidtelecom.net</origioi>
				</PChargingVector>
			</SignalingInfo>
			<IcSeizeTime>1510225865894</IcSeizeTime>
			<OgSeizeTime>1510225866144</OgSeizeTime>
			<RingingTime>1510225866756</RingingTime>
			<ConnectTime>1510225876243</ConnectTime>
			<DisconnectTime>1510225916144</DisconnectTime>
			<ReleaseTime>1510225916144</ReleaseTime>
			<CompleteTime>1510225916144</CompleteTime>
		</Call>
	</CDRs>
	<FileFooter>
		<LastModTime>1510227591467</LastModTime>
		<NumCDRs>3</NumCDRs>
		<DataErroredCDRs>0</DataErroredCDRs>
		<WriteErroredCDRs>0</WriteErroredCDRs>
	</FileFooter>
</File>
`

func TestXMLElementText3(t *testing.T) {
	doc, err := xmlquery.Parse(strings.NewReader(xmlContent))
	if err != nil {
		t.Error(err)
	}
	hPath2 := utils.ParseHierarchyPath("File.CDRs.Call", "")
	cdrs := xmlquery.Find(doc, hPath2.AsString("/", true))
	if len(cdrs) != 3 {
		t.Errorf("Expecting: 3, received: %+v", len(cdrs))
	}
	if _, err := ElementText(cdrs[0], "SignalingInfo/PChargingVector/test"); err != utils.ErrNotFound {
		t.Error(err)
	}

	if val, err := ElementText(cdrs[1], "SignalingInfo/PChargingVector/icidvalue"); err != nil {
		t.Error(err)
	} else if val != "46d7974398c2671016afccc3f2c428c7" {
		t.Errorf("Expecting: 46d7974398c2671016afccc3f2c428c7, received: %s", val)
	}
}

var xmlMultipleIndex = `<complete-success-notification callid="109870">
	<createtime>2005-08-26T14:16:42</createtime>
	<connecttime>2005-08-26T14:16:56</connecttime>
	<endtime>2005-08-26T14:17:34</endtime>
	<reference>My Call Reference</reference>
	<userid>386</userid>
	<username>sampleusername</username>
	<customerid>1</customerid>
	<companyname>Conecto LLC</companyname>
	<totalcost amount="0.21" currency="USD">US$0.21</totalcost>
	<hasrecording>yes</hasrecording>
	<hasvoicemail>no</hasvoicemail>
	<agenttotalcost amount="0.13" currency="USD">US$0.13</agenttotalcost>
	<agentid>44</agentid>
	<callleg calllegid="222146">
		<number>+441624828505</number>
		<description>Isle of Man</description>
		<seconds>38</seconds>
		<perminuterate amount="0.0200" currency="USD">US$0.0200</perminuterate>
		<cost amount="0.0140" currency="USD">US$0.0140</cost>
		<agentperminuterate amount="0.0130" currency="USD">US$0.0130</agentperminuterate>
		<agentcost amount="0.0082" currency="USD">US$0.0082</agentcost>
	</callleg>
	<callleg calllegid="222147">
		<number>+44 7624 494075</number>
		<description>Isle of Man</description>
		<seconds>37</seconds>
		<perminuterate amount="0.2700" currency="USD">US$0.2700</perminuterate>
		<cost amount="0.1890" currency="USD">US$0.1890</cost>
		<agentperminuterate amount="0.1880" currency="USD">US$0.1880</agentperminuterate>
		<agentcost amount="0.1159" currency="USD">US$0.1159</agentcost>
	</callleg>
</complete-success-notification>
`

func TestXMLIndexes(t *testing.T) {
	doc, err := xmlquery.Parse(strings.NewReader(xmlMultipleIndex))
	if err != nil {
		t.Error(err)
	}
	dP := NewXmlProvider(doc, utils.HierarchyPath([]string{}))
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "userid"}); err != nil {
		t.Error(err)
	} else if data != "386" {
		t.Errorf("expecting: 386, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "username"}); err != nil {
		t.Error(err)
	} else if data != "sampleusername" {
		t.Errorf("expecting: sampleusername, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "38" {
		t.Errorf("expecting: 38, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg[1]", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "37" {
		t.Errorf("expecting: 37, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg[@calllegid='222147']", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "37" {
		t.Errorf("expecting: 37, received: <%s>", data)
	}
}

func TestXmlProviderString(t *testing.T) {
	x := XmlProvider{}

	rcv := x.String()
	exp := utils.ToJSON(x)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}

func TestXmlProviderRemoteHost(t *testing.T) {
	x := XmlProvider{}

	rcv := x.RemoteHost()
	exp := utils.LocalAddr()

	if rcv.String() != exp.String() {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}

func TestXmlProviderFieldAsInterface(t *testing.T) {

	x := XmlProvider{
		cache: utils.MapStorage{"test": "val1"},
	}

	type exp struct {
		data any
		err  error
	}

	tests := []struct {
		name string
		arg  []string
		exp  exp
	}{
		{
			name: "check error not found",
			arg:  []string{},
			exp:  exp{data: nil, err: utils.ErrNotFound},
		},
		{
			name: "item found in cache",
			arg:  []string{"test"},
			exp:  exp{data: "val1", err: nil},
		},
		{
			name: "check error filter rule needs to end in ]",
			arg:  []string{"test[0"},
			exp:  exp{data: nil, err: fmt.Errorf("filter rule <[0> needs to end in ]")},
		},
		{
			name: "check strconv.Atoi error",
			arg:  []string{"test[a]"},
			exp:  exp{data: nil, err: fmt.Errorf(`strconv.Atoi: parsing "a": invalid syntax`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv, err := x.FieldAsInterface(tt.arg)

			if err != nil {
				if err.Error() != tt.exp.err.Error() {
					t.Fatalf("recived %s, expected %s", err, tt.exp.err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp.data) {
				t.Errorf("recived %s, expected %s", rcv, tt.exp.data)
			}
		})
	}
}

func TestXmlProviderFieldAsString(t *testing.T) {
	x := XmlProvider{}

	_, err := x.FieldAsString([]string{})
	exp := utils.ErrNotFound

	if err.Error() != exp.Error() {
		t.Fatalf("recived %d, expected %s", err, exp)
	}
}
