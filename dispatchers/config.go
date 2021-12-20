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

// do not modify this code because it's generated
package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ConfigSv1GetConfig(ctx *context.Context, args *config.SectionWithAPIOpts, reply *map[string]interface{}) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1GetConfig, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1GetConfig, args, reply)
}
func (dS *DispatcherService) ConfigSv1GetConfigAsJSON(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1GetConfigAsJSON, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1GetConfigAsJSON, args, reply)
}
func (dS *DispatcherService) ConfigSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1Ping, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1Ping, args, reply)
}
func (dS *DispatcherService) ConfigSv1ReloadConfig(ctx *context.Context, args *config.ReloadArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1ReloadConfig, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1ReloadConfig, args, reply)
}
func (dS *DispatcherService) ConfigSv1SetConfig(ctx *context.Context, args *config.SetConfigArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1SetConfig, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1SetConfig, args, reply)
}
func (dS *DispatcherService) ConfigSv1SetConfigFromJSON(ctx *context.Context, args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1SetConfigFromJSON, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1SetConfigFromJSON, args, reply)
}
func (dS *DispatcherService) ConfigSv1StoreCfgInDB(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.ConfigSv1StoreCfgInDB, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaConfig, utils.ConfigSv1StoreCfgInDB, args, reply)
}
