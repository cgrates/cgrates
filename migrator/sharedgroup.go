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
package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1SharedGroup struct {
	Id                string
	AccountParameters map[string]*engine.SharingParameters
	MemberIds         []string
}

func (v1SG v1SharedGroup) AsSharedGroup() (sg engine.SharedGroup) {
	sg = engine.SharedGroup{
		Id:                v1SG.Id,
		AccountParameters: v1SG.AccountParameters,
		MemberIds:         make(utils.StringMap),
	}
	for _, accID := range v1SG.MemberIds {
		sg.MemberIds[accID] = true
	}
	return
}
