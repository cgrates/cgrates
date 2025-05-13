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
package rates

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRatesCostForEventRateIDxSelects(t *testing.T) {
	jsonCfg := `{
"rates": {
    "enabled": true,
    "rate_indexed_selects": true,
  },
}
`
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	if err != nil {
		t.Error(err)
	}
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(db, cfg, nil)
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
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
            "rate_indexed_selects": true,
          },
        }
        `

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	if err != nil {
		t.Fatal(err)
	}

	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(db, cfg, nil)
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
            "rate_indexed_selects": true
        }
    }`

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	if err != nil {
		t.Fatal(err)
	}

	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(db, cfg, nil)
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
					},
				},
				RateProfileID: "RATE_TEST",
			},
			expected: []string{"RT_STANDARD", "RT_PREMIUM"},
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
