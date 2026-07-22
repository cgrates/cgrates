// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
