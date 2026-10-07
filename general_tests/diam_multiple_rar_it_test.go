//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"os/exec"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/agents"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

var (
	daCfgPathM, diamConfigDIRM string
	daCfgM                     *config.CGRConfig
	apierRpcM                  *birpc.Client
	diamClntM                  *agents.DiameterClient

	rplyTimeoutM time.Duration

	sTestsDiamActionsMultipleRAR = []func(t *testing.T){
		testDiamItMultipleInitCfg,
		testDiamItMultipleResetDataDb,
		testDiamItMultipleResetStorDb,
		testDiamItMultipleStartEngine,
		testDiamItMultipleConnectDiameterClient,
		testDiamItMultipleApierRpcConn,
		testDiamItMultipleTPFromFolder,
		testDiamItMultipleCCRInit,
		testDiamItMultipleCCRUpdate,
		testDiamItMultipleCCRInit2,
		testDiamItMultipleCCRUpdate2,

		testDiamItMultipleCCRInitDifferentAccount,
		testDiamItMultipleCCRUpdateDifferentAccount,
		testDiamItMultipleCCRInit2DifferentAccount,
		testDiamItMultipleCCRUpdate2DifferentAccount,

		testDiamItMultipleRARActions,
		testDiamItMultipleKillEngine,
	}
)

// Tests RAR sent to all sessions from an account (sessions have different SessionID)
func TestDiamItTcpMultipleRARFromActions(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIRM = "diamagent_internal"
	case utils.MetaMySQL:
		diamConfigDIRM = "diamagent_mysql"
	case utils.MetaMongo:
		diamConfigDIRM = "diamagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDiamActionsMultipleRAR {
		t.Run(diamConfigDIRM, stest)
	}
}

func testDiamItMultipleInitCfg(t *testing.T) {
	daCfgPathM = path.Join(*utils.DataDir, "conf", "samples", diamConfigDIRM)
	// Init config first
	var err error
	daCfgM, err = config.NewCGRConfigFromPath(daCfgPathM)
	if err != nil {
		t.Fatal(err)
	}
	daCfgM.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
	rplyTimeoutM, _ = utils.ParseDurationWithSecs(*replyTimeout)
}

// Remove data in both rating and accounting db
func testDiamItMultipleResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(daCfgM); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDiamItMultipleResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(daCfgM); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDiamItMultipleStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(daCfgPathM, 500); err != nil {
		t.Fatal(err)
	}
}

func testDiamItMultipleConnectDiameterClient(t *testing.T) {
	var err error
	diamClntM, err = agents.NewDiameterClient(daCfgM.DiameterAgentCfg().Listeners[0].Address, "INTEGRATION_TESTS",
		daCfgM.DiameterAgentCfg().OriginRealm, daCfgM.DiameterAgentCfg().VendorID,
		daCfgM.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		daCfgM.DiameterAgentCfg().DictionariesPath, daCfgM.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDiamItMultipleApierRpcConn(t *testing.T) {
	var err error
	switch *utils.Encoding {
	case utils.MetaJSON: // We connect over JSON so we can also troubleshoot if needed
		if apierRpcM, err = jsonrpc.Dial(utils.TCP, daCfgM.ListenCfg().RPCJSONListen); err != nil {
			t.Fatal(err)
		}
	case utils.MetaGOB:
		if apierRpcM, err = birpc.Dial(utils.TCP, daCfgM.ListenCfg().RPCGOBListen); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("UNSUPPORTED_RPC")
	}
}

// Load the tariff plan, creating accounts and their balances
func testDiamItMultipleTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := apierRpcM.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond) // Give time for scheduler to execute topups
}

func testDiamItMultipleTPLoadData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", daCfgPathM, "-path", path.Join(*utils.DataDir, "tariffplans", "dispatchers"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(time.Second):
		t.Errorf("cgr-loader failed: ")
	}
}

