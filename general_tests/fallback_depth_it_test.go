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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
TestRatingPlansFallbackDepth tests fallback_depth configuration.

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

func TestRatingPlansFallbackDepth(t *testing.T) {
	switch *dbType {
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
		Name: "TestRatingPlansFallbackDepth",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
		LogBuffer:  buf,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
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

// goos: linux
// goarch: amd64
// pkg: github.com/cgrates/cgrates/general_tests
// cpu: Intel(R) Core(TM) i7-10510U CPU @ 1.80GHz
// GetCostNoFallback
// GetCostNoFallback         	     176	   6478657 ns/op	    3382 B/op	      74 allocs/op
// GetCostNoFallback         	     169	   6567780 ns/op	    3401 B/op	      74 allocs/op
// GetCostNoFallback         	     181	   6520835 ns/op	    3376 B/op	      74 allocs/op
// GetCostNoFallback         	     174	   7159966 ns/op	    3386 B/op	      74 allocs/op
// GetCostNoFallback         	     174	   8267882 ns/op	    3386 B/op	      74 allocs/op
// GetCostFallbackSubject
// GetCostFallbackSubject    	     159	   6913372 ns/op	    3437 B/op	      74 allocs/op
// GetCostFallbackSubject    	     272	   5235043 ns/op	    3495 B/op	      74 allocs/op
// GetCostFallbackSubject    	     205	   5740902 ns/op	    3666 B/op	      74 allocs/op
// GetCostFallbackSubject    	     276	   5500421 ns/op	    3486 B/op	      74 allocs/op
// GetCostFallbackSubject    	     256	   6885937 ns/op	    3530 B/op	      74 allocs/op
// GetCostFallbackCategory
// GetCostFallbackCategory   	     147	   7001982 ns/op	    3475 B/op	      74 allocs/op
// GetCostFallbackCategory   	     163	   7489421 ns/op	    3415 B/op	      74 allocs/op
// GetCostFallbackCategory   	     164	   7048463 ns/op	    3413 B/op	      74 allocs/op
// GetCostFallbackCategory   	     180	   7447074 ns/op	    3388 B/op	      74 allocs/op
// GetCostFallbackCategory   	     188	   5950831 ns/op	    3360 B/op	      74 allocs/op
// GetCostFallbackSubjectAndCategory
// GetCostFallbackSubjectAndCategory         	     162	   7412799 ns/op	    3581 B/op	      74 allocs/op
// GetCostFallbackSubjectAndCategory         	     165	   7436020 ns/op	    3569 B/op	      74 allocs/op
// GetCostFallbackSubjectAndCategory         	     192	   7406939 ns/op	    3507 B/op	      74 allocs/op
// GetCostFallbackSubjectAndCategory         	     168	   7313485 ns/op	    3568 B/op	      74 allocs/op
// GetCostFallbackSubjectAndCategory         	     180	   7771994 ns/op	    3530 B/op	      74 allocs/op
func TestRatingPlansFallbackCategoryAndSubject(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"rals": {
	"enabled": true
},

"apiers": {
	"enabled": true
}

}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_NO_FALLBACK,DST_TEST,RT_1UNITS_PER_SEC,*up,1,0,
DR_FALLBACK_ANY_SUBJECT,DST_TEST,RT_2UNITS_PER_SEC,*up,1,0,
DR_FALLBACK_ANY_CATEGORY,DST_TEST,RT_3UNITS_PER_SEC,*up,1,0,
DR_FALLBACK_ANY_SUBJECT_AND_CATEGORY,DST_TEST,RT_4UNITS_PER_SEC,*up,1,0,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_TEST,+49`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_1UNITS_PER_SEC,0,1,1s,1s,0s
RT_2UNITS_PER_SEC,0,2,1s,1s,0s
RT_3UNITS_PER_SEC,0,3,1s,1s,0s
RT_4UNITS_PER_SEC,0,4,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_NO_FALLBACK,DR_NO_FALLBACK,*any,10
RP_FALLBACK_ANY_SUBJECT,DR_FALLBACK_ANY_SUBJECT,*any,10
RP_FALLBACK_ANY_CATEGORY,DR_FALLBACK_ANY_CATEGORY,*any,10
RP_FALLBACK_ANY_SUBJECT_AND_CATEGORY,DR_FALLBACK_ANY_SUBJECT_AND_CATEGORY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_NO_FALLBACK,
cgrates.org,call,*any,2014-01-14T00:00:00Z,RP_FALLBACK_ANY_SUBJECT,
cgrates.org,*any,1001,2014-01-14T00:00:00Z,RP_FALLBACK_ANY_CATEGORY,
cgrates.org,*any,*any,2014-01-14T00:00:00Z,RP_FALLBACK_ANY_SUBJECT_AND_CATEGORY,`,
	}

	testEnv := TestEnvironment{
		Name: "TestRatingPlansFallbackCategoryAndSubject",
		// Encoding:   *encoding,
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _, shutdown, err := testEnv.Setup(t, *waitRater)
	if err != nil {
		t.Fatal(err)
	}

	defer shutdown()

	t.Run("GetCostNoFallback", func(t *testing.T) {
		getCostArgs := &v1.AttrGetCost{
			Tenant:      "cgrates.org",
			Category:    "call",
			Subject:     "1001",
			Destination: "+491234567890",
			AnswerTime:  "*now",
			Usage:       "10s",
		}
		var ec engine.EventCost
		if err := client.Call(context.Background(), utils.APIerSv1GetCost, getCostArgs, &ec); err != nil {
			t.Error(err)
		} else if *ec.Cost != 10.000000 {
			t.Errorf("Unexpected cost received: %f", *ec.Cost)
		}

		expectedRP := "RP_NO_FALLBACK"
		if val, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "RatingPlanID"}); err != nil {
			t.Error(err)
		} else if val != expectedRP {
			t.Errorf("expected %v, received %v", expectedRP, val)
		}
	})
	t.Run("GetCostFallbackSubject", func(t *testing.T) {
		getCostArgs := &v1.AttrGetCost{
			Tenant:      "cgrates.org",
			Category:    "call",
			Subject:     "1234",
			Destination: "+491234567890",
			AnswerTime:  "*now",
			Usage:       "10s",
		}
		var ec engine.EventCost
		if err := client.Call(context.Background(), utils.APIerSv1GetCost, getCostArgs, &ec); err != nil {
			t.Error(err)
		} else if *ec.Cost != 20.000000 {
			t.Errorf("Unexpected cost received: %f", *ec.Cost)
		}

		expectedRP := "RP_FALLBACK_ANY_SUBJECT"
		if val, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "RatingPlanID"}); err != nil {
			t.Error(err)
		} else if val != expectedRP {
			t.Errorf("expected %v, received %v", expectedRP, val)
		}
	})
	t.Run("GetCostFallbackCategory", func(t *testing.T) {
		getCostArgs := &v1.AttrGetCost{
			Tenant:      "cgrates.org",
			Category:    "sms",
			Subject:     "1001",
			Destination: "+491234567890",
			AnswerTime:  "*now",
			Usage:       "10s",
		}
		var ec engine.EventCost
		if err := client.Call(context.Background(), utils.APIerSv1GetCost, getCostArgs, &ec); err != nil {
			t.Error(err)
		} else if *ec.Cost != 30.000000 {
			t.Errorf("Unexpected cost received: %f", *ec.Cost)
		}

		expectedRP := "RP_FALLBACK_ANY_CATEGORY"
		if val, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "RatingPlanID"}); err != nil {
			t.Error(err)
		} else if val != expectedRP {
			t.Errorf("expected %v, received %v", expectedRP, val)
		}
	})
	t.Run("GetCostFallbackSubjectAndCategory", func(t *testing.T) {
		getCostArgs := &v1.AttrGetCost{
			Tenant:      "cgrates.org",
			Category:    "sms",
			Subject:     "1234",
			Destination: "+491234567890",
			AnswerTime:  "*now",
			Usage:       "10s",
		}
		var ec engine.EventCost
		if err := client.Call(context.Background(), utils.APIerSv1GetCost, getCostArgs, &ec); err != nil {
			t.Error(err)
		} else if *ec.Cost != 40.000000 {
			t.Errorf("Unexpected cost received: %f", *ec.Cost)
		}

		expectedRP := "RP_FALLBACK_ANY_SUBJECT_AND_CATEGORY"
		if val, err := ec.FieldAsString([]string{"Charges[0]", "Rating", "RatingFilter", "RatingPlanID"}); err != nil {
			t.Error(err)
		} else if val != expectedRP {
			t.Errorf("expected %v, received %v", expectedRP, val)
		}
	})
}
