// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
