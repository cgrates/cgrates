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

	"github.com/cgrates/cgrates/utils"
)

func TestActionTimingTasks(t *testing.T) {
	//empty check
	actionTiming := new(ActionTiming)
	eOut := []*Task{&Task{Uuid: "", ActionsID: ""}}
	rcv := actionTiming.Tasks()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	//multiple check
	actionTiming.ActionsID = "test"
	actionTiming.Uuid = "test"
	actionTiming.accountIDs = utils.StringMap{"1001": true, "1002": true, "1003": true}
	eOut = []*Task{
		&Task{Uuid: "test", AccountID: "1001", ActionsID: "test"},
		&Task{Uuid: "test", AccountID: "1002", ActionsID: "test"},
		&Task{Uuid: "test", AccountID: "1003", ActionsID: "test"},
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
		//ActionTimings: []*ActionTiming{},
	}
	clned, err := at1.Clone()
	if err != nil {
		t.Error(err)
	}
	at1Cloned := clned.(*ActionPlan)
	if !reflect.DeepEqual(at1, at1Cloned) {
		t.Errorf("Expecting: %+v, received: %+v", at1, at1Cloned)
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
