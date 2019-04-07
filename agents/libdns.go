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

package agents

import (
	"errors"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

const (
	QueryType   = "QueryType"
	E164Address = "E164"
)

// e164FromNAPTR extracts the E164 address out of a NAPTR name record
func e164FromNAPTR(name string) (e164 string, err error) {
	i := strings.Index(name, ".e164.arpa")
	if i == -1 {
		return "", errors.New("unknown format")
	}
	e164 = utils.ReverseString(
		strings.Replace(name[:i], ".", "", -1))
	return
}
