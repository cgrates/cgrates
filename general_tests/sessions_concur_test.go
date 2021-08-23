//go:build performance
// +build performance

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
package general_tests

import (
	"flag"
	"fmt"
	"net/rpc"
	"path"
	"sync"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sCncrCfgDIR, sCncrCfgPath string
	sCncrCfg                  *config.CGRConfig
	sCncrRPC                  *rpc.Client

	sCncrSessions = flag.Int("sessions", 100000, "maximum concurrent sessions created")
	sCncrCps      = flag.Int("cps", 50000, "maximum requests per second sent out")

	cpsPool = make(chan struct{}, *sCncrCps)
	acntIDs = make(chan string, 1)
	wg      sync.WaitGroup
)

// Tests starting here
func TestSCncrInternal(t *testing.T) {
	sCncrCfgDIR = "sessinternal"
	for _, tst := range sTestsSCncrIT {
		t.Run("InternalConn", tst)
	}
}

// Tests starting here
func TestSCncrJSON(t *testing.T) {
	sCncrCfgDIR = "sessintjson"
	for _, tst := range sTestsSCncrIT {
		t.Run("JSONConn", tst)
	}
}

// subtests to be executed
var sTestsSCncrIT = []func(t *testing.T){
	testSCncrInitConfig,
	testSCncrInitDataDB,
	testSCncrInitStorDB,
	testSCncrStartEngine,
	testSCncrRPCConn,
	testSCncrLoadTP,
	testSCncrRunSessions,
	testSCncrKillEngine,
}

func testSCncrInitConfig(t *testing.T) {
	sCncrCfgPath = path.Join(*dataDir, "conf", "samples", sCncrCfgDIR)
	if sCncrCfg, err = config.NewCGRConfigFromPath(sCncrCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testSCncrInitDataDB(t *testing.T) {
	if err := engine.InitDataDb(sCncrCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testSCncrInitStorDB(t *testing.T) {
	if err := engine.InitStorDb(sCncrCfg); err != nil {
		t.Fatal(err)
	}
}

func testSCncrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sCncrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSCncrRPCConn(t *testing.T) {
	var err error
	sCncrRPC, err = newRPCClient(sCncrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testSCncrKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testSCncrLoadTP(t *testing.T) {
	var loadInst string
	if err := sCncrRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "tp1cnt")}, &loadInst); err != nil {
		t.Error(err)
	}
	attrPrfl := &v2.AttributeWithAPIOpts{
		ExternalAttributeProfile: &engine.ExternalAttributeProfile{
			Tenant:   "cgrates.org",
			ID:       "AttrConcurrentSessions",
			Contexts: []string{utils.MetaAny},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "TestType",
					Value: "ConcurrentSessions",
				},
			},
			Weight: 20,
		},
	}
	var resAttrSet string
	if err := sCncrRPC.Call(utils.APIerSv2SetAttributeProfile, attrPrfl, &resAttrSet); err != nil {
		t.Error(err)
	} else if resAttrSet != utils.OK {
		t.Errorf("unexpected reply returned: <%s>", resAttrSet)
	}
}

func testSCncrRunSessions(t *testing.T) {
	acntIDsSet := utils.NewStringSet(nil)
	bufferTopup := 8760 * time.Hour
	for i := 0; i < *sCncrSessions; i++ {
		acntID := fmt.Sprintf("100%d", utils.RandomInteger(100, 200))
		if !acntIDsSet.Has(acntID) {
			// Special balance BUFFER to cover concurrency on MAIN one
			argsAddBalance := &v1.AttrAddBalance{
				Tenant:      "cgrates.org",
				Account:     acntID,
				BalanceType: utils.MetaVoice,
				Value:       float64(bufferTopup.Nanoseconds()),
				Balance: map[string]interface{}{
					utils.ID: "BUFFER",
				},
			}
			var addBlcRply string
			if err = sCncrRPC.Call(utils.APIerSv1AddBalance, argsAddBalance, &addBlcRply); err != nil {
				t.Error(err)
			} else if addBlcRply != utils.OK {
				t.Errorf("received: <%s>", addBlcRply)
			}
			acntIDsSet.Add(acntID)
		}
		acntIDs <- acntID
		wg.Add(1)
		go t.Run(fmt.Sprintf("RunSession#%d", i), testRunSession)
	}
	wg.Wait()
	for acntID := range acntIDsSet.Data() {
		// make sure the account was properly refunded
		var acnt *engine.Account
		acntAttrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: acntID}
		if err = sCncrRPC.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
			return
		} else if vcBlnc := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); float64(bufferTopup.Nanoseconds())-vcBlnc > 1000000.0 { // eliminate rounding errors
			t.Errorf("unexpected voice balance received: %+v", utils.ToIJSON(acnt))
		} else if mnBlnc := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); mnBlnc != 0 {
			t.Errorf("unexpected monetary balance received: %+v", utils.ToIJSON(acnt))
		}
	}
}

