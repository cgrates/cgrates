/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	if !b.Equal(b1) {
		t.Errorf("Balance store/restore failed: expected %v was %v", b, b1)
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
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	accountingStorage.SetAccount(rifsBalance)
	ub1, err := accountingStorage.GetAccount("other")
	if err != nil || !ub1.BalanceMap[CREDIT+OUTBOUND].Equal(rifsBalance.BalanceMap[CREDIT+OUTBOUND]) {
		t.Log("UB: ", ub1)
		t.Errorf("Expected %v was %v", rifsBalance.BalanceMap[CREDIT+OUTBOUND], ub1.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 200}}}}
	cd := &CallDescriptor{
		TOR:          "0",
		Tenant:       "vdf",
		TimeStart:    time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:    0,
		CallDuration: 10 * time.Second,
		Direction:    OUTBOUND,
		Destination:  "0723",
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 110 * time.Second
	if credit != 200 || seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetSpecialPricedSeconds(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET", RatingSubject: "minu"}

	ub1 := &Account{
		Id: "OUT:CUSTOMER_1:rif",
		BalanceMap: map[string]BalanceChain{
			MINUTES + OUTBOUND: BalanceChain{b1, b2},
			CREDIT + OUTBOUND:  BalanceChain{&Balance{Value: 21}},
		},
	}
	cd := &CallDescriptor{
		TOR:         "0",
		Tenant:      "vdf",
		TimeStart:   time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 4, 15, 46, 60, 0, time.UTC),
		LoopIndex:   0,
		Direction:   OUTBOUND,
		Destination: "0723",
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 20 * time.Second
	if credit != 0 || seconds != expected || len(bucketList) != 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestAccountStorageStore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	accountingStorage.SetAccount(rifsBalance)
	result, err := accountingStorage.GetAccount(rifsBalance.Id)
	if err != nil || rifsBalance.Id != result.Id ||
		len(rifsBalance.BalanceMap[MINUTES+OUTBOUND]) < 2 || len(result.BalanceMap[MINUTES+OUTBOUND]) < 2 ||
		!(rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Equal(result.BalanceMap[MINUTES+OUTBOUND][0])) ||
		!(rifsBalance.BalanceMap[MINUTES+OUTBOUND][1].Equal(result.BalanceMap[MINUTES+OUTBOUND][1])) ||
		!rifsBalance.BalanceMap[CREDIT+OUTBOUND].Equal(result.BalanceMap[CREDIT+OUTBOUND]) {
		t.Errorf("Expected %v was %v", rifsBalance, result)
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" {
		t.Logf("%+v", cc.Timespans[0])
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Value: 21}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	t.Logf("%+v", cc.Timespans)
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 21 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
}
func TestDebitCreditZeroMixedMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testm", Value: 70, Weight: 5, DestinationId: "NAT", RatingSubject: "*zero1m"}
	b2 := &Balance{Uuid: "tests", Value: 10, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				ratingInfo:   &RatingInfo{},
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1, b2},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Value: 21}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "tests" ||
		cc.Timespans[1].Increments[0].BalanceInfo.MinuteBalanceUuid != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0], cc.Timespans[1].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][1].Value != 0 ||
		rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err == nil {
		t.Error("Showing no enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditHasCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != time.Minute || cc.Timespans[1].GetDuration() != 20*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSplitMinutesMoney(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				CallDuration: 0,
				ratingInfo:   &RatingInfo{},
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 40 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 10*time.Second || cc.Timespans[1].GetDuration() != 10*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditMoreTimespans(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 150, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 30 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
}

func TestDebitCreditMoreTimespansMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	b2 := &Balance{Uuid: "testa", Value: 150, Weight: 5, DestinationId: "NAT", RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1, b2},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 ||
		rifsBalance.BalanceMap[MINUTES+OUTBOUND][1].Value != 130 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][1], cc.Timespans[1])
	}
}

func TestDebitCreditNoConectFeeCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{ConnectFee: 10.0, Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}

	if len(cc.Timespans) != 2 || rifsBalance.BalanceMap[CREDIT+OUTBOUND].GetTotalValue() != -20 {
		t.Error("Error cutting at no connect fee: ", rifsBalance.BalanceMap[CREDIT+OUTBOUND].GetTotalValue())
	}
}

func TestDebitCreditMoneyOnly(t *testing.T) {
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				ratingInfo:   &RatingInfo{},
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "money", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err == nil {
		t.Error("Missing noy enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "money" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Logf("%+v", cc.Timespans[0].Increments)
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0].BalanceInfo)
	}
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != -30 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][0])
	}
	if len(cc.Timespans) != 3 ||
		cc.Timespans[0].GetDuration() != 10*time.Second ||
		cc.Timespans[1].GetDuration() != 40*time.Second ||
		cc.Timespans[2].GetDuration() != 30*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans[2].GetDuration())
	}
}

