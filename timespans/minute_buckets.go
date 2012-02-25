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
	"math"
	"strconv"
	"strings"
)

type MinuteBucket struct {
	Seconds       float64
	Priority      int
	Price         float64
	DestinationId string
	destination   *Destination
	precision     int
}

/*
Serializes the minute bucket for the storage. Used for key-value storages.
*/
func (mb *MinuteBucket) store() (result string) {
	result += strconv.Itoa(int(mb.Seconds)) + "|"
	result += strconv.Itoa(int(mb.Priority)) + "|"
	result += strconv.FormatFloat(mb.Price, 'f', -1, 64) + "|"
	result += mb.DestinationId
	return
}

/*
De-serializes the minute bucket for the storage. Used for key-value storages.
*/
func (mb *MinuteBucket) restore(input string) {
	mbse := strings.Split(input, "|")
	mb.Seconds, _ = strconv.ParseFloat(mbse[0], 64)
	mb.Priority, _ = strconv.Atoi(mbse[1])
	mb.Price, _ = strconv.ParseFloat(mbse[2], 64)
	mb.DestinationId = mbse[3]
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (mb *MinuteBucket) getDestination(storage StorageGetter) (dest *Destination) {
	if mb.destination == nil {
		mb.destination, _ = storage.GetDestination(mb.DestinationId)
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
