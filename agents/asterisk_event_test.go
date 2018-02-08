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
package agents

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	stasisStart                = `{"application":"cgrates_auth","type":"StasisStart","timestamp":"2016-09-12T13:53:48.919+0200","args":["cgr_reqtype=*prepaid","cgr_supplier=supplier1", "extra1=val1", "extra2=val2"],"channel":{"id":"1473681228.6","state":"Ring","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelStateChange         = `{"application":"cgrates_auth","type":"ChannelStateChange","timestamp":"2016-09-12T13:53:52.110+0200","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelAnsweredDestroyed   = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T13:54:27.335+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"},"cause":16}`
	channelUnansweredDestroyed = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T18:00:18.121+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473696018.2","state":"Ring","name":"PJSIP/1002-00000002","caller":{"name":"1002","number":"1002"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T18:00:18.109+0200"},"cause":16}`
	channelBusyDestroyed       = `{"type":"ChannelDestroyed","timestamp":"2016-09-13T12:59:48.806+0200","application":"cgrates_auth","cause_txt":"User busy","channel":{"id":"1473764378.3","state":"Ring","name":"PJSIP/1001-00000002","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":4},"creationtime":"2016-09-13T12:59:38.839+0200"},"cause":17}`
)

func TestSMAParseStasisArgs(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	expAppArgs := map[string]string{"cgr_reqtype": "*prepaid", "cgr_supplier": "supplier1", "extra1": "val1", "extra2": "val2"}
	if !reflect.DeepEqual(smaEv.cachedFields, expAppArgs) {
		t.Errorf("Expecting: %+v, received: %+v", smaEv.cachedFields, expAppArgs)
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	expAppArgs = map[string]string{}
	if !reflect.DeepEqual(smaEv.cachedFields, expAppArgs) {
		t.Errorf("Expecting: %+v, received: %+v", smaEv.cachedFields, expAppArgs)
	}
}

func TestSMAEventType(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.EventType() != "StasisStart" {
		t.Error("Received:", smaEv.EventType())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.EventType() != "" {
		t.Error("Received:", smaEv.EventType())
	}
}

func TestSMAEventChannelID(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.ChannelID() != "1473681228.6" {
		t.Error("Received:", smaEv.ChannelID())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.ChannelID() != "" {
		t.Error("Received:", smaEv.ChannelID())
	}
}

func TestSMAEventOriginatorIP(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.OriginatorIP() != "127.0.0.1" {
		t.Error("Received:", smaEv.OriginatorIP())
	}
}

func TestSMAEventAccount(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Account() != "1001" {
		t.Error("Received:", smaEv.Account())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Account() != "" {
		t.Error("Received:", smaEv.Account())
	}
}

func TestSMAEventDestination(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Destination() != "1002" {
		t.Error("Received:", smaEv.Destination())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Destination() != "" {
		t.Error("Received:", smaEv.Destination())
	}
}

func TestSMAEventTimestamp(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Timestamp() != "2016-09-12T13:53:48.919+0200" {
		t.Error("Received:", smaEv.Timestamp())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Timestamp() != "" {
		t.Error("Received:", smaEv.Timestamp())
	}
}

func TestSMAEventChannelState(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.ChannelState() != "Up" {
		t.Error("Received:", smaEv.ChannelState())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.ChannelState() != "" {
		t.Error("Received:", smaEv.ChannelState())
	}
}

func TestSMASetupTime(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.SetupTime() != "2016-09-12T13:53:48.918+0200" {
		t.Error("Received:", smaEv.SetupTime())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.SetupTime() != "" {
		t.Error("Received:", smaEv.SetupTime())
	}
}

func TestSMAEventRequestType(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.RequestType() != "*prepaid" {
		t.Error("Received:", smaEv.RequestType())
	}
	ev = make(map[string]interface{}) // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.RequestType() != "" {
		t.Error("Received:", smaEv.RequestType())
	}
}

func TestSMAEventTenant(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Tenant() != "" {
		t.Error("Received:", smaEv.Tenant())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_tenant=cgrates.org"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Tenant() != "cgrates.org" {
		t.Error("Received:", smaEv.Tenant())
	}
}

func TestSMAEventCategory(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Category() != "" {
		t.Error("Received:", smaEv.Category())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_category=premium_call"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Category() != "premium_call" {
		t.Error("Received:", smaEv.Category())
	}
}

func TestSMAEventSubject(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Subject() != "" {
		t.Error("Received:", smaEv.Subject())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_subject=dan"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Subject() != "dan" {
		t.Error("Received:", smaEv.Subject())
	}
}

func TestSMAEventPDD(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.PDD() != "" {
		t.Error("Received:", smaEv.PDD())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_pdd=2.1"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.PDD() != "2.1" {
		t.Error("Received:", smaEv.PDD())
	}
}

func TestSMAEventSupplier(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Supplier() != "supplier1" {
		t.Error("Received:", smaEv.Supplier())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_supplier=supplier1"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.Supplier() != "supplier1" {
		t.Error("Received:", smaEv.Supplier())
	}
}

func TestSMAEventDisconnectCause(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.DisconnectCause() != "" {
		t.Error("Received:", smaEv.DisconnectCause())
	}
	ev = map[string]interface{}{"args": []interface{}{"cgr_disconnectcause=NORMAL_DISCONNECT"}} // Clear previous data
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.DisconnectCause() != "NORMAL_DISCONNECT" {
		t.Error("Received:", smaEv.DisconnectCause())
	}
}

func TestSMAEventExtraParameters(t *testing.T) {
	expExtraParams := map[string]string{
		"extra1": "val1",
		"extra2": "val2",
	}
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smaEv.PDD() != "" {
		t.Error("Received:", smaEv.PDD())
	}
	if extraParams := smaEv.ExtraParameters(); !reflect.DeepEqual(expExtraParams, extraParams) {
		t.Errorf("Expecting: %+v, received: %+v", expExtraParams, extraParams)
	}
}

/*
func TestSMAEventUpdateFromEvent(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	ev = make(map[string]interface{})
	if err := json.Unmarshal([]byte(channelAnsweredDestroyed), &ev); err != nil {
		t.Error(err)
	}
	smaEv2 := NewSMAsteriskEvent(ev, "127.0.0.1")
	smaEv.UpdateFromEvent(smaEv2)
	eSMAEv := &SMAsteriskEvent{
		ariEv: map[string]interface{}{
			"type":        "ChannelDestroyed",
			"args":        []interface{}{"cgr_reqtype=*prepaid", "cgr_destination=1003", "extra1=val1", "extra2=val2"},
			"timestamp":   "2016-09-12T13:54:27.335+0200",
			"application": "cgrates_auth",
			"cause_txt":   "Normal Clearing",
			"channel": map[string]interface{}{
				"id":    "1473681228.6",
				"state": "Up",
				"name":  "PJSIP/1001-00000004",
				"caller": map[string]interface{}{
					"name":   "1001",
					"number": "1001"},
				"language": "en",
				"connected": map[string]interface{}{
					"name":   "",
					"number": "1002"},
				"accountcode": "",
				"dialplan": map[string]interface{}{
					"context":  "internal",
					"exten":    "1002",
					"priority": 3.0},
				"creationtime": "2016-09-12T13:53:48.918+0200"},
			"cause": 16.0},
		asteriskIP: "127.0.0.1",
		cachedFields: map[string]string{"cgr_reqtype": "*prepaid",
			"cgr_destination": "1003", "extra1": "val1", "extra2": "val2"},
	}
	if !reflect.DeepEqual(eSMAEv, smaEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMAEv, smaEv)
	}
}
*/

//Here
func TestSMAEventAsSMGenericEvent(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	eSMGEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:  SMAAuthorization,
		utils.OriginID:    "1473681228.6",
		utils.RequestType: "*prepaid",
		utils.OriginHost:  "127.0.0.1",
		utils.Account:     "1001",
		utils.Destination: "1002",
		utils.SetupTime:   "2016-09-12T13:53:48.919+0200",
		utils.SUPPLIER:    "supplier1",
		"extra1":          "val1",
		"extra2":          "val2",
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if smgEv := smaEv.AsSMGenericEvent(); !reflect.DeepEqual(eSMGEv, smgEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, smgEv)
	}
}

func TestSMAEventUpdateSMGEventAnswered(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	smgEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:  SMAAuthorization,
		utils.OriginID:    "1473681228.6",
		utils.RequestType: "*prepaid",
		utils.OriginHost:  "127.0.0.1",
		utils.Account:     "1001",
		utils.Destination: "1003",
		utils.SetupTime:   "2016-09-12T13:53:48.919+0200",
		"extra1":          "val1",
		"extra2":          "val2",
	}
	eSMGEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:  SMASessionStart,
		utils.OriginID:    "1473681228.6",
		utils.RequestType: "*prepaid",
		utils.OriginHost:  "127.0.0.1",
		utils.Account:     "1001",
		utils.Destination: "1003",
		utils.SetupTime:   "2016-09-12T13:53:48.919+0200",
		utils.AnswerTime:  "2016-09-12T13:53:52.110+0200",
		"extra1":          "val1",
		"extra2":          "val2",
	}
	if err := smaEv.UpdateSMGEvent(smgEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, smgEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, smgEv)
	}
	// Apply update using a terminate event
	ev = make(map[string]interface{})
	if err = json.Unmarshal([]byte(channelAnsweredDestroyed), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev, "127.0.0.1")
	eSMGEv = &sessions.SMGenericEvent{
		utils.EVENT_NAME:       SMASessionTerminate,
		utils.OriginID:         "1473681228.6",
		utils.RequestType:      "*prepaid",
		utils.OriginHost:       "127.0.0.1",
		utils.Account:          "1001",
		utils.Destination:      "1003",
		utils.SetupTime:        "2016-09-12T13:53:48.919+0200",
		utils.AnswerTime:       "2016-09-12T13:53:52.110+0200",
		utils.Usage:            "35.225s",
		utils.DISCONNECT_CAUSE: "Normal Clearing",
		"extra1":               "val1",
		"extra2":               "val2",
	}
	if err := smaEv.UpdateSMGEvent(smgEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, smgEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, smgEv)
	}
}

func TestSMAEventUpdateSMGEventUnaswered(t *testing.T) {
	smgEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:  SMAAuthorization,
		utils.OriginID:    "1473681228.6",
		utils.RequestType: "*prepaid",
		utils.OriginHost:  "127.0.0.1",
		utils.Account:     "1001",
		utils.Destination: "1003",
		utils.SetupTime:   "2016-09-12T13:53:48.919+0200",
		"extra1":          "val1",
		"extra2":          "val2",
	}
	eSMGEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:       SMASessionTerminate,
		utils.OriginID:         "1473681228.6",
		utils.RequestType:      "*prepaid",
		utils.OriginHost:       "127.0.0.1",
		utils.Account:          "1001",
		utils.Destination:      "1003",
		utils.SetupTime:        "2016-09-12T13:53:48.919+0200",
		utils.Usage:            "0s",
		utils.DISCONNECT_CAUSE: "Normal Clearing",
		"extra1":               "val1",
		"extra2":               "val2",
	}
	// Apply update using a terminate event
	ev := make(map[string]interface{})
	if err := json.Unmarshal([]byte(channelUnansweredDestroyed), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if err := smaEv.UpdateSMGEvent(smgEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, smgEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, smgEv)
	}
}

