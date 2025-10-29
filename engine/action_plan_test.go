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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestActionTimingTasks(t *testing.T) {
	//empty check
	actionTiming := new(ActionTiming)
	eOut := []*Task{{Uuid: "", ActionsID: ""}}
	rcv := actionTiming.Tasks()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	//multiple check
	actionTiming.ActionsID = "test"
	actionTiming.Uuid = "test"
	actionTiming.accountIDs = utils.StringMap{"1001": true, "1002": true, "1003": true}
	eOut = []*Task{
		{Uuid: "test", AccountID: "1001", ActionsID: "test"},
		{Uuid: "test", AccountID: "1002", ActionsID: "test"},
		{Uuid: "test", AccountID: "1003", ActionsID: "test"},
	}
	rcv = actionTiming.Tasks()
	sort.Slice(rcv, func(i, j int) bool { return rcv[i].AccountID < rcv[j].AccountID })
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestActionTimingRemoveAccountID(t *testing.T) {
	actionTiming := &ActionTiming{
		accountIDs: utils.StringMap{"1001": true, "1002": true, "1003": true},
	}
	eOut := utils.StringMap{"1002": true, "1003": true}
	rcv := actionTiming.RemoveAccountID("1001")
	if !rcv {
		t.Errorf("Account ID not found ")
	}
	if !reflect.DeepEqual(eOut, actionTiming.accountIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, actionTiming.accountIDs)
	}
	//check for Account ID not found
	rcv = actionTiming.RemoveAccountID("1001")
	if rcv {
		t.Errorf("Expected AccountID to be not found")
	}
}

func TestActionPlanRemoveAccountID(t *testing.T) {
	actionPlan := &ActionPlan{
		AccountIDs: utils.StringMap{"1001": true, "1002": true, "1003": true},
	}
	eOut := utils.StringMap{"1002": true, "1003": true}
	rcv := actionPlan.RemoveAccountID("1001")
	if !rcv {
		t.Errorf("Account ID not found ")
	}
	if !reflect.DeepEqual(eOut, actionPlan.AccountIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, actionPlan.AccountIDs)
	}
	//check for Account ID not found
	rcv = actionPlan.RemoveAccountID("1001")
	if rcv {
		t.Errorf("Expected AccountID to be not found")
	}
}

func TestActionPlanClone(t *testing.T) {
	at1 := &ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true, "two": true, "three": true},
		ActionTimings: []*ActionTiming{
			{
				Uuid:      "Uuid_test1",
				ActionsID: "ActionsID_test1",
				Weight:    0.8,
				Timing: &RateInterval{
					Weight: 0.7,
				},
			},
		},
	}
	clned, err := at1.Clone()
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("\nExpecting: %+v,\n received: %+v", at1, at1Cloned)
	}
}

func TestActionTimingClone(t *testing.T) {
	at := &ActionTiming{
		Uuid:      "Uuid_test",
		ActionsID: "ActionsID_test",
		Weight:    0.7,
	}
	if cloned := at.Clone(); !reflect.DeepEqual(at, cloned) {
		t.Errorf("\nExpecting: %+v,\n received: %+v", at, cloned)
	}
}

func TestActionTimindSetActions(t *testing.T) {
	actionTiming := new(ActionTiming)

	actions := Actions{
		&Action{ActionType: "test", Filter: "test"},
		&Action{ActionType: "test1", Filter: "test1"},
	}
	actionTiming.SetActions(actions)
	if !reflect.DeepEqual(actions, actionTiming.actions) {
		t.Errorf("Expecting: %+v, received: %+v", actions, actionTiming.actions)
	}
}

func TestActionTimingSetAccountIDs(t *testing.T) {
	actionTiming := new(ActionTiming)
	accountIDs := utils.StringMap{"one": true, "two": true, "three": true}
	actionTiming.SetAccountIDs(accountIDs)

	if !reflect.DeepEqual(accountIDs, actionTiming.accountIDs) {
		t.Errorf("Expecting: %+v, received: %+v", accountIDs, actionTiming.accountIDs)
	}
}

