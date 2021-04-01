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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

type ArgsReplicateSessionsWithAPIOpts struct {
	APIOpts map[string]interface{}
	Tenant  string
	sessions.ArgsReplicateSessions
}

type AttrRemoteLockWithAPIOpts struct {
	APIOpts map[string]interface{}
	Tenant  string
	utils.AttrRemoteLock
}

type AttrRemoteUnlockWithAPIOpts struct {
	APIOpts map[string]interface{}
	Tenant  string
	RefID   string
}

type ArgStartServiceWithAPIOpts struct {
	APIOpts map[string]interface{}
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
