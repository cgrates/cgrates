// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/utils"
)

// NewEfSv1 initializes the EfSv1 object.
func NewEfSv1(efS *efs.EfS) *EfSv1 {
	return &EfSv1{efS: efS}
}

// EfSv1 represents the RPC object to register for export failover v1 APIs.
type EfSv1 struct {
	efS *efs.EfS
}

// ProcessEvent writes failed export events.
func (s *EfSv1) ProcessEvent(ctx *context.Context, args *utils.ArgsFailedPosts, reply *string) error {
	return s.efS.V1ProcessEvent(ctx, args, reply)
}

// ReplayEvents replays failed export events.
func (s *EfSv1) ReplayEvents(ctx *context.Context, args efs.ReplayEventsParams, reply *string) error {
	return s.efS.V1ReplayEvents(ctx, args, reply)
}
