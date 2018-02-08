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
package engine

import (
	"github.com/cgrates/cgrates/utils"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	marsh = NewCodecMsgpackMarshaler()
)

func init() {
	dm.DataDB().Flush("")
	populateDB()
}

func populateDB() {
	ats := []*Action{
		&Action{ActionType: "*topup",
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Value: &utils.ValueFormula{Static: 10}}},
		&Action{ActionType: "*topup",
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.VOICE),
				Weight:         utils.Float64Pointer(20),
				Value:          &utils.ValueFormula{Static: 10 * float64(time.Second)},
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT"))}},
	}

	ats1 := []*Action{
		&Action{ActionType: "*topup",
			Balance: &BalanceFilter{
				Type:  utils.StringPointer(utils.MONETARY),
				Value: &utils.ValueFormula{Static: 10}}, Weight: 10},
		&Action{ActionType: "*reset_account", Weight: 20},
	}

	minu := &Account{
		ID: "vdf:minu",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: 50}},
			utils.VOICE: Balances{
				&Balance{Value: 200 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"), Weight: 10},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	broker := &Account{
		ID: "vdf:broker",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	luna := &Account{
		ID: "vdf:luna",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: 0, Weight: 20},
			}},
	}
	// this is added to test if csv load tests account will not overwrite balances
	minitsboy := &Account{
		ID: "vdf:minitsboy",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"),
					Weight:         10, RatingSubject: "rif"},
				&Balance{Value: 100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			},
			utils.MONETARY: Balances{
				&Balance{Value: 100, Weight: 10},
			},
		},
	}
	max := &Account{
		ID: "cgrates.org:max",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: 11, Weight: 20},
			}},
	}
	money := &Account{
		ID: "cgrates.org:money",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: 10000, Weight: 10},
			}},
	}
	if dm.dataDB != nil {
		dm.SetActions("TEST_ACTIONS", ats, utils.NonTransactional)
		dm.SetActions("TEST_ACTIONS_ORDER", ats1, utils.NonTransactional)
		dm.DataDB().SetAccount(broker)
		dm.DataDB().SetAccount(minu)
		dm.DataDB().SetAccount(minitsboy)
		dm.DataDB().SetAccount(luna)
		dm.DataDB().SetAccount(max)
		dm.DataDB().SetAccount(money)
	} else {
		log.Fatal("Could not connect to db!")
	}
}

func debitTest(t *testing.T, wg *sync.WaitGroup) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 30, 59, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call",
		Tenant: "cgrates.org", Account: "moneyp", Subject: "nt",
		Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	if _, err := cd.Debit(); err != nil {
		t.Errorf("Error debiting balance: %s", err)
	}
	wg.Done()
}

func TestSerialDebit(t *testing.T) {
	var wg sync.WaitGroup
	initialBalance := 10000.0
	moneyConcurent := &Account{
		ID: "cgrates.org:moneyp",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: initialBalance, Weight: 10},
			}},
	}
	if err := dm.DataDB().SetAccount(moneyConcurent); err != nil {
		t.Error(err)
	}
	debitsToDo := 50
	for i := 0; i < debitsToDo; i++ {
		wg.Add(1)
		debitTest(t, &wg)
	}
	wg.Wait()
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 30, 59, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call",
		Tenant: "cgrates.org", Account: "moneyp", Subject: "nt",
		Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}

	acc, err := cd.getAccount()
	if err != nil {
		t.Errorf("Error debiting balance: %+v", err)
	}
	expBalance := initialBalance - float64(debitsToDo*60)
	if acc.BalanceMap[utils.MONETARY][0].GetValue() != expBalance {
		t.Errorf("Balance does not match: %f, expected %f",
			acc.BalanceMap[utils.MONETARY][0].GetValue(), expBalance)
	}

}

func TestParallelDebit(t *testing.T) {
	var wg sync.WaitGroup
	initialBalance := 10000.0
	moneyConcurent := &Account{
		ID: "cgrates.org:moneyp",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: initialBalance, Weight: 10},
			}},
	}
	if err := dm.DataDB().SetAccount(moneyConcurent); err != nil {
		t.Error(err)
	}
	debitsToDo := 50
	for i := 0; i < debitsToDo; i++ {
		wg.Add(1)
		go debitTest(t, &wg)
	}
	wg.Wait()
	time.Sleep(time.Duration(10 * time.Millisecond))
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 30, 59, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Account: "moneyp", Subject: "nt", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}

	acc, err := cd.getAccount()
	if err != nil {
		t.Errorf("Error debiting balance: %+v", err)
	}
	expBalance := initialBalance - float64(debitsToDo*60)
	if acc.BalanceMap[utils.MONETARY][0].GetValue() != expBalance {
		t.Errorf("Balance does not match: %f, expected %f", acc.BalanceMap[utils.MONETARY][0].GetValue(), expBalance)
	}
	/*
		out, err := json.Marshal(acc)
		if err == nil {
			t.Log("Account: %s", string(out))
		}
	*/

}

