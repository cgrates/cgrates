// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// AttributeSv1Ping interrogates AttributeS server responsible to process the event
func (dS *DispatcherService) AttributeSv1Ping(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaAttributes, utils.AttributeSv1Ping, args, reply)
}

// AttributeSv1GetAttributeForEvent is the dispatcher method for AttributeSv1.GetAttributeForEvent
func (dS *DispatcherService) AttributeSv1GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1GetAttributeForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaAttributes, utils.AttributeSv1GetAttributeForEvent, args, reply)
}

// AttributeSv1ProcessEvent .
func (dS *DispatcherService) AttributeSv1ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}

	}
	return dS.Dispatch(args, utils.MetaAttributes, utils.AttributeSv1ProcessEvent, args, reply)
}
