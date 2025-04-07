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
	"bytes"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
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
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1, b2},
			utils.MetaMonetary: {&Balance{Value: 21}}}}
	dm.SetAccount(rifsBalance)
	ub1, err := dm.GetAccount("other")
	if err != nil ||
		!ub1.BalanceMap[utils.MetaMonetary].Equal(rifsBalance.BalanceMap[utils.MetaMonetary]) {
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
			utils.MetaVoice:    {b1, b2},
			utils.MetaMonetary: {&Balance{Value: 200}}}}
	cd := &CallDescriptor{
		Category:      "0",
		Tenant:        "vdf",
		TimeStart:     time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 10 * time.Second,
		Destination:   "0723",
		ToR:           utils.MetaVoice,
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
			utils.MetaVoice:    {b1, b2},
			utils.MetaMonetary: {&Balance{Value: 21}},
		},
	}
	cd := &CallDescriptor{
		Category:    "0",
		Tenant:      "vdf",
		TimeStart:   time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 4, 15, 46, 60, 0, time.UTC),
		LoopIndex:   0,
		Destination: "0723",
		ToR:         utils.MetaVoice,
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 20 * time.Second
	if credit != 0 || seconds != expected ||
		len(bucketList) != 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestAccountStorageStore(t *testing.T) {
	b1 := &Balance{Value: 10, Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1, b2},
			utils.MetaMonetary: {&Balance{Value: 21}}}}
	dm.SetAccount(rifsBalance)
	result, err := dm.GetAccount(rifsBalance.ID)
	if err != nil || rifsBalance.ID != result.ID ||
		len(rifsBalance.BalanceMap[utils.MetaVoice]) < 2 ||
		len(result.BalanceMap[utils.MetaVoice]) < 2 ||
		!(rifsBalance.BalanceMap[utils.MetaVoice][0].Equal(result.BalanceMap[utils.MetaVoice][0])) ||
		!(rifsBalance.BalanceMap[utils.MetaVoice][1].Equal(result.BalanceMap[utils.MetaVoice][1])) ||
		!rifsBalance.BalanceMap[utils.MetaMonetary].Equal(result.BalanceMap[utils.MetaMonetary]) {
		t.Errorf("Expected %s was %s", utils.ToIJSON(rifsBalance), utils.ToIJSON(result))
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{
		Uuid: "testb", Value: 10 * float64(time.Second), Weight: 10,
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "*zero1s"}
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{Rating: &RIRate{
					Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
						Value: 100, RateIncrement: 10 * time.Second,
						RateUnit: time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Destination:  "0723045326",
		Category:     "0",
		ToR:          utils.MetaVoice,
		testCallcost: cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {b1},
			utils.MetaMonetary: {&Balance{
				Categories: utils.NewStringMap("0"), Value: 21}}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Fatal("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" {
		t.Logf("%+v", cc.Timespans[0])
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.MetaVoice][0])
	}
}

func TestDebitCreditBlocker(t *testing.T) {
	b1 := &Balance{Uuid: "testa", Value: 0.1152,
		Weight: 20, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "passmonde", Blocker: true}
	b2 := &Balance{Uuid: utils.MetaDefault, Value: 1.5, Weight: 0}
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{ConnectFee: 0.15,
						Rates: RateGroups{&RGRate{GroupIntervalStart: 0,
							Value: 0.1, RateIncrement: time.Second,
							RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
		ToR:              utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Destination:  "0723045326",
		Category:     "0",
		ToR:          utils.MetaVoice,
		testCallcost: cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{utils.MetaMonetary: {b1, b2}}}
	expCC := &CallCost{
		Category:         "0",
		Destination:      "0723045326",
		ToR:              "*voice",
		deductConnectFee: true,
	}
	if cc, err := rifsBalance.debitCreditBalance(cd, false, true, true, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expCC, cc) {
		t.Errorf("expected <%+v>, \nreceived <%+v>", expCC, cc)
	}

	if rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 0.1152 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][1].GetValue() != 1.5 {
		t.Error("should not have touched the balances: ",
			utils.ToJSON(rifsBalance.BalanceMap[utils.MetaMonetary]))
	}
}

func TestDebitFreeEmpty(t *testing.T) {
	cc := &CallCost{
		Destination: "112",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{Rating: &RIRate{
					ConnectFee: 0, Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 0,
							RateIncrement: time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		deductConnectFee: true,
		ToR:              utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Tenant:       "CUSTOMER_1",
		Subject:      "rif:from:tm",
		Destination:  "112",
		Category:     "0",
		ToR:          utils.MetaVoice,
		testCallcost: cc,
	}
	// empty account
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{utils.MetaMonetary: {}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, true, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if len(cc.Timespans) == 0 || cc.Cost != 0 {
		t.Error("Wrong call cost: ", utils.ToIJSON(cc))
	}
	if len(rifsBalance.BalanceMap[utils.MetaMonetary]) != 0 {
		t.Error("should not have touched the balances: ",
			utils.ToIJSON(rifsBalance.BalanceMap[utils.MetaMonetary]))
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject: "*zero1m"}
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Destination:  "0723045326",
		Category:     "0",
		ToR:          utils.MetaVoice,
		testCallcost: cc,
	}
	rifsBalance := &Account{
		ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	//t.Logf("%+v", cc.Timespans)
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Errorf("Error setting balance id to increment: %s",
			utils.ToJSON(cc))
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Errorf("Error extracting minutes from balance: %s",
			utils.ToJSON(rifsBalance.BalanceMap[utils.MetaVoice][0]))
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0,
							Value: 100, RateIncrement: 10 * time.Second,
							RateUnit: time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.Timespans[0].GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1, b2},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "tests" ||
		cc.Timespans[1].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans)
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][1].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Logf("TS0: %+v", cc.Timespans[0])
		t.Logf("TS1: %+v", cc.Timespans[1])
		t.Errorf("Error extracting minutes from balance: %+v", rifsBalance.BalanceMap[utils.MetaVoice][1])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		RatingSubject:  "*zero1m", Weight: 10}
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err == nil {
		t.Error("Showing no enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 10*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaVoice][0])
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0,
							Value:         1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{
		utils.MetaVoice:    {b1},
		utils.MetaMonetary: {{Uuid: "moneya", Value: 110}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				DurationIndex: 0,
				ratingInfo:    &RatingInfo{},
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1},
			utils.MetaMonetary: {&Balance{Uuid: "moneya", Value: 50}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Second {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0].Duration)
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 30 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 30*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaVoice][0])
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0,
							Value:         100,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{
		ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {b1, b2},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].Duration != time.Minute {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 10*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MetaVoice][1].GetValue() != 130*float64(time.Second) {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaVoice][1], cc.Timespans[1])
	}
}