func TestActionTimingGetAccountIDs(t *testing.T) {
	actionTiming := &ActionTiming{
		accountIDs: utils.StringMap{"one": true, "two": true, "three": true},
	}
	accIDs := utils.StringMap{"one": true, "two": true, "three": true}
	rcv := actionTiming.GetAccountIDs()

	if !reflect.DeepEqual(accIDs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", accIDs, rcv)
	}
}
func TestActionTimingSetActionPlanID(t *testing.T) {
	actionTiming := new(ActionTiming)
	id := "test"
	actionTiming.SetActionPlanID(id)
	if !reflect.DeepEqual(id, actionTiming.actionPlanID) {
		t.Errorf("Expecting: %+v, received: %+v", id, actionTiming.actionPlanID)
	}
}

func TestActionTimingGetActionPlanID(t *testing.T) {
	id := "test"
	actionTiming := new(ActionTiming)
	actionTiming.actionPlanID = id

	rcv := actionTiming.GetActionPlanID()
	if !reflect.DeepEqual(id, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", id, rcv)
	}
}

func TestActionTimingIsASAP(t *testing.T) {
	actionTiming := new(ActionTiming)
	if rcv := actionTiming.IsASAP(); rcv {
		t.Error("Expecting false return")
	}
}

func TestAtplLen(t *testing.T) {
	atpl := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "first", accountIDs: utils.StringMap{"1001": true, "1002": true}},
		&ActionTiming{Uuid: "second", accountIDs: utils.StringMap{"1004": true, "1005": true}},
		&ActionTiming{Uuid: "third", accountIDs: utils.StringMap{"1001": true, "1002": true}},
	}
	eOut := len(*atpl)
	rcv := atpl.Len()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}
func TestAtplSwap(t *testing.T) {
	atpl := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "first", accountIDs: utils.StringMap{"1001": true, "1002": true}},
		&ActionTiming{Uuid: "second", accountIDs: utils.StringMap{"1004": true, "1005": true}},
	}
	atpl2 := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "second", accountIDs: utils.StringMap{"1004": true, "1005": true}},
		&ActionTiming{Uuid: "first", accountIDs: utils.StringMap{"1001": true, "1002": true}},
	}
	atpl.Swap(0, 1)
	if !reflect.DeepEqual(atpl, atpl2) {
		t.Errorf("Expecting: %+v, received: %+v", atpl, atpl2)
	}
}

func TestAtplLess(t *testing.T) {
	atpl := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "first", Weight: 0.07},
		&ActionTiming{Uuid: "second", Weight: 1.07},
	}
	rcv := atpl.Less(1, 0)
	if !rcv {
		t.Errorf("Expecting false, Received: true")
	}
	rcv = atpl.Less(0, 1)
	if rcv {
		t.Errorf("Expecting true, Received: false")
	}
}

func TestAtplSort(t *testing.T) {

	atpl := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "first", accountIDs: utils.StringMap{"1001": true, "1002": true}},
		&ActionTiming{Uuid: "second", accountIDs: utils.StringMap{"1004": true, "1005": true}},
	}
	atpl2 := &ActionTimingWeightOnlyPriorityList{
		&ActionTiming{Uuid: "first", accountIDs: utils.StringMap{"1001": true, "1002": true}},
		&ActionTiming{Uuid: "second", accountIDs: utils.StringMap{"1004": true, "1005": true}},
	}

	sort.Sort(atpl)
	atpl2.Sort()
	if !reflect.DeepEqual(atpl, atpl2) {
		t.Errorf("Expecting: %+v, received: %+v", atpl, atpl2)
	}
}

func TestCacheGetCloned(t *testing.T) {
	at1 := &ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true, "two": true, "three": true},
	}
	Cache.Set(utils.CacheActionPlans, "MYTESTAPL", at1, nil, true, "")
	clned, err := Cache.GetCloned(utils.CacheActionPlans, "MYTESTAPL")
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("Expecting: %+v, received: %+v", at1, at1Cloned)
	}
}

