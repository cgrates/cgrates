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
