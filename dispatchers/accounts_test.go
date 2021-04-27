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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDspAccountSv1PingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountSv1PingNilArgs(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.AccountSv1Ping(nil, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAAccountsForEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]*utils.Account
	result := dspSrv.AccountsForEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountsForEventErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]*utils.Account
	result := dspSrv.AccountsForEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspMaxAbstractsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.MaxAbstracts(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspMaxAbstractsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.MaxAbstracts(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspDebitAbstractsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.DebitAbstracts(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspDebitAbstractsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.DebitAbstracts(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspMaxConcretesNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.MaxConcretes(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspMaxConcretesErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.MaxConcretes(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspDebitConcretesNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.DebitConcretes(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspDebitConcretesErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *utils.ExtEventCharges
	result := dspSrv.DebitConcretes(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountSv1ActionSetBalanceNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsActSetBalance{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1ActionSetBalance(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountSv1ActionSetBalanceErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsActSetBalance{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1ActionSetBalance(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAAccountSv1ActionRemoveBalanceNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsActRemoveBalances{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1ActionRemoveBalance(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspAccountSv1ActionRemoveBalanceErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsActRemoveBalances{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.AccountSv1ActionRemoveBalance(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
