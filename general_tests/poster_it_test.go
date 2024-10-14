//go:build integration
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
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	pstrCfg       *config.CGRConfig
	pstrRpc       *birpc.Client
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
	switch *utils.DBType {
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
	if *utils.Encoding == utils.MetaGOB {
		pstrConfigDIR += "_gob"
	}

	for _, stest := range sTestsPosterIT {
		t.Run(pstrConfigDIR, stest)
	}
}

func testPosterITInitCfg(t *testing.T) {
	pstrCfgPath = path.Join(*utils.DataDir, "conf", "samples", pstrConfigDIR)
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
	// before starting the engine, create the directories needed for failed posts or
	// clear their contents if they exist already
	if err := os.RemoveAll(pstrCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Fatal("Error removing folder: ", pstrCfg.GeneralCfg().FailedPostsDir, err)
	}
	if err := os.MkdirAll(pstrCfg.GeneralCfg().FailedPostsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StopStartEngine(pstrCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testPosterITRpcConn(t *testing.T) {
	pstrRpc = engine.NewRPCClient(t, pstrCfg.ListenCfg())
}

func testPosterReadFolder(format string) (expEv *ees.ExportEvents, err error) {
	filesInDir, _ := os.ReadDir(pstrCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		err = fmt.Errorf("No files in directory: %s", pstrCfg.GeneralCfg().FailedPostsDir)
		return
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(pstrCfg.GeneralCfg().FailedPostsDir, fileName)

		expEv, err = ees.NewExportEventsFromFile(filePath)
		if err != nil {
			return
		}
		if expEv.Type == format {
			return
		}
	}
	err = fmt.Errorf("Format not found")
	return
}

func testPosterITSetAccount(t *testing.T) {
	var reply string
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1SetAccount, pstrAccount, &reply); err != nil {
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
	if err := pstrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
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
	var acc map[string]any
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
	if err := pstrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
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
	var acc map[string]any
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
	if err := pstrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
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
	var acc map[string]any
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
	if err := pstrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
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
	var acc map[string]any
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
	if err := pstrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: pstrAccount.Tenant, Account: pstrAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := pstrRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
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
	var acc map[string]any
	if err := json.Unmarshal(body, &acc); err != nil {
		t.Fatal(err)
	}
	if acc[utils.AccountField] != utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account) {
		t.Errorf("Expected %q ,received %q", utils.ConcatenatedKey(pstrAccount.Tenant, pstrAccount.Account), acc[utils.AccountField])
	}
}

func testPosterITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