func TestDebitCreditNoConectFeeCredit(t *testing.T) {
	b1 := &Balance{Uuid: "testb", Value: 70 * float64(time.Second),
		DestinationIDs: utils.StringMap{"NAT": true},
		Weight:         10, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{ConnectFee: 10.0,
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR:              utils.MetaVoice,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {b1},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}
	if len(cc.Timespans) != 1 ||
		rifsBalance.BalanceMap[utils.MetaMonetary].GetTotalValue() != 0 {
		t.Error("Error cutting at no connect fee: ",
			rifsBalance.BalanceMap[utils.MetaMonetary])
	}
}

func TestDebitCreditMoneyOnly(t *testing.T) {
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				ratingInfo:    &RatingInfo{},
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Uuid: "money", Value: 50}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err == nil {
		t.Error("Missing noy enough credit error ")
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "money" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Logf("%+v", cc.Timespans[0].Increments)
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0].BalanceInfo)
	}
	if rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 0 {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaMonetary][0])
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0,
							Value:         1,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		ToR:              utils.MetaVoice,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      "0",
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaVoice:    {b1},
			utils.MetaMonetary: {&Balance{Uuid: "moneya", Value: 350}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Errorf("Error setting balance id to increment: %+v",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue() != 180*float64(time.Second) ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 280 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.MetaVoice][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 10 * time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR:              utils.MetaVoice,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Uuid: "moneya", Value: 75,
					DestinationIDs: utils.StringMap{"NAT": true},
					RatingSubject:  "minu"}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments[0].Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 5 {
		t.Errorf("Error extracting minutes from balance: %+v",
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
	if len(cc.Timespans) != 1 ||
		cc.Timespans[0].GetDuration() != 70*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestAccountdebitBalance(t *testing.T) {
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaSMS:  {&Balance{Value: 14}},
			utils.MetaData: {&Balance{Value: 1204}},
			utils.MetaVoice: {
				&Balance{Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	newMb := &BalanceFilter{
		Type:           utils.StringPointer(utils.MetaVoice),
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NEW": true}),
	}
	a := &Action{Balance: newMb}
	ub.debitBalanceAction(a, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 3 ||
		!ub.BalanceMap[utils.MetaVoice][2].DestinationIDs.Equal(*newMb.DestinationIDs) {
		t.Errorf("Error adding minute bucket! %d %+v %+v",
			len(ub.BalanceMap[utils.MetaVoice]), ub.BalanceMap[utils.MetaVoice][2], newMb)
	}
}

func TestAccountdebitBalanceExists(t *testing.T) {
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaSMS:  {&Balance{Value: 14}},
			utils.MetaData: {&Balance{Value: 1024}},
			utils.MetaVoice: {
				&Balance{
					Value: 15, Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	newMb := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.MetaVoice),
		Weight:         utils.Float64Pointer(20),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
	}
	a := &Action{Balance: newMb}
	ub.debitBalanceAction(a, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 2 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinuteNil(t *testing.T) {
	ub := &Account{
		ID:            "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaSMS:  {&Balance{Value: 14}},
			utils.MetaData: {&Balance{Value: 1024}},
			utils.MetaVoice: {
				&Balance{Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	ub.debitBalanceAction(nil, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinutBucketEmpty(t *testing.T) {
	mb1 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.MetaVoice),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
	}
	mb2 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.MetaVoice),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"NAT": true}),
	}
	mb3 := &BalanceFilter{
		Value:          &utils.ValueFormula{Static: -10},
		Type:           utils.StringPointer(utils.MetaVoice),
		DestinationIDs: utils.StringMapPointer(utils.StringMap{"OTHER": true}),
	}
	ub := &Account{}
	a := &Action{Balance: mb1}
	ub.debitBalanceAction(a, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.MetaVoice])
	}
	a = &Action{Balance: mb2}
	ub.debitBalanceAction(a, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 1 || ub.BalanceMap[utils.MetaVoice][0].GetValue() != 20 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.MetaVoice])
	}
	a = &Action{Balance: mb3}
	ub.debitBalanceAction(a, false, false, nil)
	if len(ub.BalanceMap[utils.MetaVoice]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.MetaVoice])
	}
}

func TestAccountExecuteTriggeredActions(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Value: 1,
						Filter: &BalanceFilter{
							Type: utils.StringPointer(utils.MetaMonetary)}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2, ThresholdType: utils.TriggerMaxEventCounter,
				ActionsID: "TEST_ACTIONS"}},
	}
	ub.countUnits(1, utils.MetaMonetary, new(CallCost), nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 110 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(),
			ub.BalanceMap[utils.MetaVoice][0].GetValue())
	}
	// are set to executed
	ub.countUnits(1, utils.MetaMonetary, nil, nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 110 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(), ub.BalanceMap[utils.MetaVoice][0].GetValue())
	}
	// we can reset them
	ub.ResetActionTriggers(nil, nil)
	ub.countUnits(10, utils.MetaMonetary, nil, nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 120 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 30*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(),
			ub.BalanceMap[utils.MetaVoice][0].GetValue())
	}
}

func TestAccountExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value: 100}},
			utils.MetaVoice: {
				&Balance{
					Value:          10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{
					Weight:         10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Filter: &BalanceFilter{
						Type: utils.StringPointer(utils.MetaMonetary)},
						Value: 1.0}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMinEventCounter,
				ActionsID:      "TEST_ACTIONS"}},
	}
	ub.countUnits(1, utils.MetaMonetary, nil, nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 110 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 20*float64(time.Second) {
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(),
			ub.BalanceMap[utils.MetaVoice][0].GetValue(),
			len(ub.BalanceMap[utils.MetaMonetary]))
	}
}

func TestAccountExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB_OREDER",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Value: 1,
						Filter: &BalanceFilter{
							Type: utils.StringPointer(utils.MetaMonetary)}}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 2,
				ThresholdType:  utils.TriggerMaxEventCounter,
				ActionsID:      "TEST_ACTIONS_ORDER"}},
	}

	ub.countUnits(1, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.BalanceMap[utils.MetaMonetary]) != 1 ||
		ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 10 {

		t.Errorf("Error executing triggered actions in order %v",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestAccountExecuteTriggeredDayWeek(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{UniqueID: "day_trigger",
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 10, ThresholdType: utils.TriggerMaxEventCounter,
				ActionsID: "TEST_ACTIONS"},
			&ActionTrigger{UniqueID: "week_trigger",
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 100, ThresholdType: utils.TriggerMaxEventCounter,
				ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.InitCounters()
	if len(ub.UnitCounters) != 1 || len(ub.UnitCounters[utils.MetaMonetary][0].Counters) != 2 {
		t.Error("Error initializing counters: ", ub.UnitCounters[utils.MetaMonetary][0].Counters[0])
	}

	ub.countUnits(1, utils.MetaMonetary, new(CallCost), nil, nil)
	if ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 1 {
		t.Error("Error incrementing both counters",
			ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value,
			ub.UnitCounters[utils.MetaMonetary][0].Counters[1].Value)
	}

	// we can reset them
	resetCountersAction(ub, &Action{
		Balance: &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary),
			ID: utils.StringPointer("day_trigger")}}, nil, nil, nil, SharedActionsData{}, ActionConnCfg{})
	if ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 0 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[1].Value != 1 {
		t.Error("Error reseting both counters",
			ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value,
			ub.UnitCounters[utils.MetaMonetary][0].Counters[1].Value)
	}
}

func TestAccountExpActionTrigger(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{Value: 100,
					ExpirationDate: time.Date(2015, time.November, 9, 9, 48, 0, 0, time.UTC)}},
			utils.MetaVoice: {
				&Balance{Value: 10 * float64(time.Second), Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10 * float64(time.Second),
					DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances", Balance: &BalanceFilter{
				Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 10, ThresholdType: utils.TriggerBalanceExpired,
				ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()) ||
		ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 10 || // expired was cleaned
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 20*float64(time.Second) ||
		ub.ActionTriggers[0].Executed != true {
		t.Log(ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()))
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(),
			ub.BalanceMap[utils.MetaVoice][0].GetValue(),
			len(ub.BalanceMap[utils.MetaMonetary]))
	}
}

