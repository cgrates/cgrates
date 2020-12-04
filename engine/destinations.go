/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

func NewDestinationFromTPDestination(tpDst *utils.TPDestination) *Destination {
	return &Destination{Id: tpDst.ID, Prefixes: tpDst.Prefixes}

}

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

type DestinationWithOpts struct {
	*Destination
	Tenant string
	Opts   map[string]interface{}
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

// Reverse search in cache to see if prefix belongs to destination id
func CachedDestHasPrefix(destId, prefix string) bool {
	if cached, err := dm.GetReverseDestination(prefix, true, true, utils.NonTransactional); err == nil {
		return utils.IsSliceMember(cached, destId)
	}
	return false
}