// runSession runs one session
func testRunSession(t *testing.T) {
	defer wg.Done()       // decrease group counter one out from test
	cpsPool <- struct{}{} // push here up to cps
	go func() {           // allow more requests after a second
		time.Sleep(time.Second)
		<-cpsPool
	}()
	acntID := <-acntIDs
	originID := utils.GenUUID() // each test with it's own OriginID
	// topup as much as we know we need for one session
	mainTopup := 90 * time.Second
	var addBlcRply string
	argsAddBalance := &v1.AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     acntID,
		BalanceType: utils.MetaVoice,
		Value:       float64(mainTopup.Nanoseconds()),
		Balance: map[string]interface{}{
			utils.ID:     "MAIN",
			utils.Weight: 10,
		},
	}
	if err = sCncrRPC.Call(utils.APIerSv1AddBalance, argsAddBalance, &addBlcRply); err != nil {
		t.Error(err)
	} else if addBlcRply != utils.OK {
		t.Errorf("received: <%s> to APIerSv1.AddBalance", addBlcRply)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond) // randomize between tests

	// Auth the session
	authDur := 5 * time.Minute
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrAuth%s", originID),
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     originID,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: acntID,
				utils.Destination:  fmt.Sprintf("%s%s", acntID, acntID),
				utils.SetupTime:    time.Now(),
				utils.Usage:        authDur,
			},
		},
	}
	var rplyAuth sessions.V1AuthorizeReply
	if err := sCncrRPC.Call(utils.SessionSv1AuthorizeEvent, authArgs, &rplyAuth); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond)

	// Init the session
	initUsage := time.Minute
	initArgs := &sessions.V1InitSessionArgs{
		InitSession:   true,
		GetAttributes: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrInit%s", originID),
			Event: map[string]interface{}{
				utils.OriginID:     originID,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: acntID,
				utils.Destination:  fmt.Sprintf("%s%s", acntID, acntID),
				utils.AnswerTime:   time.Now(),
				utils.Usage:        initUsage,
			},
		},
	}
	var rplyInit sessions.V1InitSessionReply
	if err := sCncrRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &rplyInit); err != nil {
		t.Error(err)
	} else if rplyInit.MaxUsage == 0 {
		t.Errorf("unexpected MaxUsage at init: %v", rplyInit.MaxUsage)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond)

	// Update the session with relocate
	initOriginID := originID
	originID = utils.GenUUID()
	updtUsage := time.Minute
	updtArgs := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrUpdate%s", originID),
			Event: map[string]interface{}{
				utils.OriginID:        originID,
				utils.InitialOriginID: initOriginID,
				utils.Usage:           updtUsage,
			},
		},
	}
	var rplyUpdt sessions.V1UpdateSessionReply
	if err := sCncrRPC.Call(utils.SessionSv1UpdateSession,
		updtArgs, &rplyUpdt); err != nil {
		t.Error(err)
	} else if rplyUpdt.MaxUsage == 0 {
		t.Errorf("unexpected MaxUsage at update: %v", rplyUpdt.MaxUsage)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond)

	// Terminate the session
	trmntArgs := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrTerminate%s", originID),
			Event: map[string]interface{}{
				utils.OriginID: originID,
				utils.Usage:    90 * time.Second,
			},
		},
	}
	var rplyTrmnt string
	if err := sCncrRPC.Call(utils.SessionSv1TerminateSession,
		trmntArgs, &rplyTrmnt); err != nil {
		t.Error(err)
	} else if rplyTrmnt != utils.OK {
		t.Errorf("received: <%s> to SessionSv1.Terminate", rplyTrmnt)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond)

	// processCDR
	argsCDR := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrCDR%s", originID),
			Event: map[string]interface{}{
				utils.OriginID: originID,
			},
		},
	}
	var rplyCDR string
	if err := sCncrRPC.Call(utils.SessionSv1ProcessCDR,
		argsCDR, &rplyCDR); err != nil {
		t.Error(err)
	} else if rplyCDR != utils.OK {
		t.Errorf("received: <%s> to ProcessCDR", rplyCDR)
	}
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	argCDRs := utils.RPCCDRsFilter{OriginIDs: []string{originID}}
	if err := sCncrRPC.Call(utils.APIerSv2GetCDRs, &argCDRs, &cdrs); err != nil {
		t.Error(err)
	} else if len(cdrs) != 1 {
		t.Errorf("unexpected number of CDRs returned: %d", len(cdrs))
	} else if cdrs[0].Usage != "1m30s" {
		t.Errorf("unexpected usage of CDR: %+v", cdrs[0])
	}
}
