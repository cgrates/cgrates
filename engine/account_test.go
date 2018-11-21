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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	NAT = &Destination{Id: "NAT",
		Prefixes: []string{"0257", "0256", "0723"}}
	RET = &Destination{Id: "RET",
		Prefixes: []string{"0723", "0724"}}
)

func TestBalanceStoreRestore(t *testing.T) {
	b := &Balance{Value: 14, Weight: 1, Uuid: "test",
		ExpirationDate: time.Date(2013, time.July, 15, 17, 48, 0, 0, time.UTC)}
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
	//t.Logf("INITIAL: %+v", b)
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

func TestBalancesStoreRestore(t *testing.T) {
	bc := Balances{&Balance{Value: 14,
		ExpirationDate: time.Date(2013, time.July, 15, 17, 48, 0, 0, time.UTC)},
		&Balance{Value: 1024}}
	output, err := marsh.Marshal(bc)
	if err != nil {
		t.Error("Error storing balance chain: ", err)
	}
	bc1 := Balances{}
	err = marsh.Unmarshal(output, &bc1)
	if err != nil {
		t.Error("Error restoring balance chain: ", err)
	}
	if !bc.Equal(bc1) {
		t.Errorf("Balance chain store/restore failed: expected %v was %v", bc, bc1)
	}
}

func TestAccountStorageStoreRestore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20,
		DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: 21}}}}
	dm.DataDB().SetAccount(rifsBalance)
	ub1, err := dm.DataDB().GetAccount("other")
	if err != nil ||
		!ub1.BalanceMap[utils.MONETARY].Equal(rifsBalance.BalanceMap[utils.MONETARY]) {
		t.Log("UB: ", ub1)
		t.Errorf("Expected %v was %v", rifsBalance, ub1)
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20,
		DestinationIDs: utils.StringMap{"RET": true}}
	ub1 := &Account{ID: "CUSTOMER_1:rif",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: 200}}}}
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
	b1 := &Balance{Value: 10, Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "minu"}
	b2 := &Balance{Value: 100, Weight: 20,
		DestinationIDs: utils.StringMap{"RET": true}, RatingSubject: "minu"}

	ub1 := &Account{
		ID: "OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: 21}},
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
	if credit != 0 || seconds != expected ||
		len(bucketList) != 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		for _, b := range bucketList {
			t.Logf("Balance: %+v", b)
		}
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestAccountStorageStore(t *testing.T) {
	if DB == "mongo" {
		return // mongo will have a problem with null and {} so the Equal will not work
	}
	b1 := &Balance{Value: 10, Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: 21}}}}
	dm.DataDB().SetAccount(rifsBalance)
	result, err := dm.DataDB().GetAccount(rifsBalance.ID)
	if err != nil || rifsBalance.ID != result.ID ||
		len(rifsBalance.BalanceMap[utils.VOICE]) < 2 ||
		len(result.BalanceMap[utils.VOICE]) < 2 ||
		!(rifsBalance.BalanceMap[utils.VOICE][0].Equal(result.BalanceMap[utils.VOICE][0])) ||
		!(rifsBalance.BalanceMap[utils.VOICE][1].Equal(result.BalanceMap[utils.VOICE][1])) ||
		!rifsBalance.BalanceMap[utils.MONETARY].Equal(result.BalanceMap[utils.MONETARY]) {
		t.Errorf("Expected %s was %s", utils.ToIJSON(rifsBalance), utils.ToIJSON(result))
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{
		Uuid: "testb", Value: 10 * float64(time.Second), Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{Rating: &RIRate{
					Rates: RateGroups{&Rate{GroupIntervalStart: 0,
						Value: 100, RateIncrement: 10 * time.Second,
						RateUnit: time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{b1},
			utils.MONETARY: Balances{&Balance{
				Categories: utils.NewStringMap("0"), Value: 21}}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" {
		t.Logf("%+v", cc.Timespans[0])
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditBlocker(t *testing.T) {
	b1 := &Balance{Uuid: "testa", Value: 0.1152,
		Weight: 20, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "passmonde", Blocker: true}
	b2 := &Balance{Uuid: "*default", Value: 1.5, Weight: 0}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{ConnectFee: 0.15,
						Rates: RateGroups{&Rate{GroupIntervalStart: 0,
							Value: 0.1, RateIncrement: time.Second,
							RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
		TOR:              utils.VOICE,
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{b1, b2}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, true, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if len(cc.Timespans) != 0 {
		t.Error("Wrong call cost: ", utils.ToIJSON(cc))
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 0.1152 ||
		rifsBalance.BalanceMap[utils.MONETARY][1].GetValue() != 1.5 {
		t.Error("should not have touched the balances: ",
			utils.ToIJSON(rifsBalance.BalanceMap[utils.MONETARY]))
	}
}

func TestDebitFreeEmpty(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "112",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{Rating: &RIRate{
					ConnectFee: 0, Rates: RateGroups{
						&Rate{GroupIntervalStart: 0, Value: 0,
							RateIncrement: time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		deductConnectFee: true,
		TOR:              utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Tenant:       "CUSTOMER_1",
		Subject:      "rif:from:tm",
		Destination:  "112",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	// empty account
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, true, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if len(cc.Timespans) == 0 || cc.Cost != 0 {
		t.Error("Wrong call cost: ", utils.ToIJSON(cc))
	}
	if len(rifsBalance.BalanceMap[utils.MONETARY]) != 0 {
		t.Error("should not have touched the balances: ",
			utils.ToIJSON(rifsBalance.BalanceMap[utils.MONETARY]))
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0, Value: 100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{
		ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	//t.Logf("%+v", cc.Timespans)
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Errorf("Error setting balance id to increment: %s",
			utils.ToJSON(cc))
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Errorf("Error extracting minutes from balance: %s",
			utils.ToJSON(rifsBalance.BalanceMap[utils.VOICE][0]))
	}
}

func TestDebitCreditZeroMixedMinute(t *testing.T) {
	b1 := &Balance{
		Uuid: "testm", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "*zero1m", Weight: 5}
	b2 := &Balance{Uuid: "tests", Value: 10 * float64(time.Second), Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0,
							Value: 100, RateIncrement: 10 * time.Second,
							RateUnit: time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "tests" ||
		cc.Timespans[1].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans)
	}
	if rifsBalance.BalanceMap[utils.VOICE][1].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Logf("TS0: %+v", cc.Timespans[0])
		t.Logf("TS1: %+v", cc.Timespans[1])
		t.Errorf("Error extracting minutes from balance: %+v", rifsBalance.BalanceMap[utils.VOICE][1])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "*zero1m", Weight: 10}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Showing no enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
	if len(cc.Timespans) != 1 ||
		cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditHasCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0, Value: 1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0,
							Value:         1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1},
		utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 110}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 3 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSplitMinutesMoney(t *testing.T) {
	b1 := &Balance{Uuid: "testb",
		Value:          10 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				DurationIndex: 0,
				ratingInfo:    &RatingInfo{},
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1},
			utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 50}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != 1*time.Second {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0].Duration)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 2 ||
		cc.Timespans[0].GetDuration() != 10*time.Second ||
		cc.Timespans[1].GetDuration() != 20*time.Second {
		t.Error("Error truncating extra timespans: ",
			cc.Timespans[1].GetDuration())
	}
}

func TestDebitCreditMoreTimespans(t *testing.T) {
	b1 := &Balance{Uuid: "testb",
		Value:          150 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0, Value: 100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 30*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditMoreTimespansMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1m"}
	b2 := &Balance{Uuid: "testa", Value: 150 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         5, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0,
							Value:         100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{
		ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{b1, b2},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.VOICE][1].GetValue() != 130*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][1], cc.Timespans[1])
	}
}

func TestDebitCreditNoConectFeeCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{ConnectFee: 10.0,
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}
	if len(cc.Timespans) != 1 ||
		rifsBalance.BalanceMap[utils.MONETARY].GetTotalValue() != 0 {
		t.Error("Error cutting at no connect fee: ",
			rifsBalance.BalanceMap[utils.MONETARY])
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
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				ratingInfo:    &RatingInfo{},
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Uuid: "money", Value: 50}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Missing noy enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "money" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Logf("%+v", cc.Timespans[0].Increments)
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0].BalanceInfo)
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
	b1 := &Balance{Uuid: "testb",
		Categories:     utils.NewStringMap("0"),
		Value:          250 * float64(time.Second),
		Weight:         10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "minu"}
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
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&Rate{GroupIntervalStart: 0,
							Value:         1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1},
			utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 350}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Errorf("Error setting balance id to increment: %+v",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 180*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 280 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 ||
		cc.Timespans[0].GetDuration() != 70*time.Second {
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
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Uuid: "moneya", Value: 75,
					DestinationIDs: utils.StringMap{"NAT": true},
					RatingSubject:  "minu"}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 5 {
		t.Errorf("Error extracting minutes from balance: %+v",
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 ||
		cc.Timespans[0].GetDuration() != 70*time.Second {
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
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1},
		utils.MONETARY:  Balances{&Balance{Uuid: "moneya", Value: 19500, RatingSubject: "minu"}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.Money.UUID != "moneya" ||
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
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{Value: 14}},
			utils.DATA: Balances{&Balance{Value: 1204}},
			utils.VOICE: Balances{
				&Balance{Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	newMb := &BalanceFilter{
		Type:           utils.StringPointer(utils.VOICE),
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NEW": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	a := &Action{Balance: newMb}
	ub.debitBalanceAction(a, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 3 ||
		!ub.BalanceMap[utils.VOICE][2].DestinationIDs.Equal(*newMb.DestinationIDs) {
		t.Errorf("Error adding minute bucket! %d %+v %+v",
			len(ub.BalanceMap[utils.VOICE]), ub.BalanceMap[utils.VOICE][2], newMb)
	}
}

/*func TestAccountDisableBalance(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]Balances{utils.SMS: Balances{&Balance{Value: 14}}, utils.DATA: Balances{&Balance{Value: 1204}}, utils.VOICE: Balances{&Balance{Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	newMb := &BalanceFilter{
		Type:           utils.StringPointer(utils.VOICE),
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	a := &Action{Balance: newMb}
	ub.enableDisableBalanceAction(a)
	if len(ub.BalanceMap[utils.VOICE]) != 2 || ub.BalanceMap[utils.VOICE][0].Disabled != true {
		for _, b := range ub.BalanceMap[utils.VOICE] {
			t.Logf("Balance: %+v", b)
		}
		t.Errorf("Error disabling balance! %d %+v %+v", len(ub.BalanceMap[utils.VOICE]), ub.BalanceMap[utils.VOICE][0], newMb)
	}
}*/

func TestAccountdebitBalanceExists(t *testing.T) {
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{Value: 14}},
			utils.DATA: Balances{&Balance{Value: 1024}},
			utils.VOICE: Balances{
				&Balance{
					Value: 15, Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true},
					Directions:     utils.NewStringMap(utils.OUT)},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	newMb := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.VOICE),
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	a := &Action{Balance: newMb}
	ub.debitBalanceAction(a, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 || ub.BalanceMap[utils.VOICE][0].GetValue() != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinuteNil(t *testing.T) {
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{Value: 14}},
			utils.DATA: Balances{&Balance{Value: 1024}},
			utils.VOICE: Balances{
				&Balance{Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	ub.debitBalanceAction(nil, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinutBucketEmpty(t *testing.T) {
	mb1 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.VOICE),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	mb2 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.VOICE),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	mb3 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.VOICE),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"OTHER": true}),
		Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
	}
	ub := &Account{}
	a := &Action{Balance: mb1}
	ub.debitBalanceAction(a, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{Balance: mb2}
	ub.debitBalanceAction(a, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 1 || ub.BalanceMap[utils.VOICE][0].GetValue() != 20 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{Balance: mb3}
	ub.debitBalanceAction(a, false, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
}

func TestAccountExecuteTriggeredActions(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}},
			utils.VOICE: Balances{
				&Balance{Value: 10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true},
					Directions:     utils.StringMap{utils.OUT: true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{Counters: CounterFilters{
					&CounterFilter{Value: 1,
						Filter: &BalanceFilter{
							Type:       utils.StringPointer(utils.MONETARY),
							Directions: utils.StringMapPointer(utils.StringMap{utils.OUT: true})}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2, ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				ActionsID: "TEST_ACTIONS"}},
	}
	ub.countUnits(1, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MONETARY][0].GetValue(),
			ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// are set to executed
	ub.countUnits(1, utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// we can reset them
	ub.ResetActionTriggers(nil)
	ub.countUnits(10, utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 120 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 30*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MONETARY][0].GetValue(),
			ub.BalanceMap[utils.VOICE][0].GetValue())
	}
}

func TestAccountExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Directions: utils.NewStringMap(utils.OUT),
					Value:      100}},
			utils.VOICE: Balances{
				&Balance{
					Directions:     utils.NewStringMap(utils.OUT),
					Value:          10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{
					Directions:     utils.NewStringMap(utils.OUT),
					Weight:         10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{Counters: CounterFilters{
					&CounterFilter{Filter: &BalanceFilter{
						Type:       utils.StringPointer(utils.MONETARY),
						Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
						Value: 1.0}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 100,
				ThresholdType:  utils.TRIGGER_MIN_EVENT_COUNTER,
				ActionsID:      "TEST_ACTIONS"}},
	}
	ub.countUnits(1, utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue() != 110 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MONETARY][0].GetValue(),
			ub.BalanceMap[utils.VOICE][0].GetValue(),
			len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB_OREDER",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Directions: utils.NewStringMap(utils.OUT),
					Value: 100}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{Counters: CounterFilters{
					&CounterFilter{Value: 1, Filter: &BalanceFilter{
						Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
						Type:       utils.StringPointer(utils.MONETARY)}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 2,
				ThresholdType:  utils.TRIGGER_MAX_EVENT_COUNTER,
				ActionsID:      "TEST_ACTIONS_ORDER"}},
	}

	ub.countUnits(1, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.BalanceMap[utils.MONETARY]) != 1 ||
		ub.BalanceMap[utils.MONETARY][0].GetValue() != 10 {

		t.Errorf("Error executing triggered actions in order %v",
			ub.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestAccountExecuteTriggeredDayWeek(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}, utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{UniqueID: "day_trigger", Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}, ThresholdValue: 10, ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS"},
			&ActionTrigger{UniqueID: "week_trigger", Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}, ThresholdValue: 100, ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.InitCounters()
	if len(ub.UnitCounters) != 1 || len(ub.UnitCounters[utils.MONETARY][0].Counters) != 2 {
		t.Error("Error initializing counters: ", ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}

	ub.countUnits(1, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 1 ||
		ub.UnitCounters[utils.MONETARY][0].Counters[1].Value != 1 {
		t.Error("Error incrementing both counters", ub.UnitCounters[utils.MONETARY][0].Counters[0].Value, ub.UnitCounters[utils.MONETARY][0].Counters[1].Value)
	}

	// we can reset them
	resetCountersAction(ub, &Action{Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), ID: utils.StringPointer("day_trigger")}}, nil, nil)
	if ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 0 ||
		ub.UnitCounters[utils.MONETARY][0].Counters[1].Value != 1 {
		t.Error("Error reseting both counters", ub.UnitCounters[utils.MONETARY][0].Counters[0].Value, ub.UnitCounters[utils.MONETARY][0].Counters[1].Value)
	}
}

func TestAccountExpActionTrigger(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100,
					ExpirationDate: time.Date(2015, time.November, 9, 9, 48, 0, 0, time.UTC)}},
			utils.VOICE: Balances{
				&Balance{Value: 10 * float64(time.Second), Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true},
					Directions:     utils.StringMap{utils.OUT: true}},
				&Balance{Weight: 10 * float64(time.Second),
					DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances", Balance: &BalanceFilter{
				Type:       utils.StringPointer(utils.MONETARY),
				Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))},
				ThresholdValue: 10, ThresholdType: utils.TRIGGER_BALANCE_EXPIRED,
				ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue() != 10 || // expired was cleaned
		ub.BalanceMap[utils.VOICE][0].GetValue() != 20*float64(time.Second) ||
		ub.ActionTriggers[0].Executed != true {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MONETARY][0].GetValue(),
			ub.BalanceMap[utils.VOICE][0].GetValue(),
			len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExpActionTriggerNotActivated(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}, utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances", ActivationDate: time.Date(2116, 2, 5, 18, 0, 0, 0, time.UTC), Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}, ThresholdValue: 10, ThresholdType: utils.TRIGGER_BALANCE_EXPIRED, ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue() != 100 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		ub.ActionTriggers[0].Executed != false {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExpActionTriggerExpired(t *testing.T) {
	ub := &Account{
		ID:         "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: 100}}, utils.VOICE: Balances{&Balance{Value: 10, Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances", ExpirationDate: time.Date(2016, 2, 4, 18, 0, 0, 0, time.UTC), Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY), Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}, ThresholdValue: 10, ThresholdType: utils.TRIGGER_BALANCE_EXPIRED, ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue() != 100 ||
		ub.BalanceMap[utils.VOICE][0].GetValue() != 10 ||
		len(ub.ActionTriggers) != 0 {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB_OREDER",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)}}, utils.VOICE: Balances{
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
		}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC),
			},
			&ActionTrigger{
				ActivationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC),
			},
		},
	}
	ub.CleanExpiredStuff()
	if len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Error("Error cleaning expired balances!")
	}
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error cleaning expired minute buckets!")
	}
	if len(ub.ActionTriggers) != 1 {
		t.Error("Error cleaning expired action triggers!")
	}
}

