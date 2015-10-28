/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"testing"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
)

var (
	NAT = &Destination{Id: "NAT", Prefixes: []string{"0257", "0256", "0723"}}
	RET = &Destination{Id: "RET", Prefixes: []string{"0723", "0724"}}
)

func TestBalanceStoreRestore(t *testing.T) {
	b := &Balance{Value: 14, Weight: 1, Uuid: "test", ExpirationDate: time.Date(2013, time.July, 15, 17, 48, 0, 0, time.UTC)}
	marsh := NewCodecMsgpackMarshaler()
	output, err := marsh.Marshal(b)
	if err != nil {
		t.Error("Error storing balance: ", err)
	}
	b1 := &Balance{}
	err = marsh.Unmarshal(output, b1)
	if err != nil {
		t.Error("Error restoring balance: ", err)
	}
	t.Logf("INITIAL: %+v", b)
	if !b.Equal(b1) {
		t.Errorf("Balance store/restore failed: expected %+v was %+v", b, b1)
	}
}

func TestBalanceStoreRestoreZero(t *testing.T) {
	b := &Balance{}

	output, err := marsh.Marshal(b)
	if err != nil {
		t.Error("Error storing balance: ", err)
	}
	b1 := &Balance{}
	err = marsh.Unmarshal(output, b1)
	if err != nil {
		t.Error("Error restoring balance: ", err)
	}
	if !b.Equal(b1) {
		t.Errorf("Balance store/restore failed: expected %v was %v", b, b1)
	}
}

func TestBalanceChainStoreRestore(t *testing.T) {
	bc := BalanceChain{&Balance{Value: 14, ExpirationDate: time.Date(2013, time.July, 15, 17, 48, 0, 0, time.UTC)}, &Balance{Value: 1024}}
	output, err := marsh.Marshal(bc)
	if err != nil {
		t.Error("Error storing balance chain: ", err)
	}
	bc1 := BalanceChain{}
	err = marsh.Unmarshal(output, &bc1)
	if err != nil {
		t.Error("Error restoring balance chain: ", err)
	}
	if !bc.Equal(bc1) {
		t.Errorf("Balance chain store/restore failed: expected %v was %v", bc, bc1)
	}
}

func TestAccountStorageStoreRestore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 21}}}}
	accountingStorage.SetAccount(rifsBalance)
	ub1, err := accountingStorage.GetAccount("other")
	if err != nil || !ub1.BalanceMap[utils.MONETARY].Equal(rifsBalance.BalanceMap[utils.MONETARY]) {
		t.Log("UB: ", ub1)
		t.Errorf("Expected %v was %v", rifsBalance.BalanceMap[utils.MONETARY], ub1.BalanceMap[utils.MONETARY])
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}
	ub1 := &Account{Id: "CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 200}}}}
	cd := &CallDescriptor{
		Category:      "0",
		Tenant:        "vdf",
		TimeStart:     time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 10 * time.Second,
		Direction:     utils.OUT,
		Destination:   "0723",
		TOR:           utils.VOICE,
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 110 * time.Second
	if credit != 200 || seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetSpecialPricedSeconds(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "minu"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}, RatingSubject: "minu"}

	ub1 := &Account{
		Id: "OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]BalanceChain{
			utils.VOICE:    BalanceChain{b1, b2},
			utils.MONETARY: BalanceChain{&Balance{Value: 21}},
		},
	}
	cd := &CallDescriptor{
		Category:    "0",
		Tenant:      "vdf",
		TimeStart:   time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 4, 15, 46, 60, 0, time.UTC),
		LoopIndex:   0,
		Direction:   utils.OUT,
		Destination: "0723",
		TOR:         utils.VOICE,
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 20 * time.Second
	if credit != 0 || seconds != expected || len(bucketList) != 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestAccountStorageStore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 21}}}}
	accountingStorage.SetAccount(rifsBalance)
	result, err := accountingStorage.GetAccount(rifsBalance.Id)
	if err != nil || rifsBalance.Id != result.Id ||
		len(rifsBalance.BalanceMap[utils.VOICE]) < 2 || len(result.BalanceMap[utils.VOICE]) < 2 ||
		!(rifsBalance.BalanceMap[utils.VOICE][0].Equal(result.BalanceMap[utils.VOICE][0])) ||
		!(rifsBalance.BalanceMap[utils.VOICE][1].Equal(result.BalanceMap[utils.VOICE][1])) ||
		!rifsBalance.BalanceMap[utils.MONETARY].Equal(result.BalanceMap[utils.MONETARY]) {
		t.Errorf("Expected %v was %v", rifsBalance, result)
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Destination:  "0723045326",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1}, utils.MONETARY: BalanceChain{&Balance{Categories: utils.NewStringMap("0"), Value: 21}}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" {
		t.Logf("%+v", cc.Timespans[0])
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Destination:  "0723045326",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE:    BalanceChain{b1},
		utils.MONETARY: BalanceChain{&Balance{Value: 21}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	t.Logf("%+v", cc.Timespans)
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditZeroMixedMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testm", Value: 70, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	b2 := &Balance{Uuid: "tests", Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.Timespans[0].GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE:    BalanceChain{b1, b2},
		utils.MONETARY: BalanceChain{&Balance{Value: 21}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "tests" ||
		cc.Timespans[1].Increments[0].BalanceInfo.UnitBalanceUuid != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans)
	}
	if rifsBalance.BalanceMap[utils.VOICE][1].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Logf("TS0: %+v", cc.Timespans[0])
		t.Logf("TS1: %+v", cc.Timespans[1])
		t.Errorf("Error extracting minutes from balance: %+v", rifsBalance.BalanceMap[utils.VOICE][1])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE: BalanceChain{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Showing no enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditHasCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE:    BalanceChain{b1},
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 110}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 3 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSplitMinutesMoney(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				DurationIndex: 0,
				ratingInfo:    &RatingInfo{},
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE:    BalanceChain{b1},
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 50}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != 1*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0].Duration)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 10*time.Second || cc.Timespans[1].GetDuration() != 20*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans[1].GetDuration())
	}
}