func TestAccountExpActionTriggerNotActivated(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {
				&Balance{Value: 10, Weight: 20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances",
				ActivationDate: time.Date(2116, 2, 5, 18, 0, 0, 0, time.UTC),
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 10, ThresholdType: utils.TriggerBalanceExpired,
				ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()) ||
		ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 100 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 10 ||
		ub.ActionTriggers[0].Executed != false {
		t.Log(ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()))
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MetaMonetary][0].GetValue(), ub.BalanceMap[utils.MetaVoice][0].GetValue(), len(ub.BalanceMap[utils.MetaMonetary]))
	}
}

func TestAccountExpActionTriggerExpired(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 100}},
			utils.MetaVoice: {&Balance{Value: 10, Weight: 20,
				DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{ID: "check expired balances",
				ExpirationDate: time.Date(2016, 2, 4, 18, 0, 0, 0, time.UTC),
				Balance:        &BalanceFilter{Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 10, ThresholdType: utils.TriggerBalanceExpired,
				ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.ExecuteActionTriggers(nil, nil)
	if ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()) ||
		ub.BalanceMap[utils.MetaMonetary][0].GetValue() != 100 ||
		ub.BalanceMap[utils.MetaVoice][0].GetValue() != 10 ||
		len(ub.ActionTriggers) != 0 {
		t.Log(ub.BalanceMap[utils.MetaMonetary][0].IsExpiredAt(time.Now()))
		t.Error("Error executing triggered actions",
			ub.BalanceMap[utils.MetaMonetary][0].GetValue(),
			ub.BalanceMap[utils.MetaVoice][0].GetValue(),
			len(ub.BalanceMap[utils.MetaMonetary]))
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &Account{
		ID: "TEST_UB_OREDER",
		BalanceMap: map[string]Balances{utils.MetaMonetary: {
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)}}, utils.MetaVoice: {
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
	if len(ub.BalanceMap[utils.MetaMonetary]) != 2 {
		t.Error("Error cleaning expired balances!")
	}
	if len(ub.BalanceMap[utils.MetaVoice]) != 1 {
		t.Error("Error cleaning expired minute buckets!")
	}
	if len(ub.ActionTriggers) != 1 {
		t.Error("Error cleaning expired action triggers!")
	}
}

func TestAccountUnitCounting(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{
		utils.MetaMonetary: []*UnitCounter{{
			Counters: CounterFilters{&CounterFilter{Value: 0}}}}}}
	ub.countUnits(10, utils.MetaMonetary, &CallCost{}, nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MetaMonetary, &CallCost{}, nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MetaMonetary: []*UnitCounter{
		{Counters: CounterFilters{&CounterFilter{Value: 0}}}}}}
	ub.countUnits(10, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(10, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 30 {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutboundInbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{
		utils.MetaMonetary: []*UnitCounter{
			{Counters: CounterFilters{&CounterFilter{Value: 0}}}}}}
	ub.countUnits(10, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 10 {
		t.Errorf("Error counting units: %+v",
			ub.UnitCounters[utils.MetaMonetary][0].Counters[0])
	}
	ub.countUnits(10, utils.MetaMonetary, new(CallCost), nil, nil)
	if len(ub.UnitCounters[utils.MetaMonetary]) != 1 ||
		ub.UnitCounters[utils.MetaMonetary][0].Counters[0].Value != 20 {
		t.Error("Error counting units")
	}
}

func TestDebitShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval: &RateInterval{Rating: &RIRate{
					Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 2,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MetaMonetary: {&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{ID: "groupie", BalanceMap: map[string]Balances{
		utils.MetaMonetary: {&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: utils.NewStringMap(rif.ID, groupie.ID), AccountParameters: map[string]*SharingParameters{"*any": {Strategy: STRATEGY_MINE_RANDOM}}}

	dm.SetAccount(groupie)
	dm.SetSharedGroup(sg)
	cc, err := rif.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rif.BalanceMap[utils.MetaMonetary][0].GetValue() != 0 {
		t.Errorf("Error debiting from shared group: %+v", rif.BalanceMap[utils.MetaMonetary][0])
	}
	groupie, _ = dm.GetAccount("groupie")
	if groupie.BalanceMap[utils.MetaMonetary][0].GetValue() != 10 {
		t.Errorf("Error debiting from shared group: %+v", groupie.BalanceMap[utils.MetaMonetary][0])
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
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval: &RateInterval{
					Rating: &RIRate{Rates: RateGroups{
						&RGRate{GroupIntervalStart: 0, Value: 2,
							RateIncrement: 10 * time.Second,
							RateUnit:      time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MetaMonetary: {&Balance{Uuid: "moneya", Value: 0, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{ID: "groupie", BalanceMap: map[string]Balances{
		utils.MetaMonetary: {&Balance{Uuid: "moneyc", Value: 130, SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Id: "SG_TEST", MemberIds: utils.NewStringMap(rif.ID, groupie.ID), AccountParameters: map[string]*SharingParameters{"*any": {Strategy: STRATEGY_MINE_RANDOM}}}

	dm.SetAccount(groupie)
	dm.SetSharedGroup(sg)
	duration, err := cd.getMaxSessionDuration(rif, nil)
	if err != nil {
		t.Error("Error getting max session duration from shared group: ", err)
	}
	if duration != time.Minute {
		t.Error("Wrong max session from shared group: ", duration)
	}

}

func TestMaxDurationConnectFeeOnly(t *testing.T) {
	cd := &CallDescriptor{
		Tenant:        "cgrates.org",
		Category:      "call",
		TimeStart:     time.Date(2015, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:       time.Date(2015, 9, 24, 10, 58, 1, 0, time.UTC),
		Destination:   "4444",
		Subject:       "dy",
		Account:       "dy",
		ToR:           utils.MetaVoice,
		DurationIndex: 600,
	}
	rif := &Account{ID: "rif", BalanceMap: map[string]Balances{
		utils.MetaMonetary: {&Balance{Uuid: "moneya", Value: 0.2}},
	}}

	duration, err := cd.getMaxSessionDuration(rif, nil)
	if err != nil {
		t.Error("Error getting max session duration: ", err)
	}
	if duration != 0 {
		t.Error("Wrong max session: ", duration)
	}

}

func TestDebitSMS(t *testing.T) {
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 1, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: 1,
								RateUnit:      time.Nanosecond}}}},
			},
		},
		ToR: utils.MetaSMS,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaSMS: {
				&Balance{Uuid: "testm",
					Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MetaMonetary: {
				&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaSMS][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaSMS][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestDebitGeneric(t *testing.T) {
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 1, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{
								GroupIntervalStart: 0,
								Value:              100,
								RateIncrement:      1,
								RateUnit:           time.Nanosecond,
							},
						},
					},
				},
			},
		},
		ToR: utils.MetaGeneric,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaGeneric: {
				&Balance{Uuid: "testm", Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ",
			cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue() != 99 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestDebitGenericBalance(t *testing.T) {
	cc := &CallCost{
		Category:    "call",
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value:         100,
								RateIncrement: time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		Category:      "call",
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
		ExtraFields: map[string]string{
			utils.BalanceFactorID: "call",
		},
	}
	rifsBalance := &Account{
		ID: "other", BalanceMap: map[string]Balances{
			utils.MetaGeneric: {
				&Balance{Uuid: "testm", Value: 100, Weight: 5,
					DestinationIDs: utils.StringMap{"NAT": true},
					Factors:        ValueFactors{"call": 1 / (60 * float64(time.Second))}}},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue() != 99.49999 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Logf("%+v", cc.Timespans[0].Increments[0])
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestDebitGenericBalanceWithRatingSubject(t *testing.T) {
	cc := &CallCost{
		Category:    "call",
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 0,
								RateIncrement: time.Second,
								RateUnit:      time.Second}}}},
			},
		},
		ToR: utils.MetaVoice,
	}
	cd := &CallDescriptor{
		Category:      "call",
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
		ExtraFields: map[string]string{
			utils.BalanceFactorID: "call",
		},
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaGeneric: {
				&Balance{Uuid: "testm", Value: 100,
					Weight: 5, DestinationIDs: utils.StringMap{"NAT": true},
					Factors:       ValueFactors{"call": 1 / (60 * float64(time.Second))},
					RatingSubject: "free"}},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0])
	}
	if rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue() != 99.49999 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Logf("%+v", cc.Timespans[0].Increments[0])
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaGeneric][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestDebitDataUnits(t *testing.T) {
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 0, 80, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0,
								Value: 2, RateIncrement: 1,
								RateUnit: 1},
							&RGRate{GroupIntervalStart: 60,
								Value:         1,
								RateIncrement: 1,
								RateUnit:      1},
						},
					},
				},
			},
		},
		ToR: utils.MetaData,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other",
		BalanceMap: map[string]Balances{
			utils.MetaData: {
				&Balance{Uuid: "testm", Value: 100,
					Weight:         5,
					DestinationIDs: utils.StringMap{"NAT": true}}},
			utils.MetaMonetary: {&Balance{Value: 21}},
		}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true, nil)
	// test rating information
	ts := cc.Timespans[0]
	if ts.MatchedSubject != "testm" || ts.MatchedPrefix != "0723" ||
		ts.MatchedDestId != "NAT" || ts.RatingPlanId != utils.MetaNone {
		t.Errorf("Error setting rating info: %+v", ts.ratingInfo)
	}
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if ts.Increments[0].BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", ts.Increments[0])
	}
	if rifsBalance.BalanceMap[utils.MetaData][0].GetValue() != 20 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 21 {
		t.Log(ts.Increments)
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MetaData][0].GetValue(),
			rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
	}
}

func TestDebitDataMoney(t *testing.T) {
	cc := &CallCost{
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RGRate{GroupIntervalStart: 0, Value: 2, RateIncrement: time.Minute, RateUnit: time.Second},
						},
					},
				},
			},
		},
		ToR: utils.MetaData,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Destination:   cc.Destination,
		ToR:           cc.ToR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{
		utils.MetaData:     {&Balance{Uuid: "testm", Value: 0, Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MetaMonetary: {&Balance{Value: 160}},
	}}
	if _, err := rifsBalance.debitCreditBalance(cd, false, false, true, nil); err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rifsBalance.BalanceMap[utils.MetaData][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue() != 0 {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.MetaData][0].GetValue(), rifsBalance.BalanceMap[utils.MetaMonetary][0].GetValue())
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
	tag := utils.MetaMonetary
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
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		len(a.UnitCounters[utils.MetaVoice][0].Counters) != 1 ||
		len(a.UnitCounters[utils.MetaVoice][1].Counters) != 1 ||
		len(a.UnitCounters[utils.MetaSMS][0].Counters) != 1 {
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
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaMonetary),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TriggerMaxEventCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaVoice),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TriggerMaxBalanceCounter,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TriggerMaxBalance,
				Balance: &BalanceFilter{
					Type:   utils.StringPointer(utils.MetaSMS),
					Weight: utils.Float64Pointer(10),
				},
			},
		},
	}
	a.InitCounters()
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MetaMonetary][0].Counters) != 2 ||
		len(a.UnitCounters[utils.MetaVoice][0].Counters) != 1 ||
		len(a.UnitCounters[utils.MetaVoice][1].Counters) != 1 ||
		len(a.UnitCounters[utils.MetaSMS][0].Counters) != 1 {
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
			utils.MetaMonetary: {
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) != 0 {
		t.Error("error excluding on mixed balances")
	}
}

