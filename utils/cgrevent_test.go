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
package utils

import (
	"testing"
	"time"
)

func TestLibSuppliersUsage(t *testing.T) {
	event := make(map[string]interface{})
	event[USAGE] = time.Duration(20 * time.Second)
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.FieldAsDuration(USAGE)
	if err != nil {
		t.Error(err)
	}
	if answ != event[USAGE] {
		t.Errorf("Expecting: %+v, received: %+v", event[USAGE], answ)
	}
	answ, err = seErr.FieldAsDuration(USAGE)
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsTime(t *testing.T) {
	event := make(map[string]interface{})
	event[ANSWER_TIME] = time.Now().Local()
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.FieldAsTime(ANSWER_TIME, "UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != event[ANSWER_TIME] {
		t.Errorf("Expecting: %+v, received: %+v", event[ANSWER_TIME], answ)
	}
	answ, err = seErr.FieldAsTime(ANSWER_TIME, "CET")
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsString(t *testing.T) {
	event := make(map[string]interface{})
	event["supplierprofile1"] = "Supplier"
	event["UsageInterval"] = time.Duration(1 * time.Second)
	event["PddInterval"] = "1s"
	event["Weight"] = 20.0
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierevent1",
		Event:  event,
	}
	answ, err := se.FieldAsString("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != "1s" {
		t.Errorf("Expecting: %+v, received: %+v", event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsString("PddInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != event["PddInterval"] {
		t.Errorf("Expecting: %+v, received: %+v", event["PddInterval"], answ)
	}
	answ, err = se.FieldAsString("supplierprofile1")
	if err != nil {
		t.Error(err)
	}
	if answ != event["supplierprofile1"] {
		t.Errorf("Expecting: %+v, received: %+v", event["supplierprofile1"], answ)
	}
	answ, err = se.FieldAsString("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != "20" {
		t.Errorf("Expecting: %+v, received: %+v", event["Weight"], answ)
	}
}

//float , float in string time duration
