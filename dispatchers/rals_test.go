// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestRALsRALsV1PingErr1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.RALsV1Ping(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRALsRALsV1PingErr2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.RALsV1Ping(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRALsRALsV1PingErrNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var CGREvent *utils.CGREvent
	var reply *string

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.RALsV1Ping(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRALsRALsV1GetRatingPlansCostErr1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.RatingPlanCostArg{}
	var reply *RatingPlanCost

	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	result := dspSrv.RALsV1GetRatingPlansCost(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestRALsRALsV1GetRatingPlansCostErr2(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.RatingPlanCostArg{}
	var reply *RatingPlanCost

	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	result := dspSrv.RALsV1GetRatingPlansCost(context.Background(), CGREvent, reply)

	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
