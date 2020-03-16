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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/rpcclient"
)

func TestGetStartTime(t *testing.T) {
	startCGRateSTime = time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
	eOut := startCGRateSTime.Format(time.UnixDate)
	rcv := GetStartTime()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	//only check with an empty string
	rcv := FirstNonEmpty(EmptyString)
	if rcv != EmptyString {
		t.Errorf("Expecting an empty string, received: %+v", rcv)
	}
	//normal check
	firstElmnt := ""
	sampleMap := make(map[string]string)
	sampleMap["Third"] = "third"
	fourthElmnt := "fourth"
	winnerElmnt := FirstNonEmpty(firstElmnt, sampleMap["second"], sampleMap["Third"], fourthElmnt)
	if winnerElmnt != sampleMap["Third"] {
		t.Error("Wrong elemnt returned: ", winnerElmnt)
	}
}

func TestSha1(t *testing.T) {
	//empty check
	rcv := Sha1(" ")
	eOut := "b858cb282617fb0956d960215c8e84d1ccf909c6"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", eOut, rcv)
	}
	//normal check
	rcv = Sha1("teststring")
	eOut = "b8473b86d4c2072ca9b08bd28e373e8253e865c4"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", eOut, rcv)
	}
	rcv = Sha1("test1", "test2")
	eOut = "dff964f6e3c1761b6288f5c75c319d36fb09b2b9"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", eOut, rcv)
	}
}

func TestUUID(t *testing.T) {
	uuid := GenUUID()
	if len(uuid) == 0 {
		t.Fatalf("GenUUID error %s", uuid)
	}
	uuid2 := GenUUID()
	if len(uuid2) == 0 {
		t.Fatalf("GenUUID error %s", uuid)
	}
	if uuid == uuid2 {
		t.Error("GenUUID error.")
	}
}

func TestUUIDSha1Prefix(t *testing.T) {
	rcv := UUIDSha1Prefix()
	if len(rcv) != 7 {
		t.Errorf("Expected len: 7, received %d", len(rcv))
	}
	rcv2 := UUIDSha1Prefix()
	if len(rcv2) != 7 {
		t.Errorf("Expected len: 7, received %d", len(rcv))
	}
	if rcv == rcv2 {
		t.Error("UUIDSha1Prefix error")
	}
}

