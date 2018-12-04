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
	"fmt"
	"reflect"
	"sync"
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
	tmStr = "2016-04-01T02:00:00+02:00"
	expectedTime = time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC)
	tm, err = ParseTimeDetectLayout(tmStr, "")
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
	expectedTime = time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)
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
	loc, err := time.LoadLocation("Asia/Kabul")
	if err != nil {
		t.Error(err)
	}
	expectedTime = time.Date(2013, 12, 30, 15, 0, 1, 0, loc)
	goTmStr2 := "2013-12-30 15:00:01 +0430 +0430"
	goTm, err = ParseTimeDetectLayout(goTmStr2, "")
	if err != nil {
		t.Error(err)
	} else if !goTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", goTm, expectedTime)
	}
	//goTmStr2 = "2013-12-30 15:00:01 +0430"
	//if _, err = ParseTimeDetectLayout(goTmStr2, ""); err != nil {
	//	t.Errorf("Expecting error")
	//}
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
	broadSoftTmStr := "20160419210007.037"
	broadTmS, err := ParseTimeDetectLayout(broadSoftTmStr, "")
	expectedTime = time.Date(2016, 4, 19, 21, 0, 7, 37000000, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !broadTmS.Equal(expectedTime) {
		t.Errorf("Expecting: %v, received: %v", expectedTime, broadTmS)
	}
	astTimestamp := "2016-09-14T19:37:43.665+0000"
	expectedTime = time.Date(2016, 9, 14, 19, 37, 43, 665000000, time.UTC)
	astTMS, err := ParseTimeDetectLayout(astTimestamp, "")
	if err != nil {
		t.Error(err)
	} else if !astTMS.Equal(expectedTime) {
		t.Errorf("Expecting: %v, received: %v", expectedTime, astTMS)
	}
	nowTimeStr := "+24h"
	start := time.Now().Add(time.Duration(23*time.Hour + 59*time.Minute + 58*time.Second))
	end := start.Add(time.Duration(2 * time.Second))
	parseNowTimeStr, err := ParseTimeDetectLayout(nowTimeStr, "")
	if err != nil {
		t.Error(err)
	} else if parseNowTimeStr.After(start) && parseNowTimeStr.Before(end) {
		t.Errorf("Unexpected time parsed: %v", parseNowTimeStr)
	}

	unixTmMilisecStr := "1534176053410"
	expectedTime = time.Date(2018, 8, 13, 16, 00, 53, 410000000, time.UTC)
	unixTm, err = ParseTimeDetectLayout(unixTmMilisecStr, "")
	if err != nil {
		t.Error(err)
	} else if !unixTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", unixTm, expectedTime)
	}

	tmStr = "2005-08-26T14:17:34"
	expectedTime = time.Date(2005, 8, 26, 14, 17, 34, 0, time.UTC)
	tm, err = ParseTimeDetectLayout(tmStr, "")
	if err != nil {
		t.Error(err)
	} else if !tm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", tm, expectedTime)
	}

	expected := time.Date(2013, 7, 30, 19, 33, 10, 0, time.UTC)
	date, err := ParseTimeDetectLayout("1375212790", "")
	if err != nil || !date.Equal(expected) {
		t.Error("error parsing date: ", expected.Sub(date))
	}

	date, err = ParseTimeDetectLayout("*unlimited", "")
	if err != nil || !date.IsZero() {
		t.Error("error parsing unlimited date!: ")
	}

	date, err = ParseTimeDetectLayout("", "")
	if err != nil || !date.IsZero() {
		t.Error("error parsing unlimited date!: ")
	}

	date, err = ParseTimeDetectLayout("+20s", "")
	expected = time.Now()
	if err != nil || date.Sub(expected).Seconds() > 20 || date.Sub(expected).Seconds() < 19 {
		t.Error("error parsing date: ", date.Sub(expected).Seconds())
	}

	expected = time.Now().AddDate(0, 1, 0)
	if date, err := ParseTimeDetectLayout("*monthly", ""); err != nil {
		t.Error(err)
	} else if expected.Sub(date).Seconds() > 1 {
		t.Errorf("received: %+v", date)
	}

	expected = GetEndOfMonth(time.Now())
	if date, err := ParseTimeDetectLayout("*month_end", ""); err != nil {
		t.Error(err)
	} else if !date.Equal(expected) {
		t.Errorf("received: %+v", date)
	}
	expected = GetEndOfMonth(time.Now()).Add(time.Hour).Add(2 * time.Minute)
	if date, err := ParseTimeDetectLayout("*month_end+1h2m", ""); err != nil {
		t.Error(err)
	} else if !date.Equal(expected) {
		t.Errorf("expecting: %+v, received: %+v", expected, date)
	}

	date, err = ParseTimeDetectLayout("2013-07-30T19:33:10Z", "")
	expected = time.Date(2013, 7, 30, 19, 33, 10, 0, time.UTC)
	if err != nil || !date.Equal(expected) {
		t.Error("error parsing date: ", expected.Sub(date))
	}
	date, err = ParseTimeDetectLayout("2016-04-01T02:00:00+02:00", "")
	expected = time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC)
	if err != nil || !date.Equal(expected) {
		t.Errorf("Expecting: %v, received: %v", expected, date)
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
	subj := []string{"", "*zero1024", "*zero1s", "*zero5m", "*zero10h"}
	dur := []time.Duration{time.Second, time.Duration(1024),
		time.Second, 5 * time.Minute, 10 * time.Hour}
	for i, s := range subj {
		if d, err := ParseZeroRatingSubject(VOICE, s); err != nil || d != dur[i] {
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

func TestMandatory(t *testing.T) {
	_, err := FmtFieldWidth("", "", 0, "", "", true)
	if err == nil {
		t.Errorf("Failed to detect mandatory value")
	}
}

func TestMaxLen(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 4, "", "", false)
	expected := "test"
	if err != nil || result != expected {
		t.Errorf("Expected \"test\" was \"%s\"", result)
	}
}

func TestRPadding(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 8, "", "right", false)
	expected := "test    "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestPaddingFiller(t *testing.T) {
	result, err := FmtFieldWidth("", "", 8, "", "right", false)
	expected := "        "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLPadding(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 8, "", "left", false)
	expected := "    test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestZeroLPadding(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 8, "", "zeroleft", false)
	expected := "0000test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestRStrip(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 2, "right", "", false)
	expected := "te"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXRStrip(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 3, "xright", "", false)
	expected := "tex"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLStrip(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 2, "left", "", false)
	expected := "st"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXLStrip(t *testing.T) {
	result, err := FmtFieldWidth("", "test", 3, "xleft", "", false)
	expected := "xst"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestStripNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("", "test", 3, "", "", false)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestPaddingNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("", "test", 5, "", "", false)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestCastIfToString(t *testing.T) {
	v := interface{}("somestr")
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "somestr" {
		t.Errorf("Received: %+v", sOut)
	}
	v = interface{}(1)
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "1" {
		t.Errorf("Received: %+v", sOut)
	}
	v = interface{}(1.2)
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "1.2" {
		t.Errorf("Received: %+v", sOut)
	}
}

func TestEndOfMonth(t *testing.T) {
	eom := GetEndOfMonth(time.Date(2016, time.February, 5, 10, 1, 2, 3, time.UTC))
	expected := time.Date(2016, time.February, 29, 23, 59, 59, 0, time.UTC)
	if !eom.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
	eom = GetEndOfMonth(time.Date(2015, time.February, 5, 10, 1, 2, 3, time.UTC))
	expected = time.Date(2015, time.February, 28, 23, 59, 59, 0, time.UTC)
	if !eom.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
	eom = GetEndOfMonth(time.Date(2016, time.January, 31, 10, 1, 2, 3, time.UTC))
	expected = time.Date(2016, time.January, 31, 23, 59, 59, 0, time.UTC)
	if !eom.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
	eom = GetEndOfMonth(time.Date(2016, time.December, 31, 10, 1, 2, 3, time.UTC))
	expected = time.Date(2016, time.December, 31, 23, 59, 59, 0, time.UTC)
	if !eom.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
	eom = GetEndOfMonth(time.Date(2016, time.July, 31, 23, 59, 59, 0, time.UTC))
	expected = time.Date(2016, time.July, 31, 23, 59, 59, 0, time.UTC)
	if !eom.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
}

func TestParseHierarchyPath(t *testing.T) {
	eHP := HierarchyPath([]string{"Root", "CGRateS"})
	if hp := ParseHierarchyPath("/Root/CGRateS/", ""); !reflect.DeepEqual(hp, eHP) {
		t.Errorf("Expecting: %+v, received: %+v", eHP, hp)
	}
	if hp := ParseHierarchyPath("Root.CGRateS", ""); !reflect.DeepEqual(hp, eHP) {
		t.Errorf("Expecting: %+v, received: %+v", eHP, hp)
	}
}

func TestHierarchyPathAsString(t *testing.T) {
	eStr := "/Root/CGRateS"
	hp := HierarchyPath([]string{"Root", "CGRateS"})
	if hpStr := hp.AsString("/", true); hpStr != eStr {
		t.Errorf("Expecting: %q, received: %q", eStr, hpStr)
	}
}

func TestMaskSuffix(t *testing.T) {
	dest := "+4986517174963"
	if destMasked := MaskSuffix(dest, 3); destMasked != "+4986517174***" {
		t.Error("Unexpected mask applied", destMasked)
	}
	if destMasked := MaskSuffix(dest, -1); destMasked != dest {
		t.Error("Negative maskLen should not modify destination", destMasked)
	}
	if destMasked := MaskSuffix(dest, 0); destMasked != dest {
		t.Error("Zero maskLen should not modify destination", destMasked)
	}
	if destMasked := MaskSuffix(dest, 100); destMasked != "**************" {
		t.Error("High maskLen should return complete mask", destMasked)
	}
}

func TestTimeIs0h(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	if err != nil {
		t.Error("time parsing error")
	}
	result := TimeIs0h(t1)
	if result != false {
		t.Error("time is 0 when it's supposed to be", t1)
	}
}

func TestToJSON(t *testing.T) {
	if outNilObj := ToJSON(nil); outNilObj != "null" {
		t.Errorf("Expecting null, received: <%q>", outNilObj)
	}
}

func TestClone(t *testing.T) {
	a := 15
	var b int
	err := Clone(a, &b)
	if err != nil {
		t.Error("Cloning failed")
	}
	if b != a {
		t.Error("Expected:", a, ", received:", b)
	}
	// Clone from an interface
	c := "mystr"
	ifaceC := interface{}(c)
	clndIface := reflect.Indirect(reflect.New(reflect.TypeOf(ifaceC))).Interface().(string)
	if err := Clone(ifaceC, &clndIface); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ifaceC, clndIface) {
		t.Errorf("Expecting: %+v, received: %+v", ifaceC, clndIface)
	}
}

func TestIntPointer(t *testing.T) {
	t1 := 14
	result := IntPointer(t1)
	expected := &t1
	if *expected != *result {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestInt64Pointer(t *testing.T) {
	var t1 int64 = 19
	result := Int64Pointer(t1)
	expected := &t1
	if *expected != *result {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestFloat64Pointer(t *testing.T) {
	var t1 float64 = 11.5
	result := Float64Pointer(t1)
	expected := &t1
	if *expected != *result {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestBoolPointer(t *testing.T) {
	t1 := true
	result := BoolPointer(t1)
	expected := &t1
	if *expected != *result {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestStringMapPointer(t *testing.T) {
	t1 := map[string]bool{"cgr1": true, "cgr2": true}
	expected := &t1
	result := StringMapPointer(t1)
	if *result == nil {
		t.Error("Expected:", expected, ", received: nil")
	}
}

func TestTimePointer(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	if err != nil {
		t.Error("time parsing error")
	}
	result := TimePointer(t1)
	expected := &t1
	if *expected != *result {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestLen(t *testing.T) {
	t1 := Int64Slice{2112443, 414241412, 41231241}
	result := t1.Len()
	expected := 3
	if result != expected {
		t.Error("Expected:", expected, ", received:", result)
	}
}

func TestSwap(t *testing.T) {
	t1 := Int64Slice{414241412, 41231241}
	t1.Swap(0, 1)
	var expected int64 = 414241412
	if t1[1] != expected {
		t.Error("Expected:", expected, ", received:", t1[1])
	}
}

func TestLess(t *testing.T) {
	t1 := Int64Slice{414241412, 41231241}
	expected := false
	if t1.Less(0, 1) != expected {
		t.Error("Expected:", expected, ", received:", t1.Less(1, 2))
	}
}

func TestCapitalizedMessage(t *testing.T) {
	if capMsg := CapitalizedMessage(ServiceAlreadyRunning); capMsg != "SERVICE_ALREADY_RUNNING" {
		t.Errorf("Received: <%s>", capMsg)
	}
}

func TestGetCGRVersion(t *testing.T) {
	GitLastLog = `commit 73014daa0c1d7edcb532d5fe600b8a20d588cdf8
Author: DanB <danb@cgrates.org>
Date:   Fri Dec 30 19:48:09 2016 +0100

    Fixes for db driver to avoid returning new values in case of errors`
	eVers := "CGRateS 0.9.1~rc8 git+73014da (2016-12-30T19:48:09+01:00)"
	if vers := GetCGRVersion(); vers != eVers {
		t.Errorf("Expecting: <%s>, received: <%s>", eVers, vers)
	}
}

func TestCounter(t *testing.T) {
	var cmax int64 = 10000
	var i int64
	cnter := NewCounter(0, cmax)
	for i = 1; i <= cmax; i++ {
		if i != cnter.Next() {
			t.Error("Counter has unexpected value")
		}
	}
	if cnter.Next() != 0 {
		t.Error("Counter did not reset")
	}
}

func TestCounterConcurrent(t *testing.T) {
	var nmax int64 = 10000
	ch := make(chan int64, nmax)
	wg := new(sync.WaitGroup)
	cnter := NewCounter(0, nmax-1)
	var i int64
	for i = 1; i <= nmax; i++ {
		wg.Add(1)
		go func() {
			ch <- cnter.Next()
			wg.Done()
		}()
	}
	wg.Wait()
	m := make(map[int64]bool)
	for i = 1; i <= nmax; i++ {
		m[<-ch] = true
	}
	for i = 1; i <= nmax-1; i++ {
		if !m[i] {
			t.Errorf("Missing value: %d", i)
		}
	}
	if cnter.Value() != 0 {
		t.Error("Counter was not reseted to 0")
	}
}

func TestGZIPGUnZIP(t *testing.T) {
	src := []byte("CGRateS.org")
	gzipped, err := GZIPContent(src)
	if err != nil {
		t.Fatal(err)
	}
	if dst, err := GUnZIPContent(gzipped); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(src, dst) {
		t.Error("not matching initial source")
	}
}

func TestFFNNewFallbackFileNameFronString(t *testing.T) {
	fileName := "cdr|*http_json_cdr|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid_json|1acce2c9-3f2d-4774-8662-c28872dad515.json"
	eFFN := &FallbackFileName{Module: "cdr",
		Transport:  MetaHTTPjsonCDR,
		Address:    "http://127.0.0.1:12080/invalid_json",
		RequestID:  "1acce2c9-3f2d-4774-8662-c28872dad515",
		FileSuffix: JSNSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
	fileName = "cdr|*http_post|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid|70c53d6d-dbd7-452e-a5bd-36bab59bb9ff.form"
	eFFN = &FallbackFileName{Module: "cdr",
		Transport:  META_HTTP_POST,
		Address:    "http://127.0.0.1:12080/invalid",
		RequestID:  "70c53d6d-dbd7-452e-a5bd-36bab59bb9ff",
		FileSuffix: FormSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
	fileName = "act>*call_url|*http_json|http%3A%2F%2Flocalhost%3A2080%2Flog_warning|f52cf23e-da2f-4675-b36b-e8fcc3869270.json"
	eFFN = &FallbackFileName{Module: "act>*call_url",
		Transport:  MetaHTTPjson,
		Address:    "http://localhost:2080/log_warning",
		RequestID:  "f52cf23e-da2f-4675-b36b-e8fcc3869270",
		FileSuffix: JSNSuffix}
	if ffn, err := NewFallbackFileNameFronString(fileName); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eFFN, ffn) {
		t.Errorf("Expecting: %+v, received: %+v", eFFN, ffn)
	}
}

func TestFFNFallbackFileNameAsString(t *testing.T) {
	eFn := "cdr|*http_json_cdr|http%3A%2F%2F127.0.0.1%3A12080%2Finvalid_json|1acce2c9-3f2d-4774-8662-c28872dad515.json"
	ffn := &FallbackFileName{
		Module:     "cdr",
		Transport:  MetaHTTPjsonCDR,
		Address:    "http://127.0.0.1:12080/invalid_json",
		RequestID:  "1acce2c9-3f2d-4774-8662-c28872dad515",
		FileSuffix: JSNSuffix}
	if ffnStr := ffn.AsString(); ffnStr != eFn {
		t.Errorf("Expecting: <%q>, received: <%q>", eFn, ffnStr)
	}
}
