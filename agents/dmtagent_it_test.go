/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be u297seful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package agents

import (
	"flag"
	//"net"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var interations = flag.Int("iterations", 1, "Number of iterations to do for dry run simulation")
var replyTimeout = flag.String("reply_timeout", "1s", "Maximum duration to wait for a reply")

var daCfgPath string
var daCfg *config.CGRConfig
var apierRpc *rpc.Client
var dmtClient *DiameterClient
var err error
var rplyTimeout time.Duration

func TestDmtAgentInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	daCfgPath = path.Join(*dataDir, "conf", "samples", "dmtagent")
	// Init config first
	var err error
	daCfg, err = config.NewCGRConfigFromFolder(daCfgPath)
	if err != nil {
		t.Error(err)
	}
	daCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(daCfg)
	rplyTimeout, _ = utils.ParseDurationWithSecs(*replyTimeout)
}

// Remove data in both rating and accounting db
func TestDmtAgentResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestDmtAgentResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestDmtAgentStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(daCfgPath, 4000); err != nil {
		t.Fatal(err)
	}
}

func TestDmtAgentCCRAsSMGenericEvent(t *testing.T) {
	if !*testIntegration {
		return
	}
	cfgDefaults, _ := config.NewDefaultCGRConfig()
	loadDictionaries(cfgDefaults.DiameterAgentCfg().DictionariesDir, "UNIT_TEST")
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	ccr := &CCR{
		SessionId:         "routinga;1442095190;1476802709",
		OriginHost:        cfgDefaults.DiameterAgentCfg().OriginHost,
		OriginRealm:       cfgDefaults.DiameterAgentCfg().OriginRealm,
		DestinationHost:   cfgDefaults.DiameterAgentCfg().OriginHost,
		DestinationRealm:  cfgDefaults.DiameterAgentCfg().OriginRealm,
		AuthApplicationId: 4,
		ServiceContextId:  "voice@huawei.com",
		CCRequestType:     1,
		CCRequestNumber:   0,
		EventTimestamp:    time.Date(2015, 11, 23, 12, 22, 24, 0, time.UTC),
		ServiceIdentifier: 0,
		SubscriptionId: []struct {
			SubscriptionIdType int    `avp:"Subscription-Id-Type"`
			SubscriptionIdData string `avp:"Subscription-Id-Data"`
		}{
			struct {
				SubscriptionIdType int    `avp:"Subscription-Id-Type"`
				SubscriptionIdData string `avp:"Subscription-Id-Data"`
			}{SubscriptionIdType: 0, SubscriptionIdData: "4986517174963"},
			struct {
				SubscriptionIdType int    `avp:"Subscription-Id-Type"`
				SubscriptionIdData string `avp:"Subscription-Id-Data"`
			}{SubscriptionIdType: 0, SubscriptionIdData: "4986517174963"}},
		debitInterval: time.Duration(300) * time.Second,
	}
	ccr.RequestedServiceUnit.CCTime = 300
	ccr.UsedServiceUnit.CCTime = 0
	ccr.ServiceInformation.INInformation.CallingPartyAddress = "4986517174963"
	ccr.ServiceInformation.INInformation.CalledPartyAddress = "4986517174964"
	ccr.ServiceInformation.INInformation.RealCalledNumber = "4986517174964"
	ccr.ServiceInformation.INInformation.ChargeFlowType = 0
	ccr.ServiceInformation.INInformation.CallingVlrNumber = "49123956767"
	ccr.ServiceInformation.INInformation.CallingCellIDOrSAI = "12340185301425"
	ccr.ServiceInformation.INInformation.BearerCapability = "capable"
	ccr.ServiceInformation.INInformation.CallReferenceNumber = "askjadkfjsdf"
	ccr.ServiceInformation.INInformation.MSCAddress = "123324234"
	ccr.ServiceInformation.INInformation.TimeZone = 0
	ccr.ServiceInformation.INInformation.CalledPartyNP = "4986517174964"
	ccr.ServiceInformation.INInformation.SSPTime = "20091020120101"
	var err error
	if ccr.diamMessage, err = ccr.AsDiameterMessage(); err != nil {
		t.Error(err)
	}
	eSMGE := sessionmanager.SMGenericEvent{"EventName": DIAMETER_CCR, "OriginID": "routinga;1442095190;1476802709",
		"Account": "*users", "AnswerTime": "2015-11-23 12:22:24 +0000 UTC", "Category": "call",
		"Destination": "4986517174964", "Direction": "*out", "RequestType": "*users", "SetupTime": "2015-11-23 12:22:24 +0000 UTC",
		"Subject": "*users", "SubscriberId": "4986517174963", "ToR": "*voice", "Tenant": "*users", "Usage": "300"}
	if smge, err := ccr.AsSMGenericEvent(cfgDefaults.DiameterAgentCfg().RequestProcessors[0].CCRFields); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGE, smge) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGE, smge)
	}
}