func TestDebitCreditMoreTimespans(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 150, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE: BalanceChain{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 30 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditMoreTimespansMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	b2 := &Balance{Uuid: "testa", Value: 150, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE: BalanceChain{b1, b2},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		rifsBalance.BalanceMap[utils.VOICE][1].GetValue() != 130 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][1], cc.Timespans[1])
	}
}

func TestDebitCreditNoConectFeeCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{ConnectFee: 10.0, Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE: BalanceChain{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}

	if len(cc.Timespans) != 1 || rifsBalance.BalanceMap[utils.MONETARY].GetTotalValue() != 0 {
		t.Error("Error cutting at no connect fee: ", rifsBalance.BalanceMap[utils.MONETARY])
	}
}

func TestDebitCreditMoneyOnly(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				ratingInfo:    &RatingInfo{},
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "money", Value: 50}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Missing noy enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "money" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Logf("%+v", cc.Timespans[0].Increments)
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0].BalanceInfo)
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 0 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MONETARY][0])
	}
	if len(cc.Timespans) != 2 ||
		cc.Timespans[0].GetDuration() != 10*time.Second ||
		cc.Timespans[1].GetDuration() != 40*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMinutes(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Categories: utils.NewStringMap("0"), Value: 250, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      "0",
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE:    BalanceChain{b1},
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 350}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Errorf("Error setting balance id to increment: %+v", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 180 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 280 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts)
		}
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMoney(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 75, DestinationIds: utils.StringMap{"NAT": true}, RatingSubject: "minu"}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 5 {
		t.Errorf("Error extracting minutes from balance: %+v",
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

/*func TestDebitCreditSubjectMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 40, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 55, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost: cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.VOICE: BalanceChain{b1},
		utils.MONETARY:  BalanceChain{&Balance{Uuid: "moneya", Value: 19500, RatingSubject: "minu"}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 7 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 40*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts)
		}
		t.Error("Error truncating extra timespans: ", len(cc.Timespans), cc.Timespans[0].GetDuration())
	}
}*/

func TestAccountdebitBalance(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{utils.SMS: BalanceChain{&Balance{Value: 14}}, utils.DATA: BalanceChain{&Balance{Value: 1204}}, utils.VOICE: BalanceChain{&Balance{Weight: 20, DestinationIds: utils.StringMap{"NAT": true}}, &Balance{Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
	}
	newMb := &Balance{Weight: 20, DestinationIds: utils.StringMap{"NEW": true}, Directions: utils.NewStringMap(utils.OUT)}
	a := &Action{BalanceType: utils.VOICE, Balance: newMb}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 3 || !ub.BalanceMap[utils.VOICE][2].DestinationIds.Equal(newMb.DestinationIds) {
		t.Errorf("Error adding minute bucket! %d %+v %+v", len(ub.BalanceMap[utils.VOICE]), ub.BalanceMap[utils.VOICE][2], newMb)
	}
}

func TestAccountDisableBalance(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{utils.SMS: BalanceChain{&Balance{Value: 14}}, utils.DATA: BalanceChain{&Balance{Value: 1204}}, utils.VOICE: BalanceChain{&Balance{Weight: 20, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
	}
	newMb := &Balance{Weight: 20, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT), Disabled: true}
	a := &Action{BalanceType: utils.VOICE, Balance: newMb}
	ub.enableDisableBalanceAction(a)
	if len(ub.BalanceMap[utils.VOICE]) != 2 || ub.BalanceMap[utils.VOICE][0].Disabled != true {
		for _, b := range ub.BalanceMap[utils.VOICE] {
			t.Logf("Balance: %+v", b)
		}
		t.Errorf("Error disabling balance! %d %+v %+v", len(ub.BalanceMap[utils.VOICE]), ub.BalanceMap[utils.VOICE][0], newMb)
	}
}

func TestAccountdebitBalanceExists(t *testing.T) {

	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{utils.SMS: BalanceChain{&Balance{Value: 14}}, utils.DATA: BalanceChain{&Balance{Value: 1024}}, utils.VOICE: BalanceChain{&Balance{Value: 15, Weight: 20, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
	}
	newMb := &Balance{Value: -10, Weight: 20, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT)}
	a := &Action{BalanceType: utils.VOICE, Balance: newMb}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 || ub.BalanceMap[utils.VOICE][0].GetValue() != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinuteNil(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{utils.SMS: BalanceChain{&Balance{Value: 14}}, utils.DATA: BalanceChain{&Balance{Value: 1024}}, utils.VOICE: BalanceChain{&Balance{Weight: 20, DestinationIds: utils.StringMap{"NAT": true}}, &Balance{Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
	}
	ub.debitBalanceAction(nil, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinutBucketEmpty(t *testing.T) {
	mb1 := &Balance{Value: -10, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}
	mb2 := &Balance{Value: -10, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}
	mb3 := &Balance{Value: -10, DestinationIds: utils.StringMap{"OTHER": true}, Directions: utils.StringMap{utils.OUT: true}}
	ub := &Account{}
	a := &Action{BalanceType: utils.VOICE, Balance: mb1}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{BalanceType: utils.VOICE, Balance: mb2}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 1 || ub.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{BalanceType: utils.VOICE, Balance: mb3}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
}

func TestAccountExecuteTriggeredActions(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{utils.MONETARY: BalanceChain{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}, utils.VOICE: BalanceChain{&Balance{Value: 10, Weight: 20, DestinationIds: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: utils.MONETARY, Balances: BalanceChain{&Balance{Value: 1, Directions: utils.StringMap{utils.OUT: true}}}}},
		ActionTriggers: ActionTriggers{&ActionTrigger{BalanceType: utils.MONETARY, BalanceDirections: utils.StringMap{utils.OUT: true}, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 1, Directions: utils.NewStringMap(utils.OUT)}})
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 || ub.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// are set to executed
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 1, Directions: utils.NewStringMap(utils.OUT)}})
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 || ub.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// we can reset them
	ub.ResetActionTriggers(nil)
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 120 || ub.BalanceMap[utils.VOICE][0].GetValue() != 30 {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
}

func TestAccountExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{utils.MONETARY: BalanceChain{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}, utils.VOICE: BalanceChain{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 10, Weight: 20, DestinationIds: utils.StringMap{"NAT": true}}, &Balance{Directions: utils.NewStringMap(utils.OUT), Weight: 10, DestinationIds: utils.StringMap{"RET": true}}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: utils.MONETARY, Balances: BalanceChain{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 1}}}},
		ActionTriggers: ActionTriggers{&ActionTrigger{BalanceType: utils.MONETARY, BalanceDirections: utils.NewStringMap(utils.OUT), ThresholdValue: 100, ThresholdType: TRIGGER_MIN_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 || ub.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB_OREDER",
		BalanceMap:     map[string]BalanceChain{utils.MONETARY: BalanceChain{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: utils.MONETARY, Balances: BalanceChain{&Balance{Value: 1, Directions: utils.NewStringMap(utils.OUT)}}}},
		ActionTriggers: ActionTriggers{&ActionTrigger{BalanceType: utils.MONETARY, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS_ORDER"}},
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 1}})
	if len(ub.BalanceMap[utils.MONETARY]) != 1 || ub.BalanceMap[utils.MONETARY][0].GetValue() != 10 {

		t.Errorf("Error executing triggered actions in order %v BAL: %+v", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.MONETARY][1])
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &Account{
		Id: "TEST_UB_OREDER",
		BalanceMap: map[string]BalanceChain{utils.MONETARY: BalanceChain{
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)}}, utils.VOICE: BalanceChain{
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
		}},
	}
	ub.CleanExpiredBalances()
	if len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Error("Error cleaning expired balances!")
	}
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error cleaning expired minute buckets!")
	}
}

func TestAccountUnitCounting(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 20 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutbound(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 30 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutboundInbound(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 10 {
		t.Errorf("Error counting units: %+v", ub.UnitCounters[0])
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.OUT)}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: utils.MONETARY, Balance: &Balance{Value: 10, Directions: utils.NewStringMap(utils.IN)}})
	if len(ub.UnitCounters) != 1 || (ub.UnitCounters[0].BalanceType != utils.MONETARY || ub.UnitCounters[0].Balances[0].GetValue() != 30) { // for the moment no in/out distinction
		t.Error("Error counting units")
	}
}

func TestAccountRefund(t *testing.T) {
	ub := &Account{
		BalanceMap: map[string]BalanceChain{
			utils.MONETARY: BalanceChain{
				&Balance{Uuid: "moneya", Value: 100},
			},
			utils.VOICE: BalanceChain{
				&Balance{Uuid: "minutea", Value: 10, Weight: 20, DestinationIds: utils.StringMap{"NAT": true}},
				&Balance{Uuid: "minuteb", Value: 10, DestinationIds: utils.StringMap{"RET": true}},
			},
		},
	}
	increments := Increments{
		&Increment{Cost: 2, BalanceInfo: &BalanceInfo{UnitBalanceUuid: "", MoneyBalanceUuid: "moneya"}},
		&Increment{Cost: 2, Duration: 3 * time.Second, BalanceInfo: &BalanceInfo{UnitBalanceUuid: "minutea", MoneyBalanceUuid: "moneya"}},
		&Increment{Duration: 4 * time.Second, BalanceInfo: &BalanceInfo{UnitBalanceUuid: "minuteb", MoneyBalanceUuid: ""}},
	}
	for _, increment := range increments {
		ub.refundIncrement(increment, utils.VOICE, false)
	}
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 104 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 13 ||
		ub.BalanceMap[utils.VOICE][1].GetValue() != 14 {
		t.Error("Error refounding money: ", ub.BalanceMap[utils.VOICE][1].GetValue())
	}
}

func TestDebitShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 2, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{Id: "rif", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{Id: "groupie", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: []string{rif.Id, groupie.Id}, AccountParameters: map[string]*SharingParameters{"*any": &SharingParameters{Strategy: STRATEGY_MINE_RANDOM}}}

	accountingStorage.SetAccount(groupie)
	ratingStorage.SetSharedGroup(sg)
	cache2go.Cache(utils.SHARED_GROUP_PREFIX+"SG_TEST", sg)
	cc, err := rif.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rif.BalanceMap[utils.MONETARY][0].GetValue() != 0 {
		t.Errorf("Error debiting from shared group: %+v", rif.BalanceMap[utils.MONETARY][0])
	}
	groupie, _ = accountingStorage.GetAccount("groupie")
	if groupie.BalanceMap[utils.MONETARY][0].GetValue() != 10 {
		t.Errorf("Error debiting from shared group: %+v", groupie.BalanceMap[utils.MONETARY][0])
	}

	if len(cc.Timespans) != 1 {
		t.Errorf("Wrong number of timespans: %v", cc.Timespans)
	}
	if len(cc.Timespans[0].Increments) != 6 {
		t.Errorf("Wrong number of increments: %v", cc.Timespans[0].Increments)
		for index, incr := range cc.Timespans[0].Increments {
			t.Errorf("I%d: %+v (%+v)", index, incr, incr.BalanceInfo)
		}
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[1].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[2].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[3].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[4].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[5].BalanceInfo.AccountId != "groupie" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
}

func TestMaxDurationShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 2, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{Id: "rif", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{Id: "groupie", BalanceMap: map[string]BalanceChain{
		utils.MONETARY: BalanceChain{&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: []string{rif.Id, groupie.Id}, AccountParameters: map[string]*SharingParameters{"*any": &SharingParameters{Strategy: STRATEGY_MINE_RANDOM}}}

	accountingStorage.SetAccount(groupie)
	ratingStorage.SetSharedGroup(sg)
	cache2go.Cache(utils.SHARED_GROUP_PREFIX+"SG_TEST", sg)
	duration, err := cd.getMaxSessionDuration(rif)
	if err != nil {
		t.Error("Error getting max session duration from shared group: ", err)
	}
	if duration != 1*time.Minute {
		t.Error("Wrong max session from shared group: ", duration)
	}

}

func TestDebitSMS(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 1, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.SMS,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.SMS:      BalanceChain{&Balance{Uuid: "testm", Value: 100, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}}},
		utils.MONETARY: BalanceChain{&Balance{Value: 21}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.SMS][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.SMS][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGeneric(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 1, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 1 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.GENERIC,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.GENERIC:  BalanceChain{&Balance{Uuid: "testm", Value: 100, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}}},
		utils.MONETARY: BalanceChain{&Balance{Value: 21}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.UnitBalanceUuid != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitDataUnits(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 2, RateIncrement: 1 * time.Second, RateUnit: time.Minute},
							&Rate{GroupIntervalStart: 60, Value: 1, RateIncrement: 1 * time.Second, RateUnit: time.Second},
						},
					},
				},
			},
		},
		TOR: utils.DATA,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.DATA:     BalanceChain{&Balance{Uuid: "testm", Value: 100, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}}},
		utils.MONETARY: BalanceChain{&Balance{Value: 21}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	// test rating information
	ts := cc.Timespans[0]
	if ts.MatchedSubject != "testm" || ts.MatchedPrefix != "0723" || ts.MatchedDestId != "NAT" || ts.RatingPlanId != utils.META_NONE {
		t.Errorf("Error setting rating info: %+v", ts.ratingInfo)
	}
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if ts.Increments[0].BalanceInfo.UnitBalanceUuid != "testm" {
		t.Error("Error setting balance id to increment: ", ts.Increments[0])
	}
	if rifsBalance.BalanceMap[utils.DATA][0].GetValue() != 20 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(ts.Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.DATA][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitDataMoney(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 2, RateIncrement: time.Minute, RateUnit: time.Second},
						},
					},
				},
			},
		},
		TOR: utils.DATA,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		utils.DATA:     BalanceChain{&Balance{Uuid: "testm", Value: 0, Weight: 5, DestinationIds: utils.StringMap{"NAT": true}}},
		utils.MONETARY: BalanceChain{&Balance{Value: 160}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rifsBalance.BalanceMap[utils.DATA][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 0 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.DATA][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestAccountGetDefaultMoneyBalanceEmpty(t *testing.T) {
	acc := &Account{}
	defBal := acc.GetDefaultMoneyBalance()
	if defBal == nil || len(acc.BalanceMap) != 1 || !defBal.IsDefault() {
		t.Errorf("Bad default money balance: %+v", defBal)
	}
}

func TestAccountGetDefaultMoneyBalance(t *testing.T) {
	acc := &Account{}
	acc.BalanceMap = make(map[string]BalanceChain)
	tag := utils.MONETARY
	acc.BalanceMap[tag] = append(acc.BalanceMap[tag], &Balance{Weight: 10})
	defBal := acc.GetDefaultMoneyBalance()
	if defBal == nil || len(acc.BalanceMap[tag]) != 2 || !defBal.IsDefault() {
		t.Errorf("Bad default money balance: %+v", defBal)
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}

	ub1 := &Account{Id: "other", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}

func BenchmarkAccountStorageStoreRestore(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 21}}}}
	for i := 0; i < b.N; i++ {
		accountingStorage.SetAccount(rifsBalance)
		accountingStorage.GetAccount(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIds: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIds: utils.StringMap{"RET": true}}
	ub1 := &Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{utils.VOICE: BalanceChain{b1, b2}, utils.MONETARY: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}
