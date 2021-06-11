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
	"net/rpc"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestLoadMetricsGetHosts(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_2", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_3", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_4", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_5", Params: map[string]interface{}{utils.MetaRatio: 1}},
	}
	lm, err := newLoadMetrics(dhp, 1)
	if err != nil {
		t.Fatal(err)
	}
	hostsIDs := engine.DispatcherHostIDs(dhp.HostIDs())
	// to prevent randomness we increment all loads exept the first one
	for _, hst := range hostsIDs[1:] {
		lm.incrementLoad(context.Background(), hst, utils.EmptyString)
	}
	// check only the first host because the rest may be in a random order
	// because they share the same cost
	if rply := lm.getHosts(dhp.Clone()); rply[0].ID != "DSP_1" {
		t.Errorf("Expected: %q ,received: %q", "DSP_1", rply[0].ID)
	}
	lm.incrementLoad(context.Background(), hostsIDs[0], utils.EmptyString)
	lm.decrementLoad(context.Background(), hostsIDs[1], utils.EmptyString)
	if rply := lm.getHosts(dhp.Clone()); rply[0].ID != "DSP_2" {
		t.Errorf("Expected: %q ,received: %q", "DSP_2", rply[0].ID)
	}
	for _, hst := range hostsIDs {
		lm.incrementLoad(context.Background(), hst, utils.EmptyString)
	}
	if rply := lm.getHosts(dhp.Clone()); rply[0].ID != "DSP_2" {
		t.Errorf("Expected: %q ,received: %q", "DSP_2", rply[0].ID)
	}
}

func TestNewSingleDispatcher(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
		{ID: "DSP_5"},
	}
	var exp Dispatcher = &singleResultDispatcher{hosts: dhp}
	if rply, err := newSingleDispatcher(dhp, map[string]interface{}{}, utils.EmptyString, nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  singleResultDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
		{ID: "DSP_5", Params: map[string]interface{}{utils.MetaRatio: 1}},
	}
	exp = &loadDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 1,
	}
	if rply, err := newSingleDispatcher(dhp, map[string]interface{}{}, "cgrates.org", nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
	}
	exp = &loadDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 2,
	}
	if rply, err := newSingleDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: 2}, "cgrates.org", nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	exp = &loadDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 0,
	}
	if rply, err := newSingleDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: 0}, "cgrates.org", nil); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	if _, err := newSingleDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: "A"}, "cgrates.org", nil); err == nil {
		t.Fatalf("Expected error received: %v", err)
	}
}

func TestNewLoadMetrics(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_2", Params: map[string]interface{}{utils.MetaRatio: 0}},
		{ID: "DSP_3"},
	}
	exp := &LoadMetrics{
		HostsLoad: map[string]int64{},
		HostsRatio: map[string]int64{
			"DSP_1": 1,
			"DSP_2": 0,
			"DSP_3": 2,
		},
	}
	if lm, err := newLoadMetrics(dhp, 2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, lm) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(lm))
	}
	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: "A"}},
	}
	if _, err := newLoadMetrics(dhp, 2); err == nil {
		t.Errorf("Expected error received: %v", err)
	}
}

