// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDspAttributeSv1PingError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrEvent := &utils.CGREvent{}
	var reply *string
	err := dspSrv.AttributeSv1Ping(context.Background(), cgrEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1PingErrorTenant(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrEvent := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "",
		Time:    nil,
		Event:   nil,
		APIOpts: nil,
	}
	var reply *string
	err := dspSrv.AttributeSv1Ping(context.Background(), cgrEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	err := dspSrv.AttributeSv1Ping(context.Background(), nil, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1PingErrorAttributeSConns(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrEvent := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "ID",
		Time:    nil,
		Event:   nil,
		APIOpts: nil,
	}
	var reply *string
	err := dspSrv.AttributeSv1Ping(context.Background(), cgrEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &utils.CGREvent{
		Time: &time.Time{},
	}
	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(context.Background(), processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventErrorTenant(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &utils.CGREvent{
		Tenant: "tenant",
		Time:   &time.Time{},
	}
	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(context.Background(), processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventErrorAttributeS(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &utils.CGREvent{
		Tenant: "tenant",
		Time:   &time.Time{},
	}

	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(context.Background(), processEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1ProcessEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &utils.CGREvent{
		Tenant: "tenant",
		Time:   &time.Time{},
	}

	var reply *engine.AttrSProcessEventReply
	err := dspSrv.AttributeSv1ProcessEvent(context.Background(), processEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1ProcessEventErrorAttributeSConns(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &utils.CGREvent{
		Tenant: "tenant",
		Time:   &time.Time{},
	}

	var reply *engine.AttrSProcessEventReply
	err := dspSrv.AttributeSv1ProcessEvent(context.Background(), processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}
