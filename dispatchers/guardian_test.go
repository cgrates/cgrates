// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"

	"github.com/cgrates/birpc/context"
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
	result := dspSrv.GuardianSv1Ping(context.Background(), CGREvent, reply)

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

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.GuardianSv1Ping(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestGuardianGuardianSv1PingErrNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var CGREvent *utils.CGREvent
	var reply *string

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.GuardianSv1Ping(context.Background(), CGREvent, reply)

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
	result := dspSrv.GuardianSv1RemoteLock(context.Background(), CGREvent, reply)

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

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.GuardianSv1RemoteLock(context.Background(), CGREvent, reply)

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
	result := dspSrv.GuardianSv1RemoteUnlock(context.Background(), CGREvent, reply)

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

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.GuardianSv1RemoteUnlock(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
