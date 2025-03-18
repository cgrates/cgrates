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
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionUpdateToStats(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	content := `{

"data_db": {
	"db_type": "*internal",	
},


"stor_db": {
	"db_type": "*internal",	
},

"rals": {
	"enabled": true,
},


"schedulers": {
	"enabled": true,
},

"cdrs": {
	"enabled": true,
	"chargers_conns":["*internal"],
	"rals_conns": ["*internal"],
},

"chargers": {
	"enabled": true,
	"attributes_conns": ["*internal"],
},

"attributes": {
	"enabled": true,
},

"thresholds": {
	"enabled": true,
	"store_interval": "-1",
},
"stats": {
	"enabled": true,
	"store_interval": "-1",
},

"sessions": {
	"enabled": true,
	"chargers_conns": ["*internal"],
	"rals_conns": ["*internal"],
	"cdrs_conns": ["*internal"],
	"stats_conns": ["*internal"],
	"thresholds_conns": ["*internal"],
	"attributes_conns": ["*internal"],
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"],
},
	}`

	tpFiles := map[string]string{
		utils.ChargersCsv: `#Id,ActionsId,TimingId,Weight
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,DEFAULT,*none,20`,
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_1001,ACT_TOPUP,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP,*topup_reset,,,balance_monetary,*monetary,,*any,,,*unlimited,,5,10,false,false,20`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_VOICE,*any,RT_VOICE,*up,20,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_VOICE,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_VOICE,DR_VOICE,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_VOICE,`,
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,Stat1,*string:~*req.Account:1001,,,,,*tcc;*acd;*tcd,,,,,`,
		utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,TH1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,-1,0,0,false,10,,false`,
	}
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}

	client, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)
	t.Run("TestUpdateSession", func(t *testing.T) {
		var replyUpdate sessions.V1UpdateSessionReply
		if err := client.Call(context.Background(), utils.SessionSv1UpdateSession, &sessions.V1UpdateSessionArgs{
			UpdateSession:     true,
			GetAttributes:     true,
			ProcessStats:      true,
			ProcessThresholds: true,
			ThresholdIDs:      []string{"TH1"},
			StatIDs:           []string{"Stat1"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "SessionSv1UpdateSession1",
				Event: map[string]any{
					utils.OriginID:        "prepaidSession",
					utils.Tenant:          "cgrates.org",
					utils.Category:        "call",
					utils.ToR:             utils.MetaVoice,
					utils.RequestType:     utils.MetaPrepaid,
					utils.AccountField:    "1001",
					utils.Subject:         "1001",
					utils.Destination:     "1002",
					utils.SetupTime:       time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
					utils.AnswerTime:      time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
					utils.Usage:           5 * time.Second,
					utils.BalanceFactorID: "voiceFactor",
				},
				APIOpts: map[string]any{
					utils.OptsDebitInterval: 0,
				},
			},
		}, &replyUpdate); err != nil {
			t.Error(err)
		} else if len(*replyUpdate.StatQueueIDs) != 1 || len(*replyUpdate.ThresholdIDs) != 1 {
			t.Error("expected to pass event through stats or thresholds")
		}
	})
}
