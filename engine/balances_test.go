/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterEmpty(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{}
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 2, precision: 2, RatingSubject: "2", DestinationIDs: utils.NewStringMap("NAT")}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T1"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if !mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v == %+v", mb1, mb2)
	}
}

func TestBalanceMatchFilterDiffId(t *testing.T) {
	mb1 := &Balance{ID: "T1", Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &BalanceFilter{ID: utils.StringPointer("T2"), Weight: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("1"), DestinationIDs: nil}
	if mb1.MatchFilter(mb2, false, false) {
		t.Errorf("Match filter failure: %+v != %+v", mb1, mb2)
	}
}

func TestBalanceClone(t *testing.T) {
	mb1 := &Balance{Value: 1, Weight: 2, RatingSubject: "test", DestinationIDs: utils.NewStringMap("5")}
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

func TestBalancesSaveDirtyBalances(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmpConn := connMgr
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
		dm = tmpDm
		connMgr = tmpConn
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	cfg.RalsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
	})
	acc := &Account{
		ID: "cgrates.org:cond",
		BalanceMap: map[string]Balances{
			utils.MONETARY: {
				&Balance{
					Uuid:   utils.GenUUID(),
					Value:  1,
					Weight: 10,
				},
				&Balance{
					Uuid:   utils.GenUUID(),
					Value:  6,
					Weight: 20,
				},
			},
			utils.VOICE: {
				&Balance{
					Uuid:   utils.GenUUID(),
					Value:  10,
					Weight: 10,
				},
				&Balance{
					Uuid:   utils.GenUUID(),
					Value:  100,
					Weight: 20,
				},
			},
		}}
	bAcc := &Account{
		ID: "cgrates.org:max",
		BalanceMap: map[string]Balances{
			utils.MONETARY: {
				&Balance{Value: 11, Weight: 20},
			}},
	}
	bc := Balances{
		&Balance{Value: 200 * float64(time.Second),
			DestinationIDs: utils.NewStringMap("NAT"), Weight: 10,
			dirty:   true,
			account: bAcc,
		},
	}
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	SetConnManager(connMgr)
	bc.SaveDirtyBalances(acc)
	if _, err := dm.GetAccount("cgrates.org:max"); err != nil {
		t.Error(err)
	}
}

func TestBalancePublish(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	tmpDm := dm
	tmpConn := connMgr
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
		SetDataStorage(tmpDm)
		SetConnManager(tmpConn)
	}()
	cfg.RalsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.RalsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.StatSv1ProcessEvent {

			return nil
		} else if serviceMethod == utils.ThresholdSv1ProcessEvent {

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):      clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
	})
	SetConnManager(connMgr)
	at := &ActionTrigger{
		UniqueID:      "TestTR5",
		ThresholdType: utils.TRIGGER_MAX_BALANCE,
		Balance: &BalanceFilter{
			Type:   utils.StringPointer(utils.VOICE),
			Weight: utils.Float64Pointer(10),
		},
		ActionsID: "ACT_1",
	}
	ub := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.VOICE: {
				{
					Value:          10,
					DestinationIDs: utils.NewStringMap("DEST"),
				},
			},
		},
	}
	dm.SetActions("ACT_1", Actions{
		&Action{
			ActionType: utils.MetaPublishBalance,
			Balance: &BalanceFilter{
				Type:  utils.StringPointer(utils.VOICE),
				Value: &utils.ValueFormula{Static: 15},
			},
		},
	}, utils.NonTransactional)
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	if err := at.Execute(ub); err != nil {
		t.Error(err)
	}

}

