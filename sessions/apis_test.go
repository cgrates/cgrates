// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/chargers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/routes"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type testMockClients struct {
	calls map[string]func(ctx *context.Context, method string, args, reply any) error
}

func (m *testMockClients) Call(ctx *context.Context, method string, args, reply any) error {
	if calls, has := m.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return calls(ctx, method, args, reply)
	}
}

func TestSessionSBiRPCv1AuthorizeEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.ResourceSv1AuthorizeResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*attributes.ProcessEventReply) = attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{
						{
							MatchedProfileID: "attr1",
						},
					},
					CGREvent: args.(*utils.CGREvent),
				}
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				cghrgs := []*chargers.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]any{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*chargers.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(45, 0)}
				return nil
			},
			utils.IPsV1AuthorizeIP: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
				return nil
			},
			utils.RouteSv1GetRoutes: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*routes.SortedRoutesList) = routes.SortedRoutesList{{
					Routes: []*routes.SortedRoute{
						{
							RouteID: "RouteID",
						},
					},
				}}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				return nil
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, method string, args, reply any) error {
				return utils.ErrPartiallyExecuted
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt

	for _, flag := range []string{
		utils.MetaAttributes,
		utils.MetaAccounts,
		utils.MetaResources,
		utils.MetaIPs,
		utils.MetaChargers,
		utils.MetaRoutes,
		utils.MetaThresholds,
		utils.MetaStats,
	} {
		connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
		cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
			{
				ConnIDs: []string{connID},
			},
		}
		for _, apiPrefix := range []string{
			utils.AttributeSv1,
			utils.AccountSv1,
			utils.ResourceSv1,
			utils.IPsV1,
			utils.ChargerSv1,
			utils.RouteSv1,
			utils.ThresholdSv1,
			utils.StatSv1,
		} {
			sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
		}
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			utils.MetaAccounts:   true,
			utils.MetaResources:  true,
			utils.MetaIPs:        true,
			utils.MetaChargers:   true,
			utils.MetaRoutes:     true,
			utils.MetaThresholds: true,
		},
	}
	var reply V1AuthorizeReply
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}

	//Resources
	args.APIOpts = map[string]any{
		utils.MetaResources: true,
	}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if *reply.ResourceAllocation != "OK" {
		t.Errorf("Expected ResourceAllocation OK, got %v", reply.ResourceAllocation)
	}

	//Attributes
	args.APIOpts = map[string]any{
		utils.MetaAttributes: true,
	}
	exp := []*attributes.FieldsAltered{{
		MatchedProfileID: "attr1",
	}}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply.Attributes.AlteredFields) {
		t.Errorf("Expected %v, recieved %v", exp, reply.Attributes.AlteredFields)
	}

	//Accounts
	args.APIOpts = map[string]any{
		utils.MetaAccounts: true,
	}
	expReply := utils.NewDecimal(45, 0)
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expReply, reply.MaxUsage) {
		t.Errorf("Expected %v, recieved %v", exp, reply.MaxUsage)
	}

	//IPs
	args.APIOpts = map[string]any{
		utils.MetaIPs: true,
	}
	expRpl := &utils.AllocatedIP{ProfileID: "prfIP"}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expRpl, reply.AllocatedIP) {
		t.Errorf("Expected %v, recieved %v", expRpl, reply.AllocatedIP)
	}

	//Chargers
	args.APIOpts = map[string]any{
		utils.MetaChargers: true,
	}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}

	//Routes
	args.APIOpts = map[string]any{
		utils.MetaRoutes: true,
	}
	expect := routes.SortedRoutesList{{
		Routes: []*routes.SortedRoute{
			{
				RouteID: "RouteID",
			},
		},
	}}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expect, reply.RouteProfiles) {
		t.Errorf("Expected %v, recieved %v", expect, reply.RouteProfiles)
	}

	//Thresholds
	args.APIOpts = map[string]any{
		utils.MetaRoutes:     true,
		utils.MetaThresholds: true,
	}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}

	//Stats
	cfg.SessionSCfg().Conns[utils.MetaStats] = []*config.DynamicConns{
		{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}},
	}
	sessions.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, chanInternal)
	args.APIOpts = map[string]any{
		utils.MetaRoutes: true,
		utils.MetaStats:  true,
	}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %v, recieved %v", utils.ErrPartiallyExecuted, err)
	}

	//chargers has multiple runEvents
	clnt.calls[utils.ChargerSv1ProcessEvent] = func(ctx *context.Context, m string, args, reply any) error {
		*reply.(*[]*chargers.ChrgSProcessEventReply) = []*chargers.ChrgSProcessEventReply{
			{
				ChargerSProfile: "CHRG1",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "run1Ev",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
			},
			{
				ChargerSProfile: "CHRG2",
				CGREvent: &utils.CGREvent{
					Tenant:  "cgrates.org",
					ID:      "run2Ev",
					Event:   map[string]any{},
					APIOpts: map[string]any{},
				},
			},
		}
		return nil
	}

	clnt.calls[utils.AccountSv1MaxAbstracts] = func(ctx *context.Context, m string, args, reply any) error {
		abstracts := utils.NewDecimal(70, 0)
		if args.(*utils.CGREvent).ID == "run2Ev" {
			abstracts = utils.NewDecimal(50, 0)
		}
		*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: abstracts}
		return nil
	}

	args.APIOpts = map[string]any{
		utils.MetaChargers: true,
		utils.MetaAccounts: true,
	}
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if reply.MaxUsage == nil || reply.MaxUsage.Compare(utils.NewDecimal(50, 0)) != 0 {
		t.Errorf("Expected MaxUsage to be 50, got %v", reply.MaxUsage)
	}
}

func TestSessionSBiRPCv1AuthorizeEventNotConnected(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	tests := []struct {
		name   string
		args   *utils.CGREvent
		expErr string
	}{
		{
			name:   "Nil CGREvent",
			args:   nil,
			expErr: "MANDATORY_IE_MISSING: [CGREvent]",
		},
		{
			name: "Empty Tenant and ID",
			args: &utils.CGREvent{
				Tenant: "",
				ID:     "",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{},
			},
		},
		{
			name: "Nil Event",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event:  nil,
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
				},
			},
			expErr: "MANDATORY_IE_MISSING: [Event]",
		},
		{
			name: "Nil APIOpts",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: nil,
			},
		},
		{
			name: "Empty APIOpts",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{},
			},
		},
		{
			name: "Attributes",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
				},
			},
			expErr: "ATTRIBUTES_ERROR:NOT_CONNECTED: AttributeS",
		},
		{
			name: "Accounts",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
				},
			},
			expErr: "ACCOUNTS_ERROR:NOT_CONNECTED: AccountS",
		},
		{
			name: "Resources",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaResources: true,
				},
			},
			expErr: "NOT_CONNECTED: ResourceS",
		},
		{
			name: "IPs",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaIPs: true,
				},
			},
			expErr: "NOT_CONNECTED: IPs",
		},
		{
			name: "Chargers",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaResources: true,
					utils.MetaChargers:  true,
				},
			},
			expErr: "NOT_CONNECTED: ChargerS",
		},
		{
			name: "Routes",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ev1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.MetaRoutes: true,
				},
			},
			expErr: "NOT_CONNECTED: RouteS",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reply V1AuthorizeReply
			if err := sessions.BiRPCv1AuthorizeEvent(ctx, tt.args, &reply); err != nil && err.Error() != tt.expErr {
				t.Errorf("Expected %v, recieved %v", tt.expErr, err)
			}
		})
	}
}

