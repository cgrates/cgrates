// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rates

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRatesCostForEventRateIDxSelects(t *testing.T) {
	jsonCfg := `{
"rates": {
    "enabled": true,
    "rateIndexedSelects": true,
  },
}
`
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	locker := engine.NewLocker(cfg)
	if err != nil {
		t.Error(err)
	}
	db, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rts := NewRateS(cfg, fltrs, dm)

	rtPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT_ALWAYS": {
				ID: "RT_ALWAYS",
				FilterIDs: []string{
					"*string:~*req.ToR:*voice"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Increment:     utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(2, 0),
						//FixedFee:      utils.Float64Pointer(0.3),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				FilterIDs: []string{"*prefix:~*req.Destination:+332",
					"*string:~*req.RequestType:*postpaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(4, 1),
						Increment:     utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(3, 1),
						//FixedFee:      utils.Float64Pointer(0.5),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rtPrf, false, true); err != nil {
		t.Error(err)
	}

	//math the rates with true rates index selects from config
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.RequestType:  "*postpaid",
			utils.Destination:  "+332145",
		},
		APIOpts: map[string]any{
			utils.OptsRatesUsage: "1m24s",
		},
	}
	usg, err := utils.NewDecimalFromUsage("1m24s")
	if err != nil {
		t.Error(err)
	}
	var rpCost utils.RateProfileCost
	expRpCost := &utils.RateProfileCost{
		ID:   "RATE_1",
		Cost: utils.NewDecimal(1120000000000000, 4),
		CostIntervals: []*utils.RateSIntervalCost{
			{
				Increments: []*utils.RateSIncrementCost{
					{
						Usage:             usg,
						RateID:            "random",
						RateIntervalIndex: 0,
						CompressFactor:    840000000000,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"random": {
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(4, 1),
				Increment:     utils.NewDecimal(1, 1),
				Unit:          utils.NewDecimal(3, 1),
				//	FixedFee:      utils.NewDecimal(5, 1),
			},
		},
	}

	if err := rts.V1CostForEvent(context.Background(), ev,
		&rpCost); err != nil {
		t.Error(err)
	} else if !rpCost.Equals(expRpCost) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRpCost), utils.ToJSON(rpCost))
	}

	cfg.RateSCfg().RateIndexedSelects = false
	rts = NewRateS(cfg, fltrs, dm)
	if err := rts.V1CostForEvent(context.Background(), ev,
		&rpCost); err != nil {
		t.Error(err)
	} else if !rpCost.Equals(expRpCost) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRpCost), utils.ToJSON(rpCost))
	}
}

func TestRatesCostForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	rateS := NewRateS(cfg, nil, dm)

	ev := &utils.CGREvent{
		Tenant: "tenant",
		ID:     "ID",
		Event:  nil,
		APIOpts: map[string]any{
			utils.OptsRatesProfileIDs: []string{"rtID"},
		},
	}

	rpCost := &utils.RateProfileCost{}
	err := rateS.V1CostForEvent(context.Background(), ev, rpCost)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
	expected2 := &utils.RateProfileCost{}
	if !reflect.DeepEqual(rpCost, expected2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected2, rpCost)
	}
}

func TestV1RateProfilesForEvent(t *testing.T) {
	jsonCfg := `{
        "rates": {
            "enabled": true,
            "rateIndexedSelects": true,
          },
        }
        `

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	locker := engine.NewLocker(cfg)
	if err != nil {
		t.Fatal(err)
	}

	db, err := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewRateS(cfg, fltrs, dm)

	ratePrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_MATCH",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				FilterIDs: []string{
					"*prefix:~*req.Destination:+33",
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Increment:     utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := dm.SetRateProfile(context.Background(), ratePrf, false, true); err != nil {
		t.Fatal(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "+33123456789",
		},
	}

	var matched []string
	if err := rS.V1RateProfilesForEvent(context.Background(), ev, &matched); err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := []string{"RP_MATCH"}
	if !reflect.DeepEqual(matched, expected) {
		t.Errorf("Expected matched profiles: %v, got: %v", expected, matched)
	}
}