func TestSplitSpans(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, TOR: utils.VOICE}

	cd.LoadRatingPlans()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestSplitSpansWeekend(t *testing.T) {
	cd := &CallDescriptor{Direction: utils.OUT,
		Category:        "postpaid",
		TOR:             utils.VOICE,
		Tenant:          "foehn",
		Subject:         "foehn",
		Account:         "foehn",
		Destination:     "0034678096720",
		TimeStart:       time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:         time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		LoopIndex:       0,
		DurationIndex:   176 * time.Second,
		FallbackSubject: "",
		RatingInfos: RatingInfos{
			&RatingInfo{
				MatchedSubject: "*out:foehn:postpaid:foehn",
				MatchedPrefix:  "0034678",
				MatchedDestId:  "SPN_MOB",
				ActivationTime: time.Date(2015, 4, 23, 0, 0, 0, 0, time.UTC),
				RateIntervals: []*RateInterval{
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
							StartTime: "08:00:00",
						},
						Rating: &RIRate{
							ConnectFee:       0,
							RoundingMethod:   "*up",
							RoundingDecimals: 6,
							Rates: RateGroups{
								&Rate{Value: 1, RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
							StartTime: "00:00:00",
						},
						Rating: &RIRate{
							ConnectFee:       0,
							RoundingMethod:   "*up",
							RoundingDecimals: 6,
							Rates: RateGroups{
								&Rate{Value: 1, RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Saturday, time.Sunday},
							StartTime: "00:00:00",
						},
						Rating: &RIRate{
							ConnectFee:       0,
							RoundingMethod:   "*up",
							RoundingDecimals: 6,
							Rates: RateGroups{
								&Rate{Value: 1, RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
				},
			},
		},
	}

	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
	if timespans[0].RateInterval == nil ||
		timespans[0].RateInterval.Timing.StartTime != "00:00:00" ||
		timespans[1].RateInterval == nil ||
		timespans[1].RateInterval.Timing.StartTime != "08:00:00" {
		t.Errorf("Error setting rateinterval: %+v %+v", timespans[0].RateInterval.Timing.StartTime, timespans[1].RateInterval.Timing.StartTime)
	}
}

func TestSplitSpansRoundToIncrements(t *testing.T) {
	t1 := time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC)
	t2 := time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, DurationIndex: 132 * time.Second}

	cd.LoadRatingPlans()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Logf("%+v", cd)
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
	var d time.Duration
	for _, ts := range timespans {
		d += ts.GetDuration()
	}
	if d != 132*time.Second {
		t.Error("Wrong duration for timespans: ", d)
	}
}

func TestCalldescHolliday(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart: time.Date(2015, time.May, 1, 13, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2015, time.May, 1, 13, 35, 26, 0, time.UTC),
		RatingInfos: RatingInfos{
			&RatingInfo{
				RateIntervals: RateIntervalList{
					&RateInterval{
						Timing: &RITiming{WeekDays: utils.WeekDays{1, 2, 3, 4, 5}, StartTime: "00:00:00"},
						Weight: 10,
					},
					&RateInterval{
						Timing: &RITiming{WeekDays: utils.WeekDays{6, 7}, StartTime: "00:00:00"},
						Weight: 10,
					},
					&RateInterval{
						Timing: &RITiming{Months: utils.Months{time.May}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
						Weight: 20,
					},
				},
			},
		},
	}
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 1 {
		t.Error("Error assiging holidy rate interval: ", timespans)
	}
	if timespans[0].RateInterval.Timing.MonthDays == nil {
		t.Errorf("Error setting holiday rate interval: %+v", timespans[0].RateInterval.Timing)
	}
}

func TestGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2701}
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetCostRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	if result.Cost != 0.3001 || result.GetConnectFee() != 0 { // should be 0.3 :(
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestDebitRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.Debit()
	if result.Cost != 0.30006 || result.GetConnectFee() != 0 { // should be 0.3 :(
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestDebitPerformRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "cgrates.org", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, PerformRounding: true}
	if result, err := cd.Debit(); err != nil {
		t.Error(err)
	} else if result.Cost != 0.3001 || result.GetConnectFee() != 0 { // should be 0.3 :(
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestGetCostZero(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 0}
	if result.Cost != expected.Cost || result.GetConnectFee() != 0 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetCostTimespans(t *testing.T) {
	t1 := time.Date(2013, time.October, 8, 9, 23, 2, 0, time.UTC)
	t2 := time.Date(2013, time.October, 8, 9, 24, 27, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: 85 * time.Second}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "trp", Destination: "0256", Cost: 85}
	if result.Cost != expected.Cost || result.GetConnectFee() != 0 || len(result.Timespans) != 2 {
		t.Errorf("Expected %+v was %+v", expected, result)
	}

}

func TestGetCostRatingPlansAndRatingIntervals(t *testing.T) {
	t1 := time.Date(2012, time.February, 27, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 28, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "CUSTOMER_1", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 3 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 3, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMore(t *testing.T) {
	t1 := time.Date(2012, time.February, 27, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 28, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "CUSTOMER_1", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 4 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 4, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMoreDays(t *testing.T) {
	t1 := time.Date(2012, time.February, 20, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 23, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "CUSTOMER_1", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 8 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) ||
		!result.Timespans[3].TimeEnd.Equal(result.Timespans[4].TimeStart) ||
		!result.Timespans[4].TimeEnd.Equal(result.Timespans[5].TimeStart) ||
		!result.Timespans[5].TimeEnd.Equal(result.Timespans[6].TimeStart) ||
		!result.Timespans[6].TimeEnd.Equal(result.Timespans[7].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 4, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMoreDaysWeekend(t *testing.T) {
	t1 := time.Date(2012, time.February, 24, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 27, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "CUSTOMER_1", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 5 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) ||
		!result.Timespans[3].TimeEnd.Equal(result.Timespans[4].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 4, len(result.Timespans))
	}
}

func TestGetCostRateGroups(t *testing.T) {
	t1 := time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC)
	t2 := time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, DurationIndex: 132 * time.Second}

	result, err := cd.GetCost()
	if err != nil {
		t.Error("Error getting cost: ", err)
	}
	if result.Cost != 132 {
		t.Error("Error calculating cost: ", result.Timespans)
	}
}

func TestGetCostNoConnectFee(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 1}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2700}
	// connect fee is not added because LoopIndex is 1
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetCostAccount(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Account: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2701}
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestFullDestNotFound(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2701}
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		t.Log(cd.RatingInfos)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSubjectNotFound(t *testing.T) {
	t1 := time.Date(2013, time.February, 1, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2013, time.February, 1, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "not_exiting", Destination: "025740532", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 2701}
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		//t.Logf("%+v", result.Timespans[0].RateInterval)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSubjectNotFoundCostNegativeOne(t *testing.T) {
	t1 := time.Date(2013, time.February, 1, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2013, time.February, 1, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "cgrates.org", Subject: "not_exiting", Destination: "025740532", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	if result.Cost != -1 || result.GetConnectFee() != 0 {
		//t.Logf("%+v", result.Timespans[0].RateInterval)
		t.Errorf("Expected -1 was %v", result)
	}
}

func TestMultipleRatingPlans(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 2701}
	if result.Cost != expected.Cost || result.GetConnectFee() != 1 {
		t.Log(result.Timespans)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSpansMultipleRatingPlans(t *testing.T) {
	t1 := time.Date(2012, time.February, 7, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 0, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	cc, _ := cd.GetCost()
	if cc.Cost != 2100 || cc.GetConnectFee() != 0 {
		utils.LogFull(cc)
		t.Errorf("Expected %v was %v (%v)", 2100, cc, cc.GetConnectFee())
	}
}

func TestLessThanAMinute(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 30, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 15}
	if result.Cost != expected.Cost || result.GetConnectFee() != 0 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestUniquePrice(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 21, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0723045326", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0723", Cost: 1810.5}
	if result.Cost != expected.Cost || result.GetConnectFee() != 0 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMinutesCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 22, 51, 50, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0723", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "minutosu", Destination: "0723", Cost: 55}
	if result.Cost != expected.Cost || result.GetConnectFee() != 0 {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "ttttttt",
		Destination: "0723"}
	result, err := cd.GetMaxSessionDuration()
	if result != 0 || err == nil {
		t.Errorf("Expected %v was %v (%v)", 0, result, err)
	}
}

func TestMaxSessionTimeWithAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu",
		Destination: "0723",
	}
	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeWithMaxRate(t *testing.T) {
	ap, err := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	if err != nil {
		t.FailNow()
	}
	//log.Print(ap)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	//acc, _ := dataStorage.GetAccount("cgrates.org:12345")
	//log.Print("ACC: ", utils.ToIJSON(acc))
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "12345",
		Account:     "12345",
		Destination: "447956",
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 0, 0, time.UTC),
		MaxRate:     1.0,
		MaxRateUnit: time.Minute,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 40 * time.Second
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeWithMaxCost(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 10 * time.Second
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetMaxSessiontWithBlocker(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("BLOCK_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	acc, err := dm.DataDB().GetAccount("cgrates.org:block")
	if err != nil {
		t.Error("error getting account: ", err)
	}
	if len(acc.BalanceMap[utils.MONETARY]) != 2 ||
		acc.BalanceMap[utils.MONETARY][0].Blocker != true {
		for _, b := range acc.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("Error executing action  plan on account: ", acc.BalanceMap[utils.MONETARY])
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "block",
		Account:      "block",
		Destination:  "0723",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 17 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
	cd = &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "block",
		Account:      "block",
		Destination:  "444",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err = cd.GetMaxSessionDuration()
	expected = 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
}

func TestGetMaxSessiontWithBlockerEmpty(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("BLOCK_EMPTY_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	acc, err := dm.DataDB().GetAccount("cgrates.org:block_empty")
	if err != nil {
		t.Error("error getting account: ", err)
	}
	if len(acc.BalanceMap[utils.MONETARY]) != 2 ||
		acc.BalanceMap[utils.MONETARY][0].Blocker != true {
		for _, b := range acc.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("Error executing action  plan on account: ", acc.BalanceMap[utils.MONETARY])
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "block",
		Account:      "block_empty",
		Destination:  "0723",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 0 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
	cd = &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "block",
		Account:      "block_empty",
		Destination:  "444",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err = cd.GetMaxSessionDuration()
	expected = 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
}

func TestGetCostWithMaxCost(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.GetCost()
	expected := 1800.0
	if cc.Cost != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, cc.Cost)
	}
}

func TestGetCostRoundingIssue(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 51, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.GetCost()
	expected := 0.17
	if cc.Cost != expected || err != nil {
		t.Log(utils.ToIJSON(cc))
		t.Errorf("Expected %v was %+v", expected, cc)
	}
}

func TestGetCostRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.GetCost()
	if err != nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestId != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:cgrates.org:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanId != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestDebitRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.Debit()
	if err != nil ||
		cc == nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestId != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:cgrates.org:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanId != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestMaxDebitRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.MaxDebit()
	if err != nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestId != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:cgrates.org:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanId != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestMaxDebitUnknowDest(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "9999999999",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 29, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.MaxDebit()
	if err == nil || err != utils.ErrUnauthorizedDestination {
		t.Errorf("Bad error reported %+v: %v", cc, err)
	}
}

func TestMaxDebitRoundingIssue(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:       "*out",
		Category:        "call",
		Tenant:          "cgrates.org",
		Subject:         "dy",
		Account:         "dy",
		Destination:     "0723123113",
		TimeStart:       time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:         time.Date(2015, 10, 26, 13, 29, 51, 0, time.UTC),
		MaxCostSoFar:    0,
		PerformRounding: true,
	}
	acc, err := dm.DataDB().GetAccount("cgrates.org:dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value != 1 {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}

	cc, err := cd.MaxDebit()
	expected := 0.17
	if cc.Cost != expected || err != nil {
		t.Log(utils.ToIJSON(cc))
		t.Errorf("Expected %v was %+v (%v)", expected, cc, err)
	}
	acc, err = dm.DataDB().GetAccount("cgrates.org:dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value != 1-expected {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}
}

func TestDebitRoundingRefund(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:       "*out",
		Category:        "call",
		Tenant:          "cgrates.org",
		Subject:         "dy",
		Account:         "dy",
		Destination:     "0723123113",
		TimeStart:       time.Date(2016, 3, 4, 13, 50, 00, 0, time.UTC),
		TimeEnd:         time.Date(2016, 3, 4, 13, 53, 00, 0, time.UTC),
		MaxCostSoFar:    0,
		PerformRounding: true,
	}
	acc, err := dm.DataDB().GetAccount("cgrates.org:dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value != 1 {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}

	cc, err := cd.Debit()
	expected := 0.3
	if cc.Cost != expected || err != nil {
		t.Log(utils.ToIJSON(cc))
		t.Errorf("Expected %v was %+v (%v)", expected, cc, err)
	}
	acc, err = dm.DataDB().GetAccount("cgrates.org:dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value != 1-expected {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}
}

func TestMaxSessionTimeWithMaxCostFree(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxDebitWithMaxCostFree(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cc, err := cd.MaxDebit()
	expected := 10.0
	if cc.Cost != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, cc.Cost)
	}
}

func TestGetCostWithMaxCostFree(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}

	cc, err := cd.GetCost()
	expected := 10.0
	if cc.Cost != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, cc.Cost)
	}
}

func TestMaxSessionTimeWithAccountAlias(t *testing.T) {
	aliasService = NewAliasHandler(dm)
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "a1",
		Account:     "a1",
		Destination: "0723",
	}
	LoadAlias(
		&AttrMatchingAlias{
			Destination: cd.Destination,
			Direction:   cd.Direction,
			Tenant:      cd.Tenant,
			Category:    cd.Category,
			Account:     cd.Account,
			Subject:     cd.Subject,
			Context:     utils.MetaRating,
		}, cd, utils.EXTRA_FIELDS)

	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v, %v", expected, result, err)
	}
}

func TestMaxSessionTimeWithAccountShared(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP_SHARED0_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	ap, _ = dm.DataDB().GetActionPlan("TOPUP_SHARED10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}

	cd0 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "rif",
		Account:     "empty0",
		Destination: "0723",
	}

	cd1 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "rif",
		Account:     "empty10",
		Destination: "0723",
	}

	result0, err := cd0.GetMaxSessionDuration()
	result1, err := cd1.GetMaxSessionDuration()
	if result0 != result1/2 || err != nil {
		t.Errorf("Expected %v was %v, %v", result1/2, result0, err)
	}
}

func TestMaxDebitWithAccountShared(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP_SHARED0_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	ap, _ = dm.DataDB().GetActionPlan("TOPUP_SHARED10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu",
		Account:     "empty0",
		Destination: "0723",
	}

	cc, err := cd.MaxDebit()
	if err != nil || cc.Cost != 2.5 {
		t.Errorf("Wrong callcost in shared debit: %+v, %v", cc, err)
	}
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 1 || balanceMap[0].GetValue() != 0 {
		t.Errorf("Wrong shared balance debited: %+v", balanceMap[0])
	}
	other, err := dm.DataDB().GetAccount("vdf:empty10")
	if err != nil || other.BalanceMap[utils.MONETARY][0].GetValue() != 7.5 {
		t.Errorf("Error debiting shared balance: %+v", other.BalanceMap[utils.MONETARY][0])
	}
}

func TestMaxSessionTimeWithAccountAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoCredit(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "broker",
		Destination: "0723",
		TOR:         utils.VOICE,
	}
	result, err := cd.GetMaxSessionDuration()
	if result != time.Minute || err != nil {
		t.Errorf("Expected %v was %v", time.Minute, result)
	}
}

func TestMaxSessionModifiesCallDesc(t *testing.T) {
	t1 := time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC)
	t2 := time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC)
	cd := &CallDescriptor{
		TimeStart:     t1,
		TimeEnd:       t2,
		Direction:     "*out",
		Category:      "0",
		Tenant:        "vdf",
		Subject:       "minu_from_tm",
		Account:       "minu",
		Destination:   "0723",
		DurationIndex: t2.Sub(t1),
		TOR:           utils.VOICE,
	}
	initial := cd.Clone()
	_, err := cd.GetMaxSessionDuration()
	if err != nil {
		t.Error("Got error from max duration: ", err)
	}
	cd.account = nil // it's OK to cache the account
	if !reflect.DeepEqual(cd, initial) {
		t.Errorf("GetMaxSessionDuration is changing the call descriptor %+v != %+v", cd, initial)
	}
}

func TestMaxDebitDurationNoGreatherThanInitialDuration(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	initialDuration := cd.TimeEnd.Sub(cd.TimeStart)
	result, err := cd.GetMaxSessionDuration()
	if err != nil {
		t.Error("Got error from max duration: ", err)
	}
	if result > initialDuration {
		t.Error("max session duration greather than initial duration", initialDuration, result)
	}
}

func TestDebitAndMaxDebit(t *testing.T) {
	cd1 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 10, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	cd2 := cd1.Clone()
	cc1, err1 := cd1.Debit()
	cc2, err2 := cd2.MaxDebit()
	if err1 != nil || err2 != nil {
		t.Error("Error debiting and/or maxdebiting: ", err1, err2)
	}
	if cc1.Timespans[0].Increments[0].BalanceInfo.Unit.Value != 90*float64(time.Second) ||
		cc2.Timespans[0].Increments[0].BalanceInfo.Unit.Value != 80*float64(time.Second) {
		t.Error("Error setting the Unit.Value: ",
			cc1.Timespans[0].Increments[0].BalanceInfo.Unit.Value,
			cc2.Timespans[0].Increments[0].BalanceInfo.Unit.Value)
	}
	// make Unit.Values have the same value
	cc1.Timespans[0].Increments[0].BalanceInfo.Unit.Value = 0
	cc2.Timespans[0].Increments[0].BalanceInfo.Unit.Value = 0
	cc1.AccountSummary, cc2.AccountSummary = nil, nil // account changes, enforce here emptying the info so reflect can pass
	if !reflect.DeepEqual(cc1.Timespans, cc2.Timespans) {
		t.Log("CC1: ", utils.ToIJSON(cc1))
		t.Log("CC2: ", utils.ToIJSON(cc2))
		t.Error("Debit and MaxDebit differ")
	}
}

func TestMaxSesionTimeEmptyBalance(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "minu_from_tm",
		Account:     "luna",
		Destination: "0723",
	}
	acc, _ := dm.DataDB().GetAccount("vdf:luna")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime != 0 {
		t.Error("Error get max session for 0 acount", err)
	}
}