func TestDmtAgentPopulateCCTotalOctets(t *testing.T) {
	if !*testIntegration {
		return
	}
	daRP := &config.DARequestProcessor{CCAFields: []*config.CfgCdrField{
		&config.CfgCdrField{Tag: "GrantedUnit", FieldFilter: utils.ParseRSRFieldsMustCompile("CGRError(^$)", utils.INFIELD_SEP),
			FieldId: "Multiple-Services-Credit-Control>Granted-Service-Unit>CC-Time", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("CGRMaxUsage", utils.INFIELD_SEP), Mandatory: true},
		&config.CfgCdrField{Tag: "GrantedOctet", FieldFilter: utils.ParseRSRFieldsMustCompile("CGRError(^$)", utils.INFIELD_SEP),
			FieldId: "Multiple-Services-Credit-Control>Granted-Service-Unit>CC-Total-Octets", Type: utils.META_COMPOSED, Value: utils.ParseRSRFieldsMustCompile("CGRMaxUsage", utils.INFIELD_SEP), Mandatory: true},
	}}
	ccr := new(CCR)
	ccr.diamMessage = ccr.AsBareDiameterMessage()
	cca := NewBareCCAFromCCR(ccr, "cgr-da", "cgrates.org")
	if err := cca.SetProcessorAVPs(daRP, map[string]string{CGRError: "", CGRMaxUsage: "153600"}); err != nil {
		t.Error(err)
	}
	if avps, err := cca.diamMessage.FindAVPsWithPath([]interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Not found")
	} else if strResult := avpValAsString(avps[0]); strResult != "153600" { // Result-Code set in the template
		t.Errorf("Expecting 153600, received: %s", strResult)
	}
	if avps, err := cca.diamMessage.FindAVPsWithPath([]interface{}{"Multiple-Services-Credit-Control", "Granted-Service-Unit", "CC-Total-Octets"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Not found")
	} else if strResult := avpValAsString(avps[0]); strResult != "153600" { // Result-Code set in the template
		t.Errorf("Expecting 153600, received: %s", strResult)
	}
}

// Connect rpc client to rater
func TestDmtAgentApierRpcConn(t *testing.T) {
	if !*testIntegration {
		return
	}
	var err error
	apierRpc, err = jsonrpc.Dial("tcp", daCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestDmtAgentTPFromFolder(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := apierRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestConnectDiameterClient(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:47:26Z"'
func TestDmtAgentSendCCRInit(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("testccr1", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "testccr1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(0), PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DebitInterval, false)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	fmt.Printf("Received msg: %+v", msg)
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s", strCCTime)
	}
	if result, err := msg.FindAVP("Result-Code", dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if resultStr := avpValAsString(result); resultStr != "2001" {
		t.Errorf("Expecting 2001, received: %s", resultStr)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.5008
	if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:52:26Z"'
func TestDmtAgentSendCCRUpdate(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("testccr1", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "testccr1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(300) * time.Second, PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DebitInterval, false)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s, (%+v)", strCCTime, avps)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.251800
	if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:57:26Z"'
func TestDmtAgentSendCCRUpdate2(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("testccr1", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "testccr1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(600) * time.Second, PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DebitInterval, false)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s", strCCTime)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.002800
	if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if utils.Round(acnt.BalanceMap[utils.MONETARY].GetTotalValue(), 5, utils.ROUNDING_MIDDLE) != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestDmtAgentSendCCRTerminate(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("testccr1", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "testccr1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(610) * time.Second, PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DebitInterval, true)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No answer to CCR terminate received")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "0" {
		t.Errorf("Expecting 0, received: %s", strCCTime)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.243500
	if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal { // Should also consider derived charges which double the cost of 6m10s - 2x0.7584
		t.Errorf("Expected: %v, received: %v", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestDmtAgentSendCCRSMS(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testccr2"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1002")), // Address-Data
						}}),
				},
			}),
		}})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
	dmtClient.ReceivedMessage(rplyTimeout) // Discard the received message so we can test next one
	/*
		if msg == nil {
			t.Fatal("No message returned")
		}
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Granted-Service-Unit not found")
		} else if strCCTime := avpValAsString(avps[0]); strCCTime != "0" {
			t.Errorf("Expecting 0, received: %s", strCCTime)
		}
		var acnt *engine.Account
		attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
		eAcntVal := 9.205
		if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
			t.Error(err)
		} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal { // Should also consider derived charges which double the cost of 6m10s - 2x0.7584
			t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
		}
	*/
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, ToRs: []string{utils.SMS}}
	if err := apierRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "1" {
			t.Errorf("Unexpected CDR Usage received, cdr: %+v ", cdrs[0])
		}
		if cdrs[0].Cost != 0.6 {
			t.Errorf("Unexpected CDR Cost received, cdr: %+v ", cdrs[0])
		}
	}
}

func TestDmtAgentSendCCRSMSWrongAccount(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testccr3"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("non_existent")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),             // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("49602200011")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),             // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("49780029555")), // Address-Data
						}}),
				},
			}),
		}})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout) // Discard the received message so we can test next one
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code")
	} else if strResult := avpValAsString(avps[0]); strResult != "5030" { // Result-Code set in the template
		t.Errorf("Expecting 5030, received: %s", strResult)
	}
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:47:26Z"'
func TestDmtAgentSendCCRInitWrongAccount(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("testccr4", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "testccr4", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "non_existent", Subject: "non_existent", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(0) * time.Second, PDD: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DebitInterval, false)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout) // Discard the received message so we can test next one
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code")
	} else if strResult := avpValAsString(avps[0]); strResult != "5030" { // Result-Code set in the template
		t.Errorf("Expecting 5030, received: %s", strResult)
	}
}

