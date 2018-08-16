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

package v2

import (
	"time"

	"github.com/cgrates/cgrates/apier/v1"
)

type SMGenericV2 struct {
	v1.SMGenericV1
}

// GetMaxUsage returns maxUsage as time.Duration/int64
func (smgv2 *SMGenericV2) GetMaxUsage(ev map[string]interface{},
	maxUsage *time.Duration) error {
	return smgv2.SMG.BiRPCV2GetMaxUsage(nil, ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (smgv2 *SMGenericV2) InitiateSession(ev map[string]interface{},
	maxUsage *time.Duration) error {
	return smgv2.SMG.BiRPCV2InitiateSession(nil, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (smgv2 *SMGenericV2) UpdateSession(ev map[string]interface{},
	maxUsage *time.Duration) error {
	return smgv2.SMG.BiRPCV2UpdateSession(nil, ev, maxUsage)
}

// Called on individual Events (eg SMS)
func (smgv2 *SMGenericV2) ChargeEvent(ev map[string]interface{},
	maxUsage *time.Duration) error {
	return smgv2.SMG.BiRPCV2ChargeEvent(nil, ev, maxUsage)
}
