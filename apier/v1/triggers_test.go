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

package v1

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAttrSetActionTriggerUpdateActionTrigger(t *testing.T) {
	ast := AttrSetActionTrigger{}
	if _, err := ast.UpdateActionTrigger(nil, ""); err == nil || err.Error() != "Empty ActionTrigger" {
		t.Errorf("Expected error \"Empty ActionTrigger\", received: %v", err)
	}
	expErr := utils.NewErrMandatoryIeMissing(utils.GroupID)
	args := &engine.ActionTrigger{}
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error %s , received: %v", expErr, err)
	}
	ast.GroupID = "GroupID"
	expErr = utils.NewErrMandatoryIeMissing("ThresholdType", "ThresholdValue")
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error %s , received: %v", expErr, err)
	}
	tNow := time.Now()
	ast = AttrSetActionTrigger{
		GroupID:  "GroupID",
		UniqueID: "ID",
		ActionTrigger: map[string]interface{}{
			utils.ThresholdType:         "THR",
			utils.ThresholdValue:        10,
			utils.Recurrent:             false,
			utils.Executed:              false,
			utils.MinSleep:              time.Second,
			utils.ExpirationDate:        tNow,
			utils.ActivationDate:        tNow,
			utils.BalanceID:             "*default",
			utils.BalanceType:           "*call",
			utils.BalanceDestinationIds: []interface{}{"DST1", "DST2"},
			utils.BalanceWeight:         10,
			utils.BalanceExpirationDate: tNow,
			utils.BalanceTimingTags:     []string{"*asap"},
			utils.BalanceRatingSubject:  "*zero",
			utils.BalanceCategories:     []string{utils.Call},
			utils.BalanceSharedGroups:   []string{"SHRGroup"},
			utils.BalanceBlocker:        true,
			utils.BalanceDisabled:       false,
			utils.ActionsID:             "ACT1",
			utils.MinQueuedItems:        5,
		},
	}
	exp := &engine.ActionTrigger{
		ID:             "GroupID",
		UniqueID:       "ID",
		ThresholdType:  "THR",
		ThresholdValue: 10,
		Recurrent:      false,
		MinSleep:       time.Second,
		ExpirationDate: tNow,
		ActivationDate: tNow,
		Balance: &engine.BalanceFilter{
			ID:             utils.StringPointer(utils.MetaDefault),
			Type:           utils.StringPointer("*call"),
			ExpirationDate: utils.TimePointer(tNow),
			Weight:         utils.Float64Pointer(10),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("DST1", "DST2")),
			RatingSubject:  utils.StringPointer("*zero"),
			Categories:     utils.StringMapPointer(utils.NewStringMap(utils.Call)),
			SharedGroups:   utils.StringMapPointer(utils.NewStringMap("SHRGroup")),
			TimingIDs:      utils.StringMapPointer(utils.NewStringMap("*asap")),
			Disabled:       utils.BoolPointer(false),
			Blocker:        utils.BoolPointer(true),
		},
		ActionsID:      "ACT1",
		MinQueuedItems: 5,
		Executed:       false,
	}
	if updated, err := ast.UpdateActionTrigger(args, ""); err != nil {
		t.Error(err)
	} else if !updated {
		t.Errorf("Expected to be updated")
	} else if !reflect.DeepEqual(exp, args) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(args))
	}

	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.MinQueuedItems] = "NotInt"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceDisabled] = "NotBool"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceBlocker] = "NotBool"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceSharedGroups] = "notSlice"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceCategories] = "notSlice"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceTimingTags] = "notSlice"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceExpirationDate] = "notTime"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceWeight] = "notFloat"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.BalanceDestinationIds] = "notSlice"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.ActivationDate] = "notTime"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.ExpirationDate] = "notTime"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.MinSleep] = "notDuration"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.Executed] = "notBool"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.Recurrent] = "notBool"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{}
	ast.ActionTrigger[utils.ThresholdValue] = "notFloat"
	if _, err := ast.UpdateActionTrigger(args, ""); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
	args = &engine.ActionTrigger{
		ID:       "GroupID2",
		UniqueID: "ID2",
	}
	ast = AttrSetActionTrigger{
		GroupID:  "GroupID",
		UniqueID: "ID",
	}
	if updated, err := ast.UpdateActionTrigger(args, ""); err != nil {
		t.Error(err)
	} else if updated {
		t.Errorf("Expected to not be updated")
	}
	args = &engine.ActionTrigger{
		ID:       "GroupID",
		UniqueID: "ID2",
	}
	ast = AttrSetActionTrigger{
		GroupID:  "GroupID",
		UniqueID: "ID",
	}
	if updated, err := ast.UpdateActionTrigger(args, ""); err != nil {
		t.Error(err)
	} else if updated {
		t.Errorf("Expected to not be updated")
	}

}
