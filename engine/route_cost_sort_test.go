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
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestPopulateCostForRoutesConnRefused(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, connMgr, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:             "local",
				RateProfileIDs: []string{"RP_LOCAL"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesProfileIDs: []string{},
		},
	}
	extraOpts := &optsGetRoutes{}
	cfg.RouteSCfg().RateSConns = []string{"*localhost"}
	_, err := populateCostForRoutes(context.Background(), cfg, connMgr, fltrS, routes, ev, extraOpts)
	errExpect := "RATES_ERROR:dial tcp 127.0.0.1:2012: connect: connection refused"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestLeastCostSorterSortRoutesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, cM, nil)
	lcs := NewLeastCostSorter(cfg, cM, fltrS)

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := lcs.SortRoutes(context.Background(), "", map[string]*RouteWithWeight{}, &utils.CGREvent{}, &optsGetRoutes{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received <%+v>", expErr, err)
	}

}

func TestLeastCostSorterSortRoutesOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	cfg.RouteSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, args, reply interface{}) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, cc)

	fltrS := NewFilterS(cfg, cM, nil)
	lcs := NewLeastCostSorter(cfg, cM, fltrS)

	routeWW := map[string]*RouteWithWeight{
		"RW1": {
			Route: &Route{
				ID:              "RouteId",
				RouteParameters: "RouteParam",
				Weights:         utils.DynamicWeights{{Weight: 1}},
				Blockers:        utils.DynamicBlockers{{Blocker: false}},
				AccountIDs:      []string{"AccId"},
				RateProfileIDs:  []string{"RateProfileId"},
				ResourceIDs:     []string{"ResourceId"},
				StatIDs:         []string{"StatId"},
			},
			Weight: 1,
		},
		"RW2": {
			Route: &Route{
				ID:              "RouteId2",
				RouteParameters: "RouteParam2",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				Blockers:        utils.DynamicBlockers{{Blocker: false}},
				AccountIDs:      []string{"AccId2"},
				RateProfileIDs:  []string{"RateProfileId2"},
				ResourceIDs:     []string{"ResourceId2"},
				StatIDs:         []string{"StatId2"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.MetaUsage: 400,
		},
	}

	exp := &SortedRoutes{
		ProfileID: "profID1",
		Sorting:   utils.MetaLC,
		Routes: []*SortedRoute{
			{
				RouteID:         "RouteId2",
				RouteParameters: "RouteParam2",
				SortingData: map[string]interface{}{
					utils.AccountIDs: []string{},
					utils.Cost:       nil,
					utils.Weight:     10,
				},
			},
			{
				RouteID:         "RouteId",
				RouteParameters: "RouteParam",
				SortingData: map[string]interface{}{
					utils.AccountIDs: []string{},
					utils.Cost:       nil,
					utils.Weight:     1,
				},
			},
		},
	}
	if rcv, err := lcs.SortRoutes(context.Background(), "profID1", routeWW, ev, &optsGetRoutes{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expecting \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestHightCostSorterSortRoutesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, cM, nil)
	hcs := NewHighestCostSorter(cfg, cM, fltrS)

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := hcs.SortRoutes(context.Background(), "", map[string]*RouteWithWeight{}, &utils.CGREvent{}, &optsGetRoutes{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received <%+v>", expErr, err)
	}

}

func TestHightCostSorterSortRoutesOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	cfg.RouteSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, args, reply interface{}) error {
				return nil
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, cc)

	fltrS := NewFilterS(cfg, cM, nil)
	hcs := NewHighestCostSorter(cfg, cM, fltrS)

	routeWW := map[string]*RouteWithWeight{
		"RW1": {
			Route: &Route{
				ID:              "RouteId",
				RouteParameters: "RouteParam",
				Weights:         utils.DynamicWeights{{Weight: 1}},
				Blockers:        utils.DynamicBlockers{{Blocker: false}},
				AccountIDs:      []string{"AccId"},
				RateProfileIDs:  []string{"RateProfileId"},
				ResourceIDs:     []string{"ResourceId"},
				StatIDs:         []string{"StatId"},
			},
			Weight: 10,
		},
		"RW2": {
			Route: &Route{
				ID:              "RouteId2",
				RouteParameters: "RouteParam2",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				Blockers:        utils.DynamicBlockers{{Blocker: false}},
				AccountIDs:      []string{"AccId2"},
				RateProfileIDs:  []string{"RateProfileId2"},
				ResourceIDs:     []string{"ResourceId2"},
				StatIDs:         []string{"StatId2"},
			},
			Weight:  1,
			blocker: true,
		},
	}
	ev := &utils.CGREvent{
		APIOpts: map[string]interface{}{
			utils.MetaUsage: 400,
		},
	}

	exp := &SortedRoutes{
		ProfileID: "profID1",
		Sorting:   utils.MetaHC,
		Routes: []*SortedRoute{
			{
				RouteID:         "RouteId",
				RouteParameters: "RouteParam",
				SortingData: map[string]interface{}{
					utils.AccountIDs: []string{},
					utils.Cost:       nil,
					utils.Weight:     10,
				},
			},
			{
				RouteID:         "RouteId2",
				RouteParameters: "RouteParam2",
				SortingData: map[string]interface{}{
					utils.AccountIDs: []string{},
					utils.Cost:       nil,
					utils.Weight:     1,
					utils.Blocker:    true,
				},
			},
		},
	}
	if rcv, err := hcs.SortRoutes(context.Background(), "profID1", routeWW, ev, &optsGetRoutes{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(rcv)) {
		t.Errorf("Expecting \n<%+v>,\n received \n<%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestPopulateCostForRoutesGetDecimalBigOptsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	cfg.RouteSCfg().Opts.Usage = []*utils.DynamicDecimalBigOpt{
		{
			FilterIDs: []string{"*string.invalid:filter"},
			Tenant:    "cgrates.org",
			Value:     decimal.New(-1, 0),
		},
	}
	connMgr := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, connMgr, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:             "local",
				RateProfileIDs: []string{"RP_LOCAL"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	experr := `inline parse error for string: <*string.invalid:filter>`
	_, err := populateCostForRoutes(context.Background(), cfg, connMgr, fltrS, routes, ev, extraOpts)
	if err == nil || err.Error() != experr {
		t.Errorf("Expected error \n<%v>\n but received \n<%v>", experr, err)
	}

}

func TestPopulateCostForRoutesMissingIdsErr(t *testing.T) {

	defer func() {
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}

	connMgr := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, connMgr, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:             "local",
				RateProfileIDs: []string{},
				AccountIDs:     []string{},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	experr := `MANDATORY_IE_MISSING: [RateProfileIDs or AccountIDs]`
	_, err := populateCostForRoutes(context.Background(), cfg, connMgr, fltrS, routes, ev, extraOpts)
	if err == nil || err.Error() != experr {
		t.Errorf("Expected error \n<%v>\n but received \n<%v>", experr, err)
	}

}

func TestPopulateCostForRoutesAccountSConnsIgnoreErr(t *testing.T) {

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	cfg.RouteSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, cc)

	fltrS := NewFilterS(cfg, cM, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:         "local",
				AccountIDs: []string{"accID1"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{
		ignoreErrors: true,
	}

	rcv, err := populateCostForRoutes(context.Background(), cfg, cM, fltrS, routes, ev, extraOpts)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, []*SortedRoute{}) {
		t.Errorf("Received <%+v>", rcv)
	}

	expected := "CGRateS <> [WARNING] <RouteS> ignoring route with ID: local, err: NOT_IMPLEMENTED"
	if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected <%+v>, received <%+v>", expected, rcv)
	}

}

func TestPopulateCostForRoutesAccountSConnsErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	cfg.RouteSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1MaxAbstracts: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}

	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, cc)

	fltrS := NewFilterS(cfg, cM, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:         "local",
				AccountIDs: []string{"accID1"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEvent",
		Event: map[string]interface{}{
			utils.AccountField: 1001,
		},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	expErr := "ACCOUNTS_ERROR:NOT_IMPLEMENTED"
	_, err := populateCostForRoutes(context.Background(), cfg, cM, fltrS, routes, ev, extraOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error \n<%v>\n but received \n<%v>", expErr, err)
	}

}
