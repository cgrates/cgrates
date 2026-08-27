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
