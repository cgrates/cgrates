//go:build integration
// +build integration

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

package agents

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
)

var (
	interations  = flag.Int("iterations", 1, "Number of iterations to do for dry run simulation")
	replyTimeout = flag.String("reply_timeout", "1s", "Maximum duration to wait for a reply")

	daCfgPath, diamConfigDIR string
	daCfg                    *config.CGRConfig
	apierRpc                 *birpc.Client
	diamClnt                 *DiameterClient

	rplyTimeout time.Duration

	isDispatcherActive bool

	sTestsDiam = []func(t *testing.T){
		testDiamItInitCfg,
		testDiamItResetDataDb,
		testDiamItResetStorDb,
		testDiamItStartEngine,
		testDiamItConnectDiameterClient,
		testDiamItApierRpcConn,
		testDiamItTPFromFolder,
		testDiamItDryRun,
		testDiamItCCRInit,
		testDiamItCCRUpdate,

		testDiamItRAR,

		testDiamItCCRTerminate,
		testDiamItCCRSMS,
		testDiamItCCRMMS,

		testDiamItEmulateTerminate,

		testDiamItTemplateErr,
		testDiamItCCRInitWithForceDuration,

		testDiamItDRR,

		testDiamItKillEngine,
	}
)

// Test start here
func TestDiamItTcp(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIR = "diamagent_internal"
	case utils.MetaMySQL:
		diamConfigDIR = "diamagent_mysql"
	case utils.MetaMongo:
		diamConfigDIR = "diamagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
}

func TestDiamItDispatcher(t *testing.T) {
	if *utils.Encoding == utils.MetaGOB {
		t.SkipNow()
		return
	}
	testDiamItResetAllDB(t)
	isDispatcherActive = true
	engine.StartEngine(path.Join(*utils.DataDir, "conf", "samples", "dispatchers", "all"), 200)
	engine.StartEngine(path.Join(*utils.DataDir, "conf", "samples", "dispatchers", "all2"), 200)
	diamConfigDIR = "dispatchers/diamagent"
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
	isDispatcherActive = false
}

func TestDiamItSctp(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIR = "diamsctpagent_internal"
	case utils.MetaMySQL:
		diamConfigDIR = "diamsctpagent_mysql"
	case utils.MetaMongo:
		diamConfigDIR = "diamsctpagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
}

func TestDiamItBiRPC(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIR = "diamagent_internal_%sbirpc"
	case utils.MetaMySQL:
		diamConfigDIR = "diamagent_mysql_%sbirpc"
	case utils.MetaMongo:
		diamConfigDIR = "diamagent_mongo_%sbirpc"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	diamConfigDIR = fmt.Sprintf(diamConfigDIR, strings.TrimPrefix(*utils.Encoding, utils.Meta))
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
}

func TestDiamItSessionDisconnect(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIR = "diamagent_internal"
	case utils.MetaMySQL:
		diamConfigDIR = "diamagent_mysql"
	case utils.MetaMongo:
		diamConfigDIR = "diamagent_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDiam[:7] {
		t.Run(diamConfigDIR, stest)
	}
	t.Run(diamConfigDIR, testDiamInitWithSessionDisconnect)
	t.Run(diamConfigDIR, testDiamItKillEngine)
}

