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

	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) AccountSv1Ping(args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaAccounts, utils.AccountSv1Ping, args, rpl)
}

func (dS *DispatcherService) AccountsForEvent(args *utils.ArgsAccountsForEvent, reply *[]*utils.Account) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1AccountsForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAccounts, utils.AccountSv1AccountsForEvent, args, reply)
}

func (dS *DispatcherService) MaxAbstracts(args *utils.ArgsAccountsForEvent, reply *utils.ExtEventCharges) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1MaxAbstracts, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAccounts, utils.AccountSv1MaxAbstracts, args, reply)
}

func (dS *DispatcherService) DebitAbstracts(args *utils.ArgsAccountsForEvent, reply *utils.ExtEventCharges) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1DebitAbstracts, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAccounts, utils.AccountSv1DebitAbstracts, args, reply)
}

func (dS *DispatcherService) MaxConcretes(args *utils.ArgsAccountsForEvent, reply *utils.ExtEventCharges) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1MaxConcretes, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAccounts, utils.AccountSv1MaxConcretes, args, reply)
}

func (dS *DispatcherService) DebitConcretes(args *utils.ArgsAccountsForEvent, reply *utils.ExtEventCharges) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1DebitConcretes, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAccounts, utils.AccountSv1DebitConcretes, args, reply)
}

func (dS *DispatcherService) AccountSv1ActionSetBalance(args *utils.ArgsActSetBalance, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1ActionSetBalance, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaAccounts, utils.AccountSv1ActionSetBalance, args, reply)
}

func (dS *DispatcherService) AccountSv1ActionRemoveBalance(args *utils.ArgsActRemoveBalances, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AccountSv1ActionRemoveBalance, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaAccounts, utils.AccountSv1ActionRemoveBalance, args, reply)
}