func TestBalancesMatchFilter(t *testing.T) {
	b := Balance{
		ID: "test",
	}

	rcv := b.MatchFilter(nil, false, false)

	if rcv != true {
		t.Error(rcv)
	}

	str := "test"
	bf := BalanceFilter{
		ID:   &str,
		Uuid: &str,
	}

	rcv = b.MatchFilter(&bf, false, false)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestBalancesHardMatchFilter(t *testing.T) {
	b := Balance{
		ID: "test",
	}

	rcv := b.HardMatchFilter(nil, false)

	if rcv != true {
		t.Error(rcv)
	}

	str := "test"
	bf := BalanceFilter{
		ID:   &str,
		Uuid: &str,
	}

	rcv = b.HardMatchFilter(&bf, false)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestBalancesIsActive(t *testing.T) {
	b := Balance{
		Disabled: true,
	}

	tm := time.Now()
	rcv := b.IsActiveAt(tm)

	if rcv != false {
		t.Error(rcv)
	}

	b2 := Balance{
		Disabled: false,
		Timings: []*RITiming{
			{
				Years:      utils.Years{2020, 2022},
				Months:     utils.Months{time.Now().Month()},
				MonthDays:  utils.MonthDays{time.Now().Day()},
				WeekDays:   utils.WeekDays{time.Now().Weekday()},
				StartTime:  "00:00:00",
				EndTime:    "00:00:01",
				cronString: "test",
				tag:        "test",
			},
		},
	}

	tm2 := time.Now()
	rcv = b2.IsActiveAt(tm2)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestBalancesHasDestination(t *testing.T) {
	b := Balance{
		DestinationIDs: utils.StringMap{"*any": false},
	}

	rcv := b.HasDestination()

	if rcv != true {
		t.Error(rcv)
	}
}

func TestBalancesMatchDestination(t *testing.T) {

	b := Balance{
		DestinationIDs: utils.StringMap{
			"*any": false,
			"test": true,
		},
	}

	rcv := b.MatchDestination("test")

	if rcv != true {
		t.Error(rcv)
	}
}

func TestBalancegetMatchingPrefixAndDestID(t *testing.T) {
	b := Balance{}

	pre, dest := b.getMatchingPrefixAndDestID("")

	if pre != "" && dest != "" {
		t.Error(pre, dest)
	}
}

func TestBalancesEqual(t *testing.T) {
	b := Balances{}

	o := Balances{{}, {}}

	rcv := b.Equal(o)

	if rcv != false {
		t.Error(rcv)
	}

	b2 := Balances{{
		Uuid: "test",
		ID:   "test",
	}}

	o2 := Balances{{
		Uuid: "test2",
		ID:   "test2",
	}}

	rcv = b2.Equal(o2)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestBalanceHasBalance(t *testing.T) {
	bc := Balances{}
	b := Balance{}

	rcv := bc.HasBalance(&b)

	if rcv != false {
		t.Error(rcv)
	}

	bc2 := Balances{&b}

	rcv = bc2.HasBalance(&b)

	if rcv != true {
		t.Error(rcv)
	}
}

func TestBalancesGetValue(t *testing.T) {
	f := ValueFactor{}

	rcv := f.GetValue("test")

	if rcv != 1.0 {
		t.Error(rcv)
	}
}

func TestBalancesFieldAsInterface(t *testing.T) {
	bl := BalanceSummary{}
	flp := []string{}

	rcv, err := bl.FieldAsInterface(flp)

	if err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if rcv != nil {
		t.Error(rcv)
	}

	bl2 := BalanceSummary{}
	flp2 := []string{"test"}

	rcv, err = bl2.FieldAsInterface(flp2)

	if err.Error() != "unsupported field prefix: <test>" {
		t.Fatal(err)
	}

	if rcv != nil {
		t.Error(rcv)
	}

	bl3 := BalanceSummary{
		Disabled: true,
	}
	flp3 := []string{"Disabled"}

	rcv, err = bl3.FieldAsInterface(flp3)

	if err != nil {
		t.Fatal(err)
	}

	if rcv != true {
		t.Error(rcv)
	}
}

func TestBalanceCloneNil(t *testing.T) {
	var b *Balance

	rcv := b.Clone()

	if rcv != nil {
		t.Error(rcv)
	}
}