func TestLoadMetricsGetHosts2(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: 2}},
		{ID: "DSP_2", Params: map[string]interface{}{utils.MetaRatio: 3}},
		{ID: "DSP_3", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_4", Params: map[string]interface{}{utils.MetaRatio: 5}},
		{ID: "DSP_5", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_6", Params: map[string]interface{}{utils.MetaRatio: 0}},
	}
	lm, err := newLoadMetrics(dhp, 1)
	if err != nil {
		t.Fatal(err)
	}
	hostsIDs := engine.DispatcherHostIDs(dhp.HostIDs())
	exp := []string(hostsIDs.Clone())[:5]
	if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
	for i := 0; i < 100; i++ {
		for _, dh := range dhp {
			for j := int64(0); j < lm.HostsRatio[dh.ID]; j++ {
				if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
					t.Errorf("Expected for id<%s>: %+v ,received: %+v", dh.ID, exp, rply)
				}
				lm.incrementLoad(context.Background(), dh.ID, utils.EmptyString)
			}
			exp = append(exp[1:], exp[0])
		}
		exp = []string{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5"}
		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.decrementLoad(context.Background(), "DSP_4", utils.EmptyString)
		lm.decrementLoad(context.Background(), "DSP_4", utils.EmptyString)
		lm.decrementLoad(context.Background(), "DSP_2", utils.EmptyString)
		exp = []string{"DSP_2", "DSP_4", "DSP_1", "DSP_3", "DSP_5"}
		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad(context.Background(), "DSP_2", utils.EmptyString)

		exp = []string{"DSP_4", "DSP_1", "DSP_2", "DSP_3", "DSP_5"}
		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad(context.Background(), "DSP_4", utils.EmptyString)

		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad(context.Background(), "DSP_4", utils.EmptyString)
		exp = []string{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5"}
		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
	}

	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: -1}},
		{ID: "DSP_2", Params: map[string]interface{}{utils.MetaRatio: 3}},
		{ID: "DSP_3", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_4", Params: map[string]interface{}{utils.MetaRatio: 5}},
		{ID: "DSP_5", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_6", Params: map[string]interface{}{utils.MetaRatio: 0}},
	}
	lm, err = newLoadMetrics(dhp, 1)
	if err != nil {
		t.Fatal(err)
	}
	hostsIDs = engine.DispatcherHostIDs(dhp.HostIDs())
	exp = []string(hostsIDs.Clone())[:5]
	if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
	for i := 0; i < 100; i++ {
		if rply := lm.getHosts(dhp.Clone()); !reflect.DeepEqual(exp, rply.HostIDs()) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad(context.Background(), exp[0], utils.EmptyString)
	}
}

func TestLibDispatcherNewDispatcherMetaWeight(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaWeight,
	}
	result, err := newDispatcher(pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &singleResultDispatcher{
		hosts:  engine.DispatcherHostProfiles{},
		sorter: new(noSort),
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*singleResultDispatcher).hosts)
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).sorter, expected.sorter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", result.(*singleResultDispatcher).sorter, expected.sorter)
	}
}

func TestLibDispatcherNewDispatcherMetaWeightErr(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaWeight,
	}
	_, err := newDispatcher(pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherMetaRandom(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaRandom,
	}
	result, err := newDispatcher(pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &singleResultDispatcher{
		hosts:  engine.DispatcherHostProfiles{},
		sorter: new(randomSort),
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).sorter, expected.sorter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.sorter, result.(*singleResultDispatcher).sorter)
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*singleResultDispatcher).hosts)
	}
}

func TestLibDispatcherNewDispatcherMetaRandomErr(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaRandom,
	}
	_, err := newDispatcher(pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherMetaRoundRobin(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaRoundRobin,
	}
	result, err := newDispatcher(pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &singleResultDispatcher{
		hosts:  engine.DispatcherHostProfiles{},
		sorter: new(roundRobinSort),
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).sorter, expected.sorter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.sorter, result.(*singleResultDispatcher).sorter)
	}
	if !reflect.DeepEqual(result.(*singleResultDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*singleResultDispatcher).hosts)
	}
}

func TestLibDispatcherNewDispatcherMetaRoundRobinErr(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaRoundRobin,
	}
	_, err := newDispatcher(pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherPoolBroadcast(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: rpcclient.PoolBroadcast,
	}
	result, err := newDispatcher(pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &broadcastDispatcher{
		hosts:    engine.DispatcherHostProfiles{},
		strategy: pfl.Strategy,
	}
	if !reflect.DeepEqual(result.(*broadcastDispatcher).strategy, expected.strategy) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.strategy, result.(*broadcastDispatcher).strategy)
	}
	if !reflect.DeepEqual(result.(*broadcastDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*broadcastDispatcher).hosts)
	}
}

