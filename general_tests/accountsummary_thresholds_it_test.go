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
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountSummaryThresholds(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaPostgres, utils.MetaMongo, utils.MetaMySQL:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	handler := http.NewServeMux()
	var receivedRequest *http.Request
	handler.HandleFunc("/balance_exhausted", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		receivedRequest = r
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	content := `{
"general": {
	"log_level": 7,
	"reply_timeout": "50s"
},

"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080"
},

"data_db": {
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true,
	"thresholds_conns": ["*localhost"],
},
"schedulers": {
	"enabled": true,
},
"cdrs": {
	"enabled": true,
	 "rals_conns": ["*localhost"],
},
"thresholds": {
	"enabled": true,
	"indexed_selects": false,
	"store_interval": "-1"
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
},
	}`
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpPath:     path.Join(*utils.DataDir, "tariffplans", "tutorial"),
		TpFiles: map[string]string{
			utils.ActionsCsv: fmt.Sprintf(`#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_HTTP,*http_post_async,%s/balance_exhausted,,,,,,,,,,,,false,false,10`, server.URL),
			utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,THD_1,*lt:~*asm.BalanceSummaries.test.Value:1,2024-07-29T15:00:00Z,1,1,,false,10,ACT_HTTP,true`,
		},
	}
	client, _ := ng.Run(t)

	t.Run("AccountSummaryThresholds", func(t *testing.T) {
		args := &engine.ArgV1ProcessEvent{
			Flags: []string{},
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.OriginID:     utils.GenUUID(),
					utils.Source:       "test",
					utils.RequestType:  utils.MetaPrepaid,
					utils.Category:     "call",
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
					utils.Usage:        100 * time.Minute,
				},
			},
		}
		var reply string
		if err := client.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
		time.Sleep(100 * time.Millisecond)
		if receivedRequest == nil {
			t.Error("Expected HTTP request was not received")
		}
	})

}
