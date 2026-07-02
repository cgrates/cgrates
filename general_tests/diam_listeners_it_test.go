//go:build integration
// +build integration

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

package general_tests

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
)

// TestDiamMultipleListeners boots a diameter agent with two TCP listeners
// on different ports and verifies that each listener accepts a CCR and that
// the reply carries the agent's Origin-Host. This exercises the multi-listener
// bind path and the fact that all listeners share a single state machine.
func TestDiamMultipleListeners(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	cfgJSON := `{
"sessions": {
	"enabled": true
},
"diameterAgent": {
	"enabled": true,
	"listeners": [
		{"address": "127.0.0.1:13868", "network": "tcp"},
		{"address": "127.0.0.1:13869", "network": "tcp"}
	],
	"requestProcessors": [{
		"id": "multilistener",
		"flags": ["*dryRun"],
		"requestFields": [{
			"tag": "OriginID",
			"path": "*cgreq.OriginID",
			"type": "*variable",
			"value": "~*req.Session-Id",
			"mandatory": true
		}],
		"replyFields": [
			{"tag": "Result-Code",  "path": "Result-Code",  "type": "*constant", "value": "2001",            "mandatory": true},
			{"tag": "Origin-Host",  "path": "Origin-Host",  "type": "*variable", "value": "~*req.OriginHost","mandatory": true},
			{"tag": "Origin-Realm", "path": "Origin-Realm", "type": "*variable", "value": "~*req.OriginRealm","mandatory": true}
		]
	}]
}}`

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  bytes.NewBuffer(nil),
	}
	_, cfg := ng.Run(t)
	time.Sleep(2 * time.Second)

	listeners := cfg.DiameterAgentCfg().Listeners
	if len(listeners) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(listeners))
	}

	for i, lstnr := range listeners {
		t.Run(fmt.Sprintf("listener_%d_%s", i, lstnr.Address), func(t *testing.T) {
			cli, err := agents.NewDiameterClient(lstnr.Address, "INTEGRATION_TEST_CLIENT",
				cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
				cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
				cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().DictionariesAppendDefaults, lstnr.Network)
			if err != nil {
				t.Fatalf("client for %s: %v", lstnr.Address, err)
			}

			m := diam.NewRequest(diam.CreditControl, 4, nil)
			m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(fmt.Sprintf("session-%d", i)))
			m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("INTEGRATION_TEST_CLIENT"))
			m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
			m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
			m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
			m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
			m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity(cfg.DiameterAgentCfg().OriginHost))
			m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity(cfg.DiameterAgentCfg().OriginRealm))

			if err := cli.SendMessage(m); err != nil {
				t.Fatalf("send to %s: %v", lstnr.Address, err)
			}

			reply := cli.ReceivedMessage(2 * time.Second)
			if reply == nil {
				t.Fatalf("no reply from %s", lstnr.Address)
			}
			avps, err := reply.FindAVPsWithPath([]any{avp.OriginHost}, dict.UndefinedVendorID)
			if err != nil {
				t.Fatalf("find Origin-Host AVP from %s: %v", lstnr.Address, err)
			}
			if len(avps) == 0 {
				t.Fatalf("Origin-Host AVP missing in reply from %s", lstnr.Address)
			}
			oh, ok := avps[0].Data.(datatype.DiameterIdentity)
			if !ok {
				t.Fatalf("Origin-Host AVP not DiameterIdentity from %s (got %T)",
					lstnr.Address, avps[0].Data)
			}
			if got, want := string(oh), cfg.DiameterAgentCfg().OriginHost; got != want {
				t.Errorf("Origin-Host mismatch from %s: got %q want %q",
					lstnr.Address, got, want)
			}
		})
	}
}

