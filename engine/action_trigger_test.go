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
	"encoding/json"
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
		MinSleep:          time.Duration(10),
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
		MinSleep:          time.Duration(10),
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

func TestActionTriggerExecute(t *testing.T) {
	at := ActionTrigger{
		Recurrent:         true,
		LastExecutionTime: time.Now(),
		MinSleep:          1 * time.Hour,
	}
	at2 := ActionTrigger{}

	tests := []struct {
		name string
		arg  *Account
		err  string
	}{
		{
			name: "min sleep time",
			arg:  &Account{},
			err:  "",
		},
		{
			name: "account disabled",
			arg: &Account{
				Disabled: true,
			},
			err: "User  is disabled and there are triggers in action!",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if i == 0 {
				err = at.Execute(tt.arg)
			} else if i == 1 {
				err = at2.Execute(tt.arg)
			}

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}

func TestActionCreateBalance(t *testing.T) {
	str := "test"
	fl := 1.4
	tm := time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)
	sm := utils.StringMap{"test": true}
	rt := RITiming{
		Years:      utils.Years{2019},
		Months:     utils.Months{1},
		MonthDays:  utils.MonthDays{6},
		WeekDays:   utils.WeekDays{5},
		StartTime:  str,
		EndTime:    str,
		cronString: str,
		tag:        str,
	}
	bl := true

	at := ActionTrigger{
		UniqueID: "test",
		Balance: &BalanceFilter{
			Uuid: &str,
			ID:   &str,
			Type: &str,
			Value: &utils.ValueFormula{
				Method: str,
				Params: map[string]any{"test": 1},
				Static: fl,
			},
			ExpirationDate: &tm,
			Weight:         &fl,
			DestinationIDs: &sm,
			RatingSubject:  &str,
			Categories:     &sm,
			SharedGroups:   &sm,
			TimingIDs:      &sm,
			Timings:        []*RITiming{&rt},
			Disabled:       &bl,
			Factor:         &ValueFactor{},
			Blocker:        &bl,
		},
	}

	rcv := at.CreateBalance()

	exp := &Balance{
		Uuid:           str,
		ID:             str,
		Value:          0.0,
		ExpirationDate: tm,
		Weight:         fl,
		DestinationIDs: sm,
		RatingSubject:  str,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        []*RITiming{&rt},
		TimingIDs:      sm,
		Disabled:       bl,
		Factor:         nil,
		Blocker:        bl,
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(rcv), utils.ToJSON(exp))
	}
}

func TestActionTriggerEquals(t *testing.T) {
	str := "test"
	fl := 1.4
	tm := time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)
	sm := utils.StringMap{"test": true}
	rt := RITiming{
		Years:      utils.Years{2019},
		Months:     utils.Months{1},
		MonthDays:  utils.MonthDays{6},
		WeekDays:   utils.WeekDays{5},
		StartTime:  str,
		EndTime:    str,
		cronString: str,
		tag:        str,
	}
	bl := true

	at := ActionTrigger{
		ID:       "test",
		UniqueID: "test",
		Balance: &BalanceFilter{
			Uuid: &str,
			ID:   &str,
			Type: &str,
			Value: &utils.ValueFormula{
				Method: str,
				Params: map[string]any{"test": 1},
				Static: fl,
			},
			ExpirationDate: &tm,
			Weight:         &fl,
			DestinationIDs: &sm,
			RatingSubject:  &str,
			Categories:     &sm,
			SharedGroups:   &sm,
			TimingIDs:      &sm,
			Timings:        []*RITiming{&rt},
			Disabled:       &bl,
			Factor:         &ValueFactor{},
			Blocker:        &bl,
		},
	}

	arg := ActionTrigger{
		ID:       "test1",
		UniqueID: "test1",
		Balance: &BalanceFilter{
			Uuid: &str,
			ID:   &str,
			Type: &str,
			Value: &utils.ValueFormula{
				Method: str,
				Params: map[string]any{"test": 1},
				Static: fl,
			},
			ExpirationDate: &tm,
			Weight:         &fl,
			DestinationIDs: &sm,
			RatingSubject:  &str,
			Categories:     &sm,
			SharedGroups:   &sm,
			TimingIDs:      &sm,
			Timings:        []*RITiming{&rt},
			Disabled:       &bl,
			Factor:         &ValueFactor{},
			Blocker:        &bl,
		},
	}

	rcv := at.Equals(&arg)

	if rcv != false {
		t.Error(rcv)
	}
}

func TestActionTriggerMatch(t *testing.T) {
	strT := "*string"
	str := "test"

	at := ActionTrigger{
		UniqueID: "test",
		Balance: &BalanceFilter{
			Type: &strT,
		},
	}

	tj := struct {
		GroupID       string
		UniqueID      string
		ThresholdType string
	}{
		GroupID:       "",
		UniqueID:      str,
		ThresholdType: str,
	}

	path, err := json.Marshal(tj)
	if err != nil {
		t.Error(err)
	}

	a := Action{
		Balance: &BalanceFilter{
			Type: &strT,
		},
		ExtraParameters: string(path),
	}

	rcv := at.Match(&a)

	if rcv != true {
		t.Error(rcv)
	}
}
