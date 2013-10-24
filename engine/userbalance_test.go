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
)

var (
	NAT = &Destination{Id: "NAT", Prefixes: []string{"0257", "0256", "0723"}}
	RET = &Destination{Id: "RET", Prefixes: []string{"0723", "0724"}}
)

func init() {
	populateTestActionsForTriggers()
}

func populateTestActionsForTriggers() {
	ats := []*Action{
		&Action{ActionType: "*topup", BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}},
		&Action{ActionType: "*topup", BalanceId: MINUTES, Direction: OUTBOUND, Balance: &Balance{Weight: 20, Value: 10, DestinationId: "NAT"}},
	}
	storageGetter.SetActions("TEST_ACTIONS", ats)
	ats1 := []*Action{
		&Action{ActionType: "*topup", BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}, Weight: 20},
		&Action{ActionType: "*reset_prepaid", Weight: 10},
	}
	storageGetter.SetActions("TEST_ACTIONS_ORDER", ats1)
}

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

func TestUserBalanceStorageStoreRestore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	storageGetter.SetUserBalance(rifsBalance)
	ub1, err := storageGetter.GetUserBalance("other")
	if err != nil || !ub1.BalanceMap[CREDIT+OUTBOUND].Equal(rifsBalance.BalanceMap[CREDIT+OUTBOUND]) {
		t.Log("UB: ", ub1)
		t.Errorf("Expected %v was %v", rifsBalance.BalanceMap[CREDIT+OUTBOUND], ub1.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 200}}}}
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
	seconds, credit, bucketList := ub1.getSecondsForPrefix(cd)
	expected := 110.0
	if credit != 200 || seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetSpecialPricedSeconds(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT", RateSubject: "minu"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET", RateSubject: "minu"}

	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
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
	seconds, credit, bucketList := ub1.getSecondsForPrefix(cd)
	expected := 21.0
	if credit != 0 || seconds != expected || len(bucketList) < 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBalanceStorageStore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	storageGetter.SetUserBalance(rifsBalance)
	result, err := storageGetter.GetUserBalance(rifsBalance.Id)
	if err != nil || rifsBalance.Id != result.Id ||
		len(rifsBalance.BalanceMap[MINUTES+OUTBOUND]) < 2 || len(result.BalanceMap[MINUTES+OUTBOUND]) < 2 ||
		!(rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Equal(result.BalanceMap[MINUTES+OUTBOUND][0])) ||
		!(rifsBalance.BalanceMap[MINUTES+OUTBOUND][1].Equal(result.BalanceMap[MINUTES+OUTBOUND][1])) ||
		!rifsBalance.BalanceMap[CREDIT+OUTBOUND].Equal(result.BalanceMap[CREDIT+OUTBOUND]) {
		t.Errorf("Expected %v was %v", rifsBalance, result)
	}
}

func TestDebitMoneyBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	result := rifsBalance.debitGenericBalance(CREDIT, 6, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 15 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 15, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
}

func TestDebitAllMoneyBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	rifsBalance.debitGenericBalance(CREDIT, 21, false)
	result := rifsBalance.debitGenericBalance(CREDIT, 0, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 0 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
}

func TestDebitMoreMoneyBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	result := rifsBalance.debitGenericBalance(CREDIT, 22, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != -1 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", -1, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
}

func TestDebitNegativeMoneyBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	result := rifsBalance.debitGenericBalance(CREDIT, -15, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 36 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 36, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationId: "NAT", RateSubject: ZEROSECOND}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Value: 21}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
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
	b1 := &Balance{Uuid: "testm", Value: 70, Weight: 5, DestinationId: "NAT", RateSubject: ZEROMINUTE}
	b2 := &Balance{Uuid: "tests", Value: 10, Weight: 10, DestinationId: "NAT", RateSubject: ZEROSECOND}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1, b2},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Value: 21}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "tests" ||
		cc.Timespans[1].Increments[0].BalanceUuids[0] != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0], cc.Timespans[1].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][1].Value != 0 ||
		rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 10 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditHasCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
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
	b1 := &Balance{Uuid: "testb", Value: 10, Weight: 10, DestinationId: "NAT", RateSubject: ZEROSECOND}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				CallDuration: 0,
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
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
	b1 := &Balance{Uuid: "testb", Value: 150, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 30 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0])
	}
}

func TestDebitCreditMoreTimespansMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
	b2 := &Balance{Uuid: "testa", Value: 150, Weight: 5, DestinationId: "NAT", RateSubject: ZEROSECOND}
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1, b2},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
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
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: ZEROMINUTE}
	cc := &CallCost{
		Direction:   OUTBOUND,
		Destination: "0723045326",
		ConnectFee:  10.0,
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
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}

	if len(cc.Timespans) != 0 || cc.GetDuration() != 0 {
		t.Error("Error cutting at no connect fee: ", cc.Timespans)
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
				RateInterval: &RateInterval{Rating: &RIRate{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "money", Value: 50}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}

	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "money" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 0 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][0])
	}
	if len(cc.Timespans) != 2 ||
		cc.Timespans[0].GetDuration() != 10*time.Second ||
		cc.Timespans[1].GetDuration() != 40*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans[1].Increments[0])
	}
}

func TestDebitCreditSubjectMinutes(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 250, Weight: 10, DestinationId: "NAT", RateSubject: "minu"}
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
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 350}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
		cc.Timespans[0].Increments[0].BalanceUuids[1] != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 180 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 280 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
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
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		CREDIT + OUTBOUND: BalanceChain{&Balance{Uuid: "moneya", Value: 75, DestinationId: "NAT", RateSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 5 {
		t.Errorf("Error extracting minutes from balance: %+v",
			rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMixed(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 40, Weight: 10, DestinationId: "NAT", RateSubject: "minu"}
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
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 150, RateSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
		cc.Timespans[0].Increments[0].BalanceUuids[1] != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value != 0 ||
		rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value != 80 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[MINUTES+OUTBOUND][0].Value, rifsBalance.BalanceMap[CREDIT+OUTBOUND][0].Value)
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 40*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans[0].GetDuration())
	}
}

/*
func TestDebitCreditSubjectMixedMoreTS(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70, Weight: 10, DestinationId: "NAT", RateSubject: "minu"}
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
				RateInterval: &RateInterval{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 100, RateIncrement: 10 * time.Second, RateUnit: time.Second}}},
			},
			&TimeSpan{
				TimeStart:    time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:      time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				CallDuration: 10 * time.Second,
				RateInterval: &RateInterval{Rates: RateGroups{&Rate{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}},
			},
		},
	}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{
		MINUTES + OUTBOUND: BalanceChain{b1},
		CREDIT + OUTBOUND:  BalanceChain{&Balance{Uuid: "moneya", Value: 50, RateSubject: "minu"}},
	}}
	err := rifsBalance.debitCreditBalance(cc, false)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	t.Errorf("%+v %+v", cc.Timespans[0], cc.Timespans[1])
	if cc.Timespans[0].Increments[0].BalanceUuids[0] != "testb" ||
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
*/

func TestDebitSMSBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}, SMS + OUTBOUND: BalanceChain{&Balance{Value: 100}}}}
	result := rifsBalance.debitGenericBalance(SMS, 12, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value != 88 || result != rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 88, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestDebitAllSMSBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}, SMS + OUTBOUND: BalanceChain{&Balance{Value: 100}}}}
	result := rifsBalance.debitGenericBalance(SMS, 100, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value != 0 || result != rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestDebitMoreSMSBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}, SMS + OUTBOUND: BalanceChain{&Balance{Value: 100}}}}
	result := rifsBalance.debitGenericBalance(SMS, 110, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value != -10 || result != rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", -10, rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value)
	}
}

