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
	"reflect"
	"sync"
	"testing"

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
		lm.incrementLoad(hst, utils.EmptyString)
	}
	// check only the first host because the rest may be in a random order
	// because they share the same cost
	if rply := lm.getHosts(hostsIDs.Clone()); rply[0] != "DSP_1" {
		t.Errorf("Expected: %q ,received: %q", "DSP_1", rply[0])
	}
	lm.incrementLoad(hostsIDs[0], utils.EmptyString)
	lm.decrementLoad(hostsIDs[1], utils.EmptyString)
	if rply := lm.getHosts(hostsIDs.Clone()); rply[0] != "DSP_2" {
		t.Errorf("Expected: %q ,received: %q", "DSP_2", rply[0])
	}
	for _, hst := range hostsIDs {
		lm.incrementLoad(hst, utils.EmptyString)
	}
	if rply := lm.getHosts(hostsIDs.Clone()); rply[0] != "DSP_2" {
		t.Errorf("Expected: %q ,received: %q", "DSP_2", rply[0])
	}
}

func TestNewSingleStrategyDispatcher(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
		{ID: "DSP_5"},
	}
	var exp strategyDispatcher = new(singleResultstrategyDispatcher)
	if rply, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{}, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  singleResultstrategyDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
		{ID: "DSP_5", Params: map[string]interface{}{utils.MetaRatio: 1}},
	}
	exp = &loadStrategyDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 1,
	}
	if rply, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{}, "cgrates.org"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadStrategyDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	dhp = engine.DispatcherHostProfiles{
		{ID: "DSP_1"},
		{ID: "DSP_2"},
		{ID: "DSP_3"},
		{ID: "DSP_4"},
	}
	exp = &loadStrategyDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 2,
	}
	if rply, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: 2}, "cgrates.org"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadStrategyDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	exp = &loadStrategyDispatcher{
		hosts:        dhp,
		tntID:        "cgrates.org",
		defaultRatio: 0,
	}
	if rply, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: 0}, "cgrates.org"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected:  loadStrategyDispatcher structure,received: %s", utils.ToJSON(rply))
	}

	if _, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: "A"}, "cgrates.org"); err == nil {
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
	if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
	for i := 0; i < 100; i++ {
		for _, dh := range dhp {
			for j := int64(0); j < lm.HostsRatio[dh.ID]; j++ {
				if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
					t.Errorf("Expected for id<%s>: %+v ,received: %+v", dh.ID, exp, rply)
				}
				lm.incrementLoad(dh.ID, utils.EmptyString)
			}
			exp = append(exp[1:], exp[0])
		}
		exp = []string{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5"}
		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.decrementLoad("DSP_4", utils.EmptyString)
		lm.decrementLoad("DSP_4", utils.EmptyString)
		lm.decrementLoad("DSP_2", utils.EmptyString)
		exp = []string{"DSP_2", "DSP_4", "DSP_1", "DSP_3", "DSP_5"}
		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad("DSP_2", utils.EmptyString)

		exp = []string{"DSP_4", "DSP_1", "DSP_2", "DSP_3", "DSP_5"}
		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad("DSP_4", utils.EmptyString)

		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad("DSP_4", utils.EmptyString)
		exp = []string{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5"}
		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
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
	if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
	for i := 0; i < 100; i++ {
		if rply := lm.getHosts(hostsIDs.Clone()); !reflect.DeepEqual(exp, rply) {
			t.Errorf("Expected: %+v ,received: %+v", exp, rply)
		}
		lm.incrementLoad(exp[0], utils.EmptyString)
	}
}

func TestLibDispatcherNewDispatcherMetaWeight(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaWeight,
	}
	result, err := newDispatcher(dataMng, pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	strategy, err := newSingleStrategyDispatcher(pfl.Hosts, pfl.StrategyParams, pfl.TenantID())
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &WeightDispatcher{
		hosts:    engine.DispatcherHostProfiles{},
		dm:       dataMng,
		strategy: strategy,
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).dm, expected.dm) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.dm, result.(*WeightDispatcher).dm)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).strategy, expected.strategy) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.strategy, result.(*WeightDispatcher).strategy)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*WeightDispatcher).hosts)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).tnt, expected.tnt) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", result.(*WeightDispatcher).tnt, expected.tnt)
	}
}

