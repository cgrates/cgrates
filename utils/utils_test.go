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
	"time"
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

func TestParseDateUnix(t *testing.T) {
	date, err := ParseDate("1375212790")
	expected := time.Date(2013, 7, 30, 19, 33, 10, 0, time.UTC)
	if err != nil || !date.Equal(expected) {
		t.Error("error parsing date: ", expected.Sub(date))
	}
}

func TestParseDateUnlimited(t *testing.T) {
	date, err := ParseDate("*unlimited")
	if err != nil || !date.IsZero() {
		t.Error("error parsing unlimited date!: ")
	}
}

func TestParseDateEmpty(t *testing.T) {
	date, err := ParseDate("")
	if err != nil || !date.IsZero() {
		t.Error("error parsing unlimited date!: ")
	}
}

func TestParseDatePlus(t *testing.T) {
	date, err := ParseDate("+20s")
	expected := time.Now()
	if err != nil || date.Sub(expected).Seconds() > 20 || date.Sub(expected).Seconds() < 19 {
		t.Error("error parsing date: ", date.Sub(expected).Seconds())
	}
}

func TestParseDateMonthly(t *testing.T) {
	date, err := ParseDate("*monthly")
	expected := time.Now().AddDate(0, 1, 0)
	if err != nil || expected.Sub(date).Seconds() > 1 {
		t.Error("error parsing date: ", expected.Sub(date).Seconds())
	}
}

func TestParseDateRFC3339(t *testing.T) {
	date, err := ParseDate("2013-07-30T19:33:10Z")
	expected := time.Date(2013, 7, 30, 19, 33, 10, 0, time.UTC)
	if err != nil || !date.Equal(expected) {
		t.Error("error parsing date: ", expected.Sub(date))
	}
}

func TestMissingStructFieldsCorrect(t *testing.T) {
	var attr = struct {
		Tenant          string
		Direction       string
		Account         string
		Type            string
		ActionTimingsId string
	}{"bevoip.eu", "OUT", "danconns0001", "prepaid", "mama"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestMissingStructFieldsIncorrect(t *testing.T) {
	var attr = struct {
		Tenant          string
		Direction       string
		Account         string
		Type            string
		ActionTimingsId string
	}{Tenant: "bevoip.eu", Direction: "OUT", Account: "danconns0001", Type: "prepaid"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 1 || missing[0] != "ActionTimingsId" {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestRound(t *testing.T) {
	minute := time.Minute
	result := RoundTo(minute, 0*time.Second)
	expected := 0 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute1: expected %v was %v", expected, result)
	}
	result = RoundTo(time.Second, 1*time.Second+500*time.Millisecond)
	expected = 2 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute1: expected %v was %v", expected, result)
	}
	result = RoundTo(minute, 1*time.Second)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute2: expected %v was %v", expected, result)
	}
	result = RoundTo(minute, 5*time.Second)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute3: expected %v was %v", expected, result)
	}
	result = RoundTo(minute, minute)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute4: expected %v was %v", expected, result)
	}
	result = RoundTo(minute, 90*time.Second)
	expected = 120 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute5: expected %v was %v", expected, result)
	}
	result = RoundTo(60, 120)
	expected = 120.0
	if result != expected {
		t.Errorf("Error rounding to minute5: expected %v was %v", expected, result)
	}
}

func TestSplitPrefix(t *testing.T) {
	a := SplitPrefix("0123456789")
	if len(a) != 9 {
		t.Error("Error splitting prefix: ", a)
	}
}

func TestSplitPrefixEmpty(t *testing.T) {
	a := SplitPrefix("")
	if len(a) != 0 {
		t.Error("Error splitting prefix: ", a)
	}
}

func TestParseDurationWithSecs(t *testing.T) {
	durStr := "2"
	durExpected := time.Duration(2) * time.Second
	if parsed, err := ParseDurationWithSecs(durStr); err != nil {
		t.Error(err)
	} else if parsed != durExpected {
		t.Error("Parsed different than expected")
	}
	durStr = "2s"
	if parsed, err := ParseDurationWithSecs(durStr); err != nil {
		t.Error(err)
	} else if parsed != durExpected {
		t.Error("Parsed different than expected")
	}
	durStr = "2ms"
	durExpected = time.Duration(2) * time.Millisecond
	if parsed, err := ParseDurationWithSecs(durStr); err != nil {
		t.Error(err)
	} else if parsed != durExpected {
		t.Error("Parsed different than expected")
	}
}
