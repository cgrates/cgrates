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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func Testv2ActionTriggerAsThreshold(t *testing.T) {
	var filters []*engine.FilterRule
	v2ATR := &v2ActionTrigger{
		ID:                "test2",              // original csv tag
		UniqueID:          "testUUID",           // individual id
		ThresholdType:     "*min_event_counter", //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
		ThresholdValue:    5.32,
		Recurrent:         false,                          // reset excuted flag each run
		MinSleep:          time.Duration(5) * time.Second, // Minimum duration between two executions in case of recurrent triggers
		ExpirationDate:    time.Now(),
		ActivationDate:    time.Now(),
		Balance:           new(engine.BalanceFilter),
		Weight:            0,
		ActionsID:         "Action1",
		MinQueuedItems:    10, // Trigger actions only if this number is hit (stats only)
		Executed:          false,
		LastExecutionTime: time.Now(),
	}
	x, _ := engine.NewFilterRule(engine.MetaRSR, "Directions", v2ATR.Balance.Directions.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(engine.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
	filters = append(filters, x)

	filter := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     *v2ATR.Balance.ID,
		Rules:  filters}

	thp := &engine.ThresholdProfile{
		ID:                 v2ATR.ID,
		Tenant:             config.CgrConfig().GeneralCfg().DefaultTenant,
		Blocker:            false,
		Weight:             v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{v2ATR.ExpirationDate, v2ATR.ActivationDate},
		MinSleep:           v2ATR.MinSleep,
	}
	th := &engine.Threshold{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v2ATR.ID,
	}

	newthp, newth, fltr, err := v2ATR.AsThreshold()
	if err != nil {
		t.Errorf("err")
	}
	if !reflect.DeepEqual(thp, newthp) {
		t.Errorf("Expecting: %+v, received: %+v", thp, newthp)
	}
	if !reflect.DeepEqual(th, newth) {
		t.Errorf("Expecting: %+v, received: %+v", th, newth)
	}
	if !reflect.DeepEqual(filter, fltr) {
		t.Errorf("Expecting: %+v, received: %+v", filter, fltr)
	}
}
