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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestBalanceSortPrecision(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by weight!")
	}
}

func TestBalanceSortPrecisionWeightEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortPrecisionWeightGreater(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 1}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight2(t *testing.T) {
	bs := Balances{
		&Balance{ID: "B1", Weight: 2, precision: 1},
		&Balance{ID: "B2", Weight: 1, precision: 1},
	}
	bs.Sort()
	if bs[0].ID != "B1" && bs[1].ID != "B2" {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight3(t *testing.T) {
	bs := Balances{
		&Balance{ID: "B1", Weight: 2, Value: 10.0},
		&Balance{ID: "B2", Weight: 4, Value: 10.0},
	}
	bs.Sort()
	if bs[0].ID != "B2" && bs[1].ID != "B1" {
		t.Error(utils.ToJSON(bs))
	}
}

func TestBalanceSortWeightLess(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb2 || bs[1] != mb1 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb3 := &Balance{Weight: 1, precision: 1, RatingSubject: "2", DestinationIDs: utils.StringMap{}}
	if !mb1.Equal(mb2) || mb2.Equal(mb3) {
		t.Error("Equal failure!", mb1 == mb2, mb3)
	}
}

func TestBalanceMatchFilter(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{Weight: utils.Float64Pointer(1), RatingSubject: nil, DestinationIDs: nil}
	if !mb1.MatchFilter(mb2, "", false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}

	mb1.Uuid, mb2.Uuid = "id", utils.StringPointer("id")
	if !mb1.MatchFilter(mb2, "", false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}

}

func TestBalanceMatchFilterEmpty(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{}
	if !mb1.MatchFilter(mb2, "", false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 2, precision: 2, RatingSubject: "2", DestinationIDs: utils.NewStringMap("NAT")}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T1"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if !mb1.MatchFilter(mb2, "", false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterDiffId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T2"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if mb1.MatchFilter(mb2, "", false, false) {
		t.Errorf("Match filter failure: %+v != %+v", mb1, mb2)
	}
}

func TestBalanceClone(t *testing.T) {
	var mb1 *Balance
	if mb2 := mb1.Clone(); mb2 != nil {
		t.Errorf("Balance should be %v", mb2)
	}

	mb1 = &Balance{Value: 1, Weight: 2, RatingSubject: "test", DestinationIDs: utils.NewStringMap("5")}
	mb2 := mb1.Clone()
	if mb1 == mb2 || !mb1.Equal(mb2) {
		t.Errorf("Cloning failure: \n%+v\n%+v", mb1, mb2)
	}
}

func TestBalanceMatchActionTriggerId(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{ID: utils.StringPointer("test")}}
	b := &Balance{ID: "test"}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test1"
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = ""
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test"
	at.Balance.ID = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerDestination(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{DestinationIDs: utils.StringMapPointer(utils.NewStringMap("test"))}}
	b := &Balance{DestinationIDs: utils.NewStringMap("test")}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test1")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test")
	at.Balance.DestinationIDs = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerWeight(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{Weight: utils.Float64Pointer(1)}}
	b := &Balance{Weight: 1}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 2
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 0
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 1
	at.Balance.Weight = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerRatingSubject(t *testing.T) {

	at := &ActionTrigger{Balance: &BalanceFilter{RatingSubject: utils.StringPointer("test")}}
	b := &Balance{RatingSubject: "test"}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test1"
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = ""
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test"
	at.Balance.RatingSubject = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerSharedGroup(t *testing.T) {
	at := &ActionTrigger{Balance: &BalanceFilter{SharedGroups: utils.StringMapPointer(utils.NewStringMap("test"))}}
	b := &Balance{SharedGroups: utils.NewStringMap("test")}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test1")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test")
	at.Balance.SharedGroups = nil
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceIsDefault(t *testing.T) {
	b := &Balance{Weight: 0}
	if b.IsDefault() {
		t.Errorf("Balance should not be default: %+v", b)
	}
	b = &Balance{ID: utils.MetaDefault}
	if !b.IsDefault() {
		t.Errorf("Balance should be default: %+v", b)
	}
}

func TestBalanceIsExpiredAt(t *testing.T) {
	//expiration date is 0
	balance := &Balance{}
	var date2 time.Time
	if rcv := balance.IsExpiredAt(date2); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}
	//expiration date before time t
	balance.ExpirationDate = time.Date(2020, time.April, 18, 23, 0, 3, 0, time.UTC)
	date2 = time.Date(2020, time.April, 18, 23, 0, 4, 0, time.UTC)
	if rcv := balance.IsExpiredAt(date2); !rcv {
		t.Errorf("Expecting: true , received: %+v", rcv)
	}
	//expiration date after time t
	date2 = time.Date(2020, time.April, 18, 23, 0, 2, 0, time.UTC)
	if rcv := balance.IsExpiredAt(date2); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}
	//time t = 0
	var date3 time.Time
	if rcv := balance.IsExpiredAt(date3); rcv {
		t.Errorf("Expecting: false , received: %+v", rcv)
	}

}

