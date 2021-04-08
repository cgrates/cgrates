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

func TestGuardianGuardianSv1PingErr1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.GuardianSv1Ping(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1PingErr2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string

	expected := "DISPATCHER_ERROR:NOT_FOUND"
	result := dspSrv.GuardianSv1Ping(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1PingErrNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var CGREvent *utils.CGREvent
	var reply *string

	expected := "DISPATCHER_ERROR:NOT_FOUND"
	result := dspSrv.GuardianSv1Ping(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1RemoteLockErr1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := AttrRemoteLockWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.GuardianSv1RemoteLock(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1RemoteLockErr2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := AttrRemoteLockWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string

	expected := "DISPATCHER_ERROR:NOT_FOUND"
	result := dspSrv.GuardianSv1RemoteLock(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1RemoteUnlockErr1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := AttrRemoteUnlockWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.GuardianSv1RemoteUnlock(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1RemoteUnlockErr2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := AttrRemoteUnlockWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string

	expected := "DISPATCHER_ERROR:NOT_FOUND"
	result := dspSrv.GuardianSv1RemoteUnlock(CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
