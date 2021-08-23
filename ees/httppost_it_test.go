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

package ees

import (
	"io"
	"net/http"
	"net/rpc"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	httpPostConfigDir string
	httpPostCfgPath   string
	httpPostCfg       *config.CGRConfig
	httpPostRpc       *rpc.Client
	httpValues        url.Values

	sTestsHTTPPost = []func(t *testing.T){
		testCreateDirectory,
		testHTTPPostLoadConfig,
		testHTTPPostResetDataDB,
		testHTTPPostResetStorDb,
		testHTTPPostStartEngine,
		testHTTPPostRPCConn,
		testHTTPStartHTTPServer,
		testHTTPExportEvent,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestHTTPPostExport(t *testing.T) {
	httpPostConfigDir = "ees"
	for _, stest := range sTestsHTTPPost {
		t.Run(httpPostConfigDir, stest)
	}
}

func testHTTPPostLoadConfig(t *testing.T) {
	var err error
	httpPostCfgPath = path.Join(*dataDir, "conf", "samples", httpPostConfigDir)
	if httpPostCfg, err = config.NewCGRConfigFromPath(httpPostCfgPath); err != nil {
		t.Error(err)
	}
}

func testHTTPPostResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(httpPostCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(httpPostCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpPostCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostRPCConn(t *testing.T) {
	var err error
	httpPostRpc, err = newRPCClient(httpPostCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testHTTPStartHTTPServer(t *testing.T) {
	http.HandleFunc("/event_http", func(writer http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Error(err)
		}
		httpValues, err = url.ParseQuery(string(b))
		if err != nil {
			t.Errorf("Cannot parse body: %s", string(b))
		}
		httpJsonHdr = r.Header.Clone()
	})
	go http.ListenAndServe(":12080", nil)
}

func testHTTPExportEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPPostExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
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
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPPostExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "dataEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("abcdef", time.Unix(1383813745, 0).UTC().String()),
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
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPPostExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("sdfwer", time.Unix(1383813745, 0).UTC().String()),
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
		},
	}

	eventSMSNoFields := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPPostExporterWithNoFields"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaSMS,
				utils.OriginID:     "sms2",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.RunID:        utils.MetaDefault,
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := httpPostRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPValues for eventVoice
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventVoice.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventVoice.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventVoice.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventVoice.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventVoice.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventVoice.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventVoice.Event[utils.Cost]),
	} {
		if rcv := httpValues.Get(key); rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader := "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}

	if err := httpPostRpc.Call(utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPValues for eventData
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventData.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventData.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventData.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventData.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventData.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventData.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventData.Event[utils.Cost]),
	} {
		if rcv := httpValues.Get(key); rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader = "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}

	if err := httpPostRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPValues for eventSMS
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventSMS.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventSMS.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventSMS.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventSMS.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventSMS.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventSMS.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventSMS.Event[utils.Cost]),
	} {
		if rcv := httpValues.Get(key); rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader = "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}
	if err := httpPostRpc.Call(utils.EeSv1ProcessEvent, eventSMSNoFields, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPValues for eventSMS
	for key, strVal := range eventSMSNoFields.Event {
		if rcv := httpValues.Get(key); rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
}
