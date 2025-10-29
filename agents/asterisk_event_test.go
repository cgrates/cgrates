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
package agents

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	stasisStart                = `{"application":"cgrates_auth","type":"StasisStart","timestamp":"2016-09-12T13:53:48.919+0200","args":["cgr_reqtype=*prepaid","cgr_route=supplier1", "extra1=val1", "extra2=val2"],"channel":{"id":"1473681228.6","state":"Ring","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelStateChange         = `{"application":"cgrates_auth","type":"ChannelStateChange","timestamp":"2016-09-12T13:53:52.110+0200","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelAnsweredDestroyed   = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T13:54:27.335+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"},"cause":16}`
	channelUnansweredDestroyed = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T18:00:18.121+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473696018.2","state":"Ring","name":"PJSIP/1002-00000002","caller":{"name":"1002","number":"1002"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T18:00:18.109+0200"},"cause":16}`
	channelBusyDestroyed       = `{"type":"ChannelDestroyed","timestamp":"2016-09-13T12:59:48.806+0200","application":"cgrates_auth","cause_txt":"User busy","channel":{"id":"1473764378.3","state":"Ring","name":"PJSIP/1001-00000002","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":4},"creationtime":"2016-09-13T12:59:38.839+0200"},"cause":17}`
)

func TestSMAParseStasisArgs(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	expAppArgs := map[string]string{"cgr_reqtype": "*prepaid", "cgr_route": "supplier1", "extra1": "val1", "extra2": "val2"}
	if !reflect.DeepEqual(smaEv.cachedFields, expAppArgs) {
		t.Errorf("Expecting: %+v, received: %+v", smaEv.cachedFields, expAppArgs)
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	expAppArgs = map[string]string{}
	if !reflect.DeepEqual(smaEv.cachedFields, expAppArgs) {
		t.Errorf("Expecting: %+v, received: %+v", smaEv.cachedFields, expAppArgs)
	}
}

func TestSMAEventType(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.EventType() != "StasisStart" {
		t.Error("Received:", smaEv.EventType())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.EventType() != "" {
		t.Error("Received:", smaEv.EventType())
	}
}

func TestSMAEventChannelID(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.ChannelID() != "1473681228.6" {
		t.Error("Received:", smaEv.ChannelID())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.ChannelID() != "" {
		t.Error("Received:", smaEv.ChannelID())
	}
}

func TestSMAEventOriginatorIP(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.OriginatorIP() != "127.0.0.1" {
		t.Error("Received:", smaEv.OriginatorIP())
	}
}

func TestSMAEventAccount(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Account() != "1001" {
		t.Error("Received:", smaEv.Account())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Account() != "" {
		t.Error("Received:", smaEv.Account())
	}
}

func TestSMAEventDestination(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Destination() != "1002" {
		t.Error("Received:", smaEv.Destination())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Destination() != "" {
		t.Error("Received:", smaEv.Destination())
	}
}

func TestSMAEventTimestamp(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Timestamp() != "2016-09-12T13:53:48.919+0200" {
		t.Error("Received:", smaEv.Timestamp())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Timestamp() != "" {
		t.Error("Received:", smaEv.Timestamp())
	}
}

func TestSMAEventChannelState(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.ChannelState() != "Up" {
		t.Error("Received:", smaEv.ChannelState())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.ChannelState() != "" {
		t.Error("Received:", smaEv.ChannelState())
	}
}

func TestSMASetupTime(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.SetupTime() != "2016-09-12T13:53:48.918+0200" {
		t.Error("Received:", smaEv.SetupTime())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.SetupTime() != "" {
		t.Error("Received:", smaEv.SetupTime())
	}
}

func TestSMAEventRequestType(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.RequestType() != "*prepaid" {
		t.Error("Received:", smaEv.RequestType())
	}
	ev = make(map[string]any) // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.RequestType() != config.CgrConfig().GeneralCfg().DefaultReqType {
		t.Error("Received:", smaEv.RequestType())
	}
}

func TestSMAEventTenant(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Tenant() != "" {
		t.Error("Received:", smaEv.Tenant())
	}
	ev = map[string]any{"args": []any{"cgr_tenant=cgrates.org"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Tenant() != "cgrates.org" {
		t.Error("Received:", smaEv.Tenant())
	}
}

func TestSMAEventCategory(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Category() != "" {
		t.Error("Received:", smaEv.Category())
	}
	ev = map[string]any{"args": []any{"cgr_category=premium_call"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Category() != "premium_call" {
		t.Error("Received:", smaEv.Category())
	}
}

func TestSMAEventSubject(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Subject() != "" {
		t.Error("Received:", smaEv.Subject())
	}
	ev = map[string]any{"args": []any{"cgr_subject=dan"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Subject() != "dan" {
		t.Error("Received:", smaEv.Subject())
	}
}

func TestSMAEventPDD(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.PDD() != "" {
		t.Error("Received:", smaEv.PDD())
	}
	ev = map[string]any{"args": []any{"cgr_pdd=2.1"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.PDD() != "2.1" {
		t.Error("Received:", smaEv.PDD())
	}
}

func TestSMAEventRoute(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Route() != "supplier1" {
		t.Error("Received:", smaEv.Route())
	}
	ev = map[string]any{"args": []any{"cgr_route=supplier1"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.Route() != "supplier1" {
		t.Error("Received:", smaEv.Route())
	}
}

func TestSMAEventDisconnectCause(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.DisconnectCause() != "" {
		t.Error("Received:", smaEv.DisconnectCause())
	}
	ev = map[string]any{"args": []any{"cgr_disconnectcause=NORMAL_DISCONNECT"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.DisconnectCause() != "NORMAL_DISCONNECT" {
		t.Error("Received:", smaEv.DisconnectCause())
	}

	ev = map[string]any{"cause": 16} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.DisconnectCause() != "16" {
		t.Error("Received:", smaEv.DisconnectCause())
	}
	ev = map[string]any{"cause_txt": "NORMAL_DISCONNECT"} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.DisconnectCause() != "NORMAL_DISCONNECT" {
		t.Error("Received:", smaEv.DisconnectCause())
	}
}

func TestSMAEventExtraParameters(t *testing.T) {
	expExtraParams := map[string]string{
		"extra1": "val1",
		"extra2": "val2",
	}
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.PDD() != "" {
		t.Error("Received:", smaEv.PDD())
	}
	if extraParams := smaEv.ExtraParameters(); !reflect.DeepEqual(expExtraParams, extraParams) {
		t.Errorf("Expecting: %+v, received: %+v", expExtraParams, extraParams)
	}
}

func TestSMAEventV1AuthorizeArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	cgrEv, err := smaEv.AsCGREvent(timezone)
	if err != nil {
		t.Error(err)
	}
	exp := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    cgrEv,
	}
	if rcv := smaEv.V1AuthorizeArgs(); !reflect.DeepEqual(exp.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", exp.GetMaxUsage, rcv.GetMaxUsage)
	}

	stasisStart2 := `{"type":"StasisStart","timestamp":"2018-11-25T05:03:26.464-0500","args":["cgr_reqtype=*prepaid","cgr_route=route1","cgr_flags=*accounts+*attributes+*resources+*stats+*routes+*thresholds"],"channel":{"id":"1543140206.0","dialplan":{"context":"internal","exten":"1002","priority":4},"caller":{"name":"","number":"1001"},"name":"PJSIP/1001-00000000","state":"Ring","connected":{"name":"","number":""},"language":"en","accountcode":"","creationtime":"2018-11-25T05:03:26.463-0500"},"asterisk_id":"08:00:27:b7:b8:1f","application":"cgrates_auth"}`
	var ev2 map[string]any
	if err := json.Unmarshal([]byte(stasisStart2), &ev2); err != nil {
		t.Error(err)
	}
	smaEv2 := NewSMAsteriskEvent(ev2, "127.0.0.1", "")
	//smaEv2.parseStasisArgs()
	cgrEv2, err := smaEv2.AsCGREvent(timezone)
	if err != nil {
		t.Error(err)
	}

	exp2 := &sessions.V1AuthorizeArgs{
		GetAttributes:      true,
		AuthorizeResources: true,
		GetMaxUsage:        true,
		ProcessThresholds:  true,
		ProcessStats:       true,
		GetRoutes:          true,
		CGREvent:           cgrEv2,
	}
	if rcv := smaEv2.V1AuthorizeArgs(); !reflect.DeepEqual(exp2.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", exp2.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(exp2.AuthorizeResources, rcv.AuthorizeResources) {
		t.Errorf("Expecting: %+v, received: %+v", exp2.AuthorizeResources, rcv.AuthorizeResources)
	} else if !reflect.DeepEqual(exp2.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", exp2.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(exp2.ProcessThresholds, rcv.ProcessThresholds) {
		t.Errorf("Expecting: %+v, received: %+v", exp2.ProcessThresholds, rcv.ProcessThresholds)
	} else if !reflect.DeepEqual(exp2.ProcessStats, rcv.ProcessStats) {
		t.Errorf("Expecting: %+v, received: %+v", exp2.ProcessStats, rcv.ProcessStats)
	}
}

func TestSMAEventV1InitSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AsteriskEvent",
		Event: map[string]any{
			"MissingCGRSubsustems": "",
		},
	}
	exp := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent:    cgrEv,
	}
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if rcv := smaEv.V1InitSessionArgs(*cgrEv); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	exp2 := &sessions.V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		CGREvent:          cgrEv,
	}
	cgrEv.Event[utils.CGRFlags] = "*resources+*accounts+*attributes"
	if rcv := smaEv.V1InitSessionArgs(*cgrEv); !reflect.DeepEqual(exp2, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp2), utils.ToJSON(rcv))
	}
}

func TestSMAEventV1TerminateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AsteriskEvent",
		Event: map[string]any{
			"MissingCGRSubsustems": "",
		},
	}
	exp := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent:         cgrEv,
	}
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if rcv := smaEv.V1TerminateSessionArgs(*cgrEv); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	exp2 := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		ReleaseResources: true,
		ProcessStats:     true,
		CGREvent:         cgrEv,
	}
	cgrEv.Event[utils.CGRFlags] = "*resources+*accounts+*stats"
	if rcv := smaEv.V1TerminateSessionArgs(*cgrEv); !reflect.DeepEqual(exp2, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp2), utils.ToJSON(rcv))
	}
}

func TestRequestType(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.RequestType() != "*prepaid" {
		t.Error("Received:", smaEv.RequestType())
	}
	ev = make(map[string]any) // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1", "")
	if smaEv.RequestType() != config.CgrConfig().GeneralCfg().DefaultReqType {
		t.Error("Received:", smaEv.RequestType())
	}
}

func TestSMAsteriskEventUpdateCGREvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event:  make(map[string]any),
	}

	testCases := []struct {
		smaEv     *SMAsteriskEvent
		wantErr   bool
		eventName string
	}{
		{
			smaEv: &SMAsteriskEvent{
				cachedFields: map[string]string{
					eventType: ARIChannelStateChange,
					timestamp: "2023-05-05T10:00:00Z",
				},
			},
			wantErr:   false,
			eventName: SMASessionStart,
		},
		{
			smaEv: &SMAsteriskEvent{
				cachedFields: map[string]string{
					eventType: ARIChannelDestroyed,
					timestamp: "2023-05-05T10:00:00Z",
				},
			},
			wantErr:   false,
			eventName: SMASessionTerminate,
		},
	}

	for _, testCase := range testCases {
		err := testCase.smaEv.UpdateCGREvent(cgrEv)
		if testCase.wantErr && err == nil {
			t.Fatal("expected an error, but got no error")
		}
		if !testCase.wantErr && err != nil {
			t.Fatalf("expected no error, but got error: %v", err)
		}

		if !testCase.wantErr {
			if eventName, ok := cgrEv.Event[utils.EventName]; !ok || eventName != testCase.eventName {
				t.Fatalf("expected event name: %s, but got: %v", testCase.eventName, eventName)
			}
		}

	}
}

func TestRestoreAndUpdateFieldsOk(t *testing.T) {
	ariEv := map[string]any{
		"application": "cgrates_auth",
		"asterisk_id": "08:00:27:18:d8:cf",
		"cause":       "16",
		"cause_txt":   "Normal Clearing",
		"channel": map[string]any{
			"accountcode": "",
			"caller": map[string]any{
				"name":   "1001",
				"number": "1001",
			},
			"channelvars": map[string]any{
				"CDR(answer)":  "2024-05-03 08:53:06",
				"CDR(billsec)": "4",
				"cgr_flags":    "*accounts *attributes *resources *stats *routes *thresholds cgr_reqtype:*prepaid",
				"cgr_reqtype":  "*prepaid",
			},
			"connected": map[string]any{
				"name":   "",
				"number": "1002",
			},
			"creationtime": "2024-05-03T08:53:05.234+0200",
			"dialplan": map[string]any{
				"app_data": "",
				"app_name": "",
				"context":  "internal",
				"exten":    "1002",
				"priority": "9",
			},
			"id":          "1714719185.3",
			"language":    "en",
			"name":        "PJSIP/1001-00000002",
			"protocol_id": "cb1bb28866dd7d52b42484e5b38765ec@0:0:0:0:0:0:0:0",
			"state":       "Up",
		},
		"timestamp": "2024-05-03T08:53:11.511+0200",
		"type":      "ChannelDestroyed",
	}
	smaEv := NewSMAsteriskEvent(ariEv, "127.0.0.1", utils.EmptyString)
	cgrEv := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ea36649",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.EventName:    "SMA_SESSION_TERMINATE",
			utils.OriginHost:   "127.0.0.1",
			utils.OriginID:     "1714734552.6",
			utils.RequestType:  utils.MetaRated,
			utils.SetupTime:    time.Date(2013, 12, 30, 15, 01, 31, 0, time.UTC),
			utils.Source:       utils.AsteriskAgent,
		},
	}
	exp := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ea36649",
		Event: map[string]any{
			utils.AccountField:    "1001",
			utils.AnswerTime:      "2024-05-03 08:53:06",
			utils.Destination:     "1002",
			utils.DisconnectCause: "Normal Clearing",
			utils.EventName:       "SMA_SESSION_TERMINATE",
			utils.OriginHost:      "127.0.0.1",
			utils.OriginID:        "1714734552.6",
			utils.RequestType:     utils.MetaPrepaid,
			utils.SetupTime:       time.Date(2013, 12, 30, 15, 01, 31, 0, time.UTC),
			utils.Source:          utils.AsteriskAgent,
			utils.Usage:           "4s",
			utils.CGRFlags:        "*accounts+*attributes+*resources+*stats+*routes+*thresholds+cgr_reqtype:*prepaid",
		},
	}
	if err := smaEv.RestoreAndUpdateFields(&cgrEv); err != nil {
		t.Error(err)
	} else if utils.ToJSON(cgrEv) != utils.ToJSON(exp) {
		t.Errorf("Expected <%v>, \nreceived <%v>", utils.ToJSON(exp), utils.ToJSON(cgrEv))
	}

}

