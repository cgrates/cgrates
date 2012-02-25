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
	"strconv"
	"strings"
)

/*
Structure describing a tariff plan's number of bonus items. It is uset to restore
these numbers to the user budget every month.
*/
type TariffPlan struct {
	Id            string
	SmsCredit     float64
	Traffic       float64
	MinuteBuckets []*MinuteBucket
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) store() (result string) {
	result += strconv.FormatFloat(tp.SmsCredit, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(tp.Traffic, 'f', -1, 64) + ";"
	for _, mb := range tp.MinuteBuckets {
		var mbs string
		mbs += strconv.Itoa(int(mb.Seconds)) + "|"
		mbs += strconv.Itoa(int(mb.Priority)) + "|"
		mbs += strconv.FormatFloat(mb.Price, 'f', -1, 64) + "|"
		mbs += mb.DestinationId
		result += mbs + ";"
	}
	return
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) restore(input string) {
	elements := strings.Split(input, ";")
	tp.SmsCredit, _ = strconv.ParseFloat(elements[0], 64)
	tp.Traffic, _ = strconv.ParseFloat(elements[1], 64)
	for _, mbs := range elements[2 : len(elements)-1] {
		mb := &MinuteBucket{}
		mbse := strings.Split(mbs, "|")
		mb.Seconds, _ = strconv.ParseFloat(mbse[0], 64)
		mb.Priority, _ = strconv.Atoi(mbse[1])
		mb.Price, _ = strconv.ParseFloat(mbse[2], 64)
		mb.DestinationId = mbse[3]

		tp.MinuteBuckets = append(tp.MinuteBuckets, mb)
	}
}