func TestLibDispatcherNewDispatcherMetaWeightErr(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaWeight,
	}
	_, err := newDispatcher(dataMng, pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherMetaRandom(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaRandom,
	}
	result, err := newDispatcher(dataMng, pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	strategy, err := newSingleStrategyDispatcher(pfl.Hosts, pfl.StrategyParams, pfl.TenantID())
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &RandomDispatcher{
		hosts:    engine.DispatcherHostProfiles{},
		dm:       dataMng,
		strategy: strategy,
	}
	if !reflect.DeepEqual(result.(*RandomDispatcher).dm, expected.dm) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.dm, result.(*WeightDispatcher).dm)
	}
	if !reflect.DeepEqual(result.(*RandomDispatcher).strategy, expected.strategy) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.strategy, result.(*WeightDispatcher).strategy)
	}
	if !reflect.DeepEqual(result.(*RandomDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*WeightDispatcher).hosts)
	}
	if !reflect.DeepEqual(result.(*RandomDispatcher).tnt, expected.tnt) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", result.(*WeightDispatcher).tnt, expected.tnt)
	}
}

func TestLibDispatcherNewDispatcherMetaRandomErr(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaRandom,
	}
	_, err := newDispatcher(dataMng, pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherMetaRoundRobin(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: utils.MetaRoundRobin,
	}
	result, err := newDispatcher(dataMng, pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	strategy, err := newSingleStrategyDispatcher(pfl.Hosts, pfl.StrategyParams, pfl.TenantID())
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	expected := &RoundRobinDispatcher{
		hosts:    engine.DispatcherHostProfiles{},
		dm:       dataMng,
		strategy: strategy,
	}
	if !reflect.DeepEqual(result.(*RoundRobinDispatcher).dm, expected.dm) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.dm, result.(*WeightDispatcher).dm)
	}
	if !reflect.DeepEqual(result.(*RoundRobinDispatcher).strategy, expected.strategy) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.strategy, result.(*WeightDispatcher).strategy)
	}
	if !reflect.DeepEqual(result.(*RoundRobinDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*WeightDispatcher).hosts)
	}
	if !reflect.DeepEqual(result.(*RoundRobinDispatcher).tnt, expected.tnt) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", result.(*WeightDispatcher).tnt, expected.tnt)
	}
}

func TestLibDispatcherNewDispatcherMetaRoundRobinErr(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{},
		StrategyParams: map[string]interface{}{
			utils.MetaDefaultRatio: false,
		},
		Strategy: utils.MetaRoundRobin,
	}
	_, err := newDispatcher(dataMng, pfl)
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherNewDispatcherPoolBroadcast(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: rpcclient.PoolBroadcast,
	}
	result, err := newDispatcher(dataMng, pfl)
	if err != nil {
		t.Errorf("\nExpected <nil>, \nReceived <%+v>", err)
	}
	strategy := &broadcastStrategyDispatcher{strategy: pfl.Strategy}
	expected := &WeightDispatcher{
		hosts:    engine.DispatcherHostProfiles{},
		dm:       dataMng,
		strategy: strategy,
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).dm, expected.dm) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.dm, result.(*WeightDispatcher).dm)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).strategy, expected.strategy) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.strategy, result.(*WeightDispatcher).strategy)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).hosts, expected.hosts) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected.hosts, result.(*WeightDispatcher).hosts)
	}
	if !reflect.DeepEqual(result.(*WeightDispatcher).tnt, expected.tnt) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", result.(*WeightDispatcher).tnt, expected.tnt)
	}
}

func TestLibDispatcherNewDispatcherError(t *testing.T) {
	dataMng := &engine.DataManager{}
	pfl := &engine.DispatcherProfile{
		Hosts:    engine.DispatcherHostProfiles{},
		Strategy: "badStrategy",
	}
	expected := "unsupported dispatch strategy: <badStrategy>"
	_, err := newDispatcher(dataMng, pfl)
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}

}

