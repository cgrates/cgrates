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

package config

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// tagInternalConns adds subsystem to internal connections.
func tagInternalConns(conns []string, subsystem string) []string {
	if len(conns) == 0 {
		return conns
	}
	suffix := utils.ConcatenatedKeySep + subsystem
	result := make([]string, len(conns))
	for i, conn := range conns {
		switch conn {
		case utils.MetaInternal, rpcclient.BiRPCInternal:
			result[i] = conn + suffix
		default:
			result[i] = conn
		}
	}
	return result
}

// stripInternalConns resets all internal connection variants to base type (by
// removing the subsystem suffix).
func stripInternalConns(conns []string) []string {
	if len(conns) == 0 {
		return conns
	}
	result := make([]string, len(conns))
	for i, conn := range conns {
		switch {
		case strings.HasPrefix(conn, utils.MetaInternal):
			result[i] = utils.MetaInternal
		case strings.HasPrefix(conn, rpcclient.BiRPCInternal):
			result[i] = rpcclient.BiRPCInternal
		default:
			result[i] = conn
		}
	}
	return result
}
