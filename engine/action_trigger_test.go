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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestActionTriggerClone(t *testing.T) {
	var at *ActionTrigger
	if rcv := at.Clone(); at != nil {
		t.Errorf("Expecting : nil, received: %s", utils.ToJSON(rcv))
	}
	at = &ActionTrigger{}
	eOut := &ActionTrigger{}
	if rcv := at.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	ttime := time.Now()
	at = &ActionTrigger{
		ID:                "testID",
		UniqueID:          "testUniqueID",
		ThresholdType:     "testThresholdType",
		ThresholdValue:    0.1,
		Recurrent:         true,
		MinSleep:          10,
		ExpirationDate:    ttime,
		ActivationDate:    ttime.Add(5),
		Weight:            0.1,
		ActionsID:         "testActionsID",
		MinQueuedItems:    7,
		Executed:          true,
		LastExecutionTime: ttime.Add(1),
		Balance: &BalanceFilter{
			Uuid: utils.StringPointer("testUuid"),
		},
	}
	eOut = &ActionTrigger{
		ID:                "testID",
		UniqueID:          "testUniqueID",
		ThresholdType:     "testThresholdType",
		ThresholdValue:    0.1,
		Recurrent:         true,
		MinSleep:          10,
		ExpirationDate:    ttime,
		ActivationDate:    ttime.Add(5),
		Weight:            0.1,
		ActionsID:         "testActionsID",
		MinQueuedItems:    7,
		Executed:          true,
		LastExecutionTime: ttime.Add(1),
		Balance: &BalanceFilter{
			Uuid: utils.StringPointer("testUuid"),
		},
	}
	rcv := at.Clone()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rcv.Balance.Uuid = utils.StringPointer("modified")
	if at.Balance.Uuid == utils.StringPointer("modified") {
		t.Errorf("Expedcting: modified, received %s", utils.ToJSON(at.Balance.Uuid))
	}

}

func TestActionTriggersClone(t *testing.T) {
	var atpl ActionTriggers
	if rcv := atpl.Clone(); rcv != nil {
		t.Errorf("Expecting : nil, received: %s", utils.ToJSON(rcv))
	}
	atpl = make(ActionTriggers, 0)
	eOut := make(ActionTriggers, 0)
	if rcv := atpl.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))

	}
	atpl = []*ActionTrigger{
		{
			ID: "test1",
		},
		{
			ID: "test2",
		},
	}
	eOut = []*ActionTrigger{
		{
			ID: "test1",
		},
		{
			ID: "test2",
		},
	}
	if rcv := atpl.Clone(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestActionTriggerFieldAsInterface(t *testing.T) {
	at := &ActionTrigger{}
	if _, err := at.FieldAsInterface([]string{}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"test"}); err == nil {
		t.Error(err)
	}
	at = &ActionTrigger{
		ID:                "id",
		UniqueID:          "unId",
		ThresholdType:     "*max_balance_counter",
		ThresholdValue:    16.1,
		Recurrent:         true,
		MinSleep:          1 * time.Second,
		ExpirationDate:    time.Date(2023, 02, 22, 1, 0, 0, 0, time.UTC),
		ActivationDate:    time.Date(2022, 02, 22, 1, 0, 0, 0, time.UTC),
		Balance:           &BalanceFilter{},
		Weight:            1.02,
		ActionsID:         "acID",
		MinQueuedItems:    5,
		Executed:          true,
		LastExecutionTime: time.Date(2022, 2, 22, 1, 0, 0, 0, time.UTC),
	}
	if _, err := at.FieldAsInterface([]string{"ID"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"UniqueID"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"UniqueID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ThresholdType"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ThresholdType", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ThresholdValue"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ThresholdValue", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Recurrent"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Recurrent", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"MinSleep"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"MinSleep", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ExpirationDate"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ExpirationDate", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ActivationDate"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ActivationDate", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Balance"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Balance", "test"}); err == nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Weight"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Weight", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ActionsID"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"ActionsID", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"MinQueuedItems"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"MinQueuedItems", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Executed"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"Executed", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"LastExecutionTime"}); err != nil {
		t.Error(err)
	} else if _, err = at.FieldAsInterface([]string{"LastExecutionTime", "test"}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestActionTriggerFieldAsString(t *testing.T) {
	at := &ActionTrigger{
		ThresholdValue: 2.6,
	}
	if _, err := at.FieldAsString([]string{}); err == nil {
		t.Error(err)
	} else if val, err := at.FieldAsString([]string{"ThresholdValue"}); err != nil {
		t.Error(err)
	} else if val != "2.6" {
		t.Errorf("received %v", val)
	}

}