func TestSessionSBiRPCv1AuthorizeEventErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	var reply V1AuthorizeReply

	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.AttributeSv1GetAttributeForEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*attributes.ProcessEventReply) = attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{
						{
							MatchedProfileID: "attr1",
						},
					},
					CGREvent: args.(*utils.CGREvent),
				}
				return nil
			},
			utils.AccountSv1AccountsForEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(45, 0)}
				return nil
			},
			utils.IPsV1GetIPAllocationForEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
				return nil
			},
			utils.RouteSv1GetRouteProfilesForEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*routes.SortedRoutesList) = routes.SortedRoutesList{{
					Routes: []*routes.SortedRoute{
						{
							RouteID: "RouteID",
						},
					},
				}}
				return nil
			},
			utils.ThresholdSv1GetThresholdsForEvent: func(ctx *context.Context, m string, args, reply any) error {
				return utils.ErrNotImplemented
			},
			utils.StatSv1GetStatQueuesForEvent: func(ctx *context.Context, method string, args, reply any) error {
				return utils.ErrPartiallyExecuted
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt

	for _, flag := range []string{
		utils.MetaAttributes,
		utils.MetaAccounts,
		utils.MetaResources,
		utils.MetaIPs,
		utils.MetaRoutes,
		utils.MetaThresholds,
		utils.MetaStats,
	} {
		connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
		cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
			{
				ConnIDs: []string{connID},
			},
		}
		for _, apiPrefix := range []string{
			utils.AttributeSv1,
			utils.AccountSv1,
			utils.ResourceSv1,
			utils.IPsV1,
			utils.RouteSv1,
			utils.ThresholdSv1,
			utils.StatSv1,
		} {
			sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
		}
	}

	t.Run("Error cases: UNSUPPORTED_SERVICE_METHOD", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{},
		}

		//Resources
		args.APIOpts = map[string]any{
			utils.MetaResources: true,
		}
		expErr := "RESOURCES_ERROR:UNSUPPORTED_SERVICE_METHOD"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}

		//Attributes
		args.APIOpts = map[string]any{
			utils.MetaAttributes: true,
		}
		expErr = "ATTRIBUTES_ERROR:UNSUPPORTED_SERVICE_METHOD"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}

		//Accounts
		args.APIOpts = map[string]any{
			utils.MetaAccounts: true,
		}
		expErr = "ACCOUNTS_ERROR:UNSUPPORTED_SERVICE_METHOD"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}

		//IPs
		args.APIOpts = map[string]any{
			utils.MetaIPs: true,
		}
		expErr = "IPS_ERROR:UNSUPPORTED_SERVICE_METHOD"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}

		//Routes
		args.APIOpts = map[string]any{
			utils.MetaRoutes: true,
		}
		expErr = "ROUTES_ERROR:UNSUPPORTED_SERVICE_METHOD"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}

		//Thresholds
		clnt.calls[utils.RouteSv1GetRoutes] = func(ctx *context.Context, m string, args, reply any) error {
			*reply.(*routes.SortedRoutesList) = routes.SortedRoutesList{{
				Routes: []*routes.SortedRoute{
					{
						RouteID: "RouteID",
					},
				},
			}}
			return nil
		}
		args.APIOpts = map[string]any{
			utils.MetaRoutes:     true,
			utils.MetaThresholds: true,
		}
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err != utils.ErrPartiallyExecuted {
			t.Errorf("Expected %v, recieved %v", utils.ErrPartiallyExecuted, err)
		}
	})

	t.Run("Invalid bool value", func(t *testing.T) {
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)

		for _, flag := range []string{
			utils.MetaAttributes,
			utils.MetaAccounts,
			utils.MetaRoutes,
			utils.MetaResources,
			utils.MetaIPs,
			utils.MetaChargers,
		} {
			args := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaResources: true,
					flag:                []string{"test"},
				},
			}
			err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply)
			exp := "cannot convert field: [test] to bool"
			if err == nil || err.Error() != exp {
				t.Errorf("%s: Expected %v, recieved %v", flag, exp, err)
			}
		}

		clnt := &testMockClients{
			calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
				utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotFound
				},
			},
		}
		chanInternal := make(chan birpc.ClientConnector, 1)
		chanInternal <- clnt
		connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		cfg.SessionSCfg().Conns[utils.MetaAttributes] = []*config.DynamicConns{
			{
				ConnIDs: []string{connID},
			},
		}
		sessions.connMgr.AddInternalConn(connID, utils.AttributeSv1, chanInternal)

		for _, flag := range []string{utils.MetaThresholds, utils.MetaStats} {
			args := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
					flag:                 []string{"test"},
				},
			}
			var reply V1AuthorizeReply
			err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply)
			exp := "cannot convert field: [test] to bool"
			if err == nil || err.Error() != exp {
				t.Errorf("%s: Expected %v, recieved %v", flag, exp, err)
			}
		}
	})
	t.Run("Error cases for Resources and IPs", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaResources: true,
			},
		}
		cfg.SessionSCfg().Conns[utils.MetaResources] = []*config.DynamicConns{
			{FilterIDs: []string{"fltr"}, ConnIDs: []string{"testID"}},
		}
		exp := "NOT_FOUND:fltr"
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != exp {
			t.Errorf("Expected %v, recieved %v", exp, err)
		}

		args.APIOpts = map[string]any{
			utils.MetaIPs: true,
		}
		cfg.SessionSCfg().Conns[utils.MetaIPs] = []*config.DynamicConns{
			{FilterIDs: []string{"fltr"}, ConnIDs: []string{"testID"}},
		}
		if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err == nil || err.Error() != exp {
			t.Errorf("Expected %v, recieved %v", exp, err)
		}
	})
}

func TestSessionSBiRPCv1AuthorizeEventCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = -1
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	hits := 0
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				hits++
				*reply.(*attributes.ProcessEventReply) = attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{
						{
							MatchedProfileID: "attr1",
						},
					},
					CGREvent: args.(*utils.CGREvent),
				}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
	cfg.SessionSCfg().Conns[utils.MetaAttributes] = []*config.DynamicConns{
		{
			ConnIDs: []string{connID},
		},
	}
	sessions.connMgr.AddInternalConn(connID, utils.AttributeSv1, chanInternal)
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
		},
	}
	var reply V1AuthorizeReply
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if hits != 1 {
		t.Errorf("Expected AttributeS to be hit once, got %d", hits)
	}

	var reply2 V1AuthorizeReply
	if err := sessions.BiRPCv1AuthorizeEvent(ctx, args, &reply2); err != nil {
		t.Error(err)
	} else if hits != 1 {
		t.Errorf("Expected AttributeS to still have been hit only once, got %d", hits)
	}
}

