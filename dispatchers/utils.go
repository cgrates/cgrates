// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
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

type RatingPlanCost struct {
	EventCost    *engine.EventCost
	RatingPlanID string
}
