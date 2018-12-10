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
package sessions

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var smgRplcMasterCfgPath, smgRplcSlaveCfgPath string
var smgRplcMasterCfg, smgRplcSlaveCfg *config.CGRConfig
var smgRplcMstrRPC, smgRplcSlvRPC *rpc.Client

func TestSMGRplcInitCfg(t *testing.T) {
	smgRplcMasterCfgPath = path.Join(*dataDir, "conf", "samples", "smgreplcmaster")
	if smgRplcMasterCfg, err = config.NewCGRConfigFromFolder(smgRplcMasterCfgPath); err != nil {
		t.Fatal(err)
	}
	smgRplcMasterCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(smgRplcMasterCfg)
	smgRplcSlaveCfgPath = path.Join(*dataDir, "conf", "samples", "smgreplcslave")
	if smgRplcSlaveCfg, err = config.NewCGRConfigFromFolder(smgRplcSlaveCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestSMGRplcResetDB(t *testing.T) {
	if err := engine.InitDataDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGRplcStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil { // Start slave before master
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGRplcApierRpcConn(t *testing.T) {
	if smgRplcMstrRPC, err = jsonrpc.Dial("tcp", smgRplcMasterCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = jsonrpc.Dial("tcp", smgRplcSlaveCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGRplcTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := smgRplcMstrRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGRplcInitiate(t *testing.T) {
	var pSessions []*ActiveSession
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions",
		nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	var maxUsage time.Duration
	var reply string
	if err := smgRplcMstrRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &reply); err == nil ||
		err.Error() != rpcclient.ErrSessionNotFound.Error() { // Update should return rpcclient.ErrSessionNotFound
		t.Error(err)
	}
	if err := smgRplcMstrRPC.Call(utils.SMGenericV2InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(90*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.OriginID: "123451"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(90)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions",
		map[string]string{utils.OriginID: "123451"}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage != time.Duration(90*time.Second) {
		t.Errorf("PassiveSession: %+v", pSessions[0])
	}
}

// Update on slave
func TestSMGRplcUpdate(t *testing.T) {
	smgEv := map[string]interface{}{
		utils.EVENT_NAME: "TEST_EVENT",
		utils.OriginID:   "123451",
		utils.Usage:      "1m",
	}
	var maxUsage time.Duration
	if err := smgRplcSlvRPC.Call(utils.SMGenericV2UpdateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	} else if maxUsage != time.Duration(time.Minute) {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.OriginID: "123451"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	var pSessions []*ActiveSession
	// Make sure we don't have passive session on active host
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions", nil,
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// Master should not longer have activeSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.OriginID: "123451"}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	cgrID := GetSetCGRID(engine.NewSafEvent(smgEv))
	// Make sure session was replicated
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetPassiveSessions",
		nil, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].CGRID != cgrID {
		t.Errorf("PassiveSession: %+v", pSessions[0])
	} else if pSessions[0].Usage != time.Duration(150*time.Second) {
		t.Errorf("PassiveSession: %+v", pSessions[0])
	}

}

func TestSMGRplcTerminate(t *testing.T) {
	smgEv := map[string]interface{}{
		utils.EVENT_NAME: "TEST_EVENT",
		utils.OriginID:   "123451",
		utils.Usage:      "3m",
	}
	var reply string
	if err := smgRplcMstrRPC.Call("SMGenericV1.TerminateSession", smgEv, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.OriginID: "123451"}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.OriginID: "123451"}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	var pSessions map[string][]*SMGSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetPassiveSessions",
		nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions",
		nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestSMGRplcManualReplicate(t *testing.T) {
	masterProc, err := engine.StopStartEngine(smgRplcMasterCfgPath, *waitRater)
	if err != nil { // Kill both and start Master
		t.Fatal(err)
	}
	if smgRplcMstrRPC, err = jsonrpc.Dial("tcp", smgRplcMasterCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	smgEv1 := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	smgEv2 := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123481",
		utils.Direction:   utils.OUT,
		utils.Account:     "1002",
		utils.Subject:     "1002",
		utils.Destination: "1005",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	for _, smgEv := range []map[string]interface{}{smgEv1, smgEv2} {
		var maxUsage time.Duration
		if err := smgRplcMstrRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
			t.Error(err)
		}
		if maxUsage != time.Duration(90*time.Second) {
			t.Error("Bad max usage: ", maxUsage)
		}
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(90)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	// Start slave, should not have any active session at beginning
	slave, err := engine.StartEngine(smgRplcSlaveCfgPath, *waitRater)
	if err != nil {
		t.Fatal(err)
	}
	if err := slave.Process.Kill(); err != nil { // restart the slave
		t.Error(err)
	}
	if _, err := engine.StartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = jsonrpc.Dial("tcp", smgRplcSlaveCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions", nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	argsRepl := ArgsReplicateSessions{Connections: []*config.HaPoolConfig{
		{
			Address:     smgRplcSlaveCfg.ListenCfg().RPCJSONListen,
			Transport:   utils.MetaJSONrpc,
			Synchronous: true},
	}}
	var repply string
	if err := smgRplcMstrRPC.Call("SMGenericV1.ReplicateActiveSessions", argsRepl, &repply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	if err := smgRplcSlvRPC.Call("SMGenericV1.GetPassiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(90)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	// kill master
	if err := masterProc.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process, error: %v", err.Error())
	}
	var status map[string]interface{}
	if err := smgRplcMstrRPC.Call("Responder.Status", "", &status); err == nil { // master should not longer be reachable
		t.Error(err, status)
	}
	if err := smgRplcSlvRPC.Call("Responder.Status", "", &status); err != nil { // slave should be still operational
		t.Error(err)
	}
	// start master
	if _, err := engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if smgRplcMstrRPC, err = jsonrpc.Dial("tcp", smgRplcMasterCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	// Master should have no session active/passive
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetPassiveSessions", nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	// recover passive sessions from slave
	argsRepl = ArgsReplicateSessions{Connections: []*config.HaPoolConfig{
		{
			Address:     smgRplcMasterCfg.ListenCfg().RPCJSONListen,
			Transport:   utils.MetaJSONrpc,
			Synchronous: true},
	}}
	if err := smgRplcSlvRPC.Call("SMGenericV1.ReplicatePassiveSessions", argsRepl, &repply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	// Master should have no session active/passive
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcMstrRPC.Call("SMGenericV1.GetPassiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(90)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}

}

func TestSMGRplcStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
