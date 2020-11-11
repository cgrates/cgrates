/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ers

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func getProcessOptions(opts map[string]interface{}) (proc map[string]interface{}) {
	proc = make(map[string]interface{})
	for k, v := range opts {
		if strings.HasSuffix(k, utils.ProcessedOpt) {
			proc[k[:len(k)-9]] = v
		}
	}
	return
}
