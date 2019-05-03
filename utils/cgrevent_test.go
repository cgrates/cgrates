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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLibSuppliersUsage(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			Usage: time.Duration(20 * time.Second),
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.FieldAsDuration(Usage)
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[Usage] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[Usage], answ)
	}
	answ, err = seErr.FieldAsDuration(Usage)
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsTime(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			AnswerTime: time.Now(),
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.FieldAsTime(AnswerTime, "UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[AnswerTime] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[AnswerTime], answ)
	}
	answ, err = seErr.FieldAsTime(AnswerTime, "CET")
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsString(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			"supplierprofile1": "Supplier",
			"UsageInterval":    time.Duration(1 * time.Second),
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
	}
	answ, err := se.FieldAsString("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != "1s" {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsString("PddInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event["PddInterval"] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["PddInterval"], answ)
	}
	answ, err = se.FieldAsString("supplierprofile1")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event["supplierprofile1"] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["supplierprofile1"], answ)
	}
	answ, err = se.FieldAsString("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != "20" {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["Weight"], answ)
	}
}

func TestCGREventFieldAsFloat64(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
	}
	answ, err := se.FieldAsFloat64("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != float64(54.2) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsFloat64("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != float64(20.0) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["Weight"], answ)
	}
	answ, err = se.FieldAsFloat64("PddInterval")
	//TODO: Make an error to be expected :
	//	cgrEvent_test.go:149: strconv.ParseFloat: parsing "1s": invalid syntax
	// if err != nil {
	// 	t.Error(err)
	// }
	if answ != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, answ)
	}

	if _, err := se.FieldAsFloat64(AnswerTime); err == nil || !strings.HasPrefix(err.Error(), "cannot convert field") {
		t.Errorf("Unexpected error : %+v", err)
	}
}
func TestCGREventClone(t *testing.T) {
	now := time.Now()
	ev := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Time:   &now,
		Event: map[string]interface{}{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
	}
	cloned := ev.Clone()
	if !reflect.DeepEqual(ev, cloned) {
		t.Errorf("Expecting: %+v, received: %+v", ev, cloned)
	}
	if cloned.Time == ev.Time {
		t.Errorf("Expecting: different pointer but received: %+v", cloned.Time)
	}

}