func TestRestoreAndUpdateFieldsFail(t *testing.T) {
	ariEv := map[string]any{
		"application": "cgrates_auth",
		"asterisk_id": "08:00:27:18:d8:cf",
		"cause":       "16",
		"cause_txt":   "Normal Clearing",
		"channel": map[string]any{
			"accountcode": "",
			"caller": map[string]any{
				"name":   "1001",
				"number": "1001",
			},
			"connected": map[string]any{
				"name":   "",
				"number": "1002",
			},
			"creationtime": "2024-05-03T08:53:05.234+0200",
			"dialplan": map[string]any{
				"app_data": "",
				"app_name": "",
				"context":  "internal",
				"exten":    "1002",
				"priority": "9",
			},
			"id":          "1714719185.3",
			"language":    "en",
			"name":        "PJSIP/1001-00000002",
			"protocol_id": "cb1bb28866dd7d52b42484e5b38765ec@0:0:0:0:0:0:0:0",
			"state":       "Up",
		},
		"timestamp": "2024-05-03T08:53:11.511+0200",
		"type":      "ChannelDestroyed",
	}
	smaEv := NewSMAsteriskEvent(ariEv, "127.0.0.1", utils.EmptyString)
	cgrEv := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ea36649",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.EventName:    "SMA_SESSION_TERMINATE",
			utils.OriginHost:   "127.0.0.1",
			utils.OriginID:     "1714734552.6",
			utils.RequestType:  utils.MetaRated,
			utils.SetupTime:    time.Date(2013, 12, 30, 15, 01, 31, 0, time.UTC),
			utils.Source:       utils.AsteriskAgent,
		},
	}
	expErr := "channelvars not found in event <map[accountcode: caller:map[name:1001 number:1001] connected:map[name: number:1002] creationtime:2024-05-03T08:53:05.234+0200 dialplan:map[app_data: app_name: context:internal exten:1002 priority:9] id:1714719185.3 language:en name:PJSIP/1001-00000002 protocol_id:cb1bb28866dd7d52b42484e5b38765ec@0:0:0:0:0:0:0:0 state:Up]>"
	if err := smaEv.RestoreAndUpdateFields(&cgrEv); err.Error() != expErr {
		t.Errorf("Expected error <%v>, \nreceived <%+v>", expErr, err)
	}

}

func TestSMAsteriskEventClone(t *testing.T) {
	e := &SMAsteriskEvent{
		ariEv: map[string]any{
			"EV": "ID1",
			"Id": 1001,
		},
		asteriskIP:    "192.168.1.1",
		asteriskAlias: "pbx-1",
		cachedFields: map[string]string{
			"eventType": "ARIChannelStateChange",
		},
		opts: map[string]any{
			"opt1": true,
		},
	}
	clone := e.Clone()
	if !reflect.DeepEqual(clone.ariEv, e.ariEv) {
		t.Errorf("ariEv maps are not deeply equal")
	}
	clone.ariEv["EV"] = "modified"
	clone.cachedFields["eventType"] = "modified"
	clone.opts["opt1"] = false
	if e.ariEv["EV"] == "modified" {
		t.Errorf("Modifying clone affected original ariEv")
	}
	if e.cachedFields["eventType"] == "modified" {
		t.Errorf("Modifying clone affected original cachedFields")
	}
	if e.opts["opt1"] == false {
		t.Errorf("Modifying clone affected original opts")
	}
}
