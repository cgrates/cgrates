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

package general_tests

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
)

func TestDiamRplcOID(t *testing.T) {

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
		"sessions_conns": ["*bijson_localhost"],
		"request_processors": [{
			"id": "DiamRplcOID",
			"flags": ["*cdrs"],
			"request_fields": [{
					"tag": "OriginID",
					"path": "*cgreq.OriginID",
					"type": "*variable",
					"value": "~*req.Session-Id",
					"mandatory": true
				},
				{
					"tag": "OriginIDRandomSuffixWithFilter",
					"path": "*cgreq.OriginID",
					"type": "*composed",
					"filters": ["*string:~*req.Session-Id:session"],
					"value": "~*req.Session-Id;{*random}"
				}
			]
		}]
	}
}
`

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_sms,*sms,,,,,*unlimited,,1000000,,,,`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		},
	}
	client, cfg := ng.Run(t)

	diamClient, err := agents.NewDiameterClient(cfg.DiameterAgentCfg().Listen, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().ListenNet)
	if err != nil {
		t.Fatal(err)
	}

	sessionID := fmt.Sprintf("session")
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))

	for i := 1; i < 4; i++ {
		if err := diamClient.SendMessage(ccr); err != nil {
			t.Errorf("failed to send diameter message: %v", err)
			return
		}

		reply := diamClient.ReceivedMessage(time.Second)
		if reply == nil {
			t.Error("received empty reply")
			return
		}

		var cdrs []*engine.CDR
		argsCdr := &utils.RPCCDRsFilterWithAPIOpts{}
		if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, argsCdr, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != i {
			t.Fatalf("expected %v cdr, received %v", i, len(cdrs))
		} else if !strings.HasPrefix(cdrs[0].OriginID, "session") {
			t.Errorf("expected prefix <%v>, received <%v>", "session", cdrs[0].OriginID)
		}

		suffix := strings.TrimLeft(cdrs[0].OriginID, "session")
		_, err = strconv.Atoi(suffix)
		if err != nil {
			t.Error("suffix isnt convertable to number:", err)
		}
	}
}
