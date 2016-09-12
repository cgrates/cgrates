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
package sessionmanager

import (
	"encoding/json"
	"testing"
)

var (
	stasisStart                = `{"application":"cgrates_auth","type":"StasisStart","timestamp":"2016-09-12T13:53:48.919+0200","args":[],"channel":{"id":"1473681228.6","state":"Ring","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelStateChange         = `{"application":"cgrates_auth","type":"ChannelStateChange","timestamp":"2016-09-12T13:53:52.110+0200","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"}}`
	channelAnsweredDestroyed   = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T13:54:27.335+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473681228.6","state":"Up","name":"PJSIP/1001-00000004","caller":{"name":"1001","number":"1001"},"language":"en","connected":{"name":"","number":"1002"},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":3},"creationtime":"2016-09-12T13:53:48.918+0200"},"cause":16}`
	channelUnansweredDestroyed = `{"type":"ChannelDestroyed","timestamp":"2016-09-12T18:00:18.121+0200","application":"cgrates_auth","cause_txt":"Normal Clearing","channel":{"id":"1473696018.2","state":"Ring","name":"PJSIP/1002-00000002","caller":{"name":"1002","number":"1002"},"language":"en","connected":{"name":"","number":""},"accountcode":"","dialplan":{"context":"internal","exten":"1002","priority":2},"creationtime":"2016-09-12T18:00:18.109+0200"},"cause":16}`
)

func TestSMAEventType(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	ariEv := NewARIEvent(ev)
	if ariEv.Type() != "StasisStart" {
		t.Error("Received type:", ariEv.Type())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	ariEv = NewARIEvent(ev)
	if ariEv.Type() != "" {
		t.Error("Received type:", ariEv.Type())
	}
}

func TestSMAEventChannelID(t *testing.T) {
	var ev map[string]interface{}
	if err := json.Unmarshal([]byte(stasisStart), &ev); err != nil {
		t.Error(err)
	}
	ariEv := NewARIEvent(ev)
	if ariEv.ChannelID() != "1473681228.6" {
		t.Error("Received type:", ariEv.ChannelID())
	}
	ev = make(map[string]interface{}) // Clear previous data
	if err := json.Unmarshal([]byte("{}"), &ev); err != nil {
		t.Error(err)
	}
	ariEv = NewARIEvent(ev)
	if ariEv.ChannelID() != "" {
		t.Error("Received type:", ariEv.ChannelID())
	}
}
