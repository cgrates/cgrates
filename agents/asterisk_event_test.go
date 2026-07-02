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
	"testing"
)

var (
	stasisStart        = `{"application":"cgratesAuth","type":"StasisStart","timestamp":"2016-09-12T13:53:48.919+0200","args":["cgrReqtype=*prepaid","cgrRoute=supplier1", "extra1=val1", "extra2=val2"],"channel":{"id":"1473681228.6","state":"Ring","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelStateChange = `{"application":"cgratesAuth","type":"ChannelStateChange","timestamp":"2016-09-12T13:53:52.110+0200","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
)

func TestSMAEventType(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev)
	if smaEv.EventType() != "StasisStart" {
		t.Error("Received:", smaEv.EventType())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev)
	if smaEv.EventType() != "" {
		t.Error("Received:", smaEv.EventType())
	}
}

func TestSMAEventChannelID(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev)
	if smaEv.ChannelID() != "1473681228.6" {
		t.Error("Received:", smaEv.ChannelID())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev)
	if smaEv.ChannelID() != "" {
		t.Error("Received:", smaEv.ChannelID())
	}
}

func TestSMAEventChannelState(t *testing.T) {
	var ev map[string]any
	if err := json.Unmarshal([]byte(channelStateChange), &ev); err != nil {
		t.Error(err)
	}
	smaEv := NewSMAsteriskEvent(ev)
	if smaEv.ChannelState() != "Up" {
		t.Error("Received:", smaEv.ChannelState())
	}
	ev = make(map[string]any) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	smaEv = NewSMAsteriskEvent(ev)
	if smaEv.ChannelState() != "" {
		t.Error("Received:", smaEv.ChannelState())
	}
}