func TestMaxSesionTimeEmptyBalanceAndNoCost(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "one",
		Account:     "luna",
		Destination: "112",
	}
	acc, _ := dm.DataDB().GetAccount("vdf:luna")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime == 0 {
		t.Error("Error get max session for 0 acount", err)
	}
}

func TestMaxSesionTimeLong(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 07, 24, 13, 37, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 07, 24, 15, 37, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "money",
		Destination: "0723",
	}
	acc, _ := dm.DataDB().GetAccount("cgrates.org:money")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime != cd.TimeEnd.Sub(cd.TimeStart) {
		t.Error("Error get max session for acount:", allowedTime, err)
	}
}

func TestMaxSesionTimeLongerThanMoney(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 07, 24, 13, 37, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 07, 24, 16, 37, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "money",
		Destination: "0723",
	}
	acc, _ := dm.DataDB().GetAccount("cgrates.org:money")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	expected, err := time.ParseDuration("9999s") // 1 is the connect fee
	if err != nil || allowedTime != expected {
		t.Log(utils.ToIJSON(acc))
		t.Errorf("Expected: %v got %v", expected, allowedTime)
	}
}

func TestDebitFromShareAndNormal(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP_SHARED10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "rif",
		Account:     "empty10",
		Destination: "0723",
	}
	cc, err := cd.MaxDebit()
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if err != nil || cc.Cost != 2.5 {
		t.Errorf("Debit from share and normal error: %+v, %v", cc, err)
	}

	if balanceMap[0].GetValue() != 10 || balanceMap[1].GetValue() != 27.5 {
		t.Errorf("Error debiting from right balance: %v %v", balanceMap[0].GetValue(), balanceMap[1].GetValue())
	}
}