func TestAccountGetBalancesForPrefixAllExcl(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": false, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) == 0 {
		t.Error("error finding balance on all excluded")
	}
}

func TestAccountGetBalancesForPrefixMixedGood(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": true},
				},
			},
		},
	}

	bcs := acc.getBalancesForPrefix("999123", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) == 0 {
		t.Error("error finding on mixed balances good")
	}
}

func TestAccountGetBalancesForPrefixMixedBad(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value:          10,
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.MetaMonetary, "", time.Now())
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
			utils.MetaSMS:  {&Balance{ID: "sms1", Value: 14}},
			utils.MetaMMS:  {&Balance{ID: "mms1", Value: 140}},
			utils.MetaData: {&Balance{ID: "data1", Value: 1204}},
			utils.MetaVoice: {
				&Balance{ID: "voice1", Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Value: 3600},
				&Balance{ID: "voice2", Weight: 10, DestinationIDs: utils.StringMap{"RET": true}, Value: 1200},
			},
		},
	}
	expectacntSummary := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "account1",
		BalanceSummaries: []*BalanceSummary{
			{ID: "data1", Type: utils.MetaData, Value: 1204, Disabled: false},
			{ID: "sms1", Type: utils.MetaSMS, Value: 14, Disabled: false},
			{ID: "mms1", Type: utils.MetaMMS, Value: 140, Disabled: false},
			{ID: "voice1", Type: utils.MetaVoice, Weight: 20, Value: 3600, Disabled: false},
			{ID: "voice2", Type: utils.MetaVoice, Weight: 10, Value: 1200, Disabled: false},
		},
		AllowNegative: true,
		Disabled:      false,
	}
	acntSummary := acnt1.AsAccountSummary()
	// Since maps are unordered, slices will be too so we need to find element to compare
	if !reflect.DeepEqual(expectacntSummary, acntSummary) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectacntSummary), utils.ToJSON(acntSummary))
	}
}

func TestAccountGetBalancesGetBalanceWithSameWeight(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:     "SpecialBalance1",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance2",
					Value:  10,
					Weight: 10.0,
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) != 2 && bcs[0].ID != "SpecialBalance1" && bcs[1].ID != "SpecialBalance2" {
		t.Errorf("Unexpected order balances : %+v", utils.ToJSON(bcs))
	}
}

func TestAccountGetBalancesForPrefix2(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:     "SpecialBalance1",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance2",
					Value:  10,
					Weight: 20.0,
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) != 2 && bcs[0].ID != "SpecialBalance2" && bcs[0].Weight != 20.0 {
		t.Errorf("Unexpected order balances : %+v", utils.ToJSON(bcs))
	}
}

func TestAccountGetMultipleBalancesForPrefixWithSameWeight(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:     "SpecialBalance1",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance2",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance3",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance4",
					Value:  10,
					Weight: 10.0,
				},
				&Balance{
					ID:     "SpecialBalance5",
					Value:  10,
					Weight: 10.0,
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("", "", utils.MetaMonetary, "", time.Now())
	if len(bcs) != 5 &&
		bcs[0].ID != "SpecialBalance1" && bcs[1].ID != "SpecialBalance2" &&
		bcs[2].ID != "SpecialBalance3" && bcs[3].ID != "SpecialBalance4" &&
		bcs[4].ID != "SpecialBalance5" {
		t.Errorf("Unexpected order balances : %+v", utils.ToJSON(bcs))
	}
}