func TestDmtAgentSendCCRSimpaEvent(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testccr5"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("routing1.huawei.com"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("simpa@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 13, 16, 47, 58, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCMoney, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.UnitValue, avp.Mbit, 0, &diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(10000)),
							diam.NewAVP(avp.Exponent, avp.Mbit, 0, datatype.Integer32(-5)),
						},
					}),
					diam.NewAVP(avp.CurrencyCode, avp.Mbit, 0, datatype.Unsigned32(33)),
				},
			}),
		},
	})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ // Service-Information
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(29000, avp.Mbit, 2011, &diam.GroupedAVP{ // MC-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20938, avp.Mbit, 2011, datatype.OctetString("0x38924012914528")), // HighLayerCharacteristics
					diam.NewAVP(29002, avp.Mbit, 2011, datatype.UTF8String("12928471313847173")), // MC-Service-Id
					diam.NewAVP(29003, avp.Mbit, 2011, datatype.UTF8String("SPV123456012123")),   // TransparentData
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // MC-Information
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(0)),             // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("33780029555")), // Address-Data
						},
					}),
				},
			}),
		}})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout) // Discard the received message so we can test next one
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code")
	} else if strResult := avpValAsString(avps[0]); strResult != "2001" { // Result-Code set in the template
		t.Errorf("Expecting 2001, received: %s", strResult)
	}
}

