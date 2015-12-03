/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
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

func TestRoundByMethodUp3(t *testing.T) {
	result := Round(0.0701, 2, ROUNDING_UP)
	expected := 0.08
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

func TestParseTimeDetectLayout(t *testing.T) {
	tmStr := "2013-12-30T15:00:01Z"
	expectedTime := time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)
	tm, err := ParseTimeDetectLayout(tmStr, "")
	if err != nil {
		t.Error(err)
	} else if !tm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", tm, expectedTime)
	}
	_, err = ParseTimeDetectLayout(tmStr[1:], "")
	if err == nil {
		t.Errorf("Expecting error")
	}
	sqlTmStr := "2013-12-30 15:00:01"
	sqlTm, err := ParseTimeDetectLayout(sqlTmStr, "")
	if err != nil {
		t.Error(err)
	} else if !sqlTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", sqlTm, expectedTime)
	}
	_, err = ParseTimeDetectLayout(sqlTmStr[1:], "")
	if err == nil {
		t.Errorf("Expecting error")
	}
	unixTmStr := "1388415601"
	unixTm, err := ParseTimeDetectLayout(unixTmStr, "")
	if err != nil {
		t.Error(err)
	} else if !unixTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", unixTm, expectedTime)
	}
	_, err = ParseTimeDetectLayout(unixTmStr[1:], "")
	if err == nil {
		t.Errorf("Expecting error")
	}
	goTmStr := "2013-12-30 15:00:01 +0000 UTC"
	goTm, err := ParseTimeDetectLayout(goTmStr, "")
	if err != nil {
		t.Error(err)
	} else if !goTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", goTm, expectedTime)
	}
	_, err = ParseTimeDetectLayout(goTmStr[1:], "")
	if err == nil {
		t.Errorf("Expecting error")
	}
	goTmStr = "2013-12-30 15:00:01.000000000 +0000 UTC"
	goTm, err = ParseTimeDetectLayout(goTmStr, "")
	if err != nil {
		t.Error(err)
	} else if !goTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", goTm, expectedTime)
	}
	_, err = ParseTimeDetectLayout(goTmStr[1:], "")
	if err == nil {
		t.Errorf("Expecting error")
	}
	fsTmstampStr := "1394291049287234"
	fsTm, err := ParseTimeDetectLayout(fsTmstampStr, "")
	expectedTime = time.Date(2014, 3, 8, 15, 4, 9, 287234000, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !fsTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", fsTm, expectedTime)
	}
	fsTmstampStr = "0"
	fsTm, err = ParseTimeDetectLayout(fsTmstampStr, "")
	expectedTime = time.Time{}
	if err != nil {
		t.Error(err)
	} else if !fsTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", fsTm, expectedTime)
	}
	onelineTmstampStr := "20131023215149"
	olTm, err := ParseTimeDetectLayout(onelineTmstampStr, "")
	expectedTime = time.Date(2013, 10, 23, 21, 51, 49, 0, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !olTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", olTm, expectedTime)
	}
	oneSpaceTmStr := "08.04.2014 22:14:29"
	tsTm, err := ParseTimeDetectLayout(oneSpaceTmStr, "")
	expectedTime = time.Date(2014, 4, 8, 22, 14, 29, 0, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !tsTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", tsTm, expectedTime)
	}
	if nowTm, err := ParseTimeDetectLayout(META_NOW, ""); err != nil {
		t.Error(err)
	} else if time.Now().Sub(nowTm) > time.Duration(10)*time.Millisecond {
		t.Errorf("Unexpected time parsed: %v", nowTm)
	}
	eamonTmStr := "31/05/2015 14:46:00"
	eamonTmS, err := ParseTimeDetectLayout(eamonTmStr, "")
	expectedTime = time.Date(2015, 5, 31, 14, 46, 0, 0, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !eamonTmS.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", eamonTmS, expectedTime)
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
	}{"bevoip.eu", "OUT", "danconns0001", META_PREPAID, "mama"}
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
	}{Tenant: "bevoip.eu", Direction: "OUT", Account: "danconns0001", Type: META_PREPAID}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 1 || missing[0] != "ActionTimingsId" {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestRoundDuration(t *testing.T) {
	minute := time.Minute
	result := RoundDuration(minute, 0*time.Second)
	expected := 0 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute1: expected %v was %v", expected, result)
	}
	result = RoundDuration(time.Second, 1*time.Second+500*time.Millisecond)
	expected = 2 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute1: expected %v was %v", expected, result)
	}
	result = RoundDuration(minute, 1*time.Second)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute2: expected %v was %v", expected, result)
	}
	result = RoundDuration(minute, 5*time.Second)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute3: expected %v was %v", expected, result)
	}
	result = RoundDuration(minute, minute)
	expected = minute
	if result != expected {
		t.Errorf("Error rounding to minute4: expected %v was %v", expected, result)
	}
	result = RoundDuration(minute, 90*time.Second)
	expected = 120 * time.Second
	if result != expected {
		t.Errorf("Error rounding to minute5: expected %v was %v", expected, result)
	}
	result = RoundDuration(60, 120)
	expected = 120.0
	if result != expected {
		t.Errorf("Error rounding to minute5: expected %v was %v", expected, result)
	}
}