func TestRound(t *testing.T) {
	result := Round(12.49, 1, ROUNDING_UP)
	expected := 12.5
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}

	result = Round(12.21, 1, ROUNDING_UP)
	expected = 12.3
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}

	result = Round(0.0701, 2, ROUNDING_UP)
	expected = 0.08
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}

	result = Round(12.49, 1, ROUNDING_DOWN)
	expected = 12.4
	if result != expected {
		t.Errorf("Error rounding down: sould be %v was %v", expected, result)
	}

	result = Round(12.21, 1, ROUNDING_DOWN)
	expected = 12.2
	if result != expected {
		t.Errorf("Error rounding up: sould be %v was %v", expected, result)
	}

	//AlredyHavingPrecision
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

	result = Round(14.37, 8, ROUNDING_DOWN)
	expected = 14.37
	if result != expected {
		t.Errorf("Expecting: %v, received:  %v", expected, result)
	}
	result = Round(14.37, 8, "ROUNDING_NOWHERE")
	expected = 14.37
	if result != expected {
		t.Errorf("Expecting: %v, received:  %v", expected, result)
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
	fsTmstampStr = "9999999999999999"
	fsTm, err = ParseTimeDetectLayout(fsTmstampStr, "")
	if err == nil {
		t.Error("Error expected: 'value out of range', received nil")
	}
	fsTmstampStr = "1394291049287234286"
	fsTm, err = ParseTimeDetectLayout(fsTmstampStr, "")
	expectedTime = time.Date(2014, 3, 8, 15, 4, 9, 287234286, time.UTC)
	if err != nil {
		t.Error(err)
	} else if !fsTm.Equal(expectedTime) {
		t.Errorf("Unexpected time parsed: %v, expecting: %v", fsTm, expectedTime)
	}
	fsTmstampStr = "9999999999999999999"
	fsTm, err = ParseTimeDetectLayout(fsTmstampStr, "")
	if err == nil {
		t.Error("Error expected: 'value out of range', received nil")
	}
	var nilTime time.Time
	fsTmstampStr = "+9999999999999999999"
	fsTm, err = ParseTimeDetectLayout(fsTmstampStr, "")
	if err == nil {
		t.Error("Error expected: 'value out of range', received nil")
	} else if fsTm != nilTime {
		t.Errorf("Expecting nilTime, received: %+v", fsTm)
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
	expected = time.Now().AddDate(0, 0, 1)
	if date, err := ParseTimeDetectLayout("*daily", ""); err != nil {
		t.Error(err)
	} else if expected.Sub(date).Seconds() > 1 {
		t.Errorf("received: %+v", date)
	}
	expected = time.Now().AddDate(0, 1, 0)
	if date, err := ParseTimeDetectLayout("*monthly", ""); err != nil {
		t.Error(err)
	} else if expected.Sub(date).Seconds() > 1 {
		t.Errorf("received: %+v", date)
	}
	expected = time.Now().AddDate(1, 0, 0)
	if date, err := ParseTimeDetectLayout("*yearly", ""); err != nil {
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
	if date, err := ParseTimeDetectLayout("*month_end+xyz", ""); err == nil {
		t.Error("Expecting error 'time: invalid time duration', received: nil")
	} else if date != nilTime {
		t.Errorf("Expecting nilTime, received: %+v", date)
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

	date, err = ParseTimeDetectLayout("2014-11-25T00:00:00+01:00", "")
	expected = time.Date(2014, 11, 24, 23, 0, 0, 0, time.UTC)
	if err != nil || !date.UTC().Equal(expected.UTC()) {
		t.Errorf("Expecting: %v, received: %v", expected.UTC(), date.UTC())
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

func TestSplitPrefix(t *testing.T) {
	if a := SplitPrefix("0123456789", 1); len(a) != 10 {
		t.Error("Error splitting prefix: ", a)
	}
	if a := SplitPrefix("0123456789", 5); len(a) != 6 {
		t.Error("Error splitting prefix: ", a)
	}
	if a := SplitPrefix("", 1); len(a) != 0 {
		t.Error("Error splitting prefix: ", a)
	}
}

func TestCopyHour(t *testing.T) {
	var src, dst, eOut time.Time
	if rcv := CopyHour(src, dst); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	src = time.Date(2020, time.April, 18, 20, 10, 11, 01, time.UTC)
	dst = time.Date(2019, time.May, 25, 23, 0, 4, 0, time.UTC)
	eOut = time.Date(2019, time.May, 25, 20, 10, 11, 01, time.UTC)
	if rcv := CopyHour(src, dst); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestParseDurationWithSecs(t *testing.T) {
	var durExpected time.Duration
	if rcv, err := ParseDurationWithSecs(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, durExpected) {
		t.Errorf("Expecting: 0s, received: %+v", rcv)
	}
	durStr := "2"
	durExpected = time.Duration(2) * time.Second
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

func TestParseDurationWithNanosecs(t *testing.T) {
	var eOut time.Duration
	if rcv, err := ParseDurationWithNanosecs(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut, _ = time.ParseDuration("-1ns")
	if rcv, err := ParseDurationWithNanosecs(UNLIMITED); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut, _ = time.ParseDuration("28ns")
	if rcv, err := ParseDurationWithNanosecs("28"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
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
	dfltRatingSubject := map[string]string{
		ANY:      "*zero1ns",
		VOICE:    "*zero1s",
		DATA:     "*zero1ns",
		SMS:      "*zero1ns",
		MONETARY: "*zero1ns",
		GENERIC:  "*zero1ns",
	}
	for i, s := range subj {
		if d, err := ParseZeroRatingSubject(VOICE, s, dfltRatingSubject); err != nil || d != dur[i] {
			t.Error("Error parsing rating subject: ", s, d, err)
		}
	}
	if d, err := ParseZeroRatingSubject(DATA, EmptyString, dfltRatingSubject); err != nil || d != time.Nanosecond {
		t.Error("Error parsing rating subject: ", EmptyString, d, err)
	}
	if d, err := ParseZeroRatingSubject(MONETARY, EmptyString, dfltRatingSubject); err != nil || d != time.Nanosecond {
		t.Error("Error parsing rating subject: ", EmptyString, d, err)
	}
	expecting := "malformed rating subject: test"
	if _, err := ParseZeroRatingSubject(MONETARY, "test", dfltRatingSubject); err == nil || err.Error() != expecting {
		t.Errorf("Expecting: %+v, received: %+v ", expecting, err)
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

func TestSplitConcatenatedKey(t *testing.T) {
	key := "test1:test2:test3"
	eOut := []string{"test1", "test2", "test3"}
	if rcv := SplitConcatenatedKey(key); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestInfieldJoin(t *testing.T) {
	if rcv := InfieldJoin(""); rcv != "" {
		t.Errorf("Expecting: empty string, received: %+v", ToJSON(rcv))
	}
	key := "test1;test2;test3"

	eOut := "test1;test2;test3"
	if rcv := InfieldJoin(key); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(eOut), ToJSON(rcv))
	}
	key2 := "test10;test12"

	eOut = "test1;test2;test3;test10;test12"
	if rcv := InfieldJoin(key, key2); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", ToJSON(eOut), ToJSON(rcv))
	}
}

func TestInfieldSplit(t *testing.T) {
	key := "test1;test2;test3"
	eOut := []string{"test1", "test2", "test3"}
	if rcv := InfieldSplit(key); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestFmtFieldWidth(t *testing.T) {
	//Mandatory0
	if _, err := FmtFieldWidth("", "", 0, "", "", true); err == nil {
		t.Errorf("Failed to detect mandatory value")
	}
	//width = 0
	if result, err := FmtFieldWidth("", "test", 0, "", "", false); err != nil {
		t.Error(err)
	} else if result != "test" {
		t.Errorf("Expecting 'test', received %+q", result)
	}
	//MaxLen
	if result, err := FmtFieldWidth("", "test", 4, "", "", false); err != nil {
		t.Error(err)
	} else if result != "test" {
		t.Errorf("Expected \"test\" received: \"%s\"", result)
	}
	//RPadding
	if result, err := FmtFieldWidth("", "test", 8, "", "*right", false); err != nil {
		t.Error(err)
	} else if result != "test    " {
		t.Errorf("Expected <\"test    \"> \" received: \"%s\"", result)
	}
	//PaddingFiller
	expected := "        "
	if result, err := FmtFieldWidth("", "", 8, "", "*right", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//LPadding
	expected = "    test"
	if result, err := FmtFieldWidth("", "test", 8, "", "*left", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//ZeroLPadding
	expected = "0000test"
	if result, err := FmtFieldWidth("", "test", 8, "", "*zeroleft", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//RStrip
	expected = "te"
	if result, err := FmtFieldWidth("", "test", 2, "*right", "", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//XRStrip
	expected = "tex"
	if result, err := FmtFieldWidth("", "test", 3, "*xright", "", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//LStrip
	expected = "st"
	if result, err := FmtFieldWidth("", "test", 2, "*left", "", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//XLStrip
	expected = "xst"
	if result, err := FmtFieldWidth("", "test", 3, "*xleft", "", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
	}
	//StripNotAllowed
	if _, err := FmtFieldWidth("", "test", 3, "", "", false); err == nil {
		t.Error("Expected error")
	}
	//PaddingNotAllowed
	if _, err := FmtFieldWidth("", "test", 5, "", "", false); err == nil {
		t.Error("Expected error")
	}
	//
	expected = "test"
	if result, err := FmtFieldWidth("", "test", 3, "wrong", "", false); err != nil {
		t.Error(err)
	} else if result != expected {
		t.Errorf("Expected \"%s \" received: \"%s\"", expected, result)
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
	v = interface{}((int64)(1))
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "1" {
		t.Errorf("Received: %+v", sOut)
	}
	v = interface{}(true)
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "true" {
		t.Errorf("Received: %+v", sOut)
	}
	v = interface{}([]byte("test"))
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "test" {
		t.Errorf("Received: %+v", sOut)
	}
	v = interface{}(1.2)
	if sOut, casts := CastIfToString(v); !casts {
		t.Error("Does not cast")
	} else if sOut != "1.2" {
		t.Errorf("Received: %+v", sOut)
	}
	//default
	v = interface{}([]string{"test"})
	if _, casts := CastIfToString(v); casts {
		t.Error("Does cast")
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
	eom = GetEndOfMonth(time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC))
	if time.Now().Add(-1).Before(eom) {
		t.Errorf("Expected %v was %v", expected, eom)
	}
}

func TestSizeFmt(t *testing.T) {
	if str := SizeFmt(0, EmptyString); str != "0.0B" {
		t.Errorf("Expecting: 0.0B, received: %+q", str)
	}
	if str := SizeFmt(1023, EmptyString); str != "1023.0B" {
		t.Errorf("Expecting: 1023.0B, received: %+q", str)
	}
	if str := SizeFmt(1024, EmptyString); str != "1.0KiB" {
		t.Errorf("Expecting: 1.0KiB, received: %+q", str)
	}
	if str := SizeFmt(1048575, EmptyString); str != "1024.0KiB" {
		t.Errorf("Expecting: 1024.0KiB, received: %+q", str)
	}
	if str := SizeFmt(1048576, EmptyString); str != "1.0MiB" {
		t.Errorf("Expecting: 1.0MiB, received: %+q", str)
	}
	if str := SizeFmt(1073741823, EmptyString); str != "1024.0MiB" {
		t.Errorf("Expecting: 1024.0MiB, received: %+q", str)
	}
	if str := SizeFmt(1073741824, EmptyString); str != "1.0GiB" {
		t.Errorf("Expecting: 1.0GiB, received: %+q", str)
	}
	if str := SizeFmt(1099511627775, EmptyString); str != "1024.0GiB" {
		t.Errorf("Expecting: 1024.0GiB, received: %+q", str)
	}
	if str := SizeFmt(1099511627776, EmptyString); str != "1.0TiB" {
		t.Errorf("Expecting: 1.0TiB, received: %+q", str)
	}
	if str := SizeFmt(1125899906842623, EmptyString); str != "1024.0TiB" {
		t.Errorf("Expecting: 1024.0TiB, received: %+q", str)
	}
	if str := SizeFmt(1125899906842624, EmptyString); str != "1.0PiB" {
		t.Errorf("Expecting: 1.0PiB, received: %+q", str)
	}
	if str := SizeFmt(1152921504606847000, EmptyString); str != "1.0EiB" {
		t.Errorf("Expecting: 1.0EiB, received: %+q", str)
	}
	if str := SizeFmt(1180591620717411303424, EmptyString); str != "1.0ZiB" {
		t.Errorf("Expecting: 1.0ZiB, received: %+q", str)
	}
	if str := SizeFmt(9000000000000000000000000, EmptyString); str != "7.4YiB" {
		t.Errorf("Expecting: 7.4YiB, received: %+q", str)
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
	hp = HierarchyPath([]string{})
	if hpStr := hp.AsString(EmptyString, true); hpStr != EmptyString {
		t.Errorf("Expecting: %q, received: %q", EmptyString, hpStr)
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

func TestFib(t *testing.T) {
	fib := Fib()
	if tmp := fib(); tmp != 1 {
		t.Error("Expecting: 1, received ", tmp)
	}
	if tmp := fib(); tmp != 1 {
		t.Error("Expecting: 1, received ", tmp)
	}
	if tmp := fib(); tmp != 2 {
		t.Error("Expecting: 2, received ", tmp)
	}
	if tmp := fib(); tmp != 3 {
		t.Error("Expecting: 3, received ", tmp)
	}
	if tmp := fib(); tmp != 5 {
		t.Error("Expecting: 5, received ", tmp)
	}
}

func TestStringPointer(t *testing.T) {
	result := StringPointer("*zero")
	if *result != EmptyString {
		t.Errorf("Expecting: `%+q`, received `%+q`", EmptyString, *result)
	}
	str := "test_string"
	result = StringPointer(str)
	expected := &str
	if *result != *expected {
		t.Errorf("Expecting: %+v, received: %+v", &str, result)
	}
}
func TestIntPointer(t *testing.T) {
	t1 := 14
	result := IntPointer(t1)
	expected := &t1
	if *expected != *result {
		t.Errorf("Expecting: %+v, received: %+v", expected, result)
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

func TestMapStringStringPointer(t *testing.T) {
	mp := map[string]string{"string1": "string2"}
	result := MapStringStringPointer(mp)
	expected := &mp
	if *result == nil {
		t.Errorf("Expected: %+q, received: nil", expected)
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

func TestDurationPointer(t *testing.T) {
	duration := time.Duration(10)
	result := DurationPointer(duration)
	expected := &duration
	if *expected != *result {
		t.Errorf("Expected: %+q, received: %+q", expected, result)
	}
}

func TestToIJSON(t *testing.T) {
	str := "string"
	received := ToIJSON(str)
	expected := "\"string\""
	if !reflect.DeepEqual(received, expected) {
		t.Errorf("Expected: %+q, received: %+q", expected, received)
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

	Fixes for db driver to avoid returning new values in case of errors
`
	eVers := "CGRateS@v0.11.0~dev-20161230184809-73014daa0c1d"
	if vers, err := GetCGRVersion(); err != nil {
		t.Error(err)
	} else if vers != eVers {
		t.Errorf("Expecting: <%s>, received: <%s>", eVers, vers)
	}
	GitLastLog = ""
	if vers, err := GetCGRVersion(); err != nil {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
	GitLastLog = "\n"
	if vers, err := GetCGRVersion(); err == nil || err.Error() != "Building version - error: <EOF> reading line from file" {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
	GitLastLog = `commit . . .
`
	if vers, err := GetCGRVersion(); err == nil || err.Error() != "Building version - cannot extract commit hash" {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
	GitLastLog = `Date: : :
`
	if vers, err := GetCGRVersion(); err == nil || err.Error() != "Building version - cannot split commit date" {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
	GitLastLog = `Date: wrong format
`
	if vers, err := GetCGRVersion(); err == nil || err.Error() != `Building version - error: <parsing time "wrong format" as "Mon Jan 2 15:04:05 2006 -0700": cannot parse "wrong format" as "Mon"> compiling commit date` {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
	GitLastLog = `ommit 73014daa0c1d7edcb532d5fe600b8a20d588cdf8
Author: DanB <danb@cgrates.org>
Date:   Fri Dec 30 19:48:09 2016 +0100
	
	Fixes for db driver to avoid returning new values in case of errors
`
	if vers, err := GetCGRVersion(); err == nil || err.Error() != "Cannot find commitHash or commitDate information" {
		t.Error(err)
	} else if vers != "CGRateS@v0.11.0~dev" {
		t.Errorf("Expecting: <CGRateS@v0.11.0~dev>, received: <%s>", vers)
	}
}

func TestNewTenantID(t *testing.T) {
	eOut := &TenantID{ID: EmptyString}
	if rcv := NewTenantID(EmptyString); *rcv != *eOut {
		t.Errorf("Expecting: %+v, received %+v", eOut, rcv)
	}
	eOut = &TenantID{Tenant: "Test"}
	if rcv := NewTenantID("Test:"); *rcv != *eOut {
		t.Errorf("Expecting: %+v, received %+v", eOut, rcv)
	}
	eOut = &TenantID{Tenant: "cgrates.org", ID: "id"}
	if rcv := NewTenantID("cgrates.org:id"); *rcv != *eOut {
		t.Errorf("Expecting: %+v, received %+v", eOut, rcv)
	}
}

func TestTenantID(t *testing.T) {
	tID := &TenantID{Tenant: EmptyString, ID: EmptyString}
	eOut := ":"
	if rcv := tID.TenantID(); rcv != eOut {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	tID = &TenantID{Tenant: "cgrates.org", ID: "id"}
	eOut = "cgrates.org:id"
	if rcv := tID.TenantID(); rcv != eOut {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}

func TestTenantIDWithCache(t *testing.T) {
	tID := &TenantIDWithCache{Tenant: EmptyString, ID: EmptyString}
	eOut := ":"
	if rcv := tID.TenantID(); rcv != eOut {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
	tID = &TenantIDWithCache{Tenant: "cgrates.org", ID: "id"}
	eOut = "cgrates.org:id"
	if rcv := tID.TenantID(); rcv != eOut {
		t.Errorf("Expecting: %q, received: %q", eOut, rcv)
	}
}

type testRPC struct {
}

func (tRPC *testRPC) Name(args interface{}, reply interface{}) error {
	return errors.New("err_test")
}
func (tRPC *testRPC) V1Name(args interface{}, reply interface{}) error {
	return errors.New("V1_err_test")
}
func TestRPCCall(t *testing.T) {
	if err := RPCCall("wrong", "test", nil, nil); err == nil || err != rpcclient.ErrUnsupporteServiceMethod {
		t.Errorf("Expecting: %+v, received: %+v", rpcclient.ErrUnsupporteServiceMethod, err)
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

func TestReverseString(t *testing.T) {
	if rcv := ReverseString(EmptyString); rcv != EmptyString {
		t.Errorf("Expecting <%+q>, received: <%+q>", EmptyString, rcv)
	}
	if rcv := ReverseString("test"); rcv != "tset" {
		t.Errorf("Expecting <tset>, received: <%+q>", rcv)
	}
}

func TestGetUrlRawArguments(t *testing.T) {
	eOut := map[string]string{}
	if rcv := GetUrlRawArguments(EmptyString); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	if rcv := GetUrlRawArguments("test"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	if rcv := GetUrlRawArguments("test?"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	eOut = map[string]string{"test": "1"}
	if rcv := GetUrlRawArguments("?test=1"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	eOut = map[string]string{"test": "1", "test2": "2"}
	if rcv := GetUrlRawArguments("?test=1&test2=2"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	eOut = map[string]string{"test": "1", "test2": "2", EmptyString: "5"}
	if rcv := GetUrlRawArguments("?test=1&test2=2&=5"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
	eOut = map[string]string{"test": "1", "test2": "2"}
	if rcv := GetUrlRawArguments("?test=1&test2=2&5"); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expectinc: %+v, received: %+v", eOut, rcv)
	}
}

func TestWarnExecTime(t *testing.T) {
	//without Log
	WarnExecTime(time.Now(), "MyTestFunc", time.Duration(1*time.Second))
	//With Log
	WarnExecTime(time.Now(), "MyTestFunc", time.Duration(1*time.Nanosecond))
}

func TestCastRPCErr(t *testing.T) {
	err := errors.New("test")
	if rcv := CastRPCErr(err); rcv != err {
		t.Errorf("Expecting: %+q, received %+q", err, rcv)
	}
	if rcv := CastRPCErr(ErrNoMoreData); rcv.Error() != ErrNoMoreData.Error() {
		t.Errorf("Expecting: %+v, received %+v", ErrNoMoreData.Error(), rcv)
	}
}

func TestRandomInteger(t *testing.T) {
	a := RandomInteger(0, 100)
	b := RandomInteger(0, 100)
	c := RandomInteger(0, 100)
	if a == b && b == c {
		t.Errorf("same result over 3 attempts")
	}
	if a >= 100 || b >= 100 || c >= 100 {
		t.Errorf("one of the numbers equals or above the  max limit")
	}
	if a < 0 || b < 0 || c < 0 {
		t.Errorf("one of the numbers are below min limit")
	}
}

func TestGetPathIndex(t *testing.T) {
	if rcv, _ := GetPathIndex(EmptyString); rcv != EmptyString {
		t.Errorf("Expecting: \"\"(EmptyString), received: \"%+v\"", rcv)
	}
	eOut := "test"
	if rcv, _ := GetPathIndex("test"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	if rcv, index := GetPathIndex("test[10]"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	} else if *index != 10 {
		t.Errorf("Expecting: %+v, received: %+v", 10, *index)
	}
	if rcv, index := GetPathIndex("test[0]"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	} else if *index != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, *index)
	}
	eOut = "test[notanumber]"
	if rcv, index := GetPathIndex("test[notanumber]"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	} else if index != nil {
		t.Errorf("Expecting: nil, received: %+v", *index)
	}
	eOut = "test[]"
	if rcv, index := GetPathIndex("test[]"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	} else if index != nil {
		t.Errorf("Expecting: nil, received: %+v", *index)
	}
	eOut = "[]"
	if rcv, index := GetPathIndex("[]"); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	} else if index != nil {
		t.Errorf("Expecting: nil, received: %+v", *index)
	}
}