func TestBalanceAsInterface(t *testing.T) {

	b := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          2.21,
		ExpirationDate: time.Date(2022, 11, 22, 9, 0, 0, 0, time.UTC),
		Weight:         2.88,
		DestinationIDs: utils.StringMap{
			"destId1": true,
			"destId2": true,
		},
		RatingSubject: "rating",
		Categories: utils.StringMap{
			"ctg1": true,
			"ctg2": false,
		},
		SharedGroups: utils.StringMap{
			"shgp1": false,
			"shgp2": true,
		},
		Timings: []*RITiming{
			{
				ID:    "id",
				Years: utils.Years{2, 3},
			},
		},
		TimingIDs: utils.StringMap{
			"timingid1": true,
			"timingid2": false,
		},
		Factors: ValueFactors{
			"factor1": 2.21,
			"factor2": 1.34,
		},
	}

	if _, err := b.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"value"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"DestinationIDs[destId1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Categories[ctg1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"SharedGroups[shgp1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"TimingIDs[timingid1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[zero]"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[2]"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[2]", "val"}); err == nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Factors[factor1]", "secondVal"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err = b.FieldAsInterface([]string{"DestinationIDs[destId1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Categories[ctg1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"SharedGroups[shgp1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"TimingIDs[timingid1]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Timings[0]"}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{"Factors[factor1]"}); err != nil {
		t.Error(err)
	}
	if _, err = b.FieldAsInterface([]string{utils.Uuid}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ExpirationDate}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Weight}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.RatingSubject}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Categories}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.SharedGroups}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Timings}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Disabled}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Factors}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Blocker}); err != nil {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.TimingIDs}); err != nil {
		t.Error(err)
	}

	if _, err = b.FieldAsInterface([]string{utils.TimingIDs, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Uuid, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ID, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Value, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.ExpirationDate, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Weight, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.DestinationIDs, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.RatingSubject, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.Categories, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = b.FieldAsInterface([]string{utils.SharedGroups, "val"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestValueFactorsFieldAsInterface(t *testing.T) {
	v := &ValueFactors{
		"FACT_VAL": 20.22,
	}
	if _, err := v.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = v.FieldAsInterface([]string{"TEST"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = v.FieldAsInterface([]string{"FACT_VAL"}); err != nil {
		t.Error(err)
	}
}
func TestValueFactorFieldAsString(t *testing.T) {
	v := &ValueFactors{
		"FACT_VAL": 20.22,
	}
	if _, err = v.FieldAsString([]string{"TEST"}); err == nil {
		t.Error(err)
	} else if _, err = v.FieldAsString([]string{"FACT_VAL"}); err != nil {
		t.Error(err)
	}
}

func TestBalancesHasBalance(t *testing.T) {
	bc := Balances{
		{
			Uuid:           "uuid",
			ID:             "id",
			Value:          12.22,
			ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
			Blocker:        true,
			Disabled:       true,
			precision:      2,
		},
		{
			Uuid:           "uuid2",
			ID:             "id2",
			Value:          133.22,
			ExpirationDate: time.Date(2023, 3, 21, 5, 0, 0, 0, time.UTC),
			Blocker:        true,
			Disabled:       true,
			precision:      2,
		},
	}
	balance := &Balance{
		Uuid:           "uuid",
		ID:             "id",
		Value:          12.22,
		ExpirationDate: time.Date(2022, 11, 1, 20, 0, 0, 0, time.UTC),
		Blocker:        true,
		Disabled:       true,
		precision:      2,
	}

	if !bc.HasBalance(balance) {
		t.Error("should be true")
	}

}

func TestGetMinutesForCredi(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	b := &Balance{
		Value:          20 * float64(time.Second),
		DestinationIDs: utils.NewStringMap("NAT"),
		Weight:         10, RatingSubject: "rif",
	}
	cd := &CallDescriptor{
		Category:      "postpaid",
		ToR:           utils.MetaVoice,
		Tenant:        "foehn",
		Subject:       "foehn",
		Account:       "foehn",
		Destination:   "0034678096720",
		TimeStart:     time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:       time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 176 * time.Second,
	}
	if dur, _ := b.GetMinutesForCredit(cd, 12); dur != 0 {
		t.Error(err)
	}
	expLog := `Error getting new cost for balance subject:`
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", utils.ToJSON(expLog), utils.ToJSON(rcvLog))
	}
}

func TestBalancesStringValFactors(t *testing.T) {
	factors := ValueFactors{
		"factor1": 10.5,
		"factor2": 20.3,
	}
	jsonString := factors.String()
	if jsonString == "" {
		t.Errorf("String(): expected non-empty JSON string, got empty")
	}
}

func TestBalancesHasDestination(t *testing.T) {
	balance := &Balance{
		DestinationIDs: utils.StringMap{
			"destination1": true,
			"destination2": false,
		},
	}
	hasDest := balance.HasDestination()
	expected := true
	if hasDest != expected {
		t.Errorf("HasDestination(): expected %t, got %t", expected, hasDest)
	}
	balanceEmpty := &Balance{
		DestinationIDs: utils.StringMap{},
	}
	hasDestEmpty := balanceEmpty.HasDestination()
	expectedEmpty := false
	if hasDestEmpty != expectedEmpty {
		t.Errorf("HasDestination(): expected %t, got %t for empty DestinationIDs map", expectedEmpty, hasDestEmpty)
	}
}

func TestBalancesMatchDestination(t *testing.T) {
	balanceWithDestinations := &Balance{
		DestinationIDs: utils.StringMap{
			"destination1": true,
			"destination2": false,
		},
	}
	if !balanceWithDestinations.MatchDestination("destination1") {
		t.Errorf("MatchDestination(destination1): expected true, got false")
	}
	if balanceWithDestinations.MatchDestination("destination2") {
		t.Errorf("MatchDestination(destination2): expected false, got true")
	}
	if balanceWithDestinations.MatchDestination("nonexistent") {
		t.Errorf("MatchDestination(nonexistent): expected false, got true")
	}
	balanceNoDestinations := &Balance{
		DestinationIDs: utils.StringMap{},
	}
	if !balanceNoDestinations.MatchDestination("anydestination") {
		t.Errorf("MatchDestination(anydestination): expected true, got false for empty DestinationIDs map")
	}
}

func TestBalancesSetInitialValue(t *testing.T) {
	t.Run("OldNotNull", func(t *testing.T) {
		as := &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{UUID: "1", Value: 100, Initial: 0},
				{UUID: "2", Value: 200, Initial: 0},
			},
		}
		old := &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{UUID: "1", Value: 50, Initial: 0},
				{UUID: "3", Value: 300, Initial: 0},
			},
		}
		as.SetInitialValue(old)
		expected := &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{UUID: "1", Value: 100, Initial: 50},
				{UUID: "2", Value: 200, Initial: 0},
				{UUID: "3", Value: 0, Initial: 0},
			},
		}
		if diff := cmp.Diff(as, expected); diff != "" {
			t.Errorf("Unexpected result (-got +want):\n%s", diff)
		}
	})
	t.Run("OldNull", func(t *testing.T) {
		as := &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{UUID: "1", Value: 100, Initial: 0},
				{UUID: "2", Value: 200, Initial: 0},
			},
		}
		as.SetInitialValue(nil)
		expected := &AccountSummary{
			BalanceSummaries: BalanceSummaries{
				{UUID: "1", Value: 100, Initial: 0},
				{UUID: "2", Value: 200, Initial: 0},
			},
		}

		if diff := cmp.Diff(as, expected); diff != "" {
			t.Errorf("Unexpected result (-got +want):\n%s", diff)
		}
	})
}

