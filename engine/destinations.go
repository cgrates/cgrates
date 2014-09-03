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
	"encoding/json"
	"strings"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/history"
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

// returns prefix precision
func (d *Destination) containsPrefix(prefix string) int {
	if d == nil {
		return 0
	}
	for _, p := range d.Prefixes {
		if strings.Index(prefix, p) == 0 {
			return len(p)
		}
	}
	return 0
}

func (d *Destination) String() (result string) {
	result = d.Id + ": "
	for _, k := range d.Prefixes {
		result += k + ", "
	}
	result = strings.TrimRight(result, ", ")
	return result
}

func (d *Destination) AddPrefix(pfx string) {
	d.Prefixes = append(d.Prefixes, pfx)
}

// history record method
func (d *Destination) GetHistoryRecord() history.Record {
	js, _ := json.Marshal(d)
	return history.Record{
		Id:       d.Id,
		Filename: history.DESTINATIONS_FN,
		Payload:  js,
	}
}

// Reverse search in cache to see if prefix belongs to destination id
func CachedDestHasPrefix(destId, prefix string) bool {
	if cached, err := cache2go.GetCached(DESTINATION_PREFIX + prefix); err == nil {
		for _, cachedDstId := range cached.([]interface{}) {
			if destId == cachedDstId {
				return true
			}
		}
	}
	return false
}

func CleanStalePrefixes(destIds []string) {
	prefixMap, err := cache2go.GetAllEntries(DESTINATION_PREFIX)
	if err != nil {
		return
	}
	for prefix, idIDs := range prefixMap {
		dIDs := idIDs.Value().([]interface{})
		changed := false
		for _, searchedDID := range destIds {
			if i, found := utils.GetSliceMemberIndex(utils.ConvertInterfaceSliceToStringSlice(dIDs), searchedDID); found {
				if len(dIDs) == 1 {
					// remove de prefix from cache
					cache2go.RemKey(DESTINATION_PREFIX + prefix)
				} else {
					// delte the testination from list and put the new list in chache
					dIDs[i], dIDs = dIDs[len(dIDs)-1], dIDs[:len(dIDs)-1]
					changed = true
				}
			}
		}
		if changed {
			cache2go.Cache(DESTINATION_PREFIX+prefix, dIDs)
		}
	}
}
