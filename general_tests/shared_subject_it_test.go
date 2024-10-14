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
	"github.com/cgrates/cgrates/utils"
)

// TestSharedSubject is based on a request posted on the CGRateS google group: https://groups.google.com/g/cgrates/c/3iqrh4CJnow.
//
// The main idea of the test is to handle a scenario where there are two groups sharing one subject:
//
// Group 1: 1001,1002,1003,1004,1005
// Group 2: 1001,1010,1011,1012,1013
//
// The goal is to handle two different pricings based on the following rules:
//   - excluding 1001, any number from Group 1 calling any number from Group 2 will use a pricing of 2 units/s.
//   - excluding 1001, any number from Group 2 calling any number from Group 1 will use a pricing of 1 unit/s.
//   - if 1001 calls a number from Group 1 or is called by a number from Group 1, a pricing of 1 unit/s will be used.
//   - if 1001 calls a number from Group 2 or is called by a number from Group 2, a pricing of 2 units/s will be used.
//
// To accomplish this, the test does the following setup:
//   - creates one rating profile with pricing for each group
//   - attaches the rating profile via changing the event subject using an attribute profile
//   - sets up needed filters in order to select the correct pricing for the situation.
func TestSharedSubject(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"attributes":{
	"enabled": true,
},

"rals": {
	"enabled": true,
},

"cdrs": {
	"enabled": true,
	"attributes_conns": ["*internal"],
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
		utils.AttributesCsv: `#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_SET_SUBJECT,*any,,,,,,,false,20
cgrates.org,ATTR_SET_SUBJECT,,,,FLTR_SameGroup1,*req.Subject,*constant,Subject1,,
cgrates.org,ATTR_SET_SUBJECT,,,,FLTR_DiffGroup1,*req.Subject,*constant,Subject1,,
cgrates.org,ATTR_SET_SUBJECT,,,,FLTR_SameGroup2,*req.Subject,*constant,Subject2,,
cgrates.org,ATTR_SET_SUBJECT,,,,FLTR_DiffGroup2,*req.Subject,*constant,Subject2,,`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_Subject1,*any,RT_Subject1,*up,20,0,
DR_Subject2,*any,RT_Subject2,*up,20,0,`,
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SameGroup1,*string,~*req.Subject,1001;1002;1003;1004;1005,
cgrates.org,FLTR_SameGroup1,*string,~*req.Destination,1001;1002;1003;1004;1005,
cgrates.org,FLTR_SameGroup2,*string,~*req.Subject,1001;1010;1011;1012;1013,
cgrates.org,FLTR_SameGroup2,*string,~*req.Destination,1001;1010;1011;1012;1013,
cgrates.org,FLTR_DiffGroup1,*string,~*req.Subject,1010;1011;1012;1013,
cgrates.org,FLTR_DiffGroup1,*string,~*req.Destination,1002;1003;1004;1005,
cgrates.org,FLTR_DiffGroup2,*string,~*req.Subject,1002;1003;1004;1005,
cgrates.org,FLTR_DiffGroup2,*string,~*req.Destination,1010;1011;1012;1013,
#cgrates.org,FLTR_SET_SUBJECT,*prefix,~*req.Subject,10,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_Subject1,0,1,1s,1s,0s
RT_Subject2,0,2,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_Subject1,DR_Subject1,*any,10
RP_Subject2,DR_Subject2,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,Subject1,2014-01-14T00:00:00Z,RP_Subject1,
cgrates.org,call,Subject2,2014-01-14T00:00:00Z,RP_Subject2,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	t.Run("Cost1001->1002", func(t *testing.T) {
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
						utils.Subject:      "1001",
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
		rcvCost, ok := reply[0].Event[utils.Cost].(float64)
		if !ok {
			t.Fatal("failed to cast received cost into a float64")
		}
		if rcvCost != 120 {
			t.Errorf("expected cost to be %v, received %v", 120, rcvCost)
		}
	})

	t.Run("Cost1002->1001", func(t *testing.T) {
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
						utils.OriginID:     "processCDR2",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1002",
						utils.Subject:      "1002",
						utils.Destination:  "1001",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		rcvCost, ok := reply[0].Event[utils.Cost].(float64)
		if !ok {
			t.Fatal("failed to cast received cost into a float64")
		}
		if rcvCost != 120 {
			t.Errorf("expected cost to be %v, received %v", 120, rcvCost)
		}
	})

	t.Run("Cost1001->1010", func(t *testing.T) {
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
						utils.OriginID:     "processCDR3",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1010",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if len(reply) != 1 {
			t.Fatal("expected exactly one event in the reply slice")
		}
		rcvCost, ok := reply[0].Event[utils.Cost].(float64)
		if !ok {
			t.Fatal("failed to cast received cost into a float64")
		}
		if rcvCost != 240 {
			t.Errorf("expected cost to be %v, received %v", 240, rcvCost)
		}
	})

	t.Run("Cost1010->1001", func(t *testing.T) {
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
						utils.OriginID:     "processCDR4",
						utils.OriginHost:   "127.0.0.1",
						utils.RequestType:  utils.MetaRated,
						utils.AccountField: "1010",
						utils.Subject:      "1010",
						utils.Destination:  "1001",
						utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:        2 * time.Minute,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if len(reply) != 1 {
			t.Fatal("expected exactly one event in the reply slice")
		}
		rcvCost, ok := reply[0].Event[utils.Cost].(float64)
		if !ok {
			t.Fatal("failed to cast received cost into a float64")
		}
		if rcvCost != 240 {
			t.Errorf("expected cost to be %v, received %v", 240, rcvCost)
		}
	})
}