func TestSMAEventUpdateSMGEventBusy(t *testing.T) {
	smgEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:  SMAAuthorization,
		utils.OriginID:    "1473681228.6",
		utils.RequestType: "*prepaid",
		utils.OriginHost:  "127.0.0.1",
		utils.Account:     "1001",
		utils.Destination: "1003",
		utils.SetupTime:   "2016-09-12T13:53:48.919+0200",
		"extra1":          "val1",
		"extra2":          "val2",
	}
	eSMGEv := &sessions.SMGenericEvent{
		utils.EVENT_NAME:       SMASessionTerminate,
		utils.OriginID:         "1473681228.6",
		utils.RequestType:      "*prepaid",
		utils.OriginHost:       "127.0.0.1",
		utils.Account:          "1001",
		utils.Destination:      "1003",
		utils.SetupTime:        "2016-09-12T13:53:48.919+0200",
		utils.Usage:            "0s",
		utils.DISCONNECT_CAUSE: "User busy",
		"extra1":               "val1",
		"extra2":               "val2",
	}
	// Apply update using a terminate event
	ev := make(map[string]interface{})
	if err := json.Unmarshal([]byte(channelBusyDestroyed), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev, "127.0.0.1")
	if err := smaEv.UpdateSMGEvent(smgEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSMGEv, smgEv) {
		t.Errorf("Expecting: %+v, received: %+v", eSMGEv, smgEv)
	}
}
