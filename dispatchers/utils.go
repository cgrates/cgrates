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

package dispatchers

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/servmanager"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var ( //var used in all tests
	dspDelay   = 1000
	dspDataDir = "/usr/share/cgrates"
	nowTime    = time.Now()
)

type DispatcherEvent struct {
	utils.CGREvent
	*utils.ArgDispatcher
	Subsystem string
}

type ArgsReplicateSessionsWithApiKey struct {
	*utils.ArgDispatcher
	utils.TenantArg
	sessions.ArgsReplicateSessions
}

type AttrRemoteLockWithApiKey struct {
	*utils.ArgDispatcher
	utils.TenantArg
	utils.AttrRemoteLock
}

type AttrRemoteUnlockWithApiKey struct {
	*utils.ArgDispatcher
	utils.TenantArg
	RefID string
}

type StringWithApiKey struct {
	*utils.ArgDispatcher
	utils.TenantArg
	Arg string
}

type ArgStartServiceWithApiKey struct {
	*utils.ArgDispatcher
	utils.TenantArg
	servmanager.ArgStartService
}

func ParseStringMap(s string) utils.StringMap {
	if s == utils.ZERO {
		return make(utils.StringMap)
	}
	return utils.StringMapFromSlice(strings.Split(s, utils.ANDSep))
}