func TestV1RateProfileRatesForEvent(t *testing.T) {
	jsonCfg := `{
        "rates": {
            "enabled": true,
            "rateIndexedSelects": true
        }
    }`

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	locker := engine.NewLocker(cfg)
	if err != nil {
		t.Fatal(err)
	}

	db, err := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewRateS(cfg, fltrs, dm)

	ratePrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_TEST",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT_STANDARD": {
				ID: "RT_STANDARD",
				FilterIDs: []string{
					"*string:~*req.ToR:*voice",
					"*gt:~*req.Cost:10",
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Increment:     utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(2, 0),
					},
				},
			},
			"RT_PREMIUM": {
				ID: "RT_PREMIUM",
				FilterIDs: []string{
					"*prefix:~*req.Destination:+44",
					"*string:~*req.RequestType:*postpaid",
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Increment:     utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := dm.SetRateProfile(context.Background(), ratePrf, false, true); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name     string
		args     *utils.CGREventWithRateProfile
		expected []string
		err      error
	}{
		{
			name: "NilArgs",
			args: nil,
			err:  utils.NewErrMandatoryIeMissing(utils.CGREventString),
		},
		{
			name: "EmptyRateProfileID",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				RateProfileID: "",
			},
			err: utils.NewErrMandatoryIeMissing(utils.RateProfileID),
		},
		{
			name: "RateProfileNotFound",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
				},
				RateProfileID: "NON_EXISTENT",
			},
			err: utils.ErrNotFound,
		},
		{
			name: "NoMatchingRates",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.ToR:          "*data",
					},
				},
				RateProfileID: "RATE_TEST",
			},
			err: utils.ErrNotFound,
		},
		{
			name: "OneMatchingRate",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.ToR:          "*voice",
						utils.Cost:         11,
					},
				},
				RateProfileID: "RATE_TEST",
			},
			expected: []string{"RT_STANDARD"},
		},
		{
			name: "TwoMatchingRates",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.ToR:          "*voice",
						utils.Destination:  "+44123456789",
						utils.RequestType:  "*postpaid",
						utils.Cost:         11,
					},
				},
				RateProfileID: "RATE_TEST",
			},
			expected: []string{"RT_STANDARD", "RT_PREMIUM"},
		},
		{
			name: "DefaultTenant",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.ToR:          "*voice",
						utils.Cost:         20,
					},
				},
				RateProfileID: "RATE_TEST",
			},
			expected: []string{"RT_STANDARD"},
		},
		{
			name: "FilterNotPassing",
			args: &utils.CGREventWithRateProfile{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.ToR:          "*voice",
						utils.Cost:         5,
					},
				},
				RateProfileID: "RATE_TEST",
			},
			err: utils.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rateIDs []string
			err := rS.V1RateProfileRatesForEvent(context.Background(), tc.args, &rateIDs)

			if tc.err != nil {
				if err == nil || err.Error() != tc.err.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if len(rateIDs) != len(tc.expected) {
				t.Errorf("Expected %d matching rates, got %d", len(tc.expected), len(rateIDs))
			} else {
				expectedSet := make(map[string]bool)
				for _, id := range tc.expected {
					expectedSet[id] = true
				}

				resultSet := make(map[string]bool)
				for _, id := range rateIDs {
					resultSet[id] = true
				}

				if !reflect.DeepEqual(resultSet, expectedSet) {
					t.Errorf("Expected rates: %v, got: %v", tc.expected, rateIDs)
				}
			}
		})
	}
}

func TestV1CostForEventFixedFeeIntervalStart(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Fatal(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(3, 1),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name          string
		intervalStart string
		usage         string
		expCost       *utils.Decimal
	}{
		{
			name:    "From start",
			usage:   "10m",
			expCost: utils.NewDecimal(14, 1),
		},
		{
			name:    "Less than one increment from start",
			usage:   "20s",
			expCost: utils.NewDecimal(5, 1),
		},
		{
			name:          "Interval start on boundary",
			intervalStart: "1m",
			usage:         "1m",
			expCost:       utils.NewDecimal(1, 1),
		},
		{
			name:          "Interval start, rounds up increment",
			intervalStart: "1m",
			usage:         "90s",
			expCost:       utils.NewDecimal(2, 1),
		},
		{
			name:          "Usage crosses boundary",
			intervalStart: "30s",
			usage:         "90s",
			expCost:       utils.NewDecimal(6, 1),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				ID: "RATE_1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsRatesStartTime:     time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
					utils.OptsRatesUsage:         tc.usage,
					utils.OptsRatesIntervalStart: tc.intervalStart,
				},
			}

			var rpCost utils.RateProfileCost
			if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
				t.Error(err)
			}
			if rpCost.Cost.Compare(tc.expCost) != 0 {
				t.Errorf("Expected cost %v, received %v", tc.expCost, rpCost.Cost)
			}
		})
	}
}