func TestAccountClone(t *testing.T) {
	account := &Account{}
	eOut := &Account{}
	if rcv := account.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	account = &Account{
		ID: "testID",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 10, Weight: 10}}},
		ActionTriggers: []*ActionTrigger{
			{
				ID: "ActionTriggerID1",
			},
			{
				ID: "ActionTriggerID2",
			},
		},
		AllowNegative: true,
		Disabled:      true,
	}
	eOut = &Account{
		ID: "testID",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {&Balance{Value: 10, Weight: 10}}},
		ActionTriggers: []*ActionTrigger{
			{
				ID: "ActionTriggerID1",
			},
			{
				ID: "ActionTriggerID2",
			},
		},
		AllowNegative: true,
		Disabled:      true,
	}

	if rcv := account.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAccountGetBalanceWithID(t *testing.T) {
	account := &Account{
		BalanceMap: map[string]Balances{
			"type1": {&Balance{ID: "test1", Value: 0.7}},
			"type2": {&Balance{ID: "test2", Value: 0.8}},
		},
	}
	if rcv := account.GetBalanceWithID("type1", "test1"); rcv.Value != 0.7 {
		t.Errorf("Expecting: 0.7, received: %+v", rcv)
	}
	if rcv := account.GetBalanceWithID("type2", "test2"); rcv.Value != 0.8 {
		t.Errorf("Expecting: 0.8, received: %+v", rcv)
	}
	if rcv := account.GetBalanceWithID("unknown", "unknown"); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
}

func TestAccountSetRecurrent(t *testing.T) {
	aType := "testType"
	var ptrType *string = &aType
	acc := &Account{
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				ID:     "testID1",
				Weight: 10,
				Balance: &BalanceFilter{
					Type: ptrType,
				},
			},
			&ActionTrigger{
				ID:     "testID2",
				Weight: 10,
			},
		},
	}
	act := &Action{
		Id:     "testID",
		Weight: 10,
		Balance: &BalanceFilter{
			Type: ptrType,
		},
	}
	exp := ActionTriggers{
		&ActionTrigger{
			ID:        "testID1",
			Weight:    10,
			Recurrent: true,
			Balance: &BalanceFilter{
				Type: ptrType,
			},
		},
		&ActionTrigger{
			ID:     "testID2",
			Weight: 10,
		},
	}
	acc.SetRecurrent(act, true)
	rcv := acc.ActionTriggers

	if len(rcv) != len(exp) {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
	for i := range rcv {
		if rcv[i].Recurrent != exp[i].Recurrent {
			t.Fatalf("\nexpected: <%+v>,\nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
		}
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: 10, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}

	ub1 := &Account{ID: "other", BalanceMap: map[string]Balances{utils.MetaVoice: {b1, b2}, utils.MetaMonetary: {&Balance{Value: 21}}}}
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
	rifsBalance := &Account{ID: "other", BalanceMap: map[string]Balances{utils.MetaVoice: {b1, b2}, utils.MetaMonetary: {&Balance{Value: 21}}}}
	for i := 0; i < b.N; i++ {
		dm.SetAccount(rifsBalance)
		dm.GetAccount(rifsBalance.ID)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: 10, Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: 100, Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	ub1 := &Account{ID: "OUT:CUSTOMER_1:rif", BalanceMap: map[string]Balances{utils.MetaVoice: {b1, b2}, utils.MetaMonetary: {&Balance{Value: 21}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}

func TestAccountSummaryFieldAsInterface(t *testing.T) {
	as := AccountSummary{
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				UUID:    "uId",
				ID:      "id",
				Type:    "*data",
				Initial: 20.54,
				Value:   1,
			}},
		Tenant: "tenant",
		ID:     "accID",
	}

	if _, err := as.FieldAsInterface(nil); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := as.FieldAsInterface([]string{"test"}); err == nil {
		t.Error(err)
	}
	if val, err := as.FieldAsInterface([]string{"BalanceSummaries[0]"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, as.BalanceSummaries[0]) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(val), utils.ToJSON(as.BalanceSummaries[0]))
	} else if _, err = as.FieldAsInterface([]string{"Tenant"}); err != nil {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"Tenant", "Value"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"ID"}); err != nil {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"ID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if val, err := as.FieldAsInterface([]string{"BalanceSummaries"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, as.BalanceSummaries) {
		t.Errorf("expected %v ,received %v ", utils.ToJSON(val), utils.ToJSON(as.BalanceSummaries))
	}
	as.BalanceSummaries = BalanceSummaries{
		&BalanceSummary{

			UUID:    "uId",
			ID:      "id",
			Type:    "*data",
			Initial: 20.54,
			Value:   1,
		},
		&BalanceSummary{
			UUID:    "uId2",
			ID:      "id2",
			Type:    "*data",
			Initial: 20.54,
			Value:   1,
		}}
	if _, err := as.FieldAsInterface([]string{"BalanceSummaries", "id3"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)

	} else if val, err := as.FieldAsInterface([]string{"BalanceSummaries", "id2"}); err != nil {
		t.Error(err)

	} else if !reflect.DeepEqual(val, as.BalanceSummaries[1]) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(val), utils.ToJSON(as.BalanceSummaries[1]))
	}
	if _, err = as.FieldAsInterface([]string{"AllowNegative"}); err != nil {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"AllowNegative", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"Disabled"}); err != nil {
		t.Error(err)
	} else if _, err = as.FieldAsInterface([]string{"Disabled", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestAccountSummaryFieldAsString(t *testing.T) {
	as := AccountSummary{
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				UUID:    "uId",
				ID:      "id",
				Type:    "*data",
				Initial: 20.54,
				Value:   1,
			}},
	}
	rec := &BalanceSummary{}
	if _, err := as.FieldAsString([]string{}); err == nil {
		t.Error(err)
	}
	if val, err := as.FieldAsString([]string{"BalanceSummaries[0]"}); err != nil {
		t.Error(err)
	} else if err := json.Unmarshal([]byte(val), rec); err != nil {
		t.Error("Error converting value")
	} else if !reflect.DeepEqual(rec, as.BalanceSummaries[0]) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(as.BalanceSummaries[0]), utils.ToJSON(rec))
	}

}
func TestAccountFieldAsInterface(t *testing.T) {
	var acc *Account
	if _, err := acc.FieldAsInterface([]string{}); err != nil {
		t.Error(err)
	}
	acc = &Account{
		ID: "id",
		BalanceMap: map[string]Balances{
			"first": {
				&Balance{
					Uuid:  "uuid1",
					ID:    "id",
					Value: 20.44,
				},
				&Balance{
					Uuid:  "uuid2",
					ID:    "id2",
					Value: 12.2,
				},
			}},
		UnitCounters: UnitCounters{
			"first": []*UnitCounter{
				{
					CounterType: "balance",
					Counters: CounterFilters{
						{
							Value: 20.44,
							Filter: &BalanceFilter{
								Uuid: utils.StringPointer("filterUuid"),
								ID:   utils.StringPointer("filterId"),
								Type: utils.StringPointer("type"),
							},
						},
					},
				}}},
		ActionTriggers: ActionTriggers{
			{
				UniqueID:       "uniId",
				ID:             "id",
				ThresholdType:  "*min_event_counter",
				ThresholdValue: 20.55,
			},
			{
				UniqueID:       "uniId2",
				ID:             "id2",
				ThresholdType:  "*max_event_counter",
				ThresholdValue: 19.22,
				Recurrent:      true,
			},
		},
	}
	if _, err := acc.FieldAsInterface([]string{"Balance"}); err == nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{"BalanceMap[second]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if val, err := acc.FieldAsInterface([]string{"BalanceMap[first]"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acc.BalanceMap["first"], val) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(val), utils.ToJSON(acc.BalanceMap["first"]))
	} else if _, err := acc.FieldAsInterface([]string{"BalanceMap[first]", "UnitCounters[test]"}); err == nil {
		t.Error(err)
	} else if _, err := acc.FieldAsInterface([]string{"UnitCounters[test]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if val, err := acc.FieldAsInterface([]string{"UnitCounters[first]"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, acc.UnitCounters["first"]) {
		t.Errorf("expected %v ,received %v", val, acc.UnitCounters["first"])
	}
	if _, err := acc.FieldAsInterface([]string{"ActionTriggers[3]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := acc.FieldAsInterface([]string{"ActionTriggers[three]"}); err == nil {
		t.Error(err)
	} else if val, err := acc.FieldAsInterface([]string{"ActionTriggers[0]"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, acc.ActionTriggers[0]) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(acc.ActionTriggers[0]), utils.ToJSON(val))
	} else if _, err = acc.FieldAsInterface([]string{"ActionTriggers[0]", "test"}); err == nil {
		t.Error(err)
	}
	if _, err := acc.FieldAsInterface([]string{"ID"}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{"ID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := acc.FieldAsInterface([]string{"BalanceMap"}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{"BalanceMap", "first[0]"}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{"BalanceMap", "first[3]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{"BalanceMap", "first[first]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UnitCounters}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UnitCounters, "first"}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UnitCounters, "first", "sec"}); err == nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UnitCounters, "first", "2"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err := acc.FieldAsInterface([]string{utils.UnitCounters, "first", "0"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UnitCounters, "sec"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.ActionTriggers}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.ActionTriggers, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if val, err := acc.FieldAsInterface([]string{utils.ActionTriggers, "id"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, acc.ActionTriggers[0]) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(val), utils.ToJSON(acc.ActionTriggers[0]))
	} else if _, err = acc.FieldAsInterface([]string{utils.AllowNegative}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.AllowNegative, "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.Disabled}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.Disabled, "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UpdateTime}); err != nil {
		t.Error(err)
	} else if _, err = acc.FieldAsInterface([]string{utils.UpdateTime, "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestAccountFieldAsString(t *testing.T) {
	acc := &Account{
		ID: "id",
		BalanceMap: map[string]Balances{
			"first": {
				&Balance{
					Uuid:  "uuid1",
					ID:    "id",
					Value: 20.44,
				},
				&Balance{
					Uuid:  "uuid2",
					ID:    "id2",
					Value: 12.2,
				},
			}},
		UnitCounters: UnitCounters{
			"first": []*UnitCounter{
				{
					CounterType: "balance",
					Counters: CounterFilters{
						{
							Value: 20.44,
							Filter: &BalanceFilter{
								Uuid: utils.StringPointer("filterUuid"),
								ID:   utils.StringPointer("filterId"),
								Type: utils.StringPointer("type"),
							},
						},
					},
				}}},
		ActionTriggers: ActionTriggers{
			{
				UniqueID:       "uniId",
				ID:             "id",
				ThresholdType:  "*min_event_counter",
				ThresholdValue: 20.55,
			},
			{
				UniqueID:       "uniId2",
				ID:             "id2",
				ThresholdType:  "*max_event_counter",
				ThresholdValue: 19.22,
				Recurrent:      true,
			},
		},
	}
	if _, err := acc.FieldAsString([]string{"val"}); err == nil {
		t.Error(err)
	} else if _, err := acc.FieldAsString([]string{"ID"}); err != nil {
		t.Error(err)
	}
}

func TestGetSharedGroups(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			"first": {
				{SharedGroups: utils.StringMap{
					"firstval":  true,
					"firstval2": true,
				}}},
			"second": {
				{
					SharedGroups: utils.StringMap{
						"secondval":  true,
						"secondval2": false,
					}}},
		},
	}
	exp := []string{"firstval", "firstval2", "secondval", "secondval2"}

	val := acc.GetSharedGroups()

	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(val), utils.ToJSON(exp))
	}
}