func TestAccountUnitCounting(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{
		utils.MONETARY: []*UnitCounter{&UnitCounter{
			Counters: CounterFilters{&CounterFilter{Value: 0}}}}}}
	ub.countUnits(10, utils.MONETARY, &CallCost{}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 ||
		ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MONETARY, &CallCost{}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 ||
		ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 0, Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}}}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 30 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutboundInbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: 0, Filter: &BalanceFilter{Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT))}}}}}}}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 10 {
		t.Errorf("Error counting units: %+v", ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MONETARY, &CallCost{Direction: utils.IN}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
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
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{ID: "groupie", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: utils.NewStringMap(rif.ID, groupie.ID), AccountParameters: map[string]*SharingParameters{"*any": &SharingParameters{Strategy: STRATEGY_MINE_RANDOM}}}

	dm.DataDB().SetAccount(groupie)
	dm.SetSharedGroup(sg, utils.NonTransactional)
	cc, err := rif.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rif.BalanceMap[utils.MONETARY][0].GetValue() != 0 {
		t.Errorf("Error debiting from shared group: %+v", rif.BalanceMap[utils.MONETARY][0])
	}
	groupie, _ = dm.DataDB().GetAccount("groupie")
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
	if cc.Timespans[0].Increments[0].BalanceInfo.AccountID != "groupie" ||
		cc.Timespans[0].Increments[1].BalanceInfo.AccountID != "groupie" ||
		cc.Timespans[0].Increments[2].BalanceInfo.AccountID != "groupie" ||
		cc.Timespans[0].Increments[3].BalanceInfo.AccountID != "groupie" ||
		cc.Timespans[0].Increments[4].BalanceInfo.AccountID != "groupie" ||
		cc.Timespans[0].Increments[5].BalanceInfo.AccountID != "groupie" {
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
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{ID: "groupie", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: utils.NewStringMap(rif.ID, groupie.ID), AccountParameters: map[string]*SharingParameters{"*any": &SharingParameters{Strategy: STRATEGY_MINE_RANDOM}}}

	dm.DataDB().SetAccount(groupie)
	dm.SetSharedGroup(sg, utils.NonTransactional)
	duration, err := cd.getMaxSessionDuration(rif)
	if err != nil {
		t.Error("Error getting max session duration from shared group: ", err)
	}
	if duration != 1*time.Minute {
		t.Error("Wrong max session from shared group: ", duration)
	}

}

func TestMaxDurationConnectFeeOnly(t *testing.T) {
	cd := &CallDescriptor{
		Tenant:        "cgrates.org",
		Category:      "call",
		TimeStart:     time.Date(2015, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:       time.Date(2015, 9, 24, 10, 58, 1, 0, time.UTC),
		Direction:     utils.OUT,
		Destination:   "4444",
		Subject:       "dy",
		Account:       "dy",
		TOR:           utils.VOICE,
		DurationIndex: 600,
	}
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{Uuid: "moneya", Value: 0.2}},
	}}

	duration, err := cd.getMaxSessionDuration(rif)
	if err != nil {
		t.Error("Error getting max session duration: ", err)
	}
	if duration != 0 {
		t.Error("Wrong max session: ", duration)
	}

}

