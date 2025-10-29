//go:build benchmark

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

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestMsgPackBackwardsCompatibility verifies whether the msgpack library update is backwards-incompatible
// even with the TimeNotBuiltin legacy option set to true (see https://github.com/ugorji/go/issues/269).
// To use:
//   - switch to a previous commit, that's using the original library
//   - checkout this file
//   - execute the test with the action variable set to 'SET'
//   - return on the master branch and execute the test with action == 'GET'.
//
// If the 'GET' test passes, we can assume it's backwards-compatible.
//
// List of types that could be affected by the library update:
//
//   - RatingProfile
//   - Action
//   - Account
//   - utils.LoadInstance
//   - ActionTrigger
//   - ActionPlan
//   - ResourceProfile
//   - Resource
//   - StatQueueProfile
//   - StatQueue
//   - ThresholdProfile
//   - Threshold
//   - Filter
//   - RouteProfile
//   - AttributeProfile
//   - ChargerProfile
//   - DispatcherProfile
func TestMsgPackBackwardsCompatibility(t *testing.T) {
	t.SkipNow()
	// Define reference dates.
	currentDate := time.Now()
	refTime := time.Date(currentDate.Year(), currentDate.Month(), 16, 14, 20, 33, 123456789, time.UTC)
	refTimeNextYear := refTime.AddDate(1, 0, 0)

	// Create the profiles. Use them to set on 'SET' action or to validate what
	// we receive on 'GET' action.
	ratingPrf := RatingProfile{
		Id: "ratingPrf1",
		RatingPlanActivations: RatingPlanActivations{
			{
				ActivationTime: refTime,
				RatingPlanId:   "ratingPlan1",
			},
		},
	}
	actionsKey := "Actions1"
	actions := Actions{
		{
			Id:               "action1",
			ActionType:       "*log",
			ExtraParameters:  "{\"Field\":\"Value\"}",
			Filters:          []string{"FLTR_1", "FLTR_2"},
			ExpirationString: "1month",
			Weight:           10,
			Balance: &BalanceFilter{
				ID:   utils.StringPointer("balance_1"),
				Type: utils.StringPointer("*monetary"),
				Value: &utils.ValueFormula{
					Static: 15,
				},
				ExpirationDate: &refTimeNextYear,
			},
		},
	}
	account := Account{
		ID: "1001",
		BalanceMap: map[string]Balances{
			"*monetary": {
				{
					ID:             "balance_1",
					Value:          15,
					ExpirationDate: refTimeNextYear,
				},
			},
		},

		// Might want to ignore this field as it's possible
		// to get updated on its own.
		UpdateTime: refTime,
	}
	fltr := Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: refTime,
			ExpiryTime:     refTimeNextYear,
		},
	}

	cfg := config.NewDefaultCGRConfig() // datadb opts has redis defaults
	db, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatalf("failed to establish connection to redis: %v", err)
	}

	// Decide whether to set or get the profiles by manually
	// changing the following variable (SET|GET).
	action := "GET"

	switch action {
	case "SET":
		err = db.SetRatingProfileDrv(&ratingPrf)
		if err != nil {
			t.Errorf("SetRatingProfileDrv(%v) failed unexpectedly: %v", ratingPrf, err)
		}
		err = db.SetActionsDrv(actionsKey, actions)
		if err != nil {
			t.Errorf("SetActionsDrv(%q,%v) failed unexpectedly: %v", actionsKey, actions, err)
		}
		err = db.SetAccountDrv(&account)
		if err != nil {
			t.Errorf("SetAccountDrv(%v) failed unexpectedly: %v", account, err)
		}
		err = db.SetFilterDrv(&fltr)
		if err != nil {
			t.Errorf("SetFilterDrv(%v) failed unexpectedly: %v", fltr, err)
		}
	case "GET":
		if got, err := db.GetRatingProfileDrv(ratingPrf.Id); err != nil {
			t.Errorf("GetRatingProfileDrv(%q) failed unexpectedly: %v", ratingPrf.Id, err)
		} else if diff := cmp.Diff(&ratingPrf, got); diff != "" {
			t.Errorf("GetRatingProfileDrv(%q) returned unexpected profile (-want +got): \n%s", ratingPrf.Id, diff)
		}

		if got, err := db.GetActionsDrv(actionsKey); err != nil {
			t.Errorf("GetActionsDrv(%q) failed unexpectedly: %v", actionsKey, err)
		} else if diff := cmp.Diff(actions, got, cmpopts.IgnoreUnexported(Action{})); diff != "" {
			t.Errorf("GetActionsDrv(%q) returned unexpected profile (-want +got): \n%s", ratingPrf.Id, diff)
		}

		if got, err := db.GetAccountDrv(account.ID); err != nil {
			t.Errorf("GetAccountDrv(%q) failed unexpectedly: %v", account.ID, err)
		} else if diff := cmp.Diff(&account, got, cmpopts.IgnoreUnexported(Account{})); diff != "" {
			t.Errorf("GetAccountDrv(%q) returned unexpected profile (-want +got): \n%s", ratingPrf.Id, diff)
		}

		if got, err := db.GetFilterDrv(fltr.Tenant, fltr.ID); err != nil {
			t.Errorf("GetFilterDrv(%q,%q) failed unexpectedly: %v", fltr.Tenant, fltr.ID, err)
		} else if diff := cmp.Diff(&fltr, got, cmpopts.IgnoreUnexported(FilterRule{})); diff != "" {
			t.Errorf("GetFilterDrv(%q) returned unexpected profile (-want +got): \n%s", ratingPrf.Id, diff)
		}
	}
}

