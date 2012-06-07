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
	"bytes"
	"encoding/gob"
)

// Amount of a trafic of a certain type (TOR)
type UnitsCounter struct {
	Direction     string
	TOR           string
	Units         float64
	Weight        float64
	DestinationId string
	destination   *Destination
}

// Structure to store actions according to weight
type countersorter []*UnitsCounter

func (s countersorter) Len() int {
	return len(s)
}

func (s countersorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s countersorter) Less(j, i int) bool {
	return s[i].Weight < s[j].Weight
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (uc *UnitsCounter) getDestination() (dest *Destination) {
	if uc.destination == nil {
		uc.destination, _ = storageGetter.GetDestination(uc.DestinationId)
	}
	return uc.destination
}

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	Direction      string
	TOR            string
	Units          float64
	balanceMap     map[string]float64
	MinuteBuckets  []*MinuteBucket
	Weight         float64
	DestinationsId string
	destination    *Destination
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (a *Action) store() (result string) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(a)
	return buf.String()
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (a *Action) restore(input string) {
	gob.NewDecoder(bytes.NewBuffer([]byte(input))).Decode(a)
}

// Structure to store actions according to weight
type actionsorter []*Action

func (s actionsorter) Len() int {
	return len(s)
}

func (s actionsorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s actionsorter) Less(j, i int) bool {
	return s[i].Weight < s[j].Weight
}