func TestV1RateProfilesForEventErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*gt:~*req.Cost:10"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name   string
		args   *utils.CGREvent
		expErr string
	}{
		{
			name:   "Nil args",
			args:   nil,
			expErr: "MANDATORY_IE_MISSING: [CGREvent]",
		},
		{
			name: "no index match",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "2002",
					utils.Cost:         20,
				},
			},
			expErr: "NOT_FOUND",
		},
		{
			name: "filter not passing",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Cost:         5,
				},
			},
			expErr: "NOT_FOUND",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rpIDs []string
			err := rateS.V1RateProfilesForEvent(context.Background(), tc.args, &rpIDs)
			if err == nil || err.Error() != tc.expErr {
				t.Errorf("Expected error %v, received %v", tc.expErr, err)
			}
		})
	}
}

func TestV1RateProfilesForEventInvalidFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rPrfID",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	ev := &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var rpIDs *[]string

	rPrf.FilterIDs = []string{"*inv:~*req.Account:1001"}
	expErr := "Unsupported filter Type: *inv"
	if err := rateS.V1RateProfilesForEvent(context.Background(), ev, rpIDs); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %v, received %v", expErr, err)
	}

	if err := data.RemoveRateProfileDrv(context.Background(), rPrf.Tenant, rPrf.ID, rpIDs); err != nil {
		t.Error(err)
	}
	cacheS.Clear(nil)
	if err := rateS.V1RateProfilesForEvent(context.Background(), ev, rpIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v, received %v", utils.ErrNotFound, err)
	}
}

func TestRateSV1CostForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT_STANDART": {
				ID: "RT_STANDART",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(3, 1),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 25 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		startTime time.Time
		usage     string
		expCost   *utils.Decimal
	}{
		{
			name:      "normal day",
			startTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(7, 1),
		},
		{
			name:      "normal day, short call",
			startTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
			usage:     "30s",
			expCost:   utils.NewDecimal(5, 1),
		},
		{
			name:      "normal day, long call",
			startTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
			usage:     "10m",
			expCost:   utils.NewDecimal(14, 1),
		},
		{
			name:      "christmas day",
			startTime: time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(6, 2),
		},
		{
			name:      "christmas day long call",
			startTime: time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC),
			usage:     "10m",
			expCost:   utils.NewDecimal(20, 2),
		},
		{
			name:      "christmas day, starts during call",
			startTime: time.Date(2026, 12, 24, 23, 59, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(54, 2),
		},
		{
			name:      "christmas day, ends during call",
			startTime: time.Date(2026, 12, 25, 23, 59, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(22, 2),
		},
		{
			name:      "call after christmas day",
			startTime: time.Date(2026, 12, 26, 0, 0, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(7, 1),
		},
		{
			name:      "christmas day starts at midnight",
			startTime: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
			usage:     "3m",
			expCost:   utils.NewDecimal(6, 2),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsRatesStartTime: tc.startTime,
					utils.OptsRatesUsage:     tc.usage,
				},
			}
			rpCost := utils.RateProfileCost{}
			if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(rpCost.Cost, tc.expCost) {
				t.Errorf("Expected %v, recieved %v", tc.expCost, rpCost.Cost)
			}
		})
	}
}

func TestV1CostForEventRerateAfterPriceChange(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
			utils.OptsRatesUsage:     utils.NewDecimal(3, 0),
		},
	}

	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	var rpCost utils.RateProfileCost
	if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(6, 1)) != 0 {
		t.Errorf("Expected cost %v, received %v", utils.NewDecimal(6, 1), rpCost.Cost)
	}

	rPrf.Rates["RATE1"].IntervalRates[0].RecurrentFee = utils.NewDecimal(4, 1)
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(12, 1)) != 0 {
		t.Errorf("Expected cost %v, received %v", utils.NewDecimal(12, 1), rpCost.Cost)
	}

	rPrf.Rates["RATE1"].IntervalRates[0].RecurrentFee = utils.NewDecimal(1, 1)
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(3, 1)) != 0 {
		t.Errorf("Expected cost %v, received %v", utils.NewDecimal(3, 1), rpCost.Cost)
	}

	rPrf.Rates["RATE1"].IntervalRates[0].RecurrentFee = utils.NewDecimal(0, 0)
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Expected cost %v, received %v", utils.NewDecimal(0, 0), rpCost.Cost)
	}

	rPrf.Rates["RATE1"].IntervalRates[0].RecurrentFee = nil
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Expected cost %v, received %v", utils.NewDecimal(0, 0), rpCost.Cost)
	}
}

