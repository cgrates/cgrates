//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cenkalti/rpc2"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSyncSessions(t *testing.T) {
	cfgJSON := `{
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"general": {
	"reply_timeout": "1s"
},
"chargers": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"channel_sync_interval": "0",
	"chargers_conns": ["*internal"]
},
"apiers": {
	"enabled": true
}
}`
	ng := TestEnvironment{
		ConfigJSON: cfgJSON,
	}
	client, cfg := ng.Setup(t, 0)
	replyTimeout := cfg.GeneralCfg().ReplyTimeout

	var reply string
	if err := client.Call(utils.APIerSv1SetChargerProfile, &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "DEFAULT",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{utils.META_NONE},
		Weight:       10,
	}, &reply); err != nil {
		t.Fatalf("SetChargerProfile: %v", err)
	}

	disconnectCh := make(chan *utils.AttrDisconnectSession, 1)
	type activeSessionIDsHandler func(*rpc2.Client, *string, *[]*sessions.SessionID) error
	newClient := func(name string, getActive activeSessionIDsHandler) *rpc2.Client {
		t.Helper()
		handlers := map[string]any{
			utils.SessionSv1GetActiveSessionIDs: getActive,
			utils.SessionSv1DisconnectSession: func(_ *rpc2.Client, args *utils.AttrDisconnectSession, reply *string) error {
				disconnectCh <- args
				*reply = utils.OK
				return nil
			},
		}
		biJSONClient, err := utils.NewBiJSONrpcClient(cfg.SessionSCfg().ListenBijson, handlers)
		if err != nil {
			t.Fatalf("BiJSON dial %s: %v", name, err)
		}
		t.Cleanup(func() { biJSONClient.Close() })
		return biJSONClient
	}

	keepClient := newClient("keep", func(_ *rpc2.Client, _ *string, reply *[]*sessions.SessionID) error {
		*reply = []*sessions.SessionID{{
			OriginID:   "session-keep",
			OriginHost: "127.0.0.1",
		}}
		return nil
	})
	timeoutHandler := func(_ *rpc2.Client, _ *string, _ *[]*sessions.SessionID) error {
		time.Sleep(2 * replyTimeout)
		return nil
	}
	timeoutClient := newClient("timeout", timeoutHandler)
	newClient("timeout-extra", timeoutHandler)

	initSession := func(biJSONClient *rpc2.Client, originID string) {
		t.Helper()
		initArgs := &sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     originID,
				Event: map[string]any{
					utils.OriginID:    originID,
					utils.OriginHost:  "127.0.0.1",
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "1002",
					utils.Category:    "call",
					utils.Tenant:      "cgrates.org",
					utils.RequestType: utils.META_NONE,
					utils.SetupTime:   time.Now(),
					utils.AnswerTime:  time.Now(),
					utils.Usage:       time.Hour,
				},
			},
		}
		var initReply sessions.V1InitSessionReply
		if err := biJSONClient.Call(utils.SessionSv1InitiateSession, initArgs, &initReply); err != nil {
			t.Fatalf("InitiateSession %s: %v", originID, err)
		}
	}
	initSession(keepClient, "session-keep")
	initSession(timeoutClient, "session-timeout")

	var activeSessions []*sessions.ExternalSession
	if err := client.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &activeSessions); err != nil {
		t.Fatalf("GetActiveSessions: %v", err)
	}
	if len(activeSessions) != 2 {
		t.Fatalf("want 2 active sessions, got %d", len(activeSessions))
	}

	start := time.Now()
	if err := client.Call(utils.SessionSv1SyncSessions, "", &reply); err != nil {
		t.Fatalf("SyncSessions: %v", err)
	}
	if elapsed := time.Since(start); elapsed > replyTimeout+replyTimeout/2 {
		t.Fatalf("SyncSessions took %v, want parallel timeout around %v", elapsed, replyTimeout)
	}

	select {
	case disconnect := <-disconnectCh:
		originID, _ := disconnect.EventStart[utils.OriginID].(string)
		if originID != "session-timeout" {
			t.Fatalf("disconnect OriginID = %q, want session-timeout", originID)
		}
	case <-time.After(replyTimeout):
		t.Fatal("expected force-disconnect callback")
	}

	activeSessions = nil
	if err := client.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &activeSessions); err != nil {
		t.Fatalf("GetActiveSessions after sync: %v", err)
	}
	if len(activeSessions) != 1 {
		t.Fatalf("want 1 active session after sync, got %d", len(activeSessions))
	}
	activeOriginID := activeSessions[0].OriginID
	if activeOriginID != "session-keep" {
		t.Fatalf("active OriginID = %q, want session-keep", activeOriginID)
	}
}
