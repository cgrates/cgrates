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
// 	"net/rpc"
// 	"os"
// 	"path"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/ees"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	amqpMCCCfgPath string
// 	amqpMCCCfg     *config.CGRConfig
// 	amqpMCCRPC     *rpc.Client
// 	amqpMCCConfDIR string //run tests for specific configuration
// 	amqpMCCDelay   int

// 	amqpMCCTests = []func(t *testing.T){
// 		// testCreateDirectory,
// 		testAMQPMCCLoadConfig,
// 		testAMQPMCCInitDataDb,
// 		testAMQPMCCResetStorDb,
// 		// testAMQPMCCStartEngine,
// 		testAMQPMCCRPCConn,
// 		testAMQPMCCProcessEvent,
// 		// testAMQPMCCStopEngine,
// 		// testCleanDirectory,
// 	}
// )

// // Test start here
// func TestAMQPMCC(t *testing.T) {
// 	amqpMCCConfDIR = "amqp_multiplecalls_internal"
// 	for _, stest := range amqpMCCTests {
// 		t.Run(amqpMCCConfDIR, stest)
// 	}
// }

// var exportPath = []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV", "/tmp/testCSVMasked",
// 	"/tmp/testCSVfromVirt", "/tmp/testCSVExpTemp"}

// func testCreateDirectory(t *testing.T) {
// 	for _, dir := range exportPath {
// 		if err := os.RemoveAll(dir); err != nil {
// 			t.Fatal("Error removing folder: ", dir, err)
// 		}
// 		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
// 			t.Fatal("Error creating folder: ", dir, err)
// 		}
// 	}
// }

// func testCleanDirectory(t *testing.T) {
// 	for _, dir := range exportPath {
// 		if err := os.RemoveAll(dir); err != nil {
// 			t.Fatal("Error removing folder: ", dir, err)
// 		}
// 	}
// }

// func testAMQPMCCLoadConfig(t *testing.T) {
// 	var err error
// 	amqpMCCCfgPath = path.Join(*dataDir, "conf", "samples", amqpMCCConfDIR)
// 	if amqpMCCCfg, err = config.NewCGRConfigFromPath(amqpMCCCfgPath); err != nil {
// 		t.Error(err)
// 	}
// 	amqpMCCDelay = 1000
// }

// func testAMQPMCCInitDataDb(t *testing.T) {
// 	if err := engine.InitDataDb(amqpMCCCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testAMQPMCCResetStorDb(t *testing.T) {
// 	if err := engine.InitStorDb(amqpMCCCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testAMQPMCCStartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(amqpMCCCfgPath, amqpMCCDelay); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testAMQPMCCRPCConn(t *testing.T) {
// 	var err error
// 	amqpMCCRPC, err = newRPCClient(amqpMCCCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal("Could not connect to rater: ", err.Error())
// 	}
// }

// func testAMQPMCCStopEngine(t *testing.T) {
// 	if err := engine.KillEngine(amqpMCCDelay); err != nil {
// 		t.Error(err)
// 	}
// }

// func exportCDR(idx string, client *rpc.Client, channel chan string, t *testing.T) {
// 	var reply map[string]map[string]interface{}
// 	if err := client.Call(utils.EeSv1ProcessEvent, &engine.CGREventWithEeIDs{
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "event" + idx,
// 			Event: map[string]interface{}{
// 				utils.RunID:        "run_" + idx,
// 				utils.CGRID:        "CGRID" + idx,
// 				utils.Tenant:       "cgrates.org",
// 				utils.Category:     "call",
// 				utils.ToR:          utils.MetaVoice,
// 				utils.OriginID:     "processCDR" + idx,
// 				utils.OriginHost:   "OriginHost" + idx,
// 				utils.RequestType:  utils.MetaPseudoPrepaid,
// 				utils.AccountField: "1001",
// 				utils.Destination:  "1002",
// 				utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
// 				utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
// 				utils.Usage:        2 * time.Minute,
// 			},
// 		},
// 	}, &reply); err != nil {
// 		channel <- err.Error()
// 		t.Error(err)
// 	} else {
// 		channel <- utils.ToJSON(reply)
// 	}

// }

// func testAMQPMCCProcessEvent(t *testing.T) {
// 	noOfExports := 2000
// 	channel := make(chan string, noOfExports)
// 	for i := 0; i < noOfExports; i++ {
// 		idxStr := strconv.Itoa(i)
// 		go func() {
// 			var reply map[string]map[string]interface{}
// 			if err := amqpMCCRPC.Call(utils.EeSv1ProcessEvent, &engine.CGREventWithEeIDs{
// 				CGREvent: &utils.CGREvent{
// 					Tenant: "cgrates.org",
// 					ID:     "event" + idxStr,
// 					Event: map[string]interface{}{
// 						utils.RunID:        "run_" + idxStr,
// 						utils.CGRID:        "CGRID" + idxStr,
// 						utils.Tenant:       "cgrates.org",
// 						utils.Category:     "call",
// 						utils.ToR:          utils.MetaVoice,
// 						utils.OriginID:     "processCDR" + idxStr,
// 						utils.OriginHost:   "OriginHost" + idxStr,
// 						utils.RequestType:  utils.MetaPseudoPrepaid,
// 						utils.AccountField: "1001",
// 						utils.Destination:  "1002",
// 						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
// 						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
// 						utils.Usage:        2 * time.Minute,
// 					},
// 				},
// 			}, &reply); err != nil {
// 				channel <- err.Error()
// 				t.Error(err)
// 			} else {
// 				channel <- utils.ToJSON(reply)
// 			}
// 		}()
// 	}
// 	for i := 0; i < noOfExports; i++ {
// 		chanStr := <-channel
// 		fmt.Println(chanStr)
// 	}
// 	time.Sleep(10 * time.Second)

// }

// func TestAMQPMC1CExport(t *testing.T) {
// 	var err error
// 	amqpMCCCfgPath = path.Join(*dataDir, "conf", "samples", "amqp_multiplecalls_internal")
// 	if amqpMCCCfg, err = config.NewCGRConfigFromPath(amqpMCCCfgPath); err != nil {
// 		t.Error(err)
// 	}
// 	exporter, err := ees.NewEventExporter(amqpMCCCfg.EEsCfg().Exporters[0], amqpMCCCfg, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	for i := 1; i <= 10000; i++ {
// 		idx := strconv.Itoa(i)
// 		cgrEv := &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "event" + idx,
// 			Event: map[string]interface{}{
// 				utils.RunID:        "run_" + idx,
// 				utils.CGRID:        "CGRID" + idx,
// 				utils.Tenant:       "cgrates.org",
// 				utils.Category:     "call",
// 				utils.ToR:          utils.MetaVoice,
// 				utils.OriginID:     "processCDR" + idx,
// 				utils.OriginHost:   "OriginHost" + idx,
// 				utils.RequestType:  utils.MetaPseudoPrepaid,
// 				utils.AccountField: "1001",
// 				utils.Destination:  "1002",
// 				utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
// 				utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
// 				utils.Usage:        2 * time.Minute,
// 			},
// 		}
// 		ev, err := exporter.PrepareMap(cgrEv)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		go ees.ExportWithAttempts(exporter, ev, "")
// 	}
// 	time.Sleep(time.Second)
// }
