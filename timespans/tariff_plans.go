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
	"bytes"
	"encoding/gob"
)

const (
	// Direction type
	INBOUND  = "IN"
	OUTBOUND = "OUT"
	// Balance types
	CREDIT  = "MONETARY"
	SMS     = "SMS"
	TRAFFIC = "INTERNET"
)

/*
Structure describing a tariff plan's number of bonus items. It is uset to restore
these numbers to the user balance every month.
*/
type TariffPlan struct {
	Id            string
	balanceMap    map[string]float64
	Actions       []*Action
	MinuteBuckets []*MinuteBucket
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) store() (result string) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(tp)
	return buf.String()
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) restore(input string) {
	gob.NewDecoder(bytes.NewBuffer([]byte(input))).Decode(tp)
}
