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

// import (
// 	"fmt"
// 	"os"
// 	"path"
// 	"testing"

// 	"github.com/cgrates/birpc"
// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	testDblTpDifDestDir  string
// 	testDblTpDifDestPath string
// 	testDblTpDifDestCfg  *config.CGRConfig
// 	testDblTpDifDestRPC  *birpc.Client

// 	testDblTpDifDestTests = []func(t *testing.T){
// 		//testGenerateCsv,
// 		// testDblTpDifDestLoadConfig,
// 		// testDblTpDifDestResetDataDB,
// 		// testDblTpDifDestResetStorDb,
// 		// //testDblTpDifDestStartEngine,
// 		// testDblTpDifDestRPCConn,
// 		// testDblTpDifDestLoadersLoad,
// 		//testDblTpDifDestStopCgrEngine,
// 	}
// )

// func TestDblTpDifDest(t *testing.T) {
// 	switch *dbType {
// 	case utils.MetaInternal:
// 		testDblTpDifDestDir = "dblTpDifInternal"
// 	case utils.MetaMySQL:
// 		testDblTpDifDestDir = "dblTpDifMySql"
// 	case utils.MetaMongo:
// 		testDblTpDifDestDir = "dblTpDifInternalMongo"
// 	case utils.MetaPostgres:
// 		t.SkipNow()
// 	default:
// 		t.Fatal("Unknown Database type")
// 	}
// 	for _, testDblTpDifDest := range testDblTpDifDestTests {
// 		t.Run(testDblTpDifDestDir, testDblTpDifDest)
// 	}
// }

// func testDblTpDifDestLoadConfig(t *testing.T) {
// 	var err error
// 	testDblTpDifDestPath = path.Join(*dataDir, "conf", "samples", testDblTpDifDestDir)
// 	if testDblTpDifDestCfg, err = config.NewCGRConfigFromPath(testDblTpDifDestPath); err != nil {
// 		t.Error(err)
// 	}
// }

// func testDblTpDifDestResetDataDB(t *testing.T) {
// 	if err := engine.InitDataDB(testDblTpDifDestCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testDblTpDifDestResetStorDb(t *testing.T) {
// 	if err := engine.InitStorDB(testDblTpDifDestCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testDblTpDifDestStartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(testDblTpDifDestPath, *waitRater); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testDblTpDifDestRPCConn(t *testing.T) {
// 	var err error
// 	testDblTpDifDestRPC, err = newRPCClient(testDblTpDifDestCfg.ListenCfg())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testDblTpDifDestLoadersLoad(t *testing.T) {
// 	// var reply string
// 	// if err := testDblTpDifDestRPC.Call(context.Background(), utils.LoaderSv1Load,
// 	// 	&loaders.ArgsProcessFolder{
// 	// 		LoaderID:    "LoaderRatesTest",
// 	// 		StopOnError: true,
// 	// 		Caching:     utils.StringPointer(utils.MetaReload),
// 	// 		ForceLock:   true,
// 	// 	}, &reply); err != nil {
// 	// 	t.Error(err)
// 	// } else if reply != utils.OK {
// 	// 	t.Error("Unexpected reply returned:", reply)
// 	// }

// }
// func testGenerateCsv(t *testing.T) {
// 	filename := "/tmp/dat2"
// 	os.Remove(filename)
// 	f, _ := os.Create(filename)
// 	fmt.Fprintf(f, `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement`+"\n")
// 	n := 100000
// 	for i := 1; i <= n; i++ {
// 		fmt.Fprintf(f, `cgrates.org,RT_RETAIL%v,,,,,,RT_1,*prefix:~*req.Destination:%v,"* * * * *",,,0s,,0.4,1m,30s`, i, i)
// 		fmt.Fprintf(f, "\n")
// 		fmt.Fprintf(f, `cgrates.org,RT_RETAIL%v,,,,,,RT_1_CHRISTMAS,*prefix:~*req.Destination:%v,* * 24 12 *,;30,false,0s,,0.06,1m,1s`, i, i)
// 		fmt.Fprintf(f, "\n")
// 	}
// }
// func testDblTpDifDestStopCgrEngine(t *testing.T) {
// 	if err := engine.KillEngine(100); err != nil {
// 		t.Error(err)
// 	}
// }