func testDiamItInitCfg(t *testing.T) {
	daCfgPath = path.Join(*utils.DataDir, "conf", "samples", diamConfigDIR)
	// Init config first
	var err error
	daCfg, err = config.NewCGRConfigFromPath(daCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	daCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
	rplyTimeout, _ = utils.ParseDurationWithSecs(*replyTimeout)
	if isDispatcherActive {
		daCfg.ListenCfg().RPCJSONListen = ":6012"
	}
}

func testDiamItResetAllDB(t *testing.T) {
	cfgPath1 := path.Join(*utils.DataDir, "conf", "samples", "dispatchers", "all")
	allCfg, err := config.NewCGRConfigFromPath(cfgPath1)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(allCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(allCfg); err != nil {
		t.Fatal(err)
	}

	cfgPath2 := path.Join(*utils.DataDir, "conf", "samples", "dispatchers", "all2")
	allCfg2, err := config.NewCGRConfigFromPath(cfgPath2)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(allCfg2); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(allCfg2); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testDiamItResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDiamItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDiamItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(daCfgPath, 500); err != nil {
		t.Fatal(err)
	}
}

func testDiamItConnectDiameterClient(t *testing.T) {
	diamClnt, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "INTEGRATION_TESTS",
		daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorID,
		daCfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		daCfg.DiameterAgentCfg().DictionariesPath, daCfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDiamItApierRpcConn(t *testing.T) {
	var err error
	apierRpc, err = newRPCClient(daCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testDiamItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := apierRpc.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	if isDispatcherActive {
		testDiamItTPLoadData(t)
	}
	time.Sleep(100 * time.Millisecond) // Give time for scheduler to execute topups
}

func testDiamItTPLoadData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", daCfgPath, "-path", path.Join(*utils.DataDir, "tariffplans", "dispatchers"))

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

func testDiamItDryRun(t *testing.T) {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("cgrates;1451911932;00082"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("CGR-DA"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("TestDiamItDryRun")) // Match specific DryRun profile
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
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
	ccr.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(7640)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(5337)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(2303)), // CC-Output-Octets
				},
			}),
			diam.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(65000)), // Service-Identifier
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(1)),           // Rating-Group
			diam.NewAVP(avp.ReportingReason, avp.Mbit, 0, datatype.Enumerated(2)),       // Reporting-Reason
		},
	})
	ccr.NewAVP(avp.MultipleServicesCreditControl, avp.Mbit, 0, &diam.GroupedAVP{ // Multiple-Services-Credit-Control
		AVP: []*diam.AVP{
			diam.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{ // Used-Service-Unit
				AVP: []*diam.AVP{
					diam.NewAVP(avp.CCTotalOctets, avp.Mbit, 0, datatype.Unsigned64(3000)),  // CC-Total-Octets
					diam.NewAVP(avp.CCInputOctets, avp.Mbit, 0, datatype.Unsigned64(2000)),  // CC-Input-Octets
					diam.NewAVP(avp.CCOutputOctets, avp.Mbit, 0, datatype.Unsigned64(1000)), // CC-Output-Octets
				},
			}),
			diam.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(65000)), // Service-Identifier
			diam.NewAVP(avp.RatingGroup, avp.Mbit, 0, datatype.Unsigned32(2)),           // Rating-Group
			diam.NewAVP(avp.ReportingReason, avp.Mbit, 0, datatype.Enumerated(2)),       // Reporting-Reason
		},
	})
	if _, err := ccr.NewAVP("Framed-IP-Address", avp.Mbit, 0, datatype.UTF8String("10.228.16.4")); err != nil {
		t.Error(err)
	}
	for i := 0; i < *interations; i++ {
		// ============================================
		// prevent nil pointer dereference
		// ============================================
		if diamClnt == nil {
			t.Fatal("Diameter client should not be nil")
		}
		if diamClnt.conn == nil {
			t.Fatal("Diameter connection should not be nil")
		}
		if ccr == nil {
			t.Fatal("The mesage to diameter should not be nil")
		}
		// ============================================

		if err := diamClnt.SendMessage(ccr); err != nil {
			t.Error(err)
		}
		msg := diamClnt.ReceivedMessage(rplyTimeout)
		if msg == nil {
			t.Fatal("No message returned")
		}
		// Result-Code
		eVal := "2002"
		if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}

		eVal = "cgrates;1451911932;00082"
		if avps, err := msg.FindAVPsWithPath([]any{"Session-Id"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "CGR-DA"
		if avps, err := msg.FindAVPsWithPath([]any{"Origin-Host"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "cgrates.org"
		if avps, err := msg.FindAVPsWithPath([]any{"Origin-Realm"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "4"
		if avps, err := msg.FindAVPsWithPath([]any{"Auth-Application-Id"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "1"
		if avps, err := msg.FindAVPsWithPath([]any{"CC-Request-Type"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "1"
		if avps, err := msg.FindAVPsWithPath([]any{"CC-Request-Number"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		if avps, err := msg.FindAVPsWithPath([]any{"Multiple-Services-Credit-Control", "Rating-Group"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) != 2 {
			t.Errorf("Unexpected number of Multiple-Services-Credit-Control.Rating-Group : %d", len(avps))
		} else {
			if val, err := diamAVPAsString(avps[0]); err != nil {
				t.Error(err)
			} else if val != "1" {
				t.Errorf("expecting: 1, received: <%s>", val)
			}
			if val, err := diamAVPAsString(avps[1]); err != nil {
				t.Error(err)
			} else if val != "2" {
				t.Errorf("expecting: 2, received: <%s>", val)
			}
		}
		if avps, err := msg.FindAVPsWithPath([]any{"Multiple-Services-Credit-Control", "Used-Service-Unit", "CC-Total-Octets"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) != 2 {
			t.Errorf("Unexpected number of Multiple-Services-Credit-Control.Used-Service-Unit.CC-Total-Octets : %d", len(avps))
		} else {
			if val, err := diamAVPAsString(avps[0]); err != nil {
				t.Error(err)
			} else if val != "7640" {
				t.Errorf("expecting: 7640, received: <%s>", val)
			}
			if val, err := diamAVPAsString(avps[1]); err != nil {
				t.Error(err)
			} else if val != "3000" {
				t.Errorf("expecting: 3000, received: <%s>", val)
			}
		}
		eVal = "6" // sum of items
		if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
	}
}

func testDiamItCCRInit(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
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
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "300" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItCCRInitWithForceDuration(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbfx1"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("forceDurationVoice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(3000000000))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "5030"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}

}

func testDiamItCCRUpdate(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
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
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "300" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItCCRTerminate(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(3))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(2))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 15, 12, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),      // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("1006")), // Subscription-Id-Data
		}})
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RequestedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(0))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(250))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}}
	if err := apierRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != 550*time.Second {
			t.Errorf("Unexpected Usage CDR: %+v", cdrs[0])
		}
		// in case of sctp we get OriginHost ip1/ip2/ip3/...
		if !strings.Contains(cdrs[0].OriginHost, "127.0.0.1") {
			t.Errorf("Unexpected OriginHost CDR: %+v", cdrs[0])
		}
	}
}

func testDiamItCCRSMS(t *testing.T) {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("TestDmtAgentSendCCRSMS"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("message@DiamItCCRSMS"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
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
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1001")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if ccr == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(ccr); err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond)
	diamClnt.ReceivedMessage(rplyTimeout)

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, ToRs: []string{utils.MetaSMS}}}
	if err := apierRpc.Call(context.Background(), utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else if cdrs[0].Usage != 1 {
		t.Errorf("Unexpected Usage CDR: %+v", cdrs[0])
	}
}

func testDiamItCCRMMS(t *testing.T) {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("TestDmtAgentSendCCRMMS"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("mms@DiamItCCRMMS"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
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
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1001")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if ccr == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(ccr); err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond)
	diamClnt.ReceivedMessage(rplyTimeout)

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, ToRs: []string{utils.MetaMMS}}}
	if err := apierRpc.Call(context.Background(), utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else if cdrs[0].Usage != 1 {
		t.Errorf("Unexpected Usage CDR: %+v", cdrs[0])
	}
}

func testDiamInitWithSessionDisconnect(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:     "testDiamInitWithSessionDisconnect",
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Second),
		Balance: map[string]any{
			utils.ID:            "testDiamInitWithSessionDisconnect",
			utils.RatingSubject: "*zero1ms",
		},
	}
	var reply string
	if err := apierRpc.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	sessID := "bb97be2b9f37c2be9614fff71c8b1d08bdisconnect"
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessID))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("testSessionDisconnect"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("testDiamInitWithSessionDisconnect"))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),                                   // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("testDiamInitWithSessionDisconnect")), // Subscription-Id-Data
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
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1001")),                                         // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                         // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                        // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                             // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                            // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                            // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                           // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08bdisconnect")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                            // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                             // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                            // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                            // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	time.Sleep(2 * time.Second)
	msg = diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	if avps, err := msg.FindAVPsWithPath([]any{"Session-Id"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != sessID {
		t.Errorf("expecting: %s, received: <%s>", sessID, val)
	}
}

func testDiamItKillEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}

