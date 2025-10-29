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

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ErSv1Ping(ctx *context.Context, cgrEv *utils.CGREvent, reply *string) error {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if cgrEv != nil && len(cgrEv.Tenant) != 0 {
		tnt = cgrEv.Tenant
	}
	ev := make(map[string]any)
	if cgrEv != nil {
		ev = cgrEv.Event
	}
	opts := make(map[string]any)
	if cgrEv != nil {
		opts = cgrEv.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err := dS.authorize(utils.ErSv1Ping, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return err
		}
	}
	return dS.Dispatch(
		&utils.CGREvent{
			Tenant:  tnt,
			Event:   ev,
			APIOpts: opts,
		},
		utils.MetaERs,
		utils.ErSv1Ping, cgrEv, reply,
	)
}

func (dS *DispatcherService) ErSv1RunReader(ctx *context.Context, params ers.V1RunReaderParams, reply *string) error {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if params.Tenant != "" {
		tnt = params.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err := dS.authorize(utils.ErSv1RunReader, tnt,
			utils.IfaceAsString(params.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return err
		}
	}
	return dS.Dispatch(
		&utils.CGREvent{
			Tenant:  tnt,
			ID:      params.ID,
			APIOpts: params.APIOpts,
		},
		utils.MetaERs,
		utils.ErSv1RunReader, params, reply,
	)
}
