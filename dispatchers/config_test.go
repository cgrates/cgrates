// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
)

func TestDspConfigSv1GetConfigNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &config.SectionWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *map[string]any
	result := dspSrv.ConfigSv1GetConfig(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1GetConfigErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &config.SectionWithAPIOpts{}
	var reply *map[string]any
	result := dspSrv.ConfigSv1GetConfig(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1ReloadConfigNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &config.ReloadArgs{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ConfigSv1ReloadConfig(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1ReloadConfigErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &config.ReloadArgs{}
	var reply *string
	result := dspSrv.ConfigSv1ReloadConfig(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1SetConfigNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &config.SetConfigArgs{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ConfigSv1SetConfig(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1SetConfigErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &config.SetConfigArgs{}
	var reply *string
	result := dspSrv.ConfigSv1SetConfig(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1SetConfigFromJSONNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &config.SetConfigFromJSONArgs{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ConfigSv1SetConfigFromJSON(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1SetConfigFromJSONErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &config.SetConfigFromJSONArgs{}
	var reply *string
	result := dspSrv.ConfigSv1SetConfigFromJSON(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1GetConfigAsJSONNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &config.SectionWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ConfigSv1GetConfigAsJSON(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspConfigSv1GetConfigAsJSONErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &config.SectionWithAPIOpts{}
	var reply *string
	result := dspSrv.ConfigSv1GetConfigAsJSON(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
