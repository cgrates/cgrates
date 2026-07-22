// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			{
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

func TestMigratorIsASAPreturn(t *testing.T) {
	actionPlan := &v1ActionPlan{}

	if actual := actionPlan.IsASAP(); actual {
		t.Errorf("Expected IsASAP to return false but got true")
	}
}
