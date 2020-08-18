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
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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

	if _, err := newSingleStrategyDispatcher(dhp, map[string]interface{}{utils.MetaDefaultRatio: "A"}, "cgrates.org"); err == nil {
		t.Fatalf("Expected error received: %v", err)
	}
}

func TestNewLoadMetrics(t *testing.T) {
	dhp := engine.DispatcherHostProfiles{
		{ID: "DSP_1", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_2", Params: map[string]interface{}{utils.MetaRatio: 1}},
		{ID: "DSP_3"},
	}
	exp := &LoadMetrics{
		HostsLoad: map[string]int64{},
		HostsRatio: map[string]int64{
			"DSP_1": 1,
			"DSP_2": 1,
			"DSP_3": 2,
		},
		SumRatio: 4,
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