func TestDmtAgentCdrs(t *testing.T) {
	if !*testIntegration {
		return
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, ToRs: []string{utils.VOICE}}
	if err := apierRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "610" {
			t.Errorf("Unexpected CDR Usage received, cdr: %+v ", cdrs[0])
		}
		if cdrs[0].Cost != 0.7565 {
			t.Errorf("Unexpected CDR Cost received, cdr: %+v ", cdrs[0])
		}
	}
}

func TestDmtAgentSendDataGrpInit(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testdatagrp"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("gprs@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1002")), // Address-Data
						}}),
				},
			}),
		}})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code not found")
	} else if resCode := avpValAsString(avps[0]); resCode != "2001" {
		t.Errorf("Expecting 2001, received: %s", resCode)
	}
}

func TestDmtAgentSendDataGrpUpdate(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testdatagrp"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("gprs@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1002")), // Address-Data
						}}),
				},
			}),
		}})
	ccr.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(446, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(0)),    // Tariff-Change-Usage
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(20)),   // CC-Time
					diam.NewAVP(412, avp.Mbit, 0, datatype.Unsigned64(1000)), // CC-Input-Octets
					diam.NewAVP(414, avp.Mbit, 0, datatype.Unsigned64(24)),   // CC-Output-Octets
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(1)), // Data session for group 1
		},
	})
	ccr.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(446, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(0)),    // Tariff-Change-Usage
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(20)),   // CC-Time
					diam.NewAVP(412, avp.Mbit, 0, datatype.Unsigned64(1024)), // CC-Input-Octets
					diam.NewAVP(414, avp.Mbit, 0, datatype.Unsigned64(512)),  // CC-Output-Octets
				},
			}),
			diam.NewAVP(432, avp.Mbit, 0, datatype.Unsigned32(2)), // Data session for group 2
		},
	})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code not found")
	} else if resCode := avpValAsString(avps[0]); resCode != "2001" {
		t.Errorf("Expecting 2001, received: %s", resCode)
	}
}

func TestDmtAgentSendDataGrpTerminate(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("testdatagrp"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("gprs@huawei.com"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(3))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(0)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Enumerated(1)),
			diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("104502200011")), // Subscription-Id-Data
		}})
	ccr.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Enumerated(0))
	ccr.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(avp.CCTime, avp.Mbit, 0, datatype.Unsigned32(1))}})
	ccr.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{ //
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("22509")), // Calling-Vlr-Number
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("4002")),  // Called-Party-NP
				},
			}),
			diam.NewAVP(2000, avp.Mbit, 10415, &diam.GroupedAVP{ // SMS-Information
				AVP: []*diam.AVP{
					diam.NewAVP(886, avp.Mbit, 10415, &diam.GroupedAVP{ // Originator-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1002")), // Address-Data
						}}),
				},
			}),
		}})
	ccr.NewAVP("Multiple-Services-Credit-Control", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(446, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(452, avp.Mbit, 0, datatype.Enumerated(0)),   // Tariff-Change-Usage
					diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(20)),  // CC-Time
					diam.NewAVP(412, avp.Mbit, 0, datatype.Unsigned64(512)), // CC-Input-Octets
					diam.NewAVP(414, avp.Mbit, 0, datatype.Unsigned64(0)),   // CC-Output-Octets
				},
			}),
		},
	})
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(3000) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code not found")
	} else if resCode := avpValAsString(avps[0]); resCode != "2001" {
		t.Errorf("Expecting 2001, received: %s", resCode)
	}
}

