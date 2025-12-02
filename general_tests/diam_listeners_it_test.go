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

func TestDiamListeners(t *testing.T) {

	content := `
{
	"general": {
		"log_level": 7
	},
	"apiers": {
		"enabled": true
	},
	"cdrs": {
		"enabled": true,
		"rals_conns": ["*localhost"]
	},
	"sessions": {
		"enabled": true,
		"rals_conns": ["*localhost"],
		"cdrs_conns": ["*localhost"]
	},
	"diameter_agent": {
		"enabled": true,
		"listeners": [
        {"address": "127.0.0.1:3868", "network": "tcp"},
        {"address": "127.0.0.2:3869", "network": "tcp"}
       ],
		"sessions_conns": ["*bijson_localhost"],
		"request_processors": [{
            "id": "DiamRplcOID",
            "flags": ["*cdrs"],
            "request_fields": [
                {
                    "tag": "OriginID",
                    "path": "*cgreq.OriginID",
                    "type": "*variable",
                    "value": "~*req.Session-Id",
                    "mandatory": true
                }
            ],
            "reply_fields": [
                {
                    "tag": "Result-Code",
                    "path": "Result-Code",
                    "type": "*constant",
                    "value": "2001",
                    "mandatory": true
                },
                {
                    "tag": "Origin-Host",
                    "path": "Origin-Host",
                    "type": "*variable",
                    "value": "~*req.OriginHost", 
                    "mandatory": true
                },
                {
                    "tag": "Origin-Realm",
                    "path": "Origin-Realm",
                    "type": "*variable",
                    "value": "~*req.OriginRealm",
                    "mandatory": true
                }
            ]
        }]
	}
}
`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	_, cfg := ng.Run(t)

	t.Run("TestMultipleListeners", func(t *testing.T) {
		listeners := cfg.DiameterAgentCfg().Listeners
		if len(listeners) == 0 {
			t.Fatal("No listeners configured in diameter_agent!")
		}

		for i, lCfg := range listeners {

			cli, err := agents.NewDiameterClient(lCfg.Address, "INTEGRATION_TEST_CLIENT",
				cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
				cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
				cfg.DiameterAgentCfg().DictionariesPath, lCfg.Network)

			if err != nil {
				t.Fatalf("Failed to create client for %s: %v", lCfg.Address, err)
			}
			m := diam.NewRequest(diam.CreditControl, 4, nil)
			m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(fmt.Sprintf("test-conn-%d", i)))
			m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("INTEGRATION_TEST_CLIENT"))
			m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("cgrates.org"))
			m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
			m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Enumerated(1))
			m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
			m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.DiameterIdentity(cfg.DiameterAgentCfg().OriginHost))
			m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.DiameterIdentity(cfg.DiameterAgentCfg().OriginRealm))

			if err := cli.SendMessage(m); err != nil {
				t.Errorf("Failed to send message to listener %s: %v", lCfg.Address, err)
				continue
			}

			reply := cli.ReceivedMessage(time.Second)
			if reply == nil {
				t.Errorf("No reply received from listener %s", lCfg.Address)
				continue
			}
			avps, err := reply.FindAVPsWithPath([]any{avp.OriginHost}, dict.UndefinedVendorID)
			if err != nil {
				t.Errorf("Error finding Origin-Host AVP: %v", err)
				continue
			}
			if len(avps) == 0 {
				t.Errorf("Origin-Host AVP missing in reply from %s", lCfg.Address)
				continue
			}
			var replyOriginHost string
			if val, ok := avps[0].Data.(datatype.DiameterIdentity); ok {
				replyOriginHost = string(val)
			} else {
				t.Errorf("AVP data is not DiameterIdentity! Got type: %T, Value: %v", avps[0].Data, avps[0].Data)
				continue
			}
			if replyOriginHost != "CGR-DA" {
				t.Errorf("Identity Mismatch! Expected 'CGR-DA', got '%s'", replyOriginHost)
			}
		}
	})
}
