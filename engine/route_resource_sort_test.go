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
)

func TestPopulateResourcesForRoutesNoResourceSConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	routes := map[string]*RouteWithWeight{}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
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
	cM := NewConnManager(cfg)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{},
		},
	}
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: make(map[string]*ResourceUsage),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply interface{}) error {
				rplCast, canCast := reply.(*Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *res
				return nil
			},
		},
	}
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
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
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	exp := []*SortedRoute{
		{
			RouteID:         "Route1",
			RouteParameters: "param1",
			SortingData: map[string]interface{}{
				utils.Blocker:       true,
				utils.ResourceUsage: 0,
				utils.Weight:        10,
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
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
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	exp := []*SortedRoute{
		{
			RouteID:         "Route1",
			RouteParameters: "param1",
			SortingData: map[string]interface{}{
				utils.Blocker:       true,
				utils.ResourceUsage: 0,
				utils.Weight:        10,
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
		Cache = NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.RouteSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}

	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RSC1",
		Usages: make(map[string]*ResourceUsage),
	}

	cc := make(chan birpc.ClientConnector, 1)
	cc <- &ccMock{

		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply interface{}) error {
				rplCast, canCast := reply.(*Resource)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				*rplCast = *res
				return nil
			},
		},
	}
	cM := NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, cc)

	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:              "Route1",
				ResourceIDs:     []string{"RSC1"},
				RouteParameters: "param1",
			},
			lazyCheckRules: []*FilterRule{
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
		Event:   map[string]interface{}{},
		APIOpts: map[string]interface{}{},
	}
	extraOpts := &optsGetRoutes{}

	expErr := "NOT_IMPLEMENTED:inexistent"
	_, err := populateResourcesForRoutes(context.Background(), cfg, cM, routes, ev, extraOpts)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
}
