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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestFallbackDepth tests both the fallback_depth configuration (previously, max depth
// was hardcoded to 3) and that fallback keys are not ordered automatically.
func TestFallbackDepth(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{
"general": {
	"logger": "*stdout",
	"log_level": 3
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true,
	"fallback_depth": 4
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_MainSubj,DST_MainSubj,RT_MainSubj,*up,4,0,
DR_FBSubj2,DST_FBSubj2,RT_FBSubj2,*up,4,0,
DR_FBSubj1,DST_FBSubj1,RT_FBSubj1,*up,4,0,
DR_FBSubj3,DST_FBSubj3,RT_FBSubj3,*up,4,0,
DR_FBSubj4,DST_FBSubj4,RT_FBSubj4,*up,4,0,
DR_DEFAULT,DST_DEFAULT,RT_DEFAULT,*up,4,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_MainSubj,1001
DST_FBSubj2,2001
DST_FBSubj1,3001
DST_FBSubj3,4001
DST_FBSubj4,5001
DST_DEFAULT,6001`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_MainSubj,0,1,1s,1s,0s
RT_FBSubj2,0,2,1s,1s,0s
RT_FBSubj1,0,3,1s,1s,0s
RT_FBSubj3,0,4,1s,1s,0s
RT_FBSubj4,0,5,1s,1s,0s
RT_DEFAULT,0,6,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_MainSubj,DR_MainSubj,*any,
RP_FBSubj2,DR_FBSubj2,*any,
RP_FBSubj1,DR_FBSubj1,*any,
RP_FBSubj3,DR_FBSubj3,*any,
RP_FBSubj4,DR_FBSubj4,*any,
RP_DEFAULT,DR_DEFAULT,*any,`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,MainSubj,,RP_MainSubj,FBSubj2;FBSubj1
cgrates.org,call,FBSubj2,,RP_FBSubj2,
cgrates.org,call,FBSubj1,,RP_FBSubj1,FBSubj3
cgrates.org,call,FBSubj3,,RP_FBSubj3,FBSubj4
cgrates.org,call,FBSubj4,,RP_FBSubj4,DEFAULT
cgrates.org,call,DEFAULT,,RP_DEFAULT,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
		LogBuffer:  buf,
	}
	client, _ := ng.Run(t)

	cdrIdx := 0
	processCDR := func(t *testing.T, dest string, shouldFail bool) engine.EventCost {
		t.Helper()
		cdrIdx++
		var reply []*utils.EventWithFlags
		err := client.Call(context.Background(), utils.CDRsV2ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        fmt.Sprintf("run_%d", cdrIdx),
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     fmt.Sprintf("processCDR%d", cdrIdx),
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Subject:      "MainSubj",
						utils.Destination:  dest,
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        10 * time.Second,
					},
				},
			}, &reply)
		if err != nil {
			t.Error(err)
		}

		if shouldFail {
			return engine.EventCost{}
		}

		// Convert CostDetails value from map[string]any to engine.EventCost in order to be able to retrieve
		// the RatingSubject used using FieldAsString method.
		ecIface, has := reply[0].Event[utils.CostDetails]
		if !has {
			t.Errorf("expected CGREvent to have CostDetails populated")
			return engine.EventCost{}
		}
		b, err := json.Marshal(ecIface)
		if err != nil {
			t.Error(err)
			return engine.EventCost{}
		}
		var ec engine.EventCost
		err = json.Unmarshal(b, &ec)
		if err != nil {
			t.Error(err)
		}
		return ec
	}

	checkSubjectAndCost := func(t *testing.T, ec engine.EventCost, wantCost float64, wantSubj, wantLog string) {
		t.Helper()
		if wantLog != "" {
			if !strings.Contains(buf.String(), wantLog) {
				t.Error("expected unauthorized destination log")
			}
			return
		}
		subj, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "Subject"})
		if err != nil {
			t.Error(err)
			return
		}
		if subj != wantSubj {
			t.Errorf("*req.CostDetails.Charges[0].Rating.RatingFilter.Subject = %s, want %s", subj, wantSubj)
		}
		if ec.Cost == nil {
			t.Error("nil cost in EventCost")
			return
		}
		rcvCost := *ec.Cost
		if rcvCost != wantCost {
			t.Errorf("ec.Cost = %v, want %v", rcvCost, wantCost)
		}
	}

	// checkRP := func(t *testing.T, subj string) {
	// 	var rpl engine.RatingProfile
	// 	if err := client.Call(context.Background(), utils.APIerSv1GetRatingProfile,
	// 		&utils.AttrGetRatingProfile{
	// 			Tenant:   "cgrates.org",
	// 			Category: "call",
	// 			Subject:  subj,
	// 		}, &rpl); err != nil {
	// 		t.Error(err)
	// 	}
	// 	fmt.Printf("%s: %s\n", subj, utils.ToJSON(rpl))
	// }
	//
	// checkRP(t, "MainSubj")
	// checkRP(t, "FBSubj2")
	// checkRP(t, "FBSubj1")
	// checkRP(t, "FBSubj3")
	// checkRP(t, "FBSubj4")
	// checkRP(t, "DEFAULT")

	ec := processCDR(t, "1001", false)
	checkSubjectAndCost(t, ec, 10, "*out:cgrates.org:call:MainSubj", "")

	// When calling 2001, we expect FBSubj2 to match.
	// Previously, this would have failed due to ordered fallback keys:
	// MainSubj -> FBSubj1 and FBSubj1 doesn't have FBSubj2 as fallback subject
	ec = processCDR(t, "2001", false) // MainSubj -> FBSubj2
	checkSubjectAndCost(t, ec, 20, "*out:cgrates.org:call:FBSubj2", "")

	ec = processCDR(t, "3001", false) // MainSubj -> FBSubj2 -> FBSubj1
	checkSubjectAndCost(t, ec, 30, "*out:cgrates.org:call:FBSubj1", "")
	ec = processCDR(t, "4001", false) // MainSubj -> FBSubj2 -> FBSubj1 -> FBSubj3
	checkSubjectAndCost(t, ec, 40, "*out:cgrates.org:call:FBSubj3", "")
	ec = processCDR(t, "5001", false) // MainSubj -> FBSubj2 -> FBSubj1 -> FBSubj3 -> FBSubj4
	checkSubjectAndCost(t, ec, 50, "*out:cgrates.org:call:FBSubj4", "")
	processCDR(t, "6001", true) // fallback needs to be increased by 1 for this to be successful
	checkSubjectAndCost(t, engine.EventCost{}, 0, "",
		"Destination 6001 not authorized for account: cgrates.org:1001, subject: *out:cgrates.org:call:MainSubj")
}
