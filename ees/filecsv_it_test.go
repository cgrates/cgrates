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

package ees

import (
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	csvConfigDir string
	csvCfgPath   string
	csvCfg       *config.CGRConfig
	csvRpc       *rpc.Client

	sTestsCsv = []func(t *testing.T){
		testCsvCreateDirectory,
		testCsvLoadConfig,
		testCsvResetDataDB,
		testCsvResetStorDb,
		testCsvStartEngine,
		testCsvRPCConn,
		testCsvExportEvent,
		testCsvStopCgrEngine,
		//testCsvCleanDirectory,
	}
)

func TestCsvExport(t *testing.T) {
	csvConfigDir = "ees"
	for _, stest := range sTestsCsv {
		t.Run(csvConfigDir, stest)
	}
}

func testCsvCreateDirectory(t *testing.T) {
	for _, dir := range []string{"/tmp/testExport"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCsvLoadConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", csvConfigDir)
	if csvCfg, err = config.NewCGRConfigFromPath(csvCfgPath); err != nil {
		t.Error(err)
	}
}

func testCsvResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func testCsvResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func testCsvStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCsvRPCConn(t *testing.T) {
	var err error
	csvRpc, err = newRPCClient(csvCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testCsvExportEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithOpts{
		CGREventWithArgDispatcher: &utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:       utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "dsafdsaf",
					utils.OriginHost:  "192.168.1.1",
					utils.RequestType: utils.META_RATED,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "1002",
					utils.SetupTime:   time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:  time.Unix(1383813746, 0).UTC(),
					utils.Usage:       time.Duration(10) * time.Second,
					utils.RunID:       utils.MetaDefault,
					utils.Cost:        1.01,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
			},
		},
	}

	eventData := &utils.CGREventWithOpts{
		CGREventWithArgDispatcher: &utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dataEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:       utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:         utils.DATA,
					utils.OriginID:    "abcdef",
					utils.OriginHost:  "192.168.1.1",
					utils.RequestType: utils.META_RATED,
					utils.Tenant:      "AnotherTenant",
					utils.Category:    "call", //for data CDR use different Tenant
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "1002",
					utils.SetupTime:   time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:  time.Unix(1383813746, 0).UTC(),
					utils.Usage:       time.Duration(10) * time.Nanosecond,
					utils.RunID:       utils.MetaDefault,
					utils.Cost:        0.012,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
			},
		},
	}

	eventSMS := &utils.CGREventWithOpts{
		CGREventWithArgDispatcher: &utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SMSEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:       utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:         utils.SMS,
					utils.OriginID:    "sdfwer",
					utils.OriginHost:  "192.168.1.1",
					utils.RequestType: utils.META_RATED,
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "1002",
					utils.SetupTime:   time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:  time.Unix(1383813746, 0).UTC(),
					utils.Usage:       time.Duration(1),
					utils.RunID:       utils.MetaDefault,
					utils.Cost:        0.15,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
			},
		},
	}
	var reply string
	if err := csvRpc.Call(utils.EventExporterSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v, received: %+v", utils.OK, reply)
	}
	if err := csvRpc.Call(utils.EventExporterSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v, received: %+v", utils.OK, reply)
	}
	if err := csvRpc.Call(utils.EventExporterSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v, received: %+v", utils.OK, reply)
	}
	time.Sleep(1 * time.Second)
}

func testCsvStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testCsvCleanDirectory(t *testing.T) {
	for _, dir := range []string{"/tmp/testExport"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}
