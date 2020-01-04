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
package general_tests

import (
	"flag"
	"fmt"
	"net/rpc"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sCncrCfgDIR, sCncrCfgPath string
	sCncrCfg                  *config.CGRConfig
	sCncrRPC                  *rpc.Client

	sCncrSessions = flag.Int("concurrent_sessions", 500000, "maximum concurrent sessions created")
	sCncrCps      = flag.Int("cps", 50000, "maximum requests per second sent out")

	cpsPool = make(chan struct{}, *sCncrCps)
)

// Tests starting here
func TestSCncrInternal(t *testing.T) {
	sCncrCfgDIR = "sessinternal"
	for _, tst := range sTestsSCncrIT {
		t.Run(sCncrCfgDIR, tst)
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
	if err := sCncrRPC.Call(utils.ApierV1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
}

func testSCncrRunSessions(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < *sCncrSessions; i++ {
		wg.Add(1)
		go func(y int) {
			cpsPool <- struct{}{} // push here up to cps
			go func() {           // allow more requests after a second
				time.Sleep(time.Duration(time.Second))
				<-cpsPool
			}()
			err := runSession(fmt.Sprintf("100%d", y))
			wg.Done()
			if err != nil {
				t.Error(err)
			}
		}(utils.RandomInteger(100, 200))
	}
	wg.Wait()
}

// runSession runs one session
func runSession(acntID string) (err error) {
	originID := utils.GenUUID() // each test with it's own OriginID

	// topup as much as we know we need
	topupDur := time.Duration(13000) * time.Hour
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     acntID,
		BalanceType: utils.VOICE,
		Balance: map[string]interface{}{
			utils.ID:     "testSCncr",
			utils.Value:  topupDur.Nanoseconds(),
			utils.Weight: 20,
		},
	}
	var reply string
	if err = sCncrRPC.Call(utils.ApierV1SetBalance,
		attrSetBalance, &reply); err != nil {
		return
	} else if reply != utils.OK {
		return fmt.Errorf("received: <%s> to ApierV1.SetBalance", reply)
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
				utils.Tenant:      "cgrates.org",
				utils.OriginID:    originID,
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     acntID,
				utils.Destination: fmt.Sprintf("%s%s", acntID, acntID),
				utils.SetupTime:   time.Now(),
				utils.Usage:       authDur,
			},
		},
	}
	var rplyAuth sessions.V1AuthorizeReply
	if err := sCncrRPC.Call(utils.SessionSv1AuthorizeEvent, authArgs, &rplyAuth); err != nil {
		return err
	}
	if rplyAuth.MaxUsage != authDur {
		return fmt.Errorf("unexpected MaxUsage: %v to auth", rplyAuth.MaxUsage)
	}
	time.Sleep(time.Duration(
		utils.RandomInteger(0, 100)) * time.Millisecond) // randomize between tests

	// Init the session
	initUsage := 90 * time.Second
	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     fmt.Sprintf("TestSCncrInit%s", originID),
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.OriginID:    originID,
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     acntID,
				utils.Destination: fmt.Sprintf("%s%s", acntID, acntID),
				utils.AnswerTime:  time.Now(),
				utils.Usage:       initUsage,
			},
		},
	}
	var rplyInit sessions.V1InitSessionReply
	if err := sCncrRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &rplyInit); err != nil {
		return err
	} else if rplyInit.MaxUsage != initUsage {
		return fmt.Errorf("unexpected MaxUsage at init: %v", rplyInit.MaxUsage)
	}
	return
}
