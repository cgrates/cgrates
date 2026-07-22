// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

type ArgsReplicateSessionsWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	sessions.ArgsReplicateSessions
}

type AttrRemoteLockWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	utils.AttrRemoteLock
}

type AttrRemoteUnlockWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	RefID   string
}

type ArgStartServiceWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	servmanager.ArgStartService
}

func ParseStringSet(s string) utils.StringSet {
	if s == utils.MetaZero {
		return make(utils.StringSet)
	}
	return utils.NewStringSet(strings.Split(s, utils.ANDSep))
}

type RatingPlanCost struct {
	EventCost    *engine.EventCost
	RatingPlanID string
}
