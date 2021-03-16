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
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	pstrCfg       *config.CGRConfig
	pstrRpc       *rpc.Client
	pstrCfgPath   string
	pstrConfigDIR string

	sTestsPosterIT = []func(t *testing.T){
		testPosterITInitCfg,
		testPosterITInitCdrDb,
		testPosterITStartEngine,
		testPosterITRpcConn,
		testPosterITSetAccount,

		testPosterITAMQP,
		testPosterITAMQPv1,
		testPosterITSQS,
		testPosterITS3,
		testPosterITKafka,

		testPosterITStopCgrEngine,
	}
	pstrAccount = &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
)

func TestPosterIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		pstrConfigDIR = "actions_internal"
	case utils.MetaMySQL:
		pstrConfigDIR = "actions_mysql"
	case utils.MetaMongo:
		pstrConfigDIR = "actions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *encoding == utils.MetaGOB {
		pstrConfigDIR += "_gob"
	}

	for _, stest := range sTestsPosterIT {
		t.Run(pstrConfigDIR, stest)
	}
}

func testPosterITInitCfg(t *testing.T) {
	pstrCfgPath = path.Join(*dataDir, "conf", "samples", pstrConfigDIR)
	var err error
	pstrCfg, err = config.NewCGRConfigFromPath(pstrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testPosterITInitCdrDb(t *testing.T) {
	if err := engine.InitDataDb(pstrCfg); err != nil { // need it for versions
		t.Fatal(err)
	}
	if err := engine.InitStorDb(pstrCfg); err != nil {
		t.Fatal(err)
	}
}

func testPosterITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(pstrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testPosterITRpcConn(t *testing.T) {
	var err error
	pstrRpc, err = newRPCClient(pstrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testPosterReadFolder(format string) (expEv *engine.ExportEvents, err error) {
	filesInDir, _ := os.ReadDir(pstrCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		err = fmt.Errorf("No files in directory: %s", pstrCfg.GeneralCfg().FailedPostsDir)
		return
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(pstrCfg.GeneralCfg().FailedPostsDir, fileName)

		expEv, err = engine.NewExportEventsFromFile(filePath)
		if err != nil {
			return
		}
		if expEv.Format == format {
			return
		}
	}
	err = fmt.Errorf("Format not found")
	return
}

func testPosterITSetAccount(t *testing.T) {
	var reply string
	if err := pstrRpc.Call(utils.APIerSv1SetAccount, pstrAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
}

func testPosterITAMQP(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACT_AMQP",
		Actions: []*utils.TPAction{ // set a action with a wrong endpoint to easily check if it was executed
			{Identifier: utils.MetaExport, ExtraParameters: "amqp_fail"},
		},
	}
	if err := pstrRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	// verify if acction was executed
	time.Sleep(100 * time.Millisecond)
	ev, err := testPosterReadFolder(utils.MetaAMQPjsonMap)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Events) != 1 {
		t.Fatalf("Expected 1 event received: %d events", len(ev.Events))
	}
	body := ev.Events[0].([]byte)
	var acc map[string]interface{}
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITAMQPv1(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACT_AMQPv1",
		Actions: []*utils.TPAction{ // set a action with a wrong endpoint to easily check if it was executed
			{Identifier: utils.MetaExport, ExtraParameters: "aws_fail"},
		},
	}
	if err := pstrRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	// verify if acction was executed
	time.Sleep(150 * time.Millisecond)
	ev, err := testPosterReadFolder(utils.MetaAMQPV1jsonMap)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Events) != 1 {
		t.Fatalf("Expected 1 event received: %d events", len(ev.Events))
	}
	body := ev.Events[0].([]byte)
	var acc map[string]interface{}
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITSQS(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACT_SQS",
		Actions: []*utils.TPAction{ // set a action with a wrong endpoint to easily check if it was executed
			{Identifier: utils.MetaExport, ExtraParameters: "sqs_fail"},
		},
	}
	if err := pstrRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	// verify if acction was executed
	time.Sleep(100 * time.Millisecond)
	ev, err := testPosterReadFolder(utils.MetaSQSjsonMap)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Events) != 1 {
		t.Fatalf("Expected 1 event received: %d events", len(ev.Events))
	}
	body := ev.Events[0].([]byte)
	var acc map[string]interface{}
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITS3(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACT_S3",
		Actions: []*utils.TPAction{ // set a action with a wrong endpoint to easily check if it was executed
			{Identifier: utils.MetaExport, ExtraParameters: "s3_fail"},
		},
	}
	if err := pstrRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	// verify if acction was executed
	time.Sleep(100 * time.Millisecond)
	ev, err := testPosterReadFolder(utils.MetaS3jsonMap)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Events) != 1 {
		t.Fatalf("Expected 1 event received: %d events", len(ev.Events))
	}
	body := ev.Events[0].([]byte)
	var acc map[string]interface{}
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITKafka(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "ACT_Kafka",
		Actions: []*utils.TPAction{ // set a action with a wrong endpoint to easily check if it was executed
			{Identifier: utils.MetaExport, ExtraParameters: "kafka_fail"},
		},
	}
	if err := pstrRpc.Call(utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	// verify if acction was executed
	time.Sleep(500 * time.Millisecond)
	ev, err := testPosterReadFolder(utils.MetaKafkajsonMap)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev.Events) != 1 {
		t.Fatalf("Expected 1 event received: %d events", len(ev.Events))
	}
	body := ev.Events[0].([]byte)
	var acc map[string]interface{}
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
