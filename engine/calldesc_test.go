/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/utils"
)

var (
	marsh = NewCodecMsgpackMarshaler()
)

func init() {
	historyScribe, _ = history.NewMockScribe()
	ratingStorage.Flush("")
	accountingStorage.Flush("")
	populateDB()
}

func populateDB() {
	ats := []*Action{
		&Action{ActionType: "*topup", BalanceType: utils.MONETARY, Balance: &Balance{Value: 10}},
		&Action{ActionType: "*topup", BalanceType: utils.VOICE, Balance: &Balance{Weight: 20, Value: 10, DestinationIds: utils.NewStringMap("NAT")}},
	}

	ats1 := []*Action{
		&Action{ActionType: "*topup", BalanceType: utils.MONETARY, Balance: &Balance{Value: 10}, Weight: 10},
		&Action{ActionType: "*reset_account", Weight: 20},
	}

	minu := &Account{
		Id: "vdf:minu",
		BalanceMap: map[string]BalanceChain{
			utils.MONETARY: BalanceChain{&Balance{Value: 50}},
			utils.VOICE: BalanceChain{
				&Balance{Value: 200, DestinationIds: utils.NewStringMap("NAT"), Weight: 10},
				&Balance{Value: 100, DestinationIds: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	broker := &Account{
		Id: "vdf:broker",
		BalanceMap: map[string]BalanceChain{
			utils.VOICE: BalanceChain{
				&Balance{Value: 20, DestinationIds: utils.NewStringMap("NAT"), Weight: 10, RatingSubject: "rif"},
				&Balance{Value: 100, DestinationIds: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	luna := &Account{
		Id: "vdf:luna",
		BalanceMap: map[string]BalanceChain{
			utils.MONETARY: BalanceChain{
				&Balance{Value: 0, Weight: 20},
			}},
	}
	// this is added to test if csv load tests account will not overwrite balances
	minitsboy := &Account{
		Id: "vdf:minitsboy",
		BalanceMap: map[string]BalanceChain{
			utils.VOICE: BalanceChain{
				&Balance{Value: 20, DestinationIds: utils.NewStringMap("NAT"), Weight: 10, RatingSubject: "rif"},
				&Balance{Value: 100, DestinationIds: utils.NewStringMap("RET"), Weight: 20},
			},
			utils.MONETARY: BalanceChain{
				&Balance{Value: 100, Weight: 10},
			},
		},
	}
	max := &Account{
		Id: "cgrates.org:max",
		BalanceMap: map[string]BalanceChain{
			utils.MONETARY: BalanceChain{
				&Balance{Value: 11, Weight: 20},
			}},
	}
	money := &Account{
		Id: "cgrates.org:money",
		BalanceMap: map[string]BalanceChain{
			utils.MONETARY: BalanceChain{
				&Balance{Value: 10000, Weight: 10},
			}},
	}
	if accountingStorage != nil && ratingStorage != nil {
		ratingStorage.SetActions("TEST_ACTIONS", ats)
		ratingStorage.SetActions("TEST_ACTIONS_ORDER", ats1)
		accountingStorage.SetAccount(broker)
		accountingStorage.SetAccount(minu)
		accountingStorage.SetAccount(minitsboy)
		accountingStorage.SetAccount(luna)
		accountingStorage.SetAccount(max)
		accountingStorage.SetAccount(money)
	} else {
		log.Fatal("Could not connect to db!")
	}
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
	}
	//acc, _ := accountingStorage.GetAccount("cgrates.org:12345")
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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

func TestGetCostWithMaxCost(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
		t.Errorf("Expected %v was %+v", expected, cc)
	}
}

func TestMaxSessionTimeWithMaxCostFree(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	aliasService = NewAliasHandler(accountingStorage)
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
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cd, utils.EXTRA_FIELDS)

	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v, %v", expected, result, err)
	}
}

func TestMaxSessionTimeWithAccountShared(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlans("TOPUP_SHARED0_AT")
	for _, at := range ap {
		at.Execute()
	}
	ap, _ = ratingStorage.GetActionPlans("TOPUP_SHARED10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP_SHARED0_AT")
	for _, at := range ap {
		at.Execute()
	}
	ap, _ = ratingStorage.GetActionPlans("TOPUP_SHARED10_AT")
	for _, at := range ap {
		at.Execute()
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
	other, err := accountingStorage.GetAccount("vdf:empty10")
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
	if !reflect.DeepEqual(cc1, cc2) {
		t.Log("===============================")
		t.Logf("CC1: %+v", cc1)
		for _, ts := range cc1.Timespans {
			t.Logf("TS: %+v", ts)
		}
		t.Logf("CC2: %+v", cc2)
		for _, ts := range cc2.Timespans {
			t.Logf("TS: %+v", ts)
		}
		t.Log("===============================")
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
	acc, _ := accountingStorage.GetAccount("vdf:luna")
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
	acc, _ := accountingStorage.GetAccount("vdf:luna")
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
	acc, _ := accountingStorage.GetAccount("cgrates.org:money")
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
	acc, _ := accountingStorage.GetAccount("cgrates.org:money")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	expected, err := time.ParseDuration("2h46m40s")
	if err != nil || allowedTime != expected {
		t.Errorf("Expected: %v got %v", expected, allowedTime)
	}
}

func TestDebitFromShareAndNormal(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlans("TOPUP_SHARED10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP_EMPTY_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("POST_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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

func TestMaxDebitZeroDefinedRateOnlyMinutes(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	ap, _ := ratingStorage.GetActionPlans("TOPUP10_AT")
	for _, at := range ap {
		at.Execute()
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
	if cd1.account.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error using minutes: ", cd1.account.BalanceMap[utils.VOICE][0].GetValue())
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

/*************** BENCHMARKS ********************/
func BenchmarkStorageGetting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ratingStorage.GetRatingProfile(cd.GetKey(cd.Subject), false)
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
