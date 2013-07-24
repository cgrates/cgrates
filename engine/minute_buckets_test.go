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

package engine

import (
	"reflect"
	"testing"
)

func TestMinutBucketSortWeight(t *testing.T) {
	mb1 := &MinuteBucket{Weight: 1, precision: 2, Price: 2}
	mb2 := &MinuteBucket{Weight: 2, precision: 1, Price: 1}
	var bs bucketsorter
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by weight!")
	}
}

func TestMinutBucketSortPrecision(t *testing.T) {
	mb1 := &MinuteBucket{Weight: 1, precision: 2, Price: 2}
	mb2 := &MinuteBucket{Weight: 1, precision: 1, Price: 1}
	var bs bucketsorter
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestMinutBucketSortPrice(t *testing.T) {
	mb1 := &MinuteBucket{Weight: 1, precision: 1, Price: 1}
	mb2 := &MinuteBucket{Weight: 1, precision: 1, Price: 2}
	var bs bucketsorter
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by price!")
	}
}

func TestMinutBucketEqual(t *testing.T) {
	mb1 := &MinuteBucket{Weight: 1, precision: 1, Price: 1, PriceType: ABSOLUTE, DestinationId: ""}
	mb2 := &MinuteBucket{Weight: 1, precision: 1, Price: 1, PriceType: ABSOLUTE, DestinationId: ""}
	mb3 := &MinuteBucket{Weight: 1, precision: 1, Price: 2, PriceType: ABSOLUTE, DestinationId: ""}
	if !mb1.Equal(mb2) || mb2.Equal(mb3) {
		t.Error("Equal failure!", mb1, mb2, mb3)
	}
}

func TestMinutBucketClone(t *testing.T) {
	mb1 := &MinuteBucket{Seconds: 1, Weight: 2, Price: 3, PriceType: ABSOLUTE, DestinationId: "5"}
	mb2 := mb1.Clone()
	if mb1 == mb2 || !reflect.DeepEqual(mb1, mb2) {
		t.Error("Cloning failure: ", mb1, mb2)
	}
}
