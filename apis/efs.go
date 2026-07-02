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
