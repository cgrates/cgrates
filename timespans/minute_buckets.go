/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	// "log"
	"math"
)

type MinuteBucket struct {
	Seconds       float64
	Weight        float64
	Price         float64
	Percent       float64 // percentage from standard price
	DestinationId string
	destination   *Destination
	precision     int
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (mb *MinuteBucket) getDestination() (dest *Destination) {
	if mb.destination == nil {
		mb.destination, _ = storageGetter.GetDestination(mb.DestinationId)
	}
	return mb.destination
}

func (mb *MinuteBucket) GetSecondsForCredit(credit float64) (seconds float64) {
	seconds = mb.Seconds
	if mb.Price > 0 {
		seconds = math.Min(credit/mb.Price, mb.Seconds)
	}
	return
}
