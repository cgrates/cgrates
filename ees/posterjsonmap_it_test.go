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
	"encoding/json"
	"net/http"
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	httpJSONMapConfigDir string
	httpJSONMapCfgPath   string
	httpJSONMapCfg       *config.CGRConfig
	httpJSONMapRpc       *rpc.Client
	httpJsonMap          map[string]string
	httpJsonHdr          http.Header

	sTestsHTTPJsonMap = []func(t *testing.T){
		testCreateDirectory,
		testHTTPJsonMapLoadConfig,
		testHTTPJsonMapResetDataDB,
		testHTTPJsonMapResetStorDb,
		testHTTPJsonMapStartEngine,
		testHTTPJsonMapRPCConn,
		testHTTPJsonMapStartHTTPServer,
		testHTTPJsonMapExportEvent,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestHTTPJsonMapExport(t *testing.T) {
	httpJSONMapConfigDir = "ees"
	for _, stest := range sTestsHTTPJsonMap {
		t.Run(httpJSONMapConfigDir, stest)
	}
}

func testHTTPJsonMapLoadConfig(t *testing.T) {
	var err error
	httpJSONMapCfgPath = path.Join(*dataDir, "conf", "samples", httpJSONMapConfigDir)
	if httpJSONMapCfg, err = config.NewCGRConfigFromPath(httpJSONMapCfgPath); err != nil {
		t.Error(err)
	}
}

func testHTTPJsonMapResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(httpJSONMapCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(httpJSONMapCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpJSONMapCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapRPCConn(t *testing.T) {
	var err error
	httpJSONMapRpc, err = newRPCClient(httpJSONMapCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapStartHTTPServer(t *testing.T) {
	http.HandleFunc("/event_json_map_http", func(writer http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&httpJsonMap)
		r.Body.Close()
		if err != nil {
			t.Error(err)
		}
		httpJsonHdr = r.Header.Clone()
	})

	go http.ListenAndServe(":12081", nil)
}

func testHTTPJsonMapExportEvent(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPJsonMapExporter"},
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
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.CDR,
			},
		},
	}

	eventData := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPJsonMapExporter"},
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
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.CDR,
			},
		},
	}

	eventSMS := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPJsonMapExporter"},
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
				utils.OrderID:      10,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.CDR,
			},
		},
	}

	eventSMSNoFields := &utils.CGREventWithEeIDs{
		EeIDs: []string{"HTTPJsonMapExporterWithNoFields"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "SMSEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("sms2", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaSMS,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.RunID:        utils.MetaDefault,
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := httpJSONMapRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPJsonMap for eventVoice
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventVoice.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventVoice.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventVoice.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventVoice.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventVoice.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventVoice.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventVoice.Event[utils.Cost]),
		utils.EventType:    utils.CDR,
	} {
		if rcv := httpJsonMap[key]; rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader := "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}
	if err := httpJSONMapRpc.Call(utils.EeSv1ProcessEvent, eventData, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPJsonMap for eventData
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventData.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventData.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventData.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventData.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventData.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventData.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventData.Event[utils.Cost]),
		utils.EventType:    utils.CDR,
	} {
		if rcv := httpJsonMap[key]; rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader = "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}
	if err := httpJSONMapRpc.Call(utils.EeSv1ProcessEvent, eventSMS, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPJsonMap for eventSMS
	for key, strVal := range map[string]string{
		utils.CGRID:        utils.IfaceAsString(eventSMS.Event[utils.CGRID]),
		utils.ToR:          utils.IfaceAsString(eventSMS.Event[utils.ToR]),
		utils.Category:     utils.IfaceAsString(eventSMS.Event[utils.Category]),
		utils.AccountField: utils.IfaceAsString(eventSMS.Event[utils.AccountField]),
		utils.Subject:      utils.IfaceAsString(eventSMS.Event[utils.Subject]),
		utils.Destination:  utils.IfaceAsString(eventSMS.Event[utils.Destination]),
		utils.Cost:         utils.IfaceAsString(eventSMS.Event[utils.Cost]),
		utils.EventType:    utils.CDR,
	} {
		if rcv := httpJsonMap[key]; rcv != strVal {
			t.Errorf("Expected %+v, received: %+v", strVal, rcv)
		}
	}
	expHeader = "http://www.cgrates.org"
	if len(httpJsonHdr["Origin"]) == 0 || httpJsonHdr["Origin"][0] != expHeader {
		t.Errorf("Expected %+v, received: %+v", expHeader, httpJsonHdr["Origin"])
	}

	if err := httpJSONMapRpc.Call(utils.EeSv1ProcessEvent, eventSMSNoFields, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	// verify HTTPJsonMap for eventSMS
	for key, strVal := range eventSMSNoFields.Event {
		if rcv := httpJsonMap[key]; rcv != strVal {
			t.Errorf("Expected %q, received: %q", strVal, rcv)
		}
	}
}