func TestAccountAsOldStructure(t *testing.T) {
	acc := &Account{
		ID: "TEST_ID",
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					Value: 100}},
			utils.MetaVoice: {
				&Balance{
					Value:          10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{
					Weight:         10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MetaMonetary: []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Filter: &BalanceFilter{
						Type: utils.StringPointer(utils.MetaMonetary)},
						Value: 1.0}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMinEventCounter,
				ActionsID:      "TEST_ACTIONS"}},

		AllowNegative: true,
		Disabled:      true,
	}
	expAcc := &Account{
		ID: "*out:TEST_ID",
		BalanceMap: map[string]Balances{
			"*monetary*out": {
				&Balance{
					Value: 100}},
			"*voice*out": {
				&Balance{
					Value:          10 * float64(time.Second),
					Weight:         20,
					DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{
					Weight:         10,
					DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			"*monetary": []*UnitCounter{
				{Counters: CounterFilters{
					&CounterFilter{Filter: &BalanceFilter{
						Type: utils.StringPointer(utils.MetaMonetary)},
						Value: 1.0}}}}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMinEventCounter,
				ActionsID:      "TEST_ACTIONS"}},

		AllowNegative: true,
		Disabled:      true,
	}

	if val := acc.AsOldStructure(); reflect.DeepEqual(val, expAcc) {
		t.Errorf("expected %+v ,received %+v ", utils.ToJSON(expAcc), utils.ToJSON(val))
	}
}

func TestAccountSummary(t *testing.T) {

	as := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "CGRATES_1",
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				ID:       "summary_1",
				UUID:     "summary_uuid",
				Type:     "*voice",
				Initial:  2.0,
				Value:    12.2,
				Disabled: true,
			},
			&BalanceSummary{
				ID:       "summary_2",
				UUID:     "summary_uuid2",
				Type:     "*voice",
				Initial:  4.0,
				Value:    20.2,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}

	asOld := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "CGRATES_1",
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				ID:       "oldsummary_1",
				UUID:     "oldsummary_uuid",
				Type:     "*voice",
				Initial:  2.0,
				Value:    12.2,
				Disabled: true,
			},
			&BalanceSummary{
				ID:       "old_summary_2",
				UUID:     "old_summary_uuid2",
				Type:     "*data",
				Initial:  4.0,
				Value:    20.2,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}
	expAs := &AccountSummary{
		Tenant: "cgrates.org",
		ID:     "CGRATES_1",
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				ID:       "summary_1",
				UUID:     "summary_uuid",
				Type:     "*voice",
				Initial:  2.0,
				Value:    12.2,
				Disabled: true,
			},
			&BalanceSummary{
				ID:       "summary_2",
				UUID:     "summary_uuid2",
				Type:     "*voice",
				Initial:  4.0,
				Value:    20.2,
				Disabled: false,
			},
			&BalanceSummary{
				ID:       "oldsummary_1",
				UUID:     "oldsummary_uuid",
				Type:     "*voice",
				Initial:  0,
				Value:    0,
				Disabled: true,
			},
			&BalanceSummary{
				ID:       "old_summary_2",
				UUID:     "old_summary_uuid2",
				Type:     "*data",
				Initial:  0,
				Value:    0,
				Disabled: false,
			},
		},
		AllowNegative: false,
		Disabled:      false,
	}
	as.SetInitialValue(asOld)
	if !reflect.DeepEqual(as, expAs) {
		t.Errorf("expected %+v ,received %+v", utils.ToJSON(expAs), utils.ToJSON(as))
	}
}