func TestDebitNegativeSMSBalance(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}, SMS + OUTBOUND: BalanceChain{&Balance{Value: 100}}}}
	result := rifsBalance.debitGenericBalance(SMS, -15, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value != 115 || result != rifsBalance.BalanceMap[SMS+OUTBOUND][0].Value {
		t.Errorf("Expected %v was %v", 115, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestUserBalancedebitBalance(t *testing.T) {
	ub := &UserBalance{
		Id:         "rif",
		Type:       UB_TYPE_POSTPAID,
		BalanceMap: map[string]BalanceChain{SMS: BalanceChain{&Balance{Value: 14}}, TRAFFIC: BalanceChain{&Balance{Value: 1204}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	newMb := &Balance{Weight: 20, DestinationId: "NEW"}
	a := &Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: newMb}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 3 || ub.BalanceMap[MINUTES+OUTBOUND][2] != newMb {
		t.Error("Error adding minute bucket!", len(ub.BalanceMap[MINUTES+OUTBOUND]), ub.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestUserBalancedebitBalanceExists(t *testing.T) {

	ub := &UserBalance{
		Id:         "rif",
		Type:       UB_TYPE_POSTPAID,
		BalanceMap: map[string]BalanceChain{SMS + OUTBOUND: BalanceChain{&Balance{Value: 14}}, TRAFFIC + OUTBOUND: BalanceChain{&Balance{Value: 1024}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 15, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	newMb := &Balance{Value: -10, Weight: 20, DestinationId: "NAT"}
	a := &Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: newMb}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUserBalanceAddMinuteNil(t *testing.T) {
	ub := &UserBalance{
		Id:         "rif",
		Type:       UB_TYPE_POSTPAID,
		BalanceMap: map[string]BalanceChain{SMS + OUTBOUND: BalanceChain{&Balance{Value: 14}}, TRAFFIC + OUTBOUND: BalanceChain{&Balance{Value: 1024}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
	}
	ub.debitBalanceAction(nil)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUserBalanceAddMinutBucketEmpty(t *testing.T) {
	mb1 := &Balance{Value: -10, DestinationId: "NAT"}
	mb2 := &Balance{Value: -10, DestinationId: "NAT"}
	mb3 := &Balance{Value: -10, DestinationId: "OTHER"}
	ub := &UserBalance{}
	a := &Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: mb1}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
	a = &Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: mb2}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 1 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
	a = &Action{BalanceId: MINUTES, Direction: OUTBOUND, Balance: mb3}
	ub.debitBalanceAction(a)
	if len(ub.BalanceMap[MINUTES+OUTBOUND]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[MINUTES+OUTBOUND])
	}
}

func TestUserBalanceExecuteTriggeredActions(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 10, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	// are set to executed
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	// we can reset them
	ub.resetActionTriggers(nil)
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 120 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 30 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
}

func TestUserBalanceExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}, MINUTES + OUTBOUND: BalanceChain{&Balance{Value: 10, Weight: 20, DestinationId: "NAT"}, &Balance{Weight: 10, DestinationId: "RET"}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 100, ThresholdType: TRIGGER_MIN_COUNTER, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Balance: &Balance{Value: 1}})
	if ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 110 || ub.BalanceMap[MINUTES+OUTBOUND][0].Value != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND][0].Value, ub.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
}

func TestUserBalanceExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB_OREDER",
		BalanceMap:     map[string]BalanceChain{CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 100}}},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ThresholdType: TRIGGER_MAX_COUNTER, ActionsId: "TEST_ACTIONS_ORDER"}},
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 1}})
	if len(ub.BalanceMap[CREDIT+OUTBOUND]) != 1 || ub.BalanceMap[CREDIT+OUTBOUND][0].Value != 10 {
		t.Error("Error executing triggered actions in order", ub.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &UserBalance{
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

func TestUserBalanceUnitCounting(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
}

func TestUserBalanceUnitCountingOutbound(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 30 {
		t.Error("Error counting units")
	}
}

func TestUserBalanceUnitCountingOutboundInbound(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: INBOUND, Balance: &Balance{Value: 10}})
	if len(ub.UnitCounters) != 2 && ub.UnitCounters[1].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 || ub.UnitCounters[1].Units != 10 {
		t.Error("Error counting units")
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}

	ub1 := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(cd)
	}
}

func BenchmarkUserBalanceStorageStoreRestore(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	for i := 0; i < b.N; i++ {
		storageGetter.SetUserBalance(rifsBalance)
		storageGetter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &Balance{Value: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b1, b2}, CREDIT + OUTBOUND: BalanceChain{&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(cd)
	}
}