func TestBalancesValueFactorsGetValue(t *testing.T) {
	testCases := []struct {
		factors  ValueFactors
		category string
		expected float64
	}{
		{
			factors:  nil,
			category: "category1",
			expected: 1.0,
		},
		{
			factors:  ValueFactors{"category2": 2.0, "category3": 3.0},
			category: "category1",
			expected: 1.0,
		},
		{
			factors:  ValueFactors{"category1": 1.5, "category2": 2.0, "category3": 3.0},
			category: "category1",
			expected: 1.5,
		},
		{
			factors:  ValueFactors{"category1": 0.0, "category2": 2.0, "category3": 3.0},
			category: "category1",
			expected: 0.0,
		},
	}
	for _, tc := range testCases {
		result := tc.factors.GetValue(tc.category)
		if result != tc.expected {
			t.Errorf("Test failed for category '%s': expected %.2f, got %.2f", tc.category, tc.expected, result)
		}
	}
}
func TestBalancesStringJson(t *testing.T) {
	balances := Balances{
		&Balance{
			Uuid:           "uuid123",
			ID:             "balance123",
			Value:          100.0,
			ExpirationDate: time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			Weight:         1.5,
			DestinationIDs: utils.StringMap{},
			RatingSubject:  "ratingSub",
			Categories:     utils.StringMap{},
			SharedGroups:   utils.StringMap{},
			Timings:        []*RITiming{},
			TimingIDs:      utils.StringMap{},
			Disabled:       false,
			Blocker:        true,
			precision:      2,
			account:        nil,
			dirty:          false,
		},
	}

	result := balances.String()
	if result == "" {
		t.Error("Expected non-empty JSON string, but got empty string")
	}
}

