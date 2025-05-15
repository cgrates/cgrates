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
	"bytes"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionInitiateDisableAccount(t *testing.T) {
	var cfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgDir = "tutinternal"
	case utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMySQL:
		cfgDir = "tutmysql"
	case utils.MetaMongo:
		cfgDir = "tutmongo"
	default:
		t.Fatal("Unkwown database type")
	}

	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", cfgDir),
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP1,DISABLE_TRIGGER,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP1,ACT_MON,*asap,10`,
			utils.ActionsCsv: `#ActionsId,Action,ExtraParameters,Filter,BalanceId,BalanceType,Categories,DestinationIds,RatingSubject,SharedGroup,ExpiryTime,TimingIds,Units,BalanceWeight,BalanceBlocker,BalanceDisabled,Weight
ACT_MON,*topup_reset,,,balance1,*monetary,,,,,,,10,,false,,
DISABLE_ACC,*disable_account,,,,,,,,,,,,,,,`,
			utils.ActionTriggersCsv: `#Tag[0],UniqueId[1],ThresholdType[2],ThresholdValue[3],Recurrent[4],MinSleep[5],ExpiryTime[6],ActivationTime[7],BalanceTag[8],BalanceType[9],BalanceCategories[10],BalanceDestinationIds[11],BalanceRatingSubject[12],BalanceSharedGroup[13],BalanceExpiryTime[14],BalanceTimingIds[15],BalanceWeight[16],BalanceBlocker[17],BalanceDisabled[18],ActionsId[19],Weight[20]
DISABLE_TRIGGER,,*max_event_connect,1,false,0,,,,*event_connect,,,,,,,,,,DISABLE_ACC,10`,
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP,`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP,DR_RP,*any,10`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_RP,DST_1002,RT1,*up,4,0,`,
			utils.DestinationsCsv: `#Id,Prefix
DST_1002,1002`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT1,0.2,1,1s,1s,0`,
		},
		LogBuffer: bytes.NewBuffer(nil),
	}
	client, _ := ng.Run(t)
	t.Run("TestAuthorizeEvent", func(t *testing.T) {
		var accRepl engine.Account
		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &accRepl); err != nil {
			t.Error(err)
		}
		if accRepl.Disabled {
			t.Errorf("account should not be disabled")
		}
		var reply sessions.V1AuthorizeReply
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent, &sessions.V1AuthorizeArgs{GetAttributes: true,
			GetMaxUsage: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{"Account": "1001",
					"Destination": "1002",
					"OriginHost":  "127.0.0.1:8448",
					"RequestType": "*prepaid",
					"SetupTime":   "1747212851",
					"Source":      "KamailioAgent"}}}, &reply); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &accRepl); err != nil {
			t.Error(err)
		} else if *reply.MaxUsage != (9 * time.Second) {
			t.Errorf("expected to get usage even though account got disabled")
		}

		if !accRepl.Disabled {
			t.Errorf("expected account to be disabled")
		}
	})

}
