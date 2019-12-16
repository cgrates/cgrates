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
	"net/rpc"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	auCfgPath, usCfgPath, dspCfgPath string
	auCfg, usCfg, dspCfg             *config.CGRConfig
	auRPC, usRPC, dspRPC             *rpc.Client
	auEngine, usEngine, dspEngine    *exec.Cmd
	sTestsGOCS                       = []func(t *testing.T){
		testGOCSInitCfg,
		testGOCSResetDB,
		testGOCSStartEngine,
		testGOCSApierRpcConn,
		testGOCSLoadData,
		testGOCSAuthSession,
		testGOCSInitSession,
		testGOCSKillUSEngine,
		testGOCSUpdateSession,
		testGOCSStartUSEngine,
		testGOCSUpdateSession2,
		testGOCSTerminateSession,
		testGOCSProcessCDR,
		testGOCSStopCgrEngine,
	}
)

// Test start here
func TestGOCSIT(t *testing.T) {
	for _, stest := range sTestsGOCS {
		t.Run("TestGOCSIT", stest)
	}
}

//Init Config
func testGOCSInitCfg(t *testing.T) {
	auCfgPath = path.Join(*dataDir, "conf", "samples", "gocs", "au_site")
	if auCfg, err = config.NewCGRConfigFromPath(auCfgPath); err != nil {
		t.Fatal(err)
	}
	auCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(auCfg)
	usCfgPath = path.Join(*dataDir, "conf", "samples", "gocs", "us_site")
	if usCfg, err = config.NewCGRConfigFromPath(usCfgPath); err != nil {
		t.Fatal(err)
	}
	dspCfgPath = path.Join(*dataDir, "conf", "samples", "gocs", "dsp_site")
	if dspCfg, err = config.NewCGRConfigFromPath(dspCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testGOCSResetDB(t *testing.T) {
	if err := engine.InitDataDb(auCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(usCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(dspCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testGOCSStartEngine(t *testing.T) {
	if usEngine, err = engine.StopStartEngine(usCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if auEngine, err = engine.StartEngine(auCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if dspEngine, err = engine.StartEngine(dspCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

}

// Connect rpc client to rater
func testGOCSApierRpcConn(t *testing.T) {
	if auRPC, err = newRPCClient(auCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	if usRPC, err = newRPCClient(usCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	if dspRPC, err = newRPCClient(dspCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testGOCSLoadData(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "gocs", "us_site")}
	var loadInst utils.LoadInstance
	if err := usRPC.Call(utils.ApierV2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "gocs", "au_site")}
	if err := auRPC.Call(utils.ApierV2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups on au_site
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "gocs", "dsp_site")}
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspCfgPath, "-path", attrs.FolderPath)

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(1 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}
}

func testGOCSAuthSession(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := dspRPC.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
}

func testGOCSInitSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := dspRPC.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	// give a bit of time to session to be replicate
	time.Sleep(10 * time.Millisecond)

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := auRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	} else if aSessions[0].NodeID != "AU_SITE" {
		t.Errorf("Expecting : %+v, received: %+v", "AU_SITE", aSessions[0].NodeID)
	} else if aSessions[0].MaxCostSoFar != 1.0008 {
		t.Errorf("Expecting : %+v, received: %+v", 1.0008, aSessions[0].MaxCostSoFar)
	} else if aSessions[0].Usage != time.Duration(5*time.Minute) {
		t.Errorf("Expecting : %+v, received: %+v", time.Duration(5*time.Minute), aSessions[0].MaxCostSoFar)
	}

	aSessions = make([]*sessions.ExternalSession, 0)
	if err := usRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	} else if aSessions[0].NodeID != "US_SITE" {
		t.Errorf("Expecting : %+v, received: %+v", "US_SITE", aSessions[0].NodeID)
	} else if aSessions[0].MaxCostSoFar != 1.0008 {
		t.Errorf("Expecting : %+v, received: %+v", 1.0008, aSessions[0].MaxCostSoFar)
	} else if aSessions[0].Usage != time.Duration(5*time.Minute) {
		t.Errorf("Expecting : %+v, received: %+v", time.Duration(5*time.Minute), aSessions[0].Usage)
	}

	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}

	// 10 - 1.0008 = 8.9992
	if err := auRPC.Call(utils.ApierV2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 8.9992 {
		t.Errorf("Expecting : %+v, received: %+v", 8.9992, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	if err := usRPC.Call(utils.ApierV2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 8.9992 {
		t.Errorf("Expecting : %+v, received: %+v", 8.9992, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

}

func testGOCSKillUSEngine(t *testing.T) {
	if err := usEngine.Process.Kill(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func testGOCSUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
			},
		},
	}
	var rply sessions.V1UpdateSessionReply

	// right now dispatcher receive utils.ErrPartiallyExecuted
	// in case of of engines fails
	if err := dspRPC.Call(utils.SessionSv1UpdateSession, args, &rply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("Expecting : %+v, received: %+v", utils.ErrPartiallyExecuted, err)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := auRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	} else if aSessions[0].NodeID != "AU_SITE" {
		t.Errorf("Expecting : %+v, received: %+v", "AU_SITE", aSessions[0].NodeID)
	} else if aSessions[0].MaxCostSoFar != 1.5017999999999998 {
		t.Errorf("Expecting : %+v, received: %+v", 1.5017999999999998, aSessions[0].MaxCostSoFar)
	} else if aSessions[0].Usage != time.Duration(10*time.Minute) {
		t.Errorf("Expecting : %+v, received: %+v", time.Duration(5*time.Minute), aSessions[0].Usage)
	}

	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}

	// balanced changed in AU_SITE
	if err := auRPC.Call(utils.ApierV2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 8.4982 {
		t.Errorf("Expecting : %+v, received: %+v", 8.4982, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

}

func testGOCSStartUSEngine(t *testing.T) {
	if usEngine, err = engine.StartEngine(usCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if usRPC, err = newRPCClient(usCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testGOCSUpdateSession2(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession2",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
			},
		},
	}
	var rply sessions.V1UpdateSessionReply

	if err := dspRPC.Call(utils.SessionSv1UpdateSession, args, &rply); err != nil {
		t.Errorf("Expecting : %+v, received: %+v", nil, err)
	} else if rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := auRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	} else if aSessions[0].NodeID != "AU_SITE" {
		t.Errorf("Expecting : %+v, received: %+v", "AU_SITE", aSessions[0].NodeID)
	} else if aSessions[0].MaxCostSoFar != 2.0027999999999997 {
		t.Errorf("Expecting : %+v, received: %+v", 2.0027999999999997, aSessions[0].MaxCostSoFar)
	} else if aSessions[0].Usage != time.Duration(15*time.Minute) {
		t.Errorf("Expecting : %+v, received: %+v", time.Duration(15*time.Minute), aSessions[0].Usage)
	}

	aSessions = make([]*sessions.ExternalSession, 0)
	if err := usRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	} else if aSessions[0].NodeID != "US_SITE" {
		t.Errorf("Expecting : %+v, received: %+v", "US_SITE", aSessions[0].NodeID)
	} else if aSessions[0].MaxCostSoFar != 1.0008 {
		t.Errorf("Expecting : %+v, received: %+v", 1.0008, aSessions[0].MaxCostSoFar)
	} else if aSessions[0].Usage != time.Duration(5*time.Minute) {
		t.Errorf("Expecting : %+v, received: %+v", time.Duration(5*time.Minute), aSessions[0].Usage)
	}

	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}

	// because the session don't exist on US_SITE
	// this update behaves as an init
	// 8.9992 - 1.0008 = 7.9984
	if err := auRPC.Call(utils.ApierV2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 7.9984 {
		t.Errorf("Expecting : %+v, received: %+v", 7.9984, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	if err := usRPC.Call(utils.ApierV2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 7.9984 {
		t.Errorf("Expecting : %+v, received: %+v", 7.9984, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func testGOCSTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testGOCSTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       15 * time.Minute,
			},
		},
	}
	var rply string
	if err := dspRPC.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := auRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error %s received error %v and reply %s", utils.ErrNotFound, err, utils.ToJSON(aSessions))
	}
	if err := usRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error %s received error %v and reply %s", utils.ErrNotFound, err, utils.ToJSON(aSessions))
	}
}

func testGOCSProcessCDR(t *testing.T) {
	args := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessCDR",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.Category:    "call",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply string
	if err := usRPC.Call(utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testGOCSStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
