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
	"strings"
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

type destinationCacheMap map[string]*Destination

var (
	DestinationCacheMap = make(destinationCacheMap)
)

// Gets the specified destination from the storage and caches it.
func GetDestination(dId string) (d *Destination, err error) {
	d, exists := DestinationCacheMap[dId]
	if !exists {
		d, err = storageGetter.GetDestination(dId)
		if err == nil && d != nil {
			DestinationCacheMap[dId] = d
		}
	}
	return
}

/*
De-serializes the destination for the storage. Used for key-value storages.
*/
func (d *Destination) containsPrefix(prefix string) (bool, int) {
	for i := len(prefix); i >= MinPrefixLength; {
		for _, p := range d.Prefixes {
			if p == prefix[:i] {
				return true, i
			}
		}
		i--
	}

	return false, 0
}

func (d *Destination) store() (result string) {
	for _, p := range d.Prefixes {
		result += p + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (d *Destination) restore(input string) {
	d.Prefixes = strings.Split(input, ",")
}
