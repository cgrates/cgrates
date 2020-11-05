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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ConfigSv1GetConfig(args *config.SectionWithOpts, reply *map[string]interface{}) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ConfigSv1GetConfig, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaConfig, utils.ConfigSv1GetConfig, args, reply)
}

func (dS *DispatcherService) ConfigSv1ReloadConfigFromPath(args *config.ConfigReloadWithOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ConfigSv1ReloadConfigFromPath, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaConfig, utils.ConfigSv1ReloadConfigFromPath, args, reply)
}

func (dS *DispatcherService) ConfigSv1ReloadConfig(args *config.ArgsReloadWithOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ConfigSv1ReloadConfig, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaConfig, utils.ConfigSv1ReloadConfig, args, reply)
}

func (dS *DispatcherService) ConfigSv1ReloadConfigFromJSON(args *config.JSONStringReloadWithOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ConfigSv1ReloadConfigFromJSON, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaConfig, utils.ConfigSv1ReloadConfigFromJSON, args, reply)
}

func (dS *DispatcherService) ConfigSv1GetConfigAsJSON(args *config.SectionWithOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ConfigSv1GetConfigAsJSON, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaConfig, utils.ConfigSv1GetConfigAsJSON, args, reply)
}
