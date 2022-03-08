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
)

func TestNewPASSporTHeader(t *testing.T) {
	expected := &PASSporTHeader{
		Alg: STIRAlg,
		Ppt: STIRPpt,
		Typ: STIRTyp,
		X5u: "path/to/certificate",
	}
	if rply := NewPASSporTHeader("path/to/certificate"); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestNewPASSporTDestinationsIdentity(t *testing.T) {
	expected := &PASSporTDestinationsIdentity{
		Tn:  []string{"1001"},
		URI: []string{"1002@cgrates.org"},
	}
	if rply := NewPASSporTDestinationsIdentity([]string{"1001"}, []string{"1002@cgrates.org"}); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestNewPASSporTOriginsIdentity(t *testing.T) {
	expected := &PASSporTOriginsIdentity{
		Tn: "1001",
	}
	if rply := NewPASSporTOriginsIdentity("1001", ""); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply))
	}
}

func TestNewPASSporTPayload(t *testing.T) {
	dst := NewPASSporTDestinationsIdentity([]string{"1001"}, nil)
	orig := NewPASSporTOriginsIdentity("1002", "")
	expected := &PASSporTPayload{
		ATTest: "A",
		Dest:   *dst,
		IAT:    0,
		Orig:   *orig,
		OrigID: "123456",
	}
	rply := NewPASSporTPayload("A", "123456", *dst, *orig)
	rply.IAT = 0
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply))
	}
}