func TestV1CostForEventDataBytes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_DATA",
		FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.Category:*data"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Rates: map[string]*utils.Rate{
			"RT_DATA": {
				ID: "RT_DATA",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 0),
						Unit:          utils.NewDecimal(10, 0),
						Increment:     utils.NewDecimal(10, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name    string
		usage   *utils.Decimal
		expCost *utils.Decimal
	}{
		{
			name:    "two units",
			usage:   utils.NewDecimal(20, 0),
			expCost: utils.NewDecimal(4, 0),
		},
		{
			name:    "exactly 3 increment",
			usage:   utils.NewDecimal(30, 0),
			expCost: utils.NewDecimal(6, 0),
		},
		{
			name:    "partial increment rounds up above 1 increment",
			usage:   utils.NewDecimal(15, 0),
			expCost: utils.NewDecimal(4, 0),
		},
		{
			name:    "partial increment rounds up above 2 increments",
			usage:   utils.NewDecimal(25, 0),
			expCost: utils.NewDecimal(6, 0),
		},
		{
			name:    "usage below one increment",
			usage:   utils.NewDecimal(9, 0),
			expCost: utils.NewDecimal(2, 0),
		},
		{
			name:    "one byte above one increment",
			usage:   utils.NewDecimal(11, 0),
			expCost: utils.NewDecimal(4, 0),
		},
		{
			name:    "one byte above 2 increments",
			usage:   utils.NewDecimal(21, 0),
			expCost: utils.NewDecimal(6, 0),
		},
		{
			name:    "0 usage",
			usage:   utils.NewDecimal(0, 0),
			expCost: utils.NewDecimal(0, 0),
		},
		{
			name:    "usage at increment boundary",
			usage:   utils.NewDecimal(10, 0),
			expCost: utils.NewDecimal(2, 0),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Category:     "*data",
				},
				APIOpts: map[string]any{
					utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
					utils.OptsRatesUsage:     tc.usage,
				},
			}
			var rpCost utils.RateProfileCost
			if err := rateS.V1CostForEvent(context.Background(), ev, &rpCost); err != nil {
				t.Error(err)
			}
			if rpCost.Cost.Compare(tc.expCost) != 0 {
				t.Errorf("Expected cost %v, received %v", tc.expCost, rpCost.Cost)
			}
		})
	}
}

func TestV1CostForEventErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	rp1 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_1001",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	rp2 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_1002",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID:        "RATE1",
				FilterIDs: []string{"*string:~*req.ToR:*voice"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rp1, false, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetRateProfile(context.Background(), rp2, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name      string
		account   string
		opts      map[string]any
		verbosity int
		expErr    string
	}{
		{
			name:    "Invalid ProfileIDs APIOpts",
			account: "1001",
			opts: map[string]any{
				utils.OptsRatesStartTime:  time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:      utils.NewDecimal(3, 0),
				utils.OptsRatesProfileIDs: 123,
			},
			verbosity: 1000,
			expErr:    "cannot convert field: 123 to []string",
		},
		{
			name:    "Invalid ProfileIgnoreFilters format",
			account: "1001",
			opts: map[string]any{
				utils.OptsRatesStartTime:       time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:           utils.NewDecimal(3, 0),
				utils.MetaProfileIgnoreFilters: "trueee",
			},
			verbosity: 1000,
			expErr:    "strconv.ParseBool: parsing \"trueee\": invalid syntax",
		},
		{
			name:    "No rate matches event",
			account: "1002",
			opts: map[string]any{
				utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     utils.NewDecimal(3, 0),
			},
			verbosity: 1000,
			expErr:    "NOT_FOUND",
		},
		{
			name:    "Invalid OptsRatesUsage format",
			account: "1001",
			opts: map[string]any{
				utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "inv",
			},
			verbosity: 1,
			expErr:    "can't convert <inv> to decimal",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rateS.cfg.RateSCfg().Verbosity = tc.verbosity
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event: map[string]any{
					utils.AccountField: tc.account,
				},
				APIOpts: tc.opts,
			}
			var rpCost utils.RateProfileCost
			err := rateS.V1CostForEvent(context.Background(), ev, &rpCost)
			if err == nil || err.Error() != tc.expErr {
				t.Errorf("Expected error %v, received %v", tc.expErr, err)
			}
		})
	}
}
