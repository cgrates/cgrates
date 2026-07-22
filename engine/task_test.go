// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestTaskString(t *testing.T) {
	task := &Task{
		Uuid:      "test",
		AccountID: "test2",
		ActionsID: "test3",
	}
	eOut := "{\"Uuid\":\"test\",\"AccountID\":\"test2\",\"ActionsID\":\"test3\"}"
	rcv := task.String()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}

func TestTaskFieldAsinterface(t *testing.T) {
	//empty check
	task := new(Task)
	fldPath := []string{utils.MetaAct, utils.UUID, utils.ActionsID}
	rcv, err := task.FieldAsInterface(fldPath)
	eOut := ""
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//Uuid check
	task = &Task{
		Uuid:      "test",
		AccountID: "test2",
		ActionsID: "test3",
	}
	rcv, err = task.FieldAsInterface(fldPath)
	eOut = "test"
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}

}

func TestTaskFieldAsString(t *testing.T) {
	//empty check
	task := new(Task)
	eOut := ""
	var fldPath []string // := {"string1","string2"}
	rcv, err := task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//Uuid check
	task = &Task{
		Uuid:      "test",
		AccountID: "test2",
		ActionsID: "test3",
	}
	fldPath = []string{utils.MetaAct, utils.UUID, "string2"}
	eOut = "test"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//AccountID check
	fldPath = []string{utils.MetaAct, utils.AccountID, "string2"}
	eOut = "test2"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//ActionsID check
	fldPath = []string{utils.MetaAct, utils.ActionsID, "string2"}
	eOut = "test3"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//default check
	fldPath = []string{utils.MetaAct, "default", "case"}
	eOut = ""
	rcv, err = task.FieldAsString(fldPath)
	if err == nil {
		t.Error("Expecting NOT_FOUND error, received nil")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	if _, err = task.FieldAsString([]string{"default"}); err == nil {
		t.Error(err)
	}
}

func TestTaskExecute(t *testing.T) {

	task := &Task{
		AccountID: "accID",
	}
	if err := task.Execute(&FilterS{}); err == nil {
		t.Error(err)
	}
}