func TestSessionSBiRPCv1AuthorizeEventWithDigest(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*attributes.ProcessEventReply) = attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{
						{
							MatchedProfileID: "cgrates.org:ATTR_1",
							Fields:           []string{"*req.Field1"},
						},
						{
							MatchedProfileID: "cgrates.org:ATTR_2",
							Fields:           []string{"*req.Field2"},
						},
					},
					CGREvent: args.(*utils.CGREvent),
				}
				return nil
			},
			utils.ResourceSv1AuthorizeResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(45, 0)}
				return nil
			},
			utils.RouteSv1GetRoutes: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*routes.SortedRoutesList) = routes.SortedRoutesList{{
					Routes: []*routes.SortedRoute{
						{RouteID: "RouteID"},
					},
				}}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"THD1"}
				return nil
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"STQ1"}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt

	for _, flag := range []string{
		utils.MetaAttributes,
		utils.MetaAccounts,
		utils.MetaResources,
		utils.MetaRoutes,
		utils.MetaThresholds,
		utils.MetaStats,
	} {
		connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
		cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
			{ConnIDs: []string{connID}},
		}
		for _, apiPrefix := range []string{
			utils.AttributeSv1,
			utils.AccountSv1,
			utils.ResourceSv1,
			utils.RouteSv1,
			utils.ThresholdSv1,
			utils.StatSv1,
		} {
			sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
		}
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evDigest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			utils.MetaAccounts:   true,
			utils.MetaResources:  true,
			utils.MetaRoutes:     true,
			utils.MetaThresholds: true,
			utils.MetaStats:      true,
		},
	}

	var reply V1AuthorizeReplyWithDigest
	if err := sessions.BiRPCv1AuthorizeEventWithDigest(ctx, args, &reply); err != nil {
		t.Error(err)
	}

	if reply.ResourceAllocation == nil || *reply.ResourceAllocation != utils.OK {
		t.Errorf("Expected %v, got %v", utils.OK, reply.ResourceAllocation)
	}

	if reply.MaxUsage == 0 || reply.MaxUsage != 45 {
		t.Errorf("Expected 45, recieved %v", reply.MaxUsage)
	}

	exp := utils.StringPointer("RouteID")
	if !reflect.DeepEqual(reply.RoutesDigest, exp) {
		t.Errorf("Expected %#+v, recieved %#+v", exp, reply.RoutesDigest)
	}

	expected := "THD1"
	if reply.Thresholds == nil || *reply.Thresholds != expected {
		t.Errorf("Expected %v, got %v", expected, reply.Thresholds)
	}

	expReply := "STQ1"
	if reply.StatQueues == nil || *reply.StatQueues != expReply {
		t.Errorf("Expected StatQueues %v, got %v", expReply, reply.StatQueues)
	}

	if reply.AttributesDigest == nil {
		t.Error("Expected AttributesDigest to be set")
	}

	clnt.calls[utils.AccountSv1MaxAbstracts] = func(ctx *context.Context, m string, args, reply any) error {
		return utils.ErrNotImplemented
	}

	expect := "ACCOUNTS_ERROR:NOT_IMPLEMENTED"
	if err := sessions.BiRPCv1AuthorizeEventWithDigest(ctx, args, &reply); err == nil || err.Error() != expect {
		t.Errorf("Expected %v, recieved %v", expect, err)
	}
}

func TestSessionSBiRPCv1ProcessEventRefund(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	refundCalls, debitCalls := 0, 0
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AccountSv1RefundCharges: func(_ *context.Context, _ string, args, _ any) error {
				refundCalls++
				charges := args.(*utils.APIEventCharges).EventCharges
				if charges.Concretes.Compare(utils.NewDecimal(1, 0)) != 0 {
					t.Errorf("Expected Concretes to be 1, got %v", charges.Concretes)
				}
				return utils.ErrNotImplemented
			},
			utils.AccountSv1DebitAbstracts: func(_ *context.Context, _ string, _, reply any) error {
				debitCalls++
				*reply.(*utils.EventCharges) = *utils.NewEventCharges()
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)
	cfg.SessionSCfg().Conns[utils.MetaAccounts] = []*config.DynamicConns{
		{ConnIDs: []string{connID}},
	}
	sessions.connMgr.AddInternalConn(connID, utils.AccountSv1, chanInternal)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event:  map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.MetaOriginID:        "refund",
			utils.MetaRefund:          true,
			utils.MetaAccountsCost:    map[string]any{utils.Concretes: 1.0},
			utils.MetaBlockerErrorCfg: true,
		},
	}
	var reply V1ProcessEventReply
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != utils.ErrNotImplemented {
		t.Errorf("Expected %v, received %v", utils.ErrNotImplemented, err)
	}

	delete(args.APIOpts, utils.MetaBlockerErrorCfg)
	args.APIOpts[utils.MetaAccountsDebitCfg] = true
	args.APIOpts[utils.MetaUsage] = 1
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %v, received %v", utils.ErrPartiallyExecuted, err)
	}

	delete(args.APIOpts, utils.MetaAccountsCost)
	args.APIOpts[utils.MetaBlockerErrorCfg] = true
	errMissing := utils.NewErrMandatoryIeMissing(utils.MetaAccountsCost)
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != errMissing.Error() {
		t.Errorf("Expected %v, received %v", errMissing, err)
	}

	delete(args.APIOpts, utils.MetaBlockerErrorCfg)
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %v, received %v", utils.ErrPartiallyExecuted, err)
	}
	if refundCalls != 2 {
		t.Errorf("Expected refund to be called twice, got %d", refundCalls)
	}
	if debitCalls != 0 {
		t.Errorf("Expected debit not to be called, got %d", debitCalls)
	}
}

