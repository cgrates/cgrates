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

package v1

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDebitUsageWithOptions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	engine.SetDataStorage(dm)
	apierDebit := &APIerSv1{
		DataManager: dm,
		Config:      cfg,
		Responder: &engine.Responder{
			MaxComputedUsage: cfg.RalsCfg().MaxComputedUsage,
		},
	}

	cgrTenant := "cgrates.org"
	cgrAcnt1 := &engine.Account{
		ID: utils.ConcatenatedKey(cgrTenant, "account1"),
		BalanceMap: map[string]engine.Balances{
			utils.MetaMonetary: {
				{
					Value:  10,
					Weight: 10,
				},
			},
		},
	}
	if err := dataDB.SetAccountDrv(cgrAcnt1); err != nil {
		t.Error(err)
	}
	dstDe := &engine.Destination{Id: "*any", Prefixes: []string{"*any"}}
	if err := dataDB.SetDestinationDrv(dstDe, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dataDB.SetReverseDestinationDrv(dstDe.Id, dstDe.Prefixes, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	rp1 := &engine.RatingPlan{
		Id: "RP1",
		Timings: map[string]*engine.RITiming{
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*engine.RIRate{
			"b457f86d": {
				ConnectFee: 0,
				Rates: []*engine.RGRate{
					{
						GroupIntervalStart: 0,
						Value:              0.03,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.MetaRoundingMiddle,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]engine.RPRateList{
			dstDe.Id: []*engine.RPRate{
				{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	if err := dm.SetRatingPlan(rp1); err != nil {
		t.Error(err)
	}

	rpfl := &engine.RatingProfile{Id: "*out:cgrates.org:call:account1",
		RatingPlanActivations: engine.RatingPlanActivations{&engine.RatingPlanActivation{
			ActivationTime: time.Date(2001, 1, 1, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rp1.Id,
			FallbackKeys:   []string{},
		}},
	}
	if err := dm.SetRatingProfile(rpfl); err != nil {
		t.Error(err)
	}

	usageRecord := &engine.UsageRecord{
		Tenant:      cgrTenant,
		Account:     "account1",
		Destination: "*any",
		Usage:       "1",
		ToR:         utils.MetaMonetary,
		Category:    "call",
		SetupTime:   time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
		AnswerTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
	}

	var reply string
	if err := apierDebit.DebitUsageWithOptions(
		context.Background(),
		&AttrDebitUsageWithOptions{
			UsageRecord:          &engine.UsageRecordWithAPIOpts{UsageRecord: usageRecord},
			AllowNegativeAccount: false}, &reply); err != nil {
		t.Error(err)
	}

	// Reload the account and verify that the usage of $1 was removed from the monetary balance
	resolvedAccount, err := dataDB.GetAccountDrv(cgrAcnt1.ID)
	if err != nil {
		t.Error(err)
	}
	eAcntVal := 9.0
	if resolvedAccount.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal,
			resolvedAccount.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}
