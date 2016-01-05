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

var daCfgPath string
var daCfg *config.CGRConfig
var apierRpc *rpc.Client
var dmtClient *DiameterClient
var err error

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
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
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
	var loadInst engine.LoadInstance
	if err := apierRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:47:26Z"'
func TestDmtAgentSendCCRInit(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
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
	msg := dmtClient.ReceivedMessage()
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s", strCCTime)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.484
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
	cdr := &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
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
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage()
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s", strCCTime)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.214
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
	cdr := &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
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
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage()
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Granted-Service-Unit not found")
	} else if strCCTime := avpValAsString(avps[0]); strCCTime != "300" {
		t.Errorf("Expecting 300, received: %s", strCCTime)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.944000
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
	cdr := &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Direction: "*out",
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
	time.Sleep(time.Duration(100) * time.Millisecond)
	msg := dmtClient.ReceivedMessage()
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
}

func TestDmtAgentSendCCRSMS(t *testing.T) {
	if !*testIntegration {
		return
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("cgrates;1451911932;00082"))
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
	/*
		time.Sleep(time.Duration(100) * time.Millisecond)
		msg := dmtClient.ReceivedMessage()
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
}

func TestDmtAgentCdrs(t *testing.T) {
	if !*testIntegration {
		return
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
	if err := apierRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "610" {
			t.Errorf("Unexpected CDR Usage received, cdr: %+v ", cdrs[0])
		}
		if cdrs[0].Cost != 0.795 {
			t.Errorf("Unexpected CDR Cost received, cdr: %+v ", cdrs[0])
		}
	}
}

func TestDmtAgentStopEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
