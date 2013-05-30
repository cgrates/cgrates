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
	"math"
	"sort"
)

type MinuteBucket struct {
	Seconds       float64
	Weight        float64
	Price         float64
	Percent       float64 // percentage from standard price
	DestinationId string
	precision     int
}

// Returns the available number of seconds for a specified credit
func (mb *MinuteBucket) GetSecondsForCredit(credit float64) (seconds float64) {
	seconds = mb.Seconds
	if mb.Price > 0 {
		seconds = math.Min(credit/mb.Price, mb.Seconds)
	}
	return
}

// Creates a similar minute
func (mb *MinuteBucket) Clone() *MinuteBucket {
	return &MinuteBucket{
		Seconds:       mb.Seconds,
		Weight:        mb.Weight,
		Price:         mb.Price,
		Percent:       mb.Percent,
		DestinationId: mb.DestinationId,
	}
}

// Equal method
func (mb *MinuteBucket) Equal(o *MinuteBucket) bool {
	return mb.DestinationId == o.DestinationId &&
		mb.Weight == o.Weight &&
		mb.Price == o.Price &&
		mb.Percent == o.Percent
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type bucketsorter []*MinuteBucket

func (bs bucketsorter) Len() int {
	return len(bs)
}

func (bs bucketsorter) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs bucketsorter) Less(j, i int) bool {
	return bs[i].Weight < bs[j].Weight ||
		bs[i].precision < bs[j].precision ||
		bs[i].Price > bs[j].Price
}

func (bs bucketsorter) Sort() {
	sort.Sort(bs)
}