func TestDebitFromEmptyShare(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP_EMPTY_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "rif",
		Account:     "emptyX",
		Destination: "0723",
	}

	cc, err := cd.MaxDebit()
	if err != nil || cc.Cost != 2.5 {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 2 || balanceMap[0].GetValue() != 0 || balanceMap[1].GetValue() != -2.5 {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[1].GetValue())
	}
}

func TestDebitNegatve(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("POST_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "vdf",
		Subject:     "rif",
		Account:     "post",
		Destination: "0723",
	}
	cc, err := cd.MaxDebit()
	//utils.PrintFull(cc)
	if err != nil || cc.Cost != 2.5 {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	acc, _ := cd.getAccount()
	//utils.PrintFull(acc)
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 1 || balanceMap[0].GetValue() != -2.5 {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[0].GetValue())
	}
	cc, err = cd.MaxDebit()
	acc, _ = cd.getAccount()
	balanceMap = acc.BalanceMap[utils.MONETARY]
	//utils.LogFull(balanceMap)
	if err != nil || cc.Cost != 2.5 {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	if len(balanceMap) != 1 || balanceMap[0].GetValue() != -5 {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[0].GetValue())
	}
}

func TestMaxDebitZeroDefinedRate(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 0, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cc, err := cd1.MaxDebit()
	if err != nil {
		t.Error("Error maxdebiting: ", err)
	}
	if cc.GetDuration() != 49*time.Second {
		t.Error("Error obtaining max debit duration: ", cc.GetDuration())
	}

	if cc.Cost != 0.91 {
		t.Error("Error in max debit cost: ", cc.Cost)
	}
}

