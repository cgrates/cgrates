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

package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/utils"
)

func NewEfSv1(efSv1 *efs.EfS) *EfSv1 {
	return &EfSv1{
		efs: efSv1,
	}
}

// EfSv1 export RPC calls for EventFailover Service
type EfSv1 struct {
	efs *efs.EfS
	ping
}

// ProcessEvent will write into gob formnat file the Events that were failed to be exported.
func (efS *EfSv1) ProcessEvent(ctx *context.Context, args *utils.ArgsFailedPosts, reply *string) error {
	return efS.efs.V1ProcessEvent(ctx, args, reply)
}

// ReplayEvents will read the Events from gob files that were failed to be exported and try to re-export them again.
func (efS *EfSv1) ReplayEvents(ctx *context.Context, args *utils.ArgsReplayFailedPosts, reply *string) error {
	return efS.efs.V1ReplayEvents(ctx, args, reply)
}
