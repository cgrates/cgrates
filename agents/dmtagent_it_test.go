/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
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
)

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

var daCfgPath string
var daCfg *config.CGRConfig
var apierRpc *rpc.Client

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
	eSMGE := sessionmanager.SMGenericEvent{"EventName": "DIAMETER_CCR", "AccId": "routinga;1442095190;1476802709",
		"Account": "*users", "AnswerTime": "2015-11-23 12:22:24 +0000 UTC", "Category": "call",
		"Destination": "4986517174964", "Direction": "*out", "ReqType": "*users", "SetupTime": "2015-11-23 12:22:24 +0000 UTC",
		"Subject": "*users", "SubscriberId": "4986517174963", "TOR": "*voice", "Tenant": "*users", "Usage": "300"}
	if smge, err := ccr.AsSMGenericEvent(cfgDefaults.DiameterAgentCfg().RequestProcessors[0].ContentFields); err != nil {
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

func TestDmtAgentSendCCRInit(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err := NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(0) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
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
}

func TestDmtAgentSendCCRUpdate(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err := NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(610) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
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
}

func TestDmtAgentSendCCRTerminate(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err := NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(610) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
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
}

func TestDmtAgentStopEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