func TestDebitCreditSubjectMinutes(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 250, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 350}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 180 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 279 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
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
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "moneya", Value: 75, DestinationId: "NAT", RatingSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 4 {
		t.Errorf("Error extracting minutes from balance: %+v",
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 40, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 55, 0, time.UTC),
				CallDuration: 55 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 150, RatingSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.MoneyBalanceUuid != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 94 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 40*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts)
		}
		t.Error("Error truncating extra timespans: ", len(cc.Timespans), cc.Timespans[0].GetDuration())
	}
}

func TestDebitCreditSubjectMixedMoreTS(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50, RatingSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}
	//t.Logf("%+v %+v", cc.Timespans[0], cc.Timespans[1])
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][1].Value != -31 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value,
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][1].Value)
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 50*time.Second || cc.Timespans[1].GetDuration() != 30*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts.GetDuration())
		}
		t.Error("Error truncating extra timespans: ", len(cc.Timespans))
	}
}

func TestDebitCreditSubjectMixedPartPay(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 75, RatingSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}
	//t.Logf("%+v %+v", cc.Timespans[0], cc.Timespans[1])
	if cc.Timespans[0].Increments[0].BalanceInfo.MinuteBalanceUuid != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][1].Value != -11 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value,
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][1].Value)
	}
	if len(cc.Timespans) != 3 || cc.Timespans[0].GetDuration() != 70*time.Second || cc.Timespans[1].GetDuration() != 5*time.Second || cc.Timespans[2].GetDuration() != 5*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts.GetDuration())
		}
		t.Error("Error truncating extra timespans: ", len(cc.Timespans))
	}
}

