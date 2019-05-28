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
package migrator

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV1ActionPlanAsActionPlan(t *testing.T) {
	v1ap := &v1ActionPlan{
		Id:         "test",
		AccountIds: []string{"one"},
		Timing:     &engine.RateInterval{Timing: new(engine.RITiming)},
	}
	ap := &engine.ActionPlan{
		Id:         "test",
		AccountIDs: utils.StringMap{"one": true},
		ActionTimings: []*engine.ActionTiming{
			&engine.ActionTiming{
				Timing: &engine.RateInterval{
					Timing: new(engine.RITiming),
				},
			},
		},
	}
	newap := v1ap.AsActionPlan()
	if ap.Id != newap.Id || !reflect.DeepEqual(ap.AccountIDs, newap.AccountIDs) {
		t.Errorf("Expecting: %+v, received: %+v", *ap, newap)
	} else if !reflect.DeepEqual(ap.ActionTimings[0].Timing, newap.ActionTimings[0].Timing) {
		t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Timing, newap.ActionTimings[0].Timing)
	} else if ap.ActionTimings[0].Weight != newap.ActionTimings[0].Weight || ap.ActionTimings[0].ActionsID != newap.ActionTimings[0].ActionsID {
		t.Errorf("Expecting: %+v, received: %+v", ap.ActionTimings[0].Weight, newap.ActionTimings[0].Weight)
	}
}