func TestMaxDebitForceDuration(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 40, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0,
		ForceDuration: true,
	}
	_, err := cd1.MaxDebit()
	if err != utils.ErrInsufficientCredit {
		t.Fatal("Error forcing duration: ", err)
	}
}

func TestMaxDebitZeroDefinedRateOnlyMinutes(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 0, 40, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cc, err := cd1.MaxDebit()
	if err != nil {
		t.Fatal("Error maxdebiting: ", err)
	}
	if cc.GetDuration() != 40*time.Second {
		t.Error("Error obtaining max debit duration: ", cc.GetDuration())
	}
	if cc.Cost != 0.01 {
		t.Error("Error in max debit cost: ", cc.Cost)
	}
}

func TestMaxDebitConsumesMinutes(t *testing.T) {
	ap, _ := dm.DataDB().GetActionPlan("TOPUP10_AT", false, utils.NonTransactional)
	for _, at := range ap.ActionTimings {
		at.accountIDs = ap.AccountIDs
		at.Execute(nil, nil)
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 0, 5, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cd1.MaxDebit()
	if cd1.account.BalanceMap[utils.VOICE][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error using minutes: ",
			cd1.account.BalanceMap[utils.VOICE][0].GetValue())
	}
}

func TestCDGetCostANY(t *testing.T) {
	cd1 := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 0, 1, 0, time.UTC),
		TOR:         utils.DATA,
	}
	cc, err := cd1.GetCost()
	if err != nil || cc.Cost != 60 {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
}

