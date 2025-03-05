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
	db := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
