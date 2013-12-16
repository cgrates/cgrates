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

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes map[string]interface{}
}

func (d *Destination) containsPrefix(prefix string) (precision int) {
	if d == nil {
		return
	}
	for i, p := range utils.SplitPrefix(prefix) {
		if _, found := d.Prefixes[p]; found {
			return len(prefix) - i
		}
	}
	return
}

func (d *Destination) String() (result string) {
	result = d.Id + ": "
	for k, _ := range d.Prefixes {
		result += k + ", "
	}
	result = strings.TrimRight(result, ", ")
	return result
}
