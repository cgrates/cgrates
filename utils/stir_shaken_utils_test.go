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
	"bytes"
	"math"
	"reflect"
	"testing"
)

func TestRemoveWhiteSpaces(t *testing.T) {
	strWithWS := `   A	String
	With	White Spaces`
	expected := `AStringWithWhiteSpaces`
	if rply := RemoveWhiteSpaces(strWithWS); rply != expected {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestEncodeBase64JSON(t *testing.T) {
	var args interface{}
	args = math.NaN()
	if _, err := EncodeBase64JSON(args); err == nil {
		t.Errorf("Expected error")
	}
	args = map[string]interface{}{"Q": 1}
	expected := `eyJRIjoxfQ`
	if rply, err := EncodeBase64JSON(args); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %q,received: %q", expected, rply)
	}
}

func TestDecodeBase64JSON(t *testing.T) {
	args := `eyJRIjoxfQ`
	var rply1 string
	if err := DecodeBase64JSON(args, &rply1); err == nil {
		t.Errorf("Expected error")
	}
	var rply2 map[string]interface{}
	expected := map[string]interface{}{"Q": 1.}
	if err := DecodeBase64JSON(args, &rply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply2) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply2))
	}
	args = `eyJRIjoxfQ,`
	if err := DecodeBase64JSON(args, &rply2); err == nil {
		t.Errorf("Expected error")
	}
}

type testErrReader struct{}

func (testErrReader) Read([]byte) (int, error) { return 0, ErrNotFound }

func TestNewECDSAPrvKeyFromReader(t *testing.T) {
	if _, err := NewECDSAPrvKeyFromReader(new(testErrReader)); err == nil {
		t.Errorf("Expected error")
	}
	r := bytes.NewBuffer([]byte("invalid certificate"))
	if _, err := NewECDSAPrvKeyFromReader(r); err == nil {
		t.Errorf("Expected error")
	}
}

func TestNewECDSAPubKeyFromReader(t *testing.T) {
	if _, err := NewECDSAPubKeyFromReader(new(testErrReader)); err == nil {
		t.Errorf("Expected error")
	}
	r := bytes.NewBuffer([]byte("invalid certificate"))
	if _, err := NewECDSAPubKeyFromReader(r); err == nil {
		t.Errorf("Expected error")
	}
}
