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
package sessionmanager

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
	if !*testIntegration {
		return
	}
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
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGRplcStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil { // Start slave before master
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGRplcApierRpcConn(t *testing.T) {
	if !*testIntegration {
		return
	}
	if smgRplcMstrRPC, err = jsonrpc.Dial("tcp", smgRplcMasterCfg.RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = jsonrpc.Dial("tcp", smgRplcSlaveCfg.RPCJSONListen); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGRplcTPFromFolder(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := smgRplcMstrRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGRplcInitiate(t *testing.T) {
	if !*testIntegration {
		return
	}
	var pSessions []*ActiveSession
	if err := smgRplcSlvRPC.Call("SMGenericV1.PassiveSessions", nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123451",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1004",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m30s",
	}
	var maxUsage float64
	if err := smgRplcMstrRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err == nil && err.Error() != rpcclient.ErrSessionNotFound.Error() { // Update should return rpcclient.ErrSessionNotFound
		t.Error(err)
	}
	var reply string
	if err := smgRplcMstrRPC.Call("SMGenericV1.TerminateSession", smgEv, &reply); err == nil && err.Error() != rpcclient.ErrSessionNotFound.Error() { // Update should return rpcclient.ErrSessionNotFound
		t.Error(err)
	}
	if err := smgRplcMstrRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 90 {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.ActiveSessions", map[string]string{utils.ACCID: "123451"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(90)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.PassiveSessions", map[string]string{utils.ACCID: "123451"}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage != time.Duration(90*time.Second) {
		t.Errorf("PassiveSession: %+v", aSessions[0])
	}
}

// Update on slave
func TestSMGRplcUpdate(t *testing.T) {
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME: "TEST_EVENT",
		utils.ACCID:      "123451",
		utils.USAGE:      "1m",
	}
	var maxUsage float64
	if err := smgRplcSlvRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	} else if maxUsage != 60 {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcSlvRPC.Call("SMGenericV1.ActiveSessions", map[string]string{utils.ACCID: "123451"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	var pSessions []*ActiveSession
	// Make sure we don't have passive session on active host
	if err := smgRplcSlvRPC.Call("SMGenericV1.PassiveSessions", nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// Master should not longer have activeSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.ActiveSessions", map[string]string{utils.ACCID: "123451"}, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	cgrID := smgEv.GetCGRID(utils.META_DEFAULT)
	// Make sure session was replicated
	if err := smgRplcMstrRPC.Call("SMGenericV1.PassiveSessions", nil, &pSessions); err != nil {
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
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME: "TEST_EVENT",
		utils.ACCID:      "123451",
		utils.USAGE:      "3m",
	}
	var reply string
	if err := smgRplcMstrRPC.Call("SMGenericV1.TerminateSession", smgEv, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ActiveSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.ActiveSessions", map[string]string{utils.ACCID: "123451"}, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.ActiveSessions", map[string]string{utils.ACCID: "123451"}, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	var pSessions map[string][]*SMGSession
	if err := smgRplcMstrRPC.Call("SMGenericV1.PassiveSessions", nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := smgRplcSlvRPC.Call("SMGenericV1.PassiveSessions", nil, &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestSMGRplcStopCgrEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
