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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

const (
	LONG_PREFIX_SLICE_LENGTH = 30
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id              string
	Prefixes        []string
	longPrefixesMap map[string]interface{}
}

// returns prefix precision
func (d *Destination) containsPrefix(prefix string) int {
	if d == nil {
		return 0
	}
	if d.Prefixes != nil {
		for _, p := range d.Prefixes {
			if strings.Index(prefix, p) == 0 {
				return len(p)
			}
		}
	}
	if d.longPrefixesMap != nil {
		for _, p := range utils.SplitPrefix(prefix) {
			if _, found := d.longPrefixesMap[p]; found {
				return len(p)
			}
		}
	}
	return 0
}

func (d *Destination) String() (result string) {
	result = d.Id + ": "
	if d.Prefixes != nil {
		for _, k := range d.Prefixes {
			result += k + ", "
		}
	}
	if d.longPrefixesMap != nil {
		for k, _ := range d.longPrefixesMap {
			result += k + ", "
		}
	}
	result = strings.TrimRight(result, ", ")
	return result
}

func (d *Destination) AddPrefix(pfx string) {
	d.Prefixes = append(d.Prefixes, pfx)
}

func (d *Destination) OptimizePrefixes() {
	if len(d.Prefixes) > LONG_PREFIX_SLICE_LENGTH {
		d.longPrefixesMap = make(map[string]interface{})
		for _, p := range d.Prefixes {
			d.longPrefixesMap[p] = nil
		}
		d.Prefixes = nil
	}
}
