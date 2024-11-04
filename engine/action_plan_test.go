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
	"sort"
	"testing"
	"time"

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
		t.Errorf("Expecting: %+v,\n received: %+v", at1, at1Cloned)
	}
}

func TestActionTimingClone(t *testing.T) {
	var at *ActionTiming
	val := at.Clone()
	if val != nil {
		t.Errorf("expected nil ,received %v", val)
	}
	at = &ActionTiming{
		Uuid:      "Uuid_test",
		ActionsID: "ActionsID_test",
		Weight:    0.7,
	}
	if cloned := at.Clone(); !reflect.DeepEqual(at, cloned) {
		t.Errorf("Expecting: %+v,\n received: %+v", at, cloned)
	}
}

func TestActionTimindSetActions(t *testing.T) {
	actionTiming := new(ActionTiming)

	actions := Actions{
		&Action{ActionType: "test"},
		&Action{ActionType: "test1"},
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
	if err := Cache.Set(utils.CacheActionPlans, "MYTESTAPL", at1, nil, true, ""); err != nil {
		t.Errorf("Expecting nil, received: %s", err)
	}
	clned, err := Cache.GetCloned(utils.CacheActionPlans, "MYTESTAPL")
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("Expecting: %+v, received: %+v", at1, at1Cloned)
	}
}

func TestActionTimingExErr(t *testing.T) {
	tmpDm := dm
	tmp := Cache
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		dm = tmpDm
		Cache = tmp
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	at := &ActionTiming{
		Timing: &RateInterval{},
		actions: []*Action{
			{
				Filters:          []string{},
				ExpirationString: "*yearly",
				ActionType:       "test",
				Id:               "ZeroMonetary",
				Balance: &BalanceFilter{
					Type: utils.StringPointer(utils.MetaMonetary),

					Value:  &utils.ValueFormula{Static: 11},
					Weight: utils.Float64Pointer(30),
				},
			},
		},
	}
	fltrs := NewFilterS(cfg, nil, dm)
	if err := at.Execute(nil, ""); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
	at.actions[0].ActionType = utils.MetaDebitReset
	if err := at.Execute(nil, ""); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
	at.accountIDs = utils.StringMap{"cgrates.org:zeroNegative": true}
	at.actions[0].ActionType = utils.MetaResetStatQueue
	if err := at.Execute(nil, ""); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
	Cache.Set(utils.CacheFilters, "cgrates.org:*string:~*req.BalanceMap.*monetary[0].ID:*default", nil, []string{}, true, utils.NonTransactional)
	at.actions[0].Filters = []string{"*string:~*req.BalanceMap.*monetary[0].ID:*default"}
	if err := at.Execute(fltrs, ""); err != nil {
		t.Error(err)
	}
	SetDataStorage(nil)
	if err := at.Execute(nil, ""); err != nil {
		t.Error(err)
	}
}

func TestGetDayOrEndOfMonth(t *testing.T) {
	if val := getDayOrEndOfMonth(31, time.Date(2022, 12, 22, 12, 0, 0, 0, time.UTC)); val != 31 {
		t.Errorf("Should Receive Last Day %v", val)
	}
}

func TestActionTimingGetNextStartTimesMonthlyEstimated(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		at       *ActionTiming
		expected time.Time
	}{
		{
			name: "February 7 to February 29",
			t1:   time.Date(2020, 2, 7, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "February 17 to March 16",
			t1:   time.Date(2020, 2, 17, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{16},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "December 17 to January 16",
			t1:   time.Date(2020, 12, 17, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{16},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2021, 1, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "December 17 to December 31",
			t1:   time.Date(2020, 12, 17, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "July 31 same day",
			t1:   time.Date(2020, 7, 31, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "15:00:00",
					},
				},
			},
			expected: time.Date(2020, 7, 31, 15, 0, 0, 0, time.UTC),
		},
		{
			name: "February 17 same day",
			t1:   time.Date(2020, 2, 17, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{17},
						StartTime: "15:00:00",
					},
				},
			},
			expected: time.Date(2020, 2, 17, 15, 0, 0, 0, time.UTC),
		},
		{
			name: "February 17 to March 17",
			t1:   time.Date(2020, 2, 17, 15, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{17},
						StartTime: "10:00:00",
					},
				},
			},
			expected: time.Date(2020, 3, 17, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "September 29 to September 30",
			t1:   time.Date(2020, 9, 29, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 9, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "September 30 to October 31",
			t1:   time.Date(2020, 9, 30, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 10, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "September 30 same day",
			t1:   time.Date(2020, 9, 30, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "15:00:00",
					},
				},
			},
			expected: time.Date(2020, 9, 30, 15, 0, 0, 0, time.UTC),
		},
		{
			name: "September 30 same second",
			t1:   time.Date(2020, 9, 30, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "14:25:01",
					},
				},
			},
			expected: time.Date(2020, 9, 30, 14, 25, 1, 0, time.UTC),
		},
		{
			name: "December 31 same second",
			t1:   time.Date(2020, 12, 31, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "14:25:01",
					},
				},
			},
			expected: time.Date(2020, 12, 31, 14, 25, 1, 0, time.UTC),
		},
		{
			name: "December 31 to January 31",
			t1:   time.Date(2020, 12, 31, 14, 25, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
						StartTime: "14:25:00",
					},
				},
			},
			expected: time.Date(2021, 1, 31, 14, 25, 0, 0, time.UTC),
		},
		// {
		// 	name: "Non-Leap Year: January 29 to Feb 28",
		// 	t1:   time.Date(2021, 1, 29, 0, 0, 0, 0, time.UTC),
		// 	at: &ActionTiming{
		// 		Timing: &RateInterval{
		// 			Timing: &RITiming{
		// 				ID:        utils.MetaMonthlyEstimated,
		// 				MonthDays: utils.MonthDays{29},
		// 			},
		// 		},
		// 	},
		// 	expected: time.Date(2021, 2, 28, 0, 0, 0, 0, time.UTC),
		// },

		{
			name: "Non-Leap Year: February 28 to March 28",
			t1:   time.Date(2021, 2, 28, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{28},
					},
				},
			},
			expected: time.Date(2021, 3, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "October 31 to November 30",
			t1:   time.Date(2021, 10, 31, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{30},
					},
				},
			},
			expected: time.Date(2021, 11, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Leap Year: January 29 to February 29",
			t1:   time.Date(2020, 1, 29, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{29},
					},
				},
			},
			expected: time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Leap Year: February 29 to March 29",
			t1:   time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{29},
					},
				},
			},
			expected: time.Date(2020, 3, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Leap Year: February 29 to March 30",
			t1:   time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{30},
						StartTime: "00:00:00",
					},
				},
			},
			expected: time.Date(2020, 3, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "November 30 to December 31",
			t1:   time.Date(2021, 11, 30, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
					},
				},
			},
			expected: time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "September 29 to September 30",
			t1:   time.Date(2021, 9, 29, 0, 0, 0, 0, time.UTC),
			at: &ActionTiming{
				Timing: &RateInterval{
					Timing: &RITiming{
						ID:        utils.MetaMonthlyEstimated,
						MonthDays: utils.MonthDays{31},
					},
				},
			},
			expected: time.Date(2021, 9, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if st := tt.at.GetNextStartTime(tt.t1); !st.Equal(tt.expected) {
				t.Errorf("Expecting: %+v, received: %+v", tt.expected, st)
			}
		})
	}
}