func TestSessionSBiRPCv1InitiateSession(t *testing.T) {
	ctx := context.TODO()
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)

	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.ResourceSv1AllocateResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.IPsV1AllocateIP: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"THD1"}
				return nil
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"STQ1"}
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				cghrgs := []*chargers.ChrgSProcessEventReply{
					{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "TestID",
							Event: map[string]any{
								utils.Usage: "10s",
							},
						},
					},
				}
				*reply.(*[]*chargers.ChrgSProcessEventReply) = cghrgs
				return nil
			},
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				rplCast, canCast := reply.(*attributes.ProcessEventReply)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customEv := &attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "EV",
						Event: map[string]any{
							"CustomField2": "CustomValue2",
						},
						APIOpts: map[string]any{},
					},
				}
				*rplCast = *customEv
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt

	for flag, apiPrefix := range map[string]string{
		utils.MetaAttributes: utils.AttributeSv1,
		utils.MetaResources:  utils.ResourceSv1,
		utils.MetaIPs:        utils.IPsV1,
		utils.MetaThresholds: utils.ThresholdSv1,
		utils.MetaStats:      utils.StatSv1,
		utils.MetaChargers:   utils.ChargerSv1,
	} {
		connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
		sessions.cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
			{ConnIDs: []string{connID}},
		}
		sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
	}
	tempMaxUsage := time.Duration(utils.InvalidUsage)

	//OriginID
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: "originID",
		},
	}
	var reply V1InitSessionReply
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if reply.ResourceAllocation != nil || reply.AllocatedIP != nil {
		t.Errorf("Expected no allocation, recieved %+v", reply)
	} else if reply.MaxUsage == nil || *reply.MaxUsage != time.Duration(utils.InvalidUsage) {
		t.Errorf("Expected MaxUsage %v, recieved %v", tempMaxUsage, reply.MaxUsage)
	}

	//Resources
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:  "originID",
			utils.MetaResources: true,
		},
	}
	var reply1 V1InitSessionReply
	exprep := "OK"
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply1); err != nil {
		t.Error(err)
	} else if *reply1.ResourceAllocation != exprep {
		t.Errorf("Expected ResourceAllocation %v, recieved %v", exprep, reply1.ResourceAllocation)
	} else if reply1.AllocatedIP != nil {
		t.Errorf("Expected no IP allocation, recieved %v", reply1.AllocatedIP)
	}

	//IPs
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: "originID",
			utils.MetaIPs:      true,
		},
	}
	var reply2 V1InitSessionReply
	exp := utils.AllocatedIP{ProfileID: "prfIP"}
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply2); err != nil {
		t.Error(err)
	} else if reply2.AllocatedIP == nil || !reflect.DeepEqual(*reply2.AllocatedIP, exp) {
		t.Errorf("Expected AllocatedIP %v, recieved %v", exp, reply2.AllocatedIP)
	} else if reply2.ResourceAllocation != nil {
		t.Errorf("Expected no resource allocation, recieved %v", reply2.ResourceAllocation)
	}

	//Chargers
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:  "originID",
			utils.MetaResources: true,
			utils.MetaChargers:  true,
		},
	}
	var reply3 V1InitSessionReply
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply3); err != nil {
		t.Error(err)
	} else if reply3.ResourceAllocation == nil || *reply3.ResourceAllocation != "OK" {
		t.Errorf("Expected ResourceAllocation OK, recieved %v", reply3.ResourceAllocation)
	} else if reply3.AllocatedIP != nil {
		t.Errorf("Expected no IP allocation, recieved %v", reply3.AllocatedIP)
	} else if reply3.MaxUsage == nil || *reply3.MaxUsage != tempMaxUsage {
		t.Errorf("Expected the temporary MaxUsage %v, recieved %v", tempMaxUsage, reply3.MaxUsage)
	}

	//Inits
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        "10s",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:  "originID",
			utils.MetaResources: true,
			utils.MetaInitiate:  true,
		},
	}
	var reply4 V1InitSessionReply
	expMaxUsage := time.Duration(0 * time.Second)
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply4); err != nil {
		t.Error(err)
	} else if reply4.ResourceAllocation == nil || *reply4.ResourceAllocation != "OK" {
		t.Errorf("Expected ResourceAllocation OK, recieved %v", reply4.ResourceAllocation)
	} else if reply4.AllocatedIP != nil {
		t.Errorf("Expected no IP allocation, recieved %v", reply4.AllocatedIP)
	} else if reply4.MaxUsage == nil || utils.ToJSON(reply4.MaxUsage) != utils.ToJSON(&expMaxUsage) {
		t.Errorf("Expected MaxUsage %v, recieved %v", expMaxUsage, reply4.MaxUsage)
	}

	//IPs
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        "10s",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: "originID",
			utils.MetaIPs:      true,
			utils.MetaInitiate: true,
		},
	}
	var reply5 V1InitSessionReply
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply5); err == nil || err.Error() != "UNSUPPORTED_SERVICE_METHOD" {
		t.Error(err)
	} else if reply5.ResourceAllocation != nil {
		t.Errorf("Expected no resource allocation, recieved %v", reply5.ResourceAllocation)
	} else if reply5.AllocatedIP == nil || reply5.AllocatedIP.ProfileID != "prfIP" {
		t.Errorf("Expected AllocatedIP prfIP, recieved %v", reply5.AllocatedIP)
	}

	//Attributes
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "originID",
			utils.MetaResources:  true,
			utils.MetaAttributes: true,
		},
	}
	var reply6 V1InitSessionReply
	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply6); err != nil {
		t.Error(err)
	} else if reply6.ResourceAllocation == nil || *reply6.ResourceAllocation != "OK" {
		t.Errorf("Expected ResourceAllocation OK, recieved %v", reply6.ResourceAllocation)
	} else if reply6.AllocatedIP != nil {
		t.Errorf("Expected no IP allocation, recieved %v", reply6.AllocatedIP)
	} else if reply6.MaxUsage == nil || *reply6.MaxUsage != tempMaxUsage {
		t.Errorf("Expected the temporary MaxUsage %v, recieved %v", tempMaxUsage, reply6.MaxUsage)
	}

	//Resources, IPs, Thresholds and Stats
	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "originID",
			utils.MetaResources:  true,
			utils.MetaIPs:        true,
			utils.MetaThresholds: true,
			utils.MetaStats:      true,
		},
	}
	var reply7 V1InitSessionReply
	repThresholdIDs := []string{"THD1"}

	if err := sessions.BiRPCv1InitiateSession(ctx, args, &reply7); err != nil {
		t.Error(err)
	}
	if reply7.ResourceAllocation == nil || *reply7.ResourceAllocation != "OK" {
		t.Errorf("Expected ResourceAllocation OK, recieved %v", reply7.ResourceAllocation)
	}
	if reply7.AllocatedIP == nil || reply7.AllocatedIP.ProfileID != "prfIP" {
		t.Errorf("Expected AllocatedIP prfIP, recieved %v", reply7.AllocatedIP)
	}
	if reply7.ThresholdIDs == nil || !reflect.DeepEqual(*reply7.ThresholdIDs, repThresholdIDs) {
		t.Errorf("Expected ThresholdIDs %v, recieved %v", repThresholdIDs, reply7.ThresholdIDs)
	}
	if reply7.MaxUsage == nil || *reply7.MaxUsage != tempMaxUsage {
		t.Errorf("Expected the temporary MaxUsage %v, recieved %v", tempMaxUsage, reply7.MaxUsage)
	}
}