func TestDebitSMS(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 1, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 1,
								RateUnit:      time.Nanosecond}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.SMS: Balances{
				&Balance{Uuid: "testm",
					Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MONETARY: Balances{
				&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.SMS][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.SMS][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGeneric(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 1, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 1,
								RateUnit:      time.Nanosecond}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.GENERIC: Balances{
				&Balance{Uuid: "testm", Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGenericBalance(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 1 * time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{
		ID: "other", BalanceMap: map[string]Balances{
			utils.GENERIC: Balances{
				&Balance{Uuid: "testm", Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true},
					Factor:         ValueFactor{utils.VOICE: 60 * float64(time.Second)}}},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue() != 99.49999 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Logf("%+v", cc.Timespans[0].Increments[0])
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGenericBalanceWithRatingSubject(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0, Value: 0,
								RateIncrement: time.Second,
								RateUnit:      time.Second}}}},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.GENERIC: Balances{
				&Balance{Uuid: "testm", Value: 100,
					Weight: 5, DestinationIDs: utils.StringMap{"NAT": true},
					Factor:        ValueFactor{utils.VOICE: 60 * float64(time.Second)},
					RatingSubject: "free"}},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0])
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue() != 99.49999 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Logf("%+v", cc.Timespans[0].Increments[0])
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitDataUnits(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 80, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value: 2, RateIncrement: 1,
								RateUnit: 1},
							&Rate{GroupIntervalStart: 60,
								Value:         1,
								RateIncrement: 1,
								RateUnit:      1},
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
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.DATA: Balances{
				&Balance{Uuid: "testm", Value: 100,
					Weight:         5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MONETARY: Balances{&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	// test rating information
	ts := cc.Timespans[0]
	if ts.MatchedSubject != "testm" || ts.MatchedPrefix != "0723" ||
		ts.MatchedDestId != "NAT" || ts.RatingPlanId != utils.META_NONE {
		t.Errorf("Error setting rating info: %+v", ts.ratingInfo)
	}
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if ts.Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", ts.Increments[0])
	}
	if rifsBalance.BalanceMap[utils.DATA][0].GetValue() != 20 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 21 {
		t.Log(ts.Increments)
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.DATA][0].GetValue(),
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
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
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{
		utils.DATA:     Balances{&Balance{Uuid: "testm", Value: 0, Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MONETARY: Balances{&Balance{Value: 160}},
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
	acc.BalanceMap = make(map[string]Balances)
	tag := utils.MONETARY
	acc.BalanceMap[tag] = append(acc.BalanceMap[tag], &Balance{Weight: 10})
	defBal := acc.GetDefaultMoneyBalance()
	if defBal == nil || len(acc.BalanceMap[tag]) != 2 || !defBal.IsDefault() {
		t.Errorf("Bad default money balance: %+v", defBal)
	}
}

func TestAccountInitCounters(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.VOICE),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.VOICE),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.SMS),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.SMS),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 1 ||
		len(a.UnitCounters[utils.VOICE][1].Counters) != 1 ||
		len(a.UnitCounters[utils.SMS][0].Counters) != 1 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, c := range uc.Counters {
					t.Logf("B: %+v", c)
				}
			}
		}
		t.Errorf("Error Initializing unit counters: %v", len(a.UnitCounters))
	}
}