func TestDiamMultipleListenersSessionID(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	cfgJSON := `{
"sessions": {
	"enabled": true
},
"diameterAgent": {
	"enabled": true,
	"listeners": [
		{"address": "127.0.0.1:13870", "network": "tcp"},
		{"address": "127.0.0.1:13871", "network": "tcp"}
	],
	"requestProcessors": [{
		"id": "shared_sessionid",
		"flags": ["*dryRun"],
		"requestFields": [{
			"tag": "OriginID",
			"path": "*cgreq.OriginID",
			"type": "*variable",
			"value": "~*req.Session-Id",
			"mandatory": true
		}],
		"replyFields": [
			{"tag": "Result-Code",  "path": "Result-Code",  "type": "*constant", "value": "2001",            "mandatory": true},
			{"tag": "Origin-Host",  "path": "Origin-Host",  "type": "*variable", "value": "~*req.OriginHost","mandatory": true},
			{"tag": "Origin-Realm", "path": "Origin-Realm", "type": "*variable", "value": "~*req.OriginRealm","mandatory": true},
			{"tag": "Session-Id",   "path": "Session-Id",   "type": "*variable", "value": "~*req.Session-Id","mandatory": true}
		]
	}]
}}`

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  bytes.NewBuffer(nil),
	}
	_, cfg := ng.Run(t)
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("engine log:\n%s", ng.LogBuffer)
		}
	})
	time.Sleep(2 * time.Second)

	listeners := cfg.DiameterAgentCfg().Listeners
	if len(listeners) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(listeners))
	}

	daCfg := cfg.DiameterAgentCfg()
	cliA, err := agents.NewDiameterClient(listeners[0].Address, "INTEGRATION_TEST_CLIENT",
		daCfg.OriginRealm, daCfg.VendorID,
		daCfg.ProductName, utils.DiameterFirmwareRevision,
		daCfg.DictionariesPath, daCfg.DictionariesAppendDefaults, listeners[0].Network)
	if err != nil {
		t.Fatalf("client A: %v", err)
	}
	defer cliA.Close()

	cliB, err := agents.NewDiameterClient(listeners[1].Address, "INTEGRATION_TEST_CLIENT",
		daCfg.OriginRealm, daCfg.VendorID,
		daCfg.ProductName, utils.DiameterFirmwareRevision,
		daCfg.DictionariesPath, daCfg.DictionariesAppendDefaults, listeners[1].Network)
	if err != nil {
		t.Fatalf("client B: %v", err)
	}
	defer cliB.Close()

	const sessionID = "session42"
	agentHost := daCfg.OriginHost
	agentRealm := daCfg.OriginRealm

	steps := []struct {
		name    string
		cli     *agents.DiameterClient
		reqType int
		reqNum  int
	}{
		{"Initial-on-A", cliA, 1, 0},
		{"Update-on-B", cliB, 2, 1},
		{"Terminate-on-A", cliA, 3, 2},
	}
	for _, s := range steps {
		t.Run(s.name, func(t *testing.T) {
			m := diam.NewRequest(diam.CreditControl, 4, nil)
			m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))
			m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("INTEGRATION_TEST_CLIENT"))
			m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
			m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
			m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(s.reqType))
			m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(s.reqNum))
			m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity(agentHost))
			m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity(agentRealm))

			if err := s.cli.SendMessage(m); err != nil {
				t.Fatalf("send: %v", err)
			}
			reply := s.cli.ReceivedMessage(3 * time.Second)
			if reply == nil {
				t.Fatalf("no reply")
			}

			ohAVPs, err := reply.FindAVPsWithPath([]any{avp.OriginHost}, dict.UndefinedVendorID)
			if err != nil || len(ohAVPs) == 0 {
				t.Fatalf("Origin-Host AVP missing: %v", err)
			}
			if got := string(ohAVPs[0].Data.(datatype.DiameterIdentity)); got != agentHost {
				t.Errorf("Origin-Host: got %q want %q", got, agentHost)
			}

			sidAVPs, err := reply.FindAVPsWithPath([]any{avp.SessionID}, dict.UndefinedVendorID)
			if err != nil || len(sidAVPs) == 0 {
				t.Fatalf("Session-Id AVP missing: %v", err)
			}
			if got := string(sidAVPs[0].Data.(datatype.UTF8String)); got != sessionID {
				t.Errorf("Session-Id: got %q want %q", got, sessionID)
			}
		})
	}
}