func testDiamItRAR(t *testing.T) {
	if diamConfigDIR == "dispatchers/diamagent" {
		t.SkipNow()
	}
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	// ============================================
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		var reply string
		if err := apierRpc.Call(context.Background(), utils.SessionSv1AlterSessions, utils.SessionFilterWithEvent{}, &reply); err != nil {
			t.Error(err)
		}
		wait.Done()
	}()
	rar := diamClnt.ReceivedMessage(rplyTimeout)
	if rar == nil {
		t.Fatal("No message returned")
	}

	raa := rar.Answer(2001)
	raa.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))

	if err := diamClnt.SendMessage(raa); err != nil {
		t.Error(err)
	}

	wait.Wait()

	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
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
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(301))}})
	m.NewAVP(avp.UsedServiceUnit, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(420, avp.Mbit, 0, datatype.Unsigned32(301))}})
	m.NewAVP(873, avp.Mbit, 10415, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(20300, avp.Mbit, 2011, &diam.GroupedAVP{ // IN-Information
				AVP: []*diam.AVP{
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "301" // 5 mins of session
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItDRR(t *testing.T) {
	if diamConfigDIR == "dispatchers/diamagent" {
		t.SkipNow()
	}
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	// ============================================
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		var reply string
		if err := apierRpc.Call(context.Background(), utils.SessionSv1DisconnectPeer, &utils.DPRArgs{
			OriginHost:      "INTEGRATION_TESTS",
			OriginRealm:     "cgrates.org",
			DisconnectCause: 1, // BUSY
		}, &reply); err != nil {
			t.Error(err)
		}
		wait.Done()
	}()
	drr := diamClnt.ReceivedMessage(rplyTimeout)
	if drr == nil {
		t.Fatal("No message returned")
	}

	dra := drr.Answer(2001)
	// dra.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("INTEGRATION_TESTS"))
	// dra.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))

	if err := diamClnt.SendMessage(dra); err != nil {
		t.Error(err)
	}

	wait.Wait()

	eVal := "1"
	if avps, err := drr.FindAVPsWithPath([]any{avp.DisconnectCause}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItTemplateErr(t *testing.T) {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("TestDmtAgentSendCCRError"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("error@DiamItError"))
	ccr.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(4))
	ccr.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	ccr.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 5, 11, 43, 10, 0, time.UTC)))
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
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1001")), // Address-Data
						}}),
					diam.NewAVP(1201, avp.Mbit, 10415, &diam.GroupedAVP{ // Recipient-Address
						AVP: []*diam.AVP{
							diam.NewAVP(899, avp.Mbit, 10415, datatype.Enumerated(1)),      // Address-Type
							diam.NewAVP(897, avp.Mbit, 10415, datatype.UTF8String("1003")), // Address-Data
						}}),
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if ccr == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(ccr); err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond)
	msg := diamClnt.ReceivedMessage(rplyTimeout)

	if msg == nil {
		t.Fatal("Message should not be nil")
	}
	// Result-Code
	eVal := "5012" // error code diam.UnableToComply
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
}

