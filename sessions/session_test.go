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

package sessions

import (
	"testing"
	"time"
)

//Test1 ExtraDuration 0 and LastUsage < initial
func TestSRunDebitReserve(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(0),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	//start with extraDuration 0 and the difference go in rDur
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test2 ExtraDuration 0 and LastUsage > initial
func TestSRunDebitReserve2(t *testing.T) {
	lastUsage := time.Duration(2*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(0),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test3 ExtraDuration ( 1m < duration) and LastUsage < initial
func TestSRunDebitReserve3(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != (duration - lastUsage) {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test4 ExtraDuration 1m and LastUsage > initial
func TestSRunDebitReserve4(t *testing.T) {
	lastUsage := time.Duration(2*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	//We have extraDuration 1 minute and 30s different
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != time.Duration(1*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(1*time.Minute+30*time.Second), rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test5 ExtraDuration 3m ( > initialDuration) and LastUsage < initial
func TestSRunDebitReserve5(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(3 * time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	//in debit reserve we start with an extraDuration 3m
	//after we add the different dur-lastUsed (+30s)
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), rDur)
	}
	//ExtraDuration (3m30s - 2m)
	if sr.ExtraDuration != time.Duration(1*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(1*time.Minute+30*time.Second), sr.ExtraDuration)
	}
	if sr.LastUsage != duration {
		t.Errorf("Expecting: %+v, received: %+v", duration, sr.LastUsage)
	}
	if sr.TotalUsage != time.Duration(3*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(3*time.Minute+30*time.Second), sr.TotalUsage)
	}
}