func TestBalanceFieldAsString(t *testing.T) {
	balance := &Balance{
		Uuid: "uuid123",
	}
	val, err := balance.FieldAsString([]string{"Uuid"})
	if err != nil {
		t.Errorf("Unexpected error for field 'Uuid': %v", err)
	}
	expected := balance.Uuid
	if val != expected {
		t.Errorf("Expected value '%s' for field 'Uuid', but got '%s'", expected, val)
	}
	_, err = balance.FieldAsString([]string{"InvalidField"})
	if err == nil {
		t.Error("Expected error for invalid field path, but got nil")
	}
}

func TestBalancesString(t *testing.T) {
	balance := &Balance{
		Uuid: "123e4567-e89b-12d3-a456-426614174000",
	}
	jsonStr := balance.String()
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Errorf("Error unmarshalling JSON string: %v", err)
	}
	expectedUuid := balance.Uuid
	if uuid, ok := data["Uuid"].(string); !ok || uuid != expectedUuid {
		t.Errorf("Expected Uuid '%s' in JSON, but got '%v'", expectedUuid, data["Uuid"])
	}
}
func TestBalancesHasBalanceReturn(t *testing.T) {
	balances := Balances{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}
	existingBalance := &Balance{ID: "2"}
	if !balances.HasBalance(existingBalance) {
		t.Errorf("Expected balance with ID '%s' to exist, but it does not", existingBalance.ID)
	}
	nonExistingBalance := &Balance{ID: "4"}
	if balances.HasBalance(nonExistingBalance) {
		t.Errorf("Expected balance with ID '%s' to not exist, but it does", nonExistingBalance.ID)
	}
}

