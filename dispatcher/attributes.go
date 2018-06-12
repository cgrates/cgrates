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

package dispatcher

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) AttributeSv1Ping(ign string, reply *string) error {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.ResourceS)
	}
	return dS.attrS.Call(utils.AttributeSv1Ping, ign, reply)
}

func (dS *DispatcherService) AttributeSv1GetAttributeForEvent(args *CGREvWithApiKey,
	reply *engine.AttributeProfile) (err error) {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.ResourceS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Time:    args.CGREvent.Time,
		Event: map[string]interface{}{
			utils.APIKey: args.APIKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	var apiMethods string
	if apiMethods, err = rplyEv.CGREvent.FieldAsString(utils.APIMethods); err != nil {
		return
	}
	if !utils.ParseStringMap(apiMethods).HasKey(utils.AttributeSv1GetAttributeForEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.attrS.Call(utils.AttributeSv1GetAttributeForEvent, args.CGREvent, reply)

}

func (dS *DispatcherService) AttributeSv1ProcessEvent(args *CGREvWithApiKey,
	reply *engine.AttrSProcessEventReply) (err error) {
	if dS.attrS == nil {
		return utils.NewErrNotConnected(utils.ResourceS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Time:    args.CGREvent.Time,
		Event: map[string]interface{}{
			utils.APIKey: args.APIKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	var apiMethods string
	if apiMethods, err = rplyEv.CGREvent.FieldAsString(utils.APIMethods); err != nil {
		return
	}
	if !utils.ParseStringMap(apiMethods).HasKey(utils.AttributeSv1ProcessEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.attrS.Call(utils.AttributeSv1ProcessEvent, args.CGREvent, reply)

}
