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

var daCfgPath string
var daCfg *config.CGRConfig
var apierRpc *rpc.Client
var dmtClient *DiameterClient

var rplyTimeout time.Duration

func TestDiamITInitCfg(t *testing.T) {
	daCfgPath = path.Join(*dataDir, "conf", "samples", "diamagent")
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
func TestDiamITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestDiamITResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

/*
// Start CGR Engine
func TestDiamITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(daCfgPath, 4000); err != nil {
		t.Fatal(err)
	}
}
*/

func TestDiamITConnectDiameterClient(t *testing.T) {
	dmtClient, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "INTEGRATION_TESTS",
		daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName,
		utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesPath)
	if err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestDiamITApierRpcConn(t *testing.T) {
	var err error
	apierRpc, err = jsonrpc.Dial("tcp", daCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestDiamITTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := apierRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(1000) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestDiamITDryRun(t *testing.T) {
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("cgrates;1451911932;00082"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
	ccr.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity("CGR-DA"))
	ccr.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("CGR-DA"))
	ccr.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	ccr.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("TestDiamITDryRun")) // Match specific DryRun profile
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
	if _, err := ccr.NewAVP("Framed-IP-Address", avp.Mbit, 0, datatype.UTF8String("10.228.16.4")); err != nil {
		t.Error(err)
	}
	for i := 0; i < *interations; i++ {
		if err := dmtClient.SendMessage(ccr); err != nil {
			t.Error(err)
		}
		msg := dmtClient.ReceivedMessage(rplyTimeout)
		if msg == nil {
			t.Fatal("No message returned")
		}
		avps, err := msg.FindAVPsWithPath([]interface{}{"Result-Code"}, dict.UndefinedVendorID)
		if err != nil {
			t.Error(err)
		}
		if len(avps) == 0 {
			t.Error("Result-Code")
		}
		eVal := "300"
		if val, err := diamAVPAsString(avps[0]); err != nil {
			t.Error(err)
		} else if val != eVal {
			t.Errorf("expecting: %s, received: <%s>", eVal, val)
		}
	}
}