func TestCDSplitInDataSlots(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "data",
		Tenant:        "cgrates.org",
		Subject:       "rif",
		Destination:   utils.ANY,
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:           utils.DATA,
		DurationIndex: 65 * time.Second,
	}
	cd.LoadRatingPlans()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos[0])
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestCDDataGetCost(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:         utils.DATA,
	}
	cc, err := cd.GetCost()
	if err != nil || cc.Cost != 65 {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
}

func TestCDRefundIncrements(t *testing.T) {
	ub := &Account{
		ID: "test:ref",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Uuid: "moneya", Value: 100},
			},
			utils.VOICE: Balances{
				&Balance{Uuid: "minutea",
					Value:          10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Uuid: "minuteb",
					Value:          10 * float64(time.Second),
					DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	dm.DataDB().SetAccount(ub)
	increments := Increments{
		&Increment{Cost: 2, BalanceInfo: &DebitInfo{
			Monetary: &MonetaryInfo{UUID: "moneya"}, AccountID: ub.ID}},
		&Increment{Cost: 2, Duration: 3 * time.Second, BalanceInfo: &DebitInfo{
			Unit:     &UnitInfo{UUID: "minutea"},
			Monetary: &MonetaryInfo{UUID: "moneya"}, AccountID: ub.ID}},
		&Increment{Duration: 4 * time.Second, BalanceInfo: &DebitInfo{
			Unit: &UnitInfo{UUID: "minuteb"}, AccountID: ub.ID}},
	}
	cd := &CallDescriptor{TOR: utils.VOICE, Increments: increments}
	cd.RefundIncrements()
	ub, _ = dm.DataDB().GetAccount(ub.ID)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 104 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 13*float64(time.Second) ||
		ub.BalanceMap[utils.VOICE][1].GetValue() != 14*float64(time.Second) {
		t.Error("Error refunding money: ", utils.ToIJSON(ub.BalanceMap))
	}
}

func TestCDRefundIncrementsZeroValue(t *testing.T) {
	ub := &Account{
		ID: "test:ref",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Uuid: "moneya", Value: 100},
			},
			utils.VOICE: Balances{
				&Balance{Uuid: "minutea", Value: 10, Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Uuid: "minuteb", Value: 10, DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	dm.DataDB().SetAccount(ub)
	increments := Increments{
		&Increment{Cost: 0, BalanceInfo: &DebitInfo{AccountID: ub.ID}},
		&Increment{Cost: 0, Duration: 3 * time.Second, BalanceInfo: &DebitInfo{AccountID: ub.ID}},
		&Increment{Cost: 0, Duration: 4 * time.Second, BalanceInfo: &DebitInfo{AccountID: ub.ID}},
	}
	cd := &CallDescriptor{TOR: utils.VOICE, Increments: increments}
	cd.RefundIncrements()
	ub, _ = dm.DataDB().GetAccount(ub.ID)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 100 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		ub.BalanceMap[utils.VOICE][1].GetValue() != 10 {
		t.Error("Error refunding money: ", utils.ToIJSON(ub.BalanceMap))
	}
}

func TestCDDebitBalanceSubjectWithFallback(t *testing.T) {
	acnt := &Account{
		ID: "TCDDBSWF:account1",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{
				&Balance{ID: "voice1", Value: 60 * float64(time.Second),
					RatingSubject: "SubjTCDDBSWF"},
			}},
	}
	dm.DataDB().SetAccount(acnt)
	dst := &Destination{Id: "DST_TCDDBSWF", Prefixes: []string{"1716"}}
	dm.DataDB().SetDestination(dst, utils.NonTransactional)
	dm.DataDB().SetReverseDestination(dst, utils.NonTransactional)
	rpSubj := &RatingPlan{
		Id: "RP_TCDDBSWF",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{

				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      60 * time.Second,
						RateUnit:           60 * time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			dst.Id: []*RPRate{
				&RPRate{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	rpDflt := &RatingPlan{
		Id: "RP_DFLT",
		Timings: map[string]*RITiming{
			"30eab301": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f861": &RIRate{
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			dst.Id: []*RPRate{
				&RPRate{
					Timing: "30eab301",
					Rating: "b457f861",
					Weight: 10,
				},
			},
		},
	}
	for _, rpl := range []*RatingPlan{rpSubj, rpDflt} {
		dm.SetRatingPlan(rpl, utils.NonTransactional)
	}
	rpfTCDDBSWF := &RatingProfile{Id: "*out:TCDDBSWF:call:SubjTCDDBSWF",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rpSubj.Id,
		}},
	}
	rpfAny := &RatingProfile{Id: "*out:TCDDBSWF:call:*any",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rpDflt.Id,
		}},
	}
	for _, rpf := range []*RatingProfile{rpfTCDDBSWF, rpfAny} {
		dm.SetRatingProfile(rpf, utils.NonTransactional)
	}
	cd1 := &CallDescriptor{ // test the cost for subject within balance setup
		Direction:   "*out",
		Category:    "call",
		Tenant:      "TCDDBSWF",
		Subject:     "SubjTCDDBSWF",
		Destination: "1716",
		TimeStart:   time.Date(2015, 01, 01, 9, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 01, 01, 9, 2, 0, 0, time.UTC),
		TOR:         utils.VOICE,
	}
	if cc, err := cd1.GetCost(); err != nil || cc.Cost != 0 {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
	cd2 := &CallDescriptor{ // equivalent of the extra charge out of *monetary balance when *voice is finished
		Direction:   "*out",
		Category:    "call",
		Tenant:      "TCDDBSWF",
		Subject:     "dan",
		Destination: "1716",
		TimeStart:   time.Date(2015, 01, 01, 9, 1, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 01, 01, 9, 2, 0, 0, time.UTC),
		TOR:         utils.VOICE,
	}
	if cc, err := cd2.GetCost(); err != nil || cc.Cost != 0.6 {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
	cd := &CallDescriptor{ // test the cost
		Direction:   "*out",
		Category:    "call",
		Tenant:      "TCDDBSWF",
		Account:     "account1",
		Subject:     "account1",
		Destination: "1716",
		TimeStart:   time.Date(2015, 01, 01, 9, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 01, 01, 9, 2, 0, 0, time.UTC),
		TOR:         utils.VOICE,
	}
	if cc, err := cd.Debit(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.6 {
		t.Errorf("CallCost: %v", cc)
	}
	if resAcnt, err := dm.DataDB().GetAccount(acnt.ID); err != nil {
		t.Error(err)
	} else if resAcnt.BalanceMap[utils.VOICE][0].ID != "voice1" ||
		resAcnt.BalanceMap[utils.VOICE][0].Value != 0 {
		t.Errorf("Account: %v", resAcnt)
	} else if len(resAcnt.BalanceMap[utils.MONETARY]) == 0 ||
		resAcnt.BalanceMap[utils.MONETARY][0].ID != utils.META_DEFAULT ||
		resAcnt.BalanceMap[utils.MONETARY][0].Value != -0.600013 { // rounding issue
		t.Errorf("Account: %s", utils.ToIJSON(resAcnt))
	}
}
func TestCallDescriptorUpdateFromCGREvent(t *testing.T) {
	context := "*rating"
	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Generated",
		Context: &context,
		Event: map[string]interface{}{
			"Account":     "acc1",
			"AnswerTime":  time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
			"Category":    "call",
			"Destination": "0723123113",
			"Subject":     "acc1",
			"Tenant":      "cgrates.org",
			"ToR":         "",
			"Usage":       time.Duration(30) * time.Minute,
		},
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	cdExpected := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "acc1",
		Account:      "acc1",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	if err := cd.UpdateFromCGREvent(cgrEv, []string{utils.Account, utils.Subject}); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(cd, cdExpected) {
			t.Errorf("Expecting: %+v, received: %+v", cdExpected, cd)
		}
	}
	cgrEv = &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Generated",
		Context: &context,
		Event: map[string]interface{}{
			"Account":     "acc1",
			"AnswerTime":  time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
			"Category":    "call",
			"Destination": "0723123113",
			"Subject":     "acc1",
			"Tenant":      "cgrates.org",
			"ToR":         "",
			"Usage":       time.Duration(40) * time.Minute,
		},
	}
	if err := cd.UpdateFromCGREvent(cgrEv, []string{utils.Account, utils.Subject}); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(cd, cdExpected) {
			t.Errorf("Expecting: %+v, received: %+v", cdExpected, cd)
		}
	}

}

func TestCallDescriptorAsCGREvent(t *testing.T) {
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "cgrates.org",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: 0,
	}
	context := "*rating"
	eCGREvent := &utils.CGREvent{Tenant: "cgrates.org",
		ID:      "Generated",
		Context: &context,
		Event: map[string]interface{}{
			"Account":     "max",
			"AnswerTime":  time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
			"Category":    "call",
			"Destination": "0723123113",
			"Subject":     "max",
			"Tenant":      "cgrates.org",
			"ToR":         "",
			"Usage":       time.Duration(30) * time.Minute,
		},
	}
	cgrEvent := cd.AsCGREvent()
	if !reflect.DeepEqual(eCGREvent.Tenant, cgrEvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", eCGREvent.Tenant, cgrEvent.Tenant)
	}
	for fldName, fldVal := range eCGREvent.Event {
		if _, has := cgrEvent.Event[fldName]; !has {
			t.Errorf("Expecting: %+v, received: %+v", fldName, nil)
		} else if fldVal != cgrEvent.Event[fldName] {
			t.Errorf("Expecting: %s:%+v, received: %s:%+v", fldName, eCGREvent.Event[fldName], fldName, cgrEvent.Event[fldName])
		}
	}
}

/*************** BENCHMARKS ********************/
func BenchmarkStorageGetting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		dm.GetRatingProfile(cd.GetKey(cd.Subject), false, utils.NonTransactional)
	}
}

func BenchmarkStorageRestoring(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.LoadRatingPlans()
	}
}

func BenchmarkStorageGetCost(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cd.LoadRatingPlans()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.splitInTimeSpans()
	}
}

func BenchmarkStorageSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Tenant: "vdf", Subject: "minutosu", Destination: "0723"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionDuration()
	}
}

func BenchmarkStorageMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "minutosu", Destination: "0723"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionDuration()
	}
}