func testDiamItMultipleCCRInit(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR1SessID"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),        // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),        // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),       // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),          // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR1SessID")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),           // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),           // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),           // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleCCRUpdate(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR1SessID"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 57, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),        // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),        // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),       // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),          // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR1SessID")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),           // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),           // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),           // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleCCRInit2(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR2SessID"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),        // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),        // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),       // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),          // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR2SessID")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),           // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),           // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),           // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}
func testDiamItMultipleCCRUpdate2(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR2SessID"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 57, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),        // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),        // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),       // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),           // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),          // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR2SessID")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),           // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),            // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),           // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),           // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleCCRInitDifferentAccount(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR1SessID1002"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1002")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1002")),            // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1001")),            // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1001")),           // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),              // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR1SessID1002")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),               // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),               // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),               // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleCCRUpdateDifferentAccount(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR1SessID1002"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 57, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1002")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1002")),            // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1001")),            // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1001")),           // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),              // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR1SessID1002")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),               // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),               // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),               // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleCCRInit2DifferentAccount(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR2SessID1002"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1002")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1002")),            // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1001")),            // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1001")),           // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),              // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR2SessID1002")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),               // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),               // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),               // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}
func testDiamItMultipleCCRUpdate2DifferentAccount(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("CCR2SessID1002"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 57, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1002")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(300))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1002")),            // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1001")),            // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1001")),           // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),               // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),              // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("CCR2SessID1002")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),               // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),               // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),               // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClntM.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClntM.ReceivedMessage(rplyTimeoutM)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "Unsigned32{2001}"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "Unsigned32{300}" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val := avps[0].Data.String(); val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItMultipleRARActions(t *testing.T) { // test RAR ran from ExecuteAction *alter_sessions
	var aSessions []*sessions.ExternalSession
	if err := apierRpcM.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 8 {
		t.Errorf("expected 8 active sessions, got <%v>", utils.ToJSON(aSessions))
	}
	// t.Log(utils.ToJSON(aSessions))
	if diamConfigDIRM == "dispatchers/diamagent" {
		t.SkipNow()
	}
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClntM == nil {
		t.Fatal("Diameter client should not be nil")
	}
	// ============================================
	var reply string
	attrs1 := &v1.V1AttrSetActions{
		ActionsId: "SendRAR",
		Actions: []*v1.V1TPAction{
			{
				Identifier:      utils.MetaAlterSessions, // *alter_sessions
				ExtraParameters: "cgrates.org;*string:~*req.Account:~*req.Account;;;",
			},
		},
	}
	if err := apierRpcM.Call(context.Background(), utils.APIerSv1SetActions, &attrs1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "SendRAR"}
		if err := apierRpcM.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
			t.Errorf("APIerSv1ExecuteAction failed unexpectedly: %v", err)
		}
		wait.Done()
	}()
	exp1 := `Session-Id {Code:263,Flags:0x40,Length:20,VendorId:0,Value:UTF8String{CCR1SessID},Padding:2}`
	exp2 := `Session-Id {Code:263,Flags:0x40,Length:20,VendorId:0,Value:UTF8String{CCR2SessID},Padding:2}`
	count := 0
	for {
		rar := diamClntM.ReceivedMessage(rplyTimeoutM)
		if rar == nil && count == 2 {
			break
		}
		if rar != nil && count == 2 {
			t.Fatalf("received more messages than expected")
		}
		if rar == nil {
			t.Fatal("No message returned")
		}
		count++
		if strings.Contains(rar.String(), exp1) {
			exp1 = "finished result"
		} else if strings.Contains(rar.String(), exp2) {
			exp2 = "finished result"
		} else {
			t.Fatalf("unexpected result <%v>", rar.String())
		}
		// t.Log(rar.String())
		raa := rar.Answer(2001)
		sessID, err := rar.FindAVP(avp.SessionID, 0)
		if err != nil {
			t.Fatal(err)
		}
		raa.NewAVP(avp.SessionID, avp.Mbit, 0, sessID.Data)
		if err := diamClntM.SendMessage(raa); err != nil { // send message back to rar so it closes its channel and doesnt timeout
			t.Error(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func testDiamItMultipleKillEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