func testDiamItEmulateTerminate(t *testing.T) {
	if diamConfigDIR == "dispatchers/diamagent" {
		t.SkipNow()
	}
	//add a default charger
	tpDirPath := t.TempDir()
	filePath := path.Join(tpDirPath, utils.ChargersCsv)
	err := os.WriteFile(filePath,
		[]byte(`cgrates.com,CustomCharger,,,CustomCharger,*constant:*req.Category:custom_charger,20
cgrates.com,Default,,,*default,*none,20`),
		0644)
	if err != nil {
		t.Errorf("could not write to file %s: %v",
			filePath, err)
	}
	var reply string
	args := &utils.AttrLoadTpFromFolder{FolderPath: tpDirPath}
	err = apierRpc.Call(context.Background(),
		utils.APIerSv1LoadTariffPlanFromFolder,
		args, &reply)
	if err != nil {
		t.Errorf("%s call failed for path %s: %v",
			utils.APIerSv1LoadTariffPlanFromFolder, tpDirPath, err)
	}

	//set the account
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.com",
		Account:     "testDiamItEmulateTerminate",
		Value:       float64(time.Hour),
		BalanceType: utils.MetaVoice,
		Balance: map[string]any{
			utils.ID:         "testDiamItEmulateTerminate",
			utils.Categories: "custom_charger",
		},
	}
	if err := apierRpc.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.com", Account: "testDiamItEmulateTerminate"}
	if err := apierRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != float64(time.Hour) {
		t.Errorf("Expected: %f, received: %f", float64(time.Hour), acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.com"))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("voice@DiamItCCRInit"))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Date(2018, 10, 4, 14, 42, 20, 0, time.UTC)))
	m.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(0)),                            // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("testDiamItEmulateTerminate")), // Subscription-Id-Data
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
					diam.NewAVP(831, avp.Mbit, 10415, datatype.UTF8String("1006")),                                      // Calling-Party-Address
					diam.NewAVP(832, avp.Mbit, 10415, datatype.UTF8String("1002")),                                      // Called-Party-Address
					diam.NewAVP(20327, avp.Mbit, 2011, datatype.UTF8String("1002")),                                     // Real-Called-Number
					diam.NewAVP(20339, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Charge-Flow-Type
					diam.NewAVP(20302, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-Vlr-Number
					diam.NewAVP(20303, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Calling-CellID-Or-SAI
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.OctetString("")),                                        // Bearer-Capability
					diam.NewAVP(20321, avp.Mbit, 2011, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8")), // Call-Reference-Number
					diam.NewAVP(20322, avp.Mbit, 2011, datatype.UTF8String("")),                                         // MSC-Address
					diam.NewAVP(20324, avp.Mbit, 2011, datatype.Unsigned32(0)),                                          // Time-Zone
					diam.NewAVP(20385, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Called-Party-NP
					diam.NewAVP(20386, avp.Mbit, 2011, datatype.UTF8String("")),                                         // SSP-Time
				},
			}),
		}})
	// ============================================
	// prevent nil pointer dereference
	// ============================================
	if diamClnt == nil {
		t.Fatal("Diameter client should not be nil")
	}
	if diamClnt.conn == nil {
		t.Fatal("Diameter connection should not be nil")
	}
	if m == nil {
		t.Fatal("The mesage to diameter should not be nil")
	}
	// ============================================
	if err := diamClnt.SendMessage(m); err != nil {
		t.Error(err)
	}
	msg := diamClnt.ReceivedMessage(rplyTimeout)
	if msg == nil {
		t.Fatal("No message returned")
	}
	// Result-Code
	eVal := "2001"
	if avps, err := msg.FindAVPsWithPath([]any{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	// Result-Code
	eVal = "0" // 0 from sessions
	if avps, err := msg.FindAVPsWithPath([]any{"Granted-Service-Unit", "CC-Time"},
		dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}

	if err := apierRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != float64(time.Hour) {
		t.Errorf("Expected: %f, received: %f", float64(time.Hour), acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
}

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}