func TestLibDispatcherNewDispatcherError(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: "badStrategy",
	}
	expected := "unsupported dispatch strategy: <badStrategy>"
	_, err := newDispatcher(pfl)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherSingleResultDispatcherDispatch(t *testing.T) {
	wgDsp := &singleResultDispatcher{sorter: new(noSort)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherSingleResultDispatcherDispatchRouteID(t *testing.T) {
	wgDsp := &singleResultDispatcher{sorter: new(roundRobinSort)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "routeID", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherBroadcastDispatcherDispatch(t *testing.T) {
	wgDsp := &broadcastDispatcher{hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherBroadcastDispatcherDispatchRouteID(t *testing.T) {
	wgDsp := &broadcastDispatcher{hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "routeID", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadDispatcherDispatch(t *testing.T) {
	wgDsp := &loadDispatcher{sorter: new(randomSort)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadDispatcherDispatchHostsID(t *testing.T) {
	wgDsp := &loadDispatcher{
		hosts: engine.DispatcherHostProfiles{
			{ID: "hostID1"},
			{ID: "hostID2"},
		},
		sorter: new(noSort),
	}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "routeID", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatchCaseHosts(t *testing.T) {
	wgDsp := &loadDispatcher{
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID",
				// FilterIDs: []string{"filterID"},
				Weight: 4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
		sorter:       new(noSort),
	}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.Dispatch(dM, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatchCaseHostsError(t *testing.T) {
	wgDsp := &loadDispatcher{
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID2",
				// FilterIDs: []string{"filterID"},
				Weight: 4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatchCaseHostsCastError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	newCache := engine.NewCacheS(cfg, nil, nil)
	engine.Cache = newCache
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherLoads, "testID",
		false, nil, true, utils.NonTransactional)
	wgDsp := &loadDispatcher{
		tntID: "testID",
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID",
				// FilterIDs: []string{"filterID"},
				Weight: 4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "cannot cast false to *LoadMetrics"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherLoadStrategyDispatchCaseHostsCastError2(t *testing.T) {
	wgDsp := &loadDispatcher{
		tntID: "testID",
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID",
				// FilterIDs: []string{"filterID"},
				Weight: 4,
				Params: map[string]interface{}{
					utils.MetaRatio: false,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "", "", "", "", "", "")
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherSingleResultDispatcherCastError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     "",
			Transport:   "",
			Synchronous: false,
			TLS:         false,
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "", "testID", utils.MetaAttributes, "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon struct{}

func (*mockTypeCon) Call(ctx *context.Context, method string, args interface{}, reply interface{}) error {
	return utils.ErrNotFound
}

func TestLibDispatcherSingleResultDispatcherCastError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}

	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1Ping, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

func TestLibDispatcherBroadcastDispatcherDispatchError1(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     "",
			Transport:   "",
			Synchronous: false,
			TLS:         false,
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &broadcastDispatcher{hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherBroadcastDispatcherDispatchError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache

	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		nil, nil, true, utils.NonTransactional)
	wgDsp := &broadcastDispatcher{hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherBroadcastDispatcherDispatchError3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     "",
			Transport:   "",
			Synchronous: false,
			TLS:         false,
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &broadcastDispatcher{hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, "", "", "")
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherLoadDispatcherCacheError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     "",
			Transport:   "",
			Synchronous: false,
			TLS:         false,
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &loadDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherLoadDispatcherCacheError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}

	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1Ping, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &loadDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

func TestLibDispatcherLoadDispatcherCacheError3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}

	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeCon)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTENANT:testID",
		value, nil, true, utils.NonTransactional)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1Ping, chanRPC)
	wgDsp := &loadDispatcher{
		tntID: "testTENANT",
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID",
				// FilterIDs: []string{"filterID1", "filterID2"},
				Weight: 3,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: true,
			},
			{
				ID: "testID2",
				// FilterIDs: []string{"filterID1", "filterID2"},
				Weight: 3,
				Params: map[string]interface{}{
					utils.MetaRatio: 2,
				},
				Blocker: true,
			},
		},
		defaultRatio: 0,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(dm, nil, context.Background(), nil, "testTENANT", "testID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

func TestLibDispatcherLoadDispatcherCacheError4(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"con"}
	cfg.CacheCfg().Partitions[utils.CacheDispatcherRoutes].Replicate = true
	cfg.RPCConns()["con"] = &config.RPCConn{
		Strategy: "",
		PoolSize: 0,
		Conns: []*config.RemoteHost{
			{
				ID:          "testID",
				Address:     "",
				Transport:   "",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dm := engine.NewDataManager(nil, nil, connMng)

	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}

	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTENANT:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &loadDispatcher{
		tntID: "testTENANT",
		hosts: engine.DispatcherHostProfiles{
			{
				ID: "testID",
				// FilterIDs: []string{"filterID1", "filterID2"},
				Weight: 3,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: true,
			},
			{
				ID: "testID2",
				// FilterIDs: []string{"filterID1", "filterID2"},
				Weight: 3,
				Params: map[string]interface{}{
					utils.MetaRatio: 2,
				},
				Blocker: true,
			},
		},
		defaultRatio: 0,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(dm, nil, context.Background(), nil, "testTENANT", "testID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	expected := "DISCONNECTED"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeConDispatch struct{}

func (*mockTypeConDispatch) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	return rpc.ErrShutdown
}

func TestLibDispatcherLoadDispatcherCacheError5(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()

	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}

	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeConDispatch)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &loadDispatcher{
		tntID: "testTenant",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:     "testID",
				Weight: 3,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: true,
			},
		},
		defaultRatio: 0,
		sorter:       new(noSort),
	}
	err := wgDsp.Dispatch(nil, nil, context.Background(), nil, "testTenant", "testID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "connection is shut down", err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}
func TestLibDispatcherSingleResultDispatcherCase1(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeConDispatch)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(dm, nil, context.Background(), nil, "testTenant", "", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	if err == nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "connection is shut down", err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

type mockTypeConDispatch2 struct{}

func (*mockTypeConDispatch2) Call(ctx *context.Context, serviceMethod string, args, reply interface{}) error {
	return nil
}

func TestLibDispatcherSingleResultDispatcherCase2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeConDispatch2)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(dm, nil, context.Background(), nil, "testTenant", "routeID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

func TestLibDispatcherSingleResultDispatcherCase3(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"con"}
	cfg.CacheCfg().Partitions[utils.CacheDispatcherRoutes].Replicate = true
	cfg.RPCConns()["con"] = &config.RPCConn{
		Strategy: "",
		PoolSize: 0,
		Conns: []*config.RemoteHost{
			{
				ID:          "testID",
				Address:     "",
				Transport:   "",
				Synchronous: false,
				TLS:         false,
			},
		},
	}
	rpcCl := map[string]chan birpc.ClientConnector{}
	connMng := engine.NewConnManager(cfg, rpcCl)
	dm := engine.NewDataManager(nil, nil, connMng)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	value := &engine.DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:          "testID",
			Address:     rpcclient.InternalRPC,
			Transport:   utils.MetaInternal,
			Synchronous: false,
			TLS:         false,
		},
	}
	tmp := engine.IntRPC
	engine.IntRPC = map[string]*rpcclient.RPCClient{}
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- new(mockTypeConDispatch2)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultDispatcher{sorter: new(noSort), hosts: engine.DispatcherHostProfiles{{ID: "testID"}}}
	err := wgDsp.Dispatch(dm, nil, context.Background(), nil, "testTenant", "routeID", utils.MetaAttributes, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	expected := "DISCONNECTED"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}
