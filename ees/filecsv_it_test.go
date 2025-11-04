//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	csvConfigDir string
	csvCfgPath   string
	csvCfg       *config.CGRConfig
	csvRpc       *birpc.Client

	sTestsCsv = []func(t *testing.T){
		testCreateDirectory,
		testCsvLoadConfig,
		testCsvResetDBs,

		testCsvStartEngine,
		testCsvRPCConn,
		testCsvExportEvent,
		testCsvVerifyExports,
		testCsvExportComposedEvent,
		testCsvVerifyComposedExports,
		testCsvExportBufferedEvent,
		testCsvExportBufferedEventNoExports,
		testCsvExportEventWithInflateTemplate,
		testCsvVerifyExportsWithInflateTemplate,
		testCsvExportNotFoundExporter,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestCsvExport(t *testing.T) {
	csvConfigDir = "ees"
	for _, stest := range sTestsCsv {
		t.Run(csvConfigDir, stest)
	}
}

func testCsvLoadConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*utils.DataDir, "conf", "samples", csvConfigDir)
	if csvCfg, err = config.NewCGRConfigFromPath(context.Background(), csvCfgPath); err != nil {
		t.Error(err)
	}
}

func testCsvResetDBs(t *testing.T) {
	if err := engine.InitDB(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func testCsvStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCsvRPCConn(t *testing.T) {
	csvRpc = engine.NewRPCClient(t, csvCfg.ListenCfg(), *utils.Encoding)
}

func testCsvExportEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaData,
				utils.OriginID:     "abcdef",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "AnotherTenant",
				utils.Category:     "call", //for data CDR use different Tenant
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Nanosecond,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.012,
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]any{
				utils.ToR:          utils.MetaSMS,
				utils.OriginID:     "sdfwer",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        1,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.15,
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testCsvVerifyExports(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testCSV/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "192.168.1.1,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.01" +
		"\n" +
		"192.168.1.1,*default,*data,abcdef,*rated,AnotherTenant,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,0.012" +
		"\n" +
		"192.168.1.1,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.15" +
		"\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eCnt) != len(string(outContent1)) {
		t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportComposedEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterComposed"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{
				utils.ToR:           utils.MetaVoice,
				"ComposedOriginID1": "dsaf",
				"ComposedOriginID2": "dsaf",
				utils.OriginHost:    "192.168.1.1",
				utils.RequestType:   utils.MetaRated,
				utils.Tenant:        "cgrates.org",
				utils.Category:      "call",
				utils.AccountField:  "1001",
				utils.Subject:       "1001",
				utils.Destination:   "1002",
				utils.SetupTime:     time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:    time.Unix(1383813746, 0).UTC(),
				utils.Usage:         10 * time.Second,
				utils.RunID:         utils.MetaDefault,
				utils.Cost:          1.016374,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterComposed"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]any{
				utils.ToR:           utils.MetaSMS,
				"ComposedOriginID1": "sdf",
				"ComposedOriginID2": "wer",
				utils.OriginHost:    "192.168.1.1",
				utils.RequestType:   utils.MetaRated,
				utils.Tenant:        "cgrates.org",
				utils.Category:      "call",
				utils.AccountField:  "1001",
				utils.Subject:       "1001",
				utils.Destination:   "1002",
				utils.SetupTime:     time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:    time.Unix(1383813746, 0).UTC(),
				utils.Usage:         1,
				utils.RunID:         utils.MetaDefault,
				utils.Cost:          0.155462,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testCsvVerifyComposedExports(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testComposedCSV/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "NumberOfEvent,*originID,RunID,ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage,Cost" + "\n" +
		"1,dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.0164" + "\n" +
		"2,2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.1555" + "\n" +
		"2,10s,1ns,1.1718" + "\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportBufferedEvent(t *testing.T) {
	eventVoice := &ArchiveEventsArgs{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaExporterID: "CSVExporterBuffered",
			utils.MetaUsage:      123 * time.Nanosecond,
		},
		Events: []*utils.EventsWithOpts{
			{
				Event: map[string]any{

					utils.ToR:           utils.MetaVoice,
					"ComposedOriginID1": "dsaf",
					"ComposedOriginID2": "dsaf",
					utils.OriginHost:    "192.168.1.1",
					utils.RequestType:   utils.MetaRated,
					utils.Tenant:        "cgrates.org",
					utils.Category:      "call",
					utils.AccountField:  "1005",
					utils.Subject:       "1001",
					utils.Destination:   "1002",
					utils.SetupTime:     time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:    time.Unix(1383813746, 0).UTC(),
					utils.Usage:         10 * time.Second,
					utils.RunID:         utils.MetaDefault,
					utils.Cost:          1.016374,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
				Opts: map[string]any{
					utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.MetaChargers: true,
					utils.MetaRunID:    "random_runID",
				},
			},
			{
				Event: map[string]any{

					utils.ToR:          utils.MetaData,
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaRated,
					utils.Tenant:       "AnotherTenant",
					utils.Category:     "call", //for data CDR use different Tenant
					utils.AccountField: "1005",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
					utils.Usage:        10 * time.Nanosecond,
					utils.RunID:        utils.MetaDefault,
					utils.Cost:         0.012,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
				Opts: map[string]any{
					utils.MetaOriginID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
					utils.MetaUsage:    200 * time.Second,
				},
			},
			// this one will not match, because opts got another another ExporterID and it will be changed from the initial opt
			{
				Event: map[string]any{

					utils.AccountField: "1005",
					utils.Subject:      "1005",
					utils.Destination:  "103",
					utils.Usage:        1760 * time.Nanosecond,
					utils.RunID:        "Default_charging_id22",
					utils.Cost:         0,
				},
				Opts: map[string]any{
					utils.MetaOriginID:   utils.Sha1("qwertyiopuu", time.Unix(1383813745, 0).UTC().String()),
					utils.MetaExporterID: "CSVExporterBuffered_CHanged",
				},
			},
			{
				Event: map[string]any{

					utils.ToR:           utils.MetaData,
					"ComposedOriginID1": "abcdefghh",
					utils.RequestType:   utils.MetaNone,
					utils.Tenant:        "phone.org",
					utils.Category:      "sms", //for data CDR use different Tenant
					utils.AccountField:  "1005",
					utils.Subject:       "User2001",
					utils.Destination:   "User2002",
					utils.SetupTime:     time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:    time.Unix(1383813746, 0).UTC(),
					utils.Usage:         10 * time.Nanosecond,
					utils.RunID:         "raw",
					utils.Cost:          44.5,
					"ExtraFields": map[string]string{"extra1": "val_extra1",
						"extra2": "val_extra2", "extra3": "val_extra3"},
				},
				Opts: map[string]any{
					utils.MetaOriginID: utils.Sha1("nlllo", time.Unix(1383813745, 0).UTC().String()),
					utils.MetaRates:    true,
				},
			},
			{
				Event: map[string]any{

					utils.OriginHost:   "127.0.0.1",
					utils.RequestType:  utils.MetaPrepaid,
					utils.Tenant:       "dispatchers.org",
					utils.Category:     "photo", //for data CDR use different Tenant
					utils.AccountField: "1005",
					utils.Subject:      "1005",
					utils.Destination:  "1000",
					utils.SetupTime:    time.Unix(22383813745, 0).UTC(),
					utils.AnswerTime:   time.Unix(22383813760, 0).UTC(),
					utils.Usage:        10 * time.Nanosecond,
					utils.RunID:        "Default_charging_id",
					utils.Cost:         1.442234,
				},
				Opts: map[string]any{
					utils.MetaOriginID:  utils.Sha1("qwert", time.Unix(1383813745, 0).UTC().String()),
					utils.MetaStartTime: time.Date(2020, time.January, 7, 16, 60, 0, 0, time.UTC),
				},
			},
		},
	}
	var reply []byte
	if err := csvRpc.Call(context.Background(), utils.EeSv1ArchiveEventsInReply,
		eventVoice, &reply); err != nil {
		t.Error(err)
	}

	rdr, err := zip.NewReader(bytes.NewReader(reply), int64(len(reply)))
	if err != nil {
		t.Error(err)
	}
	csvRply := make([][]string, 6)
	for _, f := range rdr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		info := csv.NewReader(rc)
		info.FieldsPerRecord = -1
		csvRply, err = info.ReadAll()
		if err != nil {
			t.Error(err)
		}
		rc.Close()
	}

	expected := [][]string{
		{"NumberOfEvent", "*originID", "RunID", "ToR", "OriginID", "RequestType", "Tenant", "Category", "Account", "Subject", "Destination", "SetupTime", "AnswerTime", "Usage", "Cost"},
		{"1", "dbafe9c8614c785a65aabd116dd3959c3c56f7f6", "*default", "*voice", "dsafdsaf", "*rated", "cgrates.org", "call", "1005", "1001", "1002", "2013-11-07T08:42:25Z", "2013-11-07T08:42:26Z", "10000000000", "1.0164"},
		{"2", "ea1f1968cc207859672c332364fc7614c86b04c5", "*default", "*data", "", "*rated", "AnotherTenant", "call", "1005", "1001", "1002", "2013-11-07T08:42:25Z", "2013-11-07T08:42:26Z", "10", "0.012"},
		{"3", "9e0b2a4b23e0843efe522e8a611b092a16ecfba1", "raw", "*data", "abcdefghh", "*none", "phone.org", "sms", "1005", "User2001", "User2002", "2013-11-07T08:42:25Z", "2013-11-07T08:42:26Z", "10", "44.5"},
		{"4", "cd8112998c2abb0e4a7cd3a94c74817cd5fe67d3", "Default_charging_id", "", "", "*prepaid", "dispatchers.org", "photo", "1005", "1005", "1000", "2679-04-25T22:02:25Z", "2679-04-25T22:02:40Z", "10", "1.4422"},
		{"4", "10s", "46.9706"},
	}
	if !reflect.DeepEqual(expected, csvRply) {
		t.Errorf("Expected %+v \n received %+v", utils.ToJSON(expected), utils.ToJSON(csvRply))
	}

	time.Sleep(time.Second)
}

func testCsvExportBufferedEventNoExports(t *testing.T) {
	// in this case, exported does not exist in config
	eventVoice := &ArchiveEventsArgs{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaExporterID: "InexistentExport",
		},
		Events: []*utils.EventsWithOpts{
			{
				Event: map[string]any{
					utils.AccountField: "not_exported_Acc",
				},
			},
		},
	}
	var reply []byte
	expectedErr := "exporter config with ID: InexistentExport is missing"
	if err := csvRpc.Call(context.Background(), utils.EeSv1ArchiveEventsInReply,
		eventVoice, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %q \n received %q", utils.ToJSON(expectedErr), utils.ToJSON(err))
	}

	// in this case, exporter exists but the events will not match our filters (filter for Account)
	eventVoice = &ArchiveEventsArgs{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaExporterID: "CSVExporterBuffered",
		},
		Events: []*utils.EventsWithOpts{
			{
				Event: map[string]any{

					utils.ToR:           utils.MetaVoice,
					"ComposedOriginID1": "dsaf",
					"ComposedOriginID2": "dsaf",
					utils.OriginHost:    "192.168.1.1",
					utils.RequestType:   utils.MetaRated,
					utils.Tenant:        "cgrates.org",
					utils.Category:      "call",
					utils.AccountField:  "DifferentAccount12",
				},
				Opts: map[string]any{
					utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				},
			},
			{
				Event: map[string]any{

					utils.ToR:          utils.MetaData,
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaRated,
					utils.Tenant:       "AnotherTenant",
					utils.Category:     "call", //for data CDR use different Tenant
					utils.AccountField: "DifferentAccount10",
				},
				Opts: map[string]any{
					utils.MetaOriginID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
				},
			},
		},
	}
	expectedErr = "SERVER_ERROR: NO EXPORTS"
	if err := csvRpc.Call(context.Background(), utils.EeSv1ArchiveEventsInReply,
		eventVoice, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %q \n received %q", utils.ToJSON(expectedErr), utils.ToJSON(err))
	}
}

func testCsvExportEventWithInflateTemplate(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaData,
				utils.OriginID:     "abcdef",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "AnotherTenant",
				utils.Category:     "call", //for data CDR use different Tenant
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Nanosecond,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.012,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaSMS,
				utils.OriginID:     "sdfwer",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        1,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         0.15,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testCsvVerifyExportsWithInflateTemplate(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testCSVExpTemp/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.01,CSVExporterWIthTemplate" +
		"\n" +
		"ea1f1968cc207859672c332364fc7614c86b04c5,*default,*data,abcdef,*rated,AnotherTenant,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,0.012,CSVExporterWIthTemplate" +
		"\n" +
		"2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.15,CSVExporterWIthTemplate" +
		"\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportNotFoundExporter(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ExporterNotFound"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func TestCsvInitFileCSV(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = "/tmp/TestInitFileCSV"
	if err := os.MkdirAll("/tmp/TestInitFileCSV", 0666); err != nil {
		t.Error(err)
	}
	em := utils.NewExporterMetrics("", time.Local)
	fCsv := &FileCSVee{
		cgrCfg: cgrCfg,
		cfg:    cgrCfg.EEsCfg().Exporters[0],
		em:     em,
	}
	if err := fCsv.init(nil); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll("/tmp/TestInitFileCSV"); err != nil {
		t.Error(err)
	}
}
