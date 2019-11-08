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
	"net"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
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

func TestTaskAsMavigableMap(t *testing.T) {
	//empty check
	task := new(Task)
	eOut := config.NewNavigableMap(nil)
	eOut.Set([]string{utils.UUID}, "", false, false)
	eOut.Set([]string{utils.AccountID}, "", false, false)
	eOut.Set([]string{utils.ActionsID}, "", false, false)
	rcv, err := task.AsNavigableMap(nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//normal check
	task = &Task{
		Uuid:      "test",
		AccountID: "test2",
		ActionsID: "test3",
	}
	eOut.Set([]string{utils.UUID}, "test", false, false)
	eOut.Set([]string{utils.AccountID}, "test2", false, false)
	eOut.Set([]string{utils.ActionsID}, "test3", false, false)
	rcv, err = task.AsNavigableMap(nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}

func TestTaskFieldAsinterface(t *testing.T) {
	//empty check
	task := new(Task)
	fldPath := []string{utils.UUID, utils.ActionsID}
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
	fldPath = []string{utils.UUID, "string2"}
	eOut = "test"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//AccountID check
	fldPath = []string{utils.AccountID, "string2"}
	eOut = "test2"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//ActionsID check
	fldPath = []string{utils.ActionsID, "string2"}
	eOut = "test3"
	rcv, err = task.FieldAsString(fldPath)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	//default check
	fldPath = []string{"default", "case"}
	eOut = ""
	rcv, err = task.FieldAsString(fldPath)
	if err == nil {
		t.Error("Expecting NOT_FOUND error, received nil")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}

func TestTaskRemoteHost(t *testing.T) {
	task := new(Task)
	var eOut net.Addr
	rcv := task.RemoteHost()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}
