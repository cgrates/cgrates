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
	tpConfigDIR string //run tests for specific configuration

	sTestsTP = []func(t *testing.T){
		testTPInitCfg,
		testTPResetStorDb,
		testTPStartEngine,
		testTPRpcConn,
		testTPImportTPFromFolderPath,
		testTPExportTPToFolder,
		testTPExportTPToFolderWithError,
		testTPKillEngine,
	}
)

//Test start here
func TestTPIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	tpConfigDIR = "tutmysql"
	for _, stest := range sTestsTP {
		t.Run(tpConfigDIR, stest)
	}
}

func testTPInitCfg(t *testing.T) {
	var err error
	tpCfgPath = path.Join(*dataDir, "conf", "samples", tpConfigDIR)
	tpCfg, err = config.NewCGRConfigFromPath(tpCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Wipe out the cdr database
func testTPResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(tpCfg); err != nil {
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
	tpRPC, err = newRPCClient(tpCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPImportTPFromFolderPath(t *testing.T) {
	var reply string
	if err := tpRPC.Call(utils.APIerSv1ImportTariffPlanFromFolder,
		utils.AttrImportTPFromFolder{TPid: "TEST_TPID2",
			FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ImportTarrifPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.ImportTarrifPlanFromFolder got reply: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testTPExportTPToFolder(t *testing.T) {
	var reply *utils.ExportedTPStats
	expectedTPStas := &utils.ExportedTPStats{
		Compressed: true,
		ExportPath: "/tmp/",
		ExportedFiles: []string{utils.RatingProfilesCsv, utils.RatingPlansCsv, utils.ActionsCsv, utils.AccountActionsCsv,
			utils.ChargersCsv, utils.ActionPlansCsv, utils.ResourcesCsv, utils.StatsCsv, utils.ThresholdsCsv,
			utils.DestinationsCsv, utils.RatesCsv, utils.DestinationRatesCsv, utils.FiltersCsv, utils.RoutesCsv, utils.AttributesCsv},
	}
	sort.Strings(expectedTPStas.ExportedFiles)
	tpid := "TEST_TPID2"
	compress := true
	exportPath := "/tmp/"
	if err := tpRPC.Call(utils.APIerSv1ExportTPToFolder, &utils.AttrDirExportTP{TPid: &tpid, ExportPath: &exportPath, Compress: &compress}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExportTPToFolder: ", err.Error())
	} else if !reflect.DeepEqual(reply.ExportPath, expectedTPStas.ExportPath) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.ExportPath, reply.ExportPath)
	} else if !reflect.DeepEqual(reply.Compressed, expectedTPStas.Compressed) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.Compressed, reply.Compressed)
	} else if sort.Strings(reply.ExportedFiles); !reflect.DeepEqual(expectedTPStas.ExportedFiles, reply.ExportedFiles) {
		t.Errorf("Expecting : %+v, received: %+v", expectedTPStas.ExportedFiles, reply.ExportedFiles)
	}
}

func testTPExportTPToFolderWithError(t *testing.T) {
	var reply *utils.ExportedTPStats
	tpid := "UnexistedTP"
	compress := true
	exportPath := "/tmp/"
	if err := tpRPC.Call(utils.APIerSv1ExportTPToFolder,
		&utils.AttrDirExportTP{TPid: &tpid, ExportPath: &exportPath, Compress: &compress}, &reply); err == nil || err.Error() != utils.NewErrServerError(utils.ErrNotFound).Error() {
		t.Error("Expecting error, received: ", err)
	}

}

func testTPKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