func TestAccountGetCreditForPrefix(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	Cache.Clear(nil)
	acc := &Account{
		ID:            "cgrates.org:account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:             "voice1",
					Weight:         20,
					DestinationIDs: utils.StringMap{utils.MetaAny: false},
					precision:      0,
					SharedGroups:   utils.NewStringMap("SG_TEST"),
					Value:          3600},
			},
		},
	}
	Cache.Set(utils.CacheSharedGroups, "SG_TEST", &SharedGroup{
		Id: "SG_TEST", MemberIds: utils.NewStringMap("cgrates.org:account1"),
		AccountParameters: map[string]*SharingParameters{
			"*any": {Strategy: STRATEGY_MINE_RANDOM},
		},
	}, []string{}, true, utils.NonTransactional)
	cd := &CallDescriptor{
		Category:      "0",
		Tenant:        "vdf",
		TimeStart:     time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 10 * time.Second,
		Destination:   "0723",
		ToR:           utils.MetaMonetary,
	}
	if _, _, balances := acc.getCreditForPrefix(cd); len(balances) == 0 {
		t.Errorf("received %+v", utils.ToJSON(balances))
	}
}

func TestAcountSetBalanceAction(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSharedGroups: {
			Limit:     3,
			StaticTTL: true,
		},
	}
	db, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	SetDataStorage(dm)
	acc := &Account{
		ID: "vdf:minu",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{
					Uuid:           "uuid2",
					ExpirationDate: time.Date(2022, 12, 2, 2, 0, 0, 0, time.UTC),
					Value:          200 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("NAT"), Weight: 10,
					SharedGroups: utils.StringMap{
						"SharedGroups_true":  true,
						"SharedGroups_false": false,
					},
				},
				&Balance{
					Uuid:           "uuid1",
					ExpirationDate: time.Date(2021, 12, 2, 2, 0, 0, 0, time.UTC),
					Value:          100 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	fltrs := NewFilterS(cfg, nil, nil)
	a := &Action{
		Balance: &BalanceFilter{

			ID:   utils.StringPointer("id"),
			Type: utils.StringPointer("b_type"),
			Value: &utils.ValueFormula{
				Method: "value_method",
			},
			SharedGroups: utils.StringMapPointer(utils.NewStringMap("shrdGroup")),
		},
	}
	if err := dm.dataDB.SetSharedGroupDrv(&SharedGroup{
		Id:        "shrdGroup",
		MemberIds: utils.StringMap{}}); err != nil {
		t.Error(err)
	}
	if err = acc.setBalanceAction(a, fltrs); err != nil {
		t.Error(err)
	}
	exp := utils.StringMap{
		"vdf:minu": true,
	}
	if val, err := dm.dataDB.GetSharedGroupDrv("shrdGroup"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val.MemberIds, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(val.MemberIds))
	}
	if err = acc.setBalanceAction(nil, fltrs); err == nil || err.Error() != "nil action" {
		t.Error(err)
	}
}

func TestAccGetAllBalancesForPrefixLogg(t *testing.T) {
	tmp := Cache
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
		Cache = tmp
	}()
	Cache.Clear(nil)
	acc := &Account{
		ID:            "cgrates.org:account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:             "voice1",
					Weight:         20,
					DestinationIDs: utils.StringMap{utils.MetaAny: false},
					precision:      0,
					SharedGroups:   utils.NewStringMap("SG_TEST"),
					Value:          3600},
			},
		},
	}
	Cache.Set(utils.CacheSharedGroups, "SG_TEST", nil, []string{}, true, utils.NonTransactional)
	acc.getAlldBalancesForPrefix("0723", "0", utils.MetaMonetary, time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC))
	expLog := "Could not get shared group:"
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't containe %v ", rcvLog, expLog)
	}
}

func TestGetUniqueSharedGroupMemebersErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	tmp := Cache
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
		Cache = tmp
	}()
	Cache.Clear(nil)
	acc := &Account{
		ID:            "cgrates.org:account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:             "voice1",
					Weight:         20,
					DestinationIDs: utils.StringMap{utils.MetaAny: false},
					precision:      0,
					SharedGroups:   utils.NewStringMap("SG_TEST"),
					Value:          3600},
			},
		},
	}
	Cache.Set(utils.CacheSharedGroups, "SG_TEST", nil, []string{}, true, utils.NonTransactional)
	cd := &CallDescriptor{
		Category:      "0",
		Tenant:        "vdf",
		TimeStart:     time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 10 * time.Second,
		Destination:   "0723",
		ToR:           utils.MetaMonetary,
	}
	expLog := `Could not get shared group: `
	if _, err := acc.GetUniqueSharedGroupMembers(cd); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("Logger %v doesn't contain %v", rcvLog, expLog)
	}

}

func TestAccSetBalanceAction(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	a := &Action{
		ActionType: "*topup",
		Balance: &BalanceFilter{
			ID:    utils.StringPointer(utils.MetaDefault),
			Type:  utils.StringPointer(utils.MetaMonetary),
			Value: &utils.ValueFormula{Static: 10},
			SharedGroups: &utils.StringMap{
				"string1": true,
			}},
	}
	acc := &Account{
		ID:            "cgrates.org:account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MetaMonetary: {
				&Balance{
					ID:             "voice1",
					Weight:         20,
					DestinationIDs: utils.StringMap{utils.MetaAny: false},
					precision:      0,
					SharedGroups:   utils.NewStringMap("SG_TEST"),
					Value:          3600},
			},
		},
	}
	Cache.Set(utils.CacheSharedGroups, "string1", nil, []string{}, false, utils.NonTransactional)
	if err := acc.setBalanceAction(a, nil); err != nil {
		t.Error(err)
	}
}

