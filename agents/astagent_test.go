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
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/google/go-cmp/cmp"
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
	evCopy := ev.Clone()
	evCopy.cachedFields = map[string]string{"channelID": "1714719185.3"}
	sma.handleChannelDestroyed(ev)
	if diff := cmp.Diff(evCopy, ev, cmp.AllowUnexported(SMAsteriskEvent{})); diff != "" {
		t.Errorf("handleChannelDestroyed modified SMAsteriskEvent unexpectedly (-want +got): \n%s", diff)
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

func TestHandleChannelDestroyedCases(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	cM := engine.NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS): internalSessionSChan,
	})
	sma, err := NewAsteriskAgent(cfg, 1, cM)
	if err != nil {
		t.Error(err)
	}

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	t.Cleanup(func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	})

	testCases := []struct {
		name   string
		ariEv  map[string]any
		expLog string
	}{
		{
			name:   "Missing Channel",
			ariEv:  map[string]any{},
			expLog: "<AsteriskAgent> missing or invalid 'channel' field in event: {}",
		},
		{
			name:   "Invalid Channel",
			ariEv:  map[string]any{"channel": "invalid"},
			expLog: `<AsteriskAgent> missing or invalid 'channel' field in event: {"channel":"invalid"}`,
		},
		{
			name:   "Missing ChannelVars",
			ariEv:  map[string]any{"channel": map[string]any{"channel": "1"}},
			expLog: `<AsteriskAgent> missing or invalid 'channelvars' field in 'channel': {"channel":"1"}`,
		},
		{
			name:   "Invalid ChannelVars",
			ariEv:  map[string]any{"channel": map[string]any{"channelvars": "invalid"}},
			expLog: `<AsteriskAgent> missing or invalid 'channelvars' field in 'channel': {"channelvars":"invalid"}`,
		},
		{
			name:   "Valid ChannelVars",
			ariEv:  map[string]any{"channel": map[string]any{"channelvars": map[string]any{utils.CGRReqType: "test type"}}},
			expLog: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log.SetOutput(buf)
			ev := NewSMAsteriskEvent(tc.ariEv, "127.0.0.1", utils.EmptyString)
			sma.handleChannelDestroyed(ev)
			if !strings.Contains(buf.String(), tc.expLog) {
				t.Errorf("expected log warning %s", buf)
			}

		})
	}

}
