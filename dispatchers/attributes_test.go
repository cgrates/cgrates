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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDspAttributeSv1PingError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrEvent := &utils.CGREvent{}
	var reply *string
	err := dspSrv.AttributeSv1Ping(cgrEvent, reply)
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
	err := dspSrv.AttributeSv1Ping(cgrEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1PingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	err := dspSrv.AttributeSv1Ping(nil, reply)
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
	err := dspSrv.AttributeSv1Ping(cgrEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &engine.AttrArgsProcessEvent{
		AttributeIDs: nil,
		Context:      nil,
		ProcessRuns:  nil,
		CGREvent: &utils.CGREvent{
			Tenant:  "",
			ID:      "",
			Time:    &time.Time{},
			Event:   nil,
			APIOpts: nil,
		},
	}
	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventErrorTenant(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &engine.AttrArgsProcessEvent{
		AttributeIDs: nil,
		Context:      nil,
		ProcessRuns:  nil,
		CGREvent: &utils.CGREvent{
			Tenant:  "tenant",
			ID:      "",
			Time:    &time.Time{},
			Event:   nil,
			APIOpts: nil,
		},
	}
	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1GetAttributeForEventErrorAttributeS(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &engine.AttrArgsProcessEvent{
		AttributeIDs: nil,
		Context:      nil,
		ProcessRuns:  nil,
		CGREvent: &utils.CGREvent{
			Tenant:  "tenant",
			ID:      "",
			Time:    &time.Time{},
			Event:   nil,
			APIOpts: nil,
		},
	}

	var reply *engine.AttributeProfile
	err := dspSrv.AttributeSv1GetAttributeForEvent(processEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1ProcessEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &engine.AttrArgsProcessEvent{
		AttributeIDs: nil,
		Context:      nil,
		ProcessRuns:  nil,
		CGREvent: &utils.CGREvent{
			Tenant:  "tenant",
			ID:      "",
			Time:    &time.Time{},
			Event:   nil,
			APIOpts: nil,
		},
	}

	var reply *engine.AttrSProcessEventReply
	err := dspSrv.AttributeSv1ProcessEvent(processEvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspAttributeSv1ProcessEventErrorAttributeSConns(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	processEvent := &engine.AttrArgsProcessEvent{
		AttributeIDs: nil,
		Context:      nil,
		ProcessRuns:  nil,
		CGREvent: &utils.CGREvent{
			Tenant:  "tenant",
			ID:      "",
			Time:    &time.Time{},
			Event:   nil,
			APIOpts: nil,
		},
	}

	var reply *engine.AttrSProcessEventReply
	err := dspSrv.AttributeSv1ProcessEvent(processEvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}