func TestSessionSBiRPCv1InitiateSessionError(t *testing.T) {
	ctx := context.TODO()
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Connected", func(t *testing.T) {
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)

		clnt := &testMockClients{
			calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
				utils.ResourceSv1AllocateResources: func(ctx *context.Context, m string, args, reply any) error {
					return nil
				},
				utils.IPsV1GetIPAllocationForEvent: func(ctx *context.Context, m string, args, reply any) error {
					*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
					return nil
				},
				utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotImplemented
				},
				utils.StatSv1ProcessEvent: func(ctx *context.Context, method string, args, reply any) error {
					return utils.ErrPartiallyExecuted
				},
				utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotImplemented
				},
			},
		}
		chanInternal := make(chan birpc.ClientConnector, 1)
		chanInternal <- clnt

		for flag, apiPrefix := range map[string]string{
			utils.MetaResources:  utils.ResourceSv1,
			utils.MetaIPs:        utils.IPsV1,
			utils.MetaThresholds: utils.ThresholdSv1,
			utils.MetaStats:      utils.StatSv1,
			utils.MetaChargers:   utils.ChargerSv1,
		} {
			connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
			sessions.cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
				{
					ConnIDs: []string{connID},
				},
			}
			sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
		}

		tests := []struct {
			name   string
			args   *utils.CGREvent
			expErr string
		}{
			{
				name:   "Nil CGREvent",
				args:   nil,
				expErr: "MANDATORY_IE_MISSING: [CGREvent]",
			},
			{
				name: "Nil Event",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event:  nil,
				},
				expErr: "MANDATORY_IE_MISSING: [Event]",
			},
			{
				name: "Empty tenant and id",
				args: &utils.CGREvent{
					Tenant: "",
					ID:     "",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{},
				},
				expErr: "NOT_FOUND",
			},
			{
				name: "Nil APIOpts",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: nil,
				},
				expErr: "NOT_FOUND",
			},
			{
				name: "OriginID not found",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{},
				},
				expErr: "NOT_FOUND",
			},
			{
				name: "Missing OriginID",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID: "",
					},
				},
				expErr: "MANDATORY_IE_MISSING: [OriginID]",
			},
			{
				name: "IPs parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID: "originID",
						utils.MetaIPs:      "truee",
					},
				},
				expErr: `strconv.ParseBool: parsing "truee": invalid syntax`,
			},
			{
				name: "Chargers parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaChargers:  "truee",
					},
				},
				expErr: `strconv.ParseBool: parsing "truee": invalid syntax`,
			},
			{
				name: "Initiate parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaInitiate:  "truee",
					},
				},
				expErr: `strconv.ParseBool: parsing "truee": invalid syntax`,
			},
			{
				name: "Initiate: UNSUPPORTED_SERVICE_METHOD",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.Usage:        "test",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "err",
						utils.MetaResources: true,
						utils.MetaInitiate:  true,
					},
				},
				expErr: "UNSUPPORTED_SERVICE_METHOD",
			},
			{
				name: "Attributes parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:   "originID",
						utils.MetaResources:  true,
						utils.MetaAttributes: "truee",
					},
				},
				expErr: `strconv.ParseBool: parsing "truee": invalid syntax`,
			},
			{
				name: "Resources parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: "truee",
					},
				},
				expErr: `strconv.ParseBool: parsing "truee": invalid syntax`,
			},
			{
				name: "Thresholds: UNSUPPORTED_SERVICE_METHOD",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:   "originID",
						utils.MetaResources:  true,
						utils.MetaThresholds: "truee",
					},
				},
				expErr: "UNSUPPORTED_SERVICE_METHOD",
			},
			{
				name: "Stats parsing error",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaStats:     "truee",
					},
				},
				expErr: "UNSUPPORTED_SERVICE_METHOD",
			},
			{
				name: "Thresholds: UNSUPPORTED_SERVICE_METHOD",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:   "originID",
						utils.MetaResources:  true,
						utils.MetaThresholds: true,
					},
				},
				expErr: "UNSUPPORTED_SERVICE_METHOD",
			},
			{
				name: "Stats: UNSUPPORTED_SERVICE_METHOD",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaStats:     true,
					},
				},
				expErr: "UNSUPPORTED_SERVICE_METHOD",
			},
			{
				name: "IPS: UNSUPPORTED_SERVICE_METHOD",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.Usage:        "10s",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID: "originID",
						utils.MetaIPs:      true,
					},
				},
				expErr: "IPS_ERROR:UNSUPPORTED_SERVICE_METHOD",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var reply V1InitSessionReply
				err := sessions.BiRPCv1InitiateSession(ctx, tt.args, &reply)
				if err == nil || err.Error() != tt.expErr {
					t.Errorf("Expected %v, recieved %v", tt.expErr, err)
				}
			})
		}
	})

	t.Run("Not Connected", func(t *testing.T) {
		ctx := context.TODO()
		cfg := config.NewDefaultCGRConfig()
		cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
		locker := engine.NewLocker(cfg)
		data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
		if err != nil {
			t.Fatal(err)
		}
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
		tests := []struct {
			name   string
			args   *utils.CGREvent
			expErr string
		}{
			{
				name: "Resources not connected",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
					},
				},
				expErr: "NOT_CONNECTED: ResourceS",
			},
			{
				name: "IPs not connected",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID: "originID",
						utils.MetaIPs:      true,
					},
				},
				expErr: "NOT_CONNECTED: IPs",
			},
			{
				name: "Attributes not connected",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:   "originID",
						utils.MetaResources:  true,
						utils.MetaAttributes: true,
					},
				},
				expErr: "ATTRIBUTES_ERROR:NOT_CONNECTED: AttributeS",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var reply V1InitSessionReply
				err := sessions.BiRPCv1InitiateSession(ctx, tt.args, &reply)
				if err == nil || err.Error() != tt.expErr {
					t.Errorf("Expected %v, recieved %v", tt.expErr, err)
				}
			})
		}
	})

	t.Run("FilterId not found", func(t *testing.T) {
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)

		clnt := &testMockClients{
			calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
				utils.ResourceSv1AllocateResources: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotImplemented
				},
				utils.IPsV1AllocateIP: func(ctx *context.Context, m string, args, reply any) error {
					*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
					return utils.ErrNotImplemented
				},
				utils.ThresholdSv1GetThresholdsForEvent: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotImplemented
				},
				utils.StatSv1GetStatQueuesForEvent: func(ctx *context.Context, method string, args, reply any) error {
					return utils.ErrPartiallyExecuted
				},
				utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
					return utils.ErrNotImplemented
				},
			},
		}
		chanInternal := make(chan birpc.ClientConnector, 1)
		chanInternal <- clnt

		for flag, apiPrefix := range map[string]string{
			utils.MetaResources:  utils.ResourceSv1,
			utils.MetaIPs:        utils.IPsV1,
			utils.MetaThresholds: utils.ThresholdSv1,
			utils.MetaStats:      utils.StatSv1,
			utils.MetaChargers:   utils.ChargerSv1,
		} {
			connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
			sessions.cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{
				{
					FilterIDs: []string{"test"},
					ConnIDs:   []string{connID},
				},
			}
			sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
		}

		tests := []struct {
			name   string
			args   *utils.CGREvent
			expErr string
		}{
			{
				name: "IPs",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID: "originID",
						utils.MetaIPs:      true,
					},
				},
				expErr: "NOT_FOUND:test",
			},
			{
				name: "Chargers",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaChargers:  true,
					},
				},
				expErr: "NOT_FOUND:test",
			},
			{
				name: "Resources",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "ev1",
					Event: map[string]any{
						utils.AccountField: "1001",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originNC",
						utils.MetaResources: true,
					},
				},
				expErr: "NOT_FOUND:test",
			},
			{
				name: "Initiate",
				args: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "evRes",
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.Usage:        "10s",
					},
					APIOpts: map[string]any{
						utils.MetaOriginID:  "originID",
						utils.MetaResources: true,
						utils.MetaInitiate:  true,
					},
				},
				expErr: "NOT_FOUND:test",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var reply V1InitSessionReply
				err := sessions.BiRPCv1InitiateSession(ctx, tt.args, &reply)
				if err == nil || err.Error() != tt.expErr {
					t.Errorf("Expected %v, recieved %v", tt.expErr, err)
				}
			})
		}
	})
}

