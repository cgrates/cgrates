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

func TestLibratesCompile(t *testing.T) {
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
