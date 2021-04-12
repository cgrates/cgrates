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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAttributeSv1 returns the RPC Object for AttributeS
func NewAttributeSv1(attrS *engine.AttributeService) *AttributeSv1 {
	return &AttributeSv1{attrS: attrS}
}

// AttributeSv1 exports RPC from RLs
type AttributeSv1 struct {
	attrS *engine.AttributeService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (alSv1 *AttributeSv1) Call(ctx *context.Context, serviceMethod string,
	args, reply interface{}) error {
	return utils.APIerRPCCallCtx(alSv1, ctx, serviceMethod, args, reply)
}

// GetAttributeForEvent  returns matching AttributeProfile for Event
func (alSv1 *AttributeSv1) GetAttributeForEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttributeProfile) (err error) {
	return alSv1.attrS.V1GetAttributeForEvent(args, reply)
}

// ProcessEvent will replace event fields with the ones in matching AttributeProfile
func (alSv1 *AttributeSv1) ProcessEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttrSProcessEventReply) error {
	return alSv1.attrS.V1ProcessEvent(args, reply)
}

// Ping return pong if the service is active
func (alSv1 *AttributeSv1) Ping(_ *context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