func TestAccountDoubleInitCounters(t *testing.T) {
	a := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.VOICE),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.VOICE),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.SMS),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.SMS),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT, utils.IN)),
					Weight:     utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 1 ||
		len(a.UnitCounters[utils.VOICE][1].Counters) != 1 ||
		len(a.UnitCounters[utils.SMS][0].Counters) != 1 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, c := range uc.Counters {
					t.Logf("B: %+v", c)
				}
			}
		}
		t.Errorf("Error Initializing unit counters: %v", len(a.UnitCounters))
	}
}

func TestAccountGetBalancesForPrefixMixed(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) != 0 {
		t.Error("error excluding on mixed balances")
	}
}

func TestAccountGetBalancesForPrefixAllExcl(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": false, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) == 0 {
		t.Error("error finding balance on all excluded")
	}
}

func TestAccountGetBalancesForPrefixMixedGood(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": true},
				},
			},
		},
	}

	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) == 0 {
		t.Error("error finding on mixed balances good")
	}
}

func TestAccountGetBalancesForPrefixMixedBad(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) != 0 {
		t.Error("error excluding on mixed balances bad")
	}
}

func TestAccountNewAccountSummaryFromJSON(t *testing.T) {
	if acnt, err := NewAccountSummaryFromJSON("null"); err != nil {
		t.Error(err)
	} else if acnt != nil {
		t.Errorf("Expecting nil, received: %+v", acnt)
	}
}

