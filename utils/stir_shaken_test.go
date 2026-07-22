// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
