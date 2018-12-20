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
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var interations = flag.Int("iterations", 1, "Number of iterations to do for dry run simulation")
var replyTimeout = flag.String("reply_timeout", "1s", "Maximum duration to wait for a reply")

var daCfgPath, diamConfigDIR string
var daCfg *config.CGRConfig
var apierRpc *rpc.Client
var diamClnt *DiameterClient

var rplyTimeout time.Duration

var sTestsDiam = []func(t *testing.T){
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
	testDiamItCCRTerminate,
	testDiamItCCRSMS,
	testDiamItKillEngine,
}

//Test start here
func TestDiamItTcp(t *testing.T) {
	diamConfigDIR = "diamagent"
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
}

func TestDiamItSctp(t *testing.T) {
	diamConfigDIR = "diamsctpagent"
	for _, stest := range sTestsDiam {
		t.Run(diamConfigDIR, stest)
	}
}

func TestDiamItMaxConn(t *testing.T) {
	diamConfigDIR = "diamagentmaxconn"
	for _, stest := range sTestsDiam[:7] {
		t.Run(diamConfigDIR, stest)
	}
	t.Run(diamConfigDIR, testDiamItDryRunMaxConn)
	t.Run(diamConfigDIR, testDiamItKillEngine)
}

func testDiamItInitCfg(t *testing.T) {
	daCfgPath = path.Join(*dataDir, "conf", "samples", diamConfigDIR)
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
	if _, err := engine.StopStartEngine(daCfgPath, 200); err != nil {
		t.Fatal(err)
	}
}

func testDiamItConnectDiameterClient(t *testing.T) {
	if diamConfigDIR == "diamsctpagent" || diamConfigDIR == "diamagentmaxconn" {
		daCfg.DiameterAgentCfg().DictionariesPath = ""
	}
	diamClnt, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "INTEGRATION_TESTS",
		daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION,
		daCfg.DiameterAgentCfg().DictionariesPath, daCfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDiamItApierRpcConn(t *testing.T) {
	var err error
	apierRpc, err = jsonrpc.Dial("tcp", daCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testDiamItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := apierRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond) // Give time for scheduler to execute topups
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
		if err := diamClnt.SendMessage(ccr); err != nil {
			t.Error(err)
		}
		msg := diamClnt.ReceivedMessage(rplyTimeout)
		if msg == nil {
			t.Fatal("No message returned")
		}
		// Result-Code
		eVal := "2002"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "cgrates;1451911932;00082"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Session-Id"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "CGR-DA"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Origin-Host"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "cgrates.org"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Origin-Realm"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "4"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Auth-Application-Id"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "1"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"CC-Request-Type"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "1"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"CC-Request-Number"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Multiple-Services-Credit-Control", "Rating-Group"}, dict.UndefinedVendorID); err != nil {
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
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Multiple-Services-Credit-Control", "Used-Service-Unit", "CC-Total-Octets"}, dict.UndefinedVendorID); err != nil {
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
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"}, dict.UndefinedVendorID); err != nil {
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

func testDiamItDryRunMaxConn(t *testing.T) {
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
	if _, err := ccr.NewAVP("Framed-IP-Address", avp.Mbit, 0, datatype.UTF8String("10.228.16.4")); err != nil {
		t.Error(err)
	}
	for i := 0; i < *interations; i++ {
		if err := diamClnt.SendMessage(ccr); err != nil {
			t.Error(err)
		}
		msg := diamClnt.ReceivedMessage(rplyTimeout)
		if msg == nil {
			t.Fatal("No message returned")
		}
		// Result-Code
		eVal := "5012"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "cgrates;1451911932;00082"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Session-Id"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "CGR-DA"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Origin-Host"}, dict.UndefinedVendorID); err != nil {
			t.Error(err)
		} else if len(avps) == 0 {
			t.Error("Missing AVP")
		} else if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
		eVal = "cgrates.org"
		if avps, err := msg.FindAVPsWithPath([]interface{}{"Origin-Realm"}, dict.UndefinedVendorID); err != nil {
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
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(1))
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
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Bearer-Capability
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
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
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
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"},
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

func testDiamItCCRUpdate(t *testing.T) {
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("bb97be2b9f37c2be9614fff71c8b1d08b1acbff8"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("192.168.1.1"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(2))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(2))
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
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Bearer-Capability
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
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
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
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Granted-Service-Unit", "CC-Time"},
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
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(3))
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
					diam.NewAVP(20313, avp.Mbit, 2011, datatype.UTF8String("")),                                         // Bearer-Capability
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
	if avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID); err != nil {
		t.Error(err)
	} else if len(avps) == 0 {
		t.Error("Missing AVP")
	} else if val, err := diamAVPAsString(avps[0]); err != nil {
		t.Error(err)
	} else if val != eVal {
		t.Errorf("expecting: %s, received: <%s>", eVal, val)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := apierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
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
	if err := diamClnt.SendMessage(ccr); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Duration(100) * time.Millisecond)
	diamClnt.ReceivedMessage(rplyTimeout)

	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, ToRs: []string{utils.SMS}}
	if err := apierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != 1 {
			t.Errorf("Unexpected Usage CDR: %+v", cdrs[0])
		}
	}
}

func testDiamItKillEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
