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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var ( //var used in all tests
	dspDelay   = 1000
	dspDataDir = "/usr/share/cgrates"
	nowTime    = time.Now()
)

type DispatcherResource struct {
	APIKey  string
	RouteID *string // route over previous computed path
}

type CGREvWithApiKey struct {
	DispatcherResource
	utils.CGREvent
}

type TntIDWithApiKey struct {
	utils.TenantID
	DispatcherResource
}

type TntWithApiKey struct {
	utils.TenantArg
	DispatcherResource
}

type ArgsV1ResUsageWithApiKey struct {
	DispatcherResource
	utils.ArgRSv1ResourceUsage
}

type ArgsProcessEventWithApiKey struct {
	DispatcherResource
	engine.ArgsProcessEvent
}

type ArgsAttrProcessEventWithApiKey struct {
	DispatcherResource
	engine.AttrArgsProcessEvent
}

type ArgsGetSuppliersWithApiKey struct {
	DispatcherResource
	engine.ArgsGetSuppliers
}

type ArgsStatProcessEventWithApiKey struct {
	DispatcherResource
	engine.StatsArgsProcessEvent
}

type AuthorizeArgsWithApiKey struct {
	DispatcherResource
	sessions.V1AuthorizeArgs
}

type InitArgsWithApiKey struct {
	DispatcherResource
	sessions.V1InitSessionArgs
}

type ProcessEventWithApiKey struct {
	DispatcherResource
	sessions.V1ProcessEventArgs
}

type TerminateSessionWithApiKey struct {
	DispatcherResource
	sessions.V1TerminateSessionArgs
}

type UpdateSessionWithApiKey struct {
	DispatcherResource
	sessions.V1UpdateSessionArgs
}

type FilterSessionWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	Filters map[string]string
}

type ArgsReplicateSessionsWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	sessions.ArgsReplicateSessions
}

type SessionWithApiKey struct {
	DispatcherResource
	sessions.Session
}

type CallDescriptorWithApiKey struct {
	DispatcherResource
	engine.CallDescriptor
}

type ArgsGetCacheItemIDsWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	engine.ArgsGetCacheItemIDs
}

type ArgsGetCacheItemWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	engine.ArgsGetCacheItem
}

type AttrReloadCacheWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	utils.AttrReloadCache
}

type AttrCacheIDsWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	CacheIDs []string
}

type ArgsGetGroupWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	engine.ArgsGetGroup
}

type AttrRemoteLockWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	utils.AttrRemoteLock
}

type AttrRemoteUnlockWithApiKey struct {
	DispatcherResource
	utils.TenantArg
	RefID string
}

func ParseStringMap(s string) utils.StringMap {
	if s == utils.ZERO {
		return make(utils.StringMap)
	}
	return utils.StringMapFromSlice(strings.Split(s, utils.ANDSep))
}