func TestRoundAlredyHavingPrecision(t *testing.T) {
	x := 0.07
	if y := Round(x, 2, ROUNDING_UP); y != x {
		t.Error("Error rounding when already has desired precision: ", y)
	}
	if y := Round(x, 2, ROUNDING_MIDDLE); y != x {
		t.Error("Error rounding when already has desired precision: ", y)
	}
	if y := Round(x, 2, ROUNDING_DOWN); y != x {
		t.Error("Error rounding when already has desired precision: ", y)
	}
}

func TestSplitPrefix(t *testing.T) {
	a := SplitPrefix("0123456789", 1)
	if len(a) != 10 {
		t.Error("Error splitting prefix: ", a)
	}
}

func TestSplitPrefixFive(t *testing.T) {
	a := SplitPrefix("0123456789", 5)
	if len(a) != 6 {
		t.Error("Error splitting prefix: ", a)
	}
}

func TestSplitPrefixEmpty(t *testing.T) {
	a := SplitPrefix("", 1)
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
	durStr = "0.002"
	durExpected = time.Duration(2) * time.Millisecond
	if parsed, err := ParseDurationWithSecs(durStr); err != nil {
		t.Error(err)
	} else if parsed != durExpected {
		t.Error("Parsed different than expected")
	}
	durStr = "1.002"
	durExpected = time.Duration(1002) * time.Millisecond
	if parsed, err := ParseDurationWithSecs(durStr); err != nil {
		t.Error(err)
	} else if parsed != durExpected {
		t.Error("Parsed different than expected")
	}
}

func TestMinDuration(t *testing.T) {
	d1, _ := time.ParseDuration("1m")
	d2, _ := time.ParseDuration("59s")
	minD1 := MinDuration(d1, d2)
	minD2 := MinDuration(d2, d1)
	if minD1 != d2 || minD2 != d2 {
		t.Error("Error getting min duration: ", minD1, minD2)
	}
}

func TestParseZeroRatingSubject(t *testing.T) {
	subj := []string{"", "*zero1s", "*zero5m", "*zero10h"}
	dur := []time.Duration{time.Second, time.Second, 5 * time.Minute, 10 * time.Hour}
	for i, s := range subj {
		if d, err := ParseZeroRatingSubject(s); err != nil || d != dur[i] {
			t.Error("Error parsing rating subject: ", s, d, err)
		}
	}
}

func TestConcatenatedKey(t *testing.T) {
	if key := ConcatenatedKey("a"); key != "a" {
		t.Error("Unexpected key value received: ", key)
	}
	if key := ConcatenatedKey("a", "b"); key != fmt.Sprintf("a%sb", CONCATENATED_KEY_SEP) {
		t.Error("Unexpected key value received: ", key)
	}
	if key := ConcatenatedKey("a", "b", "c"); key != fmt.Sprintf("a%sb%sc", CONCATENATED_KEY_SEP, CONCATENATED_KEY_SEP) {
		t.Error("Unexpected key value received: ", key)
	}
}

func TestAvg(t *testing.T) {
	values := []float64{1, 2, 3}
	result := Avg(values)
	expected := 2.0
	if expected != result {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

func TestAvgEmpty(t *testing.T) {
	values := []float64{}
	result := Avg(values)
	expected := 0.0
	if expected != result {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

func TestConvertIfaceToString(t *testing.T) {
	val := interface{}("string1")
	if resVal, converted := ConvertIfaceToString(val); !converted || resVal != "string1" {
		t.Error(resVal, converted)
	}
	val = interface{}(123)
	if resVal, converted := ConvertIfaceToString(val); !converted || resVal != "123" {
		t.Error(resVal, converted)
	}
	val = interface{}([]byte("byte_val"))
	if resVal, converted := ConvertIfaceToString(val); !converted || resVal != "byte_val" {
		t.Error(resVal, converted)
	}
	val = interface{}(true)
	if resVal, converted := ConvertIfaceToString(val); !converted || resVal != "true" {
		t.Error(resVal, converted)
	}
}

func TestMandatory(t *testing.T) {
	_, err := FmtFieldWidth("", 0, "", "", true)
	if err == nil {
		t.Errorf("Failed to detect mandatory value")
	}
}

func TestMaxLen(t *testing.T) {
	result, err := FmtFieldWidth("test", 4, "", "", false)
	expected := "test"
	if err != nil || result != expected {
		t.Errorf("Expected \"test\" was \"%s\"", result)
	}
}

func TestRPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "right", false)
	expected := "test    "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestPaddingFiller(t *testing.T) {
	result, err := FmtFieldWidth("", 8, "", "right", false)
	expected := "        "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "left", false)
	expected := "    test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestZeroLPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "zeroleft", false)
	expected := "0000test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestRStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 2, "right", "", false)
	expected := "te"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXRStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 3, "xright", "", false)
	expected := "tex"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 2, "left", "", false)
	expected := "st"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXLStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 3, "xleft", "", false)
	expected := "xst"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestStripNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("test", 3, "", "", false)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestPaddingNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("test", 5, "", "", false)
	if err == nil {
		t.Error("Expected error")
	}
}