func TestAccountdebitBalance(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{SMS: BalanceChain{&Balance{Value: 14}}, DATA: BalanceChain{&Balance{Value: 1204}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	newMb := &Balance{Weight: 20, DestinationId: "NEW"}
	a := &Action{BalanceType: MINUTES, Direction: OUTBOUND, Balance: newMb}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 3 || ub.BalanceMap[MINUTES+OUTBOUND][2] != newMb {
		t.Error("Error adding minute bucket!", len(ub.BalanceMap[MINUTES+OUTBOUND]), ub.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestAccountdebitBalanceExists(t *testing.T) {

	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{SMS + OUTBOUND: BalanceChain{&Balance{Value: 14}}, DATA + OUTBOUND: BalanceChain{&Balance{Value: 1024}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 15, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	newMb := &Balance{Value: -10, Weight: 20, DestinationId: "NAT"}
	a := &Action{BalanceType: MINUTES, Direction: OUTBOUND, Balance: newMb}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinuteNil(t *testing.T) {
	ub := &Account{
		Id:            "rif",
		AllowNegative: true,
		BalanceMap:    map[string]BalanceChain{SMS + OUTBOUND: BalanceChain{&Balance{Value: 14}}, DATA + OUTBOUND: BalanceChain{&Balance{Value: 1024}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	ub.debitBalanceAction(nil)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinutBucketEmpty(t *testing.T) {
	mb1 := &Balance{Value: -10, DestinationId: "NAT"}
	mb2 := &Balance{Value: -10, DestinationId: "NAT"}
	mb3 := &Balance{Value: -10, DestinationId: "OTHER"}
	ub := &Account{}
	a := &Action{BalanceType: MINUTES, Direction: OUTBOUND, Balance: mb1}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
	a = &Action{BalanceType: MINUTES, Direction: OUTBOUND, Balance: mb2}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 1 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
	a = &Action{BalanceType: MINUTES, Direction: OUTBOUND, Balance: mb3}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestAccountExecuteTriggeredActions(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 10, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: CREDIT, Direction: OUTBOUND, Balances: BalanceChain{&Balance{Value: 1}}}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceType: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	// are set to executed
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	// we can reset them
	ub.resetActionTriggers(nil)
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 120 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 30 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
}

func TestAccountExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 10, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: CREDIT, Direction: OUTBOUND, Balances: BalanceChain{&Balance{Value: 1}}}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceType: CREDIT, Direction: OUTBOUND, ThresholdValue: 100, ThresholdType: TRIGGER_MIN_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
}

func TestAccountExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &Account{
		Id:             "TEST_UB_OREDER",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceType: CREDIT, Direction: OUTBOUND, Balances: BalanceChain{&Balance{Value: 1}}}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceType: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS_ORDER"}},
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 1}})
	if len(ub.BalanceMap[CREDIT+OUTBOUND]) != 1 || ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 10 {
		t.Error("Error executing triggered actions in order", ub.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &Account{
		Id: "TEST_UB_OREDER",
		BalanceMap: map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)}}, MINUTES + OUTBOUND: BalanceChain{
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
		}},
	}
	ub.CleanExpiredBalancesAndBuckets()
	if len(ub.BalanceMap[CREDIT+OUTBOUND]) != 2 {
		t.Error("Error cleaning expired balances!")
	}
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 1 {
		t.Error("Error cleaning expired minute buckets!")
	}
}

func TestAccountUnitCounting(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 20 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutbound(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 30 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutboundInbound(t *testing.T) {
	ub := &Account{}
	ub.countUnits(&Action{BalanceType: CREDIT, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 10 {
		t.Errorf("Error counting units: %+v", ub.UnitCounters[0])
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceType: CREDIT, Direction: INBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 2 && ub.UnitCounters[1].BalanceType != CREDIT || ub.UnitCounters[0].Balances[0].Value != 20 || ub.UnitCounters[1].Balances[0].Value != 10 {
		t.Error("Error counting units")
	}
}

func TestAccountRefund(t *testing.T) {
	ub := &Account{
		BalanceMap: map[string]BalanceChain{
			CREDIT + OUTBOUND: BalanceChain{
				&Balance{Uuid: "moneya", Value: 100},
			},
			MINUTES + OUTBOUND: BalanceChain{
				&Balance{Uuid: "minutea", Value: 10, Weight: 20, DestinationId: "NAT"},
				&Balance{Uuid: "minuteb", Value: 10, DestinationId: "RET"},
			},
		},
	}
	increments := Increments{
		&Increment{Cost: 2, BalanceInfo: &BalanceInfo{MinuteBalanceUuid: "", MoneyBalanceUuid: "moneya"}},
		&Increment{Cost: 2, Duration: 3 * time.Second, BalanceInfo: &BalanceInfo{MinuteBalanceUuid: "minutea", MoneyBalanceUuid: "moneya"}},
		&Increment{Duration: 4 * time.Second, BalanceInfo: &BalanceInfo{MinuteBalanceUuid: "minuteb", MoneyBalanceUuid: ""}},
	}
	for _, increment := range increments {
		ub.refundIncrement(increment, OUTBOUND, false)
	}
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 104 ||
		ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 13 ||
		ub.BalanceMap[MINUTES+OUTBOUND][1].Value != 14 {
		t.Error("Error refounding money: ", ub.BalanceMap[MINUTES+OUTBOUND][1].Value)
	}
}

func TestDebitShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		TOR:         "0",
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				CallDuration: 55 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 2, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	rif := &Account{Id: "rif", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "moneya", Value: 60, SharedGroup: "SG_TEST"}},
	}}
	groupie := &Account{Id: "groupie", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "moneyc", Value: 70, SharedGroup: "SG_TEST"}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: []string{rif.Id, groupie.Id}, AccountParameters: map[string]*SharingParameters{"*any": &SharingParameters{Strategy: STRATEGY_MINE_RANDOM}}}

	accountingStorage.SetAccount(groupie)
	accountingStorage.SetSharedGroup(sg)
	cache2go.Cache(SHARED_GROUP_PREFIX+"SG_TEST", sg)
	err := rif.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rif.BalanceMap[CREDIT+OUTBOUND][0].Value != 0 {
		t.Errorf("Error debiting from shared group: %+v", rif.BalanceMap[CREDIT+OUTBOUND][0])
	}
	groupie, _ = accountingStorage.GetAccount("groupie")
	if groupie.BalanceMap[CREDIT+OUTBOUND][0].Value != 10 {
		t.Errorf("Error debiting from shared group: %+v", groupie.BalanceMap[CREDIT+OUTBOUND][0])
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
	if cc.Timespans[0].Increments[0].BalanceInfo.AccountId != "rif" ||
		cc.Timespans[0].Increments[1].BalanceInfo.AccountId != "rif" ||
		cc.Timespans[0].Increments[2].BalanceInfo.AccountId != "rif" ||
		cc.Timespans[0].Increments[3].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[4].BalanceInfo.AccountId != "groupie" ||
		cc.Timespans[0].Increments[5].BalanceInfo.AccountId != "groupie" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}

	ub1 := &Account{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}

func BenchmarkAccountStorageStoreRestore(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &Account{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	for i := 0; i < b.N; i++ {
		accountingStorage.SetAccount(rifsBalance)
		accountingStorage.GetAccount(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &Account{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}
