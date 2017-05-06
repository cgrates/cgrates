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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

/*
Various RADIUS helpers here
*/

func radPassesFieldFilter(pkt *radigo.Packet, fieldFilter *utils.RSRField, processorVars map[string]string) (pass bool) {
	if fieldFilter == nil {
		return true
	}
	if val, hasIt := processorVars[fieldFilter.Id]; hasIt { // ProcessorVars have priority
		if fieldFilter.FilterPasses(val) {
			return true
		}
		return false
	}
	return
}