func TestAccountAsAccountDigest(t *testing.T) {
	acnt1 := &Account{
		ID:            "cgrates.org:account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{ID: "sms1", Value: 14}},
			utils.DATA: Balances{&Balance{ID: "data1", Value: 1204}},
			utils.VOICE: Balances{
				&Balance{ID: "voice1", Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Value: 3600},
				&Balance{ID: "voice2", Weight: 10, DestinationIDs: utils.StringMap{"RET": true}, Value: 1200},
			},
		},
	}
	expectacntSummary := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "account1",
		BalanceSummaries: []*BalanceSummary{
			&BalanceSummary{ID: "sms1", Type: utils.SMS, Value: 14, Disabled: false},
			&BalanceSummary{ID: "data1", Type: utils.DATA, Value: 1204, Disabled: false},
			&BalanceSummary{ID: "voice1", Type: utils.VOICE, Value: 1204, Disabled: false},
			&BalanceSummary{ID: "voice2", Type: utils.VOICE, Value: 1200, Disabled: false},
		},
		AllowNegative: true,
		Disabled:      false,
	}
	acntSummary := acnt1.AsAccountSummary()
	if expectacntSummary.Tenant != acntSummary.Tenant ||
		expectacntSummary.ID != acntSummary.ID ||
		expectacntSummary.AllowNegative != acntSummary.AllowNegative ||
		expectacntSummary.Disabled != acntSummary.Disabled ||
		len(expectacntSummary.BalanceSummaries) != len(acntSummary.BalanceSummaries) {
		t.Errorf("Expecting: %+v, received: %+v", expectacntSummary, acntSummary)
	}
	// Since maps are unordered, slices will be too so we need to find element to compare
	for _, bd := range acntSummary.BalanceSummaries {
		if bd.ID == "sms1" && !reflect.DeepEqual(expectacntSummary.BalanceSummaries[0], bd) {
			t.Errorf("Expecting: %+v, received: %+v", expectacntSummary, acntSummary)
		}
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: 10, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}

	ub1 := &Account{ID: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}

func BenchmarkAccountStorageStoreRestore(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: 21}}}}
	for i := 0; i < b.N; i++ {
		dm.DataDB().SetAccount(rifsBalance)
		dm.DataDB().GetAccount(rifsBalance.ID)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	ub1 := &Account{ID: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}