func TestAccEnableAccountAction(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpDm := dm
	defer func() {
		dm = tmpDm
	}()
	db, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{
					ID:     "BALANCE_ID",
					Value:  20 * float64(time.Second),
					Weight: 10},
				&Balance{Value: 100 * float64(time.Second),
					Weight: 20},
			}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					ID:   utils.StringPointer("BALANCE_ID"),
					Type: utils.StringPointer(utils.MetaMonetary)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMinEventCounter,
				ActionsID:      "TEST_ACTIONS"},
		},
		UnitCounters: UnitCounters{
			utils.MetaVoice: []*UnitCounter{
				{
					CounterType: "*event",
					Counters: CounterFilters{
						&CounterFilter{
							Value: 0,
							Filter: &BalanceFilter{
								ID:             utils.StringPointer("BALANCE_ID"),
								Type:           utils.StringPointer(utils.MetaVoice),
								DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
							},
						},
					},
				},
			},
		},
	}
	acc2 := &Account{
		ID: "cgrates.org:1002",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{
					ID:     "BALANCE_ID",
					Value:  20 * float64(time.Second),
					Weight: 10},
			}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					ID:   utils.StringPointer("BALANCE_ID"),
					Type: utils.StringPointer(utils.MetaVoice)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMinEventCounter,
				ActionsID:      "TEST_ACTIONS2"},
		},
		UnitCounters: UnitCounters{
			utils.MetaVoice: []*UnitCounter{
				{
					CounterType: "*event",
					Counters: CounterFilters{
						&CounterFilter{
							Value: 0,
							Filter: &BalanceFilter{
								ID:             utils.StringPointer("BALANCE_ID"),
								Type:           utils.StringPointer(utils.MetaVoice),
								DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
							},
						},
					},
				},
			},
		},
	}
	acc3 := &Account{
		ID: "cgrates.org:1003",
		BalanceMap: map[string]Balances{
			utils.MetaVoice: {
				&Balance{
					dirty:  true,
					ID:     "BALANCE_ID",
					Value:  150 * float64(time.Second),
					Weight: 10},
			}},
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				Balance: &BalanceFilter{
					ID:   utils.StringPointer("BALANCE_ID"),
					Type: utils.StringPointer(utils.MetaVoice)},
				ThresholdValue: 100,
				ThresholdType:  utils.TriggerMaxBalance,
				ActionsID:      "TEST_ACTIONS3"},
		},
	}
	testCases := []struct {
		name       string
		account    *Account
		actions    Actions
		key        string
		actiontype string
	}{
		{name: "Enable account action",
			account: acc,
			actions: Actions{
				&Action{
					ActionType: utils.MetaEnableAccount,
					Filters:    []string{"*gte:~*req.BalanceMap.*voice[0].Value:20"},
					Balance: &BalanceFilter{
						ID:             utils.StringPointer("BALANCE_ID"),
						Type:           utils.StringPointer(utils.MetaVoice),
						DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
					},
					Weight: 9,
				},
			},
			key:        "TEST_ACTIONS",
			actiontype: utils.MetaEnableAccount,
		},
		{
			name:    "Disable account action",
			account: acc2,
			actions: Actions{
				&Action{
					ActionType: utils.MetaDisableAccount,
					Filters:    []string{"*string:~*req.BalanceMap.*voice[0].ID:BALANCE_ID"},
					Balance: &BalanceFilter{
						ID:             utils.StringPointer("BALANCE_ID"),
						Type:           utils.StringPointer(utils.MetaVoice),
						DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
					},
					Weight: 9,
				},
			},
			key:        "TEST_ACTIONS2",
			actiontype: utils.MetaDisableAccount,
		},
		{
			name:    "Set recurrent Action",
			account: acc3,
			actions: Actions{
				&Action{
					ActionType: utils.MetaSetRecurrent,
					Filters:    []string{"*string:~*req.BalanceMap.*voice[0].ID:BALANCE_ID"},
					Balance: &BalanceFilter{
						ID:             utils.StringPointer("BALANCE_ID"),
						Type:           utils.StringPointer(utils.MetaVoice),
						DestinationIDs: utils.StringMapPointer(utils.NewStringMap("GERMANY_O2")),
					},
					Weight: 9,
				},
			},
			key:        "TEST_ACTIONS3",
			actiontype: utils.MetaSetRecurrent,
		},
	}
	SetDataStorage(dm)
	for _, tc := range testCases {
		dm.SetActions(tc.key, tc.actions)
		t.Run(tc.name, func(t *testing.T) {
			tc.account.ExecuteActionTriggers(&Action{}, NewFilterS(cfg, nil, dm))

			switch tc.actiontype {
			case utils.MetaEnableAccount:
				if acc, err := dm.GetAccount(tc.account.ID); err != nil {
					t.Error(err)
				} else if acc.Disabled != false {
					t.Errorf("account should be enabled")
				}

			case utils.MetaDisableAccount:
				if acc, err := dm.GetAccount(tc.account.ID); err != nil {
					t.Error(err)
				} else if acc.Disabled != true {
					t.Errorf("account should be disabled")
				}
			}

		})
	}
}

func TestEngineToStringJSON(t *testing.T) {
	acc := &Account{
		ID:                "acc_123",
		AllowNegative:     false,
		Disabled:          true,
		UpdateTime:        time.Now(),
		executingTriggers: false,
	}
	want, err := json.Marshal(acc)
	if err != nil {
		t.Error(err)
	}
	got := acc.String()
	if got != string(want) {
		t.Errorf("acc.String()=%s, want%s", got, want)
	}
}

func TestEngineGetID(t *testing.T) {
	tests := []struct {
		name  string
		accID string
		want  string
	}{
		{
			name:  "Valid ID format",
			accID: "prefix" + utils.ConcatenatedKeySep + "suffix",
			want:  "suffix",
		},
		{
			name:  "Invalid ID format (missing separator)",
			accID: "invalidID",
			want:  "",
		},
		{
			name:  "Invalid ID format (too many parts)",
			accID: "prefix" + utils.ConcatenatedKeySep + "suffix" + utils.ConcatenatedKeySep + "extra",
			want:  "",
		},
		{
			name:  "Empty ID",
			accID: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{ID: tt.accID}
			got := acc.GetID()
			if got != tt.want {
				t.Errorf("Expected ID: %s, Got: %s", tt.want, got)
			}
		})
	}
}

func TestEngineNewAccountSummaryFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		want    *AccountSummary
	}{
		{
			name:    "Valid JSON",
			jsonStr: `{"tenant": "cgrates.org", "id": "1234", "balanceSummaries": [], "allowNegative": false, "disabled": true}`,
			want: &AccountSummary{
				Tenant:           "cgrates.org",
				ID:               "1234",
				BalanceSummaries: BalanceSummaries{},
				AllowNegative:    false,
				Disabled:         true,
			},
		},
		{
			name:    "Empty JSON",
			jsonStr: "",
			want:    nil,
		},
		{
			name:    "Null JSON",
			jsonStr: "null",
			want:    nil,
		},
		{
			name:    "Invalid JSON",
			jsonStr: "{invalid: json}",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := NewAccountSummaryFromJSON(tt.jsonStr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountSummaryFromJSON(%q) got = %v, want %v", tt.jsonStr, got, tt.want)
			}
		})
	}
}

func TestAccountSummaryString(t *testing.T) {
	account := &AccountSummary{
		Tenant:        "cgrates.org",
		ID:            "2012",
		AllowNegative: true,
		Disabled:      false,
		BalanceSummaries: BalanceSummaries{
			&BalanceSummary{
				Factors: ValueFactors{
					"factor1": 1.0,
					"factor2": 2.0,
				},
			},
		},
	}
	result := account.String()
	var parsedResult map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsedResult)
	if err != nil {
		t.Errorf("Error unmarshalling result: %v", err)
	}
}

func TestAccountProcessAccountSummaryField(t *testing.T) {
	type args struct {
		fldPath    []string
		accSummary any
		event      map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "Direct access for *AccountSummary (Tenant)",
			args: args{
				fldPath:    []string{"Tenant"},
				accSummary: &AccountSummary{Tenant: "test_tenant", ID: "id1"},
				event:      make(map[string]any),
			},
			want:    "test_tenant",
			wantErr: false,
		},
		{
			name: "Unmarshal with ValueFactors field (expecting error on missing factor)",
			args: args{
				fldPath:    []string{"BalanceSummaries", "0", "Factors", "factor1"},
				accSummary: `{"BalanceSummaries": [{"UUID": "summary1", "Factors": {"factor1": 0.2}}]}`,
				event:      make(map[string]any),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error on missing ValueFactors field",
			args: args{
				fldPath:    []string{"BalanceSummaries", "0", "Factors", "missing_factor"},
				accSummary: `{"BalanceSummaries": [{"UUID": "summary1"}]}`,
				event:      make(map[string]any),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processAccountSummaryField(tt.args.fldPath, tt.args.accSummary, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAccountSummaryField error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" && !tt.wantErr {
				t.Errorf("processAccountSummaryField mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProcessAccountSummaryFieldNonStringAccSummary(t *testing.T) {
	type args struct {
		fldPath    []string
		accSummary any
		event      map[string]any
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Error on non-string accSummary (marshal)",
			args: args{
				fldPath:    []string{"Tenant"},
				accSummary: map[string]string{"Name": "Test Account"},
				event:      make(map[string]any),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processAccountSummaryField(tt.args.fldPath, tt.args.accSummary, tt.args.event)
			if (err == nil) != tt.wantErr {
				t.Errorf("processAccountSummaryField error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAccountFieldAsInterfaceNilAccount(t *testing.T) {
	var acc *Account
	fldPath := []string{"ID"}
	_, err := acc.FieldAsInterface(fldPath)
	if err != utils.ErrNotFound {
		t.Errorf("Expected error %v, got %v", utils.ErrNotFound, err)
	}
}