func TestSessionSBiRPCv1ProcessEventNotConnected(t *testing.T) {
	ctx := context.TODO()
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)

	tests := []struct {
		name   string
		args   *utils.CGREvent
		expErr string
	}{
		{
			name:   "Nil CGREvent",
			args:   nil,
			expErr: "MANDATORY_IE_MISSING: [CGREvent]",
		},
		{
			name: "Nil fields",
			args: &utils.CGREvent{
				Tenant:  "",
				ID:      "",
				Event:   map[string]any{utils.AccountField: "1001"},
				APIOpts: nil,
			},
			expErr: "MANDATORY_IE_MISSING: [*originID]",
		},
		{
			name: "Nil Event",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evAttr",
				Event:  nil,
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
					utils.MetaOriginID:   "originAttr",
				},
			},
			expErr: "MANDATORY_IE_MISSING: [Event]",
		},
		{
			name: "Attributes",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evAttr",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
					utils.MetaOriginID:   "originAttr",
				},
			},
			expErr: "ATTRIBUTES_ERROR:NOT_CONNECTED: AttributeS",
		},
		{
			name: "Attributes: parsing error",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evAttr",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaAttributes: "trueee",
					utils.MetaOriginID:   "originAttr",
				},
			},
			expErr: `strconv.ParseBool: parsing "trueee": invalid syntax`,
		},
		{
			name: "Chargers",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evChrg",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaOriginID: "originChrg",
				},
			},
			expErr: "NOT_CONNECTED: ChargerS",
		},
		{
			name: "Routes",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evRoutes",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaRoutes:          true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "originRoutes",
				},
			},
			expErr: "NOT_CONNECTED: RouteS",
		},
		{
			name: "Stats",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evStats",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaStats:           true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "originStats",
				},
			},
			expErr: "NOT_CONNECTED: StatS",
		},
		{
			name: "Thresholds",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evThresholds",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaThresholds:      true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "originThresholds",
				},
			},
			expErr: "NOT_CONNECTED: ThresholdS",
		},
		{
			name: "IPsAuthorize",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evIPs",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaIPsAuthorizeCfg: true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "originIPs",
				},
			},
			expErr: "NOT_CONNECTED: IPs",
		},
		{
			name: "Rates",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evRates",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaRates:           true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "originRates",
				},
			},
			expErr: "NOT_CONNECTED: RateS",
		},
		{
			name: "AccountsAuthorize",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evAccountsAuth",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaAccountsAuthorizeCfg: true,
					utils.MetaUsage:                1 * time.Minute,
					utils.OptsSesBlockerError:      true,
					utils.MetaOriginID:             "originAccountsAuth",
				},
			},
			expErr: "NOT_CONNECTED: AccountS",
		},
		{
			name: "ResourcesAuthorize",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evResourcesAuth",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsSesBlockerError:       true,
					utils.MetaOriginID:              "originResourcesAuth",
				},
			},
			expErr: "NOT_CONNECTED: ResourceS",
		},
		{
			name: "ResourcesAllocate",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:             "originResAlloc",
					utils.OptsSesBlockerError:      true,
					utils.MetaResourcesAllocateCfg: true,
				}},
			expErr: "NOT_CONNECTED: ResourceS",
		},
		{
			name: "ResourcesRelease",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:            "originResRelease",
					utils.OptsSesBlockerError:     true,
					utils.MetaResourcesReleaseCfg: true,
				}},
			expErr: "NOT_CONNECTED: ResourceS",
		},
		{
			name: "AccountsDebit",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:         "originAcntDebit",
					utils.OptsSesBlockerError:  true,
					utils.MetaAccountsDebitCfg: true,
					utils.MetaUsage:            1 * time.Minute,
				}},
			expErr: "NOT_CONNECTED: AccountS",
		},
		{
			name: "Routes: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaRoutes:          true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "Stats: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaStats:           true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "Thresholds: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaThresholds:      true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "IPsAuthorize: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaIPsAuthorizeCfg: true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "Rates: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaRates:           true,
					utils.MetaUsage:           time.Minute,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "ResourcesAllocate: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:             "originID",
					utils.OptsSesBlockerError:      false,
					utils.MetaResourcesAllocateCfg: true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "ResourcesRelease: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:            "originID",
					utils.OptsSesBlockerError:     false,
					utils.MetaResourcesReleaseCfg: true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "AccountsAuthorizeMaxAbstracts: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:             "originID",
					utils.OptsSesBlockerError:      false,
					utils.MetaAccountsAuthorizeCfg: true,
					utils.MetaUsage:                time.Minute,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "AccountsRefund: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:          "originID",
					utils.OptsSesBlockerError:   false,
					utils.MetaAccountsRefundCfg: true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
		{
			name: "EEs: PARTIALLY_EXECUTED",
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "evID",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: false,
					utils.MetaEEs:             true,
				},
			},
			expErr: "PARTIALLY_EXECUTED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reply V1ProcessEventReply
			if err := sessions.BiRPCv1ProcessEvent(ctx, tt.args, &reply); err == nil || err.Error() != tt.expErr {
				t.Errorf("Expected %v, recieved %v", tt.expErr, err)
			}
		})
	}
}

func TestSessionSBiRPCv1ProcessEventAttributes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	clnt1 := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				rply := attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{{
						MatchedProfileID: "attr1",
					}},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]any{
							utils.AccountField: "1002",
						},
					},
				}
				*reply.(*attributes.ProcessEventReply) = rply
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt1
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
	cfg.SessionSCfg().Conns[utils.MetaAttributes] = []*config.DynamicConns{{ConnIDs: []string{connID}}}
	sessions.connMgr.AddInternalConn(connID, utils.AttributeSv1, chanInternal)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "originID1",
			utils.MetaAttributes: true,
		},
	}

	t.Run("Attributes", func(t *testing.T) {
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := []*attributes.FieldsAltered{{MatchedProfileID: "attr1"}}
		rcv := reply.Attributes[utils.MetaPrimary].AlteredFields
		if !reflect.DeepEqual(rcv, expected) {
			t.Errorf("Expected %v, recieved %v", expected, rcv)
		}
	})

	t.Run("error case", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "originID2",
				utils.MetaAttributes: "trueee",
			},
		}
		var reply V1ProcessEventReply
		expErr := `strconv.ParseBool: parsing "trueee": invalid syntax`
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, recieved %v", expErr, err)
		}
		if reply.Attributes != nil {
			t.Errorf("Expected nil Attributes, got %v", reply.Attributes)
		}
	})
}

func TestSessionSBiRPCv1ProcessEventCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = -1
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	call := 0
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				call++
				*reply.(*attributes.ProcessEventReply) = attributes.ProcessEventReply{
					AlteredFields: []*attributes.FieldsAltered{
						{MatchedProfileID: "attr1"},
					},
					CGREvent: args.(*utils.CGREvent),
				}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
	cfg.SessionSCfg().Conns[utils.MetaAttributes] = []*config.DynamicConns{{ConnIDs: []string{connID}}}
	sessions.connMgr.AddInternalConn(connID, utils.AttributeSv1, chanInternal)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evCache",
		Event:  map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.MetaAttributes: true,
			utils.MetaOriginID:   "originCache",
		},
	}
	var reply V1ProcessEventReply
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	} else if call != 1 {
		t.Errorf("Expected call to be 1, recieved %d", call)
	}

	var reply2 V1ProcessEventReply
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply2); err != nil {
		t.Error(err)
	} else if call != 1 {
		t.Errorf("Expected call to still be 1, recieved %d", call)
	}
}

func TestSessionSBiRPCv1ProcessEventUsageOptions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{
					Abstracts: utils.NewDecimal(int64(60*time.Second), 0),
				}
				return nil
			},
			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{
					Abstracts: utils.NewDecimal(int64(90*time.Second), 0),
				}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)
	cfg.SessionSCfg().Conns[utils.MetaAccounts] = []*config.DynamicConns{{ConnIDs: []string{connID}}}
	sessions.connMgr.AddInternalConn(connID, utils.AccountSv1, chanInternal)

	t.Run("InterimConsumed", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:        "originID",
				utils.MetaAccounts:        true,
				utils.MetaDebit:           true,
				utils.MetaUsage:           90 * time.Second,
				utils.MetaInterimConsumed: 7,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := 90 * time.Second
		rcv := reply.AccountsUsage[utils.MetaPrimary]
		if !reflect.DeepEqual(rcv, expected) {
			t.Errorf("Expected %v, recieved %v", expected, rcv)
		}
	})

	t.Run("InterimConsumed: parsing error", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:        "originID",
				utils.MetaAccounts:        true,
				utils.MetaDebit:           true,
				utils.MetaUsage:           90 * time.Second,
				utils.MetaInterimConsumed: "err",
			},
		}
		var reply V1ProcessEventReply
		expErr := "can't convert <err> to decimal"
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %#v, recieved %#v", expErr, err)
		}
		expected := 0 * time.Second
		if !reflect.DeepEqual(reply.AccountsUsage[utils.MetaPrimary], expected) {
			t.Errorf("Expected %#v, recieved %#v", expected, reply.AccountsUsage[utils.MetaPrimary])
		}
	})

	t.Run("InterimUsage: parsing error", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:     "originID",
				utils.MetaAccounts:     true,
				utils.MetaDebit:        true,
				utils.MetaUsage:        90 * time.Second,
				utils.MetaInterimUsage: "err",
			},
		}
		var reply V1ProcessEventReply
		expErr := "can't convert <err> to decimal"
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %#v, recieved %#v", expErr, err)
		}
		expected := 0 * time.Second
		rcv := reply.AccountsUsage[utils.MetaPrimary]
		if !reflect.DeepEqual(rcv, expected) {
			t.Errorf("Expected %#v, recieved %#v", expected, rcv)
		}
	})

	t.Run("TotalUsage", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "originID",
				utils.MetaAccounts:   true,
				utils.MetaDebit:      true,
				utils.MetaUsage:      90 * time.Second,
				utils.MetaTotalUsage: 50,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := 90 * time.Second
		rcv := reply.AccountsUsage[utils.MetaPrimary]
		if !reflect.DeepEqual(rcv, expected) {
			t.Errorf("Expected %#v, recieved %#v", expected, rcv)
		}
	})

	t.Run("TotalUsage: parsing error", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "originID",
				utils.MetaAccounts:   true,
				utils.MetaDebit:      true,
				utils.MetaUsage:      90 * time.Second,
				utils.MetaTotalUsage: "err",
			},
		}
		var reply V1ProcessEventReply
		expErr := "can't convert <err> to decimal"
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != expErr {
			t.Errorf("Expected %#v, recieved %#v", expErr, err)
		}
		expected := 0 * time.Second
		rcv := reply.AccountsUsage[utils.MetaPrimary]
		if !reflect.DeepEqual(rcv, expected) {
			t.Errorf("Expected %#v, recieved %#v", expected, rcv)
		}
	})
}

