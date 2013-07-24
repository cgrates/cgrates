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
	"github.com/cgrates/cgrates/cache2go"
	"strings"
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

// Gets the specified destination from the storage and caches it.
func GetDestination(dId string) (d *Destination, err error) {
	x, err := cache2go.GetCached(dId)
	if err != nil {
		d, err = storageGetter.GetDestination(dId)
		if err == nil && d != nil {
			cache2go.Cache(dId, d)
		}
	} else {
		d = x.(*Destination)
	}
	return
}

/*
De-serializes the destination for the storage. Used for key-value storages.
*/
func (d *Destination) containsPrefix(prefix string) (precision int, ok bool) {
	if d == nil {
		return
	}
	for _, p := range d.Prefixes {
		if strings.Index(prefix, p) == 0 && len(p) > precision {
			precision = len(p)
			ok = true
		}
	}
	return
}

func (d *Destination) String() (result string) {
	result = d.Id + ": "
	for _, p := range d.Prefixes {
		result += p + ", "
	}
	result = strings.TrimRight(result, ", ")
	return result
}