func TestATExecute(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.SchedulerCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	tmpDm := dm
	tmpConn := connMgr
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
		SetDataStorage(tmpDm)
		SetConnManager(tmpConn)
	}()
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.CDRsV1ProcessEvent {

			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): clientConn,
	})
	acs := Actions{{
		Id:               "MINI",
		ActionType:       utils.CDRLOG,
		ExpirationString: utils.UNLIMITED,
		Weight:           10,
		Balance: &BalanceFilter{
			Type: utils.StringPointer(utils.MONETARY),
			Uuid: utils.StringPointer(utils.GenUUID()),
			Value: &utils.ValueFormula{Static: 10,
				Params: make(map[string]any)},
			Weight:   utils.Float64Pointer(10),
			Disabled: utils.BoolPointer(false),
			Timings: []*RITiming{
				{
					Years:     utils.Years{2016, 2017},
					Months:    utils.Months{time.January, time.February, time.March},
					MonthDays: utils.MonthDays{1, 2, 3, 4},
					WeekDays:  utils.WeekDays{1, 2, 3},
					StartTime: utils.ASAP,
				},
			},
			Blocker: utils.BoolPointer(false),
		},
	},
		{ActionType: utils.TOPUP,
			Balance: &BalanceFilter{Type: utils.StringPointer(utils.MONETARY),
				Value:          &utils.ValueFormula{Static: 25},
				DestinationIDs: utils.StringMapPointer(utils.NewStringMap("RET")),
				Weight:         utils.Float64Pointer(20)}},
	}
	dm.SetActions("MINI", acs, utils.NonTransactional)
	at := &ActionTiming{
		Uuid: utils.GenUUID(),
		Timing: &RateInterval{
			Timing: &RITiming{
				Years:     utils.Years{2012},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: utils.ASAP,
			},
		},
		Weight:    10,
		ActionsID: "MINI",
	}
	config.SetCgrConfig(cfg)
	SetDataStorage(dm)
	SetConnManager(connMgr)
	if err := at.Execute(nil, nil); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
}

func TestGetNextStartTimeOld(t *testing.T) {
	at := &ActionTiming{
		Uuid: utils.GenUUID(),
		Timing: &RateInterval{
			Timing: &RITiming{
				Years:     utils.Years{2023},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: utils.ASAP,
			},
		},
		Weight:    10,
		ActionsID: "MINI",
	}

	tests := []struct {
		name     string
		expected time.Time
	}{
		{
			name:     "get next start time 2023 ASAP",
			expected: time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := at.GetNextStartTimeOld(now)
			if !got.Equal(tt.expected) {
				t.Errorf("GetNextStartTimeOld() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestActionPlanCloneNilReturn(t *testing.T) {
	var at *ActionTiming

	rcv := at.Clone()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestActionPlanGetNextStartTimeOld(t *testing.T) {
	zr := time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)
	at := ActionTiming{
		stCache: zr,
	}
	var zr2 time.Time
	at2 := ActionTiming{
		stCache: zr2,
		Timing:  nil,
	}
	var zr3 time.Time
	at3 := ActionTiming{
		stCache: zr3,
		Timing: &RateInterval{
			Timing: &RITiming{
				StartTime: "test",
			},
		},
	}
	var zr4 time.Time
	at4 := ActionTiming{
		stCache: zr4,
		Timing: &RateInterval{
			Timing: &RITiming{
				StartTime:  "",
				Years:      utils.Years{2021},
				Months:     utils.Months{2},
				MonthDays:  utils.MonthDays{2},
				WeekDays:   utils.WeekDays{2},
				EndTime:    "08:09:23",
				cronString: "test",
				tag:        "test",
			},
		},
	}

	tests := []struct {
		name string
		exp  time.Time
	}{
		{
			name: "stCache is not zero",
			exp:  zr,
		},
		{
			name: "timing nil",
			exp:  zr2,
		},
		{
			name: "cannot parse start time",
			exp:  zr3,
		},
		{
			name: "empty timings",
			exp:  at4.GetNextStartTimeOld(time.Now()),
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rcv time.Time
			if i == 0 {
				rcv = at.GetNextStartTimeOld(time.Now())
			} else if i == 1 {
				rcv = at2.GetNextStartTimeOld(time.Now())
			} else if i == 2 {
				rcv = at3.GetNextStartTimeOld(time.Now())
			} else if i == 3 {
				rcv = at4.GetNextStartTimeOld(time.Now())
			}

			if rcv != tt.exp {
				t.Error("received", rcv, "expected", tt.exp)
			}
		})
	}
}
