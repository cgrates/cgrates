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
	"testing"
	"time"

	"github.com/cgrates/cron"
)

func TestLibratesTenantID(t *testing.T) {
	rp := &RateProfile{
		Tenant: "tenant",
		ID:     "testID",
	}
	expected := rp.Tenant + ":" + rp.ID
	received := rp.TenantID()
	if received != expected {
		t.Errorf("\nReceived: <%v>,\nExpected: <%v>", received, expected)
	}
}

func TestLibratesCompilerp(t *testing.T) {
	// empty struct
	rp := &RateProfile{}
	err := rp.Compile()
	if err != nil {
		t.Errorf("\nReceived: <%v>, \nExpected: <%v>", err, nil)
	}

	// non-empty
	fail := "shouldfail"
	rp.ID = "test"
	rp.Tenant = "tenant"
	rp.Rates = map[string]*Rate{
		"testKey1": &Rate{
			ID:              "ID1",
			ActivationTimes: fail,
		},
		"testKey2": &Rate{
			ID: "ID2",
		},
	}

	expected := "expected exactly 5 fields, found 1: [" + fail + "]"
	err = rp.Compile()

	if err == nil || err.Error() != expected {
		t.Errorf("\nReceived: <%v>, \nExpected: <%v>", err, expected)
	}
}

func TestLibratesUID(t *testing.T) {
	rt := &Rate{
		uID: "testString",
	}

	expected := "testString"
	received := rt.UID()

	if received != expected {
		t.Errorf("\nReceived: %q, \nExpected: %q", received, expected)
	}

}

func TestLibratesCompilert(t *testing.T) {
	rt := &Rate{
		ActivationTimes: EmptyString,
	}

	err := rt.Compile()

	if err != nil {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", err, nil)
	}
}

func TestLibratesRunTimes(t *testing.T) {
	var (
		sTime     time.Time
		eTime     time.Time
		verbosity int
	)

	// memory leak test
	verbosity = 0

	rt := &Rate{}

	received, err := rt.RunTimes(sTime, eTime, verbosity)
	var expected [][]time.Time

	if err == nil || err != ErrMaxIterationsReached {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", err, ErrMaxIterationsReached)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", received, expected)
	}

	// aTime after eTime test
	schd, err := cron.ParseStandard("* * * * *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	verbosity = 10
	eTime = sTime.Add(10 * time.Minute)

	received, err = rt.RunTimes(sTime, eTime, verbosity)

	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", received, expected)
	}

	// eTime before iTime test
	schd, err = cron.ParseStandard("* * 12 3 *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	sTime, err = time.Parse(time.RFC3339, "2022-03-11T15:04:05Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}
	eTime = sTime.Add(24 * time.Hour)

	received, err = rt.RunTimes(sTime, eTime, verbosity)

	aT1, err := time.Parse(time.RFC3339, "2022-03-12T00:00:00Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	aT2, err := time.Parse(time.RFC3339, "2022-03-13T00:00:00Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	aTsl := make([]time.Time, 0)
	aTsl = append(aTsl, aT1, aT2)
	expected = append(expected, aTsl)

	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", received, expected)
	}

	//eTime after iTime
	schd, err = cron.ParseStandard("* * 12 3 *")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	rt.sched = schd
	sTime, err = time.Parse(time.RFC3339, "2022-03-11T15:04:05Z")
	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}
	eTime = sTime.Add(48 * time.Hour)

	received, err = rt.RunTimes(sTime, eTime, verbosity)

	if err != nil {
		t.Errorf("\ndidn't expect error, got %v", err)
	}

	if !reflect.DeepEqual(received, expected) {
		t.Errorf("\nReceived: <%+v>, \nExpected: <%+v>", received, expected)
	}
}
