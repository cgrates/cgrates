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

package tpes

import (
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPRates(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: &engine.DataDBMock{
		GetRateProfileDrvF: func(ctx *context.Context, str1, str2 string) (*utils.RateProfile, error) {
			rt := &utils.RateProfile{
				Tenant:    utils.CGRateSorg,
				ID:        "TEST_RATE_IT_TEST",
				FilterIDs: []string{"*string:~*req.Account:dan"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				MaxCostStrategy: "*free",
				Rates: map[string]*utils.Rate{
					"RT_WEEK": {
						ID: "RT_WEEK",
						Weights: utils.DynamicWeights{
							{
								Weight: 0,
							},
						},
						ActivationTimes: "* * * * 1-5",
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimal(0, 0),
							},
						},
					},
				},
			}
			return rt, nil
		},
	}}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMng)
	exp := &TPRates{
		dm: dm,
	}
	rcv := newTPRates(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsRates(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpRt := TPRates{
		dm: dm,
	}
	rt := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	tpRt.dm.SetRateProfile(context.Background(), rt, false, false)
	err := tpRt.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TEST_RATE_TEST"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsRatesNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpRt := TPRates{
		dm: nil,
	}
	rt := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	tpRt.dm.SetRateProfile(context.Background(), rt, false, false)
	err := tpRt.exportItems(context.Background(), wrtr, "cgrates.org", []string{"fltr_for_prf"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsRatesIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tpRt := TPRates{
		dm: dm,
	}
	rt := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := tpRt.dm.SetRateProfile(context.Background(), rt, false, true); err != nil {
		t.Error(err)
	}
	err := tpRt.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TEST_RATE"})
	errExpect := "<NOT_FOUND> cannot find RateProfile with id: <TEST_RATE>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
