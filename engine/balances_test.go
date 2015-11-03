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

	"github.com/cgrates/cgrates/utils"
)

func TestBalanceSortPrecision(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs BalanceChain
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by weight!")
	}
}

func TestBalanceSortPrecisionWeightEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs BalanceChain
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortPrecisionWeightGreater(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs BalanceChain
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 1}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs BalanceChain
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeightLess(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs BalanceChain
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb2 || bs[1] != mb1 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	mb2 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	mb3 := &Balance{Weight: 1, precision: 1, RatingSubject: "2", DestinationIds: utils.StringMap{}}
	if !mb1.Equal(mb2) || mb2.Equal(mb3) {
		t.Error("Equal failure!", mb1 == mb2, mb3)
	}
}

func TestBalanceMatchFilter(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	mb2 := &Balance{Weight: 1, precision: 1, RatingSubject: "", DestinationIds: utils.StringMap{}}
	if !mb1.MatchFilter(mb2, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterEmpty(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	mb2 := &Balance{}
	if !mb1.MatchFilter(mb2, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterId(t *testing.T) {
	mb1 := &Balance{Id: "T1", Weight: 2, precision: 2, RatingSubject: "2", DestinationIds: utils.NewStringMap("NAT")}
	mb2 := &Balance{Id: "T1", Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	if !mb1.MatchFilter(mb2, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterDiffId(t *testing.T) {
	mb1 := &Balance{Id: "T1", Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	mb2 := &Balance{Id: "T2", Weight: 1, precision: 1, RatingSubject: "1", DestinationIds: utils.StringMap{}}
	if mb1.MatchFilter(mb2, false) {
		t.Errorf("Match filter failure: %+v != %+v", mb1, mb2)
	}
}

func TestBalanceClone(t *testing.T) {
	mb1 := &Balance{Value: 1, Weight: 2, RatingSubject: "test", DestinationIds: utils.NewStringMap("5")}
	mb2 := mb1.Clone()
	if mb1 == mb2 || !mb1.Equal(mb2) {
		t.Errorf("Cloning failure: \n%+v\n%+v", mb1, mb2)
	}
}

func TestBalanceMatchActionTriggerId(t *testing.T) {
	at := &ActionTrigger{BalanceId: "test"}
	b := &Balance{Id: "test"}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Id = "test1"
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Id = ""
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Id = "test"
	at.BalanceId = ""
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerDestination(t *testing.T) {
	at := &ActionTrigger{BalanceDestinationIds: utils.NewStringMap("test")}
	b := &Balance{DestinationIds: utils.NewStringMap("test")}
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIds = utils.NewStringMap("test1")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIds = utils.NewStringMap("")
	if b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIds = utils.NewStringMap("test")
	at.BalanceDestinationIds = utils.NewStringMap("")
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerWeight(t *testing.T) {
	at := &ActionTrigger{BalanceWeight: 1}
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
	at.BalanceWeight = 0
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerRatingSubject(t *testing.T) {
	at := &ActionTrigger{BalanceRatingSubject: "test"}
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
	at.BalanceRatingSubject = ""
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerSharedGroup(t *testing.T) {
	at := &ActionTrigger{BalanceSharedGroups: utils.NewStringMap("test")}
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
	at.BalanceSharedGroups = utils.NewStringMap("")
	if !b.MatchActionTrigger(at) {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceIsDefault(t *testing.T) {
	b := &Balance{Weight: 0}
	if b.IsDefault() {
		t.Errorf("Balance should not be default: %+v", b)
	}
	b = &Balance{Id: utils.META_DEFAULT}
	if !b.IsDefault() {
		t.Errorf("Balance should be default: %+v", b)
	}
}
