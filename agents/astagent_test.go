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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestAAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(AsteriskAgent))
}

func TestHandleChannelDestroyedFail(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	cM := engine.NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS): internalSessionSChan,
	})
	sma, err := NewAsteriskAgent(cfg, 1, cM)
	if err != nil {
		t.Error(err)
	}

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
	ev := NewSMAsteriskEvent(ariEv, "127.0.0.1", utils.EmptyString)
	evCopy := ev
	sma.handleChannelDestroyed(ev)
	if ev != evCopy {
		t.Errorf("Expected ev to not change, received <%v>", utils.ToJSON(ev))
	}
}

func TestAstAgentV1WarnDisconnect(t *testing.T) {
	tAsteriskAgent := &AsteriskAgent{}
	tMap := map[string]any{}
	tString := ""
	err := tAsteriskAgent.V1WarnDisconnect(nil, tMap, &tString)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestAsteriskAgentV1DisconnectPeer(t *testing.T) {
	tAsteriskAgent := &AsteriskAgent{}
	tDPRArgs := &utils.DPRArgs{}
	tString := ""
	err := tAsteriskAgent.V1DisconnectPeer(nil, tDPRArgs, &tString)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestAsteriskAgentV1AlterSession(t *testing.T) {
	tAsteriskAgent := &AsteriskAgent{}
	tCGREvent := utils.CGREvent{}
	tString := ""
	err := tAsteriskAgent.V1AlterSession(nil, tCGREvent, &tString)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}
