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
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam/dict"
)

func TestHaPoolInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	daCfgPath = path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater1")
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
func TestHaPoolResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestHaPoolResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestHaPoolStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	engine.KillEngine(*waitRater) // just to make sure

	cgrRater1 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater1")
	if _, err := engine.StartEngine(cgrRater1, *waitRater); err != nil {
		t.Fatal("cgrRater1: ", err)
	}
	cgrRater2 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater2")
	if _, err := engine.StartEngine(cgrRater2, *waitRater); err != nil {
		t.Fatal("cgrRater2: ", err)
	}
	cgrSmg1 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrsmg1")
	if _, err := engine.StartEngine(cgrSmg1, *waitRater); err != nil {
		t.Fatal("cgrSmg1: ", err)
	}
	cgrSmg2 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrsmg2")
	if _, err := engine.StartEngine(cgrSmg2, *waitRater); err != nil {
		t.Fatal("cgrSmg2: ", err)
	}

}

// Connect rpc client to rater
func TestHaPoolApierRpcConn(t *testing.T) {
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
func TestHaPoolTPFromFolder(t *testing.T) {
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
func TestHaPoolSendCCRInit(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err = NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
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
func TestHaPoolSendCCRUpdate(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(300) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
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
func TestHaPoolSendCCRUpdate2(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(600) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
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
	eAcntVal := 8.944
	if err := apierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestHaPoolSendCCRTerminate(t *testing.T) {
	if !*testIntegration {
		return
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1004", Supplier: "SUPPL1",
		SetupTime: time.Date(2015, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2015, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
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

func TestHaPoolCdrs(t *testing.T) {
	if !*testIntegration {
		return
	}
	var cdrs []*engine.ExternalCdr
	req := utils.RpcCdrsFilter{RunIds: []string{utils.META_DEFAULT}}
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

func TestHaPoolStopEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
