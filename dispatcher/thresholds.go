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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ThresholdSv1Ping(ign string, reply *string) error {
	if dS.thdS != nil {
		if err := dS.thdS.Call(utils.ThresholdSv1Ping, ign, reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<DispatcherS> error: %s ThresholdS.", err.Error()))
		}
	}
	return nil
}

func (dS *DispatcherService) ThresholdSv1GetThresholdsForEvent(args *ArgsProcessEventWithApiKey,
	t *engine.Thresholds) (err error) {
	if dS.thdS == nil {
		return utils.NewErrNotConnected(utils.ThresholdS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Time:    utils.TimePointer(time.Now()),
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
	if !utils.ParseStringMap(apiMethods).HasKey(utils.ThresholdSv1GetThresholdsForEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.thdS.Call(utils.ThresholdSv1GetThresholdsForEvent, args.ArgsProcessEvent, t)
}

func (dS *DispatcherService) ThresholdSv1ProcessEvent(args *ArgsProcessEventWithApiKey,
	tIDs *[]string) (err error) {
	if dS.thdS == nil {
		return utils.NewErrNotConnected(utils.ThresholdS)
	}
	ev := &utils.CGREvent{
		Tenant:  args.Tenant,
		ID:      utils.UUIDSha1Prefix(),
		Context: utils.StringPointer(utils.MetaAuth),
		Time:    utils.TimePointer(time.Now()),
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
	if !utils.ParseStringMap(apiMethods).HasKey(utils.ThresholdSv1ProcessEvent) {
		return utils.ErrUnauthorizedApi
	}
	return dS.thdS.Call(utils.ThresholdSv1ProcessEvent, args.ArgsProcessEvent, tIDs)
}
