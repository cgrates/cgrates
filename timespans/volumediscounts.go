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

type VolumeDiscount struct {
	TOR                string
	DestinationsId     string
	VolumeUnits        float64
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

type Volume struct {
	TOR           string
	Units         float64
	DestinationId string
}

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type InboundBonus struct {
	TOR                                   string
	InboundUnits                          float64
	MonetaryUnits, SMSUnits, TrafficUnits float64
	MinuteBucket                          *MinuteBucket
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (rcb *InboundBonus) store() (result string) {
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
func (rcb *RecivedCallBonus) restore(input string) {
	elements := strings.Split(input, ",")
	rcb.Credit, _ = strconv.ParseFloat(elements[0], 64)
	rcb.SmsCredit, _ = strconv.ParseFloat(elements[1], 64)
	rcb.Traffic, _ = strconv.ParseFloat(elements[2], 64)
	if len(elements) > 3 {
		rcb.MinuteBucket = &MinuteBucket{}
		rcb.MinuteBucket.restore(elements[3])
	}
}

type OutboundBonus struct {
	TOR                                   string
	OutboundUnits                         float64
	DestinationsId                        string
	destination                           *Destination
	MonetaryUnits, SMSUnits, TrafficUnits float64
	MinuteBucket                          *MinuteBucket
}
