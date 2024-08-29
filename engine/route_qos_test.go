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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestPopulatStatsForQOSRouteCallErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := populatStatsForQOSRoute(context.Background(), cfg, cM, []string{"stat1", "stat2"}, "cgrates.org"); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

}

func TestPopulatStatsForQOSRouteOK(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	metrics := &map[string]*utils.Decimal{
		"stat": utils.NewDecimal(5, 0),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*map[string]*utils.Decimal)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *metrics
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, cc)

	exp := map[string]*utils.Decimal{
		"stat": utils.NewDecimal(5, 0),
	}

	if rcv, err := populatStatsForQOSRoute(context.Background(), cfg, cM, []string{"stat1", "stat2"}, "cgrates.org"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", exp, rcv)
	}

}

func TestQOSRouteSorterRoutesNoStatSConns(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	qos := NewQOSRouteSorter(cfg, cM)
	ctx := context.Background()
	prflID := "prfId"
	routes := map[string]*RouteWithWeight{}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	extraOpts := &optsGetRoutes{}

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := qos.SortRoutes(ctx, prflID, routes, cgrEv, extraOpts); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

}

func TestQOSRouteSorterRoutesOK(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	metrics := &map[string]*utils.Decimal{
		"*tcd": utils.NewDecimal(5, 0),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*map[string]*utils.Decimal)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *metrics
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, cc)
	qos := NewQOSRouteSorter(cfg, cM)
	ctx := context.Background()
	prflID := "prfId"
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:      "local",
				StatIDs: []string{"stat1"},
			},
			blocker: true,
			Weight:  10,
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	extraOpts := &optsGetRoutes{
		sortingParameters: []string{"param1", utils.MetaTCD, utils.MetaPDD},
	}

	exp := &SortedRoutes{
		ProfileID: "prfId",
		Sorting:   "*qos",
		Routes: []*SortedRoute{{
			RouteID:         "local",
			RouteParameters: "",
			SortingData: map[string]any{
				utils.MetaPDD: 1.7976931348623157e+308,
				utils.Blocker: true,
				utils.MetaTCD: 5,
				utils.Weight:  10,
				"param1":      -1,
			},
		},
		},
	}

	if rcv, err := qos.SortRoutes(ctx, prflID, routes, cgrEv, extraOpts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestQOSRouteSorterRoutesLazyPassErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	metrics := &map[string]*utils.Decimal{
		"*tcd": utils.NewDecimal(5, 0),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*map[string]*utils.Decimal)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *metrics
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, cc)
	qos := NewQOSRouteSorter(cfg, cM)
	ctx := context.Background()
	prflID := "prfId"
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:      "local",
				StatIDs: []string{"stat1"},
			},
			Weight: 10,
			lazyCheckRules: []*FilterRule{
				{
					Type:    "inexistent",
					Element: "inexistent",
					Values:  []string{"inexistent"},
				},
			},
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	extraOpts := &optsGetRoutes{}

	expErr := "NOT_IMPLEMENTED:inexistent"
	if _, err := qos.SortRoutes(ctx, prflID, routes, cgrEv, extraOpts); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

}

func TestQOSRouteSorterRoutesIgnoreErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, cc)
	qos := NewQOSRouteSorter(cfg, cM)
	ctx := context.Background()
	prflID := "prfId"
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:      "local",
				StatIDs: []string{"stat1"},
			},
			Weight: 10,
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	extraOpts := &optsGetRoutes{
		ignoreErrors: true,
	}

	exp := &SortedRoutes{
		ProfileID: "prfId",
		Sorting:   "*qos",
		Routes:    []*SortedRoute{},
	}

	if rcv, err := qos.SortRoutes(ctx, prflID, routes, cgrEv, extraOpts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expected \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestQOSRouteSorterRoutesPopulateErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueDecimalMetrics: func(ctx *context.Context, args, reply any) error {
				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, cc)
	qos := NewQOSRouteSorter(cfg, cM)
	ctx := context.Background()
	prflID := "prfId"
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:      "local",
				StatIDs: []string{"stat1"},
			},
			Weight: 10,
		},
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	extraOpts := &optsGetRoutes{}

	if _, err := qos.SortRoutes(ctx, prflID, routes, cgrEv, extraOpts); err != utils.ErrNotImplemented {
		t.Errorf("Expected error \n<%v>\n but received \n<%v>", utils.ErrNotImplemented, err)
	}
}
