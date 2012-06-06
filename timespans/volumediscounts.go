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
//"log"
)

// Amount of a trafic of a certain type (TOR)
type TrafficVolume struct {
	TOR           string
	Units         float64
	DestinationId string
}

// Volume discount to be applyed after the Units are reached
// in a certain time period.
type VolumeDiscount struct {
	TOR                string
	DestinationsId     string
	Units              float64
	AbsoulteValue      float64 // either this or the procentage below
	DiscountProcentage float64 // use only one
	Weight             float64
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (vd *VolumeDiscount) getDestination(storage StorageGetter) (dest *Destination) {
	if vd.destination == nil {
		vd.destination, _ = storage.GetDestination(vd.DestinationId)
	}
	return vd.destination
}

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Bonus struct {
	Direction      string
	TOR            string
	Units          float64
	balanceMap     map[string]float64
	MinuteBuckets  []*MinuteBucket
	DestinationsId string
	destination    *Destination
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (rcb *Bonus) store() (result string) {
	result += strconv.FormatFloat(rcb.Credit, 'f', -1, 64) + ","
	result += strconv.FormatFloat(rcb.SmsCredit, 'f', -1, 64) + ","
	result += strconv.FormatFloat(rcb.Traffic, 'f', -1, 64)
	if rcb.MinuteBucket != nil {
		result += ","
		result += rcb.MinuteBucket.store()
	}
	return
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (rcb *Bonus) restore(input string) {
	elements := strings.Split(input, ",")
	rcb.Credit, _ = strconv.ParseFloat(elements[0], 64)
	rcb.SmsCredit, _ = strconv.ParseFloat(elements[1], 64)
	rcb.Traffic, _ = strconv.ParseFloat(elements[2], 64)
	if len(elements) > 3 {
		rcb.MinuteBucket = &MinuteBucket{}
		rcb.MinuteBucket.restore(elements[3])
	}
}
