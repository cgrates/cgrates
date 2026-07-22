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

func TestDspResponderPingEventNilEvent(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *string
	result := dspSrv.ResponderPing(context.Background(), nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderPingEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderPing(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderPingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderPing(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderDebitNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderDebit(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderDebitErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderDebit(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderGetCost(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderGetCost(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderMaxDebitNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderMaxDebit(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderMaxDebitErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.CallCost
	result := dspSrv.ResponderMaxDebit(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundIncrementsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.Account
	result := dspSrv.ResponderRefundIncrements(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundIncrementsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *engine.Account
	result := dspSrv.ResponderRefundIncrements(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundRoundingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply engine.Account
	result := dspSrv.ResponderRefundRounding(context.Background(), CGREvent, &reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderRefundRoundingErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply engine.Account
	result := dspSrv.ResponderRefundRounding(context.Background(), CGREvent, &reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetMaxSessionTimeNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *time.Duration
	result := dspSrv.ResponderGetMaxSessionTime(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetMaxSessionTimeErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant: "tenant",
		},
	}
	var reply *time.Duration
	result := dspSrv.ResponderGetMaxSessionTime(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderShutdownNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderShutdown(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderShutdownErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.TenantWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.ResponderShutdown(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostOnRatingPlans(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(idb, cfg.CacheCfg(), nil)
	dsp := NewDispatcherService(dm, cfg, nil, nil)
	args := &utils.GetCostOnRatingPlansArgs{
		Account: "1002",
		RatingPlanIDs: []string{
			"RP1",
			"RP2",
		},
		Subject:     "1002",
		Destination: "1001",
		Usage:       2 * time.Minute,
		Tenant:      "cgrates.org",
	}
	if err := dm.SetDispatcherHost(&engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID: "ALL2",
		},
	}); err != nil {
		t.Error(err)
	}
	if err := dm.SetDispatcherProfile(&engine.DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test1",
		FilterIDs:  []string{},
		Strategy:   utils.MetaRoundRobin,
		Subsystems: []string{utils.MetaAny},
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "ALL2",
				FilterIDs: []string{},
				Weight:    20,
				Params:    make(map[string]any),
			},
		},
		Weight: 20,
	}, true); err != nil {
		t.Error(err)
	}
	var reply map[string]any
	if err := dsp.ResponderGetCostOnRatingPlans(context.Background(), args, &reply); err == nil {
		t.Error(err)
	}
}

// func TestDspResponderGetMaxSessionTimeOnAccounts(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	idb, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
// 	if err != nil{t.Error(err)}
// 	dm := engine.NewDataManager(idb, cfg.CacheCfg(), nil)
// 	dsp := NewDispatcherService(dm, cfg, nil, nil)
// 	args := &utils.GetMaxSessionTimeOnAccountsArgs{
// 		Subject:     "1002",
// 		Destination: "1001",
// 		Usage:       2 * time.Minute,
// 		Tenant:      "cgrates.org",
// 	}
// 	if err := dm.SetDispatcherHost(&engine.DispatcherHost{
// 		Tenant: "cgrates.org",
// 		RemoteHost: &config.RemoteHost{
// 			ID: "ALL2",
// 		},
// 	}); err != nil {
// 		t.Error(err)
// 	}
// 	if err := dm.SetDispatcherProfile(&engine.DispatcherProfile{
// 		Tenant:     "cgrates.org",
// 		ID:         "DSP_Test1",
// 		FilterIDs:  []string{},
// 		Strategy:   utils.MetaRoundRobin,
// 		Subsystems: []string{utils.MetaAny},
// 		Hosts: engine.DispatcherHostProfiles{
// 			&engine.DispatcherHostProfile{
// 				ID:        "ALL2",
// 				FilterIDs: []string{},
// 				Weight:    20,
// 				Params:    make(map[string]any),
// 			},
// 		},
// 		Weight: 20,
// 	}, true); err != nil {
// 		t.Error(err)
// 	}
// 	var reply map[string]any
// 	if err := dsp.ResponderGetMaxSessionTimeOnAccounts(context.Background(),args, &reply); err == nil {
// 		t.Error(err)
// 	}
// }

func TestDspResponderGetMaxSessionTimeOnAccountsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.GetMaxSessionTimeOnAccountsArgs{}

	var reply *map[string]any
	err := dspSrv.ResponderGetMaxSessionTimeOnAccounts(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspResponderGetMaxSessionTimeOnAccountsErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.GetMaxSessionTimeOnAccountsArgs{}
	var reply *map[string]any
	err := dspSrv.ResponderGetMaxSessionTimeOnAccounts(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestDspResponderGetCostOnRatingPlansNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.GetCostOnRatingPlansArgs{
		Tenant: "tenant",
	}
	var reply *map[string]any
	result := dspSrv.ResponderGetCostOnRatingPlans(context.Background(), CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspResponderGetCostOnRatingPlansErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.GetCostOnRatingPlansArgs{
		Tenant: "tenant",
	}
	var reply *map[string]any
	result := dspSrv.ResponderGetCostOnRatingPlans(context.Background(), CGREvent, reply)
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