func BenchmarkMsgpackTime(b *testing.B) {
	type NestedStructWithTime struct {
		ID               string
		TimeField        time.Time
		TimePointerField *time.Time
	}
	type StructWithTime struct {
		ID          string
		TimeField   time.Time
		StructField NestedStructWithTime
	}
	refTime := time.Now()
	obj := StructWithTime{
		ID:        "structWithTime",
		TimeField: time.Now(),
		StructField: NestedStructWithTime{
			ID:               "nestedStructWithTime",
			TimeField:        time.Now(),
			TimePointerField: &refTime,
		},
	}
	ms, err := NewMarshaler(utils.MsgPack)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		byt, err := ms.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		var s StructWithTime
		if err := ms.Unmarshal(byt, &s); err != nil {
			b.Fatal(err)
		}
	}
}

/*
old version (github.com/cgrates/ugocodec):

	go test -tags=benchmark -bench=BenchmarkMsgpack  -run=^$ -benchtime=1s -count=5 -benchmem
	goos: linux
	goarch: amd64
	pkg: github.com/cgrates/cgrates/general_tests
	cpu: 12th Gen Intel(R) Core(TM) i7-1265U
	BenchmarkMsgpack-3        652118              1976 ns/op            2952 B/op         18 allocs/op
	BenchmarkMsgpack-3        565254              1890 ns/op            2952 B/op         18 allocs/op
	BenchmarkMsgpack-3        640384              1841 ns/op            2952 B/op         18 allocs/op
	BenchmarkMsgpack-3        663924              1855 ns/op            2952 B/op         18 allocs/op
	BenchmarkMsgpack-3        641320              1834 ns/op            2952 B/op         18 allocs/op


new version with legacy option TimeNotBuiltin set to true (current implementation):

	go test -tags=benchmark -bench=BenchmarkMsgpack  -run=^$ -benchtime=1s -count=5 -benchmem
	goos: linux
	goarch: amd64
	pkg: github.com/cgrates/cgrates/general_tests
	cpu: 12th Gen Intel(R) Core(TM) i7-1265U
	BenchmarkMsgpack-3        914578              1356 ns/op            2120 B/op         12 allocs/op
	BenchmarkMsgpack-3        901622              1340 ns/op            2120 B/op         12 allocs/op
	BenchmarkMsgpack-3        899760              1339 ns/op            2120 B/op         12 allocs/op
	BenchmarkMsgpack-3        899694              1335 ns/op            2120 B/op         12 allocs/op
	BenchmarkMsgpack-3        891693              1344 ns/op            2120 B/op         12 allocs/op

new version default:

	go test -tags=benchmark -bench=BenchmarkMsgpack  -run=^$ -benchtime=1s -count=5 -benchmem
	goos: linux
	goarch: amd64
	pkg: github.com/cgrates/cgrates/general_tests
	cpu: 12th Gen Intel(R) Core(TM) i7-1265U
	BenchmarkMsgpack-3        980895              1180 ns/op            2072 B/op          9 allocs/op
	BenchmarkMsgpack-3        978346              1181 ns/op            2072 B/op          9 allocs/op
	BenchmarkMsgpack-3       1008806              1209 ns/op            2072 B/op          9 allocs/op
	BenchmarkMsgpack-3       1013218              1192 ns/op            2072 B/op          9 allocs/op
	BenchmarkMsgpack-3        995960              1282 ns/op            2072 B/op          9 allocs/op

*/