func addInternalConn(sessions *SessionS, cfg *config.CGRConfig, flag, apiPrefix string, clnt *testMockClients) {
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	connID := utils.ConcatenatedKey(utils.MetaInternal, flag)
	cfg.SessionSCfg().Conns[flag] = []*config.DynamicConns{{ConnIDs: []string{connID}}}
	sessions.connMgr.AddInternalConn(connID, apiPrefix, chanInternal)
}

func TestSessionSBiRPCv1ProcessEventChargers(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]*chargers.ChrgSProcessEventReply) = []*chargers.ChrgSProcessEventReply{
					{
						ChargerSProfile: "CHRG1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "run1Ev",
							Event:  map[string]any{utils.AccountField: "1001"},
							APIOpts: map[string]any{
								utils.MetaOriginID: "originID",
								utils.MetaRunID:    "CHRG1",
							},
						},
					},
				}
				return nil
			},
		},
	}
	addInternalConn(sessions, cfg, utils.MetaChargers, utils.ChargerSv1, clnt)
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID",
		Event:  map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.MetaOriginID: "originID",
			utils.MetaSession:  true,
			utils.MetaChargers: true,
			utils.MetaUR:       true,
		},
	}
	var reply V1ProcessEventReply
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}
	if len(reply.UsageRecords) != 1 {
		t.Errorf("Expected 1 UsageRecords, recieved %d", len(reply.UsageRecords))
	}
	if _, ok := reply.UsageRecords["CHRG1"]; !ok {
		t.Errorf("Expected UsageRecord for CHRG1, recieved %v", reply.UsageRecords)
	}
}

func TestSessionSBiRPCv1ProcessEventRouting(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.RouteSv1GetRoutes: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*routes.SortedRoutesList) = routes.SortedRoutesList{{Routes: []*routes.SortedRoute{{RouteID: "RouteID"}}}}
				return nil
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"STQ1"}
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]string) = []string{"THD1"}
				return nil
			},
			utils.IPsV1AuthorizeIP: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.AllocatedIP) = utils.AllocatedIP{ProfileID: "prfIP"}
				return nil
			},
		},
	}
	addInternalConn(sessions, cfg, utils.MetaRoutes, utils.RouteSv1, clnt)
	addInternalConn(sessions, cfg, utils.MetaStats, utils.StatSv1, clnt)
	addInternalConn(sessions, cfg, utils.MetaThresholds, utils.ThresholdSv1, clnt)
	addInternalConn(sessions, cfg, utils.MetaIPs, utils.IPsV1, clnt)
	t.Run("Routes", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID: "originID",
				utils.MetaRoutes:   true,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := routes.SortedRoutesList{{Routes: []*routes.SortedRoute{{RouteID: "RouteID"}}}}
		if !reflect.DeepEqual(reply.RouteProfiles[utils.MetaPrimary], expected) {
			t.Errorf("Expected %v, recieved %v", expected, reply.RouteProfiles[utils.MetaPrimary])
		}
	})
	t.Run("Stats", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID: "originID",
				utils.MetaStats:    true,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(reply.StatQueueIDs[utils.MetaPrimary], []string{"STQ1"}) {
			t.Errorf("Expected [STQ1], recieved %#v", reply.StatQueueIDs[utils.MetaPrimary])
		}
	})
	t.Run("Thresholds", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:   "originID",
				utils.MetaThresholds: true,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(reply.ThresholdIDs[utils.MetaPrimary], []string{"THD1"}) {
			t.Errorf("Expected [THD1], recieved %#v", reply.ThresholdIDs[utils.MetaPrimary])
		}
	})

	t.Run("IPs", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:        "originID",
				utils.MetaIPsAuthorizeCfg: true,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := &utils.AllocatedIP{ProfileID: "prfIP"}
		if !reflect.DeepEqual(reply.IPsAllocation[utils.MetaPrimary], expected) {
			t.Errorf("Expected %#v, recieved %#v", expected, reply.IPsAllocation[utils.MetaPrimary])
		}
	})
}

func TestSessionSBiRPCv1ProcessEventResources(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.ResourceSv1AuthorizeResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.ResourceSv1AllocateResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
			utils.ResourceSv1ReleaseResources: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*string) = "OK"
				return nil
			},
		},
	}
	addInternalConn(sessions, cfg, utils.MetaResources, utils.ResourceSv1, clnt)
	t.Run("Authorize", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:              "originResAuth",
				utils.MetaResources:             true,
				utils.MetaResourcesAuthorizeCfg: true,
			}}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		if rcv := reply.ResourceAllocation[utils.MetaPrimary]; rcv != "OK" {
			t.Errorf("Expected OK, recieved %v", rcv)
		}
	})
	t.Run("Allocate", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:             "originResAlloc",
				utils.MetaResources:            true,
				utils.MetaResourcesAllocateCfg: true,
			}}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		if rcv := reply.ResourceAllocation[utils.MetaPrimary]; rcv != "OK" {
			t.Errorf("Expected OK, recieved %v", rcv)
		}
	})
	t.Run("Release", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:            "originResRelease",
				utils.MetaResources:           true,
				utils.MetaResourcesReleaseCfg: true,
			}}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
	})
}

func TestSessionSBiRPCv1ProcessEventEEs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.TODO()
	t.Run("ees", func(t *testing.T) {
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
		clnt := &testMockClients{
			calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
				utils.EeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
					*reply.(*map[string]map[string]any) = map[string]map[string]any{"EXPORTER1": {}}
					return nil
				},
			},
		}
		addInternalConn(sessions, cfg, utils.MetaEEs, utils.EeSv1, clnt)
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID: "originEEs",
				utils.MetaEEs:      true,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		exp := []string{"EXPORTER1"}
		if !reflect.DeepEqual(reply.EventExporters[utils.MetaPrimary], exp) {
			t.Errorf("Expected %v, recieved %v", exp, reply.EventExporters[utils.MetaPrimary])
		}
	})
	t.Run("Call returns error", func(t *testing.T) {
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
		dm := engine.NewDataManager(dbCM, cfg, nil, locker)
		dm.SetCache(cacheS)
		fltrS := engine.NewFilterS(cfg, nil, dm)
		connMgr := engine.NewConnManager(cfg)
		connMgr.SetCache(cacheS)
		sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
		clnt := &testMockClients{calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				return utils.ErrNotImplemented
			},
		}}
		addInternalConn(sessions, cfg, utils.MetaEEs, utils.EeSv1, clnt)
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event:  map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.MetaOriginID:        "originEEsWrap",
				utils.OptsSesBlockerError: true,
				utils.MetaEEs:             true,
			},
		}
		var reply2 V1ProcessEventReply
		expErr := "EES_ERROR:NOT_IMPLEMENTED"
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply2); err == nil || err.Error() != expErr {
			t.Errorf("Expected %v, received %v", expErr, err)
		}
	})
}

