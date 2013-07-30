/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

func TestFirstNonEmpty(t *testing.T) {
	firstElmnt := ""
	sampleMap := make(map[string]string)
	sampleMap["Third"] = "third"
	fourthElmnt := "fourth"
	winnerElmnt := FirstNonEmpty(firstElmnt, sampleMap["second"], sampleMap["Third"], fourthElmnt)
	if winnerElmnt != sampleMap["Third"] {
		t.Error("Wrong elemnt returned: ", winnerElmnt)
	}
}

func TestUUID(t *testing.T) {
	uuid := GenUUID()
	if len(uuid) == 0 {
		t.Fatalf("GenUUID error %s", uuid)
	}
}

func TestRoundUp(t *testing.T) {
	result := Round(12.52, 0, ROUNDING_UP)
	expected := 13.0
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundUpMiddle(t *testing.T) {
	result := Round(12.5, 0, ROUNDING_UP)
	expected := 13.0
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundDown(t *testing.T) {
	result := Round(12.49, 0, ROUNDING_MIDDLE)
	expected := 12.0
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundPrec(t *testing.T) {
	result := Round(12.49, 1, ROUNDING_UP)
	expected := 12.5
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundPrecNothing(t *testing.T) {
	result := Round(12.49, 2, ROUNDING_MIDDLE)
	expected := 12.49
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundPrecNoTouch(t *testing.T) {
	result := Round(12.49, 2, "")
	expected := 12.49
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundByMethodUp1(t *testing.T) {
	result := Round(12.49, 1, ROUNDING_UP)
	expected := 12.5
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundByMethodUp2(t *testing.T) {
	result := Round(12.21, 1, ROUNDING_UP)
	expected := 12.3
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}

func TestRoundByMethodDown1(t *testing.T) {
	result := Round(12.49, 1, ROUNDING_DOWN)
	expected := 12.4
	if result != expected {
		t.Errorf("Error rounding down: sould be %v was %v", expected, result)
	}
}

func TestRoundByMethodDown2(t *testing.T) {
	result := Round(12.21, 1, ROUNDING_DOWN)
	expected := 12.2
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}
}