func TestBalancesEqual(t *testing.T) {
	balances1 := Balances{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}
	balances2 := Balances{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}
	if !balances1.Equal(balances1) {
		t.Errorf("Expected balances1 to equal itself, but it does not")
	}
	if !balances1.Equal(balances2) {
		t.Errorf("Expected balances1 to equal balances2, but they are not equal")
	}
	balances3 := Balances{
		{ID: "1"},
		{ID: "2"},
	}
	if balances1.Equal(balances3) {
		t.Errorf("Expected balances1 to not equal balances3, but they are equal")
	}
	balances4 := Balances{
		{ID: "1"},
		{ID: "2"},
		{ID: "4"},
	}
	if balances1.Equal(balances4) {
		t.Errorf("Expected balances1 to not equal balances4, but they are equal")
	}
}

func TestBalancesFieldAsString(t *testing.T) {
	bc := Balances{
		&Balance{
			Uuid:           "uuid123",
			ID:             "balance123",
			Value:          100.0,
			ExpirationDate: time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			Weight:         1.5,
			DestinationIDs: utils.StringMap{},
			RatingSubject:  "ratingSub",
			Categories:     utils.StringMap{},
			SharedGroups:   utils.StringMap{},
			Timings:        []*RITiming{},
			TimingIDs:      utils.StringMap{},
			Disabled:       false,
			Blocker:        true,
			precision:      2,
			account:        nil,
			dirty:          false,
		},
	}

	t.Run("empty field path", func(t *testing.T) {
		val, err := bc.FieldAsString([]string{})
		if err == nil {
			t.Error("expected error, got nil")
		}
		if val != "" {
			t.Errorf("expected empty string, got '%v'", val)
		}
	})

	t.Run("invalid field path", func(t *testing.T) {
		val, err := bc.FieldAsString([]string{"invalid"})
		if err == nil {
			t.Error("expected error, got nil")
		}
		if val != "" {
			t.Errorf("expected empty string, got '%v'", val)
		}
	})

	t.Run("valid field path for non-existing ID", func(t *testing.T) {
		val, err := bc.FieldAsString([]string{"2", "ID"})
		if err == nil {
			t.Error("expected error, got nil")
		}
		if val != "" {
			t.Errorf("expected empty string, got '%v'", val)
		}
	})
}

func TestBalancesIsActiveAt(t *testing.T) {
	testTime := time.Now()

	t.Run("balance is disabled", func(t *testing.T) {
		balance := &Balance{
			Disabled: true,
		}
		if balance.IsActiveAt(testTime) {
			t.Errorf("expected false, got true")
		}
	})

}

func TestBalancesFieldAsInterfaceIndexPath(t *testing.T) {

	bc := Balances{
		&Balance{
			Uuid:           "uuid123",
			ID:             "balance123",
			Value:          100.0,
			ExpirationDate: time.Date(2024, time.December, 31, 23, 59, 59, 0, time.UTC),
			Weight:         1.5,
			DestinationIDs: utils.StringMap{},
			RatingSubject:  "ratingSub",
			Categories:     utils.StringMap{},
			SharedGroups:   utils.StringMap{},
			Timings:        []*RITiming{},
			TimingIDs:      utils.StringMap{},
			Disabled:       false,
			Blocker:        true,
			precision:      2,
			account:        nil,
			dirty:          false,
		},
	}

	t.Run("test index path in FieldAsInterface", func(t *testing.T) {

		idx := "0"
		val, err := bc.FieldAsInterface([]string{idx})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		expected := bc[0]
		if val != expected {
			t.Errorf("expected '%v', got '%v'", expected, val)
		}
	})

}
