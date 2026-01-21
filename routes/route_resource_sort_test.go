/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package routes

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestPopulateResourcesForRoutesNoResourceSConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	routes := map[string]*RouteWithWeight{}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	_, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	errExpect := "MANDATORY_IE_MISSING: [connIDs]"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestPopulateResourcesForRoutesNoResourceIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cM := engine.NewConnManager(cfg)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{},
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	_, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	errExpect := "MANDATORY_IE_MISSING: [ResourceIDs]"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestPopulateResourcesForRoutesOK(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: make(map[string]*utils.ResourceUsage),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *res
				return nil
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			Weight:  10,
			blocker: true,
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	exp := []*SortedRoute{
		{
			RouteID:         "Route1",
			RouteParameters: "param1",
			SortingData: map[string]any{
				utils.Blocker:          true,
				utils.ResourceUsageStr: 0,
				utils.Weight:           10,
			},
		},
	}

	rcv, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestPopulateResourcesForRoutesCallErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			Weight:  10,
			blocker: true,
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	exp := []*SortedRoute{
		{
			RouteID:         "Route1",
			RouteParameters: "param1",
			SortingData: map[string]any{
				utils.Blocker:          true,
				utils.ResourceUsageStr: 0,
				utils.Weight:           10,
			},
		},
	}

	rcv, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
	expected := "<RouteS> error: NOT_IMPLEMENTED getting resource for ID : RSC1"
	if log := buf.String(); !strings.Contains(log, expected) {
		t.Errorf("Expected <%+v>, received <%+v>", expected, log)
	}
}

func TestPopulateResourcesForRoutesLazyPassErr(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: make(map[string]*utils.ResourceUsage),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *res
				return nil
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			lazyCheckRules: []*engine.FilterRule{
				{
					Type:    "inexistent",
					Element: "inexistent",
					Values:  []string{"inexistent"},
				},
			},
			Weight:  10,
			blocker: true,
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	expErr := "NOT_IMPLEMENTED:inexistent"
	_, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
}

func TestResourceDescendentSorterSortRoutesNoResourceSConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	rds := NewResourceDescendentSorter(cfg, cM)

	routes := map[string]*RouteWithWeight{}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	_, err := rds.SortRoutes(context.Background(), "PROFILE1", routes, ev, extraOpts)
	errExpect := "MANDATORY_IE_MISSING: [connIDs]"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestResourceDescendentSorterSortRoutesNoResourceIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cM := engine.NewConnManager(cfg)
	rds := NewResourceDescendentSorter(cfg, cM)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{},
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	_, err := rds.SortRoutes(context.Background(), "PROFILE1", routes, ev, extraOpts)
	errExpect := "MANDATORY_IE_MISSING: [ResourceIDs]"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestResourceDescendentSorterSortRoutesOK(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res1 := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				ID:    "RU1",
				Units: 10,
			},
		},
	}
	res2 := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC2",
		Usages: map[string]*utils.ResourceUsage{
			"RU2": {
				ID:    "RU2",
				Units: 5,
			},
		},
	}
	res3 := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC3",
		Usages: map[string]*utils.ResourceUsage{
			"RU3": {
				ID:    "RU3",
				Units: 15,
			},
		},
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				argsCast := args.(*utils.TenantIDWithAPIOpts)
				switch argsCast.ID {
				case "RSC1":
					*rplCast = *res1
				case "RSC2":
					*rplCast = *res2
				case "RSC3":
					*rplCast = *res3
				}
				return nil
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	rds := NewResourceDescendentSorter(cfg, cM)

	routes := map[string]*RouteWithWeight{
		"RW1": {
			Route: &utils.Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			Weight: 10,
		},
		"RW2": {
			Route: &utils.Route{
				ID:              "Route2",
				ResourceIDs:     []string{"RSC2"},
				RouteParameters: "param2",
			},
			Weight: 20,
		},
		"RW3": {
			Route: &utils.Route{
				ID:              "Route3",
				ResourceIDs:     []string{"RSC3"},
				RouteParameters: "param3",
			},
			Weight:  15,
			blocker: true,
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	rcv, err := rds.SortRoutes(context.Background(), "PROFILE1", routes, ev, extraOpts)
	if err != nil {
		t.Fatal(err)
	}

	if rcv.ProfileID != "PROFILE1" {
		t.Errorf("Expected ProfileID <%s>, received <%s>", "PROFILE1", rcv.ProfileID)
	}

	if rcv.Sorting != utils.MetaReds {
		t.Errorf("Expected Sorting <%s>, received <%s>", utils.MetaReds, rcv.Sorting)
	}

	if len(rcv.Routes) != 3 {
		t.Fatalf("Expected 3 routes, received %d", len(rcv.Routes))
	}

	if rcv.Routes[0].RouteID != "Route3" {
		t.Errorf("Expected first route to be Route3, got %s", rcv.Routes[0].RouteID)
	}
	if rcv.Routes[1].RouteID != "Route1" {
		t.Errorf("Expected second route to be Route1, got %s", rcv.Routes[1].RouteID)
	}
	if rcv.Routes[2].RouteID != "Route2" {
		t.Errorf("Expected third route to be Route2, got %s", rcv.Routes[2].RouteID)
	}

	if _, hasBlocker := rcv.Routes[0].SortingData[utils.Blocker]; !hasBlocker {
		t.Errorf("Expected Route3 to have blocker flag")
	}
}

func TestResourceDescendentSorterSortRoutesEmptyRoutes(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	rds := NewResourceDescendentSorter(cfg, cM)

	routes := map[string]*RouteWithWeight{}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}
	extraOpts := &optsGetRoutes{}

	rcv, err := rds.SortRoutes(context.Background(), "PROFILE1", routes, ev, extraOpts)
	if err != nil {
		t.Fatal(err)
	}

	if rcv.ProfileID != "PROFILE1" {
		t.Errorf("Expected ProfileID <%s>, received <%s>", "PROFILE1", rcv.ProfileID)
	}

	if rcv.Sorting != utils.MetaReds {
		t.Errorf("Expected Sorting <%s>, received <%s>", utils.MetaReds, rcv.Sorting)
	}

	if len(rcv.Routes) != 0 {
		t.Errorf("Expected 0 routes, received %d", len(rcv.Routes))
	}
}

func TestResourceDescendentSorterSortRoutesSingleRoute(t *testing.T) {

	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				ID:    "RU1",
				Units: 25,
			},
		},
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *res
				return nil
			},
		},
	}
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	rds := NewResourceDescendentSorter(cfg, cM)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &utils.Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			Weight: 50,
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	exp := &SortedRoutes{
		ProfileID: "PROFILE1",
		Sorting:   utils.MetaReds,
		Routes: []*SortedRoute{
			{
				RouteID:         "Route1",
				RouteParameters: "param1",
				SortingData: map[string]any{
					utils.ResourceUsageStr: 25.0,
					utils.Weight:           50.0,
				},
			},
		},
	}

	rcv, err := rds.SortRoutes(context.Background(), "PROFILE1", routes, ev, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
