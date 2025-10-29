/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
)

func NewSMGenericV1(sS *sessions.SessionS) *SMGenericV1 {
	return &SMGenericV1{
		Ss: sS,
	}
}

// Exports RPC from SMGeneric
// DEPRECATED, use SessionSv1 instead
type SMGenericV1 struct {
	Ss *sessions.SessionS
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (smgv1 *SMGenericV1) GetMaxUsage(ctx *context.Context, ev map[string]any,
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1GetMaxUsage(ctx, ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (smgv1 *SMGenericV1) InitiateSession(ctx *context.Context, ev map[string]any,
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1InitiateSession(ctx, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (smgv1 *SMGenericV1) UpdateSession(ctx *context.Context, ev map[string]any,
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1UpdateSession(ctx, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (smgv1 *SMGenericV1) TerminateSession(ctx *context.Context, ev map[string]any,
	reply *string) error {
	return smgv1.Ss.BiRPCV1TerminateSession(ctx, ev, reply)
}

// Called on session end, should send the CDR to CDRS
func (smgv1 *SMGenericV1) ProcessCDR(ctx *context.Context, ev map[string]any,
	reply *string) error {
	return smgv1.Ss.BiRPCV1ProcessCDR(ctx, ev, reply)
}
