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
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		testCreateDirectory,
		testCsvLoadConfig,
		testCsvResetDataDB,
		testCsvResetStorDb,
		testCsvStartEngine,
		testCsvRPCConn,
		testCsvExportEvent,
		testCsvVerifyExports,
		testCsvExportComposedEvent,
		testCsvVerifyComposedExports,
		testCsvExportMaskedDestination,
		testCsvVerifyMaskedDestination,
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
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.VOICE,
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
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dataEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.DATA,
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
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporter"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SMSEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.SMS,
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
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
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
	eCnt := "dbafe9c8614c785a65aabd116dd3959c3c56f7f6,192.168.1.1,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.01" +
		"\n" +
		"ea1f1968cc207859672c332364fc7614c86b04c5,192.168.1.1,*default,*data,abcdef,*rated,AnotherTenant,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,0.012" +
		"\n" +
		"2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4,192.168.1.1,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.15" +
		"\n"
	if outContent1, err := ioutil.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eCnt) != len(string(outContent1)) {
		t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportComposedEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterComposed"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:         utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:           utils.VOICE,
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
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterComposed"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SMSEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:         utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:           utils.SMS,
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
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
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
	eCnt := "NumberOfEvent,CGRID,RunID,ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage,Cost" + "\n" +
		"1,dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.0164" + "\n" +
		"2,2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.1555" + "\n" +
		"2,10s,1ns,1.1718" + "\n"
	if outContent1, err := ioutil.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportMaskedDestination(t *testing.T) {

	attrs := utils.AttrSetDestination{Id: "MASKED_DESTINATIONS", Prefixes: []string{"+4986517174963"}}
	var reply string
	if err := csvRpc.Call(utils.APIerSv1SetDestination, &attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVMaskedDestination"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.VOICE,
					utils.OriginID:     "dsafdsaf",
					utils.OriginHost:   "192.168.1.1",
					utils.RequestType:  utils.MetaRated,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "+4986517174963",
					utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
					utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
					utils.Usage:        10 * time.Second,
					utils.RunID:        utils.MetaDefault,
					utils.Cost:         1.01,
				},
			},
		},
	}
	var rply map[string]utils.MapStorage
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &rply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testCsvVerifyMaskedDestination(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testCSVMasked/", func(path string, info os.FileInfo, err error) error {
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
	eCnt := "dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,+4986517174***,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.01\n"
	if outContent1, err := ioutil.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportEventWithInflateTemplate(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.VOICE,
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
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dataEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.DATA,
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
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterWIthTemplate"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SMSEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.SMS,
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
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
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
	eCnt := "dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10000000000,1.01" +
		"\n" +
		"ea1f1968cc207859672c332364fc7614c86b04c5,*default,*data,abcdef,*rated,AnotherTenant,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,0.012" +
		"\n" +
		"2478e9f18ebcd3c684f3c14596b8bfeab2b0d6d4,*default,*sms,sdfwer,*rated,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,1,0.15" +
		"\n"
	if outContent1, err := ioutil.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testCsvExportNotFoundExporter(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"ExporterNotFound"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceEvent",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]interface{}{
					utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
					utils.ToR:          utils.VOICE,
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
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := csvRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}