func TestDmtAgentSendDataGrpCDRs(t *testing.T) {
	if !*testIntegration {
		return
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{CGRIDs: []string{utils.Sha1("testdatagrp")}}
	if err := apierRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

/*
func TestDmtAgentDryRun1(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("cgrates;1451911932;00082"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("pubsub1")) // Match specific DryRun profile
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	if _, err := ccr.NewAVP("Framed-IP-Address", avp.Mbit, 0, datatype.UTF8String("10.228.16.4")); err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(ccr); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Result-Code")
	} else if strResult := avpValAsString(avps[0]); strResult != "300" { // Result-Code set in the template
		t.Errorf("Expecting 300, received: %s", strResult)
	}
}
*/

func TestDmtAgentDryRun1(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("cgrates;1451911932;00082"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("CGR-DA"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("pubsub1")) // Match specific DryRun profile
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2016, 1, 5, 11, 30, 10, 0, time.UTC)))
	ccr.NewAVP(avp.TerminationCause, avp.Mbit, 0, datatype.Enumerated(1))
	ccr.NewAVP(443, avp.Mbit, 0, &diam.GroupedAVP{ // Subscription-Id
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1001")), // Subscription-Id-Data
		}})
	ccr.NewAVP(443, avp.Mbit, 0, &diam.GroupedAVP{ // Subscription-Id
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(1)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208123456789")), // Subscription-Id-Data
		}})
	ccr.NewAVP(439, avp.Mbit, 0, datatype.Unsigned32(0)) // Service-Identifier
	ccr.NewAVP(437, avp.Mbit, 0, &diam.GroupedAVP{       // Requested-Service-Unit
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300)), // CC-Time
		}})
	ccr.NewAVP(873, avp.Mbit|avp.Vbit, 10415, &diam.GroupedAVP{ // Service-information
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(20336, avp.Mbit, 2011, datatype.UTF8String("1001")),             // CallingPartyAdress
					diam.NewAVP(20337, avp.Mbit, 2011, datatype.UTF8String("1002")),             // CalledPartyAdress
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                  // ChargeFlowType
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("33609004940")),      // CallingVlrNumber
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("208104941749984")),  // CallingCellID
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("0x8090a3")),        // BearerCapability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.OctetString("0x401c4132ed665")), // CallreferenceNumber
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("33609004940")),      // MSCAddress
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("20160501010101")),   // SSPTime
					diam.NewAVP(20938, avp.Mbit, 2011, datatype.OctetString("0x00000001")),      // HighLayerCharacteristics
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Integer32(8)),                   // Time-Zone
				},
			}),
		}})
	if _, err := ccr.NewAVP("Framed-IP-Address", avp.Mbit, 0, datatype.UTF8String("10.228.16.4")); err != nil {
		t.Error(err)
	}
	tStart := time.Now()
	maxLoops := 100000
	for i := 0; i < *interations; i++ {
		if err := dmtClient.SendMessage(ccr); err != nil {
			t.Error(err)
		}
		msg := dmtClient.ReceivedMessage(rplyTimeout)
		if msg == nil {
			t.Fatal("No message returned")
		}
		/*
			if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
				t.Error(err)
			} else if len(avps) == 0 {
				t.Error("Result-Code")
			} else if strResult := avpValAsString(avps[0]); strResult != "300" { // Result-Code set in the template
				t.Errorf("Expecting 300, received: %s", strResult)
			}
		*/
	}
	totalDur := time.Now().Sub(tStart)
	fmt.Printf("Total duration: %v resulting %f ops per second\n", totalDur, float64(maxLoops)/totalDur.Seconds())
}

/*
func TestDmtAgentLoadCER(t *testing.T) {
	if !*testIntegration {
		return
	}
	m := diam.NewRequest(diam.CapabilitiesExchange, 4, dict.Default)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, datatype.Address(net.ParseIP("127.0.0.1")))
	m.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(999))
	m.NewAVP(avp.ProductName, 0, 0, datatype.UTF8String("CGR-DA"))
	m.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.FirmwareRevision, avp.Mbit, 0, datatype.Unsigned32(1))
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
}
*/

func TestDmtAgentStopEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
