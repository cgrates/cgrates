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

package v1

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpCfgPath   string
	tpCfg       *config.CGRConfig
	tpRPC       *rpc.Client
	tpDataDir   = "/usr/share/cgrates"
	tpConfigDIR string //run tests for specific configuration
)

var sTestsTP = []func(t *testing.T){
	testTPInitCfg,
	testTPResetStorDb,
	testTPStartEngine,
	testTPRpcConn,
	testTPImportTPFromFolderPath,
	testTPExportTPToFolder,
	testTPKillEngine,
}

//Test start here
func TestTPITMySql(t *testing.T) {
	tpConfigDIR = "tutmysql"
	for _, stest := range sTestsTP {
		t.Run(tpConfigDIR, stest)
	}
}

func TestTPITMongo(t *testing.T) {
	tpConfigDIR = "tutmongo"
	for _, stest := range sTestsTP {
		t.Run(tpConfigDIR, stest)
	}
}

func TestTPITPG(t *testing.T) {
	tpConfigDIR = "tutpostgres"
	for _, stest := range sTestsTP {
		t.Run(tpConfigDIR, stest)
	}
}

func testTPInitCfg(t *testing.T) {
	var err error
	tpCfgPath = path.Join(tpDataDir, "conf", "samples", tpConfigDIR)
	tpCfg, err = config.NewCGRConfigFromPath(tpCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpCfg.DataFolderPath = tpDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpCfg)
}

// Wipe out the cdr database
func testTPResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRpcConn(t *testing.T) {
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPImportTPFromFolderPath(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV1.ImportTariffPlanFromFolder",
		utils.AttrImportTPFromFolder{TPid: "TEST_TPID2",
			FolderPath: path.Join(tpDataDir, "tariffplans", "tutorial")}, &reply); err != nil {
		t.Error("Got error on ApierV1.ImportTarrifPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ImportTarrifPlanFromFolder got reply: ", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testTPExportTPToFolder(t *testing.T) {
	var reply *utils.ExportedTPStats
	expectedTPStas := &utils.ExportedTPStats{
		Compressed: true,
		ExportPath: "/tmp/",
		ExportedFiles: []string{utils.RATING_PROFILES_CSV, utils.RATING_PLANS_CSV, utils.ACTIONS_CSV, utils.ACCOUNT_ACTIONS_CSV,
			utils.ChargersCsv, utils.TIMINGS_CSV, utils.ACTION_PLANS_CSV, utils.ResourcesCsv, utils.StatsCsv, utils.ThresholdsCsv,
			utils.DESTINATIONS_CSV, utils.RATES_CSV, utils.DESTINATION_RATES_CSV, utils.FiltersCsv, utils.SuppliersCsv, utils.AttributesCsv},
	}
	sort.Strings(expectedTPStas.ExportedFiles)
	tpid := "TEST_TPID2"
	compress := true
	exportPath := "/tmp/"
	if err := tpRPC.Call("ApierV1.ExportTPToFolder", &utils.AttrDirExportTP{TPid: &tpid, ExportPath: &exportPath, Compress: &compress}, &reply); err != nil {
		t.Error("Got error on ApierV1.ExportTPToFolder: ", err.Error())
	} else if !reflect.DeepEqual(reply.ExportPath, expectedTPStas.ExportPath) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.ExportPath, reply.ExportPath)
	} else if !reflect.DeepEqual(reply.Compressed, expectedTPStas.Compressed) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.Compressed, reply.Compressed)
	} else if sort.Strings(reply.ExportedFiles); !reflect.DeepEqual(expectedTPStas.ExportedFiles, reply.ExportedFiles) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.ExportedFiles, reply.ExportedFiles)
	}
	time.Sleep(500 * time.Millisecond)

}

func testTPKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