func TestLibDispatcherSetProfile(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "0",
				FilterIDs: []string{"FilterTest1"},
				Weight:    1,
				Params:    nil,
				Blocker:   false,
			},
		},
	}
	wgDsp := &WeightDispatcher{}
	wgDsp.SetProfile(pfl)
	expected := &engine.DispatcherProfile{
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:        "0",
				FilterIDs: []string{"FilterTest1"},
				Weight:    1,
				Params:    nil,
				Blocker:   false,
			},
		},
	}
	if !reflect.DeepEqual(expected, pfl) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(pfl))
	}
}

func TestLibDispatcherHostIDs(t *testing.T) {
	expected := engine.DispatcherHostIDs{"5", "10", "1"}

	wgDsp := &WeightDispatcher{
		RWMutex: sync.RWMutex{},
		dm:      nil,
		tnt:     "",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "5",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
			{
				ID:        "10",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
			{
				ID:        "1",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
		strategy: nil,
	}
	result := wgDsp.HostIDs()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

type strategyMockDispatcher struct{}

func (stMck strategyMockDispatcher) dispatch(dm *engine.DataManager, routeID string, subsystem, tnt string, hostIDs []string,
	serviceMethod string, args interface{}, reply interface{}) (err error) {
	return
}

func TestLibDispatcherDispatch(t *testing.T) {
	wgDsp := &WeightDispatcher{
		strategy: strategyMockDispatcher{},
	}
	result := wgDsp.Dispatch("", "", "", "", "")
	if !reflect.DeepEqual(nil, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestLibDispatcherRandomSetProfile(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: []*engine.DispatcherHostProfile{
			{
				ID:        "0",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
	}
	wgDsp := &RandomDispatcher{
		strategy: strategyMockDispatcher{},
	}
	wgDsp.SetProfile(pfl)
	expected := &engine.DispatcherProfile{
		Hosts: []*engine.DispatcherHostProfile{
			{
				ID:        "0",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
	}

	if !reflect.DeepEqual(pfl, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, pfl)
	}
}

func TestLibDispatcherRandomHostIDs(t *testing.T) {
	expected := engine.DispatcherHostIDs{"5"}
	wgDsp := &RandomDispatcher{
		RWMutex: sync.RWMutex{},
		dm:      nil,
		tnt:     "",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "5",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
		strategy: nil,
	}
	result := wgDsp.HostIDs()

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestLibDispatcherRandomDispatch(t *testing.T) {
	wgDsp := &RandomDispatcher{
		strategy: strategyMockDispatcher{},
	}
	result := wgDsp.Dispatch("", "", "", "", "")
	if !reflect.DeepEqual(nil, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestLibDispatcherRoundRobinSetProfile(t *testing.T) {
	pfl := &engine.DispatcherProfile{
		Hosts: []*engine.DispatcherHostProfile{
			{
				ID:        "0",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
	}
	wgDsp := &RoundRobinDispatcher{
		strategy: strategyMockDispatcher{},
	}
	wgDsp.SetProfile(pfl)
	expected := &engine.DispatcherProfile{
		Hosts: []*engine.DispatcherHostProfile{
			{
				ID:        "0",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
	}

	if !reflect.DeepEqual(pfl, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, pfl)
	}
}

func TestLibDispatcherRoundRobinHostIDs(t *testing.T) {
	expected := engine.DispatcherHostIDs{"5"}
	wgDsp := &RoundRobinDispatcher{
		RWMutex: sync.RWMutex{},
		dm:      nil,
		tnt:     "",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "5",
				FilterIDs: nil,
				Weight:    0,
				Params:    nil,
				Blocker:   false,
			},
		},
		strategy: nil,
	}
	result := wgDsp.HostIDs()

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestLibDispatcherRoundRobinDispatch(t *testing.T) {
	wgDsp := &RoundRobinDispatcher{
		strategy: strategyMockDispatcher{},
	}
	result := wgDsp.Dispatch("", "", "", "", "")
	if !reflect.DeepEqual(nil, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestLibDispatcherSingleResultstrategyDispatcherDispatch(t *testing.T) {
	wgDsp := &singleResultstrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "", "", "", []string{""}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherSingleResultstrategyDispatcherDispatchRouteID(t *testing.T) {
	wgDsp := &singleResultstrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "routeID", "", "", []string{""}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherBroadcastStrategyDispatcherDispatch(t *testing.T) {
	wgDsp := &broadcastStrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "", "", "", []string{""}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherBroadcastStrategyDispatcherDispatchRouteID(t *testing.T) {
	wgDsp := &broadcastStrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "routeID", "", "", []string{""}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatcherDispatch(t *testing.T) {
	wgDsp := &loadStrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "", "", "", []string{""}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatcherDispatchHostsID(t *testing.T) {
	wgDsp := &loadStrategyDispatcher{}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "routeID", "", "", []string{"hostID1", "hostID2"}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatchCaseHosts(t *testing.T) {
	wgDsp := &loadStrategyDispatcher{
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "testID",
				FilterIDs: []string{"filterID"},
				Weight:    4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
	}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dM := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)
	err := wgDsp.dispatch(dM, "", "", "", []string{"testID"}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherLoadStrategyDispatchCaseHostsError(t *testing.T) {
	wgDsp := &loadStrategyDispatcher{
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "testID2",
				FilterIDs: []string{"filterID"},
				Weight:    4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
	}
	err := wgDsp.dispatch(nil, "", "", "", []string{"testID2"}, "", "", "")
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
	wgDsp := &loadStrategyDispatcher{
		tntID: "testID",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "testID",
				FilterIDs: []string{"filterID"},
				Weight:    4,
				Params: map[string]interface{}{
					utils.MetaRatio: 1,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
	}
	err := wgDsp.dispatch(nil, "", "", "", []string{"testID"}, "", "", "")
	expected := "cannot cast false to *LoadMetrics"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherLoadStrategyDispatchCaseHostsCastError2(t *testing.T) {
	wgDsp := &loadStrategyDispatcher{
		tntID: "testID",
		hosts: engine.DispatcherHostProfiles{
			{
				ID:        "testID",
				FilterIDs: []string{"filterID"},
				Weight:    4,
				Params: map[string]interface{}{
					utils.MetaRatio: false,
				},
				Blocker: false,
			},
		},
		defaultRatio: 1,
	}
	err := wgDsp.dispatch(nil, "", "", "", []string{"testID"}, "", "", "")
	expected := "cannot convert field<bool>: false to int"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestLibDispatcherSingleResultStrategyDispatcherCastError(t *testing.T) {
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
	wgDsp := &singleResultstrategyDispatcher{}
	err := wgDsp.dispatch(nil, "testID", utils.MetaAttributes, "testTenant", []string{"testID"}, "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

type mockTypeCon struct{}

func (*mockTypeCon) Call(serviceMethod string, args, reply interface{}) error {
	return utils.ErrNotFound
}

func TestLibDispatcherSingleResultStrategyDispatcherCastError2(t *testing.T) {
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
	chanRPC := make(chan rpcclient.ClientConnector, 1)
	chanRPC <- new(mockTypeCon)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1Ping, chanRPC)
	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherRoutes, "testID:*attributes",
		value, nil, true, utils.NonTransactional)
	wgDsp := &singleResultstrategyDispatcher{}
	err := wgDsp.dispatch(nil, "testID", utils.MetaAttributes, "testTenant", []string{"testID"}, utils.AttributeSv1Ping, &utils.CGREvent{}, &wgDsp)
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
	engine.IntRPC = tmp
}

func TestLibDispatcherBroadcastStrategyDispatcherDispatchError1(t *testing.T) {
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
	wgDsp := &broadcastStrategyDispatcher{}
	err := wgDsp.dispatch(nil, "testID", utils.MetaAttributes, "testTenant", []string{"testID"}, "", "", "")
	expected := "DISPATCHER_ERROR:NO_DATABASE_CONNECTION"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherBroadcastStrategyDispatcherDispatchError2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache

	engine.Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, "testTenant:testID",
		nil, nil, true, utils.NonTransactional)
	wgDsp := &broadcastStrategyDispatcher{}
	err := wgDsp.dispatch(nil, "testID", utils.MetaAttributes, "testTenant", []string{"testID"}, "", "", "")
	expected := "HOST_NOT_FOUND"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
	engine.Cache = cacheInit
}

func TestLibDispatcherBroadcastStrategyDispatcherDispatchError3(t *testing.T) {
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
	wgDsp := &broadcastStrategyDispatcher{}
	err := wgDsp.dispatch(nil, "testID", utils.MetaAttributes, "testTenant", []string{"testID"}, "", "", "")
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	engine.Cache = cacheInit
}
