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
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) StatSv1Ping(ign string, reply *string) error {
	if dS.statS != nil {
		if err := dS.statS.Call(utils.StatSv1Ping, ign, reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s StatS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(args *CGREvWithApiKey,
	reply *[]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Event: map[string]interface{}{
			utils.APIKey: args.APIKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	mp := utils.ParseStringMap(rplyEv.CGREvent.Event[utils.APIMethods].(string))
	if !mp.HasKey(utils.StatSv1GetStatQueuesForEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.statS.Call(utils.StatSv1GetStatQueuesForEvent, args.CGREvent, reply)
}

func (dS *DispatcherService) StatSv1GetQueueStringMetrics(args *TntIDWithApiKey,
	reply *map[string]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Event: map[string]interface{}{
			utils.APIKey: args.APIKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	mp := utils.ParseStringMap(rplyEv.CGREvent.Event[utils.APIMethods].(string))
	if !mp.HasKey(utils.StatSv1GetQueueStringMetrics) {
		return utils.ErrUnauthorizedApi
	}
	return dS.statS.Call(utils.StatSv1GetQueueStringMetrics, args.TenantID, reply)
}

func (dS *DispatcherService) StatSv1ProcessEvent(args *CGREvWithApiKey,
	reply *[]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Event: map[string]interface{}{
			utils.APIKey: args.APIKey,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err = dS.authorizeEvent(ev, &rplyEv); err != nil {
		return
	}
	mp := utils.ParseStringMap(rplyEv.CGREvent.Event[utils.APIMethods].(string))
	if !mp.HasKey(utils.StatSv1ProcessEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.statS.Call(utils.StatSv1ProcessEvent, args.CGREvent, reply)
}
