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

func TestActionErrors(t *testing.T) {
	tests := []struct {
		name string
		rcv  string
		exp  string
	}{
		{
			name: "resetTriggersAction",
			rcv:  resetTriggersAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  setRecurrentAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  unsetRecurrentAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  allowNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  denyNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  resetAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  topupResetAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  debitResetAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  debitAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  resetCountersAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  genericDebit(nil, nil, false).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  enableAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  disableAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  topupZeroNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  setExpiryAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  publishAccount(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  publishBalance(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rcv != tt.exp {
				t.Errorf("expected %s, receives %s", tt.exp, tt.rcv)
			}
		})
	}
}

var sm utils.StringMap = utils.StringMap{str: bl}
var rtm RITiming = RITiming{
	Years:      utils.Years{2021},
	Months:     utils.Months{8},
	MonthDays:  utils.MonthDays{28},
	WeekDays:   utils.WeekDays{5},
	StartTime:  str2,
	EndTime:    str2,
	cronString: str2,
	tag:        str2,
}
var vf ValueFactor = ValueFactor{str2: fl2}
var vfr utils.ValueFormula = utils.ValueFormula{
	Method: str2,
	Params: map[string]any{str2: str2},
	Static: fl2,
}
var bf BalanceFilter = BalanceFilter{
	Uuid:           &str2,
	ID:             &str2,
	Type:           &str2,
	Value:          &vfr,
	ExpirationDate: &tm2,
	Weight:         &fl2,
	DestinationIDs: &sm,
	RatingSubject:  &str2,
	Categories:     &sm,
	SharedGroups:   &sm,
	TimingIDs:      &sm,
	Timings:        []*RITiming{&rtm},
	Disabled:       &bl,
	Factor:         &vf,
	Blocker:        &bl,
}
var cf CounterFilter = CounterFilter{
	Value:  fl,
	Filter: &bf,
}
var uc UnitCounter = UnitCounter{
	CounterType: "*balance",
	Counters:    CounterFilters{&cf},
}
var at ActionTrigger = ActionTrigger{
	ID:                str2,
	UniqueID:          str2,
	ThresholdType:     str2,
	Recurrent:         bl,
	MinSleep:          1 * time.Millisecond,
	ExpirationDate:    tm2,
	ActivationDate:    tm2,
	Balance:           &bf,
	Weight:            fl2,
	ActionsID:         str2,
	MinQueuedItems:    nm2,
	Executed:          bl,
	LastExecutionTime: tm2,
}

var ac Account = Account{
	ID:                "test:test",
	UnitCounters:      UnitCounters{str: {&uc}},
	ActionTriggers:    ActionTriggers{&at},
	AllowNegative:     bl,
	Disabled:          bl,
	UpdateTime:        tm2,
	executingTriggers: bl,
}
var a *Action = &Action{
	Id:               str2,
	ActionType:       str2,
	ExtraParameters:  str2,
	Filter:           str2,
	ExpirationString: str2,
	Weight:           fl2,
	Balance:          &bf,
	balanceValue:     fl2,
}

func TestActionClone(t *testing.T) {
	var a *Action

	rcv := a.Clone()

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestActionlogAction(t *testing.T) {
	ub := Account{}
	a := Action{}
	acs := Actions{}

	err := logAction(&ub, &a, acs, nil)
	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	err = logAction(nil, &a, acs, &ub)
	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}
}

func TestActionresetAction(t *testing.T) {

	acs := Actions{}

	err := debitResetAction(&ac, a, acs, nil)
	if err != nil {
		if err.Error() != "" {
			t.Error(err)
		}
	}

	if ac.BalanceMap == nil {
		t.Error("didn't reset action")
	}
}

func TestActiongetOneData(t *testing.T) {
	ub := Account{}

	rcv, err := getOneData(&ub, nil)
	if err != nil {
		t.Error(err)
	}
	exp, _ := json.Marshal(&ub)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}

	rcv, err = getOneData(nil, &ub)
	if err != nil {
		t.Error(err)
	}
	exp, _ = json.Marshal(&ub)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}

	rcv, err = getOneData(nil, nil)
	if err != nil {
		t.Error(err)
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestAccountsendErrors(t *testing.T) {
	tests := []struct {
		name string
		rcv  error
		exp  string
	}{
		{
			name: "sendAMQP error check",
			rcv:  sendAMQP(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "sendAWS error check",
			rcv:  sendAWS(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "sendSQS error check",
			rcv:  sendSQS(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "sendKafka error check",
			rcv:  sendKafka(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "sendS3 error check",
			rcv:  sendS3(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "callURL error check",
			rcv:  callURL(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "callURLAsync error check",
			rcv:  callURLAsync(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
		{
			name: "postEvent error check",
			rcv:  postEvent(nil, nil, nil, make(chan int)),
			exp:  "json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rcv != nil {
				if tt.rcv.Error() != tt.exp {
					t.Errorf("expected %s, received %s", tt.exp, tt.rcv)
				}
			} else {
				t.Error("was expecting an error")
			}
		})
	}
}

func TestActionsetBalanceAction(t *testing.T) {
	err := setBalanceAction(nil, nil, nil, nil)

	if err != nil {
		if err.Error() != `nil account for null action` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}

func TestActiontransferMonetaryDefaultAction(t *testing.T) {
	err := transferMonetaryDefaultAction(nil, nil, nil, nil)

	if err != nil {
		if err.Error() != "ACCOUNT_NOT_FOUND" {
			t.Error(err)
		}
	}

	blc := Balance{
		Uuid:           str2,
		ID:             str2,
		Value:          fl2,
		ExpirationDate: tm2,
		Weight:         fl2,
		DestinationIDs: sm,
		RatingSubject:  str2,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        []*RITiming{&rtm},
		TimingIDs:      sm,
		Disabled:       bl,
		Factor:         ValueFactor{str2: fl2},
		Blocker:        bl,
		precision:      nm2,
		dirty:          bl,
	}
	ac := Account{
		BalanceMap: map[string]Balances{str: {&blc}},
	}

	err = transferMonetaryDefaultAction(&ac, nil, nil, nil)

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Error(err)
		}
	}
}

func TestActionpublishAccount(t *testing.T) {
	sm := utils.StringMap{
		"test1": bl,
	}
	rtm := RITiming{
		Years:      utils.Years{2021},
		Months:     utils.Months{8},
		MonthDays:  utils.MonthDays{28},
		WeekDays:   utils.WeekDays{5},
		StartTime:  str,
		EndTime:    str,
		cronString: str,
		tag:        str,
	}
	blc := Balance{
		Uuid:           str2,
		ID:             str2,
		Value:          fl2,
		ExpirationDate: tm2,
		Weight:         fl2,
		DestinationIDs: sm,
		RatingSubject:  str2,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        []*RITiming{&rtm},
		TimingIDs:      sm,
		Disabled:       bl,
		Factor:         ValueFactor{str2: fl2},
		Blocker:        bl,
		precision:      nm2,
		dirty:          bl,
	}
	ac := Account{
		BalanceMap: map[string]Balances{str: {&blc}},
	}

	err := publishAccount(&ac, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestActionCloneNil(t *testing.T) {
	var apl Actions

	rcv, err := apl.Clone()
	if err != nil {
		t.Error(err)
	}

	if rcv != nil {
		t.Error(err)
	}
}

var cdrP cdrLogProvider = cdrLogProvider{
	acnt:   &ac,
	action: a,
	cache:  utils.MapStorage{},
}

func TestActionString(t *testing.T) {
	rcv := cdrP.String()
	exp := utils.ToJSON(cdrP)

	if rcv != exp {
		t.Errorf("expected %s, received %s", exp, rcv)
	}
}

func TestActionFieldAsInterface(t *testing.T) {
	cdrP := cdrLogProvider{
		acnt:   &ac,
		action: a,
		cache:  utils.MapStorage{},
	}

	tests := []struct {
		name string
		arg  []string
		exp  any
		err  string
	}{
		{
			name: "empty field path",
			arg:  []string{},
			exp:  nil,
			err:  "NOT_FOUND",
		},
		{
			name: "AccountID case",
			arg:  []string{"AccountID"},
			exp:  "test:test",
			err:  "",
		},
		{
			name: "ActionID case",
			arg:  []string{"ActionID"},
			exp:  "test",
			err:  "",
		},
		{
			name: "ActionType case",
			arg:  []string{"ActionType"},
			exp:  "test",
			err:  "",
		},
		{
			name: "BalanceUUID case",
			arg:  []string{"BalanceUUID"},
			exp:  "test",
			err:  "",
		},
		{
			name: "BalanceID case",
			arg:  []string{"BalanceID"},
			exp:  "test",
			err:  "",
		},
		{
			name: "BalanceValue case",
			arg:  []string{"BalanceValue"},
			exp:  "0",
			err:  "",
		},
		{
			name: "DestinationIDs case",
			arg:  []string{"DestinationIDs"},
			exp:  cdrP.action.Balance.DestinationIDs.String(),
			err:  "",
		},
		{
			name: "ExtraParameters case",
			arg:  []string{"ExtraParameters"},
			exp:  cdrP.action.ExtraParameters,
			err:  "",
		},
		{
			name: "RatingSubject case",
			arg:  []string{"RatingSubject"},
			exp:  "test",
			err:  "",
		},
		{
			name: "Category case",
			arg:  []string{"Category"},
			exp:  cdrP.action.Balance.Categories.String(),
			err:  "",
		},
		{
			name: "SharedGroups case",
			arg:  []string{"SharedGroups"},
			exp:  cdrP.action.Balance.SharedGroups.String(),
			err:  "",
		},
		{
			name: "default case",
			arg:  []string{"test"},
			exp:  "test",
			err:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := cdrP.FieldAsInterface(tt.arg)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %v, received %v", tt.exp, rcv)
			}
		})
	}
}

func TestActionFieldAsString(t *testing.T) {
	cdrP.acnt.ID = "test"

	_, err := cdrP.FieldAsString([]string{"AccountID"})

	if err != nil {
		if err.Error() != "Unsupported format for TenantAccount: test" {
			t.Error(err)
		}
	}
}

func TestActionRemoteHost(t *testing.T) {
	rcv := cdrP.RemoteHost()

	if rcv.String() != "local" {
		t.Error(rcv)
	}
}

func TestActionresetAccountErrors(t *testing.T) {
	type args struct {
		ub     *Account
		action *Action
		acts   Actions
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "nil account",
			args: args{ub: nil, action: nil, acts: nil},
			err:  "nil account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resetAccount(tt.args.ub, tt.args.action, tt.args.acts, nil)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}

func TestActionremoveExpired(t *testing.T) {
	sm := utils.StringMap{
		"test1": bl,
	}
	rtm := RITiming{
		Years:      utils.Years{2021},
		Months:     utils.Months{8},
		MonthDays:  utils.MonthDays{28},
		WeekDays:   utils.WeekDays{5},
		StartTime:  str,
		EndTime:    str,
		cronString: str,
		tag:        str,
	}
	blc := Balance{
		Uuid:           str2,
		ID:             str2,
		Value:          fl2,
		ExpirationDate: tm2,
		Weight:         fl2,
		DestinationIDs: sm,
		RatingSubject:  str2,
		Categories:     sm,
		SharedGroups:   sm,
		Timings:        []*RITiming{&rtm},
		TimingIDs:      sm,
		Disabled:       bl,
		Factor:         ValueFactor{str2: fl2},
		Blocker:        bl,
		precision:      nm2,
		dirty:          bl,
	}
	acc := &Account{
		BalanceMap: map[string]Balances{"test1": {&blc}},
	}

	type args struct {
		ub     *Account
		action *Action
		acts   Actions
		extra  any
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "nil account",
			args: args{nil, nil, nil, nil},
			err:  "nil account for null action",
		},
		{
			name: "nil account",
			args: args{acc, a, nil, nil},
			err:  "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := removeExpired(tt.args.ub, tt.args.action, tt.args.acts, nil)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}

func TestActiontopupZeroNegativeAction(t *testing.T) {
	ub := &Account{}

	topupZeroNegativeAction(ub, nil, nil, nil)

	if ub.BalanceMap == nil {
		t.Error("didn't update")
	}
}

func TestActionremoveAccountAction(t *testing.T) {
	ub := &Account{
		ID: "",
	}
	actn := &Action{
		ExtraParameters: str2,
	}
	actn2 := &Action{
		ExtraParameters: "",
	}
	type args struct {
		ub     *Account
		action *Action
		acts   Actions
		extra  any
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "error json unmarshal",
			args: args{nil, actn, nil, nil},
			err:  "invalid character 'e' in literal true (expecting 'r')",
		},
		{
			name: "concatenate key",
			args: args{nil, actn2, nil, nil},
			err:  "",
		},
		{
			name: "accID empty string",
			args: args{ub, actn2, nil, nil},
			err:  "INVALID_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := removeAccountAction(tt.args.ub, tt.args.action, tt.args.acts, tt.args.extra)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}
