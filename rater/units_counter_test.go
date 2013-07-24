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

package rater

import (
	"reflect"
	"testing"
)

func TestUnitsCounterStoreRestore(t *testing.T) {
	uc := &UnitsCounter{
		Direction:     OUTBOUND,
		BalanceId:     SMS,
		Units:         100,
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: ABSOLUTE, DestinationId: "RET"}},
	}
	r, err := uc.Store()
	if err != nil || r != "*out/*sms/100/0;20;1;;NAT,0;10;10;*absolute;RET" {
		t.Errorf("Error serializing units counter: %v", string(r))
	}
	o := &UnitsCounter{}
	err = o.Restore(r)
	if err != nil || !reflect.DeepEqual(o, uc) {
		t.Errorf("Expected %v was  %v", uc, o)
	}
}

func TestUnitsCounterAddMinuteBucket(t *testing.T) {
	uc := &UnitsCounter{
		Direction:     OUTBOUND,
		BalanceId:     SMS,
		Units:         100,
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: ABSOLUTE, DestinationId: "RET"}},
	}
	uc.addMinutes(20, "test")
	if len(uc.MinuteBuckets) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUnitsCounterAddMinuteBucketExists(t *testing.T) {
	uc := &UnitsCounter{
		Direction:     OUTBOUND,
		BalanceId:     SMS,
		Units:         100,
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, PriceType: ABSOLUTE, DestinationId: "RET"}},
	}
	uc.addMinutes(5, "0723")
	if len(uc.MinuteBuckets) != 2 || uc.MinuteBuckets[0].Seconds != 15 {
		t.Error("Error adding minute bucket!")
	}
}
