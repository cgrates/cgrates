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

package ers

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	jsonCfgPath string
	jsonCfgDIR  string
	jsonCfg     *config.CGRConfig
	jsonRPC     *rpc.Client

	fileContent = `
{
	"Tenant": "cgrates.org",
	"Account": "voiceAccount",
	"AnswerTime": "2018-08-24T16:00:26Z",
	"SetupTime": "2018-08-24T16:00:26Z",
	"Destination": "+4986517174963",
	"OriginHost": "192.168.1.1",
	"OriginID": "testJsonCDR",
	"RequestType": "*pseudoprepaid",
	"Source": "jsonFile",
	"Usage": 120000000000
}`
	jsonTests = []func(t *testing.T){
		testCreateDirs,
		testJSONInitConfig,
		testJSONInitCdrDb,
		testJSONResetDataDb,
		testJSONStartEngine,
		testJSONRpcConn,
		testJSONAddData,
		testJSONHandleFile,
		testJSONVerify,
		testCleanupFiles,
		testJSONKillEngine,
	}
)

func TestJSONReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		jsonCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		jsonCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		jsonCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		jsonCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, test := range jsonTests {
		t.Run(jsonCfgDIR, test)
	}
}

func testJSONInitConfig(t *testing.T) {
	var err error
	jsonCfgPath = path.Join(*dataDir, "conf", "samples", jsonCfgDIR)
	if jsonCfg, err = config.NewCGRConfigFromPath(jsonCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testJSONInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(jsonCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testJSONResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(jsonCfg); err != nil {
		t.Fatal(err)
	}
}

func testJSONStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(jsonCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testJSONRpcConn(t *testing.T) {
	var err error
	jsonRPC, err = newRPCClient(jsonCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testJSONAddData(t *testing.T) {
	var reply string
	//add a charger
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "Default",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	if err := jsonRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	attrSetAcnt := v2.AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "voiceAccount",
	}
	if err := jsonRPC.Call(utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "voiceAccount",
		BalanceType: utils.VOICE,
		Value:       600000000000,
		Balance: map[string]interface{}{
			utils.ID:        utils.MetaDefault,
			"RatingSubject": "*zero1m",
			utils.Weight:    10.0,
		},
	}
	if err := jsonRPC.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	if err := jsonRPC.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 600000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}
}

// The default scenario, out of ers defined in .cfg file
func testJSONHandleFile(t *testing.T) {
	fileName := "file1.json"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ErsJSON/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testJSONVerify(t *testing.T) {
	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			OriginIDs: []string{"testJsonCDR"},
		},
	}
	if err := jsonRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != 2*time.Minute {
			t.Errorf("Unexpected usage for CDR: %d", cdrs[0].Usage)
		}
	}

	var acnt *engine.Account
	if err := jsonRPC.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.VOICE][0].Value != 480000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.VOICE][0])
	}
}

func testJSONKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
