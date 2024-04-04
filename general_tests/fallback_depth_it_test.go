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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
TestFallbackDepth tests fallback_depth configuration.

Previously, the max depth was always 3. The test ensures that the functionality works properly when
the depth exceeds the previous hard-coded value.

The test steps are as follows:

1. Create 3 rating plans:

	- a dummy rating plan that never matches
	- one that matches only when destination is 1002
	- one that matches any destination

2. Define 5 subjects in the following manner:

   	- a main subject that will be assigned to the event args. This subject will have a fallback
	subject as backup
   	- 4 fallback subjects, each having the next one as backup
	- main subject and first 2 fallback subjects will use the dummy rating plan
	- third fallback subject will have a rating plan defined for destination 1002
	- fourth fallback subject will have a rating plan defined for any destination

3. Configure fallback_depth to be 4.
4. Process a CDR where the destination is 1002. This is expected to return CostDetails mentioning that
   FallbackSubject3 was taken into consideration during rating.
5. Process CDR where the destination is 1003. This is expected to encounter an error log saying that the
   destination is not authorized. For it to reach the fourth fallback subject, a fallback depth of 5 would
   be required.
*/

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
DR_ANY,*any,RT_ANY,*up,20,0,
DR_1002,DST_1002,RT_1002,*up,20,0,
DUMMY_DR,DUMMY_DST,DUMMY_RT,*up,20,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DUMMY_DST,1234
DST_1002,1002`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,1,1s,1s,0s
RT_1002,0,1,1s,1s,0s
DUMMY_RT,0,0.1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10
RP_1002,DR_1002,*any,10
DUMMY_RP,DUMMY_DR,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,MainSubject,2014-01-01T00:00:00Z,DUMMY_RP,FallbackSubject1
cgrates.org,call,FallbackSubject1,2014-01-01T00:00:00Z,DUMMY_RP,FallbackSubject2
cgrates.org,call,FallbackSubject2,2014-01-01T00:00:00Z,DUMMY_RP,FallbackSubject3
cgrates.org,call,FallbackSubject3,2014-01-01T00:00:00Z,RP_1002,FallbackSubject4
cgrates.org,call,FallbackSubject4,2014-01-01T00:00:00Z,RP_ANY,`,
	}

	buf := &bytes.Buffer{}
	testEnv := TestEnvironment{
		Name: "TestFallbackDepth",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
		LogBuffer:  buf,
	}
	client, _, shutdown, err := testEnv.Setup(t, *utils.WaitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("ProcessCdrFallbackSuccess", func(t *testing.T) {
		var reply []*utils.EventWithFlags
		err := client.Call(context.Background(), utils.CDRsV2ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR1",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Subject:      "MainSubject",
						utils.Destination:  "1002",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		// Convert CostDetails value from map[string]any to engine.EventCost in order to be able to retrieve
		// the RatingSubject used using FieldAsString method.
		ecIface, has := reply[0].Event[utils.CostDetails]
		if !has {
			t.Fatalf("expected CGREvent to have CostDetails populated")
		}
		b, err := json.Marshal(ecIface)
		if err != nil {
			t.Fatal(err)
		}
		var ec engine.EventCost
		err = json.Unmarshal(b, &ec)
		if err != nil {
			t.Fatal(err)
		}

		expected := `*out:cgrates.org:call:FallbackSubject3`
		subj, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "Subject"})
		if err != nil {
			t.Fatal(err)
		}
		if subj != expected {
			t.Errorf("expected %s, received %s", expected, subj)
		}

		rcvCost := reply[0].Event[utils.Cost]
		if rcvCost != 120. {
			t.Errorf("expected cost to be %v, received %v", 120., rcvCost)
		}
	})

	t.Run("ProcessCdrFallbackFail", func(t *testing.T) {
		var reply []*utils.EventWithFlags
		err := client.Call(context.Background(), utils.CDRsV2ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event2",
					Event: map[string]any{
						utils.RunID:        "run_1",
						utils.Tenant:       "cgrates.org",
						utils.Category:     "call",
						utils.ToR:          utils.MetaVoice,
						utils.OriginID:     "processCDR2",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Subject:      "MainSubject",
						utils.Destination:  "1003",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(),
			"Destination 1003 not authorized for account: cgrates.org:1001, subject: *out:cgrates.org:call:MainSubject") {
			t.Fatal("expected unauthorized destination log")
		}

		rcvCost := reply[0].Event[utils.Cost]
		if rcvCost != -1. {
			t.Errorf("expected cost to be %v, received %v", -1., rcvCost)
		}
	})

}