func TestSessionSBiRPCv1ProcessEventAccounting(t *testing.T) {
	ctx := context.TODO()
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.RateSv1CostForEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.RateProfileCost) = utils.RateProfileCost{Cost: utils.NewDecimalFromFloat64(1.0)}
				return nil
			},
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(int64(60*time.Second), 0)}
				return nil
			},
			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(int64(90*time.Second), 0)}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	addInternalConn(sessions, cfg, utils.MetaRates, utils.RateSv1, clnt)
	addInternalConn(sessions, cfg, utils.MetaAccounts, utils.AccountSv1, clnt)
	t.Run("Rates", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: "originID",
				utils.MetaRunID:    "run1",
				utils.MetaRates:    true,
				utils.MetaUsage:    1 * time.Minute,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		if _, ok := reply.RatesCost["run1"]; !ok {
			t.Error("Expected RatesCost for run1")
		}
	})
	t.Run("Authorize", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID:  "originID",
				utils.MetaAccounts:  true,
				utils.MetaAuthorize: true,
				utils.MetaUsage:     60 * time.Second,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := 60 * time.Second
		if !reflect.DeepEqual(reply.AccountsUsage[utils.MetaPrimary], expected) {
			t.Errorf("Expected %#v, got %#v", expected, reply.AccountsUsage[utils.MetaPrimary])
		}
	})
	t.Run("Debit", func(t *testing.T) {
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: "originID",
				utils.MetaAccounts: true,
				utils.MetaDebit:    true,
				utils.MetaUsage:    90 * time.Second,
			},
		}
		var reply V1ProcessEventReply
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
		expected := 90 * time.Second
		if !reflect.DeepEqual(reply.AccountsUsage[utils.MetaPrimary], expected) {
			t.Errorf("Expected %#v, got %#v", expected, reply.AccountsUsage[utils.MetaPrimary])
		}

		args.APIOpts = map[string]any{
			utils.MetaOriginID:         "originID",
			utils.OptsSesBlockerError:  false,
			utils.MetaAccountsDebitCfg: true,
			utils.MetaUsage:            0 * time.Minute,
		}
		if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
			t.Error(err)
		}
	})
}

func TestSessionSBiRPCv1ProcessEventSession(t *testing.T) {
	ctx := context.TODO()
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	clnt := &testMockClients{
		calls: map[string]func(ctx *context.Context, m string, args, reply any) error{
			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{Abstracts: utils.NewDecimal(int64(90*time.Second), 0)}
				return nil
			},
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, m string, args, reply any) error {
				*reply.(*[]*chargers.ChrgSProcessEventReply) = []*chargers.ChrgSProcessEventReply{
					{
						ChargerSProfile: "CHRG1",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "run1Ev",
							Event:  map[string]any{utils.AccountField: "1001"},
							APIOpts: map[string]any{
								utils.MetaOriginID: "originID",
								utils.MetaRunID:    "CHRG1",
							},
						},
					},
					{
						ChargerSProfile: "CHRG2",
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "run1Ev",
							Event:  map[string]any{utils.AccountField: "1001"},
							APIOpts: map[string]any{
								utils.MetaOriginID: "originID",
								utils.MetaRunID:    "CHRG2",
							},
						},
					},
				}
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- clnt
	addInternalConn(sessions, cfg, utils.MetaAccounts, utils.AccountSv1, clnt)
	addInternalConn(sessions, cfg, utils.MetaChargers, utils.ChargerSv1, clnt)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "sessInit",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: "originID",
			utils.MetaSession:  true,
			utils.MetaChargers: true,
		},
	}

	expS := &Session{
		ID: "63d6015b6fe2ea85029d91e6944f3f2223f7eedb",
		OriginCGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "sessInit",
			Event: map[string]any{
				"Account": "1001",
			},
			APIOpts: map[string]any{
				utils.MetaCGRid:    "63d6015b6fe2ea85029d91e6944f3f2223f7eedb",
				utils.MetaChargers: true,
				utils.MetaOriginID: "originID",
				utils.MetaRunID:    "*primary",
				utils.MetaSession:  true,
				utils.MetaUsage:    0,
			},
		},
		AutoChargeInterval: 0,
	}
	var reply V1ProcessEventReply
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}
	cgrId := args.APIOpts[utils.MetaCGRid].(string)
	sS := sessions.getActivateSession(cgrId)
	if !reflect.DeepEqual(utils.ToJSON(sS), utils.ToJSON(expS)) {
		t.Errorf("Expected %#+v, \nrecieved %#+v", expS, sS)
	}
	if got := len(sS.sRuns); got != 2 {
		t.Errorf("Expected 2 sRuns, got %d", got)
	}
	args.APIOpts[utils.MetaTerminate] = true
	if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err != nil {
		t.Error(err)
	}
	if got := len(sessions.getSessions("originID", false)); got != 0 {
		t.Errorf("Expected 0 active sessions, got %d", got)
	}
}

func TestSessionSBiRPCv1ProcessEventNonBlockingParseErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	locker := engine.NewLocker(cfg)
	data, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	sessions := NewSessionS(cfg, dm, cacheS, fltrS, connMgr)
	ctx := context.TODO()

	tests := []struct {
		flag        string
		blockerErr  any
		expectedErr string
	}{
		{
			flag:        utils.MetaChargers,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaSession,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaRoutes,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaStats,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaThresholds,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaIPs,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaRates,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaAccounts,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaResources,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaAuthorize,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaRefund,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaDebit,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaIPsAuthorizeCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaResourcesAuthorizeCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaResourcesAllocateCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaResourcesReleaseCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaAccountsAuthorizeCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaAccountsRefundCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaAccountsDebitCfg,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaUR,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaEEs,
			blockerErr:  true,
			expectedErr: `strconv.ParseBool: parsing "test": invalid syntax`,
		},
		{
			flag:        utils.MetaRoutes,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaStats,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaThresholds,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaIPs,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaRates,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaAccounts,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaResources,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaAuthorize,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaRefund,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaDebit,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaIPsAuthorizeCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaResourcesAuthorizeCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaResourcesAllocateCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaResourcesReleaseCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaAccountsAuthorizeCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaAccountsRefundCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaAccountsDebitCfg,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaUR,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
		{
			flag:        utils.MetaEEs,
			blockerErr:  false,
			expectedErr: "PARTIALLY_EXECUTED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			args := &utils.CGREvent{
				Tenant: "cgrates.org",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.MetaOriginID:        "originID",
					utils.OptsSesBlockerError: tt.blockerErr,
					tt.flag:                   "test",
				},
			}
			var reply V1ProcessEventReply
			if err := sessions.BiRPCv1ProcessEvent(ctx, args, &reply); err == nil || err.Error() != tt.expectedErr {
				t.Fatalf("Expected %v, received %v", tt.expectedErr, err)
			}
		})
	}
}
