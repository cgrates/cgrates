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
	lm, err := newLoadMetrics(dhp)
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
