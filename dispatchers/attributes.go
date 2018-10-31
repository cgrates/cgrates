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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) AttributeSv1Ping(ign string, reply *string) error {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.AttributeS)
	}
	return dS.attrS.Call(utils.AttributeSv1Ping, ign, reply)
}

func (dS *DispatcherService) AttributeSv1GetAttributeForEvent(args *ArgsAttrProcessEventWithApiKey,
	reply *engine.AttributeProfile) (err error) {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.AttributeS)
	}
	if err = dS.authorize(utils.AttributeSv1GetAttributeForEvent, args.AttrArgsProcessEvent.CGREvent.Tenant,
		args.APIKey, args.AttrArgsProcessEvent.CGREvent.Time); err != nil {
		return
	}
	return dS.attrS.Call(utils.AttributeSv1GetAttributeForEvent, args.AttrArgsProcessEvent, reply)

}

func (dS *DispatcherService) AttributeSv1ProcessEvent(args *ArgsAttrProcessEventWithApiKey,
	reply *engine.AttrSProcessEventReply) (err error) {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.AttributeS)
	}
	if err = dS.authorize(utils.AttributeSv1ProcessEvent, args.AttrArgsProcessEvent.CGREvent.Tenant,
		args.APIKey, args.AttrArgsProcessEvent.CGREvent.Time); err != nil {
		return
	}
	return dS.attrS.Call(utils.AttributeSv1ProcessEvent, args.AttrArgsProcessEvent, reply)

}
