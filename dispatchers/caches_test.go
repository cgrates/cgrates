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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestDspCacheSv1PingError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "",
		Time:    &time.Time{},
		Event:   nil,
		APIOpts: nil,
	}
	var reply *string
	result := dspSrv.CacheSv1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1PingErrorArgs(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	var reply *string
	result := dspSrv.CacheSv1Ping(nil, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1PingErrorAttributeSConns(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant:  "tenant",
		ID:      "",
		Time:    &time.Time{},
		Event:   nil,
		APIOpts: nil,
	}
	var reply *string
	result := dspSrv.CacheSv1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetItemIDsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetCacheItemIDsWithAPIOpts{}
	var reply *[]string
	result := dspSrv.CacheSv1GetItemIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetItemIDsErrorArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.CacheSv1GetItemIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1HasItemError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{}
	var reply *bool
	result := dspSrv.CacheSv1HasItem(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1HasItemErrorArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *bool
	result := dspSrv.CacheSv1HasItem(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetItemExpiryTimeCacheSv1GetItemExpiryTimeError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{}
	var reply *time.Time
	result := dspSrv.CacheSv1GetItemExpiryTime(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetItemExpiryTimeCacheSv1GetItemExpiryTimeErrorArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *time.Time
	result := dspSrv.CacheSv1GetItemExpiryTime(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveItemError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1RemoveItem(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveItemArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1RemoveItem(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveItemsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1RemoveItems(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveItemsArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1RemoveItems(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ClearError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1Clear(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ClearArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1Clear(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetCacheStatsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{}
	var reply *map[string]*ltcache.CacheStats
	result := dspSrv.CacheSv1GetCacheStats(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetCacheStatsArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *map[string]*ltcache.CacheStats
	result := dspSrv.CacheSv1GetCacheStats(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1PrecacheStatusError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{}
	var reply *map[string]string
	result := dspSrv.CacheSv1PrecacheStatus(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1PrecacheStatusArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrCacheIDsWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *map[string]string
	result := dspSrv.CacheSv1PrecacheStatus(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1HasGroupError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{}
	var reply *bool
	result := dspSrv.CacheSv1HasGroup(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1HasGroupArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *bool
	result := dspSrv.CacheSv1HasGroup(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetGroupItemIDsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{}
	var reply *[]string
	result := dspSrv.CacheSv1GetGroupItemIDs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1GetGroupItemIDsArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]string
	result := dspSrv.CacheSv1GetGroupItemIDs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveGroupError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1RemoveGroup(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1RemoveGroupArgsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgsGetGroupWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1RemoveGroup(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReloadCacheError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1ReloadCache(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReloadCacheNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1ReloadCache(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1LoadCacheError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{}
	var reply *string
	result := dspSrv.CacheSv1LoadCache(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1LoadCacheNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.AttrReloadCacheWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1LoadCache(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReplicateRemoveError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgCacheReplicateRemove{}
	var reply *string
	result := dspSrv.CacheSv1ReplicateRemove(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReplicateRemoveNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgCacheReplicateRemove{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1ReplicateRemove(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReplicateSetError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.ArgCacheReplicateSet{}
	var reply *string
	result := dspSrv.CacheSv1ReplicateSet(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCacheSv1ReplicateSetNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.ArgCacheReplicateSet{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CacheSv1ReplicateSet(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
